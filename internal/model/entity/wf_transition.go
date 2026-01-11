package entity

// WfTransition 流程流转
type WfTransition struct {
	BaseModel
	WfDefinitionID uint   `gorm:"column:WF_DEFINITION_ID;not null;index" json:"wfDefinitionId"`
	FromNodeID     uint   `gorm:"column:FROM_NODE_ID;not null;index" json:"fromNodeId"`
	ToNodeID       uint   `gorm:"column:TO_NODE_ID;not null;index" json:"toNodeId"`
	Name           string `gorm:"column:NAME;size:80" json:"name"`
	Condition      string `gorm:"column:CONDITION;size:500" json:"condition"` // 流转条件表达式
	Orderno        int    `gorm:"column:ORDERNO" json:"orderno"`              // 优先级顺序
}

// TableName 指定表名
func (WfTransition) TableName() string {
	return "wf_transition"
}
