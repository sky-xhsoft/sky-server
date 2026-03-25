# 直播录制列表生成时间不显示问题修复

## 问题描述

直播录制列表页面中，"生成时间"列不显示数据，显示为空白。

## 问题原因

腾讯云录制回调中的 `event_time` 字段可能为 0 或不存在，导致：
1. 后端保存时 `EventTime` 字段为 0
2. 前端显示时，`formatDateTime(0)` 返回无效日期
3. 或者前端判断为 null/0 后显示 "-"

## 解决方案

### 后端修复：使用当前时间作为备选

在保存录制事件时，如果 `event_time` 为 0，使用当前时间：

```go
// 如果 event_time 为 0，使用当前时间
eventTime := event.EventTime
if eventTime == 0 {
    eventTime = time.Now().Unix()
    logger.Warn("录制文件事件的 event_time 为 0，使用当前时间",
        zap.String("streamId", event.StreamID),
        zap.Int64("currentTime", eventTime))
}

// 保存事件到数据库
callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    eventTime,  // 使用处理后的时间
    // ...
}
```

### 添加日志

在日志中添加 `eventTime` 字段，方便调试：

```go
logger.Info("收到录制文件事件",
    zap.String("streamId", event.StreamID),
    zap.String("domain", getDomainName(event)),
    zap.String("app", getAppName(event)),
    zap.String("stream", getStreamName(event)),
    zap.String("videoUrl", event.VideoURL),
    zap.Int64("fileSize", event.FileSize),
    zap.Int64("duration", event.Duration),
    zap.String("format", event.FileFormat),
    zap.Int64("eventTime", event.EventTime))  // 添加 eventTime 日志
```

## 修改的文件

- `sky-server/api/handler/live_callback_handler.go`
  - 第527-551行：添加 eventTime 日志和备选逻辑

## 为什么 event_time 可能为 0？

### 可能的原因

1. **腾讯云回调数据不完整**：
   - 某些情况下，腾讯云可能不发送 `event_time` 字段
   - 或者发送的值为 0

2. **JSON 解析问题**：
   - 如果字段不存在，Go 会将 int64 类型默认为 0
   - 如果字段名不匹配，也会导致解析为 0

3. **测试数据问题**：
   - 手动创建的测试数据可能没有包含 `event_time` 字段

## 腾讯云实际回调数据分析

根据用户提供的实际回调数据：

```json
{
  "app" : "upload.skyzhou.cn",
  "appid" : 1301212747,
  "appname" : "live",
  "callback_ext" : "{...}",
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
  "video_url" : "http://video-1301212747.cos.ap-nanjing.myqcloud.com/..."
}
```

**注意**：这个数据中**没有 `event_time` 字段**！

### 可用的时间字段

实际回调数据中有以下时间字段：
- `start_time`: 1770031217（录制开始时间，秒）
- `end_time`: 1770031273（录制结束时间，秒）
- `start_time_usec`: 650373（开始时间微秒部分）
- `end_time_usec`: 880554（结束时间微秒部分）

但是**没有 `event_time` 字段**！

## 更好的解决方案

既然腾讯云的实际回调中没有 `event_time` 字段，我们应该使用其他时间字段作为备选：

### 方案1：使用 end_time 作为生成时间

录制文件生成时间应该接近录制结束时间，所以可以使用 `end_time`：

```go
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
```

### 方案2：添加 EndTime 到结构体（如果还没有）

确保 `RecordingFileEvent` 结构体中有 `EndTime` 字段（已经有了，第81行）。

## 实施更好的方案

让我更新代码以使用 `end_time` 作为备选：

```go
// 确定事件时间
eventTime := event.EventTime
if eventTime == 0 {
    // 优先使用录制结束时间
    if event.EndTime > 0 {
        eventTime = event.EndTime
        logger.Info("event_time 为 0，使用 end_time 作为生成时间",
            zap.String("streamId", event.StreamID),
            zap.Int64("endTime", event.EndTime))
    } else {
        // 最后使用当前时间
        eventTime = time.Now().Unix()
        logger.Warn("event_time 和 end_time 都为 0，使用当前时间",
            zap.String("streamId", event.StreamID),
            zap.Int64("currentTime", eventTime))
    }
}
```

## 前端调试

前端已经添加了调试日志（LiveRecordings.vue）：

```typescript
console.log('item.eventTime:', item.eventTime, 'type:', typeof item.eventTime)
console.log('Processed eventTime:', processedEventTime, 'from:', item.eventTime)
```

查看控制台输出可以确认：
1. 后端是否正确保存了 eventTime
2. 前端是否正确接收和处理了 eventTime

## 测试验证

### 1. 检查后端日志

重启后端服务后，当收到录制回调时，查看日志：

```
INFO  收到录制文件事件
  streamId: 0202
  domain: upload.skyzhou.cn
  app: live
  stream: 0202
  videoUrl: http://video-1301212747.cos.ap-nanjing.myqcloud.com/...
  fileSize: 43911385
  duration: 56
  format: mp4
  eventTime: 0  <-- 如果为 0，说明回调中没有 event_time
```

如果看到 `eventTime: 0`，然后应该看到：

```
INFO  event_time 为 0，使用 end_time 作为生成时间
  streamId: 0202
  endTime: 1770031273
```

### 2. 检查数据库

```sql
SELECT
    ID,
    EVENT_TYPE,
    EVENT_TIME,
    CREATE_TIME,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 1;
```

**预期结果**：
- `EVENT_TIME` 应该有值（不为 0）
- 如果原始回调没有 `event_time`，应该等于 `end_time` 的值

### 3. 检查前端显示

刷新录制列表页面，"生成时间"列应该显示时间，而不是空白或 "-"。

## 总结

### 问题根源

腾讯云的实际录制回调数据中**没有 `event_time` 字段**，导致后端保存时 `EventTime` 为 0，前端无法显示。

### 解决方案

1. **当前修复**：如果 `event_time` 为 0，使用当前时间
2. **更好的方案**：如果 `event_time` 为 0，使用 `end_time`（录制结束时间）

### 优先级

```
event_time > end_time > 当前时间
```

这样可以确保：
- 如果有 `event_time`，使用它（最准确）
- 如果没有，使用 `end_time`（接近生成时间）
- 如果都没有，使用当前时间（保底方案）

## 下一步

建议实施"更好的方案"，使用 `end_time` 作为第一备选，而不是直接使用当前时间。这样生成时间会更准确。
