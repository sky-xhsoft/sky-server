package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/totp"
	"github.com/sky-xhsoft/sky-server/internal/service/mfa"
)

func setupMFATest() (*gin.Engine, *MFAHandler, mfa.Service) {
	gin.SetMode(gin.TestMode)
	svc := mfa.NewService(mfa.NewInMemoryStore(), "Test-App")
	h := NewMFAHandler(svc)
	r := gin.New()
	return r, h, svc
}

func TestMFAHandler_Enroll_InvalidRequest(t *testing.T) {
	r, h, _ := setupMFATest()
	r.POST("/enroll", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Enroll(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enroll", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing accountName, got %d", w.Code)
	}
}

func TestMFAHandler_Enroll_Unauthorized(t *testing.T) {
	r, h, _ := setupMFATest()
	r.POST("/enroll", h.Enroll)

	body, _ := json.Marshal(map[string]string{"accountName": "alice"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestMFAHandler_Enroll_Success(t *testing.T) {
	r, h, _ := setupMFATest()
	r.POST("/enroll", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Enroll(c)
	})

	body, _ := json.Marshal(map[string]string{"accountName": "alice@test.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 响应应包含 secret 和 qrCodeUri
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	data := resp["data"].(map[string]interface{})
	if data["secret"] == "" {
		t.Error("secret should not be empty")
	}
	if data["qrCodeUri"] == "" {
		t.Error("qrCodeUri should not be empty")
	}
}

func TestMFAHandler_Verify_InvalidCodeFormat(t *testing.T) {
	r, h, _ := setupMFATest()
	r.POST("/verify", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Verify(c)
	})

	// 长度不足 6 位
	body, _ := json.Marshal(map[string]string{"code": "123"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid code format, got %d", w.Code)
	}
}

func TestMFAHandler_Verify_WrongCode(t *testing.T) {
	r, h, svc := setupMFATest()
	svc.EnrollUser(1, "alice")

	r.POST("/verify", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Verify(c)
	})

	body, _ := json.Marshal(map[string]string{"code": "000000"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for wrong code, got %d", w.Code)
	}
}

func TestMFAHandler_Verify_Success(t *testing.T) {
	r, h, svc := setupMFATest()
	secret, _, _ := svc.EnrollUser(1, "alice")

	r.POST("/verify", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Verify(c)
	})

	// 生成正确的 TOTP 码
	code, _ := totp.GenerateCode(secret, time.Now())
	body, _ := json.Marshal(map[string]string{"code": code})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMFAHandler_Status_NotEnabled(t *testing.T) {
	r, h, _ := setupMFATest()
	r.GET("/status", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Status(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["enabled"].(bool) {
		t.Error("Should not be enabled initially")
	}
}

func TestMFAHandler_Disable(t *testing.T) {
	r, h, svc := setupMFATest()

	// 先启用
	secret, _, _ := svc.EnrollUser(1, "alice")
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyCode(1, code)

	r.POST("/disable", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.Disable(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/disable", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	if svc.IsEnabled(1) {
		t.Error("MFA should be disabled")
	}
}
