package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTranslate_Chinese(t *testing.T) {
	got := Translate(Chinese, "unauthorized")
	if got != "未授权" {
		t.Errorf("Expected '未授权', got %q", got)
	}
}

func TestTranslate_English(t *testing.T) {
	got := Translate(English, "unauthorized")
	if got != "Unauthorized" {
		t.Errorf("Expected 'Unauthorized', got %q", got)
	}
}

func TestTranslate_FallbackOnMissingKey(t *testing.T) {
	got := Translate(Chinese, "non_existent_key")
	if got != "non_existent_key" {
		t.Errorf("Expected original key when not found, got %q", got)
	}
}

func TestTranslate_FallbackToDefaultLanguage(t *testing.T) {
	// 不存在的语言应回退到默认
	got := Translate(Language("fr-FR"), "unauthorized")
	if got != "未授权" {
		t.Errorf("Expected fallback to Chinese, got %q", got)
	}
}

func TestGetLanguageFromRequest_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	lang := GetLanguageFromRequest(c)
	if lang != DefaultLanguage {
		t.Errorf("Expected default language, got %s", lang)
	}
}

func TestGetLanguageFromRequest_English(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "en-US")
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	lang := GetLanguageFromRequest(c)
	if lang != English {
		t.Errorf("Expected English, got %s", lang)
	}
}

func TestGetLanguageFromRequest_Chinese(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "zh-CN")
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	lang := GetLanguageFromRequest(c)
	if lang != Chinese {
		t.Errorf("Expected Chinese, got %s", lang)
	}
}

func TestMiddleware_SetsLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware())

	var capturedLang Language
	r.GET("/test", func(c *gin.Context) {
		capturedLang = GetLanguageFromContext(c)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Language", "en-US")
	r.ServeHTTP(w, req)

	if capturedLang != English {
		t.Errorf("Expected English, got %s", capturedLang)
	}
}

func TestGetLanguageFromContext_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	lang := GetLanguageFromContext(c)
	if lang != DefaultLanguage {
		t.Errorf("Expected default when not set, got %s", lang)
	}
}

func TestTranslateContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware())

	var translated string
	r.GET("/test", func(c *gin.Context) {
		translated = TranslateContext(c, "unauthorized")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Language", "en-US")
	r.ServeHTTP(w, req)

	if translated != "Unauthorized" {
		t.Errorf("Expected 'Unauthorized', got %q", translated)
	}
}
