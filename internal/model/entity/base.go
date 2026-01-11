package entity

import "time"

// BaseModel 基础模型（包含标准审计字段）
type BaseModel struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SysCompanyID   uint      `gorm:"column:SYS_COMPANY_ID;index" json:"sysCompanyId"`
	CreateBy       string    `gorm:"column:CREATE_BY;size:80" json:"createBy"`
	CreateTime     time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateBy       string    `gorm:"column:UPDATE_BY;size:80" json:"updateBy"`
	UpdateTime     time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
	IsActive       string    `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"` // Y:有效, N:无效
}

// TableName 默认表名（子类可覆盖）
func (BaseModel) TableName() string {
	return ""
}
