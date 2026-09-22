package handler

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/domain/search"
	pkgErrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/crud"
	"github.com/xuri/excelize/v2"
)

// fakeCrudService 实现 crud.Service 接口
type fakeCrudService struct {
	getOneResult       map[string]interface{}
	getOneErr          error
	getListResult      *crud.QueryResponse
	getListErr         error
	createResult       map[string]interface{}
	createErr          error
	updateErr          error
	deleteErr          error
	batchDeleteErr     error
	submitErr          error
	unsubmitErr        error
	voidErr            error
	saveWithDetailsRes *crud.SaveWithDetailsResponse
	saveWithDetailsErr error
	searchResult       *crud.QueryResponse
	searchErr          error
	exportResult       *excelize.File
	exportErr          error
	importCount        int
	importErr          error
}

func (m *fakeCrudService) GetOne(ctx context.Context, tableName string, id uint, userID uint) (map[string]interface{}, error) {
	return m.getOneResult, m.getOneErr
}
func (m *fakeCrudService) GetList(ctx context.Context, req *crud.QueryRequest, userID uint) (*crud.QueryResponse, error) {
	return m.getListResult, m.getListErr
}
func (m *fakeCrudService) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	return m.createResult, m.createErr
}
func (m *fakeCrudService) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
	return m.updateErr
}
func (m *fakeCrudService) Delete(ctx context.Context, tableName string, id uint, userID uint) error {
	return m.deleteErr
}
func (m *fakeCrudService) BatchDelete(ctx context.Context, tableName string, ids []uint, userID uint) error {
	return m.batchDeleteErr
}
func (m *fakeCrudService) Submit(ctx context.Context, tableName string, id uint, userID uint) error {
	return m.submitErr
}
func (m *fakeCrudService) Unsubmit(ctx context.Context, tableName string, id uint, userID uint) error {
	return m.unsubmitErr
}
func (m *fakeCrudService) SaveWithDetails(ctx context.Context, req *crud.SaveWithDetailsRequest, userID uint) (*crud.SaveWithDetailsResponse, error) {
	return m.saveWithDetailsRes, m.saveWithDetailsErr
}
func (m *fakeCrudService) Void(ctx context.Context, tableName string, id uint, userID uint) error {
	return m.voidErr
}
func (m *fakeCrudService) Search(ctx context.Context, req *search.SearchRequest, tableName string, userID uint) (*crud.QueryResponse, error) {
	return m.searchResult, m.searchErr
}
func (m *fakeCrudService) Export(ctx context.Context, tableName string, filters map[string]interface{}, userID uint) (*excelize.File, error) {
	return m.exportResult, m.exportErr
}
func (m *fakeCrudService) Import(ctx context.Context, tableName string, file multipart.File, userID uint) (int, error) {
	return m.importCount, m.importErr
}
func (m *fakeCrudService) ImportTemplate(ctx context.Context, tableName string, userID uint) (*excelize.File, error) {
	return excelize.NewFile(), nil
}

func TestCrudHandler_GetOne_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.GET("/data/:tableName/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.GetOne(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data/users/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCrudHandler_GetOne_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.GET("/data/:tableName/:id", h.GetOne)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data/users/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestCrudHandler_GetOne_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{
		getOneResult: map[string]interface{}{"id": 1, "name": "Test"},
	}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.GET("/data/:tableName/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.GetOne(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data/users/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCrudHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{
		createResult: map[string]interface{}{"id": 1, "name": "New"},
	}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.POST("/data/:tableName", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Create(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"name": "New"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/data/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCrudHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.PUT("/data/:tableName/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Update(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/data/users/abc", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCrudHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.DELETE("/data/:tableName/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Delete(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/data/users/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCrudHandler_BatchDelete_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.POST("/data/:tableName/batch-delete", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.BatchDelete(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/data/users/batch-delete", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCrudHandler_Submit_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{
		submitErr: stderrors.New("submit failed"),
	}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.POST("/data/:tableName/:id/submit", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Submit(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/data/users/1/submit", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestCrudHandler_Submit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.POST("/data/:tableName/:id/submit", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Submit(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/data/users/1/submit", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

// Verify error code expectation
func TestCrudHandler_Create_BusinessError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCrudService{
		createErr: pkgErrors.New(pkgErrors.ErrValidation, "字段验证失败"),
	}
	h := NewCrudHandler(svc)

	r := gin.New()
	r.POST("/data/:tableName", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Create(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"name": ""})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/data/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}
