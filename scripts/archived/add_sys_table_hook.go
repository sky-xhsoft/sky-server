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

	// Get sys_table ID
	var tableID uint
	err = db.WithContext(ctx).Raw("SELECT ID FROM sys_table WHERE NAME = 'SYS_TABLE'").Scan(&tableID).Error
	if err != nil {
		log.Fatal("Failed to get sys_table ID:", err)
	}
	fmt.Printf("sys_table ID: %d\n", tableID)

	// Check if hook already exists
	var count int64
	err = db.WithContext(ctx).Table("sys_table_cmd").
		Where("SYS_TABLE_ID = ? AND ACTION = 'A' AND EVENT = 'end' AND CONTENT_TYPE = 'go' AND CONTENT = 'sys_table'", tableID).
		Count(&count).Error
	if err != nil {
		log.Fatal("Failed to check hook existence:", err)
	}

	if count > 0 {
		fmt.Println("✅ 钩子已存在，无需重复添加")
		return
	}

	// Insert the hook
	err = db.WithContext(ctx).Exec(`
		INSERT INTO sys_table_cmd (
			SYS_TABLE_ID, ACTION, EVENT, CONTENT_TYPE, CONTENT,
			IS_ACTIVE, ORDERNO,
			CREATE_BY, CREATE_TIME, SYS_COMPANY_ID
		) VALUES (
			?, 'A', 'end', 'go', 'sys_table',
			'Y', 10,
			'system', NOW(), 1
		)
	`, tableID).Error
	if err != nil {
		log.Fatal("Failed to insert hook:", err)
	}

	fmt.Println("✅ 已添加 sys_table 的 after create 钩子到 sys_table_cmd")
	fmt.Println("\n钩子配置：")
	fmt.Println("  表: sys_table")
	fmt.Println("  动作: A (新增)")
	fmt.Println("  事件: end (结束)")
	fmt.Println("  内容类型: go")
	fmt.Println("  内容: sys_table")
	fmt.Println("\n请重启服务器以加载新的钩子注册：")
	fmt.Println("  1. 停止当前服务器 (Ctrl+C)")
	fmt.Println("  2. 重新运行: go run cmd/server/main.go")
	fmt.Println("  3. 刷新元数据缓存: curl -X POST http://localhost:9090/api/v1/metadata/refresh")
}
