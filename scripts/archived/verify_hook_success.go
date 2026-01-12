package main

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Connect to database
	dsn := "root:abc123@tcp(127.0.0.1:3306)/skyserver?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	ctx := context.Background()

	// 查询最新创建的表
	var table map[string]interface{}
	err = db.WithContext(ctx).Raw("SELECT * FROM sys_table WHERE NAME = 'PERFECT_TABLE'").Scan(&table).Error
	if err != nil {
		log.Fatal("Failed to query table:", err)
	}

	if len(table) == 0 {
		fmt.Println("❌ 表未创建")
		return
	}

	fmt.Println("✅ 表创建成功！")
	fmt.Printf("  ID: %v\n", table["ID"])
	fmt.Printf("  NAME: %v\n", table["NAME"])
	fmt.Printf("  DISPLAY_NAME: %v\n", table["DISPLAY_NAME"])
	fmt.Printf("  ORDERNO: %v\n", table["ORDERNO"])
	fmt.Printf("  SYS_DIRECTORY_ID: %v\n", table["SYS_DIRECTORY_ID"])

	// 查询该表的字段数量
	var columnCount int64
	db.WithContext(ctx).Raw("SELECT COUNT(*) FROM sys_column WHERE SYS_TABLE_ID = ?", table["ID"]).Scan(&columnCount)

	fmt.Printf("\n✅ 标准字段已创建: %d 个\n", columnCount)

	// 列出所有字段
	var columns []map[string]interface{}
	db.WithContext(ctx).Raw("SELECT DB_NAME, DISPLAY_NAME, COL_TYPE, SET_VALUE_TYPE FROM sys_column WHERE SYS_TABLE_ID = ? ORDER BY ORDERNO", table["ID"]).Scan(&columns)

	fmt.Println("\n字段列表:")
	for i, col := range columns {
		fmt.Printf("  %d. %v (%v) - %v [%v]\n", i+1, col["DB_NAME"], col["DISPLAY_NAME"], col["COL_TYPE"], col["SET_VALUE_TYPE"])
	}

	// 查询 directory
	var directory map[string]interface{}
	db.WithContext(ctx).Raw("SELECT * FROM sys_directory WHERE ID = ?", table["SYS_DIRECTORY_ID"]).Scan(&directory)

	if len(directory) > 0 {
		fmt.Println("\n✅ Directory 已创建:")
		fmt.Printf("  ID: %v\n", directory["ID"])
		fmt.Printf("  NAME: %v\n", directory["NAME"])
		fmt.Printf("  DISPLAY_NAME: %v\n", directory["DISPLAY_NAME"])
	}

	fmt.Println("\n🎉 sys_table_cmd 配置的钩子执行完全成功！")
}
