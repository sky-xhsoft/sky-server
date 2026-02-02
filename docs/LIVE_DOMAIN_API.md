# 腾讯云直播域名管理 API 文档

## 概述

本文档描述腾讯云直播域名管理的 HTTP API 接口。

## 已实现的接口

### 1. 添加域名

**接口地址**: `POST /api/v1/live/domains`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "domainName": "push.yourdomain.com",
  "domainType": 0
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domainName | string | 是 | 域名 |
| domainType | int | 是 | 域名类型：0-推流域名，1-播放域名 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "domainName": "push.yourdomain.com",
    "message": "添加域名成功"
  },
  "timestamp": "2026-01-30T22:00:00Z"
}
```

---

### 2. 查询域名列表

**接口地址**: `GET /api/v1/live/domains`

**请求头**:
```
Authorization: Bearer {token}
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domainType | int | 否 | 域名类型：0-推流域名，1-播放域名 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "domains": [
      {
        "name": "push.yourdomain.com",
        "type": 0,
        "status": 1,
        "createTime": "2026-01-30 10:00:00",
        "updateTime": "2026-01-30 10:00:00"
      }
    ],
    "total": 1
  },
  "timestamp": "2026-01-30T22:00:00Z"
}
```

---

### 3. 查询域名信息

**接口地址**: `GET /api/v1/live/domains/{domainName}`

**请求头**:
```
Authorization: Bearer {token}
```

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domainName | string | 是 | 域名 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "name": "push.yourdomain.com",
    "type": 0,
    "status": 1,
    "createTime": "2026-01-30 10:00:00",
    "updateTime": "2026-01-30 10:00:00"
  },
  "timestamp": "2026-01-30T22:00:00Z"
}
```

---

### 4. 删除域名

**接口地址**: `DELETE /api/v1/live/domains/{domainName}`

**请求头**:
```
Authorization: Bearer {token}
```

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domainName | string | 是 | 域名 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "domainName": "push.yourdomain.com",
    "message": "删除域名成功"
  },
  "timestamp": "2026-01-30T22:00:00Z"
}
```

---

### 5. 启用域名

**接口地址**: `POST /api/v1/live/domains/{domainName}/enable`

**请求头**:
```
Authorization: Bearer {token}
```

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domainName | string | 是 | 域名 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "domainName": "push.yourdomain.com",
    "message": "启用域名成功"
  },
  "timestamp": "2026-01-30T22:00:00Z"
}
```

---

### 6. 禁用域名

**接口地址**: `POST /api/v1/live/domains/{domainName}/forbid`

**请求头**:
```
Authorization: Bearer {token}
```

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domainName | string | 是 | 域名 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "domainName": "push.yourdomain.com",
    "message": "禁用域名成功"
  },
  "timestamp": "2026-01-30T22:00:00Z"
}
```

---

## 错误响应

所有接口在发生错误时返回统一的错误格式：

```json
{
  "code": 400,
  "message": "请求参数错误: domainName is required",
  "timestamp": "2026-01-30T22:00:00Z"
}
```

**常见错误码**:
- `400` - 请求参数错误
- `401` - 未授权（token 无效或过期）
- `403` - 无权限
- `404` - 资源不存在
- `500` - 服务器内部错误

---

## 使用示例

### cURL 示例

```bash
# 1. 添加域名
curl -X POST http://localhost:9090/api/v1/live/domains \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "domainName": "push.yourdomain.com",
    "domainType": 0
  }'

# 2. 查询域名列表
curl -X GET "http://localhost:9090/api/v1/live/domains?domainType=0" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 3. 查询域名信息
curl -X GET http://localhost:9090/api/v1/live/domains/push.yourdomain.com \
  -H "Authorization: Bearer YOUR_TOKEN"

# 4. 启用域名
curl -X POST http://localhost:9090/api/v1/live/domains/push.yourdomain.com/enable \
  -H "Authorization: Bearer YOUR_TOKEN"

# 5. 禁用域名
curl -X POST http://localhost:9090/api/v1/live/domains/push.yourdomain.com/forbid \
  -H "Authorization: Bearer YOUR_TOKEN"

# 6. 删除域名
curl -X DELETE http://localhost:9090/api/v1/live/domains/push.yourdomain.com \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript 示例

```javascript
// 添加域名
async function addDomain() {
  const response = await fetch('http://localhost:9090/api/v1/live/domains', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      domainName: 'push.yourdomain.com',
      domainType: 0
    })
  });

  const data = await response.json();
  console.log(data);
}

// 查询域名列表
async function listDomains() {
  const response = await fetch('http://localhost:9090/api/v1/live/domains?domainType=0', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN'
    }
  });

  const data = await response.json();
  console.log(data);
}
```

---

## 注意事项

1. **认证要求**: 所有接口都需要在请求头中携带有效的 JWT token
2. **域名备案**: 添加的域名必须已完成 ICP 备案
3. **域名类型**: 推流域名和播放域名需要分别添加
4. **操作限制**: 删除域名前需要先禁用域名
5. **频率限制**: 腾讯云 API 有调用频率限制，建议添加重试机制

---

## 集成步骤

### 1. 在主路由中注册域名路由

编辑 `api/router/router.go`，在 `Setup` 函数中添加：

```go
// 注册直播域名路由
if cfg.TencentCloud.Live.Enabled {
    liveService, err := live.NewService(&cfg.TencentCloud)
    if err != nil {
        logger.Error("创建直播服务失败", zap.Error(err))
    } else {
        registerLiveDomainRoutes(v1, jwtUtil, liveService)
    }
}
```

### 2. 配置腾讯云密钥

在 `configs/config.yaml` 中配置：

```yaml
tencentCloud:
  secretID: "your-secret-id"
  secretKey: "your-secret-key"
  region: "ap-guangzhou"

  live:
    enabled: true
    pushDomain: "push.yourdomain.com"
    playDomain: "play.yourdomain.com"
```

### 3. 启动服务

```bash
cd sky-server
go run cmd/server/main.go
```

---

## 测试

使用 Postman 或其他 API 测试工具测试接口：

1. 先调用登录接口获取 token
2. 使用 token 调用域名管理接口
3. 验证返回结果是否符合预期

---

## 相关文档

- [腾讯云直播 API 文档](https://cloud.tencent.com/document/product/267/20456)
- [项目集成总结](./TENCENT_LIVE_INTEGRATION_SUMMARY.md)
- [完整 API 文档](./TENCENT_LIVE_API.md)
