package metadata

import (
	"context"

	"gorm.io/gorm"
)

// CheckSchema 比对元数据与当前数据库的物理表结构（MySQL），返回不一致项：
// 登记了字段但物理表不存在的表、物理表中不存在的启用字段（虚拟字段与按钮/分隔等纯界面字段除外）、
// 字段 MASK 不是 10 位 0/1 的配置错误。
// 前两类不一致会让通用查询直接报 Unknown column；MASK 格式错误会让字段不显示、不可编辑。
func CheckSchema(ctx context.Context, db *gorm.DB) ([]string, error) {
	var issues []string

	var missingTables []string
	// 没有字段的表单是纯页面菜单（如费用中心、模板管理），没有物理表，不检查
	if err := db.WithContext(ctx).Raw(`SELECT t.NAME FROM sys_table t
		WHERE t.IS_ACTIVE = 'Y'
		  AND EXISTS (SELECT 1 FROM sys_column c WHERE c.SYS_TABLE_ID = t.ID AND c.IS_ACTIVE = 'Y')
		  AND NOT EXISTS (
			SELECT 1 FROM information_schema.TABLES x
			WHERE x.TABLE_SCHEMA = DATABASE() AND LOWER(x.TABLE_NAME) = LOWER(t.NAME))`).
		Scan(&missingTables).Error; err != nil {
		return nil, err
	}
	for _, t := range missingTables {
		issues = append(issues, "物理表不存在: "+t)
	}

	var missingColumns []string
	if err := db.WithContext(ctx).Raw(`SELECT CONCAT(t.NAME, '.', c.DB_NAME) FROM sys_column c
		JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
		WHERE c.IS_ACTIVE = 'Y' AND t.IS_ACTIVE = 'Y'
		  AND IFNULL(c.IS_VIRTUAL, 'N') <> 'Y'
		  AND IFNULL(c.DISPLAY_TYPE, '') NOT IN ('button', 'blank', 'hr')
		  AND EXISTS (SELECT 1 FROM information_schema.TABLES x
			WHERE x.TABLE_SCHEMA = DATABASE() AND LOWER(x.TABLE_NAME) = LOWER(t.NAME))
		  AND NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS x
			WHERE x.TABLE_SCHEMA = DATABASE() AND LOWER(x.TABLE_NAME) = LOWER(t.NAME)
			  AND LOWER(x.COLUMN_NAME) = LOWER(c.DB_NAME))`).
		Scan(&missingColumns).Error; err != nil {
		return nil, err
	}
	for _, c := range missingColumns {
		issues = append(issues, "物理列不存在: "+c)
	}

	var badMasks []string
	if err := db.WithContext(ctx).Raw(`SELECT CONCAT(t.NAME, '.', c.DB_NAME, '=', c.MASK) FROM sys_column c
		JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
		WHERE c.IS_ACTIVE = 'Y' AND t.IS_ACTIVE = 'Y' AND IFNULL(c.MASK, '') <> ''
		  AND c.MASK NOT REGEXP '^[01]{10}$'`).
		Scan(&badMasks).Error; err != nil {
		return nil, err
	}
	for _, m := range badMasks {
		issues = append(issues, "字段 MASK 格式错误（应为 10 位 0/1）: "+m)
	}

	return issues, nil
}
