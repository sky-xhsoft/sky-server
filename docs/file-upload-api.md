# 文件上传 API 文档

> **版本**: v1.0
> **最后更新**: 2026-01-28
> **维护者**: Sky Team

---

## 目录

1. [API 概述](#api-概述)
2. [认证说明](#认证说明)
3. [API 接口](#api-接口)
4. [数据模型](#数据模型)
5. [错误码](#错误码)
6. [使用示例](#使用示例)
7. [技术实现](#技术实现)

---

## API 概述

文件上传 API 提供完整的文件管理功能，包括：

- ✅ 单文件/批量上传
- ✅ 文件下载和预览
- ✅ 文件信息查询
- ✅ 文件删除
- ✅ MD5 秒传
- ✅ 公开访问（无需认证）

### 基础信息

- **Base URL**: `/api/v1/files`
- **认证方式**: Bearer Token（部分接口需要）
- **上传限制**: 默认 100MB
- **支持格式**: `.jpg`, `.jpeg`, `.png`, `.gif`, `.pdf`, `.doc`, `.docx`, `.xls`, `.xlsx`, `.txt`, `.zip`, `.rar`

---

## 认证说明

### 需要认证的接口

以下接口需要在请求头中携带 JWT Token：

```
Authorization: Bearer {token}
```

- `POST /files/upload` - 上传文件
- `POST /files/upload/multiple` - 批量上传
- `GET /files/download/:id` - 下载文件
- `GET /files/preview/:id` - 预览文件
- `GET /files/:id` - 获取文件信息
- `POST /files/list` - 查询文件列表
- `DELETE /files/:id` - 删除文件

### 公开访问接口

以下接口无需认证即可访问：

- `GET /files/access/:storageName` - 通过存储名称访问文件

---

## API 接口

### 1. 上传单个文件

**接口**: `POST /files/upload`

**认证**: 需要

**请求格式**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 上传的文件 |
| category | string | 否 | 文件分类，默认 "default" |

**响应示例**:

```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "id": 7,
    "sysCompanyId": 0,
    "createBy": "admin",
    "createTime": "2026-01-28T18:52:53.941+08:00",
    "updateBy": "admin",
    "updateTime": "2026-01-28T18:52:53.942+08:00",
    "isActive": "Y",
    "fileName": "example.png",
    "storageName": "a16dd4fc-4412-4367-a882-668f6583b04a.png",
    "filePath": "uploads/2026/01/28/a16dd4fc-4412-4367-a882-668f6583b04a.png",
    "fileSize": 2025971,
    "fileType": "image/png",
    "fileExt": ".png",
    "storageType": "local",
    "bucketName": "",
    "accessUrl": "/api/v1/files/access/a16dd4fc-4412-4367-a882-668f6583b04a.png",
    "thumbnailUrl": "",
    "md5": "859c6e876f68fdb254e645aa665f7116",
    "uploadIp": "127.0.0.1",
    "downloadCount": 0,
    "category": "default",
    "description": "",
    "expireTime": null
  }
}
```

**cURL 示例**:

```bash
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer {token}" \
  -F "file=@/path/to/file.png" \
  -F "category=images"
```

---

### 2. 批量上传文件

**接口**: `POST /files/upload/multiple`

**认证**: 需要

**请求格式**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| files | File[] | 是 | 上传的文件数组 |
| category | string | 否 | 文件分类，默认 "default" |

**响应示例**:

```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "total": 3,
    "success": 3,
    "files": [
      { /* 文件对象1 */ },
      { /* 文件对象2 */ },
      { /* 文件对象3 */ }
    ]
  }
}
```

---

### 3. 下载文件

**接口**: `GET /files/download/:id`

**认证**: 需要

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应**: 文件流（自动下载）

**响应头**:
```
Content-Type: {文件MIME类型}
Content-Disposition: attachment; filename={文件名}
```

---

### 4. 预览文件

**接口**: `GET /files/preview/:id`

**认证**: 需要

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应**: 文件流（浏览器内预览）

**响应头**:
```
Content-Type: {文件MIME类型}
Content-Disposition: inline; filename={文件名}
```

---

### 5. 通过存储名称访问文件（公开）

**接口**: `GET /files/access/:storageName`

**认证**: 不需要

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| storageName | string | 文件存储名称（UUID + 扩展名） |

**响应**: 文件流（浏览器内预览）

**响应头**:
```
Content-Type: {文件MIME类型}
Content-Disposition: inline; filename={文件名}
Cache-Control: public, max-age=31536000
```

**特点**:
- ✅ 无需认证，可直接访问
- ✅ 支持浏览器缓存（1年）
- ✅ 适合用于图片、视频等公开资源
- ✅ 从数据库查询文件信息，安全可控

**示例**:
```
http://localhost:8080/api/v1/files/access/a16dd4fc-4412-4367-a882-668f6583b04a.png
```

---

### 6. 获取文件信息

**接口**: `GET /files/:id`

**认证**: 需要

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 7,
    "fileName": "example.png",
    "fileSize": 2025971,
    "fileType": "image/png",
    "accessUrl": "/api/v1/files/access/a16dd4fc-4412-4367-a882-668f6583b04a.png",
    "createTime": "2026-01-28T18:52:53.941+08:00"
  }
}
```

---

### 7. 查询文件列表

**接口**: `POST /files/list`

**认证**: 需要

**请求体**:

```json
{
  "page": 1,
  "pageSize": 20,
  "category": "images",
  "fileName": "example",
  "fileType": "image/png",
  "createBy": "admin",
  "startTime": "2026-01-01T00:00:00Z",
  "endTime": "2026-01-31T23:59:59Z"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "pageSize": 20,
    "list": [
      { /* 文件对象1 */ },
      { /* 文件对象2 */ }
    ]
  }
}
```

---

### 8. 删除文件

**接口**: `DELETE /files/:id`

**认证**: 需要

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应示例**:

```json
{
  "code": 0,
  "message": "删除成功"
}
```

---

## 数据模型

### SysFile 文件对象

```go
type SysFile struct {
    ID            uint       `json:"id"`
    SysCompanyID  uint       `json:"sysCompanyId"`
    CreateBy      string     `json:"createBy"`
    CreateTime    time.Time  `json:"createTime"`
    UpdateBy      string     `json:"updateBy"`
    UpdateTime    time.Time  `json:"updateTime"`
    IsActive      string     `json:"isActive"`
    FileName      string     `json:"fileName"`      // 原始文件名
    StorageName   string     `json:"storageName"`   // 存储文件名（UUID）
    FilePath      string     `json:"filePath"`      // 文件路径
    FileSize      int64      `json:"fileSize"`      // 文件大小（字节）
    FileType      string     `json:"fileType"`      // MIME类型
    FileExt       string     `json:"fileExt"`       // 文件扩展名
    StorageType   string     `json:"storageType"`   // 存储类型：local, oss, s3
    BucketName    string     `json:"bucketName"`    // 存储桶名称
    AccessURL     string     `json:"accessUrl"`     // 访问URL
    ThumbnailURL  string     `json:"thumbnailUrl"`  // 缩略图URL
    MD5           string     `json:"md5"`           // 文件MD5值
    UploadIP      string     `json:"uploadIp"`      // 上传IP
    DownloadCount int        `json:"downloadCount"` // 下载次数
    Category      string     `json:"category"`      // 文件分类
    Description   string     `json:"description"`   // 文件描述
    ExpireTime    *time.Time `json:"expireTime"`    // 过期时间
}
```

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 20002 | 未提供认证令牌 |
| 20003 | 令牌无效或已过期 |
| 40001 | 参数错误 |
| 40004 | 资源不存在 |
| 50000 | 服务器内部错误 |

**错误响应示例**:

```json
{
  "code": 40001,
  "message": "文件大小超过限制（最大100MB）"
}
```

---

## 使用示例

### JavaScript/TypeScript

```typescript
// 上传文件
async function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('category', 'images')

  const response = await fetch('/api/v1/files/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  })

  const result = await response.json()
  if (result.code === 0) {
    console.log('上传成功:', result.data.accessUrl)
    return result.data
  } else {
    throw new Error(result.message)
  }
}

// 显示图片
function displayImage(accessUrl: string) {
  const img = document.createElement('img')
  img.src = accessUrl  // 无需认证即可访问
  document.body.appendChild(img)
}
```

### Vue 3 + Arco Design

```vue
<template>
  <a-upload
    :action="uploadUrl"
    :headers="uploadHeaders"
    :file-list="fileList"
    @change="handleChange"
  >
    <template #upload-button>
      <a-button>上传文件</a-button>
    </template>
  </a-upload>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { getAccessToken } from '@/utils/token'

const uploadUrl = '/api/v1/files/upload'
const fileList = ref([])

const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${getAccessToken()}`
}))

function handleChange(fileList: any[]) {
  const files = fileList.map(file => ({
    name: file.name,
    url: file.response?.data?.accessUrl || file.url
  }))
  console.log('上传的文件:', files)
}
</script>
```

---

## 技术实现

### 文件存储结构

```
uploads/
├── 2026/
│   ├── 01/
│   │   ├── 28/
│   │   │   ├── a16dd4fc-4412-4367-a882-668f6583b04a.png
│   │   │   ├── b27ee5fd-5523-5478-b993-779g7694c15b.jpg
│   │   │   └── ...
│   │   └── 29/
│   └── 02/
└── ...
```

### 文件命名规则

- **存储名称**: `{UUID}.{扩展名}`
- **目录结构**: `uploads/{年}/{月}/{日}/`
- **访问URL**: `/api/v1/files/access/{storageName}`

### MD5 秒传机制

1. 上传文件时计算 MD5
2. 查询数据库是否存在相同 MD5 的文件
3. 如果存在，创建新的文件记录但共享物理文件
4. 如果不存在，保存新文件

### 安全机制

1. **文件类型验证**: 检查文件扩展名是否在允许列表中
2. **文件大小限制**: 默认 100MB，可配置
3. **存储名称隔离**: 使用 UUID 避免文件名冲突
4. **数据库验证**: 公开访问接口从数据库查询文件信息，防止路径遍历攻击
5. **多租户隔离**: 通过 SYS_COMPANY_ID 实现公司级别的数据隔离

### 性能优化

1. **浏览器缓存**: 公开访问接口设置 1 年缓存
2. **数据库索引**: 在 STORAGE_NAME 和 MD5 字段上建立索引
3. **秒传机制**: 相同文件只存储一次物理文件

---

## 配置说明

### 服务配置

在 `cmd/server/main.go` 中配置文件服务：

```go
fileService := file.NewService(db, &file.Config{
    UploadDir:   "./uploads",           // 上传目录
    MaxFileSize: 100 * 1024 * 1024,     // 最大文件大小（100MB）
    AllowedExts: []string{              // 允许的文件扩展名
        ".jpg", ".jpeg", ".png", ".gif",
        ".pdf", ".doc", ".docx",
        ".xls", ".xlsx", ".txt",
        ".zip", ".rar",
    },
})
```

### 路由配置

在 `api/router/router.go` 中：

```go
files := rg.Group("/files")
{
    // 公开访问（无需认证）
    files.GET("/access/:storageName", fileHandler.GetFileByPath)

    // 需要认证的接口
    authenticated := files.Group("")
    authenticated.Use(middleware.AuthRequired(jwtUtil))
    {
        authenticated.POST("/upload", fileHandler.UploadFile)
        authenticated.POST("/upload/multiple", fileHandler.UploadMultipleFiles)
        authenticated.GET("/download/:id", fileHandler.DownloadFile)
        authenticated.GET("/preview/:id", fileHandler.PreviewFile)
        authenticated.GET("/:id", fileHandler.GetFile)
        authenticated.POST("/list", fileHandler.ListFiles)
        authenticated.DELETE("/:id", fileHandler.DeleteFile)
    }
}
```

---

## 相关文档

- [前端图片上传功能](../../sky-web/docs/features/image-upload-feature.md)
- [API 设计文档](./03-API设计.md)
- [数据库设计文档](./02-数据库设计.md)

---

## 更新日志

### 2026-01-28 v1.0

- ✅ 实现文件上传 API
- ✅ 支持单文件/批量上传
- ✅ 实现 MD5 秒传
- ✅ 添加公开访问接口
- ✅ 优化文件存储结构
- ✅ 添加安全验证机制
- ✅ 实现浏览器缓存优化

---

**文档版本**: v1.0
**最后更新**: 2026-01-28
**维护者**: Sky Team
