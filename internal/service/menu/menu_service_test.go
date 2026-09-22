package menu

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func setupMenuTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	// 创建相关表
	createStmts := []string{
		`CREATE TABLE sys_subsystem (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_COMPANY_ID INTEGER DEFAULT 0,
			NAME TEXT, ENAME TEXT, ORDERNO INTEGER DEFAULT 0,
			URL TEXT, ICON TEXT, DESCRIPTION TEXT, KEY TEXT,
			CREATE_TIME DATETIME, UPDATE_TIME DATETIME,
			IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
		`CREATE TABLE sys_table_category (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_COMPANY_ID INTEGER DEFAULT 0,
			SYS_SUBSYSTEM_ID INTEGER,
			NAME TEXT, ENAME TEXT, ORDERNO INTEGER DEFAULT 0,
			URL TEXT, ICON TEXT, DESCRIPTION TEXT,
			CREATE_TIME DATETIME, UPDATE_TIME DATETIME,
			IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
		`CREATE TABLE sys_table (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_COMPANY_ID INTEGER DEFAULT 0,
			NAME TEXT, DISPLAY_NAME TEXT, ENAME TEXT,
			SYS_TABLECATEGORY_ID INTEGER,
			ORDERNO INTEGER DEFAULT 0,
			URL TEXT, ICO_IMG TEXT, IS_MENU TEXT DEFAULT 'Y',
			MASK TEXT,
			CREATE_TIME DATETIME, UPDATE_TIME DATETIME,
			IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
		`CREATE TABLE sys_user (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_COMPANY_ID INTEGER, USERNAME TEXT,
			IS_ADMIN TEXT DEFAULT 'N', IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
		`CREATE TABLE sys_user_groups (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_USER_ID INTEGER,
			SYS_GROUPS_ID INTEGER,
			SYS_DIRECTORY_ID INTEGER,
			IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
		`CREATE TABLE sys_group_prem (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_GROUPS_ID INTEGER, SYS_DIRECTORY_ID INTEGER,
			PERMISSION INTEGER DEFAULT 0,
			IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
		`CREATE TABLE sys_directory (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_TABLE_ID INTEGER, PARENT_ID INTEGER,
			NAME TEXT, IS_ACTIVE TEXT DEFAULT 'Y'
		)`,
	}

	for _, stmt := range createStmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("Failed to create table: %v\nSQL: %s", err, stmt)
		}
	}
	return db
}

func TestMenuService_GetMenuTree_EmptyDB(t *testing.T) {
	db := setupMenuTestDB(t)
	svc := NewService(db)

	tree, err := svc.GetMenuTree(context.Background(), 1, "live")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// 空 DB 时返回空切片或 nil 都可接受
	if len(tree) != 0 {
		t.Errorf("Expected empty tree, got %d nodes", len(tree))
	}
}

func TestMenuService_GetMenuTree_WithData(t *testing.T) {
	db := setupMenuTestDB(t)

	// 插入子系统
	db.Exec(`INSERT INTO sys_subsystem (NAME, KEY, ORDERNO, SYS_COMPANY_ID, IS_ACTIVE) VALUES ('Live System', 'live', 10, 1, 'Y')`)
	// 插入分类
	db.Exec(`INSERT INTO sys_table_category (NAME, SYS_SUBSYSTEM_ID, ORDERNO, SYS_COMPANY_ID, IS_ACTIVE) VALUES ('Live Cat', 1, 10, 1, 'Y')`)
	// 插入表
	db.Exec(`INSERT INTO sys_table (NAME, DISPLAY_NAME, SYS_TABLECATEGORY_ID, ORDERNO, IS_MENU, SYS_COMPANY_ID, IS_ACTIVE) VALUES ('test_table', 'Test Table', 1, 10, 'Y', 1, 'Y')`)

	svc := NewService(db)
	tree, err := svc.GetMenuTree(context.Background(), 1, "live")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tree) == 0 {
		t.Error("Expected at least 1 subsystem")
	}
}

func TestMenuService_GetUserMenuTree_AdminReturnsAll(t *testing.T) {
	db := setupMenuTestDB(t)

	// 管理员用户
	db.Exec(`INSERT INTO sys_user (USERNAME, SYS_COMPANY_ID, IS_ADMIN, IS_ACTIVE) VALUES ('admin', 1, 'Y', 'Y')`)
	db.Exec(`INSERT INTO sys_subsystem (NAME, KEY, ORDERNO, SYS_COMPANY_ID, IS_ACTIVE) VALUES ('System', 'live', 10, 1, 'Y')`)

	svc := NewService(db)
	_, err := svc.GetUserMenuTree(context.Background(), 1, 1, "live")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestMenuService_GetUserMenuTree_NonAdminNoGroups(t *testing.T) {
	db := setupMenuTestDB(t)

	// 非管理员用户，无权限组
	db.Exec(`INSERT INTO sys_user (USERNAME, SYS_COMPANY_ID, IS_ADMIN, IS_ACTIVE) VALUES ('user1', 1, 'N', 'Y')`)

	svc := NewService(db)
	tree, err := svc.GetUserMenuTree(context.Background(), 1, 1, "live")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tree) != 0 {
		t.Errorf("Expected empty tree for user without permissions, got %d", len(tree))
	}
}

func TestBuildMenuTree_PreservesOrder(t *testing.T) {
	db := setupMenuTestDB(t)
	s := &service{db: db}

	categories := []entity.SysTableCategory{
		{BaseModel: entity.BaseModel{ID: 1}, SysSubsystemID: 1, Name: "Cat1", OrderNo: 10},
		{BaseModel: entity.BaseModel{ID: 2}, SysSubsystemID: 2, Name: "Cat2", OrderNo: 20},
	}
	cat1ID := uint(1)
	cat2ID := uint(2)
	tables := []entity.SysTable{
		{BaseModel: entity.BaseModel{ID: 1}, Name: "tbl1", SysTableCategoryID: &cat1ID, IsMenu: "Y", OrderNo: 10},
		{BaseModel: entity.BaseModel{ID: 2}, Name: "tbl2", SysTableCategoryID: &cat2ID, IsMenu: "Y", OrderNo: 10},
	}

	// 允许的子系统集合（含 1、2）
	subsystemIDs := map[uint]bool{1: true, 2: true}
	tree := s.buildMenuTree(subsystemIDs, categories, tables)
	if len(tree) != 2 {
		t.Fatalf("Expected 2 root nodes, got %d", len(tree))
	}
	// 顶层现在是表类别，按 OrderNo 升序：Cat1 (10) 在前
	if tree[0].Name != "Cat1" {
		t.Errorf("Expected Cat1 in front, got %s", tree[0].Name)
	}
	// 验证子节点
	if len(tree[0].Children) != 1 || tree[0].Children[0].Name != "tbl1" {
		t.Errorf("Expected Cat1 to have tbl1 child")
	}
	// 验证过滤：子系统 3 不在集合内，其类别应被过滤
	filtered := s.buildMenuTree(map[uint]bool{1: true}, categories, tables)
	if len(filtered) != 1 || filtered[0].Name != "Cat1" {
		t.Errorf("Expected only Cat1 after filtering to subsystem 1, got %d nodes", len(filtered))
	}
}

func TestBuildMenuTree_FiltersEmptyBranches(t *testing.T) {
	db := setupMenuTestDB(t)
	s := &service{db: db}

	// 分类有但没有表 - 应被过滤
	categories := []entity.SysTableCategory{
		{BaseModel: entity.BaseModel{ID: 1}, SysSubsystemID: 1, Name: "Empty Cat"},
	}
	tables := []entity.SysTable{} // 无表

	tree := s.buildMenuTree(map[uint]bool{1: true}, categories, tables)
	if len(tree) != 0 {
		t.Errorf("Empty branches should be filtered, got %d nodes", len(tree))
	}
}
