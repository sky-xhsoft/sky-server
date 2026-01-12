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

	fmt.Println("开始执行迁移：Rename sys_column.NAME to DISPLAY_NAME")
	fmt.Println("=" + string(make([]byte, 60)))

	// Execute migration
	sql := `ALTER TABLE sys_column
		CHANGE COLUMN ` + "`NAME`" + ` ` + "`DISPLAY_NAME`" + ` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称'`

	if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("✅ Migration completed successfully")

	// Verify the change
	var count int64
	if err := db.WithContext(ctx).Raw("SELECT COUNT(*) FROM sys_column").Scan(&count).Error; err != nil {
		log.Fatal("Failed to verify:", err)
	}
	fmt.Printf("✅ Total columns: %d\n", count)

	// Show sample data
	type ColumnSample struct {
		ID          uint
		DisplayName string `gorm:"column:DISPLAY_NAME"`
		DbName      string `gorm:"column:DB_NAME"`
		ColType     string `gorm:"column:COL_TYPE"`
	}

	var samples []ColumnSample
	err = db.WithContext(ctx).Raw(`
		SELECT ID, DISPLAY_NAME, DB_NAME, COL_TYPE
		FROM sys_column
		WHERE SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'SYS_TABLE' LIMIT 1)
		LIMIT 5
	`).Scan(&samples).Error

	if err == nil && len(samples) > 0 {
		fmt.Println("\n示例数据 (SYS_TABLE 表的字段):")
		for _, s := range samples {
			fmt.Printf("  - %s (%s): %s\n", s.DisplayName, s.DbName, s.ColType)
		}
	}

	fmt.Println("\n迁移完成！")
}
