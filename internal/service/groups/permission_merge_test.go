package groups

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetUserDirectoryPermissionMergesGroups(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		`CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, IS_ADMIN TEXT, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_user_groups (ID INTEGER PRIMARY KEY, SYS_USER_ID INTEGER, SYS_DIRECTORY_ID INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_group_prem (ID INTEGER PRIMARY KEY, SYS_GROUPS_ID INTEGER, SYS_DIRECTORY_ID INTEGER, PERMISSION INTEGER, IS_ACTIVE TEXT)`,
		`INSERT INTO sys_user VALUES (1, 'N', 'Y'), (2, 'N', 'Y')`,
		// 用户 1 属于组 10（只读）和组 20（增改），用户 2 只属于组 10
		`INSERT INTO sys_user_groups VALUES (1, 1, 10, 'Y'), (2, 1, 20, 'Y'), (3, 2, 10, 'Y')`,
		`INSERT INTO sys_group_prem VALUES (1, 10, 5, 1, 'Y'), (2, 20, 5, 6, 'Y'), (3, 20, 6, 8, 'N')`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &service{db: db}
	ctx := context.Background()

	cases := []struct {
		user, dir uint
		want      int
	}{
		{1, 5, PermRead | PermCreate | PermUpdate}, // 两个组的权限合并
		{2, 5, PermRead},
		{1, 6, 0}, // 停用的授权不生效
		{2, 99, 0},
	}
	for _, c := range cases {
		got, err := s.GetUserDirectoryPermission(ctx, c.user, c.dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("user %d dir %d: got %d, want %d", c.user, c.dir, got, c.want)
		}
	}
}
