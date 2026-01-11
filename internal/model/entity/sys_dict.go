package entity

// SysDict 数据字典
type SysDict struct {
	BaseModel
	Name         string `gorm:"column:NAME;size:255;uniqueIndex;not null" json:"name"`
	DisplayName  string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	Type         int    `gorm:"column:TYPE" json:"type"` // 0:String, 1:int
	DefaultValue string `gorm:"column:DEFAULT_VALUE;size:255" json:"defaultValue"`
	Description  string `gorm:"column:DESCRIPTION;size:2000" json:"description"`
}

// TableName 指定表名
func (SysDict) TableName() string {
	return "sys_dict"
}

// SysDictItem 数据字典明细
type SysDictItem struct {
	BaseModel
	SysDictID      uint   `gorm:"column:SYS_DICT_ID;index:idx_dict_item_dict;not null" json:"sysDictId"`
	DisplayName    string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	Value          string `gorm:"column:VALUE;size:255" json:"value"`
	Orderno        int    `gorm:"column:ORDERNO" json:"orderno"`
	CssClass       string `gorm:"column:CSSCLASS;size:255" json:"cssClass"`
	IsDefaultValue string `gorm:"column:IS_DEFAULT_VALUE;size:1" json:"isDefaultValue"` // Y/N
}

// TableName 指定表名
func (SysDictItem) TableName() string {
	return "sys_dict_item"
}
