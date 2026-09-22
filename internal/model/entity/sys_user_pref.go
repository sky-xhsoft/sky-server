package entity

// SysUserPref 用户环境变量(用户 × 模块 × 键)
type SysUserPref struct {
	BaseModel
	SysUserID   uint   `gorm:"column:SYS_USER_ID;not null;default:0;index:idx_user_module_name,unique,priority:1" json:"sysUserId"` // 用户(0=全局/系统级)
	Module      string `gorm:"column:MODULE;size:64;not null;default:;index:idx_user_module_name,unique,priority:2" json:"module"`
	Name        string `gorm:"column:NAME;size:128;not null;default:;index:idx_user_module_name,unique,priority:3" json:"name"`
	Value       string `gorm:"column:VALUE;type:text" json:"value"`
	Description string `gorm:"column:DESCRIPTION;size:255" json:"description"`
}

// TableName 指定表名
func (SysUserPref) TableName() string {
	return "sys_user_pref"
}