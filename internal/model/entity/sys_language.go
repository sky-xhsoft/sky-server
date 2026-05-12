package entity

import "time"

// SysLanguage 语言配置
type SysLanguage struct {
	BaseModel
	Code       string    `gorm:"column:CODE;size:50;not null" json:"code"`            // 语言代码 zh-CN, en-US 等
	Name       string    `gorm:"column:NAME;size:255;not null" json:"name"`           // 语言名称
	IsDefault  string    `gorm:"column:IS_DEFAULT;size:1;default:N" json:"isDefault"` // Y/N
	IsActive   string    `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"`   // Y/N
	CreateBy   string    `gorm:"column:CREATE_BY;size:80" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;size:80" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (SysLanguage) TableName() string {
	return "sys_language"
}
