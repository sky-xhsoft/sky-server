package entity

// WfNode 流程节点
type WfNode struct {
	BaseModel
	WfDefinitionID uint   `gorm:"column:WF_DEFINITION_ID;not null;index" json:"wfDefinitionId"`
	Name           string `gorm:"column:NAME;size:80;not null" json:"name"`
	DisplayName    string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	NodeType       string `gorm:"column:NODE_TYPE;size:20;not null" json:"nodeType"` // start:开始, end:结束, user:用户任务, auto:自动任务, gateway:网关
	AssignType     string `gorm:"column:ASSIGN_TYPE;size:20" json:"assignType"`      // user:指定用户, role:角色, expression:表达式
	AssignValue    string `gorm:"column:ASSIGN_VALUE;size:500" json:"assignValue"`   // 分配值(用户ID/角色ID/表达式)
	ActionID       uint   `gorm:"column:ACTION_ID;index" json:"actionId"`            // 自动任务关联的动作ID
	Config         string `gorm:"column:CONFIG;type:text" json:"config"`             // JSON配置
	PosX           int    `gorm:"column:POS_X" json:"posX"`                          // 节点X坐标
	PosY           int    `gorm:"column:POS_Y" json:"posY"`                          // 节点Y坐标
}

// TableName 指定表名
func (WfNode) TableName() string {
	return "wf_node"
}
