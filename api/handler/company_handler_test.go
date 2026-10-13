package handler

import (
	"bytes"
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *fakeCompanyService) CheckRegisterAvailability(ctx context.Context, username, phone string) (*company.RegisterAvailability, error) {
	return &company.RegisterAvailability{UsernameTaken: username == "taken", PhoneValid: company.ValidPhone(phone), PhoneTaken: phone == "13900000000"}, nil
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

// 用户名、手机号已被使用返回 409 与具体提示，而不是「注册失败」500
func TestCompanyHandler_Register_Taken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		err  error
		code int
	}{
		{company.ErrUsernameTaken, http.StatusConflict},
		{company.ErrPhoneTaken, http.StatusConflict},
		{company.ErrInvalidPhone, http.StatusBadRequest},
	} {
		h := NewCompanyHandler(&fakeCompanyService{registerErr: tc.err}, &fakeSmsService{})
		r := gin.New()
		r.POST("/register", h.Register)
		w := httptest.NewRecorder()
		body := `{"name":"Test","code":"TC","username":"admin","password":"password123","confirmPassword":"password123","adminName":"Admin","phone":"13800000000","verifyCode":"123456"}`
		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != tc.code || !strings.Contains(w.Body.String(), tc.err.Error()) {
			t.Errorf("%v: 期望 %d 且提示原因，实际 %d %s", tc.err, tc.code, w.Code, w.Body.String())
		}
	}
}

func TestCompanyHandler_CheckRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCompanyHandler(&fakeCompanyService{}, nil)
	r := gin.New()
	r.GET("/register/check", h.CheckRegister)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/register/check?username=taken&phone=13900000000", nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, `"usernameTaken":true`) || !strings.Contains(body, `"phoneTaken":true`) || !strings.Contains(body, `"phoneValid":true`) {
		t.Errorf("查重结果不对: %d %s", w.Code, body)
	}
}
