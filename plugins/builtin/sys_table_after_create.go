package builtin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/plugins/core"
	"github.com/sky-xhsoft/sky-server/plugins/registry"
	"gorm.io/gorm"
)

// SysTableAfterCreatePlugin sys_table表创建后自动生成标准字段的插件
// 参考 Oracle 存储过程 AD_TABLE_AC
// 插件命名规则：表单名称_执行时机_动作
type SysTableAfterCreatePlugin struct{}

// 通过 init 函数自动注册插件
func init() {
	registry.Register(
		"sys_table_after_create",
		func() core.Plugin {
			return &SysTableAfterCreatePlugin{}
		},
		core.PluginMetadata{
			Name:        "sys_table_after_create",
			Description: "sys_table 表创建后自动生成标准字段和目录",
			Version:     "1.0.0",
			Author:      "Sky-Server",
			Enabled:     true,
			Priority:    10,
			HookPoint:   "sys_table.after.create",
		},
	)
}

// Name 返回插件名称
func (p *SysTableAfterCreatePlugin) Name() string {
	return "sys_table_after_create"
}

// Description 返回插件描述
func (p *SysTableAfterCreatePlugin) Description() string {
	return "sys_table 表创建后自动生成标准字段和目录"
}

// Version 返回插件版本
func (p *SysTableAfterCreatePlugin) Version() string {
	return "1.0.0"
}

// Execute 执行插件逻辑
// 参考 Oracle AD_TABLE_AC 存储过程实现以下功能：
// 1. 验证 MASK 字段格式
// 2. 自动生成 orderno（如果未设置）
// 3. 创建 directory（对于不以 ITEM/LINE 结尾且没有 parent_table_id 的表）
// 4. 创建标准字段到 sys_column 表
// 5. 设置表的 pk_column_id
func (p *SysTableAfterCreatePlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
	// 只在创建操作时执行
	if data.Action != "create" {
		return nil
	}

	// 获取新创建的表ID
	tableID := data.RecordID
	if tableID == 0 {
		return fmt.Errorf("表ID不能为0")
	}
	// 查询表的完整信息
	tableInfo := make(map[string]interface{})
	if err := db.WithContext(ctx).Table("sys_table").
		Where("ID = ?", tableID).
		Take(&tableInfo).Error; err != nil {
		return fmt.Errorf("查询表信息失败: %v", err)
	}

	// 1. 验证 MASK 字段（必须是 AMDSQPGU 的组合，长度8位）
	mask := getStringValue(tableInfo, "MASK")
	if mask != "" {
		validChars := "AMDSQPGUIE"
		for _, ch := range mask {
			if !strings.Contains(validChars, string(ch)) {
				return fmt.Errorf("MASK 必须由 AMDSQPGUIE 组成，当前为: %s", mask)
			}
		}
	}

	// 2. 自动生成 orderno（如果未设置）
	orderno := getIntValue(tableInfo, "ORDERNO")
	if orderno == 0 {
		categoryID := getUintValue(tableInfo, "SYS_TABLECATEGORY_ID")
		var maxOrderno int
		db.Table("sys_table").
			Where("SYS_TABLECATEGORY_ID = ?", categoryID).
			Select("COALESCE(MAX(ORDERNO), 0)").
			Scan(&maxOrderno)

		orderno = ((maxOrderno / 10) + 1) * 10 // 按10递增

		// 更新 orderno
		if err := db.Table("sys_table").
			Where("ID = ?", tableID).
			Update("ORDERNO", orderno).Error; err != nil {
			return fmt.Errorf("更新orderno失败: %v", err)
		}
	}

	// 3. 创建 directory（对于不以 ITEM/LINE 结尾且没有 parent_table_id 的表）
	tableName := strings.ToUpper(getStringValue(tableInfo, "NAME"))
	parentTableID := getUintValue(tableInfo, "PARENT_TABLE_ID")
	directoryID := getUintValue(tableInfo, "SYS_DIRECTORY_ID")

	// 判断是否需要创建 directory
	needsDirectory := !strings.HasSuffix(tableName, "ITEM") &&
		!strings.HasSuffix(tableName, "LINE") &&
		parentTableID == 0

	if needsDirectory && directoryID == 0 {
		// 创建新的 directory
		directory := map[string]interface{}{
			"NAME":                  tableName + "_LIST",
			"DISPLAY_NAME":          getStringValue(tableInfo, "DISPLAY_NAME"),
			"SYS_TABLE_CATEGORY_ID": getUintValue(tableInfo, "SYS_TABLECATEGORY_ID"),
			"SYS_TABLE_ID":          tableID,
			"IS_ACTIVE":             "Y",
			"CREATE_BY":             data.Data["CREATE_BY"],
			"CREATE_TIME":           time.Now(),
			"SYS_COMPANY_ID":        data.CompanyID,
		}

		if err := db.Table("sys_directory").Create(&directory).Error; err != nil {
			return fmt.Errorf("创建directory失败: %v", err)
		}

		// 获取新创建的 directory ID（GORM在使用map插入时不会自动填充ID）
		var newDirectoryID uint
		if err := db.Raw("SELECT LAST_INSERT_ID()").Scan(&newDirectoryID).Error; err == nil && newDirectoryID > 0 {
			// 更新 sys_table 的 directory ID
			db.Table("sys_table").
				Where("ID = ?", tableID).
				Update("SYS_DIRECTORY_ID", newDirectoryID)
			directoryID = newDirectoryID
		}
	}

	// 验证：如果既没有 directory 也没有 parent_table，报错
	if directoryID == 0 && parentTableID == 0 && needsDirectory {
		return fmt.Errorf("必须设置父表(PARENT_TABLE_ID)或安全目录(SYS_DIRECTORY_ID)")
	}

	// 4. 创建标准字段
	var pkColumnID uint
	columns := p.getStandardColumns(tableID, tableName, data)

	for _, column := range columns {
		if err := db.WithContext(ctx).Table("sys_column").Create(&column).Error; err != nil {
			return fmt.Errorf("创建标准字段失败 [%s]: %v", column["DB_NAME"], err)
		}

		// 记录 ID 字段的 column_id，用于设置表的 pk_column_id
		if column["DB_NAME"] == "ID" {
			pkColumnID = getUintValue(column, "ID")
		}
	}

	// 5. 设置表的 pk_column_id, ak_column_id, dk_column_id
	if pkColumnID > 0 {
		updates := map[string]interface{}{
			"PK_COLUMN_ID": pkColumnID,
			"AK_COLUMN_ID": pkColumnID,
			"DK_COLUMN_ID": pkColumnID,
		}
		if err := db.Table("sys_table").
			Where("ID = ?", tableID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("更新表的PK字段ID失败: %v", err)
		}
	}

	return nil
}

// getStandardColumns 获取标准字段定义
// 参考 Oracle 存储过程中的字段定义
func (p *SysTableAfterCreatePlugin) getStandardColumns(tableID uint, tableName string, data core.PluginData) []map[string]interface{} {
	now := time.Now()
	companyID := data.CompanyID
	createBy := data.Data["CREATE_BY"]

	return []map[string]interface{}{
		// 1. ID 主键字段
		{
			"DISPLAY_NAME":   tableName + ".ID",
			"DB_NAME":        "ID",
			"FULL_NAME":      tableName + ".ID",
			"DESCRIPTION":    "主键",
			"COL_TYPE":       "int",
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        1,
			"NULL_ABLE":      "N",
			"MASK":           "0000000000", // 10位MASK，参考存储过程
			"SET_VALUE_TYPE": "pk",
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "text",
			"IS_SHOW_TITLE":  "Y",
			"IS_AK":          "Y", // 是主键
			"IS_DK":          "Y", // 是显示键
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 2. SYS_COMPANY_ID 所属公司
		{
			"DISPLAY_NAME":   tableName + ".SYS_COMPANY_ID",
			"DB_NAME":        "SYS_COMPANY_ID",
			"FULL_NAME":      tableName + ".SYS_COMPANY_ID",
			"DESCRIPTION":    "所属公司",
			"COL_TYPE":       "int",
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        2,
			"NULL_ABLE":      "Y",
			"MASK":           "0000000000",
			"SET_VALUE_TYPE": "object", // 从上下文对象获取
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "text",
			"IS_SHOW_TITLE":  "Y",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 3. 日志字段（虚拟字段，用于触发器）
		{
			"DISPLAY_NAME":   tableName + ".(" + tableName + ".ID+100)",
			"DB_NAME":        "(" + tableName + ".ID+100)",
			"FULL_NAME":      tableName + ".(" + tableName + ".ID+100)",
			"DESCRIPTION":    "日志",
			"COL_TYPE":       "varchar",
			"COL_LENGTH":     30,
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        1000,
			"NULL_ABLE":      "Y",
			"MASK":           "0010011001",
			"SET_VALUE_TYPE": "trigger", // 触发器赋值
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "hr", // 水平线分隔符
			"IS_SHOW_TITLE":  "N",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 4. CREATE_BY 创建人
		{
			"DISPLAY_NAME":   tableName + ".CREATE_BY",
			"DB_NAME":        "CREATE_BY",
			"FULL_NAME":      tableName + ".CREATE_BY",
			"DESCRIPTION":    "创建人",
			"COL_TYPE":       "varchar",
			"COL_LENGTH":     80,
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        1001,
			"NULL_ABLE":      "Y",
			"MASK":           "0010001001",
			"SET_VALUE_TYPE": "createBy", // 操作人赋值
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "text",
			"IS_SHOW_TITLE":  "Y",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 5. UPDATE_BY 修改人
		{
			"DISPLAY_NAME":   tableName + ".UPDATE_BY",
			"DB_NAME":        "UPDATE_BY",
			"FULL_NAME":      tableName + ".UPDATE_BY",
			"DESCRIPTION":    "修改人",
			"COL_TYPE":       "varchar",
			"COL_LENGTH":     80,
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        1002,
			"NULL_ABLE":      "Y",
			"MASK":           "0010101101",
			"SET_VALUE_TYPE": "operator", // 操作人赋值
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "text",
			"IS_SHOW_TITLE":  "Y",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 6. CREATE_TIME 创建时间
		{
			"DISPLAY_NAME":   tableName + ".CREATE_TIME",
			"DB_NAME":        "CREATE_TIME",
			"FULL_NAME":      tableName + ".CREATE_TIME",
			"DESCRIPTION":    "创建时间",
			"COL_TYPE":       "datetime",
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        1003,
			"NULL_ABLE":      "Y",
			"MASK":           "0010001001",
			"SET_VALUE_TYPE": "sysdate", // 系统时间
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "datetime",
			"IS_SHOW_TITLE":  "Y",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 7. UPDATE_TIME 修改时间
		{
			"DISPLAY_NAME":   tableName + ".UPDATE_TIME",
			"DB_NAME":        "UPDATE_TIME",
			"FULL_NAME":      tableName + ".UPDATE_TIME",
			"DESCRIPTION":    "修改时间",
			"COL_TYPE":       "datetime",
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        1004,
			"NULL_ABLE":      "Y",
			"MASK":           "0010101101",
			"SET_VALUE_TYPE": "sysdate", // 系统时间
			"MODIFI_ABLE":    "N",
			"DISPLAY_TYPE":   "datetime",
			"IS_SHOW_TITLE":  "Y",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},

		// 8. IS_ACTIVE 是否有效
		{
			"DISPLAY_NAME":   tableName + ".IS_ACTIVE",
			"DB_NAME":        "IS_ACTIVE",
			"FULL_NAME":      tableName + ".IS_ACTIVE",
			"DESCRIPTION":    "可用",
			"COL_TYPE":       "char",
			"COL_LENGTH":     1,
			"SYS_TABLE_ID":   tableID,
			"ORDERNO":        10000,
			"NULL_ABLE":      "N",
			"MASK":           "0011101000",
			"SET_VALUE_TYPE": "select", // 下拉选择
			"DEFAULT_VALUE":  "Y",
			"MODIFI_ABLE":    "Y",
			"DISPLAY_TYPE":   "check", // 复选框
			"IS_SHOW_TITLE":  "Y",
			"IS_ACTIVE":      "Y",
			"CREATE_BY":      createBy,
			"CREATE_TIME":    now,
			"SYS_COMPANY_ID": companyID,
		},
	}
}
