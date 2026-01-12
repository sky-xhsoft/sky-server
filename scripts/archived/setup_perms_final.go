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

	// Step 1: Create admin group
	fmt.Println("Step 1: Creating admin group...")
	err = db.WithContext(ctx).Exec(`
		INSERT INTO sys_groups (NAME, DESCRIPTION, SGRADE, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
		VALUES ('管理员组', '系统管理员权限组', 99, 'Y', '1', NOW(), '1', NOW(), 1)
		ON DUPLICATE KEY UPDATE ID=LAST_INSERT_ID(ID)
	`).Error
	if err != nil {
		log.Fatal("Failed to create group:", err)
	}

	var groupID uint
	db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&groupID)
	fmt.Printf("Group ID: %d\n", groupID)

	// Step 2: Create directories for all tables
	fmt.Println("Step 2: Creating directories for all tables...")
	err = db.WithContext(ctx).Exec(`
		INSERT INTO sys_directory (NAME, DISPLAY_NAME, SYS_TABLE_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
		SELECT
			CONCAT('DIR_', t.NAME),
			t.DISPLAY_NAME,
			t.ID,
			'Y',
			'1',
			NOW(),
			'1',
			NOW(),
			1
		FROM sys_table t
		WHERE t.IS_ACTIVE = 'Y'
		ON DUPLICATE KEY UPDATE SYS_TABLE_ID=VALUES(SYS_TABLE_ID)
	`).Error
	if err != nil {
		log.Fatal("Failed to create directories:", err)
	}

	// Step 3: Link tables to directories
	fmt.Println("Step 3: Linking tables to directories...")
	err = db.WithContext(ctx).Exec(`
		UPDATE sys_table t
		JOIN sys_directory d ON d.SYS_TABLE_ID = t.ID
		SET t.SYS_DIRECTORY_ID = d.ID
		WHERE t.IS_ACTIVE = 'Y'
	`).Error
	if err != nil {
		log.Fatal("Failed to link tables:", err)
	}

	// Step 4: Create permissions
	fmt.Println("Step 4: Creating permissions for all directories...")
	err = db.WithContext(ctx).Exec(`
		INSERT INTO sys_group_prem (SYS_GROUPS_ID, SYS_DIRECTORY_ID, PERMISSION, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
		SELECT
			?,
			d.ID,
			31,
			'Y',
			'1',
			NOW(),
			'1',
			NOW(),
			1
		FROM sys_directory d
		WHERE d.IS_ACTIVE = 'Y'
		AND d.SYS_TABLE_ID IS NOT NULL
		ON DUPLICATE KEY UPDATE PERMISSION=31
	`, groupID).Error
	if err != nil {
		log.Fatal("Failed to create permissions:", err)
	}

	// Step 5: Assign user to group
	fmt.Println("Step 5: Assigning admin user to group...")
	err = db.WithContext(ctx).Exec(`
		INSERT INTO sys_user_groups (SYS_USER_ID, SYS_DIRECTORY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
		VALUES (1, ?, 'Y', '1', NOW(), '1', NOW(), 1)
		ON DUPLICATE KEY UPDATE IS_ACTIVE='Y'
	`, groupID).Error
	if err != nil {
		log.Fatal("Failed to assign user to group:", err)
	}

	fmt.Println("\n✅ Permission setup completed successfully!")
	fmt.Println("Admin user (ID=1) has been granted full permissions (31) to all tables.")
	fmt.Println("\nNext step: Refresh the metadata cache using the API.")
}
