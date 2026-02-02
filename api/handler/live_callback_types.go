package handler

// 腾讯云直播回调事件类型常量
const (
	// 基础事件
	EventTypePushStream       = "push_stream"       // 推流事件
	EventTypeDisconnectStream = "disconnect_stream" // 断流事件

	// 录制相关
	EventTypeRecordingFile   = "recording_file"   // 录制文件生成
	EventTypeRecordingStatus = "recording_status" // 录制状态变更
	EventTypeRecordException = "record_exception" // 录制异常

	// 截图和审核
	EventTypeScreenshot = "screenshot"   // 截图
	EventTypeVideoAudit = "video_audit"  // 画面审核
	EventTypeAudioAudit = "audio_audit"  // 音频审核

	// 质检和评测
	EventTypeQualityInspection = "quality_inspection" // 质检
	EventTypeQualityThreshold  = "quality_threshold"  // 评测阈值
	EventTypeQualityAverage    = "quality_average"    // 评测平均分

	// AI功能
	EventTypeSmartErase = "smart_erase" // 智能擦除
	EventTypeSubtitle   = "subtitle"    // 直播字幕
	EventTypeSummary    = "summary"     // 直播摘要
	EventTypeHighlight  = "highlight"   // 高光切片

	// 异常和监控
	EventTypePushException = "push_exception" // 推流异常
	EventTypePullStream    = "pull_stream"    // 拉流转推
	EventTypeMonitor       = "monitor"        // 监播
)

// 事件类型映射表（腾讯云event_type -> 内部事件类型）
var EventTypeMapping = map[int]string{
	0:   EventTypeDisconnectStream, // 断流
	1:   EventTypePushStream,       // 推流
	100: EventTypeRecordingFile,    // 录制文件
	200: EventTypeRecordingStatus,  // 录制状态（也可能是截图，需要根据具体字段判断）
	317: EventTypeVideoAudit,       // 画面审核
	318: EventTypeAudioAudit,       // 音频审核
	319: EventTypeQualityInspection, // 质检
	320: EventTypeQualityThreshold,  // 评测阈值
	321: EventTypeQualityAverage,    // 评测平均分
	322: EventTypeSmartErase,        // 智能擦除
	323: EventTypeSubtitle,          // 直播字幕
	324: EventTypeSummary,           // 直播摘要
	325: EventTypeHighlight,         // 高光切片
	326: EventTypePushException,     // 推流异常
	327: EventTypeRecordException,   // 录制异常
	328: EventTypePullStream,        // 拉流转推
	329: EventTypeMonitor,           // 监播
}

// GetEventTypeName 根据event_type获取事件类型名称
func GetEventTypeName(eventType int) string {
	if name, ok := EventTypeMapping[eventType]; ok {
		return name
	}
	return "unknown"
}

// 事件类型描述
var EventTypeDescriptions = map[string]string{
	EventTypePushStream:        "推流事件",
	EventTypeDisconnectStream:  "断流事件",
	EventTypeRecordingFile:     "录制文件生成",
	EventTypeRecordingStatus:   "录制状态变更",
	EventTypeRecordException:   "录制异常",
	EventTypeScreenshot:        "截图",
	EventTypeVideoAudit:        "画面审核",
	EventTypeAudioAudit:        "音频审核",
	EventTypeQualityInspection: "质检",
	EventTypeQualityThreshold:  "评测阈值",
	EventTypeQualityAverage:    "评测平均分",
	EventTypeSmartErase:        "智能擦除",
	EventTypeSubtitle:          "直播字幕",
	EventTypeSummary:           "直播摘要",
	EventTypeHighlight:         "高光切片",
	EventTypePushException:     "推流异常",
	EventTypePullStream:        "拉流转推",
	EventTypeMonitor:           "监播",
}

// GetEventTypeDescription 获取事件类型描述
func GetEventTypeDescription(eventType string) string {
	if desc, ok := EventTypeDescriptions[eventType]; ok {
		return desc
	}
	return "未知事件"
}
