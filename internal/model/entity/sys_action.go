package entity

// SysAction 动作定义
type SysAction struct {
	BaseModel
	SysTableID  int    `gorm:"column:SYS_TABLE_ID;index;not null" json:"sysTableId"`
	Name        string `gorm:"column:NAME;size:80" json:"name"`
	DisplayName string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	DisplayType string `gorm:"column:DISPLAY_TYPE;size:80" json:"displayType"` // list_button,list_menu_item,obj_button,obj_menu_item,tab_button
	ActionType  string `gorm:"column:ACTION_TYPE;size:255" json:"actionType"`   // url,sp,job,js,bsh,py,go
	Content     string `gorm:"column:CONTENT;size:255" json:"content"`
	Scripts     string `gorm:"column:SCRIPTS;size:2000" json:"scripts"`
	URLTarget   string `gorm:"column:URLTARGET;size:255" json:"urlTarget"`
	SaveObj     string `gorm:"column:SAVE_OBJ;size:80" json:"saveObj"`
	Comments    string `gorm:"column:COMMENTS;size:255" json:"comments"`
	Filter      string `gorm:"column:FILTER;size:255" json:"filter"`
	Orderno     int    `gorm:"column:ORDERNO" json:"orderno"`
}

// TableName 指定表名
func (SysAction) TableName() string {
	return "sys_action"
}
