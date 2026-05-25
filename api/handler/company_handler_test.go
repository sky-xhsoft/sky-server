package handler

import (
	"bytes"
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
)

// fakeCompanyService 实现 company.Service 接口
type fakeCompanyService struct {
	registerErr error
}

func (m *fakeCompanyService) Register(ctx context.Context, req *company.RegisterRequest) error {
	return m.registerErr
}

func TestCompanyHandler_Register_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCompanyService{}
	h := NewCompanyHandler(svc)

	r := gin.New()
	r.POST("/register", h.Register)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCompanyHandler_Register_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCompanyService{
		registerErr: stderrors.New("db error"),
	}
	h := NewCompanyHandler(svc)

	r := gin.New()
	r.POST("/register", h.Register)

	w := httptest.NewRecorder()
	body := `{"name":"Test","code":"TC","username":"admin","password":"password123","confirmPassword":"password123","adminName":"Admin"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestCompanyHandler_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCompanyService{}
	h := NewCompanyHandler(svc)

	r := gin.New()
	r.POST("/register", h.Register)

	w := httptest.NewRecorder()
	body := `{"name":"Test","code":"TC","username":"admin","password":"password123","confirmPassword":"password123","adminName":"Admin"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
