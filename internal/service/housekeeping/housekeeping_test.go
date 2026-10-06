package housekeeping

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, ddl := range []string{
		`CREATE TABLE sys_user_session (ID INTEGER PRIMARY KEY, EXPIRE_TIME DATETIME)`,
		`CREATE TABLE audit_log (ID INTEGER PRIMARY KEY, CREATED_AT DATETIME)`,
		`CREATE TABLE cloud_share (ID INTEGER PRIMARY KEY, STATUS TEXT, EXPIRE_TIME DATETIME, RESOURCE_ID INTEGER)`,
		`CREATE TABLE cloud_item (ID INTEGER PRIMARY KEY, IS_ACTIVE TEXT)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func count(t *testing.T, db *gorm.DB, sql string) int64 {
	t.Helper()
	var n int64
	if err := db.Raw(sql).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRunOnce(t *testing.T) {
	db := setupDB(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)

	db.Exec(`INSERT INTO sys_user_session (ID, EXPIRE_TIME) VALUES (1, ?), (2, ?), (3, ?)`, past, past, future)
	db.Exec(`INSERT INTO audit_log (ID, CREATED_AT) VALUES (1, ?), (2, ?), (3, ?)`,
		now.AddDate(0, 0, -181), now.AddDate(0, 0, -179), now)
	db.Exec(`INSERT INTO cloud_item (ID, IS_ACTIVE) VALUES (100, 'Y'), (200, 'N')`)
	db.Exec(`INSERT INTO cloud_share (ID, STATUS, EXPIRE_TIME, RESOURCE_ID) VALUES (1, 'active', ?, 100), (2, 'active', ?, 100), (3, 'active', NULL, 100), (4, 'disabled', ?, 100), (5, 'active', NULL, 200)`,
		past, future, past)

	res, err := RunOnce(context.Background(), db, Config{AuditRetentionDays: 180}, now)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExpiredSessions != 2 || res.OldAuditLogs != 1 || res.ExpiredShares != 1 || res.DisabledShares != 1 {
		t.Fatalf("结果不符: %+v", res)
	}
	if n := count(t, db, `SELECT COUNT(*) FROM sys_user_session`); n != 1 {
		t.Errorf("应保留 1 个未过期会话, got %d", n)
	}
	if n := count(t, db, `SELECT COUNT(*) FROM audit_log`); n != 2 {
		t.Errorf("应保留 180 天内的 2 条审计日志, got %d", n)
	}
	if n := count(t, db, `SELECT COUNT(*) FROM cloud_share WHERE STATUS = 'expired'`); n != 1 {
		t.Errorf("应只有 1 个分享被标记为过期, got %d", n)
	}
	if n := count(t, db, `SELECT COUNT(*) FROM cloud_share WHERE ID = 5 AND STATUS = 'disabled'`); n != 1 {
		t.Error("指向已删除资源的分享应被标记为 disabled")
	}
	if n := count(t, db, `SELECT COUNT(*) FROM cloud_share WHERE ID = 4 AND STATUS = 'disabled'`); n != 1 {
		t.Error("已禁用的分享状态不应被改动")
	}
}

func TestRunOnce_RetentionDisabled(t *testing.T) {
	db := setupDB(t)
	now := time.Now()
	db.Exec(`INSERT INTO audit_log (ID, CREATED_AT) VALUES (1, ?)`, now.AddDate(-2, 0, 0))

	if _, err := RunOnce(context.Background(), db, Config{AuditRetentionDays: 0}, now); err != nil {
		t.Fatal(err)
	}
	if n := count(t, db, `SELECT COUNT(*) FROM audit_log`); n != 1 {
		t.Errorf("保留天数为 0 时不应删除审计日志, got %d", n)
	}
}
