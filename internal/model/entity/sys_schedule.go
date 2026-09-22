package entity

import "time"

// SysSchedule 自动任务定义(仿 AD_PROCESS+AD_PROCESSQUEUE+AD_TRIGGER)
type SysSchedule struct {
	BaseModel
	Name          string     `gorm:"column:NAME;size:255" json:"name"`
	DisplayName   string     `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	ActionType    string     `gorm:"column:ACTION_TYPE;size:32;default:go" json:"actionType"` // url/sp/js/py/go/bsh
	ActionContent string     `gorm:"column:ACTION_CONTENT;type:text" json:"actionContent"`
	Cron          string     `gorm:"column:CRON;size:64" json:"cron"`
	LastRunAt     *time.Time `gorm:"column:LAST_RUN_AT" json:"lastRunAt"`
	NextRunAt     *time.Time `gorm:"column:NEXT_RUN_AT;index:idx_next_run" json:"nextRunAt"`
	Description   string     `gorm:"column:DESCRIPTION;size:2000" json:"description"`
}

// TableName 指定表名
func (SysSchedule) TableName() string {
	return "sys_schedule"
}