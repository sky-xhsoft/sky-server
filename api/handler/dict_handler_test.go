package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// mockDictService mock 字典服务
type mockDictService struct {
	items      []*entity.SysDictItem
	itemsErr   error
	defValue   string
	defErr     error
	refreshErr error
}

func (m *mockDictService) GetDictItems(dictID uint) ([]*entity.SysDictItem, error) {
	return m.items, m.itemsErr
}

func (m *mockDictService) GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error) {
	return m.items, m.itemsErr
}

func (m *mockDictService) GetDefaultValue(dictName string) (string, error) {
	return m.defValue, m.defErr
}

func (m *mockDictService) RefreshDictCache() error {
	return m.refreshErr
}

func TestDictHandler_GetDictItems_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{}
	h := NewDictHandler(mock)

	r := gin.New()
	r.GET("/dicts/:dictId", h.GetDictItems)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dicts/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestDictHandler_GetDictItems_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{
		items: []*entity.SysDictItem{
			{Value: "v1", DisplayName: "Value 1"},
			{Value: "v2", DisplayName: "Value 2"},
		},
	}
	h := NewDictHandler(mock)

	r := gin.New()
	r.GET("/dicts/:dictId", h.GetDictItems)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dicts/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestDictHandler_GetDictItems_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{itemsErr: errors.New("db error")}
	h := NewDictHandler(mock)

	r := gin.New()
	r.GET("/dicts/:dictId", h.GetDictItems)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dicts/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestDictHandler_GetDictItemsByName_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{}
	h := NewDictHandler(mock)

	r := gin.New()
	// Gin 默认会阻止空 path 参数；这里测试 service 调用
	r.GET("/dicts/name/:dictName", h.GetDictItemsByName)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dicts/name/test", nil)
	r.ServeHTTP(w, req)

	// 应该正常返回（即使 items 为空）
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestDictHandler_GetDefaultValue_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{defValue: "active"}
	h := NewDictHandler(mock)

	r := gin.New()
	r.GET("/dicts/default/:dictName", h.GetDefaultValue)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dicts/default/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 验证响应包含 defaultValue
	body, _ := json.Marshal(resp)
	if !bytes.Contains(body, []byte("active")) {
		t.Errorf("Expected response to contain 'active', got: %s", string(body))
	}
}

func TestDictHandler_RefreshCache_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{}
	h := NewDictHandler(mock)

	r := gin.New()
	r.POST("/dicts/refresh", h.RefreshCache)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dicts/refresh", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestDictHandler_RefreshCache_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockDictService{refreshErr: errors.New("cache error")}
	h := NewDictHandler(mock)

	r := gin.New()
	r.POST("/dicts/refresh", h.RefreshCache)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dicts/refresh", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}
