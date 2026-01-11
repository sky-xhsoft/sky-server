package entity

// SysTableRef 表关联关系
type SysTableRef struct {
	BaseModel
	SysTableID   int    `gorm:"column:SYS_TABLE_ID;index;not null" json:"sysTableId"`
	RefTableID   int    `gorm:"column:REF_TABLE_ID;not null" json:"refTableId"`
	RefColumnID  int    `gorm:"column:REF_COLUMN_ID" json:"refColumnId"`
	DisplayName  string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	Filter       string `gorm:"column:FILTER;size:255" json:"filter"`
	AssocType    string `gorm:"column:ASSOCTYPE;size:1" json:"assocType"` // 1:1对1, n:1对n
	EditType     string `gorm:"column:EDIT_TYPE;size:2" json:"editType"`   // Y:标准,N:无,NP:非内嵌允许弹出,NS:非内嵌禁止弹出,A:仅显示新增字段
	Orderno      int    `gorm:"column:ORDERNO" json:"orderno"`
}

// TableName 指定表名
func (SysTableRef) TableName() string {
	return "sys_table_ref"
}
