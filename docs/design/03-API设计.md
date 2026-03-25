# Sky-Server API设计文档

## 1. API设计原则

### 1.1 RESTful设计

- 使用标准HTTP方法：GET(查询)、POST(创建)、PUT(更新)、DELETE(删除)
- URL使用名词复数形式
- 使用HTTP状态码表示结果
- 支持版本控制（通过URL path）

### 1.2 统一响应格式

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 业务数据
  },
  "timestamp": "2026-01-11T10:30:00Z"
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "参数错误",
  "error": {
    "field": "username",
    "detail": "用户名不能为空"
  },
  "timestamp": "2026-01-11T10:30:00Z"
}
```

**分页响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [...],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
    }
  },
  "timestamp": "2026-01-11T10:30:00Z"
}
```

### 1.3 HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无返回内容） |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如唯一性约束） |
| 500 | 服务器内部错误 |

## 2. 认证与授权

### 2.1 登录认证（单点登录）

**接口**: `POST /api/v1/auth/login`

**请求**:
```json
{
  "username": "admin",
  "password": "123456",
  "companyId": 1,
  "clientType": "web",
  "deviceId": "uuid-xxxx-xxxx",
  "deviceName": "Chrome on Windows"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| companyId | int | 是 | 公司ID |
| clientType | string | 是 | 客户端类型（web/mobile/desktop） |
| deviceId | string | 否 | 设备唯一标识，如不传则自动生成 |
| deviceName | string | 否 | 设备名称，用于显示 |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 7200,
    "user": {
      "id": 1,
      "username": "admin",
      "trueName": "管理员",
      "isAdmin": "Y",
      "companyId": 1
    }
  }
}
```

### 2.2 Token刷新

**接口**: `POST /api/v1/auth/refresh`

**请求头**:
```
Authorization: Bearer {refreshToken}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 7200
  }
}
```

### 2.3 登出

**接口**: `POST /api/v1/auth/logout`

**请求头**:
```
Authorization: Bearer {token}
```

## 3. 通用CRUD API

### 3.1 查询列表

**接口**: `GET /api/v1/data/{tableName}`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| pageSize | int | 否 | 每页数量，默认20 |
| sortBy | string | 否 | 排序字段 |
| sortOrder | string | 否 | 排序方向(asc/desc) |
| filter | string | 否 | 过滤条件（JSON字符串） |

**filter示例**:
```json
{
  "NAME": {"op": "like", "value": "张%"},
  "IS_ACTIVE": {"op": "=", "value": "Y"},
  "CREATE_TIME": {"op": ">=", "value": "2026-01-01"}
}
```

**支持的操作符**:
- `=`: 等于
- `!=`: 不等于
- `>`: 大于
- `>=`: 大于等于
- `<`: 小于
- `<=`: 小于等于
- `like`: 模糊匹配
- `in`: 在列表中
- `between`: 范围查询

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "ID": 1,
        "NAME": "张三",
        "IS_ACTIVE": "Y",
        // ...其他字段
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
    }
  }
}
```

### 3.2 查询单条记录

**接口**: `GET /api/v1/data/{tableName}/{id}`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "ID": 1,
    "NAME": "张三",
    "IS_ACTIVE": "Y",
    // ...其他字段
  }
}
```

### 3.3 创建记录

**接口**: `POST /api/v1/data/{tableName}`

**请求**:
```json
{
  "NAME": "李四",
  "EMAIL": "lisi@example.com",
  // ...其他字段（根据元数据）
}
```

**响应**:
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 2
  }
}
```

### 3.4 更新记录

**接口**: `PUT /api/v1/data/{tableName}/{id}`

**请求**:
```json
{
  "NAME": "李四2",
  "EMAIL": "lisi2@example.com"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "更新成功"
}
```

### 3.5 删除记录（软删除）

**接口**: `DELETE /api/v1/data/{tableName}/{id}`

**响应**:
```json
{
  "code": 204,
  "message": "删除成功"
}
```

### 3.6 批量操作

**批量创建**: `POST /api/v1/data/{tableName}/batch`

**请求**:
```json
{
  "items": [
    {"NAME": "用户1", "EMAIL": "user1@example.com"},
    {"NAME": "用户2", "EMAIL": "user2@example.com"}
  ]
}
```

**批量删除**: `DELETE /api/v1/data/{tableName}/batch`

**请求**:
```json
{
  "ids": [1, 2, 3]
}
```

## 4. 元数据API

### 4.1 获取表元数据

**接口**: `GET /api/v1/metadata/tables/{tableName}`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "sys_user",
    "displayName": "系统用户",
    "mask": "AMDQ",
    "columns": [
      {
        "id": 1,
        "name": "用户名",
        "dbName": "USERNAME",
        "colType": "varchar",
        "colLength": 255,
        "nullable": false,
        "displayType": "text",
        "setValueType": "byPage"
      }
    ],
    "refs": [
      {
        "id": 1,
        "refTableId": 2,
        "refTableName": "sys_company",
        "assocType": "1"
      }
    ]
  }
}
```

### 4.2 获取所有表列表

**接口**: `GET /api/v1/metadata/tables`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| isMenu | string | 否 | 是否菜单(Y/N) |
| categoryId | int | 否 | 表类别ID |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "sys_user",
      "displayName": "系统用户",
      "isMenu": "Y"
    }
  ]
}
```

### 4.3 刷新元数据缓存

**接口**: `POST /api/v1/metadata/refresh`

**响应**:
```json
{
  "code": 200,
  "message": "缓存刷新成功"
}
```

## 5. 数据字典API

### 5.1 获取字典列表

**接口**: `GET /api/v1/dict`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "gender",
      "displayName": "性别",
      "type": 0
    }
  ]
}
```

### 5.2 获取字典项

**接口**: `GET /api/v1/dict/{dictName}/items`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "displayName": "男",
      "value": "M",
      "orderno": 1,
      "isDefaultValue": "Y"
    },
    {
      "id": 2,
      "displayName": "女",
      "value": "F",
      "orderno": 2,
      "isDefaultValue": "N"
    }
  ]
}
```

## 6. 动作执行API

### 6.1 执行动作

**接口**: `POST /api/v1/action/{actionId}/execute`

**请求**:
```json
{
  "context": {
    "recordId": 1,
    "tableName": "sys_user",
    "params": {
      "key1": "value1"
    }
  }
}
```

**响应**:
```json
{
  "code": 200,
  "message": "执行成功",
  "data": {
    "result": "执行结果"
  }
}
```

### 6.2 获取表的动作列表

**接口**: `GET /api/v1/action/table/{tableId}`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| displayType | string | 否 | 显示类型筛选 |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "export",
      "displayName": "导出",
      "displayType": "list_button",
      "actionType": "url"
    }
  ]
}
```

## 7. 权限API

### 7.1 检查权限

**接口**: `GET /api/v1/permission/check`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tableId | int | 是 | 表ID |
| action | string | 是 | 操作(A/M/D/Q/S/U/V) |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "hasPermission": true
  }
}
```

### 7.2 获取用户权限列表

**接口**: `GET /api/v1/permission/user`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "permissions": [
      {
        "directoryId": 1,
        "directoryName": "系统管理",
        "tableId": 1,
        "tableName": "sys_user",
        "permission": 3,
        "actions": ["read", "write"]
      }
    ]
  }
}
```

## 8. 单据编号API

### 8.1 生成编号

**接口**: `POST /api/v1/sequence/{seqName}/next`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "value": "PO202601110001"
  }
}
```

### 8.2 预览编号

**接口**: `GET /api/v1/sequence/{seqName}/preview`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "nextValue": "PO202601110002",
    "format": "PO{YYYY}{MM}{DD}{0000}"
  }
}
```

## 9. 用户管理API

### 9.1 获取当前用户信息

**接口**: `GET /api/v1/user/current`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "trueName": "管理员",
    "email": "admin@example.com",
    "isAdmin": "Y",
    "companyId": 1
  }
}
```

### 9.2 修改密码

**接口**: `PUT /api/v1/user/password`

**请求**:
```json
{
  "oldPassword": "123456",
  "newPassword": "654321"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "密码修改成功"
}
```

## 10. 文件上传API

### 10.1 上传文件

**接口**: `POST /api/v1/file/upload`

**请求**: multipart/form-data

**参数**:
- file: 文件
- tableName: 表名
- columnName: 字段名
- recordId: 记录ID（可选）

**响应**:
```json
{
  "code": 200,
  "message": "上传成功",
  "data": {
    "fileId": "uuid",
    "fileName": "document.pdf",
    "fileSize": 102400,
    "url": "/api/v1/file/download/uuid"
  }
}
```

### 10.2 下载文件

**接口**: `GET /api/v1/file/download/{fileId}`

返回文件流。

## 11. 导入导出API

### 11.1 导出数据

**接口**: `POST /api/v1/export/{tableName}`

**请求**:
```json
{
  "filter": {...},
  "columns": ["ID", "NAME", "EMAIL"],
  "format": "xlsx"
}
```

**响应**: 返回文件流

### 11.2 导入数据

**接口**: `POST /api/v1/import/{tableName}`

**请求**: multipart/form-data
- file: Excel文件

**响应**:
```json
{
  "code": 200,
  "message": "导入成功",
  "data": {
    "total": 100,
    "success": 98,
    "failed": 2,
    "errors": [
      {"row": 5, "error": "用户名重复"},
      {"row": 10, "error": "邮箱格式错误"}
    ]
  }
}
```

## 12. 单点登录管理API

### 12.1 获取用户所有活跃会话

**接口**: `GET /api/v1/sso/sessions`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "deviceId": "uuid-xxxx-1",
      "deviceName": "Chrome on Windows",
      "clientType": "web",
      "ipAddress": "192.168.1.100",
      "loginTime": "2026-01-11T10:00:00Z",
      "lastActiveTime": "2026-01-11T15:30:00Z",
      "isCurrent": true
    },
    {
      "deviceId": "uuid-xxxx-2",
      "deviceName": "iPhone 15",
      "clientType": "mobile",
      "ipAddress": "192.168.1.101",
      "loginTime": "2026-01-10T08:00:00Z",
      "lastActiveTime": "2026-01-11T14:00:00Z",
      "isCurrent": false
    }
  ]
}
```

### 12.2 踢出指定设备

**接口**: `DELETE /api/v1/sso/sessions/{deviceId}`

**响应**:
```json
{
  "code": 200,
  "message": "设备已踢出"
}
```

### 12.3 登出所有设备

**接口**: `POST /api/v1/sso/logout-all`

**响应**:
```json
{
  "code": 200,
  "message": "已登出所有设备"
}
```

## 13. 单据生命周期API

### 13.1 提交单据

**接口**: `POST /api/v1/data/{tableName}/{id}/submit`

**响应**:
```json
{
  "code": 200,
  "message": "提交成功",
  "data": {
    "status": "SUBMITTED",
    "submitTime": "2026-01-11T15:30:00Z"
  }
}
```

### 13.2 反提交

**接口**: `POST /api/v1/data/{tableName}/{id}/unsubmit`

**响应**:
```json
{
  "code": 200,
  "message": "反提交成功",
  "data": {
    "status": "DRAFT"
  }
}
```

### 13.3 作废单据

**接口**: `POST /api/v1/data/{tableName}/{id}/void`

**请求**:
```json
{
  "reason": "作废原因"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "作废成功",
  "data": {
    "status": "VOIDED"
  }
}
```

### 13.4 获取单据状态

**接口**: `GET /api/v1/data/{tableName}/{id}/status`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "SUBMITTED",
    "statusName": "已提交",
    "submitBy": "张三",
    "submitTime": "2026-01-11T15:30:00Z",
    "canEdit": false,
    "canSubmit": false,
    "canApprove": true,
    "canVoid": true
  }
}
```

## 14. 审核流程API

### 14.1 审核通过

**接口**: `POST /api/v1/workflow/approve`

**请求**:
```json
{
  "tableName": "purchase_order",
  "recordId": 123,
  "comment": "同意采购"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "审核通过",
  "data": {
    "status": "APPROVED",
    "approveTime": "2026-01-11T16:00:00Z"
  }
}
```

### 14.2 审核拒绝

**接口**: `POST /api/v1/workflow/reject`

**请求**:
```json
{
  "tableName": "purchase_order",
  "recordId": 123,
  "comment": "采购金额过高，不予批准"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "审核拒绝",
  "data": {
    "status": "DRAFT"
  }
}
```

### 14.3 获取审核历史

**接口**: `GET /api/v1/workflow/history`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tableName | string | 是 | 表名 |
| recordId | int | 是 | 记录ID |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "step": 1,
      "stepName": "部门主管审核",
      "approverName": "李四",
      "action": "APPROVE",
      "comment": "同意",
      "createTime": "2026-01-11T15:45:00Z"
    },
    {
      "step": 2,
      "stepName": "财务审核",
      "approverName": "王五",
      "action": "APPROVE",
      "comment": "财务审核通过",
      "createTime": "2026-01-11T16:00:00Z"
    }
  ]
}
```

### 14.4 获取待审核列表

**接口**: `GET /api/v1/workflow/pending`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| pageSize | int | 否 | 每页数量，默认20 |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "tableName": "purchase_order",
        "tableDisplayName": "采购订单",
        "recordId": 123,
        "recordTitle": "办公用品采购",
        "submitBy": "张三",
        "submitTime": "2026-01-11T15:30:00Z",
        "currentStep": 2,
        "currentStepName": "财务审核"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 5,
      "totalPages": 1
    }
  }
}
```

## 15. 审计日志API

### 15.1 查询审计日志

**接口**: `GET /api/v1/audit/logs`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tableName | string | 否 | 表名 |
| recordId | int | 否 | 记录ID |
| userId | int | 否 | 操作用户ID |
| action | string | 否 | 操作类型 |
| startTime | string | 否 | 开始时间 |
| endTime | string | 否 | 结束时间 |
| page | int | 否 | 页码 |
| pageSize | int | 否 | 每页数量 |

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1001,
        "tableName": "sys_user",
        "recordId": 10,
        "action": "UPDATE",
        "username": "admin",
        "changes": {
          "EMAIL": {
            "oldValue": "old@example.com",
            "newValue": "new@example.com"
          }
        },
        "ipAddress": "192.168.1.100",
        "createTime": "2026-01-11T15:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
    }
  }
}
```

### 15.2 获取记录变更历史

**接口**: `GET /api/v1/audit/history/{tableName}/{recordId}`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "action": "CREATE",
      "username": "张三",
      "createTime": "2026-01-10T10:00:00Z",
      "newValue": {
        "NAME": "测试用户",
        "EMAIL": "test@example.com"
      }
    },
    {
      "action": "UPDATE",
      "username": "李四",
      "createTime": "2026-01-11T15:30:00Z",
      "changes": {
        "EMAIL": {
          "oldValue": "test@example.com",
          "newValue": "test2@example.com"
        }
      }
    },
    {
      "action": "SUBMIT",
      "username": "张三",
      "createTime": "2026-01-11T16:00:00Z"
    }
  ]
}
```

### 15.3 导出审计日志

**接口**: `POST /api/v1/audit/export`

**请求**:
```json
{
  "tableName": "sys_user",
  "startTime": "2026-01-01T00:00:00Z",
  "endTime": "2026-01-31T23:59:59Z",
  "format": "xlsx"
}
```

**响应**: 返回Excel文件流

## 16. 系统配置API

### 12.1 获取系统参数

**接口**: `GET /api/v1/param/{paramName}`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "name": "system.title",
    "value": "Sky Server",
    "valueType": "str"
  }
}
```

### 12.2 更新系统参数

**接口**: `PUT /api/v1/param/{paramName}`

**请求**:
```json
{
  "value": "New Title"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "更新成功"
}
```

## 13. 健康检查API

### 13.1 健康检查

**接口**: `GET /api/v1/health`

**响应**:
```json
{
  "status": "UP",
  "database": "UP",
  "redis": "UP",
  "timestamp": "2026-01-11T10:30:00Z"
}
```

### 13.2 版本信息

**接口**: `GET /api/v1/version`

**响应**:
```json
{
  "version": "1.0.0",
  "buildTime": "2026-01-11T10:00:00Z",
  "gitCommit": "abc123"
}
```

## 14. WebSocket API

### 14.1 实时通知

**接口**: `WS /api/v1/ws/notification`

**连接参数**:
- token: JWT token

**消息格式**:
```json
{
  "type": "notification",
  "data": {
    "title": "新消息",
    "content": "您有一条新的待办事项",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

## 15. 错误码定义

| 错误码 | 说明 |
|--------|------|
| 10001 | 参数错误 |
| 10002 | 参数缺失 |
| 10003 | 参数类型错误 |
| 20001 | 用户名或密码错误 |
| 20002 | Token无效 |
| 20003 | Token过期 |
| 20004 | 无权限 |
| 30001 | 资源不存在 |
| 30002 | 资源已存在 |
| 30003 | 资源冲突 |
| 40001 | 数据库错误 |
| 40002 | 缓存错误 |
| 50001 | 服务器内部错误 |
| 50002 | 第三方服务错误 |

## 16. API安全

### 16.1 请求签名（可选）

对于敏感API，可以要求请求签名：

**请求头**:
```
X-Timestamp: 1641900000
X-Nonce: random_string
X-Signature: md5(timestamp + nonce + secret)
```

### 16.2 限流

- 同一IP每分钟最多100次请求
- 同一用户每分钟最多200次请求
- 超过限制返回429 Too Many Requests

### 16.3 CORS设置

```
Access-Control-Allow-Origin: https://example.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 3600
```

## 17. API版本管理

### 17.1 版本策略

- 使用URL path进行版本控制：`/api/v1/`, `/api/v2/`
- 主版本号变更表示不兼容的API变更
- 保留至少2个版本的兼容性
- 新版本发布前至少提前1个月通知

### 17.2 废弃通知

对于即将废弃的API，在响应头中添加：
```
X-API-Deprecated: true
X-API-Deprecation-Date: 2026-06-01
X-API-Sunset-Date: 2026-12-01
```

## 18. API文档

使用 Swagger/OpenAPI 3.0 规范生成API文档。

**访问地址**: `http://localhost:9090/swagger/index.html`

**Swagger注解示例**:
```go
// @Summary 获取用户列表
// @Description 分页查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} Response
// @Router /api/v1/data/sys_user [get]
// @Security BearerAuth
func GetUsers(c *gin.Context) {
    // ...
}
```
