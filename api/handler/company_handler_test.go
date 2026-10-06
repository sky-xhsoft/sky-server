package handler

import (
	"bytes"
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
)

// fakeSmsService 实现 sms.Service：验证码固定为 123456
type fakeSmsService struct{}

func (f *fakeSmsService) SendVerifyCode(ctx context.Context, phone string) error { return nil }

func (f *fakeSmsService) VerifyCode(ctx context.Context, phone, code string) (bool, error) {
	return code == "123456", nil
}

// fakeCompanyService 实现 company.Service 接口
type fakeCompanyService struct {
	registerErr error
}

func (m *fakeCompanyService) Register(ctx context.Context, req *company.RegisterRequest) error {
	return m.registerErr
}

func (m *fakeCompanyService) GetCompanyConfig(ctx context.Context, companyID uint) (*entity.SysCompanyConf, error) {
	return &entity.SysCompanyConf{}, nil
}

func (m *fakeCompanyService) UpdateCompanyConfig(ctx context.Context, companyID uint, req *company.UpdateCompanyConfigRequest) (*entity.SysCompanyConf, error) {
	return &entity.SysCompanyConf{}, nil
}

func TestCompanyHandler_Register_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCompanyService{}
	h := NewCompanyHandler(svc, nil)

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
	h := NewCompanyHandler(svc, &fakeSmsService{})

	r := gin.New()
	r.POST("/register", h.Register)

	w := httptest.NewRecorder()
	body := `{"name":"Test","code":"TC","username":"admin","password":"password123","confirmPassword":"password123","adminName":"Admin","phone":"13800000000","verifyCode":"123456"}`
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
	h := NewCompanyHandler(svc, &fakeSmsService{})

	r := gin.New()
	r.POST("/register", h.Register)

	w := httptest.NewRecorder()
	body := `{"name":"Test","code":"TC","username":"admin","password":"password123","confirmPassword":"password123","adminName":"Admin","phone":"13800000000","verifyCode":"123456"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCompanyHandler_Register_WrongVerifyCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCompanyService{}
	h := NewCompanyHandler(svc, &fakeSmsService{})

	r := gin.New()
	r.POST("/register", h.Register)

	w := httptest.NewRecorder()
	body := `{"name":"Test","code":"TC","username":"admin","password":"password123","confirmPassword":"password123","adminName":"Admin","phone":"13800000000","verifyCode":"000000"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for wrong verify code, got %d", w.Code)
	}
}
