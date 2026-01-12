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

	// Tables that need ORDERNO column
	tables := []string{"sys_table_cmd", "sys_table_ref"}

	for _, table := range tables {
		fmt.Printf("检查表 %s...\n", table)

		// Check if ORDERNO column exists
		var count int64
		err = db.WithContext(ctx).Raw(`
			SELECT COUNT(*)
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = 'skyserver'
			AND TABLE_NAME = ?
			AND COLUMN_NAME = 'ORDERNO'
		`, table).Scan(&count).Error
		if err != nil {
			log.Printf("Failed to check %s: %v\n", table, err)
			continue
		}

		if count > 0 {
			fmt.Printf("  %s 已有 ORDERNO 字段\n", table)
			continue
		}

		// Add ORDERNO column
		fmt.Printf("  添加 ORDERNO 字段到 %s...\n", table)
		err = db.WithContext(ctx).Exec(fmt.Sprintf(`
			ALTER TABLE %s ADD COLUMN ORDERNO INT NULL COMMENT '序号'
		`, table)).Error
		if err != nil {
			log.Printf("Failed to add ORDERNO to %s: %v\n", table, err)
			continue
		}

		// Set default values
		err = db.WithContext(ctx).Exec(fmt.Sprintf(`
			UPDATE %s SET ORDERNO = ID * 10 WHERE ORDERNO IS NULL
		`, table)).Error
		if err != nil {
			log.Printf("Failed to set default ORDERNO for %s: %v\n", table, err)
			continue
		}

		fmt.Printf("  ✓ %s ORDERNO 字段添加成功\n", table)
	}

	fmt.Println("\n✅ 完成！所有表已添加 ORDERNO 字段")
	fmt.Println("\n请刷新元数据缓存：")
	fmt.Println("  curl -X POST http://localhost:9090/api/v1/metadata/refresh -H \"Authorization: Bearer YOUR_TOKEN\"")
}
