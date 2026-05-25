package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/monitor"
)

// PrometheusMetrics 提供 Prometheus 格式的指标导出
// 兼容 Prometheus 文本格式 (https://prometheus.io/docs/instrumenting/exposition_formats/)
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		collector := monitor.GetMetricsCollector()
		metrics := collector.GetAllMetrics()

		var sb strings.Builder

		// 写入计数器指标
		if counters, ok := metrics["counters"].(map[string]int64); ok {
			writeCounters(&sb, counters)
		}

		// 写入仪表盘指标
		if gauges, ok := metrics["gauges"].(map[string]float64); ok {
			writeGauges(&sb, gauges)
		}

		// 写入直方图统计
		if histograms, ok := metrics["histograms"].(map[string]interface{}); ok {
			writeHistograms(&sb, histograms)
		}

		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, sb.String())
	}
}

// writeCounters 写入计数器（COUNTER 类型）
func writeCounters(sb *strings.Builder, counters map[string]int64) {
	for name, value := range counters {
		metricName := sanitizeMetricName(name)
		fmt.Fprintf(sb, "# TYPE %s counter\n", metricName)
		fmt.Fprintf(sb, "%s %d\n", metricName, value)
	}
}

// writeGauges 写入仪表盘（GAUGE 类型）
func writeGauges(sb *strings.Builder, gauges map[string]float64) {
	for name, value := range gauges {
		metricName := sanitizeMetricName(name)
		fmt.Fprintf(sb, "# TYPE %s gauge\n", metricName)
		fmt.Fprintf(sb, "%s %s\n", metricName, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

// writeHistograms 写入直方图（SUMMARY 类型）
func writeHistograms(sb *strings.Builder, histograms map[string]interface{}) {
	for name, raw := range histograms {
		stats, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		metricName := sanitizeMetricName(name)
		fmt.Fprintf(sb, "# TYPE %s summary\n", metricName)

		if count, ok := stats["count"].(int); ok {
			fmt.Fprintf(sb, "%s_count %d\n", metricName, count)
		}
		if avg, ok := stats["avg"].(float64); ok {
			fmt.Fprintf(sb, "%s_avg %s\n", metricName, strconv.FormatFloat(avg, 'f', -1, 64))
		}
		if min, ok := stats["min"].(float64); ok {
			fmt.Fprintf(sb, "%s{quantile=\"0\"} %s\n", metricName, strconv.FormatFloat(min, 'f', -1, 64))
		}
		if max, ok := stats["max"].(float64); ok {
			fmt.Fprintf(sb, "%s{quantile=\"1\"} %s\n", metricName, strconv.FormatFloat(max, 'f', -1, 64))
		}
	}
}

// sanitizeMetricName 将指标名转换为 Prometheus 兼容格式
// Prometheus 规范: [a-zA-Z_:][a-zA-Z0-9_:]*
func sanitizeMetricName(name string) string {
	name = "sky_" + strings.NewReplacer(".", "_", "-", "_", "/", "_").Replace(name)
	// 移除非法字符
	var sb strings.Builder
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			sb.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			sb.WriteRune(r)
		case r >= '0' && r <= '9' && i > 0:
			sb.WriteRune(r)
		case r == '_' || r == ':':
			sb.WriteRune(r)
		default:
			sb.WriteRune('_')
		}
	}
	return sb.String()
}