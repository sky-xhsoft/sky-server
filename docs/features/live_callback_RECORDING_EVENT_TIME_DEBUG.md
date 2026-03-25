# 直播录制列表生成时间不显示问题调试

## 问题描述

直播录制列表页面中，"生成时间"列不显示数据。

## 问题分析

### 1. 前端代码检查

**表格列定义** (LiveRecordings.vue:177-181)：
```vue
<a-table-column title="生成时间" :width="180">
  <template #cell="{ record }">
    {{ formatDateTime(record.eventTime) }}
  </template>
</a-table-column>
```

**数据解析** (LiveRecordings.vue:468)：
```typescript
eventTime: item.eventTime != null ? item.eventTime * 1000 : null
```

**时间格式化函数** (LiveRecordings.vue:589-594)：
```typescript
const formatDateTime = (timestamp: number | string) => {
  if (timestamp == null || isNaN(Number(timestamp))) return '-'
  const date = dayjs(timestamp)
  if (!date.isValid()) return '-'
  return date.format('YYYY-MM-DD HH:mm:ss')
}
```

### 2. 后端数据结构

**实体定义** (live_callback_event.go:11)：
```go
EventTime int64 `gorm:"column:EVENT_TIME;not null" json:"eventTime"` // 事件时间戳（秒）
```

**腾讯云回调数据**：
```json
{
  "event_time": 1770031273,
  ...
}
```

### 3. 可能的原因

1. **后端 eventTime 字段为 0 或 null**
   - 保存时没有正确设置 eventTime
   - 数据库中的值为 0

2. **时间戳单位问题**
   - 后端存储的是秒级时间戳
   - 前端需要毫秒级时间戳（乘以 1000）
   - 如果时间戳太小或太大，dayjs 可能无法正确解析

3. **字段名映射问题**
   - JSON 序列化时字段名不匹配

## 调试步骤

### 已添加的调试日志

在 `LiveRecordings.vue` 中添加了以下调试日志：

```typescript
console.log('Processing recording item:', item)
console.log('item.eventTime:', item.eventTime, 'type:', typeof item.eventTime)
const eventData = JSON.parse(item.eventData)
console.log('Parsed recording eventData:', eventData)
console.log('Available eventData keys:', Object.keys(eventData))

// ...

const processedEventTime = item.eventTime != null ? item.eventTime * 1000 : null
console.log('Processed eventTime:', processedEventTime, 'from:', item.eventTime)
```

### 如何查看调试信息

1. 打开浏览器开发者工具（F12）
2. 切换到 Console 标签
3. 刷新录制列表页面
4. 查看控制台输出

### 需要检查的信息

1. **item.eventTime 的值**：
   - 是否为 null 或 undefined？
   - 是否为 0？
   - 数值是否合理（应该是秒级时间戳，如 1770031273）？

2. **processedEventTime 的值**：
   - 乘以 1000 后的值是否正确？
   - 是否为 null？

3. **eventData 中的 event_time**：
   - 检查原始回调数据中是否有 event_time 字段
   - 值是否正确？

## 可能的解决方案

### 方案1：后端保存时设置 eventTime

如果 `item.eventTime` 为 0 或 null，需要检查后端保存逻辑：

**检查点** (live_callback_handler.go:533-545)：
```go
callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    event.EventTime,  // 确保这个值不为 0
    DomainName:   getDomainName(event),
    AppName:      getAppName(event),
    StreamName:   getStreamName(event),
    StreamID:     event.StreamID,
    EventData:    string(eventData),
    Sign:         event.Sign,
    TValue:       event.T,
    SysCompanyID: 1,
    IsActive:     "Y",
}
```

**可能的问题**：
- 腾讯云回调中 `event_time` 字段为 0
- 结构体解析失败，`event.EventTime` 为 0

**解决方法**：
如果 `event.EventTime` 为 0，使用当前时间：
```go
eventTime := event.EventTime
if eventTime == 0 {
    eventTime = time.Now().Unix()
}

callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    eventTime,
    // ...
}
```

### 方案2：前端使用 createTime 作为备选

如果 eventTime 不可用，使用 createTime：

```typescript
const processedEventTime = item.eventTime != null && item.eventTime > 0
  ? item.eventTime * 1000
  : (item.createTime ? new Date(item.createTime).getTime() : null)
```

### 方案3：从 eventData 中提取 event_time

如果后端的 eventTime 字段没有正确保存，可以从 eventData 中提取：

```typescript
const eventTime = item.eventTime || eventData.event_time || eventData.eventTime || null
const processedEventTime = eventTime != null && eventTime > 0
  ? eventTime * 1000
  : null
```

## 数据库检查

### SQL 查询

检查数据库中的实际数据：

```sql
-- 查看最近的录制事件
SELECT
    ID,
    EVENT_TYPE,
    EVENT_TIME,
    CREATE_TIME,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 5;

-- 检查 EVENT_TIME 为 0 的记录
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN EVENT_TIME = 0 THEN 1 ELSE 0 END) as zero_event_time,
    SUM(CASE WHEN EVENT_TIME IS NULL THEN 1 ELSE 0 END) as null_event_time
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';
```

### 预期结果

- `EVENT_TIME` 应该是秒级时间戳（如 1770031273）
- `EVENT_TIME` 不应该为 0 或 NULL
- `CREATE_TIME` 应该是 datetime 格式

## 时间戳验证

### 验证时间戳是否合理

```javascript
// 在浏览器控制台中验证
const timestamp = 1770031273 // 从数据库或日志中获取
console.log('原始时间戳（秒）:', timestamp)
console.log('转换为毫秒:', timestamp * 1000)
console.log('格式化时间:', new Date(timestamp * 1000).toLocaleString())
```

**预期输出**：
```
原始时间戳（秒）: 1770031273
转换为毫秒: 1770031273000
格式化时间: 2026/2/2 19:21:13
```

如果时间不合理（如显示为 1970年），说明时间戳有问题。

## 下一步

1. **查看控制台日志**：
   - 打开页面，查看 `item.eventTime` 的实际值
   - 确认是否为 0、null 或其他异常值

2. **检查数据库**：
   - 运行 SQL 查询，查看 `EVENT_TIME` 字段的实际值
   - 确认是否有数据

3. **根据调试结果选择解决方案**：
   - 如果 eventTime 为 0：修改后端保存逻辑
   - 如果 eventTime 为 null：使用 createTime 作为备选
   - 如果时间戳格式错误：调整转换逻辑

## 临时解决方案

在确定根本原因之前，可以使用 createTime 作为备选：

```typescript
eventTime: item.eventTime != null && item.eventTime > 0
  ? item.eventTime * 1000
  : (item.createTime ? new Date(item.createTime).getTime() : null)
```

这样至少能显示创建时间，而不是空白。

## 相关文件

- `sky-web/src/pages/LiveRecordings.vue` - 前端列表页面
- `sky-server/internal/model/entity/live_callback_event.go` - 后端实体定义
- `sky-server/api/handler/live_callback_handler.go` - 后端回调处理器
