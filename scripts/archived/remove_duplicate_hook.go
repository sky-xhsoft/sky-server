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

	// Delete the Go hook for sys_table (since we have PluginManager handling it)
	err = db.WithContext(ctx).Exec(`
		DELETE FROM sys_table_cmd
		WHERE CONTENT_TYPE = 'go'
		AND CONTENT = 'sys_table'
	`).Error
	if err != nil {
		log.Fatal("Failed to delete hook:", err)
	}

	fmt.Println("✅ 已删除sys_table_cmd中的重复Go钩子")
	fmt.Println("\n插件现在由PluginManager管理，无需通过sys_table_cmd配置")
	fmt.Println("\n请刷新元数据缓存：")
	fmt.Println("  curl -X POST http://localhost:9090/api/v1/metadata/refresh -H \"Authorization: Bearer YOUR_TOKEN\"")
}
