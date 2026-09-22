package dbmonitor

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	return db
}

func TestNewMonitor_DefaultInterval(t *testing.T) {
	db := setupTestDB(t)
	m := NewMonitor(db, 0)
	if m.interval != 30*time.Second {
		t.Errorf("Expected default 30s, got %v", m.interval)
	}
}

func TestNewMonitor_CustomInterval(t *testing.T) {
	db := setupTestDB(t)
	m := NewMonitor(db, 5*time.Second)
	if m.interval != 5*time.Second {
		t.Errorf("Expected 5s, got %v", m.interval)
	}
}

func TestMonitor_GetStats(t *testing.T) {
	db := setupTestDB(t)
	m := NewMonitor(db, time.Second)

	stats, err := m.GetStats()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
	if stats.CollectedAt.IsZero() {
		t.Error("CollectedAt should be set")
	}
}

func TestMonitor_GetLastStats_NilBeforeCollection(t *testing.T) {
	db := setupTestDB(t)
	m := NewMonitor(db, time.Second)
	if m.GetLastStats() != nil {
		t.Error("Last stats should be nil before any collection")
	}
}

func TestMonitor_GetLastStats_AfterCollection(t *testing.T) {
	db := setupTestDB(t)
	m := NewMonitor(db, time.Second)
	m.GetStats() // 触发一次采集
	if m.GetLastStats() == nil {
		t.Error("Last stats should not be nil after collection")
	}
}

func TestMonitor_StartStop(t *testing.T) {
	db := setupTestDB(t)
	m := NewMonitor(db, 100*time.Millisecond)

	m.Start()
	// 第二次启动应该被忽略（不死锁）
	m.Start()

	time.Sleep(150 * time.Millisecond)

	m.Stop()
	// 第二次停止应该被忽略
	m.Stop()
}
