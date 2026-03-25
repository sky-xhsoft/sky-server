# 腾讯云直播推流地址生成功能

## 功能说明

本功能实现了腾讯云直播推流地址的生成和鉴权，支持：

1. 生成带鉴权的RTMP推流地址
2. 生成OBS推流配置（服务器地址 + 推流密钥）
3. 生成播放地址（RTMP、FLV、HLS）
4. 验证推流地址的有效性

## 配置说明

在 `config.yaml` 中添加以下配置：

```yaml
tencentCloud:
  secretID: "your_secret_id"
  secretKey: "your_secret_key"
  region: "ap-guangzhou"
  callbackKey: "your_callback_key"
  live:
    enabled: true
    pushDomain: "push.example.com"      # 推流域名
    playDomain: "play.example.com"      # 播放域名
    appName: "live"                     # 应用名称，默认为 live
    streamKey: "your_stream_key"        # 推流密钥（在腾讯云控制台获取）
    recordBucket: "your-bucket"
    recordRegion: "ap-guangzhou"
    callbackURL: "https://your-domain.com/api/v1/live/callback"
```

## API 接口

### 1. 生成推流地址（POST）

**接口地址：** `POST /api/v1/live/push/url`

**请求参数：**

```json
{
  "streamName": "test_stream_001",    // 流名称（必填）
  "expireHours": 24,                  // 过期时间（小时），默认24小时
  "includePlayUrl": true              // 是否包含播放地址
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "pushUrl": "rtmp://push.example.com/live/test_stream_001?txSecret=abc123&txTime=5f8a1b2c",
    "pushUrlObs": "rtmp://push.example.com/live/",
    "streamKey": "test_stream_001?txSecret=abc123&txTime=5f8a1b2c",
    "expireTime": 1609459200,
    "playUrls": {
      "rtmp": "rtmp://play.example.com/live/test_stream_001",
      "flv": "http://play.example.com/live/test_stream_001.flv",
      "hls": "http://play.example.com/live/test_stream_001.m3u8"
    }
  }
}
```

### 2. 获取推流地址（GET）

**接口地址：** `GET /api/v1/live/push/url`

**请求参数：**

- `streamName`: 流名称（必填）
- `expireHours`: 过期时间（小时），默认24小时
- `includePlayUrl`: 是否包含播放地址

**示例：**

```
GET /api/v1/live/push/url?streamName=test_stream_001&expireHours=24&includePlayUrl=true
```

### 3. 验证推流地址

**接口地址：** `POST /api/v1/live/push/validate`

**请求参数：**

```json
{
  "streamName": "test_stream_001",
  "txSecret": "abc123",
  "txTime": "5f8a1b2c"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "valid": true
  }
}
```

## 使用示例

### 1. 在代码中使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

func main() {
    // 创建推流地址生成器
    generator := live.NewPushURLGenerator(
        "push.example.com",  // 推流域名
        "your_stream_key",   // 推流密钥
        "live",              // 应用名称
    )

    // 生成推流地址
    req := &live.GeneratePushURLRequest{
        StreamName: "test_stream_001",
        ExpireTime: 24 * time.Hour,
    }

    resp, err := generator.GeneratePushURL(req)
    if err != nil {
        fmt.Printf("生成推流地址失败: %v\n", err)
        return
    }

    fmt.Printf("推流地址: %s\n", resp.PushURL)
    fmt.Printf("OBS推流地址: %s\n", resp.PushURLOBS)
    fmt.Printf("OBS推流密钥: %s\n", resp.StreamKey)
    fmt.Printf("过期时间: %d\n", resp.ExpireTime)

    // 生成播放地址
    playURLs, err := generator.GeneratePlayURL(
        "test_stream_001",
        "play.example.com",
        "",  // 播放密钥，如果播放域名没有配置鉴权则为空
        24 * time.Hour,
    )
    if err != nil {
        fmt.Printf("生成播放地址失败: %v\n", err)
        return
    }

    fmt.Printf("RTMP播放地址: %s\n", playURLs["rtmp"])
    fmt.Printf("FLV播放地址: %s\n", playURLs["flv"])
    fmt.Printf("HLS播放地址: %s\n", playURLs["hls"])
}
```

### 2. 使用 curl 测试

```bash
# 生成推流地址
curl -X POST http://localhost:8080/api/v1/live/push/url \
  -H "Content-Type: application/json" \
  -d '{
    "streamName": "test_stream_001",
    "expireHours": 24,
    "includePlayUrl": true
  }'

# 获取推流地址（GET方式）
curl "http://localhost:8080/api/v1/live/push/url?streamName=test_stream_001&expireHours=24&includePlayUrl=true"

# 验证推流地址
curl -X POST http://localhost:8080/api/v1/live/push/validate \
  -H "Content-Type: application/json" \
  -d '{
    "streamName": "test_stream_001",
    "txSecret": "abc123",
    "txTime": "5f8a1b2c"
  }'
```

### 3. OBS 推流配置

使用生成的推流地址在 OBS 中配置：

1. 打开 OBS Studio
2. 点击 "设置" -> "推流"
3. 服务选择 "自定义"
4. 服务器：填入 `pushUrlObs`（例如：`rtmp://push.example.com/live/`）
5. 串流密钥：填入 `streamKey`（例如：`test_stream_001?txSecret=abc123&txTime=5f8a1b2c`）
6. 点击 "确定" 保存配置
7. 点击 "开始推流"

## 鉴权算法说明

腾讯云直播推流鉴权算法：

1. **计算过期时间戳 txTime**
   - 当前时间 + 过期时长
   - 转换为十六进制字符串

2. **计算签名 txSecret**
   - 签名字符串：`streamKey + streamName + txTime`
   - 计算 MD5 哈希值
   - 转换为十六进制字符串

3. **生成推流地址**
   - 格式：`rtmp://domain/AppName/StreamName?txSecret=xxx&txTime=xxx`

## 注意事项

1. **推流密钥获取**
   - 登录腾讯云控制台
   - 进入 "云直播" -> "域名管理"
   - 选择推流域名，点击 "管理"
   - 在 "推流配置" 中查看或生成推流密钥

2. **域名配置**
   - 推流域名和播放域名需要在腾讯云控制台完成配置
   - 域名需要完成 ICP 备案
   - 需要配置 CNAME 解析

3. **鉴权配置**
   - 推流域名必须开启推流鉴权
   - 播放域名可选择是否开启播放鉴权
   - 如果播放域名开启了鉴权，需要在生成播放地址时传入播放密钥

4. **过期时间**
   - 推流地址有效期建议设置为 24 小时
   - 过期后需要重新生成推流地址
   - 可以根据实际需求调整过期时间

5. **流名称规范**
   - 流名称只能包含字母、数字、下划线、连字符
   - 建议使用有意义的命名，如：`user_123_live`
   - 避免使用特殊字符

## 测试

运行单元测试：

```bash
cd internal/pkg/tencent/live
go test -v
```

## 相关文档

- [腾讯云直播推流配置](https://cloud.tencent.com/document/product/267/32833)
- [腾讯云直播鉴权配置](https://cloud.tencent.com/document/product/267/32735)
- [OBS 推流指南](https://cloud.tencent.com/document/product/267/32726)
