package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/monitor"
	"go.uber.org/zap"
)

func healthEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	if logger.Logger == nil {
		logger.Logger = zap.NewNop()
	}
	r := gin.New()
	r.Use(Metrics())
	r.GET(healthPath, HealthCheck())
	r.GET("/thumb", func(c *gin.Context) { c.Status(http.StatusNotFound) })
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
	return r
}

func hit(r *gin.Engine, path string, n int) int {
	code := 0
	for i := 0; i < n; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		code = w.Code
	}
	return code
}

// 大量 4xx（如缩略图不存在 404）不应判为不健康；健康检查自身不计入统计
func TestHealthCheck_ClientErrorsDoNotMakeUnhealthy(t *testing.T) {
	monitor.GetMetricsCollector().Reset()
	r := healthEngine()
	hit(r, "/ok", 10)
	hit(r, "/thumb", 90)
	if code := hit(r, healthPath, 5); code != http.StatusOK {
		t.Fatalf("只有 4xx 时应健康（200），got %d", code)
	}
	if total := monitor.GetMetricsCollector().GetCounter("http.requests.total"); total != 100 {
		t.Errorf("健康检查请求不应计入总数，total=%d", total)
	}
}

// 服务端错误（5xx）超过 30% 判为不健康
func TestHealthCheck_ServerErrorsMakeUnhealthy(t *testing.T) {
	monitor.GetMetricsCollector().Reset()
	r := healthEngine()
	hit(r, "/ok", 6)
	hit(r, "/boom", 4)
	if code := hit(r, healthPath, 1); code != http.StatusServiceUnavailable {
		t.Fatalf("5xx 占 40%% 时应不健康（503），got %d", code)
	}
}
