package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

// RunNowWithDB 手动执行一个任务(公共,供 Scheduler/service 复用)
func RunNowWithDB(ctx context.Context, db *gorm.DB, run TaskRunner, id uint) (string, error) {
	var sc entity.SysSchedule
	// 手动执行不要求 IS_ACTIVE=Y(允许手动触发停用任务)
	if err := db.WithContext(ctx).Where("ID = ?", id).First(&sc).Error; err != nil {
		return "", err
	}

	start := time.Now()
	var message, errStr string
	status := "SUCCESS"
	if run != nil {
		out, err := run(ctx, sc.ActionType, sc.ActionContent, sc.SysCompanyID)
		if err != nil {
			status = "FAIL"
			errStr = err.Error()
			message = out
		} else {
			message = out
		}
	}

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
	if lerr := db.WithContext(ctx).Create(&log).Error; lerr != nil {
		logger.Warn("schedule runnow log write failed", zap.Error(lerr))
	}

	next, nerr := NextRunAfter(sc.Cron, time.Now())
	updates := map[string]interface{}{"LAST_RUN_AT": now}
	if nerr == nil {
		updates["NEXT_RUN_AT"] = next
	} else {
		updates["NEXT_RUN_AT"] = nil
	}
	_ = db.WithContext(ctx).Model(&sc).Updates(updates).Error

	if errStr != "" {
		return message, &ScheduleError{msg: errStr}
	}
	return message, nil
}

// RunAllEnabled 执行所有启用(IS_ACTIVE=Y)的任务,返回执行结果汇总。
// 供 action local:schedule:run-all /手动"执行全部"按钮调用。
func RunAllEnabled(ctx context.Context, db *gorm.DB, run TaskRunner) (string, error) {
	var scs []entity.SysSchedule
	if err := db.WithContext(ctx).Where("IS_ACTIVE = 'Y'").Find(&scs).Error; err != nil {
		return "", err
	}
	if len(scs) == 0 {
		return "没有启用的自动任务", nil
	}
	var okN, failN int
	var errs []string
	for _, sc := range scs {
		if _, err := RunNowWithDB(ctx, db, run, sc.ID); err != nil {
			failN++
			errs = append(errs, fmt.Sprintf("任务#%d %s: %v", sc.ID, sc.Name, err))
		} else {
			okN++
		}
	}
	msg := fmt.Sprintf("执行完成: 成功%d, 失败%d", okN, failN)
	if len(errs) > 0 {
		return msg, fmt.Errorf("%s; %s", msg, strings.Join(errs, "; "))
	}
	return msg, nil
}
type ScheduleError struct {
	msg string
}

func (e *ScheduleError) Error() string { return e.msg }

// NextRunAfter 便捷引用(实际实现在 cron_parse.go)
var _ = time.Time{}