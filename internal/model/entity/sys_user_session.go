package entity

import "time"

// SysUserSession 用户会话记录（支持SSO和多设备管理）
type SysUserSession struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint      `gorm:"column:USER_ID;index:idx_session_user;not null" json:"userId"`
	CompanyID      uint      `gorm:"column:COMPANY_ID;not null" json:"companyId"`
	Token          string    `gorm:"column:TOKEN;size:500;index:idx_session_token" json:"token"`
	RefreshToken   string    `gorm:"column:REFRESH_TOKEN;size:500" json:"refreshToken"`
	ClientType     string    `gorm:"column:CLIENT_TYPE;size:20" json:"clientType"` // web, mobile, desktop
	DeviceID       string    `gorm:"column:DEVICE_ID;size:255;uniqueIndex:idx_session_device" json:"deviceId"`
	DeviceName     string    `gorm:"column:DEVICE_NAME;size:255" json:"deviceName"`
	IPAddress      string    `gorm:"column:IP_ADDRESS;size:50" json:"ipAddress"`
	UserAgent      string    `gorm:"column:USER_AGENT;size:500" json:"userAgent"`
	LoginTime      time.Time `gorm:"column:LOGIN_TIME;not null" json:"loginTime"`
	LastActiveTime time.Time `gorm:"column:LAST_ACTIVE_TIME" json:"lastActiveTime"`
	ExpireTime     time.Time `gorm:"column:EXPIRE_TIME" json:"expireTime"`
	IsActive       string    `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"` // Y/N
}

// TableName 指定表名
func (SysUserSession) TableName() string {
	return "sys_user_session"
}
