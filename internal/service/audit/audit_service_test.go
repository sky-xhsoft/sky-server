package audit

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func setupAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&entity.AuditLog{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	return db
}

func TestLogBuilder_Build(t *testing.T) {
	log := NewLogBuilder().
		WithUser(123, "alice").
		WithAction("create").
		WithResource("user", "456", "Test User").
		WithRequest("POST", "/api/users", "127.0.0.1", "test-agent").
		WithStatus(entity.StatusSuccess).
		WithDuration(150).
		WithCompanyID(1).
		Build()

	if log.UserID != 123 {
		t.Errorf("UserID mismatch: %v", log.UserID)
	}
	if log.Username != "alice" {
		t.Errorf("Username mismatch: %s", log.Username)
	}
	if log.Action != "create" {
		t.Errorf("Action mismatch: %s", log.Action)
	}
	if log.Resource != "user" {
		t.Errorf("Resource mismatch: %s", log.Resource)
	}
	if log.ResourceID != "456" {
		t.Errorf("ResourceID mismatch: %s", log.ResourceID)
	}
	if log.Method != "POST" {
		t.Errorf("Method mismatch: %s", log.Method)
	}
	if log.Status != entity.StatusSuccess {
		t.Errorf("Status mismatch: %s", log.Status)
	}
	if log.Duration != 150 {
		t.Errorf("Duration mismatch: %d", log.Duration)
	}
}

func TestLogBuilder_WithErrorSetsFailureStatus(t *testing.T) {
	log := NewLogBuilder().
		WithAction("test").
		WithError(context.DeadlineExceeded).
		Build()

	if log.Status != entity.StatusFailure {
		t.Errorf("Expected status=failure when error set, got %s", log.Status)
	}
	if log.ErrorMessage == "" {
		t.Error("ErrorMessage should be set")
	}
}

func TestAuditService_Log_Sync(t *testing.T) {
	db := setupAuditTestDB(t)
	svc := NewService(db)

	log := NewLogBuilder().
		WithAction("login").
		WithResource("auth", "0", "Login attempt").
		WithStatus(entity.StatusSuccess).
		Build()

	err := svc.Log(context.Background(), log)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证已写入
	var count int64
	db.Model(&entity.AuditLog{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 log, got %d", count)
	}
}

func TestAuditService_GetLog_NotFound(t *testing.T) {
	db := setupAuditTestDB(t)
	svc := NewService(db)

	_, err := svc.GetLog(context.Background(), 99999)
	if err == nil {
		t.Error("Expected error for non-existent log")
	}
}

func TestAuditService_QueryLogs_Empty(t *testing.T) {
	db := setupAuditTestDB(t)
	svc := NewService(db)

	logs, total, err := svc.QueryLogs(context.Background(), &QueryRequest{
		Page:     1,
		PageSize: 10,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected 0 total, got %d", total)
	}
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(logs))
	}
}

func TestAuditService_QueryLogs_WithFilters(t *testing.T) {
	db := setupAuditTestDB(t)
	svc := NewService(db)

	// 插入测试数据
	uid := uint(100)
	for i := 0; i < 5; i++ {
		log := NewLogBuilder().
			WithUser(uid, "user1").
			WithAction("create").
			WithResource("table", "1", "Test").
			Build()
		svc.Log(context.Background(), log)
	}

	logs, total, err := svc.QueryLogs(context.Background(), &QueryRequest{
		UserID:   uid,
		Page:     1,
		PageSize: 10,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected 5 total, got %d", total)
	}
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(logs))
	}
}

func TestAuditService_CleanExpiredLogs(t *testing.T) {
	db := setupAuditTestDB(t)
	svc := NewService(db)

	// 插入一些数据
	log := NewLogBuilder().WithAction("test").Build()
	svc.Log(context.Background(), log)

	// 清理一个未来的日期，不应删除任何东西
	deleted, err := svc.CleanExpiredLogs(context.Background(), time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted (logs are newer), got %d", deleted)
	}

	// 清理未来的时间，应该删除所有
	deleted, err = svc.CleanExpiredLogs(context.Background(), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if deleted < 1 {
		t.Errorf("Expected at least 1 deleted, got %d", deleted)
	}
}
