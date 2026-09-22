package entity

import "time"

// SysActionLog 动作执行日志(自含字段,与物理表 SYS_ACTION_LOG 完全一致,不嵌入 BaseModel)
type SysActionLog struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint       `gorm:"column:SYS_COMPANY_ID" json:"sysCompanyId"`
	ActionID     uint       `gorm:"column:ACTION_ID;index:idx_action_id" json:"actionId"`
	ActionName   string     `gorm:"column:ACTION_NAME;size:80" json:"actionName"`
	Table        string     `gorm:"column:TABLE_NAME;size:255" json:"tableName"`
	RecordID     uint       `gorm:"column:RECORD_ID" json:"recordId"`
	Params       string     `gorm:"column:PARAMS;type:text" json:"params"`
	Result       string     `gorm:"column:RESULT;size:255" json:"result"`
	Error        string     `gorm:"column:ERROR;type:text" json:"error"`
	DurationMs   int64      `gorm:"column:DURATION_MS" json:"durationMs"`
	ExecTime     *time.Time `gorm:"column:EXEC_TIME" json:"execTime"`
	ExecBy       string     `gorm:"column:EXEC_BY;size:80" json:"execBy"`
	CreateTime   time.Time  `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time  `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
	IsActive     string     `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"`
}

// TableName 指定表名
func (SysActionLog) TableName() string {
	return "sys_action_log"
}