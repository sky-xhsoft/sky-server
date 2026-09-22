package entity

import "time"

// SysScheduleLog 自动任务执行日志(仿 AD_PINSTANCE + AUTOTASK_LOG)
type SysScheduleLog struct {
	BaseModel
	ScheduleID uint       `gorm:"column:SCHEDULE_ID;not null;index:idx_schedule_id" json:"scheduleId"`
	RunAt      *time.Time `gorm:"column:RUN_AT" json:"runAt"`
	Status     string     `gorm:"column:STATUS;size:16" json:"status"` // SUCCESS/FAIL
	Message    string     `gorm:"column:MESSAGE;size:2000" json:"message"`
	Result     string     `gorm:"column:RESULT;type:text" json:"result"`
	Error      string     `gorm:"column:ERROR;type:text" json:"error"`
	DurationMs int64      `gorm:"column:DURATION_MS" json:"durationMs"`
}

// TableName 指定表名
func (SysScheduleLog) TableName() string {
	return "sys_schedule_log"
}