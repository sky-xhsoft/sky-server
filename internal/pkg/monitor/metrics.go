package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu              sync.RWMutex
	counters        map[string]int64
	gauges          map[string]float64
	histograms      map[string][]time.Duration
	requestDurations map[string][]time.Duration
	maxSamples      int
}

var (
	collectorInstance *MetricsCollector
	collectorOnce     sync.Once
)

// GetMetricsCollector 获取指标收集器单例
func GetMetricsCollector() *MetricsCollector {
	collectorOnce.Do(func() {
		collectorInstance = &MetricsCollector{
			counters:        make(map[string]int64),
			gauges:          make(map[string]float64),
			histograms:      make(map[string][]time.Duration),
			requestDurations: make(map[string][]time.Duration),
			maxSamples:      1000,
		}
	})
	return collectorInstance
}

// IncrementCounter 增加计数器
func (mc *MetricsCollector) IncrementCounter(name string, value int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters[name] += value
}

// GetCounter 获取计数器值
func (mc *MetricsCollector) GetCounter(name string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.counters[name]
}

// SetGauge 设置仪表盘值
func (mc *MetricsCollector) SetGauge(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.gauges[name] = value
}

// GetGauge 获取仪表盘值
func (mc *MetricsCollector) GetGauge(name string) float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.gauges[name]
}

// RecordDuration 记录耗时
func (mc *MetricsCollector) RecordDuration(name string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.histograms[name] == nil {
		mc.histograms[name] = make([]time.Duration, 0, mc.maxSamples)
	}

	mc.histograms[name] = append(mc.histograms[name], duration)
	if len(mc.histograms[name]) > mc.maxSamples {
		mc.histograms[name] = mc.histograms[name][1:]
	}
}

// RecordRequestDuration 记录请求耗时
func (mc *MetricsCollector) RecordRequestDuration(path string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.requestDurations[path] == nil {
		mc.requestDurations[path] = make([]time.Duration, 0, mc.maxSamples)
	}

	mc.requestDurations[path] = append(mc.requestDurations[path], duration)
	if len(mc.requestDurations[path]) > mc.maxSamples {
		mc.requestDurations[path] = mc.requestDurations[path][1:]
	}
}

// GetDurationStats 获取耗时统计
func (mc *MetricsCollector) GetDurationStats(name string) (min, max, avg time.Duration, count int) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	durations := mc.histograms[name]
	if len(durations) == 0 {
		return 0, 0, 0, 0
	}

	min = durations[0]
	max = durations[0]
	sum := time.Duration(0)

	for _, d := range durations {
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
		sum += d
	}

	avg = sum / time.Duration(len(durations))
	return min, max, avg, len(durations)
}

// GetRequestStats 获取请求统计
func (mc *MetricsCollector) GetRequestStats(path string) (min, max, avg time.Duration, count int) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	durations := mc.requestDurations[path]
	if len(durations) == 0 {
		return 0, 0, 0, 0
	}

	min = durations[0]
	max = durations[0]
	sum := time.Duration(0)

	for _, d := range durations {
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
		sum += d
	}

	avg = sum / time.Duration(len(durations))
	return min, max, avg, len(durations)
}

// GetAllMetrics 获取所有指标
func (mc *MetricsCollector) GetAllMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]interface{})

	counters := make(map[string]int64)
	for k, v := range mc.counters {
		counters[k] = v
	}
	result["counters"] = counters

	gauges := make(map[string]float64)
	for k, v := range mc.gauges {
		gauges[k] = v
	}
	result["gauges"] = gauges

	durationStats := make(map[string]map[string]interface{})
	for name := range mc.histograms {
		min, max, avg, count := mc.GetDurationStats(name)
		durationStats[name] = map[string]interface{}{
			"min":    min.Milliseconds(),
			"max":    max.Milliseconds(),
			"avg":    avg.Milliseconds(),
			"count":  count,
		}
	}
	result["durations"] = durationStats

	return result
}

// Reset 重置所有指标
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.counters = make(map[string]int64)
	mc.gauges = make(map[string]float64)
	mc.histograms = make(map[string][]time.Duration)
	mc.requestDurations = make(map[string][]time.Duration)
}

// Timer 计时器
type Timer struct {
	name      string
	startTime time.Time
	collector *MetricsCollector
}

// NewTimer 创建计时器
func NewTimer(name string) *Timer {
	return &Timer{
		name:      name,
		startTime: time.Now(),
		collector: GetMetricsCollector(),
	}
}

// Stop 停止计时器并记录
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.startTime)
	t.collector.RecordDuration(t.name, duration)
	return duration
}

// Record 直接记录耗时
func (t *Timer) Record(duration time.Duration) {
	t.collector.RecordDuration(t.name, duration)
}

// HTTPMetrics HTTP指标
type HTTPMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	ErrorRequests   int64
	RequestDuration time.Duration
}

// MetricsMiddleware 指标中间件
func MetricsMiddleware() func(ctx context.Context, next func(ctx context.Context) error) error {
	return func(ctx context.Context, next func(ctx context.Context) error) error {
		start := time.Now()
		collector := GetMetricsCollector()

		// 增加请求计数
		collector.IncrementCounter("http.requests.total", 1)

		err := next(ctx)

		duration := time.Since(start)

		if err != nil {
			collector.IncrementCounter("http.requests.error", 1)
		} else {
			collector.IncrementCounter("http.requests.success", 1)
		}

		collector.RecordDuration("http.request.duration", duration)

		if duration > time.Second {
			logger.Warn("Slow request",
				zap.Duration("duration", duration))
		}

		return err
	}
}

// RequestStats 请求统计信息
type RequestStats struct {
	Path          string
	Method        string
	Count         int
	MinDuration   time.Duration
	MaxDuration   time.Duration
	AvgDuration   time.Duration
	ErrorCount    int
	SuccessCount  int
}

// GetRequestStatsByPath 获取路径请求统计
func (mc *MetricsCollector) GetRequestStatsByPath(path string) *RequestStats {
	min, max, avg, count := mc.GetRequestStats(path)

	return &RequestStats{
		Path:         path,
		Count:        count,
		MinDuration:  min,
		MaxDuration:  max,
		AvgDuration:  avg,
		ErrorCount:   int(mc.GetCounter(path + ".errors")),
		SuccessCount: count - int(mc.GetCounter(path + ".errors")),
	}
}
