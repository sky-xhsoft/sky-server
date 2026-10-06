// Package housekeeping 定时清理过期数据：过期会话、超过保留期的审计日志、已过期分享的状态
package housekeeping

import (
	"context"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// batchSize 每批删除的行数，避免一次性大删除长时间锁表
const batchSize = 5000

// Config 清理任务配置
type Config struct {
	Interval           time.Duration // 执行间隔
	AuditRetentionDays int           // 审计日志保留天数，<= 0 表示不删除
}

// Result 一次清理的结果
type Result struct {
	ExpiredSessions int64
	OldAuditLogs    int64
	ExpiredShares   int64
}

// RunOnce 执行一次清理
func RunOnce(ctx context.Context, db *gorm.DB, cfg Config, now time.Time) (Result, error) {
	var res Result
	var err error

	// 会话的 EXPIRE_TIME 为刷新令牌到期时间，过期后已无法续期
	if res.ExpiredSessions, err = deleteInBatches(ctx, db, "sys_user_session", "EXPIRE_TIME < ?", now); err != nil {
		return res, err
	}

	if cfg.AuditRetentionDays > 0 {
		cutoff := now.AddDate(0, 0, -cfg.AuditRetentionDays)
		if res.OldAuditLogs, err = deleteInBatches(ctx, db, "audit_log", "CREATED_AT < ?", cutoff); err != nil {
			return res, err
		}
	}

	// 分享过期原本只在被访问时才更新状态，这里统一更新
	r := db.WithContext(ctx).Exec("UPDATE cloud_share SET STATUS = 'expired' WHERE STATUS = 'active' AND EXPIRE_TIME IS NOT NULL AND EXPIRE_TIME < ?", now)
	if r.Error != nil {
		return res, r.Error
	}
	res.ExpiredShares = r.RowsAffected

	return res, nil
}

// deleteInBatches 按批删除满足条件的行，返回删除总数
// 先查出一批 ID 再按 ID 删除：MySQL 不支持 IN 子查询带 LIMIT，SQLite 不支持 DELETE ... LIMIT
func deleteInBatches(ctx context.Context, db *gorm.DB, table, where string, arg interface{}) (int64, error) {
	var total int64
	for {
		var ids []uint
		if err := db.WithContext(ctx).Table(table).Where(where, arg).Limit(batchSize).Pluck("ID", &ids).Error; err != nil {
			return total, err
		}
		if len(ids) == 0 {
			return total, nil
		}
		r := db.WithContext(ctx).Exec("DELETE FROM "+table+" WHERE ID IN ?", ids)
		if r.Error != nil {
			return total, r.Error
		}
		total += r.RowsAffected
		if len(ids) < batchSize {
			return total, nil
		}
	}
}

// Start 启动后台定时清理，ctx 取消时退出
func Start(ctx context.Context, db *gorm.DB, cfg Config) {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Hour
	}
	run := func() {
		res, err := RunOnce(ctx, db, cfg, time.Now())
		if err != nil {
			logger.Error("Housekeeping failed", zap.Error(err))
			return
		}
		if res.ExpiredSessions+res.OldAuditLogs+res.ExpiredShares > 0 {
			logger.Info("Housekeeping done",
				zap.Int64("expiredSessions", res.ExpiredSessions),
				zap.Int64("oldAuditLogs", res.OldAuditLogs),
				zap.Int64("expiredShares", res.ExpiredShares))
		}
	}
	go func() {
		run()
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
	logger.Info("Housekeeping task started",
		zap.Duration("interval", cfg.Interval),
		zap.Int("auditRetentionDays", cfg.AuditRetentionDays))
}
