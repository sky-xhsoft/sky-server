# 录制列表字段为空问题修复 - 腾讯云实际字段名

## 问题根源

腾讯云直播录制回调的**实际字段名**与**官方文档**不一致！

### 官方文档中的字段名
```json
{
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 实际回调中的字段名
```json
{
  "app": "upload.skyzhou.cn",
  "appname": "live",
  "stream_id": "0202"
}
```

**注意**：
- `push_domain` → 实际是 `app`
- `app_name` → 实际是 `appname`
- `stream_name` → 实际不存在，只有 `stream_id`

## 实际回调数据示例

```json
{
  "app" : "upload.skyzhou.cn",
  "appid" : 1301212747,
  "appname" : "live",
  "callback_ext" : "{\"video_codec\":\"h264\",\"session_id\":\"1807121258070784645\",\"resolution\":\"1920x1080\"}",
  "channel_id" : "0202",
  "duration" : 56,
  "end_time" : 1770031273,
  "end_time_usec" : 880554,
  "event_type" : 100,
  "file_format" : "mp4",
  "file_id" : "1301212747_d782d6455d5e42f78181e0e475e14cfc",
  "file_size" : 43911385,
  "media_start_time" : 33,
  "record_bps" : 0,
  "record_file_id" : "1301212747_d782d6455d5e42f78181e0e475e14cfc",
  "record_temp_id" : "1622658",
  "start_time" : 1770031217,
  "start_time_usec" : 650373,
  "stream_id" : "0202",
  "stream_param" : "txSecret=5d00f4090fdb09f6666cc9b278f47097&txTime=69830DEC",
  "task_id" : "1807121258070784645",
  "video_id" : "1301212747_11554a2e8cbb4326a0fe858f9418f5eb",
  "video_url" : "http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/0202/1807121258070784645-6d97196c6eec4b66b62355185cf34146/2026-02-02-19-20-17.mp4"
}
```

## 解决方案

### 1. 更新 RecordingFileEvent 结构体

添加实际的字段名，同时保持向后兼容：

```go
type RecordingFileEvent struct {
	EventType    int    `json:"event_type"`
	StreamID     string `json:"stream_id"`
	// ... 其他字段 ...

	// 实际的腾讯云字段（优先使用）
	App     string `json:"app"`     // 推流域名（实际字段）
	AppName string `json:"appname"` // 应用名称（实际字段）

	// 文档中的字段（向后兼容）
	PushDomain     string `json:"push_domain"`  // 推流域名（文档字段）
	AppNameCompat  string `json:"app_name"`     // 应用名称（文档字段）
	StreamName     string `json:"stream_name"`  // 流名称（文档字段）
}
```

### 2. 添加辅助函数

创建辅助函数来获取字段值，优先使用实际字段：

```go
// getDomainName 获取推流域名（优先使用实际字段 app）
func getDomainName(event RecordingFileEvent) string {
	if event.App != "" {
		return event.App
	}
	return event.PushDomain
}

// getAppName 获取应用名称（优先使用实际字段 appname）
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
	return event.StreamID
}
```

### 3. 更新保存逻辑

使用辅助函数来获取字段值：

```go
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "recording_file",
	EventTime:    event.EventTime,
	DomainName:   getDomainName(event),  // 优先使用 app
	AppName:      getAppName(event),     // 优先使用 appname
	StreamName:   getStreamName(event),  // 优先使用 stream_name，回退到 stream_id
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

## 修改的文件

- `sky-server/api/handler/live_callback_handler.go`
  - 第69-95行：更新 `RecordingFileEvent` 结构体
  - 第520-545行：更新保存逻辑
  - 第1010-1033行：添加辅助函数

## 字段映射关系

| 数据库字段 | 实际回调字段 | 文档字段 | 说明 |
|-----------|-------------|---------|------|
| DOMAIN_NAME | `app` | `push_domain` | 推流域名 |
| APP_NAME | `appname` | `app_name` | 应用名称 |
| STREAM_NAME | - | `stream_name` | 流名称（实际回调中不存在，使用 stream_id） |
| STREAM_ID | `stream_id` | `stream_id` | 流ID |

## 向后兼容性

这个修复保持了向后兼容：
- ✅ 如果回调使用实际字段（`app`、`appname`），会正确解析
- ✅ 如果回调使用文档字段（`push_domain`、`app_name`），也能正确解析
- ✅ 如果没有 `stream_name`，会使用 `stream_id` 作为流名称

## 测试验证

### 1. 使用实际数据测试

```bash
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "0202",
    "channel_id": "0202",
    "t": 1234567890,
    "sign": "test",
    "event_time": 1234567890,
    "video_url": "http://example.com/video.mp4",
    "file_size": 43911385,
    "duration": 56,
    "file_format": "mp4",
    "start_time": 1234567834,
    "end_time": 1234567890,
    "app": "upload.skyzhou.cn",
    "appname": "live"
  }'
```

### 2. 检查数据库

```sql
SELECT
    ID,
    DOMAIN_NAME,
    APP_NAME,
    STREAM_NAME,
    STREAM_ID,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 1;
```

**预期结果**：
- `DOMAIN_NAME` = "upload.skyzhou.cn"
- `APP_NAME` = "live"
- `STREAM_NAME` = "0202" (使用 stream_id)
- `STREAM_ID` = "0202"

### 3. 检查前端显示

刷新录制列表页面，应该能看到：
- 域名：upload.skyzhou.cn
- 应用名：live
- 流名称：0202

## 其他发现的字段

实际回调中还包含以下额外字段（可以考虑后续添加）：

- `appid`: 应用ID
- `callback_ext`: 回调扩展信息（JSON字符串）
  - `video_codec`: 视频编码
  - `session_id`: 会话ID
  - `resolution`: 分辨率
- `end_time_usec`: 结束时间微秒
- `start_time_usec`: 开始时间微秒
- `media_start_time`: 媒体开始时间
- `record_bps`: 录制码率
- `record_temp_id`: 录制临时ID
- `task_id`: 任务ID
- `file_id`: 文件ID

## 总结

这个问题的根本原因是**腾讯云的实际回调数据与官方文档不一致**。通过添加实际字段名并保持向后兼容，我们解决了字段为空的问题。

修复后，系统能够：
1. 正确解析腾讯云的实际回调数据
2. 保持对文档字段的向后兼容
3. 在前端正确显示域名、应用名和流名称
