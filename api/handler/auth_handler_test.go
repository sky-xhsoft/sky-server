package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	pkgErrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/sso"
)

// mockSSOService SSO 服务 mock
type mockSSOService struct {
	loginResp         *sso.LoginResponse
	loginErr          error
	refreshResp       *sso.TokenResponse
	refreshErr        error
	logoutErr         error
	logoutAllErr      error
	sessions          []*sso.SessionInfo
	sessionsErr       error
	kickErr           error
	changePasswordErr error
}

func (m *mockSSOService) Login(req *sso.LoginRequest) (*sso.LoginResponse, error) {
	return m.loginResp, m.loginErr
}

func (m *mockSSOService) RefreshToken(refreshToken string) (*sso.TokenResponse, error) {
	return m.refreshResp, m.refreshErr
}

func (m *mockSSOService) Logout(userID uint, deviceID string, token string) error {
	return m.logoutErr
}

func (m *mockSSOService) LogoutAll(userID uint) error {
	return m.logoutAllErr
}

func (m *mockSSOService) GetActiveSessions(userID uint) ([]*sso.SessionInfo, error) {
	return m.sessions, m.sessionsErr
}

func (m *mockSSOService) KickDevice(userID uint, deviceID string) error {
	return m.kickErr
}

func (m *mockSSOService) ChangePassword(userID uint, oldPassword string, newPassword string) error {
	return m.changePasswordErr
}

func (m *mockSSOService) GetCurrentUser(userID uint) (*sso.CurrentUserResponse, error) {
	return nil, nil
}

// setupAuthHandlerTest 创建测试用 router 和 mock 服务
func setupAuthHandlerTest(mock *mockSSOService) (*gin.Engine, *AuthHandler) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(mock)
	r := gin.New()
	return r, handler
}

func TestAuthHandler_Login_InvalidRequestBody(t *testing.T) {
	mock := &mockSSOService{}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/login", h.Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestAuthHandler_Login_MissingCompanyID(t *testing.T) {
	mock := &mockSSOService{}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/login", h.Login)

	// 不提供 companyId 且无域名识别
	body, _ := json.Marshal(map[string]interface{}{
		"username":   "test",
		"password":   "test123",
		"clientType": "web",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing companyId, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mock := &mockSSOService{
		loginErr: pkgErrors.InvalidCredentials,
	}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/login", h.Login)

	cid := uint(1)
	body, _ := json.Marshal(map[string]interface{}{
		"username":   "test",
		"password":   "wrong",
		"clientType": "web",
		"companyId":  cid,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mock := &mockSSOService{
		loginResp: &sso.LoginResponse{
			Token:        "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			User: &sso.UserInfo{
				ID:       1,
				Username: "test",
			},
		},
	}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/login", h.Login)

	cid := uint(1)
	body, _ := json.Marshal(map[string]interface{}{
		"username":   "test",
		"password":   "test123",
		"clientType": "web",
		"companyId":  cid,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_RefreshToken_Invalid(t *testing.T) {
	mock := &mockSSOService{
		refreshErr: errors.New("invalid token"),
	}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/refresh", h.RefreshToken)

	body, _ := json.Marshal(map[string]interface{}{
		"refreshToken": "invalid",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mock := &mockSSOService{
		refreshResp: &sso.TokenResponse{
			Token:     "new-token",
			ExpiresIn: 3600,
		},
	}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/refresh", h.RefreshToken)

	body, _ := json.Marshal(map[string]interface{}{
		"refreshToken": "old-refresh-token",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	mock := &mockSSOService{}
	r, h := setupAuthHandlerTest(mock)
	r.POST("/logout", h.Logout)

	body, _ := json.Marshal(map[string]interface{}{
		"deviceId": "device-1",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 未设置 userID 上下文，应返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mock := &mockSSOService{}
	r, h := setupAuthHandlerTest(mock)

	// 模拟认证中间件设置上下文
	r.POST("/logout", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("token", "fake-token")
		h.Logout(c)
	})

	body, _ := json.Marshal(map[string]interface{}{
		"deviceId": "device-1",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
