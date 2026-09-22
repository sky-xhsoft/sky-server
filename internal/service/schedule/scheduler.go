package schedule

import (
	"context"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

// TaskRunner 执行一次自动任务动作(由上层注入,复用 action/executor)
type TaskRunner func(ctx context.Context, actionType, actionContent string, sysCompanyID uint) (string, error)

// Scheduler 自动任务调度器:后台 goroutine 周期扫描 sys_schedule
type Scheduler struct {
	db       *gorm.DB
	run      TaskRunner
	interval time.Duration
	stopCh   chan struct{}
	running  bool
}

// NewScheduler 创建调度器
func NewScheduler(db *gorm.DB, run TaskRunner) *Scheduler {
	return &Scheduler{
		db:       db,
		run:      run,
		interval: 30 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动调度循环(非阻塞)
func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true
	go s.loop()
	logger.Info("schedule scheduler started", zap.Duration("interval", s.interval))
}

// Stop 停止调度
func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	close(s.stopCh)
	s.running = false
}

func (s *Scheduler) loop() {
	t := time.NewTicker(s.interval)
	defer t.Stop()
	// 启动即扫描一次
	s.scan()
	for {
		select {
		case <-s.stopCh:
			return
		case <-t.C:
			s.scan()
		}
	}
}

// scan 扫描 due 任务并执行
func (s *Scheduler) scan() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	now := time.Now()
	var due []entity.SysSchedule
	if err := s.db.WithContext(ctx).
		Where("IS_ACTIVE = 'Y' AND NEXT_RUN_AT IS NOT NULL AND NEXT_RUN_AT <= ?", now).
		Find(&due).Error; err != nil {
		logger.Warn("schedule scan query failed", zap.Error(err))
		return
	}

	for i := range due {
		sc := &due[i]
		s.exec(ctx, sc)
	}
}

// exec 执行单个任务并更新下次运行时间
func (s *Scheduler) exec(ctx context.Context, sc *entity.SysSchedule) {
	start := time.Now()
	var message, errStr string
	status := "SUCCESS"

	if s.run != nil {
		out, err := s.run(ctx, sc.ActionType, sc.ActionContent, sc.SysCompanyID)
		if err != nil {
			status = "FAIL"
			errStr = err.Error()
			message = out
		} else {
			message = out
		}
	}

	// 写日志
	dur := time.Since(start)
	now := time.Now()
	log := entity.SysScheduleLog{
		ScheduleID: sc.ID,
		RunAt:      &now,
		Status:     status,
		Message:    message,
		Error:      errStr,
		DurationMs: dur.Milliseconds(),
	}
	if lerr := s.db.WithContext(ctx).Create(&log).Error; lerr != nil {
		logger.Warn("schedule log write failed", zap.Error(lerr))
	}

	// 计算下次运行
	next, nerr := NextRunAfter(sc.Cron, time.Now())
	last := now
	updates := map[string]interface{}{"LAST_RUN_AT": last}
	if nerr == nil {
		updates["NEXT_RUN_AT"] = next
	} else {
		updates["NEXT_RUN_AT"] = nil // cron 无效则停跑
	}
	if uerr := s.db.WithContext(ctx).Model(sc).Updates(updates).Error; uerr != nil {
		logger.Warn("schedule update failed", zap.Error(uerr))
	}
}

// RunNow 手动执行一个任务(返回消息与错误)
func (s *Scheduler) RunNow(ctx context.Context, id uint) (string, error) {
	return RunNowWithDB(ctx, s.db, s.run, id)
}