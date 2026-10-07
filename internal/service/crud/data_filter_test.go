package crud

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"gorm.io/gorm"
)

func TestUserTableDataFilterUsesTableDirectory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		`CREATE TABLE sys_directory (ID INTEGER PRIMARY KEY, SYS_TABLE_ID INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_user_groups (ID INTEGER PRIMARY KEY, SYS_USER_ID INTEGER, SYS_DIRECTORY_ID INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_group_prem (ID INTEGER PRIMARY KEY, SYS_GROUPS_ID INTEGER, SYS_DIRECTORY_ID INTEGER, PERMISSION INTEGER, FILTER_OBJ TEXT, IS_ACTIVE TEXT)`,
		// 表 81 对应目录 7；目录 81 属于别的表
		`INSERT INTO sys_directory VALUES (7, 81, 'Y'), (81, 3, 'Y')`,
		`INSERT INTO sys_user_groups VALUES (1, 1, 10, 'Y')`,
		`INSERT INTO sys_group_prem VALUES (1, 10, 7, 1, '{"EXAM_TYPE":"final"}', 'Y'), (2, 10, 81, 1, '{"WRONG":"x"}', 'Y')`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &service{db: db, groupsService: groups.NewService(db)}
	ctx := context.Background()

	f, err := s.userTableDataFilter(ctx, 1, 81)
	if err != nil {
		t.Fatal(err)
	}
	if f["EXAM_TYPE"] != "final" || f["WRONG"] != nil {
		t.Errorf("应取表 81 所在目录 7 的过滤条件, got %v", f)
	}

	f, err = s.userTableDataFilter(ctx, 1, 999)
	if err != nil || f != nil {
		t.Errorf("没有安全目录的表不过滤, got %v %v", f, err)
	}
}
