# 腾讯云直播回调事件接收器

## 概述

本模块实现了腾讯云直播回调事件的接收和处理功能，支持以下3种事件类型：

1. **推断流事件通知** - 推流和断流事件
2. **录制文件事件通知** - 录制文件生成事件
3. **录制状态事件通知** - 录制状态变更事件

## 功能特性

- ✅ 接收腾讯云直播回调事件
- ✅ 签名验证（可选）
- ✅ 事件数据持久化存储
- ✅ 事件查询接口
- ✅ 支持多种事件类型
- ✅ 详细的事件日志记录

## 数据库表结构

### live_callback_event 表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| ID | bigint | 主键ID |
| EVENT_TYPE | varchar(50) | 事件类型 |
| EVENT_TIME | bigint | 事件时间戳（秒） |
| DOMAIN_NAME | varchar(255) | 推流域名 |
| APP_NAME | varchar(255) | 应用名称 |
| STREAM_NAME | varchar(255) | 流名称 |
| STREAM_ID | varchar(255) | 流ID |
| CLIENT_IP | varchar(50) | 客户端IP |
| EVENT_DATA | text | 事件详细数据（JSON格式） |
| SIGN | varchar(255) | 签名 |
| T_VALUE | bigint | 签名过期时间 |
| CREATE_TIME | datetime | 创建时间 |
| SYS_COMPANY_ID | bigint | 公司ID |
| IS_ACTIVE | char(1) | 是否有效 |

## API 接口

### 1. 推流事件回调

**接口地址**: `POST /api/v1/live/callback/push`

**说明**: 接收腾讯云推流事件通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 1,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "sequence": "seq_001",
  "node": "192.168.1.1",
  "user_ip": "1.2.3.4",
  "stream_param": "",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

**响应示例**:
```json
{
  "code": 0
}
```

### 2. 断流事件回调

**接口地址**: `POST /api/v1/live/callback/disconnect`

**说明**: 接收腾讯云断流事件通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 0,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "sequence": "seq_001",
  "node": "192.168.1.1",
  "user_ip": "1.2.3.4",
  "stream_param": "",
  "duration": 3600,
  "reason": "normal",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 3. 录制文件事件回调

**接口地址**: `POST /api/v1/live/callback/recording-file`

**说明**: 接收腾讯云录制文件生成通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 100,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "video_url": "https://example.com/record.flv",
  "file_size": 1024000,
  "duration": 3600,
  "file_format": "flv",
  "start_time": 1234567890,
  "end_time": 1234571490,
  "stream_param": "",
  "video_id": "video_001",
  "record_file_id": "file_001",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 4. 录制状态事件回调

**接口地址**: `POST /api/v1/live/callback/recording-status`

**说明**: 接收腾讯云录制状态变更通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 200,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "status": 1,
  "stream_param": "",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 5. 查询回调事件列表

**接口地址**: `GET /api/v1/live/callback/events`

**说明**: 查询直播回调事件记录（需要认证）

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

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "eventType": "push_stream",
        "eventTime": 1234567890,
        "domainName": "push.example.com",
        "appName": "live",
        "streamName": "test_stream",
        "streamId": "test_stream_001",
        "clientIp": "1.2.3.4",
        "eventData": "{...}",
        "createTime": "2026-02-02T10:00:00Z"
      }
    ],
    "total": 100,
    "pageNum": 1,
    "pageSize": 20
  }
}
```

## 配置说明

### 1. 数据库迁移

执行以下SQL创建回调事件表：

```bash
mysql -u root -p your_database < sqls/migrations/create_live_callback_event.sql
```

### 2. 配置文件

在 `configs/config.yaml` 中添加回调密钥配置：

```yaml
tencentCloud:
  secretID: "your_secret_id"
  secretKey: "your_secret_key"
  region: "ap-guangzhou"
  callbackKey: "your_callback_key"  # 回调密钥，用于验证签名
  live:
    enabled: true
    pushDomain: "push.example.com"
    playDomain: "play.example.com"
    appName: "live"
    streamKey: "your_stream_key"
    callbackURL: "https://your-domain.com/api/v1/live/callback"
```

### 3. 腾讯云控制台配置

在腾讯云直播控制台配置回调URL：

1. 登录 [腾讯云直播控制台](https://console.cloud.tencent.com/live)
2. 进入【功能配置】>【事件中心】>【直播回调】
3. 配置回调URL：
   - 推流回调URL: `https://your-domain.com/api/v1/live/callback/push`
   - 断流回调URL: `https://your-domain.com/api/v1/live/callback/disconnect`
   - 录制回调URL: `https://your-domain.com/api/v1/live/callback/recording-file`
   - 录制状态回调URL: `https://your-domain.com/api/v1/live/callback/recording-status`
4. 配置回调密钥（与config.yaml中的callbackKey保持一致）

## 签名验证

回调签名验证算法：

```
sign = MD5(callbackKey + t)
```

其中：
- `callbackKey`: 配置文件中的回调密钥
- `t`: 签名过期时间戳（秒）

如果不需要签名验证，可以将 `callbackKey` 配置为空字符串。

## 事件类型说明

| 事件类型 | event_type | 说明 |
|----------|-----------|------|
| push_stream | 1 | 推流事件 |
| disconnect_stream | 0 | 断流事件 |
| recording_file | 100 | 录制文件生成 |
| recording_status | 200 | 录制状态变更 |

## 日志记录

所有回调事件都会记录详细日志，包括：
- 事件类型
- 流信息（域名、应用名、流名称）
- 客户端IP
- 事件详细数据

日志级别：
- INFO: 正常事件接收
- WARN: 签名验证失败
- ERROR: 数据解析失败、数据库保存失败

## 故障排查

### 1. 回调未收到

- 检查腾讯云控制台回调URL配置是否正确
- 检查服务器防火墙是否开放对应端口
- 检查服务器日志是否有错误信息

### 2. 签名验证失败

- 检查 `callbackKey` 配置是否与腾讯云控制台一致
- 检查时间戳是否过期（默认10分钟有效期）

### 3. 数据未保存

- 检查数据库连接是否正常
- 检查 `live_callback_event` 表是否已创建
- 查看服务器日志中的错误信息

## 参考文档

- [腾讯云直播回调事件通知](https://cloud.tencent.com/document/product/267/32744)
- [腾讯云直播API文档](https://cloud.tencent.com/document/product/267/20456)
