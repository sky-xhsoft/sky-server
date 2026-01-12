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
	fmt.Printf("sys_table ID: %d\n\n", tableID)

	// Get all hooks for sys_table
	var hooks []map[string]interface{}
	err = db.WithContext(ctx).Table("sys_table_cmd").
		Where("SYS_TABLE_ID = ? AND IS_ACTIVE = 'Y'", tableID).
		Find(&hooks).Error
	if err != nil {
		log.Fatal("Failed to query hooks:", err)
	}

	fmt.Printf("Found %d hooks for sys_table:\n", len(hooks))
	for i, hook := range hooks {
		fmt.Printf("\n%d. Hook:\n", i+1)
		for k, v := range hook {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
}
