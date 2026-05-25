package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestNewIPRateLimiter(t *testing.T) {
	lim := NewIPRateLimiter(2, 5)
	if lim == nil {
		t.Fatal("expected non-nil limiter")
	}
	if lim.rate != 2 {
		t.Errorf("expected rate=2, got %v", lim.rate)
	}
	if lim.burst != 5 {
		t.Errorf("expected burst=5, got %d", lim.burst)
	}
	if len(lim.limiters) != 0 {
		t.Errorf("expected empty limiters map")
	}
}

func TestIPLimiter_GetLimiter_New(t *testing.T) {
	lim := NewIPRateLimiter(2, 5)
	got := lim.getLimiter("192.168.1.1")
	if got == nil {
		t.Fatal("expected non-nil limiter")
	}
	if got.Burst() != 5 {
		t.Errorf("expected burst=5, got %d", got.Burst())
	}
}

func TestIPLimiter_GetLimiter_Cached(t *testing.T) {
	lim := NewIPRateLimiter(2, 5)
	first := lim.getLimiter("10.0.0.1")
	second := lim.getLimiter("10.0.0.1")
	if first != second {
		t.Error("expected same limiter instance for same IP")
	}
}

func TestIPLimiter_Limit_Allows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := NewIPRateLimiter(100, 100) // high limits ensure no blocking

	r := gin.New()
	r.Use(lim.Limit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIPLimiter_Limit_Blocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := NewIPRateLimiter(1, 1) // rate=1/s, burst=1

	r := gin.New()
	r.Use(lim.Limit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// First request should pass (uses burst token)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "10.0.0.99:12345"
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", w1.Code)
	}

	// Second request should be rate limited (burst depleted)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.99:12345"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request should be blocked with 429, got %d", w2.Code)
	}
}

func TestIPLimiter_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := NewIPRateLimiter(1, 1) // each IP gets burst=1

	r := gin.New()
	r.Use(lim.Limit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// IP-A uses its burst token
	_ = lim.getLimiter("ip-a").Allow()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "ip-b:12345"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("different IP should not be blocked, got %d", w.Code)
	}
}

func TestIPLimiter_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lim := NewIPRateLimiter(1, 0)

	r := gin.New()
	r.Use(lim.Limit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// First request passes
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	r.ServeHTTP(w1, req1)

	// Second request blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
	if w2.Body.Len() == 0 {
		t.Error("expected response body")
	}
}

func TestIPLimiter_CleanupTick(t *testing.T) {
	lim := NewIPRateLimiter(2, 5)
	lim.getLimiter("192.168.1.1")
	if len(lim.limiters) != 1 {
		t.Fatalf("expected 1 limiter, got %d", len(lim.limiters))
	}

	// Simulate cleanup
	lim.mu.Lock()
	lim.limiters = make(map[string]*rate.Limiter)
	lim.mu.Unlock()

	if len(lim.limiters) != 0 {
		t.Errorf("expected empty limiters after cleanup")
	}
}
