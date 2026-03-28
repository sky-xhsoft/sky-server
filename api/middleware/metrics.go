package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/monitor"
	"go.uber.org/zap"
)

// Metrics 性能监控中间件
func Metrics() gin.HandlerFunc {
	collector := monitor.GetMetricsCollector()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// 记录请求开始
		collector.IncrementCounter("http.requests.total", 1)
		collector.IncrementCounter("http.requests."+method+".total", 1)
		if path != "" {
			collector.IncrementCounter("http.requests.path."+path, 1)
		}

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 记录请求耗时
		if path != "" {
			collector.RecordRequestDuration(path, duration)
		}
		collector.RecordDuration("http.request.duration", duration)

		// 记录响应状态
		status := c.Writer.Status()
		collector.IncrementCounter("http.responses."+strconv.Itoa(status), 1)

		if status >= 400 {
			collector.IncrementCounter("http.requests.error", 1)
		} else {
			collector.IncrementCounter("http.requests.success", 1)
		}

		// 慢请求告警
		if duration > 1*time.Second {
			logger.Warn("Slow request detected",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("full_path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.String("client_ip", c.ClientIP()))
		}

		// 错误请求记录
		if status >= 500 {
			logger.Error("Server error request",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("full_path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.String("client_ip", c.ClientIP()))
		}
	}
}

// MetricsStats 指标统计
func MetricsStats() gin.HandlerFunc {
	collector := monitor.GetMetricsCollector()

	return func(c *gin.Context) {
		metrics := collector.GetAllMetrics()
		c.JSON(200, gin.H{
			"code":    200,
			"message": "success",
			"data":    metrics,
		})
	}
}

// RequestTiming 请求计时中间件
func RequestTiming() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 存储开始时间到上下文
		c.Set("request_start_time", start)

		c.Next()

		// 计算总耗时
		duration := time.Since(start)
		c.Header("X-Response-Time", strconv.FormatInt(duration.Milliseconds(), 10)+"ms")
	}
}

// DatabaseMetrics 数据库监控
type DatabaseMetrics struct {
	QueryCount     int64
	QueryDuration  int64
	SlowQueryCount int64
}

var (
	dbMetrics     DatabaseMetrics
	dbMetricsLock = monitor.GetMetricsCollector()
)

// RecordQuery 记录数据库查询
func RecordQuery(duration time.Duration, isSlow bool) {
	dbMetricsLock.IncrementCounter("db.queries.total", 1)
	dbMetricsLock.RecordDuration("db.query.duration", duration)

	if isSlow {
		dbMetricsLock.IncrementCounter("db.queries.slow", 1)
		logger.Warn("Slow database query",
			zap.Duration("duration", duration))
	}
}

// GetDatabaseMetrics 获取数据库指标
func GetDatabaseMetrics() DatabaseMetrics {
	return DatabaseMetrics{
		QueryCount:    dbMetricsLock.GetCounter("db.queries.total"),
		SlowQueryCount: dbMetricsLock.GetCounter("db.queries.slow"),
	}
}

// ResetMetrics 重置指标
func ResetMetrics(c *gin.Context) {
	collector := monitor.GetMetricsCollector()
	collector.Reset()
	c.JSON(200, gin.H{
		"code":    200,
		"message": "Metrics reset successfully",
	})
}

// HealthCheck 健康检查
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		collector := monitor.GetMetricsCollector()

		totalRequests := collector.GetCounter("http.requests.total")
		errorRequests := collector.GetCounter("http.requests.error")
		successRequests := collector.GetCounter("http.requests.success")

		var errorRate float64
		if totalRequests > 0 {
			errorRate = float64(errorRequests) / float64(totalRequests) * 100
		}

		status := "healthy"
		httpStatus := 200

		// 如果错误率超过10%，标记为degraded
		if errorRate > 10 {
			status = "degraded"
		}

		// 如果错误率超过30%，标记为unhealthy
		if errorRate > 30 {
			status = "unhealthy"
			httpStatus = 503
		}

		c.JSON(httpStatus, gin.H{
			"code":    httpStatus,
			"message": status,
			"data": gin.H{
				"status":          status,
				"total_requests":  totalRequests,
				"success_requests": successRequests,
				"error_requests":  errorRequests,
				"error_rate":      errorRate,
				"timestamp":       time.Now().Format(time.RFC3339),
			},
		})
	}
}
