package entity

import "time"

// SysTranslation 翻译配置
type SysTranslation struct {
	BaseModel
	Key        string    `gorm:"column:KEY_NAME;size:255;not null" json:"key"`      // 翻译键（库表列名为 KEY_NAME）
	Language   string    `gorm:"column:LANGUAGE;size:50;not null" json:"language"`  // 语言代码
	Value      string    `gorm:"column:VALUE;type:text" json:"value"`               // 翻译值
	Module     string    `gorm:"column:MODULE;size:50" json:"module"`               // 所属模块
	IsActive   string    `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"` // Y/N
	CreateBy   string    `gorm:"column:CREATE_BY;size:80" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;size:80" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (SysTranslation) TableName() string {
	return "sys_translation"
}
