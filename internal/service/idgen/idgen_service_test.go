package idgen

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func setupIDGenTest(t *testing.T) (Service, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	// 创建测试表
	if err := db.Exec(`CREATE TABLE test_table (ID INTEGER PRIMARY KEY AUTOINCREMENT, NAME TEXT)`).Error; err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	// 插入一些数据让 MAX(ID) 不为 0
	db.Exec(`INSERT INTO test_table (NAME) VALUES ('a'), ('b'), ('c')`)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return NewService(db, client), db
}

func TestIDGenService_NextID_NewTable(t *testing.T) {
	svc, _ := setupIDGenTest(t)
	ctx := context.Background()

	// 首次获取 ID - 应基于 DB MAX(ID) 初始化
	id1, err := svc.GetNextID(ctx, "test_table")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id1 < 4 {
		t.Errorf("Expected ID >= 4 (after MAX=3), got %d", id1)
	}

	// 再次获取应该递增
	id2, err := svc.GetNextID(ctx, "test_table")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id2 <= id1 {
		t.Errorf("Expected id2 > id1, got id1=%d, id2=%d", id1, id2)
	}
}

func TestIDGenService_NextID_EmptyTable(t *testing.T) {
	svc, db := setupIDGenTest(t)
	db.Exec(`DELETE FROM test_table`)
	ctx := context.Background()

	// 空表时第一个 ID 应该是 1
	id1, err := svc.GetNextID(ctx, "test_table")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id1 != 1 {
		t.Errorf("Expected first ID=1, got %d", id1)
	}

	id2, err := svc.GetNextID(ctx, "test_table")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id2 != 2 {
		t.Errorf("Expected second ID=2, got %d", id2)
	}
}

func TestIDGenService_ResetCache(t *testing.T) {
	svc, _ := setupIDGenTest(t)
	ctx := context.Background()

	// 先获取一个 ID 来初始化缓存
	_, err := svc.GetNextID(ctx, "test_table")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 重置缓存
	err = svc.ResetCache(ctx, "test_table")
	if err != nil {
		t.Errorf("ResetCache failed: %v", err)
	}
}