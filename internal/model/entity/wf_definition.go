package entity

// WfDefinition 工作流定义
type WfDefinition struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:80;not null" json:"name"`
	DisplayName string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	Version     int    `gorm:"column:VERSION;not null;default:1" json:"version"`
	Status      string `gorm:"column:STATUS;size:20;not null;default:'draft'" json:"status"` // draft:草稿, published:已发布, archived:已归档
	SysTableID  int    `gorm:"column:SYS_TABLE_ID;index" json:"sysTableId"`                   // 关联的业务表
	Description string `gorm:"column:DESCRIPTION;size:2000" json:"description"`
	Config      string `gorm:"column:CONFIG;type:text" json:"config"` // JSON配置
}

// TableName 指定表名
func (WfDefinition) TableName() string {
	return "wf_definition"
}
