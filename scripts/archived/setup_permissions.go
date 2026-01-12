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

	// 1. Get all tables from sys_table
	type Table struct {
		ID          uint
		Name        string
		DisplayName string
	}
	var tables []Table
	if err := db.WithContext(ctx).Table("sys_table").
		Select("ID, NAME, DISPLAY_NAME").
		Where("IS_ACTIVE = ?", "Y").
		Find(&tables).Error; err != nil {
		log.Fatal("Failed to query tables:", err)
	}

	fmt.Printf("Found %d tables\n", len(tables))

	// 2. Create or get admin group
	var groupID uint
	err = db.WithContext(ctx).Raw(`
		INSERT INTO sys_groups (NAME, CODE, DESCRIPTION, SGRADE, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
		VALUES ('管理员组', 'ADMIN_GROUP', '系统管理员权限组', 99, 'Y', 1, NOW(), 1, NOW(), 1)
		ON DUPLICATE KEY UPDATE ID=LAST_INSERT_ID(ID)
	`).Error
	if err != nil {
		log.Fatal("Failed to create admin group:", err)
	}

	err = db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&groupID).Error
	if err != nil {
		log.Fatal("Failed to get group ID:", err)
	}
	fmt.Printf("Admin group ID: %d\n", groupID)

	// 3. For each table, create directory, assign permission, link table
	for _, table := range tables {
		fmt.Printf("Processing table: %s (%d)\n", table.Name, table.ID)

		// Create directory
		var dirID uint
		err = db.WithContext(ctx).Raw(`
			INSERT INTO sys_directory (NAME, CODE, SYS_TABLE_ID, PARENT_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
			VALUES (?, ?, ?, NULL, 'Y', 1, NOW(), 1, NOW(), 1)
			ON DUPLICATE KEY UPDATE ID=LAST_INSERT_ID(ID), SYS_TABLE_ID=?
		`, table.DisplayName, table.Name, table.ID, table.ID).Error
		if err != nil {
			log.Printf("Failed to create directory for %s: %v\n", table.Name, err)
			continue
		}

		err = db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&dirID).Error
		if err != nil {
			log.Printf("Failed to get directory ID for %s: %v\n", table.Name, err)
			continue
		}

		// Update table to link directory
		err = db.WithContext(ctx).Exec(`
			UPDATE sys_table SET SYS_DIRECTORY_ID = ? WHERE ID = ?
		`, dirID, table.ID).Error
		if err != nil {
			log.Printf("Failed to update table %s: %v\n", table.Name, err)
			continue
		}

		// Assign full permissions (31 = all permissions) to admin group
		err = db.WithContext(ctx).Raw(`
			INSERT INTO sys_group_prem (SYS_GROUPS_ID, SYS_DIRECTORY_ID, PREM, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
			VALUES (?, ?, 31, 'Y', 1, NOW(), 1, NOW(), 1)
			ON DUPLICATE KEY UPDATE PREM=31
		`, groupID, dirID).Error
		if err != nil {
			log.Printf("Failed to assign permissions for %s: %v\n", table.Name, err)
			continue
		}

		fmt.Printf("  ✓ Created directory %d, assigned permissions\n", dirID)
	}

	// 4. Assign admin user (ID=1) to admin group
	err = db.WithContext(ctx).Raw(`
		INSERT INTO sys_user_groups (SYS_USER_ID, SYS_DIRECTORY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
		VALUES (1, ?, 'Y', 1, NOW(), 1, NOW(), 1)
		ON DUPLICATE KEY UPDATE IS_ACTIVE='Y'
	`, groupID).Error
	if err != nil {
		log.Fatal("Failed to assign user to group:", err)
	}

	fmt.Println("\n✅ Permission setup completed successfully!")
	fmt.Println("Admin user (ID=1) has been granted full permissions to all tables.")
}
