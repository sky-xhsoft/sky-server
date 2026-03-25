# 录制列表字段为空问题排查

## 问题描述

直播录制列表页面中，以下字段显示为空：
- 流名称 (streamName)
- 域名 (domainName)
- 应用名称 (appName)

## 代码检查结果

### 1. 后端保存逻辑 ✓ 正确

`live_callback_handler.go:533-545`：
```go
callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    event.EventTime,
    DomainName:   event.PushDomain,      // ✓ 正确映射
    AppName:      event.AppName,          // ✓ 正确映射
    StreamName:   event.StreamName,       // ✓ 正确映射
    StreamID:     event.StreamID,
    EventData:    string(eventData),
    Sign:         event.Sign,
    TValue:       event.T,
    SysCompanyID: 1,
    IsActive:     "Y",
}
```

### 2. 数据库表结构 ✓ 正确

`create_live_callback_event.sql`：
```sql
`DOMAIN_NAME` varchar(255) DEFAULT NULL COMMENT '推流域名',
`APP_NAME` varchar(255) DEFAULT NULL COMMENT '应用名称',
`STREAM_NAME` varchar(255) DEFAULT NULL COMMENT '流名称',
```

### 3. 实体定义 ✓ 正确

`live_callback_event.go:12-14`：
```go
DomainName   string    `gorm:"column:DOMAIN_NAME;type:varchar(255);index" json:"domainName"`
AppName      string    `gorm:"column:APP_NAME;type:varchar(255);index" json:"appName"`
StreamName   string    `gorm:"column:STREAM_NAME;type:varchar(255);index" json:"streamName"`
```

### 4. 查询接口 ✓ 正确

`live_callback_handler.go:685-696`：
```go
if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(size).Find(&events).Error; err != nil {
    // ...
}

utils.Success(c, gin.H{
    "list":     events,  // 直接返回实体数组
    "total":    total,
    "pageNum":  page,
    "pageSize": size,
})
```

### 5. 前端解析逻辑 ✓ 正确

`LiveRecordings.vue:442-444`：
```typescript
const streamName = item.streamName || eventData.stream_name || eventData.streamName || ''
const domainName = item.domainName || eventData.push_domain || eventData.pushDomain || ''
const appName = item.appName || eventData.app_name || eventData.appName || ''
```

## 问题根源

代码逻辑都是正确的，问题很可能是：

### 1. 数据库中的数据本身就是空的

可能的原因：
- 测试数据是手动创建的，只包含了 `event_type` 和 `stream_id` 等最基本字段
- 实际的腾讯云回调还没有触发，或者回调数据不完整
- 早期的测试数据没有包含这些字段

### 2. GORM 查询时的字段选择问题

虽然不太可能，但可以尝试显式指定要查询的字段。

## 解决方案

### 方案1：检查数据库数据（推荐）

执行以下 SQL 查询，检查数据库中的实际数据：

```sql
-- 查看最近的录制事件数据
SELECT
    ID,
    EVENT_TYPE,
    DOMAIN_NAME,
    APP_NAME,
    STREAM_NAME,
    STREAM_ID,
    EVENT_DATA,
    CREATE_TIME
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 5;
```

**预期结果**：
- 如果 `DOMAIN_NAME`、`APP_NAME`、`STREAM_NAME` 列都是 NULL 或空字符串，说明数据本身就是空的
- 如果这些列有值，但前端显示为空，说明是查询或序列化问题

### 方案2：使用完整的测试数据

如果数据库中的数据确实是空的，使用以下完整的测试数据：

```bash
#!/bin/bash

# 发送完整的录制文件回调测试数据
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '$(date +%s)',
    "sign": "test_sign",
    "event_time": '$(date +%s)',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "start_time": '$(date +%s)',
    "end_time": '$(date +%s)',
    "stream_param": "",
    "video_id": "video_001",
    "record_file_id": "file_001",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

**关键字段**：
- `push_domain` → 保存到 `DOMAIN_NAME`
- `app_name` → 保存到 `APP_NAME`
- `stream_name` → 保存到 `STREAM_NAME`

### 方案3：修改后端查询（如果方案1和2都不行）

如果确认数据库有数据但查询不到，可以尝试显式指定查询字段：

```go
// 在 QueryCallbackEvents 函数中
if err := query.
    Select("ID, EVENT_TYPE, EVENT_TIME, DOMAIN_NAME, APP_NAME, STREAM_NAME, STREAM_ID, CLIENT_IP, EVENT_DATA, SIGN, T_VALUE, CREATE_TIME, SYS_COMPANY_ID, IS_ACTIVE").
    Order("CREATE_TIME DESC").
    Offset(offset).
    Limit(size).
    Find(&events).Error; err != nil {
    // ...
}
```

### 方案4：添加调试日志

在后端查询接口中添加日志，查看实际返回的数据：

```go
// 在 QueryCallbackEvents 函数中，返回之前添加
logger.Info("查询到的事件数据",
    zap.Int("count", len(events)),
    zap.Any("firstEvent", events[0])) // 打印第一条数据

utils.Success(c, gin.H{
    "list":     events,
    "total":    total,
    "pageNum":  page,
    "pageSize": size,
})
```

## 验证步骤

### 1. 检查数据库

```sql
-- 检查是否有数据
SELECT COUNT(*) FROM live_callback_event WHERE EVENT_TYPE = 'recording_file';

-- 检查字段是否为空
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN DOMAIN_NAME IS NULL OR DOMAIN_NAME = '' THEN 1 ELSE 0 END) as empty_domain,
    SUM(CASE WHEN APP_NAME IS NULL OR APP_NAME = '' THEN 1 ELSE 0 END) as empty_app,
    SUM(CASE WHEN STREAM_NAME IS NULL OR STREAM_NAME = '' THEN 1 ELSE 0 END) as empty_stream
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';
```

### 2. 发送测试数据

使用上面的 curl 命令发送完整的测试数据。

### 3. 检查前端控制台

打开浏览器控制台，查看以下日志：
- `Processing recording item:` - 查看 `item.domainName`、`item.appName`、`item.streamName` 的值
- `Parsed recording eventData:` - 查看 `eventData` 中的字段

### 4. 检查后端日志

查看后端日志中的 "收到录制文件事件" 日志，确认：
- `domain`: 是否有值
- `app`: 是否有值
- `stream`: 是否有值

## 最可能的原因

根据之前的调试信息（控制台显示 `eventData` 只有 `event_type` 和 `stream_id`），最可能的原因是：

**数据库中的数据本身就是不完整的测试数据**

解决方法：
1. 清空现有的测试数据
2. 使用完整的测试脚本重新创建数据
3. 或者配置腾讯云回调，等待真实的回调数据

## 清空测试数据

```sql
-- 删除所有录制事件测试数据
DELETE FROM live_callback_event WHERE EVENT_TYPE = 'recording_file';

-- 或者只删除字段为空的数据
DELETE FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
  AND (DOMAIN_NAME IS NULL OR DOMAIN_NAME = '')
  AND (APP_NAME IS NULL OR APP_NAME = '')
  AND (STREAM_NAME IS NULL OR STREAM_NAME = '');
```

## 总结

1. **首先**：执行 SQL 查询检查数据库中的实际数据
2. **如果数据为空**：使用完整的测试脚本重新创建数据
3. **如果数据不为空**：添加后端日志，检查查询和序列化过程
4. **最后**：检查前端控制台日志，确认数据传输是否正确
