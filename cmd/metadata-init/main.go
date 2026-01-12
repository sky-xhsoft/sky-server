package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/go-sql-driver/mysql"
)

// TableInfo 表信息
type TableInfo struct {
	TableName    string
	TableComment string
}

// ColumnInfo 字段信息
type ColumnInfo struct {
	ColumnName    string
	DataType      string
	ColumnType    string
	ColumnComment string
	IsNullable    string
	ColumnKey     string
	ColumnDefault *string
	CharMaxLength *int
}

// 命令行参数
var (
	excludeSys = flag.Bool("exclude-sys", false, "排除 sys_ 开头的系统表")
	onlySys    = flag.Bool("only-sys", false, "只初始化 sys_ 开头的系统表")
	tables     = flag.String("tables", "", "指定要初始化的表名（逗号分隔），如：user,order,product")
	force      = flag.Bool("force", false, "强制重新初始化已存在的表")
	initDB     = flag.Bool("init-db", false, "在元数据初始化前先执行 sqls/init.sql 初始化数据库")
	help       = flag.Bool("help", false, "显示帮助信息")
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 显示帮助信息
	if *help {
		printHelp()
		os.Exit(0)
	}

	// 参数冲突检查
	if *excludeSys && *onlySys {
		fmt.Println("错误: --exclude-sys 和 --only-sys 不能同时使用")
		os.Exit(1)
	}
	// 1. 加载配置
	cfg, e := config.Load()
	if e != nil {
		fmt.Printf("Failed to load config: %v\n", e)
		os.Exit(1)
	}

	// 2. 初始化日志
	log, e := logger.Init(&cfg.Log)
	if e != nil {
		fmt.Printf("Failed to initialize logger: %v\n", e)
		os.Exit(1)
	}
	defer log.Sync()

	logger.Info("Starting metadata initialization")

	// 3. 初始化数据库连接
	db, e := mysql.Init(&cfg.Database.MySQL, log)
	if e != nil {
		logger.Fatal("Failed to connect to database", zap.Error(e))
	}
	defer mysql.Close()
	logger.Info("Database connected successfully")

	ctx := context.Background()

	// 4. 获取数据库名称
	dbName := cfg.Database.MySQL.Database
	logger.Info("Database name", zap.String("database", dbName))

	// 5. 执行 init.sql（如果指定）
	if *initDB {
		logger.Info("Executing init.sql before metadata initialization")
		if err := executeInitSQL(&cfg.Database.MySQL); err != nil {
			logger.Fatal("Failed to execute init.sql", zap.Error(err))
		}
		logger.Info("init.sql executed successfully")
	}

	// 6. 初始化基础数据字典
	if err := initBaseDictionaries(ctx, db); err != nil {
		logger.Fatal("Failed to initialize base dictionaries", zap.Error(err))
	}
	logger.Info("Base dictionaries initialized")

	// 7. 初始化 sys_directory（为 sys_table 创建对应的安全目录）
	if err := initDirectoriesFromTables(ctx, db); err != nil {
		logger.Fatal("Failed to initialize directories from tables", zap.Error(err))
	}
	logger.Info("Directories initialized from sys_table")

	// 8. 获取要初始化的表
	var tableList []TableInfo
	var err error

	if *tables != "" {
		// 指定了具体的表名
		tableNames := strings.Split(*tables, ",")
		for i := range tableNames {
			tableNames[i] = strings.TrimSpace(tableNames[i])
		}
		tableList, err = getSpecificTables(ctx, db, dbName, tableNames)
		if err != nil {
			logger.Fatal("Failed to get specified tables", zap.Error(err))
		}
		logger.Info("Found specified tables", zap.Int("count", len(tableList)))
	} else {
		// 获取所有表或根据过滤条件获取
		tableList, err = getTables(ctx, db, dbName, *excludeSys, *onlySys)
		if err != nil {
			logger.Fatal("Failed to get tables", zap.Error(err))
		}

		filterInfo := "all tables"
		if *excludeSys {
			filterInfo = "business tables (excluding sys_*)"
		} else if *onlySys {
			filterInfo = "system tables (sys_* only)"
		}
		logger.Info("Found tables", zap.Int("count", len(tableList)), zap.String("filter", filterInfo))
	}

	// 9. 为每个表初始化元数据
	successCount := 0
	failCount := 0
	skippedCount := 0

	for _, table := range tableList {
		logger.Info("Processing table", zap.String("table", table.TableName))

		err := initTableMetadata(ctx, db, dbName, table, *force)
		if err != nil {
			if err.Error() == "table already exists" {
				logger.Info("Table already exists in sys_table, skipping",
					zap.String("table", table.TableName))
				skippedCount++
				continue
			}

			logger.Error("Failed to initialize table metadata",
				zap.String("table", table.TableName),
				zap.Error(err),
			)
			failCount++
			continue
		}

		logger.Info("Table metadata initialized", zap.String("table", table.TableName))
		successCount++
	}

	// 10. 输出结果
	logger.Info("Metadata initialization completed",
		zap.Int("success", successCount),
		zap.Int("skipped", skippedCount),
		zap.Int("failed", failCount),
		zap.Int("total", len(tableList)),
	)

	// 11. 再次初始化 sys_directory（为新增的表创建目录）
	if err := initDirectoriesFromTables(ctx, db); err != nil {
		logger.Error("Failed to initialize directories after metadata creation", zap.Error(err))
	} else {
		logger.Info("Directories synchronized after metadata creation")
	}
}

// printHelp 显示帮助信息
func printHelp() {
	fmt.Println("元数据初始化工具 - Sky-Server Metadata Initializer")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  metadata-init [参数]")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  --exclude-sys      排除 sys_ 开头的系统表（只初始化业务表）")
	fmt.Println("  --only-sys         只初始化 sys_ 开头的系统表")
	fmt.Println("  --tables <names>   指定要初始化的表名（逗号分隔），如：user,order,product")
	fmt.Println("  --force            强制重新初始化已存在的表（会删除原有元数据）")
	fmt.Println("  --init-db          在元数据初始化前先执行 sqls/init.sql 初始化数据库")
	fmt.Println("  --help             显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 初始化所有表（默认）")
	fmt.Println("  metadata-init")
	fmt.Println()
	fmt.Println("  # 只初始化业务表")
	fmt.Println("  metadata-init --exclude-sys")
	fmt.Println()
	fmt.Println("  # 只初始化系统表")
	fmt.Println("  metadata-init --only-sys")
	fmt.Println()
	fmt.Println("  # 初始化指定的表")
	fmt.Println("  metadata-init --tables user,order,product")
	fmt.Println()
	fmt.Println("  # 强制重新初始化所有表")
	fmt.Println("  metadata-init --force")
	fmt.Println()
	fmt.Println("  # 先执行 init.sql 初始化数据库，再初始化元数据")
	fmt.Println("  metadata-init --init-db")
	fmt.Println()
	fmt.Println("注意:")
	fmt.Println("  - 已存在的表默认会跳过，使用 --force 参数可强制重新初始化")
	fmt.Println("  - --exclude-sys 和 --only-sys 不能同时使用")
	fmt.Println("  - 指定 --tables 时会忽略 --exclude-sys 和 --only-sys")
	fmt.Println("  - --init-db 会执行 sqls/init.sql，请确保该文件存在且内容正确")
}

// getTables 获取表（根据过滤条件）
func getTables(ctx context.Context, db *gorm.DB, dbName string, excludeSys, onlySys bool) ([]TableInfo, error) {
	var tables []TableInfo

	query := `
		SELECT
			TABLE_NAME as table_name,
			COALESCE(TABLE_COMMENT, '') as table_comment
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_TYPE = 'BASE TABLE'
	`

	// 根据参数添加过滤条件
	if excludeSys {
		query += " AND TABLE_NAME NOT LIKE 'sys_%'"
	} else if onlySys {
		query += " AND TABLE_NAME LIKE 'sys_%'"
	}
	// 如果都不指定，则返回所有表

	query += " ORDER BY TABLE_NAME"

	err := db.WithContext(ctx).Raw(query, dbName).Scan(&tables).Error
	return tables, err
}

// getSpecificTables 获取指定的表
func getSpecificTables(ctx context.Context, db *gorm.DB, dbName string, tableNames []string) ([]TableInfo, error) {
	if len(tableNames) == 0 {
		return []TableInfo{}, nil
	}

	var tables []TableInfo

	query := `
		SELECT
			TABLE_NAME as table_name,
			COALESCE(TABLE_COMMENT, '') as table_comment
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_TYPE = 'BASE TABLE'
		  AND TABLE_NAME IN (?)
		ORDER BY TABLE_NAME
	`

	err := db.WithContext(ctx).Raw(query, dbName, tableNames).Scan(&tables).Error
	return tables, err
}

// getColumns 获取表的所有字段
func getColumns(ctx context.Context, db *gorm.DB, dbName, tableName string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	query := `
		SELECT
			COLUMN_NAME as column_name,
			DATA_TYPE as data_type,
			COLUMN_TYPE as column_type,
			COALESCE(COLUMN_COMMENT, '') as column_comment,
			IS_NULLABLE as is_nullable,
			COLUMN_KEY as column_key,
			COLUMN_DEFAULT as column_default,
			CHARACTER_MAXIMUM_LENGTH as char_max_length
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	err := db.WithContext(ctx).Raw(query, dbName, tableName).Scan(&columns).Error
	return columns, err
}

// initTableMetadata 初始化表的元数据
func initTableMetadata(ctx context.Context, db *gorm.DB, dbName string, table TableInfo, force bool) error {
	// 1. 检查表是否已存在于sys_table
	var existingTableID uint
	err := db.WithContext(ctx).Table("sys_table").
		Select("ID").
		Where("NAME = ?", strings.ToUpper(table.TableName)).
		Scan(&existingTableID).Error
	if err != nil {
		return fmt.Errorf("check table existence failed: %w", err)
	}

	if existingTableID > 0 {
		if !force {
			// 不强制更新，返回特殊错误
			return fmt.Errorf("table already exists")
		}

		// 强制更新：删除旧的元数据
		logger.Info("Force mode enabled, deleting existing metadata",
			zap.String("table", table.TableName),
			zap.Uint("table_id", existingTableID))

		err = db.Transaction(func(tx *gorm.DB) error {
			// 删除字段记录
			if err := tx.Exec("DELETE FROM sys_column WHERE SYS_TABLE_ID = ?", existingTableID).Error; err != nil {
				return fmt.Errorf("delete columns failed: %w", err)
			}

			// 删除表记录
			if err := tx.Exec("DELETE FROM sys_table WHERE ID = ?", existingTableID).Error; err != nil {
				return fmt.Errorf("delete table failed: %w", err)
			}

			logger.Info("Deleted existing metadata",
				zap.String("table", table.TableName),
				zap.Uint("table_id", existingTableID))
			return nil
		})

		if err != nil {
			return fmt.Errorf("delete existing metadata failed: %w", err)
		}
	}

	// 2. 获取表的字段信息
	columns, err := getColumns(ctx, db, dbName, table.TableName)
	if err != nil {
		return fmt.Errorf("get columns failed: %w", err)
	}

	// 3. 在事务中创建表和字段记录
	return db.Transaction(func(tx *gorm.DB) error {
		// 创建sys_table记录
		tableData := map[string]interface{}{
			"NAME":                 strings.ToUpper(table.TableName),
			"DISPLAY_NAME":         getDisplayName(table.TableName, table.TableComment),
			"DESCRIPTION":          table.TableComment,
			"MASK":                 "AMDQ", // 默认权限：增删改查提交打印授权反提交
			"SYS_TABLECATEGORY_ID": 1,      // 默认类别
			"IS_ACTIVE":            "Y",
			"CREATE_BY":            "system",
			"CREATE_TIME":          time.Now(),
			"SYS_COMPANY_ID":       1,
		}

		if err := tx.Table("sys_table").Create(&tableData).Error; err != nil {
			return fmt.Errorf("create sys_table record failed: %w", err)
		}

		// 获取最后插入的ID
		var tableID uint
		if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&tableID).Error; err != nil {
			return fmt.Errorf("failed to get table ID: %w", err)
		}

		if tableID == 0 {
			return fmt.Errorf("table ID is 0")
		}

		// 创建sys_column记录
		orderno := 10
		for _, col := range columns {
			columnData := createColumnData(tableID, strings.ToUpper(table.TableName), col, orderno)
			if err := tx.Table("sys_column").Create(&columnData).Error; err != nil {
				return fmt.Errorf("create sys_column record failed [%s]: %w", col.ColumnName, err)
			}
			orderno += 10
		}

		logger.Info("Created table metadata",
			zap.String("table", table.TableName),
			zap.Uint("table_id", tableID),
			zap.Int("columns", len(columns)),
		)

		return nil
	})
}

// createColumnData 创建字段数据
func createColumnData(tableID uint, tableName string, col ColumnInfo, orderno int) map[string]interface{} {
	dbName := strings.ToUpper(col.ColumnName)
	displayName := getDisplayName(col.ColumnName, col.ColumnComment)

	// 映射数据类型
	colType := mapDataType(col.DataType)
	colLength := getColumnLength(col)

	// 确定字段属性
	nullAble := "Y"
	if col.IsNullable == "NO" {
		nullAble = "N"
	}

	setValueType := determineSetValueType(col)
	modifiAble := determineModifiAble(col)
	displayType := determineDisplayType(col)

	data := map[string]interface{}{
		"SYS_TABLE_ID":   tableID,
		"DISPLAY_NAME":   displayName,
		"DB_NAME":        dbName,
		"FULL_NAME":      tableName + "." + dbName,
		"DESCRIPTION":    col.ColumnComment,
		"COL_TYPE":       colType,
		"ORDERNO":        orderno,
		"NULL_ABLE":      nullAble,
		"SET_VALUE_TYPE": setValueType,
		"MODIFI_ABLE":    modifiAble,
		"DISPLAY_TYPE":   displayType,
		"IS_ACTIVE":      "Y",
		"CREATE_BY":      "system",
		"CREATE_TIME":    time.Now(),
		"SYS_COMPANY_ID": 1,
	}

	if colLength > 0 {
		data["COL_LENGTH"] = colLength
	}

	if col.ColumnDefault != nil {
		data["DEFAULT_VALUE"] = *col.ColumnDefault
	}

	// 主键标记
	if col.ColumnKey == "PRI" {
		data["MASK"] = "0000000000"
		data["IS_AK"] = "Y"
		data["IS_DK"] = "Y"
	}

	return data
}

// mapDataType 映射MySQL数据类型到系统类型
func mapDataType(mysqlType string) string {
	switch strings.ToLower(mysqlType) {
	case "int", "bigint", "smallint", "tinyint", "mediumint":
		return "int"
	case "decimal", "numeric", "float", "double":
		return "decimal"
	case "date":
		return "date"
	case "datetime", "timestamp":
		return "datetime"
	case "time":
		return "time"
	case "char":
		return "char"
	case "text", "mediumtext", "longtext":
		return "text"
	default:
		return "varchar"
	}
}

// getColumnLength 获取字段长度
func getColumnLength(col ColumnInfo) int {
	if col.CharMaxLength != nil && *col.CharMaxLength > 0 {
		return *col.CharMaxLength
	}

	// 从COLUMN_TYPE中解析长度，如 varchar(100) -> 100
	colType := col.ColumnType
	if idx := strings.Index(colType, "("); idx > 0 {
		endIdx := strings.Index(colType, ")")
		if endIdx > idx {
			var length int
			fmt.Sscanf(colType[idx+1:endIdx], "%d", &length)
			return length
		}
	}

	return 0
}

// determineSetValueType 确定赋值类型
func determineSetValueType(col ColumnInfo) string {
	dbName := strings.ToUpper(col.ColumnName)

	// 主键
	if col.ColumnKey == "PRI" {
		return "pk"
	}

	// 标准字段
	switch dbName {
	case "CREATE_BY":
		return "createBy"
	case "UPDATE_BY":
		return "operator"
	case "CREATE_TIME", "UPDATE_TIME":
		return "sysdate"
	case "SYS_COMPANY_ID", "SYS_ORG_ID":
		return "object"
	case "IS_ACTIVE":
		return "select"
	}

	// 外键字段（以_ID结尾）
	if strings.HasSuffix(dbName, "_ID") && dbName != "ID" {
		return "fk"
	}

	// 默认为界面输入
	return "byPage"
}

// determineModifiAble 确定是否可修改
func determineModifiAble(col ColumnInfo) string {
	dbName := strings.ToUpper(col.ColumnName)

	// 主键和系统字段不可修改
	if col.ColumnKey == "PRI" {
		return "N"
	}

	switch dbName {
	case "CREATE_BY", "CREATE_TIME", "UPDATE_BY", "UPDATE_TIME", "SYS_COMPANY_ID":
		return "N"
	}

	return "Y"
}

// determineDisplayType 确定显示类型
func determineDisplayType(col ColumnInfo) string {
	dbName := strings.ToUpper(col.ColumnName)
	dataType := strings.ToLower(col.DataType)

	// 主键
	if col.ColumnKey == "PRI" {
		return "text"
	}

	// IS_ACTIVE 使用复选框
	if dbName == "IS_ACTIVE" {
		return "check"
	}

	// 日期时间类型
	switch dataType {
	case "date":
		return "date"
	case "datetime", "timestamp":
		return "datetime"
	case "time":
		return "time"
	}

	// 文本类型
	if dataType == "text" || dataType == "mediumtext" || dataType == "longtext" {
		return "textarea"
	}

	// 外键使用下拉选择
	if strings.HasSuffix(dbName, "_ID") && dbName != "ID" {
		return "select"
	}

	return "text"
}

// getDisplayName 获取显示名称
func getDisplayName(name, comment string) string {
	if comment != "" {
		return comment
	}
	return strings.ToUpper(name)
}

// executeInitSQL 执行 init.sql 脚本
func executeInitSQL(mysqlConfig *config.MySQLConfig) error {
	// 1. 读取 init.sql 文件
	sqlFilePath := filepath.Join("sqls", "init.sql")
	sqlBytes, err := os.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read init.sql: %w", err)
	}

	sqlContent := string(sqlBytes)
	if len(sqlContent) == 0 {
		return fmt.Errorf("init.sql is empty")
	}

	// 2. 构建 DSN（添加 multiStatements=true 参数）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
	)

	// 3. 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// 4. 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 5. 执行 SQL
	logger.Info("Executing SQL script", zap.String("file", sqlFilePath))
	_, err = db.Exec(sqlContent)
	if err != nil {
		return fmt.Errorf("failed to execute init.sql: %w", err)
	}

	logger.Info("SQL script executed successfully")
	return nil
}

// initDirectoriesFromTables 为 sys_table 中的表创建对应的 sys_directory
func initDirectoriesFromTables(ctx context.Context, db *gorm.DB) error {
	// 1. 查询所有 sys_table 记录（包括已存在的）
	var tables []struct {
		ID          uint
		Name        string
		DisplayName string
		URL         string
	}

	err := db.WithContext(ctx).
		Table("sys_table").
		Select("ID, NAME, DISPLAY_NAME, URL").
		Where("IS_ACTIVE = ?", "Y").
		Find(&tables).Error

	if err != nil {
		return fmt.Errorf("query sys_table failed: %w", err)
	}

	if len(tables) == 0 {
		logger.Info("No tables found in sys_table, skipping directory initialization")
		return nil
	}

	logger.Info("Found tables in sys_table", zap.Int("count", len(tables)))

	// 2. 为每个表创建或更新 sys_directory
	createdCount := 0
	updatedCount := 0
	skippedCount := 0

	for _, table := range tables {
		// 检查是否已存在对应的 sys_directory
		var existingDirID uint
		err := db.WithContext(ctx).
			Table("sys_directory").
			Select("ID").
			Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", table.ID, "Y").
			Scan(&existingDirID).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error("Failed to check existing directory",
				zap.String("table", table.Name),
				zap.Error(err))
			continue
		}

		if existingDirID > 0 {
			// 目录已存在，检查 sys_table 的 SYS_DIRECTORY_ID 是否已设置
			var tableDirectoryID *uint
			db.WithContext(ctx).
				Table("sys_table").
				Select("SYS_DIRECTORY_ID").
				Where("ID = ?", table.ID).
				Scan(&tableDirectoryID)

			if tableDirectoryID == nil || *tableDirectoryID == 0 {
				// 更新 sys_table 的 SYS_DIRECTORY_ID
				if err := db.WithContext(ctx).
					Table("sys_table").
					Where("ID = ?", table.ID).
					Update("SYS_DIRECTORY_ID", existingDirID).Error; err != nil {
					logger.Error("Failed to update table's directory ID",
						zap.String("table", table.Name),
						zap.Uint("dirID", existingDirID),
						zap.Error(err))
					continue
				}
				updatedCount++
				logger.Info("Updated table's directory link",
					zap.String("table", table.Name),
					zap.Uint("dirID", existingDirID))
			} else {
				skippedCount++
			}
			continue
		}

		// 3. 在事务中创建 sys_directory 并更新 sys_table
		err = db.Transaction(func(tx *gorm.DB) error {
			// 创建 sys_directory 记录
			directoryData := map[string]interface{}{
				"NAME":         table.Name,
				"DISPLAY_NAME": table.DisplayName,
				"URL":          table.URL,
				"SYS_TABLE_ID": table.ID,
				"IS_ACTIVE":    "Y",
				"CREATE_BY":    "system",
				"CREATE_TIME":  time.Now(),
				"SYS_COMPANY_ID": 1,
			}

			if err := tx.Table("sys_directory").Create(&directoryData).Error; err != nil {
				return fmt.Errorf("create sys_directory failed: %w", err)
			}

			// 获取新创建的目录ID
			var dirID uint
			if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&dirID).Error; err != nil {
				return fmt.Errorf("failed to get directory ID: %w", err)
			}

			if dirID == 0 {
				return fmt.Errorf("directory ID is 0")
			}

			// 更新 sys_table 的 SYS_DIRECTORY_ID
			if err := tx.Table("sys_table").
				Where("ID = ?", table.ID).
				Update("SYS_DIRECTORY_ID", dirID).Error; err != nil {
				return fmt.Errorf("update sys_table.SYS_DIRECTORY_ID failed: %w", err)
			}

			logger.Info("Created directory and linked to table",
				zap.String("table", table.Name),
				zap.Uint("tableID", table.ID),
				zap.Uint("dirID", dirID))

			return nil
		})

		if err != nil {
			logger.Error("Failed to create directory for table",
				zap.String("table", table.Name),
				zap.Error(err))
			continue
		}

		createdCount++
	}

	// 4. 输出统计结果
	logger.Info("Directory initialization completed",
		zap.Int("created", createdCount),
		zap.Int("updated", updatedCount),
		zap.Int("skipped", skippedCount),
		zap.Int("total", len(tables)))

	return nil
}

// initBaseDictionaries 初始化基础数据字典
func initBaseDictionaries(ctx context.Context, db *gorm.DB) error {
	// 检查YESNO字典是否已存在
	var count int64
	err := db.WithContext(ctx).Table("sys_dict").
		Where("NAME = ?", "YESNO").
		Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		logger.Info("Base dictionaries already exist, skipping")
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 创建YESNO字典
		dictData := map[string]interface{}{
			"NAME":           "YESNO",
			"DISPLAY_NAME":   "是否",
			"TYPE":           0, // String类型
			"DESCRIPTION":    "是/否选择",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      "system",
			"CREATE_TIME":    time.Now(),
			"SYS_COMPANY_ID": 1,
		}

		result := tx.Table("sys_dict").Create(&dictData)
		if result.Error != nil {
			return fmt.Errorf("create YESNO dict failed: %w", result.Error)
		}

		// 获取最后插入的ID
		var dictID uint
		if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&dictID).Error; err != nil {
			return fmt.Errorf("failed to get dict ID: %w", err)
		}

		if dictID == 0 {
			return fmt.Errorf("dict ID is 0")
		}

		// 创建字典项
		items := []map[string]interface{}{
			{
				"SYS_DICT_ID":    dictID,
				"DISPLAY_NAME":   "是",
				"VALUE":          "Y",
				"ORDERNO":        10,
				"IS_ACTIVE":      "Y",
				"CREATE_BY":      "system",
				"CREATE_TIME":    time.Now(),
				"SYS_COMPANY_ID": 1,
			},
			{
				"SYS_DICT_ID":    dictID,
				"DISPLAY_NAME":   "否",
				"VALUE":          "N",
				"ORDERNO":        20,
				"IS_ACTIVE":      "Y",
				"CREATE_BY":      "system",
				"CREATE_TIME":    time.Now(),
				"SYS_COMPANY_ID": 1,
			},
		}

		for _, item := range items {
			if err := tx.Table("sys_dict_item").Create(&item).Error; err != nil {
				return fmt.Errorf("create dict item failed: %w", err)
			}
		}

		logger.Info("Created YESNO dictionary", zap.Uint("dict_id", dictID))
		return nil
	})
}
