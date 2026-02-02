package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LiveCallbackHandler 直播回调处理器
type LiveCallbackHandler struct {
	db        *gorm.DB
	callbackKey string // 回调密钥，用于验证签名
}

// NewLiveCallbackHandler 创建直播回调处理器
func NewLiveCallbackHandler(db *gorm.DB, callbackKey string) *LiveCallbackHandler {
	return &LiveCallbackHandler{
		db:          db,
		callbackKey: callbackKey,
	}
}

// PushStreamEvent 推流事件通知
type PushStreamEvent struct {
	EventType  int    `json:"event_type"`  // 事件类型：1-推流
	StreamID   string `json:"stream_id"`   // 流ID
	ChannelID  string `json:"channel_id"`  // 频道ID
	T          int64  `json:"t"`           // 过期时间
	Sign       string `json:"sign"`        // 签名
	EventTime  int64  `json:"event_time"`  // 事件时间
	Sequence   string `json:"sequence"`    // 序列号
	Node       string `json:"node"`        // 接入点IP
	UserIP     string `json:"user_ip"`     // 用户IP
	StreamParam string `json:"stream_param"` // 推流参数
	PushDomain string `json:"push_domain"` // 推流域名
	AppName    string `json:"app_name"`    // 应用名称
	StreamName string `json:"stream_name"` // 流名称
}

// DisconnectStreamEvent 断流事件通知
type DisconnectStreamEvent struct {
	EventType  int    `json:"event_type"`  // 事件类型：0-断流
	StreamID   string `json:"stream_id"`   // 流ID
	ChannelID  string `json:"channel_id"`  // 频道ID
	T          int64  `json:"t"`           // 过期时间
	Sign       string `json:"sign"`        // 签名
	EventTime  int64  `json:"event_time"`  // 事件时间
	Sequence   string `json:"sequence"`    // 序列号
	Node       string `json:"node"`        // 接入点IP
	UserIP     string `json:"user_ip"`     // 用户IP
	StreamParam string `json:"stream_param"` // 推流参数
	Duration   int64  `json:"duration"`    // 推流时长（秒）
	Reason     string `json:"reason"`      // 断流原因
	PushDomain string `json:"push_domain"` // 推流域名
	AppName    string `json:"app_name"`    // 应用名称
	StreamName string `json:"stream_name"` // 流名称
}

// RecordingFileEvent 录制文件事件通知
type RecordingFileEvent struct {
	EventType    int    `json:"event_type"`     // 事件类型：100-录制文件生成
	StreamID     string `json:"stream_id"`      // 流ID
	ChannelID    string `json:"channel_id"`     // 频道ID
	T            int64  `json:"t"`              // 过期时间
	Sign         string `json:"sign"`           // 签名
	EventTime    int64  `json:"event_time"`     // 事件时间
	VideoURL     string `json:"video_url"`      // 录制文件下载URL
	FileSize     int64  `json:"file_size"`      // 文件大小（字节）
	Duration     int64  `json:"duration"`       // 录制时长（秒）
	FileFormat   string `json:"file_format"`    // 文件格式：flv, hls, mp4, aac
	StartTime    int64  `json:"start_time"`     // 录制开始时间
	EndTime      int64  `json:"end_time"`       // 录制结束时间
	StreamParam  string `json:"stream_param"`   // 推流参数
	VideoID      string `json:"video_id"`       // 点播文件ID
	RecordFileID string `json:"record_file_id"` // 录制文件ID

	// 实际的腾讯云字段（优先使用）
	App     string `json:"app"`     // 推流域名（实际字段）
	AppName string `json:"appname"` // 应用名称（实际字段）

	// 文档中的字段（向后兼容）
	PushDomain     string `json:"push_domain"`  // 推流域名（文档字段，向后兼容）
	AppNameCompat  string `json:"app_name"`     // 应用名称（文档字段，向后兼容）
	StreamName     string `json:"stream_name"`  // 流名称（文档字段，向后兼容）
}

// RecordingStatusEvent 录制状态事件通知
type RecordingStatusEvent struct {
	EventType   int    `json:"event_type"`   // 事件类型：200-录制状态变更
	StreamID    string `json:"stream_id"`    // 流ID
	ChannelID   string `json:"channel_id"`   // 频道ID
	T           int64  `json:"t"`            // 过期时间
	Sign        string `json:"sign"`         // 签名
	EventTime   int64  `json:"event_time"`   // 事件时间
	Status      int    `json:"status"`       // 录制状态：0-未录制，1-录制中
	StreamParam string `json:"stream_param"` // 推流参数
	PushDomain  string `json:"push_domain"`  // 推流域名
	AppName     string `json:"app_name"`     // 应用名称
	StreamName  string `json:"stream_name"`  // 流名称
}

// ScreenshotEvent 截图事件通知
type ScreenshotEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：200-截图
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	PicURL      string   `json:"pic_url"`      // 截图URL
	PicFullURL  string   `json:"pic_full_url"` // 完整截图URL
	CreateTime  int64    `json:"create_time"`  // 截图生成时间
	StreamParam string   `json:"stream_param"` // 推流参数
	Width       int      `json:"width"`        // 图片宽度
	Height      int      `json:"height"`       // 图片高度
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// VideoAuditEvent 画面审核事件通知
type VideoAuditEvent struct {
	EventType      int      `json:"event_type"`      // 事件类型：317-画面审核
	StreamID       string   `json:"stream_id"`       // 流ID
	ChannelID      string   `json:"channel_id"`      // 频道ID
	T              int64    `json:"t"`               // 过期时间
	Sign           string   `json:"sign"`            // 签名
	EventTime      int64    `json:"event_time"`      // 事件时间
	Confidence     int      `json:"confidence"`      // 置信度
	Label          string   `json:"label"`           // 审核标签
	Suggestion     string   `json:"suggestion"`      // 建议：pass, review, block
	ScreenshotURL  string   `json:"screenshot_url"`  // 截图URL
	ScreenshotTime int64    `json:"screenshot_time"` // 截图时间
	SendTime       int64    `json:"send_time"`       // 发送时间
	StreamParam    string   `json:"stream_param"`    // 推流参数
	PushDomain     string   `json:"push_domain"`     // 推流域名
	AppName        string   `json:"app_name"`        // 应用名称
	StreamName     string   `json:"stream_name"`     // 流名称
}

// AudioAuditEvent 音频审核事件通知
type AudioAuditEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：318-音频审核
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	Confidence  int      `json:"confidence"`   // 置信度
	Label       string   `json:"label"`        // 审核标签
	Suggestion  string   `json:"suggestion"`   // 建议：pass, review, block
	AudioText   string   `json:"audio_text"`   // 音频转文字内容
	AudioTime   int64    `json:"audio_time"`   // 音频时间
	SendTime    int64    `json:"send_time"`    // 发送时间
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// QualityInspectionEvent 质检事件通知
type QualityInspectionEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：319-质检
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	DiagnoseType string  `json:"diagnose_type"` // 诊断类型
	Level       string   `json:"level"`        // 级别：warning, error
	Description string   `json:"description"`  // 描述
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// QualityThresholdEvent 评测阈值事件通知
type QualityThresholdEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：320-评测阈值
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	MetricType  string   `json:"metric_type"`  // 指标类型
	Threshold   float64  `json:"threshold"`    // 阈值
	CurrentValue float64 `json:"current_value"` // 当前值
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// QualityAverageEvent 评测平均分事件通知
type QualityAverageEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：321-评测平均分
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	Score       float64  `json:"score"`        // 平均分
	Duration    int64    `json:"duration"`     // 统计时长（秒）
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// SmartEraseEvent 智能擦除事件通知
type SmartEraseEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：322-智能擦除
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	TaskID      string   `json:"task_id"`      // 任务ID
	Status      string   `json:"status"`       // 状态：success, failed
	OutputURL   string   `json:"output_url"`   // 输出URL
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// SubtitleEvent 直播字幕事件通知
type SubtitleEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：323-直播字幕
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	Text        string   `json:"text"`         // 字幕文本
	Language    string   `json:"language"`     // 语言
	StartTime   int64    `json:"start_time"`   // 开始时间
	EndTime     int64    `json:"end_time"`     // 结束时间
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// SummaryEvent 直播摘要事件通知
type SummaryEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：324-直播摘要
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	Summary     string   `json:"summary"`      // 摘要内容
	Keywords    []string `json:"keywords"`     // 关键词
	Duration    int64    `json:"duration"`     // 时长（秒）
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// HighlightEvent 高光切片事件通知（实际格式）
type HighlightEvent struct {
	AppID     int64           `json:"appid"`      // 应用ID
	Domain    string          `json:"domain"`     // 域名
	EventType int             `json:"event_type"` // 事件类型：349-高光切片
	Items     []HighlightItem `json:"items"`      // 高光切片列表
	Path      string          `json:"path"`       // 路径（如：live）
	StreamID  string          `json:"stream_id"`  // 流ID
	T         int64           `json:"t"`          // 过期时间（可选）
	Sign      string          `json:"sign"`       // 签名（可选）
}

// HighlightItem 单个高光切片信息
type HighlightItem struct {
	BeginTime       int64    `json:"begin_time"`         // 开始时间（Unix时间戳）
	EndTime         int64    `json:"end_time"`           // 结束时间（Unix时间戳）
	CovImgStoreURL  string   `json:"cov_img_store_url"`  // 封面图片存储URL
	VideoStoreURL   string   `json:"video_store_url"`    // 视频存储URL
	Title           string   `json:"title"`              // 标题
	Summary         string   `json:"summary"`            // 摘要
	KeyWords        []string `json:"key_words"`          // 关键词列表
}

// PushExceptionEvent 推流异常事件通知
type PushExceptionEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：326-推流异常
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	ErrorCode   int      `json:"error_code"`   // 错误码
	ErrorMsg    string   `json:"error_msg"`    // 错误信息
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// RecordExceptionEvent 录制异常事件通知
type RecordExceptionEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：327-录制异常
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	ErrorCode   int      `json:"error_code"`   // 错误码
	ErrorMsg    string   `json:"error_msg"`    // 错误信息
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// PullStreamEvent 拉流转推事件通知
type PullStreamEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：328-拉流转推
	TaskID      string   `json:"task_id"`      // 任务ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	Status      string   `json:"status"`       // 状态：start, stop, error
	SourceURL   string   `json:"source_url"`   // 源URL
	TargetURL   string   `json:"target_url"`   // 目标URL
	ErrorCode   int      `json:"error_code"`   // 错误码
	ErrorMsg    string   `json:"error_msg"`    // 错误信息
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// MonitorEvent 监播事件通知
type MonitorEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：329-监播
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	AlertType   string   `json:"alert_type"`   // 告警类型
	AlertLevel  string   `json:"alert_level"`  // 告警级别：info, warning, error
	Description string   `json:"description"`  // 描述
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}

// verifySign 验证回调签名
func (h *LiveCallbackHandler) verifySign(t int64, sign string) bool {
	// 检查时间戳是否过期（默认10分钟有效期）
	now := time.Now().Unix()
	if now > t {
		logger.Warn("回调签名已过期",
			zap.Int64("t", t),
			zap.Int64("now", now))
		return false
	}

	// 计算签名：MD5(key + t)
	expected := fmt.Sprintf("%s%d", h.callbackKey, t)
	hash := md5.Sum([]byte(expected))
	expectedSign := hex.EncodeToString(hash[:])

	return expectedSign == sign
}

// HandlePushStream 处理推流事件
// @Summary 推流事件回调
// @Description 接收腾讯云推流事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body PushStreamEvent true "推流事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/push [post]
func (h *LiveCallbackHandler) HandlePushStream(c *gin.Context) {
	var event PushStreamEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析推流事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	// 验证签名
	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("推流事件签名验证失败",
			zap.String("streamId", event.StreamID),
			zap.String("sign", event.Sign))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	// 记录事件日志
	logger.Info("收到推流事件",
		zap.String("streamId", event.StreamID),
		zap.String("domain", event.PushDomain),
		zap.String("app", event.AppName),
		zap.String("stream", event.StreamName),
		zap.String("userIp", event.UserIP))

	// 保存事件到数据库
	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "push_stream",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		ClientIP:     event.UserIP,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1, // TODO: 从上下文获取公司ID
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存推流事件失败", zap.Error(err))
	}

	// 返回成功响应
	c.JSON(200, gin.H{"code": 0})
}

// HandleDisconnectStream 处理断流事件
// @Summary 断流事件回调
// @Description 接收腾讯云断流事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body DisconnectStreamEvent true "断流事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/disconnect [post]
func (h *LiveCallbackHandler) HandleDisconnectStream(c *gin.Context) {
	var event DisconnectStreamEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析断流事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	// 验证签名
	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("断流事件签名验证失败",
			zap.String("streamId", event.StreamID),
			zap.String("sign", event.Sign))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	// 记录事件日志
	logger.Info("收到断流事件",
		zap.String("streamId", event.StreamID),
		zap.String("domain", event.PushDomain),
		zap.String("app", event.AppName),
		zap.String("stream", event.StreamName),
		zap.Int64("duration", event.Duration),
		zap.String("reason", event.Reason))

	// 保存事件到数据库
	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "disconnect_stream",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		ClientIP:     event.UserIP,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1, // TODO: 从上下文获取公司ID
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存断流事件失败", zap.Error(err))
	}

	// 返回成功响应
	c.JSON(200, gin.H{"code": 0})
}

// HandleRecordingFile 处理录制文件事件
// @Summary 录制文件事件回调
// @Description 接收腾讯云录制文件生成通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body RecordingFileEvent true "录制文件事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/recording-file [post]
func (h *LiveCallbackHandler) HandleRecordingFile(c *gin.Context) {
	var event RecordingFileEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析录制文件事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	// 验证签名
	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("录制文件事件签名验证失败",
			zap.String("streamId", event.StreamID),
			zap.String("sign", event.Sign))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	// 记录事件日志
	logger.Info("收到录制文件事件",
		zap.String("streamId", event.StreamID),
		zap.String("domain", getDomainName(event)),
		zap.String("app", getAppName(event)),
		zap.String("stream", getStreamName(event)),
		zap.String("videoUrl", event.VideoURL),
		zap.Int64("fileSize", event.FileSize),
		zap.Int64("duration", event.Duration),
		zap.String("format", event.FileFormat),
		zap.Int64("eventTime", event.EventTime))

	// 确定事件时间：优先使用 event_time，如果为 0 则使用 end_time，最后使用当前时间
	eventTime := event.EventTime
	if eventTime == 0 {
		if event.EndTime > 0 {
			eventTime = event.EndTime
			logger.Info("event_time 为 0，使用 end_time 作为生成时间",
				zap.String("streamId", event.StreamID),
				zap.Int64("endTime", event.EndTime))
		} else {
			eventTime = time.Now().Unix()
			logger.Warn("event_time 和 end_time 都为 0，使用当前时间",
				zap.String("streamId", event.StreamID),
				zap.Int64("currentTime", eventTime))
		}
	}

	// 保存事件到数据库
	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "recording_file",
		EventTime:    eventTime,
		DomainName:   getDomainName(event),  // 优先使用 app 字段
		AppName:      getAppName(event),     // 优先使用 appname 字段
		StreamName:   getStreamName(event),  // 优先使用 stream_id 作为流名称
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1, // TODO: 从上下文获取公司ID
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存录制文件事件失败", zap.Error(err))
	}

	// 返回成功响应
	c.JSON(200, gin.H{"code": 0})
}

// HandleRecordingStatus 处理录制状态事件
// @Summary 录制状态事件回调
// @Description 接收腾讯云录制状态变更通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body RecordingStatusEvent true "录制状态事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/recording-status [post]
func (h *LiveCallbackHandler) HandleRecordingStatus(c *gin.Context) {
	var event RecordingStatusEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析录制状态事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	// 验证签名
	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("录制状态事件签名验证失败",
			zap.String("streamId", event.StreamID),
			zap.String("sign", event.Sign))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	// 记录事件日志
	statusText := "未录制"
	if event.Status == 1 {
		statusText = "录制中"
	}
	logger.Info("收到录制状态事件",
		zap.String("streamId", event.StreamID),
		zap.String("domain", event.PushDomain),
		zap.String("app", event.AppName),
		zap.String("stream", event.StreamName),
		zap.String("status", statusText))

	// 保存事件到数据库
	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "recording_status",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1, // TODO: 从上下文获取公司ID
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存录制状态事件失败", zap.Error(err))
	}

	// 返回成功响应
	c.JSON(200, gin.H{"code": 0})
}

// QueryCallbackEvents 查询回调事件列表
// @Summary 查询回调事件列表
// @Description 查询直播回调事件记录
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param eventType query string false "事件类型"
// @Param streamId query string false "流ID"
// @Param domainName query string false "域名"
// @Param appName query string false "应用名称"
// @Param streamName query string false "流名称"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(20)
// @Success 200 {object} utils.Response
// @Router /api/v1/live/callback/events [get]
func (h *LiveCallbackHandler) QueryCallbackEvents(c *gin.Context) {
	// 获取查询参数
	eventType := c.Query("eventType")
	streamID := c.Query("streamId")
	domainName := c.Query("domainName")
	appName := c.Query("appName")
	streamName := c.Query("streamName")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	pageNum := c.DefaultQuery("pageNum", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	// 构建查询
	query := h.db.Model(&entity.LiveCallbackEvent{}).Where("IS_ACTIVE = ?", "Y")

	if eventType != "" {
		query = query.Where("EVENT_TYPE = ?", eventType)
	}
	if streamID != "" {
		query = query.Where("STREAM_ID = ?", streamID)
	}
	if domainName != "" {
		query = query.Where("DOMAIN_NAME = ?", domainName)
	}
	if appName != "" {
		query = query.Where("APP_NAME = ?", appName)
	}
	if streamName != "" {
		query = query.Where("STREAM_NAME = ?", streamName)
	}
	if startTime != "" {
		query = query.Where("CREATE_TIME >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("CREATE_TIME <= ?", endTime)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("统计回调事件失败", zap.Error(err))
		utils.InternalError(c, "查询失败")
		return
	}

	// 分页查询
	var events []entity.LiveCallbackEvent
	page, _ := utils.ParseInt(pageNum, 1)
	size, _ := utils.ParseInt(pageSize, 20)
	offset := (page - 1) * size

	if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(size).Find(&events).Error; err != nil {
		logger.Error("查询回调事件失败", zap.Error(err))
		utils.InternalError(c, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":     events,
		"total":    total,
		"pageNum":  page,
		"pageSize": size,
	})
}

// HandleScreenshot 处理截图事件
// @Summary 截图事件回调
// @Description 接收腾讯云截图事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body ScreenshotEvent true "截图事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/screenshot [post]
func (h *LiveCallbackHandler) HandleScreenshot(c *gin.Context) {
	var event ScreenshotEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析截图事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("截图事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到截图事件",
		zap.String("streamId", event.StreamID),
		zap.String("picUrl", event.PicURL),
		zap.Int("width", event.Width),
		zap.Int("height", event.Height))

	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "screenshot",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存截图事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleVideoAudit 处理画面审核事件
// @Summary 画面审核事件回调
// @Description 接收腾讯云画面审核事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body VideoAuditEvent true "画面审核事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/video-audit [post]
func (h *LiveCallbackHandler) HandleVideoAudit(c *gin.Context) {
	var event VideoAuditEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析画面审核事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("画面审核事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到画面审核事件",
		zap.String("streamId", event.StreamID),
		zap.String("label", event.Label),
		zap.String("suggestion", event.Suggestion),
		zap.Int("confidence", event.Confidence))

	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "video_audit",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存画面审核事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleAudioAudit 处理音频审核事件
// @Summary 音频审核事件回调
// @Description 接收腾讯云音频审核事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body AudioAuditEvent true "音频审核事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/audio-audit [post]
func (h *LiveCallbackHandler) HandleAudioAudit(c *gin.Context) {
	var event AudioAuditEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析音频审核事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("音频审核事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到音频审核事件",
		zap.String("streamId", event.StreamID),
		zap.String("label", event.Label),
		zap.String("suggestion", event.Suggestion),
		zap.String("audioText", event.AudioText))

	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "audio_audit",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存音频审核事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleQualityInspection 处理质检事件
// @Summary 质检事件回调
// @Description 接收腾讯云质检事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body QualityInspectionEvent true "质检事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/quality-inspection [post]
func (h *LiveCallbackHandler) HandleQualityInspection(c *gin.Context) {
	var event QualityInspectionEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析质检事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("质检事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到质检事件",
		zap.String("streamId", event.StreamID),
		zap.String("diagnoseType", event.DiagnoseType),
		zap.String("level", event.Level),
		zap.String("description", event.Description))

	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "quality_inspection",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存质检事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleQualityThreshold 处理评测阈值事件
// @Summary 评测阈值事件回调
// @Description 接收腾讯云评测阈值事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body QualityThresholdEvent true "评测阈值事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/quality-threshold [post]
func (h *LiveCallbackHandler) HandleQualityThreshold(c *gin.Context) {
	var event QualityThresholdEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析评测阈值事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("评测阈值事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到评测阈值事件",
		zap.String("streamId", event.StreamID),
		zap.String("metricType", event.MetricType),
		zap.Float64("threshold", event.Threshold),
		zap.Float64("currentValue", event.CurrentValue))

	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "quality_threshold",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存评测阈值事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleQualityAverage 处理评测平均分事件
// @Summary 评测平均分事件回调
// @Description 接收腾讯云评测平均分事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body QualityAverageEvent true "评测平均分事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/quality-average [post]
func (h *LiveCallbackHandler) HandleQualityAverage(c *gin.Context) {
	var event QualityAverageEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析评测平均分事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("评测平均分事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到评测平均分事件",
		zap.String("streamId", event.StreamID),
		zap.Float64("score", event.Score),
		zap.Int64("duration", event.Duration))

	eventData, _ := json.Marshal(event)
	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "quality_average",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存评测平均分事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}


// getDomainName 获取推流域名（优先使用实际字段 app，回退到文档字段 push_domain）
func getDomainName(event RecordingFileEvent) string {
	if event.App != "" {
		return event.App
	}
	return event.PushDomain
}

// getAppName 获取应用名称（优先使用实际字段 appname，回退到文档字段 app_name）
func getAppName(event RecordingFileEvent) string {
	if event.AppName != "" {
		return event.AppName
	}
	return event.AppNameCompat
}

// getStreamName 获取流名称（优先使用 stream_name，回退到 stream_id）
func getStreamName(event RecordingFileEvent) string {
	if event.StreamName != "" {
		return event.StreamName
	}
	// 如果没有 stream_name，使用 stream_id 作为流名称
	return event.StreamID
}
