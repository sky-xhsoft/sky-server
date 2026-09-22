package entity

import "time"

// LiveCut 直播精彩片段/切片
type LiveCut struct {
	ID           int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID int64     `gorm:"column:SYS_COMPANY_ID;not null;index" json:"sysCompanyId"`
	RoomID       *int64    `gorm:"column:ROOM_ID;index" json:"roomId"`
	StreamID     string    `gorm:"column:STREAM_ID;type:varchar(255)" json:"streamId"`
	ClipName     string    `gorm:"column:CLIP_NAME;type:varchar(255);not null" json:"clipName"`
	ClipURL      string    `gorm:"column:CLIP_URL;type:varchar(500)" json:"clipUrl"`
	StartTime    int64     `gorm:"column:START_TIME;default:0" json:"startTime"`
	EndTime      int64     `gorm:"column:END_TIME;default:0" json:"endTime"`
	Duration     int       `gorm:"column:DURATION;default:0" json:"duration"`
	Thumbnail    string    `gorm:"column:THUMBNAIL;type:varchar(500)" json:"thumbnail"`
	FileSize     int64     `gorm:"column:FILE_SIZE;default:0" json:"fileSize"`
	Status       string    `gorm:"column:STATUS;type:varchar(50);default:'processing'" json:"status"`
	Description  string    `gorm:"column:DESCRIPTION;type:varchar(500)" json:"description"`
	CreateBy     string    `gorm:"column:CREATE_BY;type:varchar(80)" json:"createBy"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME;not null;autoCreateTime" json:"createTime"`
	UpdateBy     string    `gorm:"column:UPDATE_BY;type:varchar(80)" json:"updateBy"`
	UpdateTime   time.Time `gorm:"column:UPDATE_TIME;not null;autoUpdateTime" json:"updateTime"`
	IsActive     string    `gorm:"column:IS_ACTIVE;type:char(1);default:'Y'" json:"isActive"`
}

// TableName 指定表名
func (LiveCut) TableName() string {
	return "live_cut"
}
