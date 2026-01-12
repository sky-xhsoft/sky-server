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

	// Step 1: Add ORDERNO column
	fmt.Println("添加 ORDERNO 字段到 sys_table 表...")
	err = db.WithContext(ctx).Exec(`
		ALTER TABLE sys_table ADD COLUMN ORDERNO INT NULL COMMENT '序号' AFTER SYS_TABLECATEGORY_ID
	`).Error
	if err != nil {
		// 如果字段已存在，忽略错误
		if err.Error() == "Error 1060 (42S21): Duplicate column name 'ORDERNO'" ||
		   contains(err.Error(), "Duplicate column name") {
			fmt.Println("ORDERNO 字段已存在，跳过...")
		} else {
			log.Fatal("Failed to add ORDERNO column:", err)
		}
	} else {
		fmt.Println("✓ ORDERNO 字段添加成功")
	}

	// Step 2: Set default values for existing records
	fmt.Println("\n为现有记录设置 ORDERNO 默认值...")
	err = db.WithContext(ctx).Exec(`
		UPDATE sys_table SET ORDERNO = ID * 10 WHERE ORDERNO IS NULL
	`).Error
	if err != nil {
		log.Fatal("Failed to set default ORDERNO values:", err)
	}
	fmt.Println("✓ ORDERNO 默认值设置成功")

	fmt.Println("\n✅ 完成！sys_table 表已添加 ORDERNO 字段")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
