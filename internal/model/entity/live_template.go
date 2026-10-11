package entity

import "time"

// 模板与规则的同步状态
const (
	TemplateSyncPending = "pending" // 未同步：本地有，云端还没有
	TemplateSyncSynced  = "synced"  // 已同步：本地与云端一致
	TemplateSyncFailed  = "failed"  // 同步失败：SYNC_ERROR 记录原因
	TemplateSyncDrift   = "drift"   // 云端不一致：有人在腾讯云控制台改过
)

// 模板来源
const (
	TemplateSourceLocal    = "local"    // 本系统创建
	TemplateSourceImport   = "import"   // 从腾讯云导入
	TemplateSourceStandard = "standard" // 标准配置自动创建
)

// 规则类型
const (
	TemplateKindCallback = "callback"
	TemplateKindRecord   = "record"
)

// TemplateSync 模板与规则共用的同步字段
type TemplateSync struct {
	SyncStatus string     `gorm:"column:SYNC_STATUS;type:varchar(20);not null;default:'pending'" json:"syncStatus"`
	SyncError  string     `gorm:"column:SYNC_ERROR;type:varchar(500)" json:"syncError"`
	SyncTime   *time.Time `gorm:"column:SYNC_TIME;type:datetime" json:"syncTime"`
}

// LiveCallbackTemplate 直播回调模板（本地为准，保存后同步到公司的腾讯云账号）
type LiveCallbackTemplate struct {
	ID                    uint   `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID          uint   `gorm:"column:SYS_COMPANY_ID;not null;default:0;index" json:"sysCompanyId"`
	Name                  string `gorm:"column:NAME;type:varchar(255);not null" json:"name"`
	Description           string `gorm:"column:DESCRIPTION;type:varchar(1024)" json:"description"`
	StreamBeginURL        string `gorm:"column:STREAM_BEGIN_URL;type:varchar(500)" json:"streamBeginUrl"`
	StreamEndURL          string `gorm:"column:STREAM_END_URL;type:varchar(500)" json:"streamEndUrl"`
	RecordURL             string `gorm:"column:RECORD_URL;type:varchar(500)" json:"recordUrl"`
	RecordStatusURL       string `gorm:"column:RECORD_STATUS_URL;type:varchar(500)" json:"recordStatusUrl"`
	RecordExceptionURL    string `gorm:"column:RECORD_EXCEPTION_URL;type:varchar(500)" json:"recordExceptionUrl"`
	RecordExceptionLevels string `gorm:"column:RECORD_EXCEPTION_LEVELS;type:varchar(100)" json:"recordExceptionLevels"` // 逗号分隔：error,alarm
	PushExceptionURL      string `gorm:"column:PUSH_EXCEPTION_URL;type:varchar(500)" json:"pushExceptionUrl"`
	CallbackKey           string `gorm:"column:CALLBACK_KEY;type:varchar(500);serializer:encrypted" json:"callbackKey"`
	CloudTemplateID       int64  `gorm:"column:CLOUD_TEMPLATE_ID;not null;default:0" json:"cloudTemplateId"`
	Source                string `gorm:"column:SOURCE;type:varchar(20);not null;default:'local'" json:"source"`
	TemplateSync
	CreateBy   string    `gorm:"column:CREATE_BY;type:varchar(255)" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;type:datetime;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;type:varchar(255)" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;type:datetime;autoUpdateTime" json:"updateTime"`
	IsActive   string    `gorm:"column:IS_ACTIVE;type:char(1);default:'Y'" json:"isActive"`
}

// TableName 指定表名
func (LiveCallbackTemplate) TableName() string { return "live_callback_template" }

// LiveRecordTemplate 直播录制模板（本系统创建的为存储至 COS 的 MP4 录制）
type LiveRecordTemplate struct {
	ID              uint   `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID    uint   `gorm:"column:SYS_COMPANY_ID;not null;default:0;index" json:"sysCompanyId"`
	Name            string `gorm:"column:NAME;type:varchar(255);not null" json:"name"`
	Description     string `gorm:"column:DESCRIPTION;type:varchar(1024)" json:"description"`
	CosStore        string `gorm:"column:COS_STORE;type:char(1);not null;default:'Y'" json:"cosStore"` // Y：COS；N：云点播 VOD（仅导入）
	CosBucket       string `gorm:"column:COS_BUCKET;type:varchar(255)" json:"cosBucket"`
	CosRegion       string `gorm:"column:COS_REGION;type:varchar(50)" json:"cosRegion"`
	CosPath         string `gorm:"column:COS_PATH;type:varchar(500)" json:"cosPath"`
	RecordInterval  int64  `gorm:"column:RECORD_INTERVAL;not null;default:1800" json:"recordInterval"` // 分段时长（秒）
	RemoveWatermark string `gorm:"column:REMOVE_WATERMARK;type:char(1);not null;default:'Y'" json:"removeWatermark"`
	Formats         string `gorm:"column:FORMATS;type:varchar(50);default:'mp4'" json:"formats"` // 逗号分隔
	CloudTemplateID int64  `gorm:"column:CLOUD_TEMPLATE_ID;not null;default:0" json:"cloudTemplateId"`
	Source          string `gorm:"column:SOURCE;type:varchar(20);not null;default:'local'" json:"source"`
	TemplateSync
	CreateBy   string    `gorm:"column:CREATE_BY;type:varchar(255)" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;type:datetime;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;type:varchar(255)" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;type:datetime;autoUpdateTime" json:"updateTime"`
	IsActive   string    `gorm:"column:IS_ACTIVE;type:char(1);default:'Y'" json:"isActive"`
}

// TableName 指定表名
func (LiveRecordTemplate) TableName() string { return "live_record_template" }

// LiveTemplateRule 模板绑定规则：回调规则为 域名 + AppName 唯一，录制规则为 域名 + AppName + StreamName 唯一
type LiveTemplateRule struct {
	ID           uint   `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint   `gorm:"column:SYS_COMPANY_ID;not null;default:0;index" json:"sysCompanyId"`
	Kind         string `gorm:"column:KIND;type:varchar(20);not null" json:"kind"` // callback / record
	DomainName   string `gorm:"column:DOMAIN_NAME;type:varchar(128);not null" json:"domainName"`
	AppName      string `gorm:"column:APP_NAME;type:varchar(64);not null;default:''" json:"appName"`
	StreamName   string `gorm:"column:STREAM_NAME;type:varchar(128);not null;default:''" json:"streamName"`
	TemplateID   uint   `gorm:"column:TEMPLATE_ID;not null" json:"templateId"` // 本地模板 ID
	TemplateSync
	CreateBy   string    `gorm:"column:CREATE_BY;type:varchar(255)" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;type:datetime;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;type:varchar(255)" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;type:datetime;autoUpdateTime" json:"updateTime"`
}

// TableName 指定表名
func (LiveTemplateRule) TableName() string { return "live_template_rule" }
