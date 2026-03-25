# 高光切片事件结构体修复 - 匹配腾讯云实际格式

## 问题描述

原有的 `HighlightEvent` 结构体定义与腾讯云实际返回的数据格式不匹配，导致：
1. 无法正确解析请求数据
2. 所有字段都是空值
3. 数据无法保存到数据库

## 实际数据格式

腾讯云实际返回的高光切片事件格式：

```json
{
  "appid": 1301212747,
  "domain": "upload.skyzhou.cn",
  "event_type": 349,
  "items": [
    {
      "begin_time": 1770026056,
      "cov_img_store_url": "",
      "end_time": 1770026081,
      "key_words": ["全球标准", "技术规范", "新法规", "高性能"],
      "summary": "官方解读Hima高性能、多功能、全地形能力的技术规范，标志全球皮卡行业新标准。",
      "title": "新法规解读：Hima技术规范与全球标准",
      "video_store_url": "http://video-1301212747.cos.ap-nanjing.myqcloud.com/..."
    }
  ],
  "path": "live",
  "stream_id": "0202"
}
```

**关键特点：**
1. 包含 `items` 数组，一个请求可能包含多个高光切片
2. 使用 `begin_time` 和 `end_time` 而不是 `start_time` 和 `end_time`
3. 包含 `title`、`summary`、`key_words` 等 AI 生成的内容
4. 使用 `video_store_url` 而不是 `clip_url`
5. 包含 `appid`、`domain`、`path` 等顶层字段

## 修复方案

### 1. 更新结构体定义

**文件：** `sky-server/api/handler/live_callback_handler.go`

**修改前：**
```go
// HighlightEvent 高光切片事件通知
type HighlightEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：325-高光切片
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	ClipURL     string   `json:"clip_url"`     // 切片URL
	StartTime   int64    `json:"start_time"`   // 开始时间
	EndTime     int64    `json:"end_time"`     // 结束时间
	Score       float64  `json:"score"`        // 精彩度评分
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}
```

**修改后：**
```go
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
```

### 2. 更新处理逻辑

**文件：** `sky-server/api/handler/live_callback_handler_ext.go`

**关键改动：**

1. **循环处理多个切片**：
```go
// 为每个高光切片创建一条记录
for i, item := range event.Items {
	// 处理每个切片...
}
```

2. **构建兼容的数据结构**：
```go
itemData := map[string]interface{}{
	"event_type":        event.EventType,
	"stream_id":         event.StreamID,
	"appid":             event.AppID,
	"domain":            event.Domain,
	"path":              event.Path,
	"begin_time":        item.BeginTime,
	"end_time":          item.EndTime,
	"video_store_url":   item.VideoStoreURL,
	"cov_img_store_url": item.CovImgStoreURL,
	"title":             item.Title,
	"summary":           item.Summary,
	"key_words":         item.KeyWords,
	// 兼容旧字段名
	"clip_url":          item.VideoStoreURL,
	"start_time":        item.BeginTime,
	"end_time":          item.EndTime,
}
```

3. **保存到数据库**：
```go
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "highlight",
	EventTime:    item.BeginTime,
	DomainName:   event.Domain,
	AppName:      event.Path,
	StreamName:   "",
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	ClientIP:     c.ClientIP(),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

## 数据映射关系

| 腾讯云字段 | 数据库字段 | 说明 |
|-----------|-----------|------|
| `appid` | `event_data.appid` | 应用ID |
| `domain` | `domain_name` | 域名 |
| `event_type` | `event_data.event_type` | 事件类型（349） |
| `path` | `app_name` | 路径（如：live） |
| `stream_id` | `stream_id` | 流ID |
| `items[].begin_time` | `event_time` | 开始时间 |
| `items[].end_time` | `event_data.end_time` | 结束时间 |
| `items[].video_store_url` | `event_data.video_store_url` | 视频URL |
| `items[].title` | `event_data.title` | 标题 |
| `items[].summary` | `event_data.summary` | 摘要 |
| `items[].key_words` | `event_data.key_words` | 关键词 |

## 兼容性处理

为了保持前端兼容性，在 `event_data` 中同时保存新旧字段名：

```json
{
  "begin_time": 1770026056,
  "end_time": 1770026081,
  "video_store_url": "http://...",
  "start_time": 1770026056,    // 兼容旧字段名
  "end_time": 1770026081,      // 兼容旧字段名
  "clip_url": "http://..."     // 兼容旧字段名
}
```

这样前端可以使用 `start_time` 或 `begin_time`，`clip_url` 或 `video_store_url`。

## 前端更新建议

虽然后端已经做了兼容处理，但建议前端也更新字段名以匹配实际格式：

**文件：** `sky-web/src/pages/LiveHighlightClips.vue`

```typescript
// 当前使用（兼容）
startTime: eventData.start_time != null ? eventData.start_time * 1000 : null
clipUrl: eventData.clip_url

// 建议更新为
startTime: eventData.begin_time != null ? eventData.begin_time * 1000 : null
clipUrl: eventData.video_store_url

// 或者使用兼容写法
startTime: (eventData.begin_time || eventData.start_time) != null
  ? (eventData.begin_time || eventData.start_time) * 1000
  : null
clipUrl: eventData.video_store_url || eventData.clip_url
```

## 新增字段展示

前端可以展示更多 AI 生成的内容：

1. **标题（title）**：高光切片的标题
2. **摘要（summary）**：内容摘要
3. **关键词（key_words）**：关键词列表
4. **封面图（cov_img_store_url）**：封面图片URL

**示例：**
```vue
<template>
  <a-table-column title="标题" :width="200">
    <template #cell="{ record }">
      {{ record.title || '-' }}
    </template>
  </a-table-column>

  <a-table-column title="摘要" :width="300">
    <template #cell="{ record }">
      <a-tooltip :content="record.summary">
        <div class="text-ellipsis">{{ record.summary || '-' }}</div>
      </a-tooltip>
    </template>
  </a-table-column>

  <a-table-column title="关键词" :width="200">
    <template #cell="{ record }">
      <a-space wrap>
        <a-tag v-for="keyword in record.keyWords" :key="keyword">
          {{ keyword }}
        </a-tag>
      </a-space>
    </template>
  </a-table-column>
</template>

<script setup lang="ts">
const tableData = ref([])

const loadData = async () => {
  // ...
  tableData.value = res.data.list.map((item: any) => {
    const eventData = JSON.parse(item.eventData)
    return {
      id: item.id,
      streamId: item.streamId,
      streamName: item.streamName,
      domainName: item.domainName,
      appName: item.appName,
      clipUrl: eventData.video_store_url || eventData.clip_url,
      startTime: (eventData.begin_time || eventData.start_time) != null
        ? (eventData.begin_time || eventData.start_time) * 1000
        : null,
      endTime: (eventData.end_time) != null
        ? eventData.end_time * 1000
        : null,
      title: eventData.title,           // 新增
      summary: eventData.summary,       // 新增
      keyWords: eventData.key_words,    // 新增
      coverUrl: eventData.cov_img_store_url, // 新增
      eventTime: item.eventTime != null ? item.eventTime * 1000 : null,
      createTime: item.createTime
    }
  })
}
</script>
```

## 测试验证

### 1. 发送测试请求

使用您提供的实际数据格式：

```bash
curl -X POST http://localhost:3000/api/v1/live/callback/highlight \
  -H "Content-Type: application/json" \
  -d '{
    "appid": 1301212747,
    "domain": "upload.skyzhou.cn",
    "event_type": 349,
    "items": [
      {
        "begin_time": 1770026056,
        "cov_img_store_url": "",
        "end_time": 1770026081,
        "key_words": ["全球标准", "技术规范", "新法规", "高性能"],
        "summary": "官方解读Hima高性能、多功能、全地形能力的技术规范，标志全球皮卡行业新标准。",
        "title": "新法规解读：Hima技术规范与全球标准",
        "video_store_url": "http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/0202/LIVE_0228577F359F9989378520BF9280AA1274-1770025101592/2026-02-02_17-59-14_954866.mp4"
      }
    ],
    "path": "live",
    "stream_id": "0202"
  }'
```

### 2. 验证数据保存

```sql
SELECT
  id,
  stream_id,
  domain_name,
  app_name,
  event_time,
  JSON_EXTRACT(event_data, '$.title') as title,
  JSON_EXTRACT(event_data, '$.summary') as summary,
  JSON_EXTRACT(event_data, '$.video_store_url') as video_url,
  JSON_EXTRACT(event_data, '$.key_words') as keywords,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
ORDER BY id DESC
LIMIT 1;
```

**预期结果：**
- `stream_id` = "0202"
- `domain_name` = "upload.skyzhou.cn"
- `app_name` = "live"
- `event_time` = 1770026056
- `title` = "新法规解读：Hima技术规范与全球标准"
- `summary` = "官方解读Hima高性能..."
- `video_url` = "http://video-1301212747.cos..."
- `keywords` = ["全球标准", "技术规范", "新法规", "高性能"]

### 3. 验证前端显示

1. 打开高光切片列表页面
2. 查询数据
3. 验证表格中显示的数据是否正确
4. 检查时间、URL、标题等字段

## 日志输出

修改后的日志会显示更详细的信息：

```
INFO  收到高光切片事件  streamId=0202  domain=upload.skyzhou.cn  itemCount=1
INFO  保存高光切片成功  itemIndex=0  title=新法规解读：Hima技术规范与全球标准  videoUrl=http://video-1301212747.cos...
```

## 注意事项

1. **多个切片处理**：一个请求可能包含多个高光切片，每个切片会创建一条独立的数据库记录

2. **时间戳单位**：`begin_time` 和 `end_time` 是 Unix 时间戳（秒），前端需要乘以 1000 转换为毫秒

3. **签名验证**：签名字段是可选的，只有在提供了签名时才进行验证

4. **字段兼容性**：`event_data` 中同时保存新旧字段名，确保前端兼容

5. **空值处理**：使用 `!= null` 检查，避免将 `0` 当作空值过滤掉

## 相关文件

- `sky-server/api/handler/live_callback_handler.go` - 结构体定义
- `sky-server/api/handler/live_callback_handler_ext.go` - 处理逻辑
- `sky-web/src/pages/LiveHighlightClips.vue` - 前端页面

## 更新日期

2026-02-02
