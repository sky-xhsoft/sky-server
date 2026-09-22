package dict

import (
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	pkgErrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// mockDictRepository mock 字典仓储
type mockDictRepository struct {
	items    []*entity.SysDictItem
	itemsErr error
	dict     *entity.SysDict
	dictErr  error
	allDicts []*entity.SysDict
}

func (m *mockDictRepository) GetAllDicts() ([]*entity.SysDict, error) {
	return m.allDicts, nil
}

func (m *mockDictRepository) GetDictByName(name string) (*entity.SysDict, error) {
	return m.dict, m.dictErr
}

func (m *mockDictRepository) GetDictByID(id uint) (*entity.SysDict, error) {
	return m.dict, m.dictErr
}

func (m *mockDictRepository) GetDictItems(dictID uint) ([]*entity.SysDictItem, error) {
	return m.items, m.itemsErr
}

func (m *mockDictRepository) GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error) {
	return m.items, m.itemsErr
}

// setupDictService 创建测试用字典服务（包含 miniredis 客户端）
func setupDictService(t *testing.T, repo *mockDictRepository) Service {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return NewService(repo, client, 60)
}

func TestDictService_GetDictItems_RepoError(t *testing.T) {
	repo := &mockDictRepository{itemsErr: errors.New("db error")}
	svc := setupDictService(t, repo)

	_, err := svc.GetDictItems(1)
	if err == nil {
		t.Fatal("Expected error")
	}
	if pkgErrors.GetCode(err) != pkgErrors.ErrDatabase {
		t.Errorf("Expected ErrDatabase, got %d", pkgErrors.GetCode(err))
	}
}

func TestDictService_GetDictItems_Empty(t *testing.T) {
	repo := &mockDictRepository{items: []*entity.SysDictItem{}}
	svc := setupDictService(t, repo)

	items, err := svc.GetDictItems(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

func TestDictService_GetDictItems_Success(t *testing.T) {
	repo := &mockDictRepository{
		items: []*entity.SysDictItem{
			{Value: "v1", DisplayName: "Value 1"},
			{Value: "v2", DisplayName: "Value 2"},
		},
	}
	svc := setupDictService(t, repo)

	items, err := svc.GetDictItems(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestDictService_GetDictItemsByName_Success(t *testing.T) {
	repo := &mockDictRepository{
		items: []*entity.SysDictItem{
			{Value: "active", DisplayName: "活跃"},
		},
	}
	svc := setupDictService(t, repo)

	items, err := svc.GetDictItemsByName("status")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}
}

func TestDictService_GetDefaultValue_NoDefault(t *testing.T) {
	repo := &mockDictRepository{
		items: []*entity.SysDictItem{
			{Value: "v1", DisplayName: "V1", IsDefaultValue: "N"},
			{Value: "v2", DisplayName: "V2", IsDefaultValue: "N"},
		},
	}
	svc := setupDictService(t, repo)

	val, err := svc.GetDefaultValue("test_dict")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// 无显式默认值时返回首项
	if val != "v1" {
		t.Errorf("Expected v1 (first), got %s", val)
	}
}

func TestDictService_GetDefaultValue_WithDefault(t *testing.T) {
	repo := &mockDictRepository{
		items: []*entity.SysDictItem{
			{Value: "v1", DisplayName: "V1", IsDefaultValue: "N"},
			{Value: "v2", DisplayName: "V2", IsDefaultValue: "Y"},
		},
	}
	svc := setupDictService(t, repo)

	val, err := svc.GetDefaultValue("test_dict")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != "v2" {
		t.Errorf("Expected v2 (default), got %s", val)
	}
}

func TestDictService_GetDefaultValue_EmptyItems(t *testing.T) {
	repo := &mockDictRepository{items: []*entity.SysDictItem{}}
	svc := setupDictService(t, repo)

	val, err := svc.GetDefaultValue("empty_dict")
	if err != nil {
		t.Errorf("Unexpected error for empty dict: %v", err)
	}
	if val != "" {
		t.Errorf("Expected empty string for empty dict, got %s", val)
	}
}
