// encrypt-secrets 将元数据中 SET_VALUE_TYPE=encrypt 字段的存量明文加密（幂等，已加密的值会跳过）
//
// 用法：go run ./cmd/encrypt-secrets [-dry-run]
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/sqlutils"
	"github.com/sky-xhsoft/sky-server/internal/repository/mysql"
)

var dryRun = flag.Bool("dry-run", false, "只统计待加密的记录数，不写库")

// entityEncryptedColumns 实体上使用 serializer:encrypted、但不一定登记在元数据中的列
var entityEncryptedColumns = []secretColumn{
	{TableName: "sys_company_conf", DbName: "ALIYUN_OSS_ACCESS_KEY_SECRET"},
	{TableName: "sys_company_conf", DbName: "TENCENT_COS_SECRET_KEY"},
	{TableName: "sys_company_conf", DbName: "TENCENT_CLOUD_SECRET_KEY"},
	{TableName: "sys_company_conf", DbName: "TENCENT_CLOUD_CALLBACK_KEY"},
	{TableName: "sys_company_conf", DbName: "SECRET_KEY"},
	{TableName: "sys_company_conf", DbName: "VOLCENGINE_ACCESS_KEY_SECRET"},
	{TableName: "cloud_storage_config", DbName: "ALIYUN_OSS_ACCESS_KEY_SECRET"},
	{TableName: "cloud_storage_config", DbName: "TENCENT_COS_SECRET_KEY"},
	{TableName: "sys_email_config", DbName: "SMTP_PASSWORD"},
	{TableName: "cloud_share", DbName: "PASSWORD"},
}

type secretColumn struct {
	TableName string `gorm:"column:TABLE_NAME"`
	DbName    string `gorm:"column:DB_NAME"`
}

type secretRow struct {
	ID    uint   `gorm:"column:ID"`
	Value string `gorm:"column:VAL"`
}

func main() {
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	log, err := logger.Init(&cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 与服务端 setupFieldCrypto 保持一致的密钥选取规则
	key := cfg.Security.FieldEncryptKey
	if key == "" {
		fmt.Println("警告: security.fieldEncryptKey 未配置，回退使用 jwt.secret")
		key = cfg.JWT.Secret
	}
	fieldcrypto.Setup(key)

	db, err := mysql.Init(&cfg.Database.MySQL, log)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer mysql.CloseDB(db)

	var columns []secretColumn
	if err := db.Raw(`SELECT t.NAME AS TABLE_NAME, c.DB_NAME
		FROM sys_column c JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
		WHERE c.SET_VALUE_TYPE = 'encrypt' AND c.IS_ACTIVE = 'Y'`).Scan(&columns).Error; err != nil {
		fmt.Printf("查询加密字段失败: %v\n", err)
		os.Exit(1)
	}

	// 合并实体加密列并去重（表名大小写不敏感）
	seen := make(map[string]bool)
	var targets []secretColumn
	for _, col := range append(columns, entityEncryptedColumns...) {
		k := strings.ToLower(col.TableName + "." + col.DbName)
		if !seen[k] {
			seen[k] = true
			targets = append(targets, col)
		}
	}

	total := 0
	for _, col := range targets {
		table, _ := sqlutils.SanitizeIdentifier(col.TableName)
		column, _ := sqlutils.SanitizeIdentifier(col.DbName)
		if table == "" || column == "" {
			fmt.Printf("跳过非法标识符: %s.%s\n", col.TableName, col.DbName)
			continue
		}

		var rows []secretRow
		query := fmt.Sprintf("SELECT ID, %s AS VAL FROM %s WHERE %s IS NOT NULL AND %s <> '' AND %s NOT LIKE ?",
			column, table, column, column, column)
		if err := db.Raw(query, fieldcrypto.EncPrefix+"%").Scan(&rows).Error; err != nil {
			fmt.Printf("查询 %s.%s 失败: %v\n", table, column, err)
			os.Exit(1)
		}

		for _, row := range rows {
			if *dryRun {
				continue
			}
			encrypted, err := fieldcrypto.EncryptField(row.Value)
			if err != nil {
				fmt.Printf("加密 %s.%s ID=%d 失败: %v\n", table, column, row.ID, err)
				os.Exit(1)
			}
			update := fmt.Sprintf("UPDATE %s SET %s = ? WHERE ID = ?", table, column)
			if err := db.Exec(update, encrypted, row.ID).Error; err != nil {
				fmt.Printf("更新 %s.%s ID=%d 失败: %v\n", table, column, row.ID, err)
				os.Exit(1)
			}
		}
		fmt.Printf("%s.%s: %d 条明文%s\n", table, column, len(rows), map[bool]string{true: "待加密", false: "已加密"}[*dryRun])
		total += len(rows)
	}
	fmt.Printf("完成，共 %d 条\n", total)
}
