package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupOverviewDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, ddl := range []string{
		`CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_groups (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_table (ID INTEGER PRIMARY KEY, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_company (ID INTEGER PRIMARY KEY, IS_ACTIVE TEXT)`,
		`CREATE TABLE audit_log (ID INTEGER PRIMARY KEY, USER_ID INTEGER, USERNAME TEXT, ACTION TEXT, PATH TEXT, CREATED_AT DATETIME)`,
		`CREATE TABLE cloud_item (ID INTEGER PRIMARY KEY, ITEM_TYPE TEXT, NAME TEXT, OWNER_ID INTEGER, FILE_TYPE TEXT, FILE_EXT TEXT, FILE_SIZE INTEGER, CREATE_TIME DATETIME, IS_ACTIVE TEXT)`,
		`CREATE TABLE cloud_share (ID INTEGER PRIMARY KEY, SHARER_ID INTEGER, RESOURCE_ID INTEGER, SHARE_CODE TEXT, STATUS TEXT, CREATE_TIME DATETIME, IS_ACTIVE TEXT)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestGetSystemOverview(t *testing.T) {
	db := setupOverviewDB(t)
	now := time.Now()
	db.Exec(`INSERT INTO sys_user VALUES (1, 10, 'Y'), (2, 10, 'Y'), (3, 10, 'N'), (4, 20, 'Y')`)
	db.Exec(`INSERT INTO sys_groups VALUES (1, 10, 'Y'), (2, 0, 'Y'), (3, 20, 'Y')`)
	db.Exec(`INSERT INTO sys_table VALUES (1, 'Y'), (2, 'Y'), (3, 'N')`)
	db.Exec(`INSERT INTO sys_company VALUES (10, 'Y'), (20, 'Y')`)
	db.Exec(`INSERT INTO audit_log VALUES
		(1, 1, 'alice', 'login', '/api/v1/auth/login', ?),
		(2, 1, 'alice', 'update', '/api/v1/data/SYS_USER/1', ?),
		(3, 2, 'bob', 'create', '/api/v1/data/STUDENT', ?),
		(4, 4, 'other', 'delete', '/api/v1/data/SECRET', ?),
		(5, 1, 'alice', 'update', '/api/v1/old', ?)`,
		now, now, now.AddDate(0, 0, -1), now, now.AddDate(0, 0, -30))

	svc := &service{db: db}
	ov, err := svc.GetSystemOverview(context.Background(), 10, 7)
	if err != nil {
		t.Fatal(err)
	}
	if ov.Users != 2 || ov.Groups != 2 || ov.Tables != 2 || ov.Companies != 2 {
		t.Errorf("统计不符: users=%d groups=%d tables=%d companies=%d", ov.Users, ov.Groups, ov.Tables, ov.Companies)
	}
	if len(ov.Trend.Dates) != 7 || ov.Trend.Logins[6] != 1 || ov.Trend.Ops[6] != 1 || ov.Trend.Ops[5] != 1 {
		t.Errorf("趋势不符: %+v", ov.Trend)
	}
	for _, op := range ov.RecentOps {
		if op.Username == "other" {
			t.Error("不应包含其他公司的操作")
		}
		if op.Action == "login" {
			t.Error("最近操作不应包含登录")
		}
	}
	if len(ov.RecentOps) != 3 || ov.RecentOps[0].Target == "" {
		t.Errorf("最近操作不符: %d 条", len(ov.RecentOps))
	}
}

func TestGetCloudOverview(t *testing.T) {
	db := setupOverviewDB(t)
	now := time.Now()
	db.Exec(`INSERT INTO cloud_item VALUES
		(1, 'folder', '素材', 7, NULL, NULL, 0, ?, 'Y'),
		(2, 'file', 'a.mp4', 7, 'video', 'mp4', 1000, ?, 'Y'),
		(3, 'file', 'b.pdf', 7, 'document', 'pdf', 500, ?, 'Y'),
		(4, 'file', 'deleted.txt', 7, 'text', 'txt', 9999, ?, 'N'),
		(5, 'file', 'others.doc', 8, 'document', 'doc', 7777, ?, 'Y')`,
		now.Add(-3*time.Hour), now.Add(-2*time.Hour), now.AddDate(0, 0, -1), now, now)
	db.Exec(`INSERT INTO cloud_share VALUES (1, 7, 2, 'abc', 'active', ?, 'Y'), (2, 7, 3, 'def', 'expired', ?, 'Y'), (3, 8, 5, 'ghi', 'active', ?, 'Y')`,
		now.Add(-time.Hour), now.AddDate(0, 0, -2), now)

	svc := &service{db: db}
	ov, err := svc.GetCloudOverview(context.Background(), 7, 7)
	if err != nil {
		t.Fatal(err)
	}
	if ov.UsedStorage != 1500 || ov.FileCount != 2 || ov.FolderCount != 1 || ov.ShareCount != 1 {
		t.Errorf("统计不符: storage=%d files=%d folders=%d shares=%d", ov.UsedStorage, ov.FileCount, ov.FolderCount, ov.ShareCount)
	}
	if ov.Trend.Uploads[6] != 1 || ov.Trend.Files[6] != 2 || ov.Trend.Uploads[5] != 1 {
		t.Errorf("趋势不符: %+v", ov.Trend)
	}
	for _, f := range ov.RecentFiles {
		if f.Name == "others.doc" || f.Name == "deleted.txt" {
			t.Errorf("最近文件不应包含 %s", f.Name)
		}
	}
	if len(ov.RecentFiles) != 2 || ov.RecentFiles[0].Name != "a.mp4" {
		t.Errorf("最近文件不符: %+v", ov.RecentFiles)
	}
	if len(ov.Activities) != 4 || ov.Activities[0].Action != "createdShare" || ov.Activities[0].Target != "a.mp4" {
		t.Errorf("最近动态不符: %+v", ov.Activities[0])
	}
}
