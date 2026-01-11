package entity

import "time"

// WfInstance 流程实例
type WfInstance struct {
	BaseModel
	WfDefinitionID uint      `gorm:"column:WF_DEFINITION_ID;not null;index" json:"wfDefinitionId"`
	SysTableID     int       `gorm:"column:SYS_TABLE_ID;index" json:"sysTableId"`             // 关联的业务表
	BusinessID     uint      `gorm:"column:BUSINESS_ID;index" json:"businessId"`              // 业务数据ID
	Status         string    `gorm:"column:STATUS;size:20;not null" json:"status"`            // running:运行中, completed:已完成, terminated:已终止, suspended:已挂起
	CurrentNodeID  uint      `gorm:"column:CURRENT_NODE_ID;index" json:"currentNodeId"`       // 当前节点ID
	StartUserID    uint      `gorm:"column:START_USER_ID;index" json:"startUserId"`           // 发起人
	StartTime      time.Time `gorm:"column:START_TIME" json:"startTime"`                      // 开始时间
	EndTime        time.Time `gorm:"column:END_TIME" json:"endTime"`                          // 结束时间
	Variables      string    `gorm:"column:VARIABLES;type:text" json:"variables"`             // 流程变量(JSON)
	Title          string    `gorm:"column:TITLE;size:255" json:"title"`                      // 流程标题
}

// TableName 指定表名
func (WfInstance) TableName() string {
	return "wf_instance"
}
