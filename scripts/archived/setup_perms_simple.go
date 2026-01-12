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

	// Execute SQL directly
	sql := `
-- 1. Create admin group
INSERT INTO sys_groups (NAME, DESCRIPTION, SGRADE, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
VALUES ('管理员组', '系统管理员权限组', 99, 'Y', '1', NOW(), '1', NOW(), 1)
ON DUPLICATE KEY UPDATE ID=LAST_INSERT_ID(ID);

SET @group_id = LAST_INSERT_ID();

-- 2. For each table in sys_table, create directory and permissions
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
ON DUPLICATE KEY UPDATE SYS_TABLE_ID=VALUES(SYS_TABLE_ID);

-- 3. Update sys_table to link directories
UPDATE sys_table t
JOIN sys_directory d ON d.SYS_TABLE_ID = t.ID
SET t.SYS_DIRECTORY_ID = d.ID
WHERE t.IS_ACTIVE = 'Y';

-- 4. Create permissions for all directories
INSERT INTO sys_group_prem (SYS_GROUPS_ID, SYS_DIRECTORY_ID, PREM, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
SELECT
    @group_id,
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
ON DUPLICATE KEY UPDATE PREM=31;

-- 5. Assign admin user to the group
INSERT INTO sys_user_groups (SYS_USER_ID, SYS_DIRECTORY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, SYS_COMPANY_ID)
VALUES (1, @group_id, 'Y', '1', NOW(), '1', NOW(), 1)
ON DUPLICATE KEY UPDATE IS_ACTIVE='Y';
`

	if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
		log.Fatal("Failed to execute SQL:", err)
	}

	fmt.Println("✅ Permission setup completed successfully!")
	fmt.Println("Admin user (ID=1) has been granted full permissions (31) to all tables.")
	fmt.Println("\nPlease refresh the metadata cache on the server using:")
	fmt.Println("  curl -X POST http://localhost:9090/api/v1/metadata/refresh -H \"Authorization: Bearer YOUR_TOKEN\"")
}
