package billing

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	ratedDoneKey     = "rated_done" // 已完成计价的最后一天
	dailyDoneKey     = "daily_done" // 已完成每日计量的最后一天
	dailyMaxAttempts = 6            // 每日计量连续失败的最多重试次数
)

// SchedulerConfig 计量定时任务配置
type SchedulerConfig struct {
	Interval  time.Duration // 推流段落整理与巡检间隔，默认 5 分钟
	DailyHour int           // 每日计量的执行时刻（腾讯云次日 3 点出前一天数据），默认 4 点
}

// Tick 一次定时计量：整理回调、巡检在线流、刷新今天和昨天的时长；到点后计算昨天的全部计费点
func (m *Meter) Tick(ctx context.Context, cfg SchedulerConfig) {
	if _, err := m.ProcessEvents(ctx); err != nil {
		zap.L().Error("计量：整理推流回调失败", zap.Error(err))
	}
	if err := m.Patrol(ctx); err != nil {
		zap.L().Error("计量：巡检在线流失败", zap.Error(err))
	}
	now := m.now()
	today := now.Format(dayLayout)
	yesterday := now.AddDate(0, 0, -1).Format(dayLayout)
	for _, day := range []string{yesterday, today} {
		if err := m.ComputeDuration(ctx, day); err != nil {
			zap.L().Error("计量：计算直播时长失败", zap.String("day", day), zap.Error(err))
		}
	}
	if err := m.RenewDue(ctx); err != nil {
		zap.L().Error("计费：处理到期套餐失败", zap.Error(err))
	}
	if now.Hour() < cfg.DailyHour {
		return
	}
	if m.getState(ctx, dailyDoneKey) < yesterday {
		m.dailyMeter(ctx, yesterday)
	}
	m.rateAndClose(ctx, now)
}

// rateAndClose 每日计量完成后计价昨天；上月最后一天计价后为上月出账
func (m *Meter) rateAndClose(ctx context.Context, now time.Time) {
	yesterday := now.AddDate(0, 0, -1).Format(dayLayout)
	if m.getState(ctx, dailyDoneKey) >= yesterday && m.getState(ctx, ratedDoneKey) < yesterday {
		if err := m.RateDay(ctx, yesterday); err != nil {
			zap.L().Error("计费：计价未全部完成", zap.String("day", yesterday), zap.Error(err))
		}
		_ = m.setState(ctx, ratedDoneKey, yesterday)
	}
	month := now.Format(monthLayoutBill)
	last := prevMonth(month)
	lastDay := time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.Local).Format(dayLayout)
	if m.getState(ctx, statementClosedKey) < last && m.getState(ctx, ratedDoneKey) >= lastDay {
		if err := m.GenerateStatements(ctx, last); err != nil {
			zap.L().Error("计费：出账未全部完成", zap.String("month", last), zap.Error(err))
		} else {
			zap.L().Info("计费：月账单已生成", zap.String("month", last))
		}
	}
}

// dailyMeter 计算前一天的全部计费点（流量拉取失败时重试，达到次数上限后不再重试）
func (m *Meter) dailyMeter(ctx context.Context, yesterday string) {
	if err := m.RunDaily(ctx, yesterday); err != nil {
		// 流量拉取失败时下一轮重试（成功的计费点已写入）；连续失败 dailyMaxAttempts 次后不再重试，
		// 失败原因留在流量对账中，由平台在对账页手动重新计量
		attemptsKey := "daily_attempts:" + yesterday
		n := 0
		fmt.Sscanf(m.getState(ctx, attemptsKey), "%d", &n)
		n++
		_ = m.setState(ctx, attemptsKey, fmt.Sprint(n))
		zap.L().Error("计量：每日计量未全部完成", zap.String("day", yesterday), zap.Int("attempt", n), zap.Error(err))
		if n < dailyMaxAttempts {
			return
		}
	}
	if err := m.setState(ctx, dailyDoneKey, yesterday); err != nil {
		zap.L().Error("计量：保存每日计量状态失败", zap.Error(err))
	}
	zap.L().Info("计量：每日计量完成", zap.String("day", yesterday))
}

// Start 启动计量定时任务
func (m *Meter) Start(ctx context.Context, cfg SchedulerConfig) {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Minute
	}
	if cfg.DailyHour <= 0 || cfg.DailyHour > 23 {
		cfg.DailyHour = 4
	}
	go func() {
		m.Tick(ctx, cfg)
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.Tick(ctx, cfg)
			}
		}
	}()
}
