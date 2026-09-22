// Package dbmonitor 提供数据库连接池监控功能
package dbmonitor

import (
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Stats 数据库连接池统计信息
type Stats struct {
	MaxOpenConnections int           `json:"maxOpenConnections"`
	OpenConnections    int           `json:"openConnections"`
	InUse              int           `json:"inUse"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"waitCount"`
	WaitDuration       time.Duration `json:"waitDuration"`
	MaxIdleClosed      int64         `json:"maxIdleClosed"`
	MaxLifetimeClosed  int64         `json:"maxLifetimeClosed"`
	CollectedAt        time.Time     `json:"collectedAt"`
}

// Monitor 数据库监控器
type Monitor struct {
	db        *gorm.DB
	interval  time.Duration
	mu        sync.RWMutex
	lastStats *Stats
	stopCh    chan struct{}
	running   bool
}

// NewMonitor 创建数据库监控器
func NewMonitor(db *gorm.DB, interval time.Duration) *Monitor {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Monitor{
		db:       db,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动监控
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go m.loop()
}

// Stop 停止监控
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	m.running = false
	close(m.stopCh)
}

// GetStats 获取当前统计信息（实时采集）
func (m *Monitor) GetStats() (*Stats, error) {
	sqlDB, err := m.db.DB()
	if err != nil {
		return nil, err
	}

	dbStats := sqlDB.Stats()
	stats := &Stats{
		MaxOpenConnections: dbStats.MaxOpenConnections,
		OpenConnections:    dbStats.OpenConnections,
		InUse:              dbStats.InUse,
		Idle:               dbStats.Idle,
		WaitCount:          dbStats.WaitCount,
		WaitDuration:       dbStats.WaitDuration,
		MaxIdleClosed:      dbStats.MaxIdleClosed,
		MaxLifetimeClosed:  dbStats.MaxLifetimeClosed,
		CollectedAt:        time.Now(),
	}

	m.mu.Lock()
	m.lastStats = stats
	m.mu.Unlock()

	return stats, nil
}

// GetLastStats 获取最近一次采集的统计信息
func (m *Monitor) GetLastStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastStats
}

// loop 定时采集
func (m *Monitor) loop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			stats, err := m.GetStats()
			if err != nil {
				continue
			}

			// 警告日志：连接池使用率过高
			if stats.MaxOpenConnections > 0 {
				usage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
				if usage > 0.8 {
					logger.Warn("DB connection pool usage high",
						zap.Float64("usage", usage),
						zap.Int("inUse", stats.InUse),
						zap.Int("max", stats.MaxOpenConnections))
				}
			}

			// 警告日志：等待时间过长
			if stats.WaitDuration > time.Second {
				logger.Warn("DB connection wait time high",
					zap.Duration("waitDuration", stats.WaitDuration),
					zap.Int64("waitCount", stats.WaitCount))
			}
		}
	}
}
