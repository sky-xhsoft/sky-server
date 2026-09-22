package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/monitor"
)

func TestPrometheusMetrics_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics", PrometheusMetrics())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %s", ct)
	}
}

func TestPrometheusMetrics_OutputFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	collector := monitor.GetMetricsCollector()
	collector.IncrementCounter("test.counter", 42)
	collector.SetGauge("test.gauge", 3.14)

	r := gin.New()
	r.GET("/metrics", PrometheusMetrics())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	if !strings.Contains(body, "# TYPE") {
		t.Error("expected TYPE comments in output")
	}
	if !strings.Contains(body, "sky_test_counter") {
		t.Errorf("expected sanitized counter name, got: %s", body)
	}
}

func TestSanitizeMetricName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http.requests.total", "sky_http_requests_total"},
		{"path./api/users", "sky_path__api_users"},
		{"my-metric", "sky_my_metric"},
	}

	for _, tt := range tests {
		actual := sanitizeMetricName(tt.input)
		if actual != tt.expected {
			t.Errorf("sanitizeMetricName(%q): expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}

func TestCompression_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Compression())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, strings.Repeat("hello world ", 200))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	encoding := w.Header().Get("Content-Encoding")
	if encoding != "gzip" {
		t.Errorf("expected gzip encoding, got %s", encoding)
	}
}

func TestRequestID_GeneratesNew(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, exists := c.Get("RequestID")
		if !exists {
			t.Error("RequestID not set in context")
		}
		if id == "" {
			t.Error("RequestID is empty")
		}
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header not set in response")
	}
}

func TestRequestID_UsesClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	clientID := "client-provided-id-12345"

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("RequestID")
		if id != clientID {
			t.Errorf("expected %s, got %v", clientID, id)
		}
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", clientID)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != clientID {
		t.Errorf("X-Request-ID header should echo client value")
	}
}
