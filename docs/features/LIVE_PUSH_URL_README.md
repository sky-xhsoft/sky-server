# 腾讯云直播推流地址生成功能

## 概述

本次为项目添加了腾讯云直播推流地址生成和鉴权功能，支持生成带鉴权的推流地址和播放地址。

## 新增文件

### 1. 核心功能文件

- **`internal/pkg/tencent/live/push_url.go`**
  - 推流地址生成器核心实现
  - 实现腾讯云推流鉴权算法
  - 支持生成推流地址和播放地址
  - 支持验证推流地址有效性

### 2. API 处理器

- **`api/handler/live_push_url_handler.go`**
  - HTTP API 接口实现
  - 提供 POST 和 GET 两种方式生成推流地址
  - 提供推流地址验证接口

### 3. 测试文件

- **`internal/pkg/tencent/live/push_url_test.go`**
  - 单元测试
  - 覆盖主要功能场景

### 4. 文档和示例

- **`docs/live_push_url.md`**
  - 详细的功能说明文档
  - API 接口文档
  - 使用示例和注意事项

- **`examples/live_push_url_example.go`**
  - 完整的使用示例程序
  - 演示各种使用场景

## 主要功能

### 1. 生成推流地址

```go
generator := live.NewPushURLGenerator(pushDomain, streamKey, appName)

req := &live.GeneratePushURLRequest{
    StreamName: "test_stream_001",
    ExpireTime: 24 * time.Hour,
}

resp, err := generator.GeneratePushURL(req)
// resp.PushURL: 完整的 RTMP 推流地址
// resp.PushURLOBS: OBS 推流服务器地址
// resp.StreamKey: OBS 推流密钥
// resp.ExpireTime: 过期时间戳
```

### 2. 生成播放地址

```go
playURLs, err := generator.GeneratePlayURL(
    streamName,
    playDomain,
    playKey,  // 播放密钥，不需要鉴权时传空字符串
    24 * time.Hour,
)
// playURLs["rtmp"]: RTMP 播放地址
// playURLs["flv"]: FLV 播放地址
// playURLs["hls"]: HLS 播放地址
```

### 3. 验证推流地址

```go
isValid := generator.ValidatePushURL(streamName, txSecret, txTime)
```

## API 接口

### 1. 生成推流地址

**POST** `/api/v1/live/push/url`

```json
{
  "streamName": "test_stream_001",
  "expireHours": 24,
  "includePlayUrl": true
}
```

**GET** `/api/v1/live/push/url?streamName=test_stream_001&expireHours=24&includePlayUrl=true`

### 2. 验证推流地址

**POST** `/api/v1/live/push/validate`

```json
{
  "streamName": "test_stream_001",
  "txSecret": "abc123",
  "txTime": "5f8a1b2c"
}
```

## 配置说明

在 `config.yaml` 中添加以下配置：

```yaml
tencentCloud:
  secretID: "your_secret_id"
  secretKey: "your_secret_key"
  region: "ap-guangzhou"
  live:
    enabled: true
    pushDomain: "push.example.com"      # 推流域名
    playDomain: "play.example.com"      # 播放域名
    appName: "live"                     # 应用名称
    streamKey: "your_stream_key"        # 推流密钥
```

## 鉴权算法

腾讯云直播推流鉴权算法实现：

1. **计算过期时间戳 txTime**
   - 当前时间 + 过期时长
   - 转换为十六进制字符串

2. **计算签名 txSecret**
   - 签名字符串：`streamKey + streamName + txTime`
   - 计算 MD5 哈希值
   - 转换为十六进制字符串

3. **生成推流地址**
   - 格式：`rtmp://domain/AppName/StreamName?txSecret=xxx&txTime=xxx`

## 使用步骤

### 1. 配置腾讯云直播

1. 登录腾讯云控制台
2. 进入"云直播"服务
3. 添加推流域名和播放域名
4. 完成域名 CNAME 解析
5. 开启推流鉴权，获取推流密钥

### 2. 配置项目

在 `config.yaml` 中填入推流域名、播放域名和推流密钥。

### 3. 注册路由

在路由配置中添加推流地址生成接口：

```go
// 创建处理器
pushURLHandler := handler.NewLivePushURLHandler(cfg)

// 注册路由
liveGroup := router.Group("/api/v1/live")
{
    pushGroup := liveGroup.Group("/push")
    {
        pushGroup.POST("/url", pushURLHandler.GeneratePushURL)
        pushGroup.GET("/url", pushURLHandler.GetPushURL)
        pushGroup.POST("/validate", pushURLHandler.ValidatePushURL)
    }
}
```

### 4. 使用 API

调用 API 生成推流地址，然后在 OBS 或其他推流工具中配置推流。

## 测试

### 运行单元测试

```bash
cd internal/pkg/tencent/live
go test -v
```

### 运行示例程序

```bash
cd examples
go run live_push_url_example.go
```

### 使用 curl 测试 API

```bash
# 生成推流地址
curl -X POST http://localhost:8080/api/v1/live/push/url \
  -H "Content-Type: application/json" \
  -d '{
    "streamName": "test_stream_001",
    "expireHours": 24,
    "includePlayUrl": true
  }'
```

## OBS 推流配置

1. 打开 OBS Studio
2. 点击"设置" -> "推流"
3. 服务选择"自定义"
4. 服务器：填入 API 返回的 `pushUrlObs`
5. 串流密钥：填入 API 返回的 `streamKey`
6. 点击"确定"保存配置
7. 点击"开始推流"

## 注意事项

1. **推流密钥安全**
   - 推流密钥是敏感信息，不要泄露
   - 建议定期更换推流密钥
   - 不要将推流密钥提交到代码仓库

2. **域名配置**
   - 推流域名和播放域名需要完成 ICP 备案
   - 需要在腾讯云控制台完成域名配置
   - 需要配置 CNAME 解析

3. **过期时间**
   - 推流地址有效期建议设置为 24 小时
   - 过期后需要重新生成推流地址
   - 可以根据实际需求调整过期时间

4. **流名称规范**
   - 流名称只能包含字母、数字、下划线、连字符
   - 建议使用有意义的命名
   - 避免使用特殊字符

## 相关文档

- [腾讯云直播推流配置](https://cloud.tencent.com/document/product/267/32833)
- [腾讯云直播鉴权配置](https://cloud.tencent.com/document/product/267/32735)
- [OBS 推流指南](https://cloud.tencent.com/document/product/267/32726)

## 技术栈

- Go 1.x
- Gin Web Framework
- 腾讯云直播 SDK
- MD5 加密算法

## 作者

Sky-XH Software Team

## 许可证

本项目遵循项目主许可证。
