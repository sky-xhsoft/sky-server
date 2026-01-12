package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Directory struct {
	ID          uint
	Name        string
	Code        string
	SysTableID  uint
	ParentID    *uint
	IsActive    string
	CreateBy    string
	CreateTime  time.Time
	UpdateBy    string
	UpdateTime  time.Time
	SysCompanyID uint
}

type Group struct {
	ID           uint
	Name         string
	Description  string
	Sgrade       int
	IsActive     string
	CreateBy     string
	CreateTime   time.Time
	UpdateBy     string
	UpdateTime   time.Time
	SysCompanyID uint
}

type GroupPerm struct {
	ID             uint
	SysGroupsID    uint
	SysDirectoryID uint
	Prem           int
	IsActive       string
	CreateBy       string
	CreateTime     time.Time
	UpdateBy       string
	UpdateTime     time.Time
	SysCompanyID   uint
}

type UserGroup struct {
	ID             uint
	SysUserID      uint
	SysDirectoryID uint
	IsActive       string
	CreateBy       string
	CreateTime     time.Time
	UpdateBy       string
	UpdateTime     time.Time
	SysCompanyID   uint
}

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

	// 2. Create admin group
	var group Group
	err = db.WithContext(ctx).Table("sys_groups").
		Where("NAME = ?", "管理员组").
		First(&group).Error

	if err == gorm.ErrRecordNotFound {
		// Create new group
		now := time.Now()
		group = Group{
			Name:         "管理员组",
			Description:  "系统管理员权限组",
			Sgrade:       99,
			IsActive:     "Y",
			CreateBy:     "1",
			CreateTime:   now,
			UpdateBy:     "1",
			UpdateTime:   now,
			SysCompanyID: 1,
		}
		if err := db.WithContext(ctx).Table("sys_groups").Create(&group).Error; err != nil {
			log.Fatal("Failed to create admin group:", err)
		}
		fmt.Printf("Created admin group ID: %d\n", group.ID)
	} else if err != nil {
		log.Fatal("Failed to query group:", err)
	} else {
		fmt.Printf("Found existing admin group ID: %d\n", group.ID)
	}

	// 3. For each table, create directory, assign permission, link table
	for _, table := range tables {
		fmt.Printf("Processing table: %s (ID:%d)\n", table.Name, table.ID)

		// Check if directory exists
		var dir Directory
		err = db.WithContext(ctx).Table("sys_directory").
			Where("CODE = ?", table.Name).
			First(&dir).Error

		if err == gorm.ErrRecordNotFound {
			// Create new directory
			now := time.Now()
			dir = Directory{
				Name:         table.DisplayName,
				Code:         table.Name,
				SysTableID:   table.ID,
				ParentID:     nil,
				IsActive:     "Y",
				CreateBy:     "1",
				CreateTime:   now,
				UpdateBy:     "1",
				UpdateTime:   now,
				SysCompanyID: 1,
			}
			if err := db.WithContext(ctx).Table("sys_directory").Create(&dir).Error; err != nil {
				log.Printf("Failed to create directory for %s: %v\n", table.Name, err)
				continue
			}
			fmt.Printf("  Created directory ID: %d\n", dir.ID)
		} else if err != nil {
			log.Printf("Failed to query directory for %s: %v\n", table.Name, err)
			continue
		} else {
			// Update existing directory to link to table
			dir.SysTableID = table.ID
			if err := db.WithContext(ctx).Table("sys_directory").Where("ID = ?", dir.ID).Updates(map[string]interface{}{
				"SYS_TABLE_ID": table.ID,
			}).Error; err != nil {
				log.Printf("Failed to update directory for %s: %v\n", table.Name, err)
			}
			fmt.Printf("  Found existing directory ID: %d\n", dir.ID)
		}

		// Update table to link directory
		err = db.WithContext(ctx).Table("sys_table").
			Where("ID = ?", table.ID).
			Updates(map[string]interface{}{
				"SYS_DIRECTORY_ID": dir.ID,
			}).Error
		if err != nil {
			log.Printf("Failed to update table %s: %v\n", table.Name, err)
			continue
		}

		// Check if permission already exists
		var existingPerm GroupPerm
		err = db.WithContext(ctx).Table("sys_group_prem").
			Where("SYS_GROUPS_ID = ? AND SYS_DIRECTORY_ID = ?", group.ID, dir.ID).
			First(&existingPerm).Error

		if err == gorm.ErrRecordNotFound {
			// Create new permission (31 = all permissions: Read|Write|Submit|Audit|Export)
			now := time.Now()
			perm := GroupPerm{
				SysGroupsID:    group.ID,
				SysDirectoryID: dir.ID,
				Prem:           31,
				IsActive:       "Y",
				CreateBy:       "1",
				CreateTime:     now,
				UpdateBy:       "1",
				UpdateTime:     now,
				SysCompanyID:   1,
			}
			if err := db.WithContext(ctx).Table("sys_group_prem").Create(&perm).Error; err != nil {
				log.Printf("Failed to assign permissions for %s: %v\n", table.Name, err)
				continue
			}
			fmt.Printf("  Assigned permissions (31)\n")
		} else if err != nil {
			log.Printf("Failed to query permissions for %s: %v\n", table.Name, err)
			continue
		} else {
			// Update existing permission to ensure it's 31
			if err := db.WithContext(ctx).Table("sys_group_prem").Where("ID = ?", existingPerm.ID).Updates(map[string]interface{}{
				"PREM": 31,
			}).Error; err != nil {
				log.Printf("Failed to update permissions for %s: %v\n", table.Name, err)
			}
			fmt.Printf("  Updated existing permissions (31)\n")
		}
	}

	// 4. Assign admin user (ID=1) to admin group
	var existingUserGroup UserGroup
	err = db.WithContext(ctx).Table("sys_user_groups").
		Where("SYS_USER_ID = ? AND SYS_DIRECTORY_ID = ?", 1, group.ID).
		First(&existingUserGroup).Error

	if err == gorm.ErrRecordNotFound {
		now := time.Now()
		userGroup := UserGroup{
			SysUserID:      1,
			SysDirectoryID: group.ID,
			IsActive:       "Y",
			CreateBy:       "1",
			CreateTime:     now,
			UpdateBy:       "1",
			UpdateTime:     now,
			SysCompanyID:   1,
		}
		if err := db.WithContext(ctx).Table("sys_user_groups").Create(&userGroup).Error; err != nil {
			log.Fatal("Failed to assign user to group:", err)
		}
		fmt.Println("\n✅ Assigned admin user to group")
	} else if err != nil {
		log.Fatal("Failed to query user groups:", err)
	} else {
		// Update existing to ensure it's active
		if err := db.WithContext(ctx).Table("sys_user_groups").Where("ID = ?", existingUserGroup.ID).Updates(map[string]interface{}{
			"IS_ACTIVE": "Y",
		}).Error; err != nil {
			log.Fatal("Failed to update user group:", err)
		}
		fmt.Println("\n✅ User group already exists and is active")
	}

	fmt.Println("\n✅ Permission setup completed successfully!")
	fmt.Println("Admin user (ID=1) has been granted full permissions (31) to all tables.")
	fmt.Println("\nPlease refresh the metadata cache on the server.")
}
