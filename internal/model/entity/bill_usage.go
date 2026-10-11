package entity

import "time"

// 计费一期：只计量不扣费。三个计费点：直播时长、播放流量、存储空间

// 计费点
const (
	MeterLiveDuration = "live_duration" // 直播时长，单位：分钟
	MeterPlayFlux     = "play_flux"     // 播放流量，单位：GB
	MeterStorage      = "storage"       // 存储空间，单位：GB（当日快照）
)

// BillLiveSession 推流段落：一次推流到断流，按直播间记录
type BillLiveSession struct {
	ID           uint       `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint       `gorm:"column:SYS_COMPANY_ID;not null;index:IDX_BLS_COMPANY_START,priority:1" json:"-"`
	RoomID       uint       `gorm:"column:ROOM_ID;not null;index" json:"roomId"`
	StreamName   string     `gorm:"column:STREAM_NAME;type:varchar(128);not null" json:"streamName"`
	PushDomain   string     `gorm:"column:PUSH_DOMAIN;type:varchar(255)" json:"pushDomain"`
	StartTime    time.Time  `gorm:"column:START_TIME;not null;index:IDX_BLS_COMPANY_START,priority:2" json:"startTime"`
	EndTime      *time.Time `gorm:"column:END_TIME" json:"endTime"`
	LastSeenTime time.Time  `gorm:"column:LAST_SEEN_TIME;not null" json:"lastSeenTime"` // 巡检最后确认在线的时间
	Seconds      int64      `gorm:"column:SECONDS;not null;default:0" json:"seconds"`
	EndSource    string     `gorm:"column:END_SOURCE;type:varchar(20)" json:"endSource"` // callback 断流回调 / patrol 巡检 / cap 超过最长时长
	PushEventID  int64      `gorm:"column:PUSH_EVENT_ID;not null;uniqueIndex:UK_BLS_PUSH_EVENT" json:"-"`
	CreateTime   time.Time  `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time  `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"-"`
}

func (BillLiveSession) TableName() string { return "bill_live_session" }

// BillUsageDaily 日用量：每个公司、直播间（存储为 0）、计费点、细项一行，按来源键幂等写入
type BillUsageDaily struct {
	ID           uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID;not null;index:IDX_BUD_COMPANY_DATE,priority:1" json:"-"`
	UsageDate    string    `gorm:"column:USAGE_DATE;type:char(10);not null;index:IDX_BUD_COMPANY_DATE,priority:2" json:"usageDate"`
	Meter        string    `gorm:"column:METER;type:varchar(32);not null" json:"meter"`
	Item         string    `gorm:"column:ITEM;type:varchar(32);not null;default:''" json:"item"` // 存储细项：record 录制 / clip 切片 / disk 云盘
	RoomID       uint      `gorm:"column:ROOM_ID;not null;default:0" json:"roomId"`
	Quantity     float64   `gorm:"column:QUANTITY;type:decimal(20,6);not null" json:"quantity"`
	Unit         string    `gorm:"column:UNIT;type:varchar(16);not null" json:"unit"`
	RawQuantity  float64   `gorm:"column:RAW_QUANTITY;type:decimal(20,6);not null;default:0" json:"rawQuantity"` // 流量：分摊前的监控流量（GB）；时长：秒
	SourceKey    string    `gorm:"column:SOURCE_KEY;type:varchar(191);not null;uniqueIndex:UK_BUD_SOURCE" json:"-"`
	UpdateTime   time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillUsageDaily) TableName() string { return "bill_usage_daily" }

// BillFluxReconcile 播放流量对账：某个腾讯云账号下某个播放域名某天的监控流量与计费流量
type BillFluxReconcile struct {
	ID             uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UsageDate      string    `gorm:"column:USAGE_DATE;type:char(10);not null;uniqueIndex:UK_BFR_DAY,priority:1" json:"usageDate"`
	Account        string    `gorm:"column:ACCOUNT;type:varchar(64);not null;uniqueIndex:UK_BFR_DAY,priority:2" json:"account"` // global 平台账号 / tc:<密钥 ID 摘要> 公司自有账号
	PlayDomain     string    `gorm:"column:PLAY_DOMAIN;type:varchar(255);not null;uniqueIndex:UK_BFR_DAY,priority:3" json:"playDomain"`
	MonitorMB      float64   `gorm:"column:MONITOR_MB;type:decimal(20,3);not null" json:"monitorMb"`
	BillingMB      float64   `gorm:"column:BILLING_MB;type:decimal(20,3);not null" json:"billingMb"`
	UnattributedMB float64   `gorm:"column:UNATTRIBUTED_MB;type:decimal(20,3);not null" json:"unattributedMb"` // 找不到直播间的流（分摊后）
	StreamCount    int       `gorm:"column:STREAM_COUNT;not null" json:"streamCount"`
	Status         string    `gorm:"column:STATUS;type:varchar(20);not null" json:"status"` // ok / diff 监控与计费差异超过阈值 / error 拉取失败
	Message        string    `gorm:"column:MESSAGE;type:varchar(500)" json:"message"`
	UpdateTime     time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillFluxReconcile) TableName() string { return "bill_flux_reconcile" }

// BillMeterState 计量任务状态（游标、每日任务完成标记）
type BillMeterState struct {
	StateKey   string    `gorm:"column:STATE_KEY;type:varchar(64);primaryKey" json:"stateKey"`
	StateValue string    `gorm:"column:STATE_VALUE;type:varchar(255);not null" json:"stateValue"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillMeterState) TableName() string { return "bill_meter_state" }
