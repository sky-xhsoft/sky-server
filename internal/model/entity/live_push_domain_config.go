package entity

import "time"

// LivePushDomainConfig 直播推流域名配置
type LivePushDomainConfig struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID;not null;index" json:"sysCompanyId"`
	DomainName   string    `gorm:"column:DOMAIN_NAME;type:varchar(255);not null;uniqueIndex:uk_company_domain" json:"domainName"`
	StreamKey    string    `gorm:"column:STREAM_KEY;type:varchar(255);not null" json:"streamKey"`
	AppName      string    `gorm:"column:APP_NAME;type:varchar(100);not null;default:live" json:"appName"`
	PlayDomain   string    `gorm:"column:PLAY_DOMAIN;type:varchar(255)" json:"playDomain"`
	IsDefault    bool      `gorm:"column:IS_DEFAULT;type:tinyint(1);not null;default:0" json:"isDefault"`
	IsActive     bool      `gorm:"column:IS_ACTIVE;type:tinyint(1);not null;default:1" json:"isActive"`
	Remark       string    `gorm:"column:REMARK;type:varchar(500)" json:"remark"`
	CreatedAt    time.Time `gorm:"column:CREATED_AT;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:UPDATED_AT;not null;autoUpdateTime" json:"updatedAt"`
	CreatedBy    *uint     `gorm:"column:CREATED_BY" json:"createdBy"`
	UpdatedBy    *uint     `gorm:"column:UPDATED_BY" json:"updatedBy"`
}

// TableName 指定表名
func (LivePushDomainConfig) TableName() string {
	return "live_push_domain_config"
}
