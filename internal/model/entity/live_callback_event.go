package entity

import (
	"time"
)

// LiveCallbackEvent 直播回调事件
type LiveCallbackEvent struct {
	ID           int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	EventType    string    `gorm:"column:EVENT_TYPE;type:varchar(50);not null;index" json:"eventType"` // 事件类型：push_stream, disconnect_stream, recording_file, recording_status
	EventTime    int64     `gorm:"column:EVENT_TIME;not null" json:"eventTime"`                        // 事件时间戳（秒）
	DomainName   string    `gorm:"column:DOMAIN_NAME;type:varchar(255);index" json:"domainName"`       // 推流域名
	AppName      string    `gorm:"column:APP_NAME;type:varchar(255);index" json:"appName"`             // 应用名称
	StreamName   string    `gorm:"column:STREAM_NAME;type:varchar(255);index" json:"streamName"`       // 流名称
	StreamID     string    `gorm:"column:STREAM_ID;type:varchar(255);index" json:"streamId"`           // 流ID
	ClientIP     string    `gorm:"column:CLIENT_IP;type:varchar(50)" json:"clientIp"`                  // 客户端IP
	EventData    string    `gorm:"column:EVENT_DATA;type:text" json:"eventData"`                       // 事件详细数据（JSON格式）
	Sign         string    `gorm:"column:SIGN;type:varchar(255)" json:"sign"`                          // 签名
	TValue       int64     `gorm:"column:T_VALUE;not null" json:"tValue"`                              // 签名过期时间
	CreateTime   time.Time `gorm:"column:CREATE_TIME;type:datetime;not null;autoCreateTime" json:"createTime"`
	SysCompanyID int64     `gorm:"column:SYS_COMPANY_ID;not null;index" json:"sysCompanyId"`
	IsActive     string    `gorm:"column:IS_ACTIVE;type:char(1);default:'Y'" json:"isActive"`
	RoomName     string    `gorm:"column:ROOM_NAME;type:varchar(255)" json:"roomName"` // 直播间名称
}

// TableName 指定表名
func (LiveCallbackEvent) TableName() string {
	return "live_callback_event"
}
