# 腾讯云直播 API 集成文档

## 概述

本项目已集成腾讯云直播 API，支持以下四大功能模块：

1. **域名管理** - 添加、删除、查询、启用/禁用直播域名
2. **拉流管理** - 创建、删除、查询、重启拉流任务
3. **审核管理** - 创建关键词库、管理审核关键词
4. **录制管理** - 创建录制模板、规则和任务

## 架构设计

```
internal/
├── config/
│   └── config.go                    # 配置定义（已扩展腾讯云配置）
├── pkg/
│   └── tencent/
│       ├── client.go                # 腾讯云客户端封装
│       └── live/
│           ├── domain.go            # 域名管理
│           ├── pull_stream.go       # 拉流管理
│           ├── audit.go             # 审核管理
│           └── record.go            # 录制管理
└── service/
    └── live/
        └── live_service.go          # 直播业务服务层
```

## 配置说明

在 `configs/config.yaml` 中添加腾讯云配置：

```yaml
tencentCloud:
  secretID: "your-tencent-secret-id"
  secretKey: "your-tencent-secret-key"
  region: "ap-guangzhou"

  live:
    enabled: true
    pushDomain: "push.yourdomain.com"
    playDomain: "play.yourdomain.com"
    appName: "live"
    streamKey: "your-stream-key"
    recordBucket: "your-record-bucket"
    recordRegion: "ap-guangzhou"
    callbackURL: "https://your-domain.com/api/live/callback"
```

## 使用示例

### 1. 初始化服务

```go
package main

import (
    "context"
    "log"

    "github.com/sky-xhsoft/sky-server/internal/config"
    "github.com/sky-xhsoft/sky-server/internal/service/live"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // 创建直播服务
    liveService, err := live.NewService(&cfg.TencentCloud)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // 使用服务...
}
```

### 2. 域名管理示例

```go
import (
    tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

// 添加推流域名
err := liveService.AddDomain(ctx, &tencentLive.AddDomainRequest{
    DomainName: "push.yourdomain.com",
    DomainType: 0, // 0-推流域名，1-播放域名
})

// 查询域名列表
domains, err := liveService.ListDomains(ctx, nil)
for _, domain := range domains {
    log.Printf("域名: %s, 状态: %d", domain.Name, domain.Status)
}

// 启用域名
err = liveService.EnableDomain(ctx, "push.yourdomain.com")

// 禁用域名
err = liveService.ForbidDomain(ctx, "push.yourdomain.com")

// 删除域名
err = liveService.DeleteDomain(ctx, "push.yourdomain.com")
```

### 3. 拉流管理示例

```go
// 创建拉流任务
taskID, err := liveService.CreatePullStreamTask(ctx, &tencentLive.CreatePullStreamTaskRequest{
    SourceType: "PullLivePushLive",
    SourceURLs: []string{"rtmp://source.example.com/live/stream"},
    DomainName: "push.yourdomain.com",
    AppName:    "live",
    StreamName: "test_stream",
    StartTime:  "2026-01-30 10:00:00",
    EndTime:    "2026-01-30 12:00:00",
    Operator:   "admin",
    Comment:    "测试拉流任务",
})

// 查询拉流任务
tasks, err := liveService.DescribePullStreamTasks(ctx, &taskID)
for _, task := range tasks {
    log.Printf("任务ID: %s, 状态: %s", task.TaskID, task.Status)
}

// 重启拉流任务
err = liveService.RestartPullStreamTask(ctx, taskID, "admin")

// 删除拉流任务
err = liveService.DeletePullStreamTask(ctx, taskID, "admin")
```

### 4. 审核管理示例

```go
// 创建关键词库
libID, err := liveService.CreateKeywordLib(ctx, &tencentLive.CreateKeywordLibRequest{
    LibName: "敏感词库",
    LibType: 1, // 1-黑名单，2-白名单
})

// 添加关键词
err = liveService.CreateKeywords(ctx, &tencentLive.CreateKeywordsRequest{
    LibID:    libID,
    Keywords: []string{"违禁词1", "违禁词2"},
})

// 查询关键词列表
keywords, err := liveService.DescribeKeywords(ctx, libID)
for _, kw := range keywords {
    log.Printf("关键词: %s, 创建时间: %s", kw.Keyword, kw.CreateTime)
}

// 删除关键词
err = liveService.DeleteKeywords(ctx, libID, []string{"违禁词1"})
```

### 5. 录制管理示例

```go
// 创建录制模板
templateID, err := liveService.CreateRecordTemplate(ctx, &tencentLive.CreateRecordTemplateRequest{
    TemplateName: "标准录制模板",
    Description:  "HLS格式录制",
    HlsParam: &tencentLive.RecordParam{
        RecordInterval: 3600,  // 录制间隔（秒）
        StorageTime:    86400, // 存储时长（秒）
        Enable:         1,     // 启用
    },
})

// 创建录制规则
err = liveService.CreateRecordRule(ctx, &tencentLive.CreateRecordRuleRequest{
    DomainName: "push.yourdomain.com",
    AppName:    "live",
    StreamName: "test_stream",
    TemplateID: templateID,
})

// 查询录制模板列表
templates, err := liveService.DescribeRecordTemplates(ctx)
for _, tpl := range templates {
    log.Printf("模板ID: %d, 名称: %s", tpl.TemplateID, tpl.TemplateName)
}

// 创建录制任务
taskID, err := liveService.CreateRecordTask(ctx, &tencentLive.CreateRecordTaskRequest{
    StreamName: "test_stream",
    DomainName: "push.yourdomain.com",
    AppName:    "live",
    StartTime:  1706601600, // Unix时间戳
    EndTime:    1706605200,
    StreamType: 0,
    TemplateID: templateID,
    Comment:    "测试录制任务",
})

// 停止录制任务
err = liveService.StopRecordTask(ctx, taskID)

// 删除录制任务
err = liveService.DeleteRecordTask(ctx, taskID)

// 删除录制规则
err = liveService.DeleteRecordRule(ctx, "push.yourdomain.com", "live", "test_stream")

// 删除录制模板
err = liveService.DeleteRecordTemplate(ctx, templateID)
```

## API 接口说明

### 域名管理接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `AddDomain` | 添加域名 | 域名名称、域名类型 |
| `DeleteDomain` | 删除域名 | 域名名称 |
| `DescribeDomain` | 查询域名信息 | 域名名称 |
| `ListDomains` | 查询域名列表 | 域名类型（可选） |
| `EnableDomain` | 启用域名 | 域名名称 |
| `ForbidDomain` | 禁用域名 | 域名名称 |

### 拉流管理接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `CreatePullStreamTask` | 创建拉流任务 | 源类型、源URL、推流域名等 |
| `DeletePullStreamTask` | 删除拉流任务 | 任务ID、操作者 |
| `DescribePullStreamTasks` | 查询拉流任务 | 任务ID（可选） |
| `RestartPullStreamTask` | 重启拉流任务 | 任务ID、操作者 |

### 审核管理接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `CreateKeywordLib` | 创建关键词库 | 词库名称、词库类型 |
| `CreateKeywords` | 创建关键词 | 词库ID、关键词列表 |
| `DeleteKeywords` | 删除关键词 | 词库ID、关键词列表 |
| `DescribeKeywords` | 查询关键词 | 词库ID |

### 录制管理接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `CreateRecordTemplate` | 创建录制模板 | 模板名称、录制参数 |
| `DeleteRecordTemplate` | 删除录制模板 | 模板ID |
| `DescribeRecordTemplates` | 查询录制模板列表 | 无 |
| `CreateRecordRule` | 创建录制规则 | 域名、应用名、流名、模板ID |
| `DeleteRecordRule` | 删除录制规则 | 域名、应用名、流名 |
| `CreateRecordTask` | 创建录制任务 | 流名、域名、时间等 |
| `StopRecordTask` | 停止录制任务 | 任务ID |
| `DeleteRecordTask` | 删除录制任务 | 任务ID |

## 错误处理

所有方法都返回 `error` 类型，建议统一处理：

```go
if err != nil {
    log.Printf("操作失败: %v", err)
    // 根据业务需求处理错误
    return
}
```

## 注意事项

1. **API 密钥安全**：不要将 SecretID 和 SecretKey 提交到代码仓库
2. **区域选择**：根据业务需求选择合适的区域（ap-guangzhou、ap-shanghai 等）
3. **域名备案**：直播域名需要完成 ICP 备案
4. **费用控制**：注意监控 API 调用量和直播流量，避免产生意外费用
5. **并发控制**：腾讯云 API 有频率限制，建议添加重试和限流机制

## 下一步

1. 根据业务需求创建 HTTP API 接口（在 `api/` 目录）
2. 添加数据库实体和持久化逻辑（在 `internal/model/entity/` 目录）
3. 实现直播流状态监控和回调处理
4. 添加单元测试和集成测试

## 参考文档

- [腾讯云直播 API 文档](https://cloud.tencent.com/document/product/267/20456)
- [腾讯云 Go SDK](https://github.com/tencentcloud/tencentcloud-sdk-go)
