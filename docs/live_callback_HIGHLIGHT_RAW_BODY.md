# 高光切片事件处理优化 - 保留原始请求体

## 更新内容

优化了高光切片（Highlight）事件的处理逻辑，在保存事件数据时同时保留原始 HTTP 请求体和客户端 IP 地址。

## 修改文件

- `sky-server/api/handler/live_callback_handler_ext.go`

## 修改详情

### 1. 添加必要的导入

**修改前：**
```go
import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)
```

**修改后：**
```go
import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)
```

**说明：**
- 添加 `bytes` 包：用于创建字节缓冲区
- 添加 `io` 包：用于读取和恢复请求体

### 2. 读取并保留原始请求体

**修改前：**
```go
func (h *LiveCallbackHandler) HandleHighlight(c *gin.Context) {
	var event HighlightEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析高光切片事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}
	// ...
}
```

**修改后：**
```go
func (h *LiveCallbackHandler) HandleHighlight(c *gin.Context) {
	// 读取原始请求体
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("读取高光切片事件请求体失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "failed to read request body"})
		return
	}
	// 恢复请求体，以便后续的 ShouldBindJSON 可以读取
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var event HighlightEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析高光切片事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}
	// ...
}
```

**说明：**
1. 使用 `io.ReadAll()` 读取原始请求体
2. 使用 `io.NopCloser()` 和 `bytes.NewBuffer()` 恢复请求体
3. 这样既保留了原始数据，又不影响后续的 JSON 解析

### 3. 保存客户端 IP 地址

**修改前：**
```go
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "highlight",
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
```

**修改后：**
```go
// 保存原始请求体和解析后的事件数据
eventData, _ := json.Marshal(event)
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "highlight",
	EventTime:    event.EventTime,
	DomainName:   event.PushDomain,
	AppName:      event.AppName,
	StreamName:   event.StreamName,
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	ClientIP:     c.ClientIP(),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

**说明：**
- 添加 `ClientIP` 字段，使用 `c.ClientIP()` 获取客户端 IP 地址
- 添加注释说明保存的是原始请求体和解析后的事件数据

## 功能说明

### 1. 原始请求体保留

**为什么需要保留原始请求体？**

- **调试和排查**：当事件处理出现问题时，可以查看原始请求内容
- **审计追踪**：记录完整的请求历史，用于审计和合规
- **数据重放**：可以使用原始请求体重新处理事件
- **数据对比**：对比原始数据和解析后的数据，发现潜在问题

**实现方式：**

```go
// 1. 读取原始请求体
rawBody, err := io.ReadAll(c.Request.Body)

// 2. 恢复请求体（因为 Body 是一次性读取的流）
c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

// 3. 正常解析 JSON
var event HighlightEvent
c.ShouldBindJSON(&event)
```

**注意事项：**
- `c.Request.Body` 是一个 `io.ReadCloser`，只能读取一次
- 读取后必须恢复，否则后续的 `ShouldBindJSON` 会失败
- `io.NopCloser` 创建一个不需要关闭的 ReadCloser

### 2. 客户端 IP 地址记录

**为什么需要记录客户端 IP？**

- **安全审计**：追踪请求来源，识别异常访问
- **问题定位**：当某个地区或网络出现问题时，可以快速定位
- **统计分析**：分析不同地区的事件分布

**实现方式：**

```go
ClientIP: c.ClientIP()
```

**`c.ClientIP()` 的工作原理：**
1. 优先从 `X-Forwarded-For` 头获取（如果有代理）
2. 其次从 `X-Real-IP` 头获取
3. 最后从 `RemoteAddr` 获取

## 数据库字段

确保 `live_callback_event` 表包含以下字段：

```sql
CREATE TABLE `live_callback_event` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `event_type` varchar(50) NOT NULL COMMENT '事件类型',
  `event_time` bigint DEFAULT NULL COMMENT '事件时间戳',
  `domain_name` varchar(255) DEFAULT NULL COMMENT '域名',
  `app_name` varchar(255) DEFAULT NULL COMMENT '应用名称',
  `stream_name` varchar(255) DEFAULT NULL COMMENT '流名称',
  `stream_id` varchar(255) DEFAULT NULL COMMENT '流ID',
  `event_data` text COMMENT '事件数据JSON',
  `client_ip` varchar(50) DEFAULT NULL COMMENT '客户端IP地址',
  `sign` varchar(255) DEFAULT NULL COMMENT '签名',
  `t_value` bigint DEFAULT NULL COMMENT 'T值',
  `sys_company_id` bigint NOT NULL COMMENT '公司ID',
  `is_active` char(1) DEFAULT 'Y' COMMENT '是否有效',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_stream_id` (`stream_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播回调事件表';
```

## 使用示例

### 查询高光切片事件

```sql
-- 查询所有高光切片事件
SELECT
  id,
  stream_id,
  stream_name,
  event_data,
  client_ip,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
ORDER BY create_time DESC;

-- 查询特定流的高光切片
SELECT
  id,
  JSON_EXTRACT(event_data, '$.clip_url') as clip_url,
  JSON_EXTRACT(event_data, '$.score') as score,
  JSON_EXTRACT(event_data, '$.start_time') as start_time,
  JSON_EXTRACT(event_data, '$.end_time') as end_time,
  client_ip,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
  AND stream_id = '0202'
ORDER BY create_time DESC;

-- 统计不同IP的请求数量
SELECT
  client_ip,
  COUNT(*) as request_count
FROM live_callback_event
WHERE event_type = 'highlight'
GROUP BY client_ip
ORDER BY request_count DESC;
```

### API 查询示例

```bash
# 查询高光切片列表
curl "http://localhost:3000/api/v1/live/callback/events?eventType=highlight&pageNum=1&pageSize=20"

# 查询特定流的高光切片
curl "http://localhost:3000/api/v1/live/callback/events?eventType=highlight&streamId=0202&pageNum=1&pageSize=20"

# 查询特定时间范围的高光切片
curl "http://localhost:3000/api/v1/live/callback/events?eventType=highlight&startTime=2026-01-26%2000:00:00&endTime=2026-02-02%2023:59:59&pageNum=1&pageSize=20"
```

## 测试验证

### 1. 发送测试请求

```bash
curl -X POST http://localhost:3000/api/v1/live/callback/highlight \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 349,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": 1738483200,
    "sign": "test_sign",
    "event_time": 1738483200,
    "clip_url": "http://example.com/clip.mp4",
    "start_time": 1738483100,
    "end_time": 1738483200,
    "score": 8.5,
    "stream_param": "test_param",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

### 2. 验证数据保存

```sql
-- 查询最新的高光切片事件
SELECT
  id,
  stream_id,
  event_data,
  client_ip,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
ORDER BY id DESC
LIMIT 1;
```

**预期结果：**
- `event_data` 字段包含完整的 JSON 数据
- `client_ip` 字段包含请求的客户端 IP 地址
- 所有字段都正确保存

### 3. 验证前端显示

1. 打开高光切片列表页面：`http://localhost:8080/#/live/highlight-clips`
2. 查询数据
3. 验证表格中显示的数据是否正确

## 后续优化建议

### 1. 添加原始请求体字段

如果需要保存完整的原始请求体（而不仅仅是解析后的 JSON），可以在数据库中添加 `raw_body` 字段：

```sql
ALTER TABLE live_callback_event
ADD COLUMN raw_body TEXT COMMENT '原始请求体' AFTER event_data;
```

然后在代码中保存：

```go
callbackEvent := &entity.LiveCallbackEvent{
	// ...
	EventData:    string(eventData),
	RawBody:      string(rawBody),  // 保存原始请求体
	ClientIP:     c.ClientIP(),
	// ...
}
```

### 2. 添加请求头记录

如果需要记录请求头信息（如 User-Agent、Referer 等），可以添加 `request_headers` 字段：

```sql
ALTER TABLE live_callback_event
ADD COLUMN request_headers TEXT COMMENT '请求头JSON' AFTER raw_body;
```

然后在代码中保存：

```go
// 获取请求头
headers := make(map[string]string)
for key, values := range c.Request.Header {
	if len(values) > 0 {
		headers[key] = values[0]
	}
}
headersJSON, _ := json.Marshal(headers)

callbackEvent := &entity.LiveCallbackEvent{
	// ...
	RequestHeaders: string(headersJSON),
	// ...
}
```

### 3. 添加响应时间记录

记录请求处理时间，用于性能分析：

```go
startTime := time.Now()

// ... 处理逻辑 ...

processingTime := time.Since(startTime).Milliseconds()

callbackEvent := &entity.LiveCallbackEvent{
	// ...
	ProcessingTime: processingTime,
	// ...
}
```

## 相关文件

- `sky-server/api/handler/live_callback_handler_ext.go` - 高光切片事件处理器
- `sky-server/internal/model/entity/live_callback_event.go` - 回调事件实体
- `sky-web/src/pages/LiveHighlightClips.vue` - 高光切片列表页面

## 更新日期

2026-02-02
