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

	// Update the hook to use correct plugin name
	err = db.WithContext(ctx).Exec(`
		UPDATE sys_table_cmd
		SET CONTENT = 'sys_table'
		WHERE CONTENT = 'sys_table_after_create'
	`).Error
	if err != nil {
		log.Fatal("Failed to update hook:", err)
	}

	fmt.Println("✅ 钩子配置已更新，插件名称从 'sys_table_after_create' 改为 'sys_table'")
	fmt.Println("\n请刷新元数据缓存：")
	fmt.Println("  curl -X POST http://localhost:9090/api/v1/metadata/refresh -H \"Authorization: Bearer YOUR_TOKEN\"")
}
