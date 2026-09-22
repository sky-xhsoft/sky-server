package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// mockMetadataService mock metadata service
type mockMetadataService struct {
	table        *entity.SysTable
	tableErr     error
	columns      []*entity.SysColumn
	columnsErr   error
	refs         []*entity.SysTableRef
	refsErr      error
	actions      []*entity.SysAction
	actionsErr   error
	refreshErr   error
	version      string
	fkOptions    []map[string]interface{}
	fkOptionsErr error
	fkDisplay    string
	fkDisplayErr error
}

func (m *mockMetadataService) GetTable(tableName string) (*entity.SysTable, error) {
	return m.table, m.tableErr
}

func (m *mockMetadataService) GetTableByID(tableID uint) (*entity.SysTable, error) {
	return m.table, m.tableErr
}

func (m *mockMetadataService) GetColumns(tableID uint) ([]*entity.SysColumn, error) {
	return m.columns, m.columnsErr
}

func (m *mockMetadataService) GetTableRefs(tableID uint) ([]*entity.SysTableRef, error) {
	return m.refs, m.refsErr
}

func (m *mockMetadataService) GetActions(tableID uint) ([]*entity.SysAction, error) {
	return m.actions, m.actionsErr
}

func (m *mockMetadataService) RefreshCache() error {
	return m.refreshErr
}

func (m *mockMetadataService) GetMetadataVersion() string {
	return m.version
}

func (m *mockMetadataService) GetForeignKeyOptions(
	tableID uint, columnID *uint, search string, filters map[string]string, page, pageSize int, userID uint,
) ([]map[string]interface{}, int64, string, error) {
	return m.fkOptions, int64(len(m.fkOptions)), "N", m.fkOptionsErr
}

func (m *mockMetadataService) GetForeignKeyDisplayValue(
	tableID uint, value string, columnID *uint, userID uint,
) (string, error) {
	return m.fkDisplay, m.fkDisplayErr
}

func (m *mockMetadataService) ResolveDisplayColumn(tableID uint, columnID *uint) (*entity.SysColumn, error) {
	if m.columnsErr != nil {
		return nil, m.columnsErr
	}
	if len(m.columns) > 0 {
		return m.columns[0], nil
	}
	return nil, nil
}

func (m *mockMetadataService) ResolveInputColumn(tableID uint, columnID *uint) (*entity.SysColumn, error) {
	return m.ResolveDisplayColumn(tableID, columnID)
}

// mockDictRepoForMeta mock DictRepository for metadata tests
type mockDictRepoForMeta struct{}

func (m *mockDictRepoForMeta) GetAllDicts() ([]*entity.SysDict, error) { return nil, nil }
func (m *mockDictRepoForMeta) GetDictByName(name string) (*entity.SysDict, error) {
	return nil, nil
}
func (m *mockDictRepoForMeta) GetDictByID(id uint) (*entity.SysDict, error) { return nil, nil }
func (m *mockDictRepoForMeta) GetDictItems(dictID uint) ([]*entity.SysDictItem, error) {
	return []*entity.SysDictItem{}, nil
}
func (m *mockDictRepoForMeta) GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error) {
	return nil, nil
}

func TestMetadataHandler_GetTable_InvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/table/:tableName", h.GetTable)

	w := httptest.NewRecorder()
	// Gin 的路由参数不会为空，但服务可能返回 nil
	req, _ := http.NewRequest("GET", "/metadata/table/nonexistent", nil)
	r.ServeHTTP(w, req)

	// tableName 存在但 service 返回 nil
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestMetadataHandler_GetTable_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{
		table: &entity.SysTable{Name: "test_table", DisplayName: "Test"},
	}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/table/:tableName", h.GetTable)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/table/test_table", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestMetadataHandler_GetColumns_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/columns/:tableId", h.GetColumns)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/columns/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestMetadataHandler_GetColumns_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{
		columns: []*entity.SysColumn{
			{DbName: "ID", DisplayName: "ID"},
			{DbName: "NAME", DisplayName: "名称"},
		},
	}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/columns/:tableId", h.GetColumns)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/columns/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestMetadataHandler_RefreshCache_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.POST("/metadata/cache/refresh", h.RefreshCache)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/metadata/cache/refresh", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestMetadataHandler_GetMetadataVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{version: "20260521120000"}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/version", h.GetMetadataVersion)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/version", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestMetadataHandler_GetForeignKeyOptions_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/fk", h.GetForeignKeyOptions)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/fk?tableId=1", nil)
	r.ServeHTTP(w, req)

	// 无 userID 上下文应返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestMetadataHandler_GetForeignKeyOptions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockMetadataService{
		fkOptions: []map[string]interface{}{
			{"id": 1, "name": "Option 1"},
		},
	}
	h := NewMetadataHandler(mock, &mockDictRepoForMeta{})

	r := gin.New()
	r.GET("/metadata/fk", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.GetForeignKeyOptions(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/fk?tableId=1&search=test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
