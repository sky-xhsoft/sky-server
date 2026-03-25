# 录制列表空数据问题修复

## 问题描述

用户报告录制列表页面显示多个空列，包括：
- 流名称
- 域名
- 应用名称
- 文件格式
- 文件大小
- 录制时长
- 录制时间

## 问题原因

### 1. 数据库中的事件数据不完整

通过控制台日志发现，数据库中存储的 `eventData` 只包含最基本的字段：
```json
{
  "event_type": 100,
  "stream_id": "0202"
}
```

而前端期望的完整字段包括：
```json
{
  "event_type": 100,
  "stream_id": "test_stream_001",
  "video_url": "https://example.com/record.flv",
  "file_size": 1024000,
  "duration": 3600,
  "file_format": "flv",
  "start_time": 1234567890,
  "end_time": 1234571490,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 2. 可能的原因

1. **测试数据不完整**：数据库中的数据是通过简化的测试脚本创建的，没有包含所有字段
2. **腾讯云回调未配置**：实际的腾讯云录制回调还未配置或触发
3. **回调数据格式变化**：腾讯云的实际回调格式可能与文档不一致

## 解决方案

### 前端修复（已完成）

更新 `LiveRecordings.vue` 以更好地处理缺失数据：

1. **增强字段提取逻辑**：
   - 支持多种可能的字段名（snake_case 和 camelCase）
   - 从 `eventData` 和 `item` 的多个位置尝试获取数据
   - 使用空值合并运算符 `??` 处理 0 值

```typescript
// 从 eventData 中提取字段，支持多种可能的字段名
const videoUrl = eventData.video_url || eventData.videoUrl || ''
const fileSize = eventData.file_size ?? eventData.fileSize ?? 0
const duration = eventData.duration ?? 0
const fileFormat = eventData.file_format || eventData.fileFormat || ''

// 如果 eventData 中没有这些字段，尝试从 item 的其他字段获取
const streamName = item.streamName || eventData.stream_name || eventData.streamName || ''
const domainName = item.domainName || eventData.push_domain || eventData.pushDomain || ''
const appName = item.appName || eventData.app_name || eventData.appName || ''
```

2. **改进空值显示**：
   - 文件大小为 0 或 null 时显示 "-" 而不是 "0 B"
   - 时长为 0 或 null 时显示 "-" 而不是 "0秒"
   - 空字符串字段显示 "-"
   - 文件格式为空时显示 "-" 而不是空的标签

3. **添加调试日志**：
```typescript
console.log('Available eventData keys:', Object.keys(eventData))
```

### 后端验证（需要检查）

需要验证后端是否正确保存了完整的事件数据：

1. **检查 `HandleRecordingFile` 函数**：
   - 确认是否正确解析了所有字段
   - 确认 `json.Marshal(event)` 是否包含所有字段

2. **检查数据库表结构**：
   - `live_callback_event` 表的 `event_data` 字段是否足够大
   - 是否有字段长度限制导致数据被截断

3. **检查实际回调数据**：
   - 查看后端日志，确认腾讯云发送的实际数据格式
   - 对比文档和实际数据的差异

## 测试方法

### 1. 使用完整测试数据

创建包含所有字段的测试数据：

```bash
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_001",
    "t": '$(date +%s)',
    "sign": "test",
    "event_time": '$(date +%s)',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "start_time": '$(date +%s)',
    "end_time": '$(date +%s)',
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream",
    "record_file_id": "file_001",
    "video_id": "video_001"
  }'
```

### 2. 检查数据库

```sql
SELECT id, stream_id, stream_name, domain_name, app_name,
       LENGTH(event_data) as data_length,
       event_data
FROM live_callback_event
WHERE event_type = 'recording_file'
ORDER BY id DESC
LIMIT 5;
```

### 3. 查看前端控制台

打开浏览器控制台，查看以下日志：
- `Available eventData keys:` - 显示实际可用的字段
- `Parsed recording eventData:` - 显示解析后的完整数据

## 预期结果

修复后，即使数据不完整，页面也应该：
1. 显示 "-" 而不是空白单元格
2. 正确显示可用的字段
3. 不会因为缺失字段而报错

## 后续改进建议

1. **后端数据验证**：
   - 在保存事件前验证必需字段
   - 记录警告日志如果字段缺失

2. **前端数据提示**：
   - 在列表顶部显示数据完整性提示
   - 如果大部分记录缺少关键字段，提示用户检查配置

3. **配置检查工具**：
   - 添加腾讯云回调配置检查功能
   - 提供测试回调的工具

## 相关文件

- `sky-web/src/pages/LiveRecordings.vue` - 录制列表页面（已修复）
- `sky-server/api/handler/live_callback_handler.go` - 录制回调处理器
- `sky-server/docs/live_callback_TEST.md` - 测试文档
- `sky-server/docs/live_callback_README.md` - API 文档
