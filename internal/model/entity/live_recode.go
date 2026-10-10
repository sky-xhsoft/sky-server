package entity

import "time"

// LiveRecode 直播录制文件
type LiveRecode struct {
	ID           int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID int64     `gorm:"column:SYS_COMPANY_ID;not null;index" json:"sysCompanyId"`
	RoomID       *int64    `gorm:"column:ROOM_ID;index" json:"roomId"`
	StreamID     string    `gorm:"column:STREAM_ID;type:varchar(255)" json:"streamId"`
	RecordURL    string    `gorm:"column:RECORD_URL;type:varchar(500)" json:"recordUrl"`
	RecordType   string    `gorm:"column:RECORD_TYPE;type:varchar(50);default:'flv'" json:"recordType"`
	FileSize     int64     `gorm:"column:FILE_SIZE;default:0" json:"fileSize"`
	Duration     int       `gorm:"column:DURATION;default:0" json:"duration"`
	StartTime    int64     `gorm:"column:START_TIME;default:0" json:"startTime"`
	EndTime      int64     `gorm:"column:END_TIME;default:0" json:"endTime"`
	Status       string    `gorm:"column:STATUS;type:varchar(50);default:'pending'" json:"status"`
	Description  string    `gorm:"column:DESCRIPTION;type:varchar(500)" json:"description"`
	CreateBy     string    `gorm:"column:CREATE_BY;type:varchar(80)" json:"createBy"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME;not null;autoCreateTime" json:"createTime"`
	UpdateBy     string    `gorm:"column:UPDATE_BY;type:varchar(80)" json:"updateBy"`
	UpdateTime   time.Time `gorm:"column:UPDATE_TIME;not null;autoUpdateTime" json:"updateTime"`
	IsActive     string    `gorm:"column:IS_ACTIVE;type:char(1);default:'Y'" json:"isActive"`

	RoomName string `gorm:"-" json:"roomName"` // 直播间名称（查询列表时按 ROOM_ID 填充，不落库）
}

// TableName 指定表名
func (LiveRecode) TableName() string {
	return "live_recode"
}
