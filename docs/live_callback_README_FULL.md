# 腾讯云直播回调事件接收器 - 完整版

## 概述

本模块实现了腾讯云直播回调事件的接收和处理功能，支持以下17种事件类型：

### 基础事件
1. **推流事件通知** - 推流开始事件
2. **断流事件通知** - 推流结束事件

### 录制相关事件
3. **录制文件事件通知** - 录制文件生成事件
4. **录制状态事件通知** - 录制状态变更事件
5. **录制异常事件通知** - 录制过程异常事件

### 截图和审核事件
6. **截图事件通知** - 直播截图生成事件
7. **画面审核事件通知** - 视频内容审核事件
8. **音频审核事件通知** - 音频内容审核事件

### 质检和评测事件
9. **质检事件通知** - 直播质量检测事件
10. **评测阈值事件通知** - 质量指标超阈值事件
11. **评测平均分事件通知** - 质量平均分统计事件

### AI功能事件
12. **智能擦除事件通知** - 智能擦除任务完成事件
13. **直播字幕事件通知** - 实时字幕生成事件
14. **直播摘要事件通知** - 直播内容摘要事件
15. **高光切片事件通知** - 精彩片段识别事件

### 异常和监控事件
16. **推流异常事件通知** - 推流过程异常事件
17. **拉流转推事件通知** - 拉流转推状态变更事件
18. **监播事件通知** - 直播监控告警事件

## 功能特性

- ✅ 支持18种腾讯云直播回调事件
- ✅ 签名验证（可选）
- ✅ 事件数据持久化存储
- ✅ 事件查询接口
- ✅ 详细的事件日志记录
- ✅ 统一的错误处理

## API 接口列表

### 基础事件

#### 1. 推流事件回调
**接口**: `POST /api/v1/live/callback/push`

#### 2. 断流事件回调
**接口**: `POST /api/v1/live/callback/disconnect`

### 录制相关

#### 3. 录制文件事件回调
**接口**: `POST /api/v1/live/callback/recording-file`

#### 4. 录制状态事件回调
**接口**: `POST /api/v1/live/callback/recording-status`

#### 5. 录制异常事件回调
**接口**: `POST /api/v1/live/callback/record-exception`

### 截图和审核

#### 6. 截图事件回调
**接口**: `POST /api/v1/live/callback/screenshot`

**请求示例**:
```json
{
  "event_type": 200,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "pic_url": "https://example.com/screenshot.jpg",
  "width": 1920,
  "height": 1080,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 7. 画面审核事件回调
**接口**: `POST /api/v1/live/callback/video-audit`

**请求示例**:
```json
{
  "event_type": 317,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "confidence": 95,
  "label": "porn",
  "suggestion": "block",
  "screenshot_url": "https://example.com/audit.jpg",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 8. 音频审核事件回调
**接口**: `POST /api/v1/live/callback/audio-audit`

**请求示例**:
```json
{
  "event_type": 318,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "confidence": 90,
  "label": "abuse",
  "suggestion": "review",
  "audio_text": "转写的音频文本内容",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 质检和评测

#### 9. 质检事件回调
**接口**: `POST /api/v1/live/callback/quality-inspection`

**请求示例**:
```json
{
  "event_type": 319,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "diagnose_type": "video_freeze",
  "level": "warning",
  "description": "检测到视频卡顿",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 10. 评测阈值事件回调
**接口**: `POST /api/v1/live/callback/quality-threshold`

**请求示例**:
```json
{
  "event_type": 320,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "metric_type": "bitrate",
  "threshold": 2000.0,
  "current_value": 1500.0,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 11. 评测平均分事件回调
**接口**: `POST /api/v1/live/callback/quality-average`

**请求示例**:
```json
{
  "event_type": 321,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "score": 85.5,
  "duration": 3600,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### AI功能

#### 12. 智能擦除事件回调
**接口**: `POST /api/v1/live/callback/smart-erase`

**请求示例**:
```json
{
  "event_type": 322,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "task_id": "task_001",
  "status": "success",
  "output_url": "https://example.com/erased.mp4",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 13. 直播字幕事件回调
**接口**: `POST /api/v1/live/callback/subtitle`

**请求示例**:
```json
{
  "event_type": 323,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "text": "这是实时生成的字幕内容",
  "language": "zh-CN",
  "start_time": 1234567890,
  "end_time": 1234567895,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 14. 直播摘要事件回调
**接口**: `POST /api/v1/live/callback/summary`

**请求示例**:
```json
{
  "event_type": 324,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "summary": "本次直播的主要内容摘要",
  "keywords": ["关键词1", "关键词2", "关键词3"],
  "duration": 3600,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 15. 高光切片事件回调
**接口**: `POST /api/v1/live/callback/highlight`

**请求示例**:
```json
{
  "event_type": 325,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "clip_url": "https://example.com/highlight.mp4",
  "start_time": 1234567890,
  "end_time": 1234567920,
  "score": 9.5,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 异常和监控

#### 16. 推流异常事件回调
**接口**: `POST /api/v1/live/callback/push-exception`

**请求示例**:
```json
{
  "event_type": 326,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "error_code": 1001,
  "error_msg": "推流连接超时",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 17. 拉流转推事件回调
**接口**: `POST /api/v1/live/callback/pull-stream`

**请求示例**:
```json
{
  "event_type": 328,
  "task_id": "task_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "status": "start",
  "source_url": "rtmp://source.example.com/live/stream",
  "target_url": "rtmp://target.example.com/live/stream",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 18. 监播事件回调
**接口**: `POST /api/v1/live/callback/monitor`

**请求示例**:
```json
{
  "event_type": 329,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "alert_type": "black_screen",
  "alert_level": "error",
  "description": "检测到黑屏",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 查询接口

#### 查询回调事件列表
**接口**: `GET /api/v1/live/callback/events`（需要认证）

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| eventType | string | 否 | 事件类型 |
| streamId | string | 否 | 流ID |
| domainName | string | 否 | 域名 |
| appName | string | 否 | 应用名称 |
| streamName | string | 否 | 流名称 |
| startTime | string | 否 | 开始时间 |
| endTime | string | 否 | 结束时间 |
| pageNum | int | 否 | 页码（默认1） |
| pageSize | int | 否 | 每页大小（默认20） |

## 事件类型映射表

| 事件类型 | event_type | 数据库存储名称 | 说明 |
|----------|-----------|---------------|------|
| 推流事件 | 1 | push_stream | 推流开始 |
| 断流事件 | 0 | disconnect_stream | 推流结束 |
| 录制文件 | 100 | recording_file | 录制文件生成 |
| 录制状态 | 200 | recording_status | 录制状态变更 |
| 截图事件 | 200 | screenshot | 截图生成 |
| 画面审核 | 317 | video_audit | 视频内容审核 |
| 音频审核 | 318 | audio_audit | 音频内容审核 |
| 质检事件 | 319 | quality_inspection | 质量检测 |
| 评测阈值 | 320 | quality_threshold | 阈值告警 |
| 评测平均分 | 321 | quality_average | 平均分统计 |
| 智能擦除 | 322 | smart_erase | 智能擦除完成 |
| 直播字幕 | 323 | subtitle | 实时字幕 |
| 直播摘要 | 324 | summary | 内容摘要 |
| 高光切片 | 325 | highlight | 精彩片段 |
| 推流异常 | 326 | push_exception | 推流异常 |
| 录制异常 | 327 | record_exception | 录制异常 |
| 拉流转推 | 328 | pull_stream | 拉流转推状态 |
| 监播事件 | 329 | monitor | 监控告警 |

## 配置说明

### 1. 数据库迁移

```bash
mysql -u root -p your_database < sqls/migrations/create_live_callback_event.sql
```

### 2. 配置文件

在 `configs/config.yaml` 中配置：

```yaml
tencentCloud:
  secretID: "your_secret_id"
  secretKey: "your_secret_key"
  region: "ap-guangzhou"
  callbackKey: "your_callback_key"  # 回调密钥
  live:
    enabled: true
    pushDomain: "push.example.com"
    playDomain: "play.example.com"
    appName: "live"
    streamKey: "your_stream_key"
```

### 3. 腾讯云控制台配置

在腾讯云直播控制台配置各类回调URL：

**基础事件**:
- 推流回调: `https://your-domain.com/api/v1/live/callback/push`
- 断流回调: `https://your-domain.com/api/v1/live/callback/disconnect`

**录制相关**:
- 录制文件: `https://your-domain.com/api/v1/live/callback/recording-file`
- 录制状态: `https://your-domain.com/api/v1/live/callback/recording-status`
- 录制异常: `https://your-domain.com/api/v1/live/callback/record-exception`

**截图和审核**:
- 截图: `https://your-domain.com/api/v1/live/callback/screenshot`
- 画面审核: `https://your-domain.com/api/v1/live/callback/video-audit`
- 音频审核: `https://your-domain.com/api/v1/live/callback/audio-audit`

**质检和评测**:
- 质检: `https://your-domain.com/api/v1/live/callback/quality-inspection`
- 评测阈值: `https://your-domain.com/api/v1/live/callback/quality-threshold`
- 评测平均分: `https://your-domain.com/api/v1/live/callback/quality-average`

**AI功能**:
- 智能擦除: `https://your-domain.com/api/v1/live/callback/smart-erase`
- 直播字幕: `https://your-domain.com/api/v1/live/callback/subtitle`
- 直播摘要: `https://your-domain.com/api/v1/live/callback/summary`
- 高光切片: `https://your-domain.com/api/v1/live/callback/highlight`

**异常和监控**:
- 推流异常: `https://your-domain.com/api/v1/live/callback/push-exception`
- 拉流转推: `https://your-domain.com/api/v1/live/callback/pull-stream`
- 监播: `https://your-domain.com/api/v1/live/callback/monitor`

## 使用场景

### 1. 内容审核
通过画面审核和音频审核事件，实时监控直播内容，自动识别违规内容。

### 2. 质量监控
通过质检、评测阈值和评测平均分事件，实时监控直播质量，及时发现和处理质量问题。

### 3. 智能运营
通过字幕、摘要和高光切片事件，自动生成直播内容的文字记录、摘要和精彩片段，提升运营效率。

### 4. 异常告警
通过推流异常、录制异常和监播事件，及时发现和处理直播过程中的各类异常。

### 5. 数据分析
通过查询接口，分析直播数据，生成运营报表。

## 故障排查

### 1. 回调未收到
- 检查腾讯云控制台回调URL配置
- 检查服务器防火墙设置
- 查看服务器日志

### 2. 签名验证失败
- 检查 `callbackKey` 配置
- 检查时间戳是否过期

### 3. 数据未保存
- 检查数据库连接
- 检查表是否已创建
- 查看错误日志

## 参考文档

- [腾讯云直播回调事件通知](https://cloud.tencent.com/document/product/267/32744)
- [腾讯云直播API文档](https://cloud.tencent.com/document/product/267/20456)
