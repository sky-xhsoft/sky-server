package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/config"
)

func TestParseSubdomainCompanyID(t *testing.T) {
	const suffix = ".zhibo.skyzhou.cn"
	cases := []struct {
		host string
		want uint
		ok   bool
	}{
		{"1.zhibo.skyzhou.cn", 1, true},
		{"46.zhibo.skyzhou.cn", 46, true},
		{"123.zhibo.skyzhou.cn", 123, true},
		{"zhibo.skyzhou.cn", 0, false},
		{"a.1.zhibo.skyzhou.cn", 0, false},
		{"foo.zhibo.skyzhou.cn", 0, false},
		{"localhost", 0, false},
		{"0.zhibo.skyzhou.cn", 0, false},
	}
	for _, c := range cases {
		got, ok := parseSubdomainCompanyID(c.host, suffix)
		if got != c.want || ok != c.ok {
			t.Errorf("host=%q got=(%d,%v) want=(%d,%v)", c.host, got, ok, c.want, c.ok)
		}
	}
}

// 未配置子域后缀（正式环境）时不按子域识别公司
func TestParseSubdomainCompanyID_EmptySuffix(t *testing.T) {
	if got, ok := parseSubdomainCompanyID("1.zhibo.skyzhou.cn", ""); ok || got != 0 {
		t.Errorf("empty suffix should not match, got=(%d,%v)", got, ok)
	}
}

func TestCORS_SubdomainSuffix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.CORSConfig{AllowMethods: []string{"GET"}, AllowHeaders: []string{"Content-Type"}}
	allowed := func(suffix, origin string) bool {
		r := gin.New()
		r.Use(CORS(cfg, suffix))
		r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Header().Get("Access-Control-Allow-Origin") == origin
	}
	if !allowed(".zhibo.skyzhou.cn", "http://2.zhibo.skyzhou.cn:5173") {
		t.Error("configured suffix should allow company subdomain origin")
	}
	if allowed(".zhibo.skyzhou.cn", "http://evil.example.com") {
		t.Error("unrelated origin should not be allowed")
	}
	if allowed("", "http://2.zhibo.skyzhou.cn") {
		t.Error("empty suffix should not allow subdomain origin")
	}
}
