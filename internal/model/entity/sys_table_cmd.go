package entity

// SysTableCmd 表命令钩子
type SysTableCmd struct {
	BaseModel
	SysTableID  int    `gorm:"column:SYS_TABLE_ID;index;not null" json:"sysTableId"`
	ActionType  string `gorm:"column:ACTION_TYPE;size:1" json:"actionType"` // 1:系统按钮
	Action      string `gorm:"column:ACTION;size:1" json:"action"`          // A:新增, M:修改, D:删除
	ActionName  string `gorm:"column:ACTION_NAME;size:255" json:"actionName"`
	Event       string `gorm:"column:EVENT;size:255" json:"event"`           // begin:开始, end:结束
	Content     string `gorm:"column:CONTENT;size:255" json:"content"`       // 执行内容
	ContentType string `gorm:"column:CONTENT_TYPE;size:255" json:"contentType"` // 内容类型
	Orderno     int    `gorm:"column:ORDERNO" json:"orderno"`
}

// TableName 指定表名
func (SysTableCmd) TableName() string {
	return "sys_table_cmd"
}
