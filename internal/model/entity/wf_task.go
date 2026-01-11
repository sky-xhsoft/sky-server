package entity

import "time"

// WfTask 工作流任务
type WfTask struct {
	BaseModel
	WfInstanceID uint      `gorm:"column:WF_INSTANCE_ID;not null;index" json:"wfInstanceId"`
	WfNodeID     uint      `gorm:"column:WF_NODE_ID;not null;index" json:"wfNodeId"`
	AssigneeID   uint      `gorm:"column:ASSIGNEE_ID;index" json:"assigneeId"`           // 任务执行人
	Status       string    `gorm:"column:STATUS;size:20;not null" json:"status"`         // pending:待处理, completed:已完成, rejected:已拒绝, transferred:已转交
	Action       string    `gorm:"column:ACTION;size:20" json:"action"`                  // approve:同意, reject:拒绝, transfer:转交
	Comment      string    `gorm:"column:COMMENT;size:2000" json:"comment"`              // 审批意见
	ClaimTime    time.Time `gorm:"column:CLAIM_TIME" json:"claimTime"`                   // 签收时间
	CompleteTime time.Time `gorm:"column:COMPLETE_TIME" json:"completeTime"`             // 完成时间
	DueTime      time.Time `gorm:"column:DUE_TIME" json:"dueTime"`                       // 截止时间
	Priority     int       `gorm:"column:PRIORITY;default:0" json:"priority"`            // 优先级
	Variables    string    `gorm:"column:VARIABLES;type:text" json:"variables"`          // 任务变量(JSON)
}

// TableName 指定表名
func (WfTask) TableName() string {
	return "wf_task"
}
