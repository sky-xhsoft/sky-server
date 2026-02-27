package entity

import (
	"time"
)

// PullStreamTask 拉流任务实体类
type PullStreamTask struct {
	ID         int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	TaskID     string    `gorm:"column:TASK_ID;type:varchar(255);uniqueIndex;not null" json:"taskId"`
	Comment    string    `gorm:"column:COMMENT;type:varchar(255)" json:"comment"`
	Region     string    `gorm:"column:REGION;type:varchar(255);not null" json:"region"`
	SourceType string    `gorm:"column:SOURCE_TYPE;type:varchar(255);not null" json:"sourceType"`
	SourceURL  string    `gorm:"column:SOURCE_URL;type:varchar(255);not null" json:"sourceUrl"`
	TargetURL  string    `gorm:"column:TARGET_URL;type:varchar(255);not null" json:"targetUrl"`
	StartTime  time.Time `gorm:"column:START_TIME;type:datetime;not null" json:"startTime"`
	EndTime    time.Time `gorm:"column:END_TIME;type:datetime;not null" json:"endTime"`
	Status     string    `gorm:"column:STATUS;type:varchar(255);default:'enable'" json:"status"`
	Operator   string    `gorm:"column:OPERATOR;type:varchar(255);not null" json:"operator"`
	CreateBy   string    `gorm:"column:CREATE_BY;type:varchar(255)" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;type:datetime;not null;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;type:varchar(255)" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;type:datetime;not null;autoUpdateTime" json:"updateTime"`
	IsActive   string    `gorm:"column:IS_ACTIVE;type:char(1);default:'Y'" json:"isActive"`
	RoomID     string    `gorm:"column:ROOM_ID;type:varchar(255)" json:"roomId"`
	RoomName   string    `gorm:"column:ROOM_NAME;type:varchar(255)" json:"roomName"`
}

// TableName 指定表名
func (PullStreamTask) TableName() string {
	return "pull_stream_task"
}
