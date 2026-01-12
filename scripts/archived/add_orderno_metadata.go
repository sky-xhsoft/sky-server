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

	// Get sys_table's ID
	var tableID uint
	err = db.WithContext(ctx).Table("sys_table").
		Select("ID").
		Where("NAME = ?", "SYS_TABLE").
		Scan(&tableID).Error
	if err != nil {
		log.Fatal("Failed to find sys_table:", err)
	}
	fmt.Printf("sys_table ID: %d\n", tableID)

	// Check if ORDERNO column metadata exists
	var count int64
	err = db.WithContext(ctx).Table("sys_column").
		Where("SYS_TABLE_ID = ? AND DB_NAME = ?", tableID, "ORDERNO").
		Count(&count).Error
	if err != nil {
		log.Fatal("Failed to check ORDERNO column:", err)
	}

	if count > 0 {
		fmt.Println("ORDERNO 字段元数据已存在，跳过...")
		return
	}

	// Add ORDERNO column metadata
	fmt.Println("添加 ORDERNO 字段元数据到 sys_column...")
	err = db.WithContext(ctx).Exec(`
		INSERT INTO sys_column (
			SYS_TABLE_ID, DB_NAME, NAME,
			COL_TYPE, COL_LENGTH, NULL_ABLE,
			IS_ACTIVE, DISPLAY_TYPE, SET_VALUE_TYPE, ORDERNO,
			CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID
		) VALUES (
			?, 'ORDERNO', '序号',
			'int', 11, 'Y',
			'Y', 'number', 'byPage', 140,
			'1', NOW(), '1', NOW(), 1
		)
	`, tableID).Error
	if err != nil {
		log.Fatal("Failed to add ORDERNO column metadata:", err)
	}

	fmt.Println("✅ ORDERNO 字段元数据添加成功")
	fmt.Println("\n请刷新元数据缓存：")
	fmt.Println("  curl -X POST http://localhost:9090/api/v1/metadata/refresh -H \"Authorization: Bearer YOUR_TOKEN\"")
}
