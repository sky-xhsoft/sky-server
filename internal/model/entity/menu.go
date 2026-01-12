package entity

// MenuNode 菜单节点（通用结构，支持三级）
type MenuNode struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName,omitempty"` // sys_table 使用
	Icon        string      `json:"icon,omitempty"`
	URL         string      `json:"url,omitempty"`
	OrderNo     int         `json:"orderno"`
	Type        string      `json:"type"`        // subsystem, category, table
	Children    []*MenuNode `json:"children,omitempty"`
}

// SysSubsystem 子系统（一级菜单）
type SysSubsystem struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255" json:"name"`
	OrderNo     int    `gorm:"column:ORDERNO" json:"orderno"`
	URL         string `gorm:"column:URL;size:255" json:"url"`
	Icon        string `gorm:"column:ICON;size:255" json:"icon"`
	Description string `gorm:"column:DESCRIPTION;size:255" json:"description"`
}

// TableName 指定表名
func (SysSubsystem) TableName() string {
	return "sys_subsystem"
}

// SysTableCategory 表类别（二级菜单）
type SysTableCategory struct {
	BaseModel
	SysSubsystemID uint   `gorm:"column:SYS_SUBSYSTEM_ID;index:idx_subsystem" json:"sysSubsystemId"`
	Name           string `gorm:"column:NAME;size:255;not null" json:"name"`
	OrderNo        int    `gorm:"column:ORDERNO" json:"orderno"`
	Icon           string `gorm:"column:ICON;size:255" json:"icon"`
	URL            string `gorm:"column:URL;size:255" json:"url"`
	Description    string `gorm:"column:DESCRIPTION;size:255" json:"description"`
}

// TableName 指定表名
func (SysTableCategory) TableName() string {
	return "sys_table_category"
}
