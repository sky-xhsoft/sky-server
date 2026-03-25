package entity

import "time"

// LiveRoom 直播间
type LiveRoom struct {
	BaseModel
	RoomName         string     `gorm:"column:ROOM_NAME;type:varchar(255);not null" json:"roomName"`                               // 直播间名称
	RoomType         string     `gorm:"column:ROOM_TYPE;type:varchar(50);not null;index" json:"roomType"`                          // 直播间类型：video(视频直播), image(图片直播), vr(VR直播), audio(语音直播), graphic(图文直播)
	BroadcastFormat  string     `gorm:"column:BROADCAST_FORMAT;type:varchar(50);not null" json:"broadcastFormat"`                  // 播出形式：live(直播), vod(点播/录播), pseudo(伪直播)
	RoomStage        string     `gorm:"column:ROOM_STAGE;type:varchar(50);not null" json:"roomStage"`                              // 直播间阶段：formal(正式直播), test(测试直播)
	DisplayMode      string     `gorm:"column:DISPLAY_MODE;type:varchar(50)" json:"displayMode"`                                   // 显示方式：landscape(横屏), portrait(竖屏), three_screen(三分屏)
	StartTime        *time.Time `gorm:"column:START_TIME;type:datetime" json:"startTime"`                                          // 开始时间
	EndTime          *time.Time `gorm:"column:END_TIME;type:datetime" json:"endTime"`                                              // 结束时间
	CoverImage       string     `gorm:"column:COVER_IMAGE;type:varchar(500)" json:"coverImage"`                                    // 直播间封面
	ViewingMethod    string     `gorm:"column:VIEWING_METHOD;type:varchar(50);not null;default:'public'" json:"viewingMethod"`     // 观看方式：public(公开), encrypted(加密), paid(付费), ticket(购票进入), enterprise(企业成员观看), custom(自建成员观看)
	ViewingPassword  string     `gorm:"column:VIEWING_PASSWORD;type:varchar(255)" json:"viewingPassword"`                          // 观看密码（加密观看时使用）
	ViewingPrice     *float64   `gorm:"column:VIEWING_PRICE;type:decimal(10,2)" json:"viewingPrice"`                               // 观看价格（付费观看时使用）
	PlaybackMethod   string     `gorm:"column:PLAYBACK_METHOD;type:varchar(50);not null;default:'post_end'" json:"playbackMethod"` // 回放方式：post_end(结束后回放), real_time(实时回放), no_playback(结束后不回放)
	PlaybackValidity string     `gorm:"column:PLAYBACK_VALIDITY;type:varchar(50);default:'unlimited'" json:"playbackValidity"`     // 回放有效期：unlimited(无限制), all_day(全天), partial(部分时段)
	StreamName       string     `gorm:"column:STREAM_NAME;type:varchar(255);index" json:"streamName"`                              // 流名称
	PushURL          string     `gorm:"column:PUSH_URL;type:varchar(500)" json:"pushUrl"`                                          // 推流地址
	PlayURL          string     `gorm:"column:PLAY_URL;type:varchar(500)" json:"playUrl"`                                          // 播放地址
	Status           string     `gorm:"column:STATUS;type:varchar(50);not null;default:'draft';index" json:"status"`               // 状态：draft(草稿), scheduled(已排期), live(直播中), ended(已结束), archived(已归档)
	ViewerCount      int        `gorm:"column:VIEWER_COUNT;default:0" json:"viewerCount"`                                          // 观看人数
	PeakViewerCount  int        `gorm:"column:PEAK_VIEWER_COUNT;default:0" json:"peakViewerCount"`                                 // 峰值观看人数
	Duration         int        `gorm:"column:DURATION;default:0" json:"duration"`                                                 // 直播时长（秒）
	Description      string     `gorm:"column:DESCRIPTION;type:text" json:"description"`                                           // 直播间描述
	Props            string     `gorm:"column:PROPS;type:text" json:"props"`                                                       // 扩展属性（JSON）
}

// TableName 指定表名
func (LiveRoom) TableName() string {
	return "live_room"
}
