package entity

// SysGroups 权限组
type SysGroups struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255;not null" json:"name"`
	Description string `gorm:"column:DESCRIPTION;size:255" json:"description"`
	Sgrade      int    `gorm:"column:SGRADE" json:"sgrade"`
}

// TableName 指定表名
func (SysGroups) TableName() string {
	return "sys_groups"
}

// SysUserGroups 用户权限组关联
type SysUserGroups struct {
	BaseModel
	SysUserID      uint `gorm:"column:SYS_USER_ID;index:idx_user_groups;not null" json:"sysUserId"`
	SysDirectoryID uint `gorm:"column:SYS_DIRECTORY_ID;index:idx_user_groups;not null" json:"sysDirectoryId"`
}

// TableName 指定表名
func (SysUserGroups) TableName() string {
	return "sys_user_groups"
}

// SysDirectory 安全目录
type SysDirectory struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255;not null" json:"name"`
	SysTableID  *uint  `gorm:"column:SYS_TABLE_ID;index" json:"sysTableId"`
	ParentID    *uint  `gorm:"column:PARENT_ID;index" json:"parentId"`
	Orderno     int    `gorm:"column:ORDERNO" json:"orderno"`
	Description string `gorm:"column:DESCRIPTION;size:255" json:"description"`
}

// TableName 指定表名
func (SysDirectory) TableName() string {
	return "sys_directory"
}

// SysGroupPrem 权限组明细
type SysGroupPrem struct {
	BaseModel
	SysGroupsID    uint   `gorm:"column:SYS_GROUPS_ID;index;not null" json:"sysGroupsId"`
	SysDirectoryID uint   `gorm:"column:SYS_DIRECTORY_ID;index;not null" json:"sysDirectoryId"`
	Permission     int    `gorm:"column:PERMISSION;not null" json:"permission"` // 权限值（位运算）
	FilterObj      string `gorm:"column:FILTER_OBJ;size:255" json:"filterObj"`  // 数据过滤条件（JSON）
}

// TableName 指定表名
func (SysGroupPrem) TableName() string {
	return "sys_group_prem"
}

// SysCompany 公司（多租户）
type SysCompany struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255;not null" json:"name"`
	Code        string `gorm:"column:CODE;size:50;uniqueIndex" json:"code"`
	Description string `gorm:"column:DESCRIPTION;size:500" json:"description"`
	Status      string `gorm:"column:STATUS;size:1;default:Y" json:"status"` // Y:启用, N:禁用
}

// TableName 指定表名
func (SysCompany) TableName() string {
	return "sys_company"
}
