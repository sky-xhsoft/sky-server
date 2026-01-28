# 云盘服务完整指南

> **版本**: v2.0
> **最后更新**: 2026-01-28
> **维护者**: Sky Team

本文档整合了云盘服务的 API 文档、快速开始指南和服务实现说明。

---

## 📚 目录

1. [快速开始](#快速开始)
2. [API 接口](#api-接口)
3. [服务实现](#服务实现)
4. [最佳实践](#最佳实践)

---

## 🚀 快速开始

### 基础信息

- **Base URL**: `/api/v1/cloud`
- **认证方式**: JWT Bearer Token
- **请求头**: `Authorization: Bearer {token}`

### 快速示例

#### 1. 创建文件夹

```bash
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "parentId": 0,
    "folderName": "我的文档"
  }'
```

#### 2. 上传文件

```bash
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/file.pdf" \
  -F "folderId=1"
```

#### 3. 列出文件

```bash
curl -X GET "http://localhost:9090/api/v1/cloud/files?folderId=1&page=1&pageSize=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 4. 下载文件

```bash
curl -X GET http://localhost:9090/api/v1/cloud/files/1/download \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o downloaded_file.pdf
```

### 常见使用场景

#### 场景 1: 创建文件夹并上传文件

```bash
# 1. 创建文件夹
FOLDER_RESPONSE=$(curl -s -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"parentId": 0, "folderName": "项目文档"}')

FOLDER_ID=$(echo $FOLDER_RESPONSE | jq -r '.data.ID')

# 2. 上传文件到该文件夹
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@document.pdf" \
  -F "folderId=$FOLDER_ID"
```

#### 场景 2: 创建分享链接

```bash
# 1. 创建分享
curl -X POST http://localhost:9090/api/v1/cloud/shares \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "itemId": 1,
    "itemType": "file",
    "expireDays": 7,
    "password": "abc123"
  }'
```

---

## 📖 API 接口

### 文件夹管理

#### 1. 创建文件夹

**请求**

```http
POST /api/v1/cloud/folders
Content-Type: application/json
Authorization: Bearer {token}

{
  "parentId": 0,              // 父文件夹ID，0表示根目录
  "folderName": "我的文件夹"   // 文件夹名称
}
```

**响应**

```json
{
  "code": 201,
  "message": "success",
  "data": {
    "ID": 1,
    "FolderName": "我的文件夹",
    "ParentID": 0,
    "UserID": 100,
    "CreateTime": "2026-01-13T10:00:00Z"
  }
}
```

#### 2. 列出文件夹

**请求**

```http
GET /api/v1/cloud/folders?parentId=0
Authorization: Bearer {token}
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| parentId | int | 否 | 父文件夹ID，默认为0（根目录） |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "ID": 1,
      "FolderName": "文档",
      "ParentID": 0,
      "UserID": 100,
      "CreateTime": "2026-01-13T10:00:00Z"
    }
  ]
}
```

#### 3. 重命名文件夹

**请求**

```http
PUT /api/v1/cloud/folders/{id}
Content-Type: application/json
Authorization: Bearer {token}

{
  "folderName": "新文件夹名称"
}
```

#### 4. 删除文件夹

**请求**

```http
DELETE /api/v1/cloud/folders/{id}
Authorization: Bearer {token}
```

**说明**: 删除文件夹会同时删除其中的所有文件和子文件夹。

---

### 文件管理

#### 1. 上传文件

**请求**

```http
POST /api/v1/cloud/files
Content-Type: multipart/form-data
Authorization: Bearer {token}

file: (binary)
folderId: 1
```

**表单参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 要上传的文件 |
| folderId | int | 否 | 目标文件夹ID，默认为0（根目录） |

**响应**

```json
{
  "code": 201,
  "message": "success",
  "data": {
    "ID": 1,
    "FileName": "document.pdf",
    "FileSize": 1024000,
    "FileType": "application/pdf",
    "FolderID": 1,
    "StoragePath": "/uploads/2026/01/28/xxx.pdf",
    "UploadTime": "2026-01-28T10:00:00Z"
  }
}
```

#### 2. 列出文件

**请求**

```http
GET /api/v1/cloud/files?folderId=1&page=1&pageSize=20
Authorization: Bearer {token}
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| folderId | int | 否 | 文件夹ID，默认为0（根目录） |
| page | int | 否 | 页码，默认为1 |
| pageSize | int | 否 | 每页数量，默认为20 |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "pageSize": 20,
    "list": [
      {
        "ID": 1,
        "FileName": "document.pdf",
        "FileSize": 1024000,
        "FileType": "application/pdf",
        "FolderID": 1,
        "UploadTime": "2026-01-28T10:00:00Z"
      }
    ]
  }
}
```

#### 3. 下载文件

**请求**

```http
GET /api/v1/cloud/files/{id}/download
Authorization: Bearer {token}
```

**响应**: 文件流（二进制数据）

**响应头**

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="document.pdf"
Content-Length: 1024000
```

#### 4. 删除文件

**请求**

```http
DELETE /api/v1/cloud/files/{id}
Authorization: Bearer {token}
```

#### 5. 批量删除文件

**请求**

```http
POST /api/v1/cloud/files/batch-delete
Content-Type: application/json
Authorization: Bearer {token}

{
  "fileIds": [1, 2, 3]
}
```

---

### 分享管理

#### 1. 创建分享

**请求**

```http
POST /api/v1/cloud/shares
Content-Type: application/json
Authorization: Bearer {token}

{
  "itemId": 1,
  "itemType": "file",        // "file" 或 "folder"
  "expireDays": 7,           // 过期天数，0表示永久
  "password": "abc123"       // 可选，访问密码
}
```

**响应**

```json
{
  "code": 201,
  "message": "success",
  "data": {
    "ID": 1,
    "ShareCode": "abc123xyz",
    "ItemID": 1,
    "ItemType": "file",
    "ExpireTime": "2026-02-04T10:00:00Z",
    "ShareURL": "http://example.com/share/abc123xyz"
  }
}
```

#### 2. 访问分享

**请求**

```http
GET /api/v1/cloud/shares/{shareCode}
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 否 | 访问密码（如果设置了密码） |

#### 3. 取消分享

**请求**

```http
DELETE /api/v1/cloud/shares/{id}
Authorization: Bearer {token}
```

---

### 配额管理

#### 1. 查询配额

**请求**

```http
GET /api/v1/cloud/quota
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "totalQuota": 10737418240,      // 总配额（字节）
    "usedQuota": 1073741824,        // 已使用（字节）
    "availableQuota": 9663676416,   // 可用配额（字节）
    "fileCount": 150,               // 文件数量
    "maxFileSize": 1073741824       // 单文件最大大小（字节）
  }
}
```

---

## 🔧 服务实现

### 架构设计

云盘服务采用分层架构：

```
┌─────────────────────────────────────┐
│         API Handler Layer           │  ← HTTP 请求处理
├─────────────────────────────────────┤
│         Service Layer               │  ← 业务逻辑
├─────────────────────────────────────┤
│         Repository Layer            │  ← 数据访问
├─────────────────────────────────────┤
│         Database (MySQL)            │  ← 数据存储
└─────────────────────────────────────┘
```

### 核心组件

#### 1. CloudService

**位置**: `internal/service/cloud/cloud_service.go`

**职责**:
- 文件夹和文件的 CRUD 操作
- 文件上传和下载
- 配额管理
- 分享管理

**关键方法**:

```go
type Service interface {
    // 文件夹管理
    CreateFolder(ctx context.Context, req *CreateFolderRequest) (*CloudFolder, error)
    ListFolders(ctx context.Context, parentID uint) ([]*CloudFolder, error)
    DeleteFolder(ctx context.Context, folderID uint) error

    // 文件管理
    UploadFile(ctx context.Context, req *UploadFileRequest) (*CloudFile, error)
    ListFiles(ctx context.Context, req *ListFilesRequest) ([]*CloudFile, int64, error)
    DownloadFile(ctx context.Context, fileID uint) (*CloudFile, error)
    DeleteFile(ctx context.Context, fileID uint) error

    // 配额管理
    GetQuota(ctx context.Context, userID uint) (*Quota, error)
    CheckQuota(ctx context.Context, userID uint, fileSize int64) error

    // 分享管理
    CreateShare(ctx context.Context, req *CreateShareRequest) (*Share, error)
    GetShare(ctx context.Context, shareCode string) (*Share, error)
}
```

#### 2. CloudHandler

**位置**: `api/handler/cloud_handler.go`

**职责**:
- HTTP 请求解析
- 参数验证
- 响应格式化
- 错误处理

#### 3. 数据模型

**CloudFolder** - 文件夹

```go
type CloudFolder struct {
    ID         uint      `gorm:"primaryKey"`
    FolderName string    `gorm:"size:255;not null"`
    ParentID   uint      `gorm:"default:0"`
    UserID     uint      `gorm:"not null;index"`
    CreateTime time.Time `gorm:"autoCreateTime"`
    UpdateTime time.Time `gorm:"autoUpdateTime"`
}
```

**CloudFile** - 文件

```go
type CloudFile struct {
    ID          uint      `gorm:"primaryKey"`
    FileName    string    `gorm:"size:255;not null"`
    FileSize    int64     `gorm:"not null"`
    FileType    string    `gorm:"size:100"`
    FolderID    uint      `gorm:"default:0;index"`
    UserID      uint      `gorm:"not null;index"`
    StoragePath string    `gorm:"size:500;not null"`
    UploadTime  time.Time `gorm:"autoCreateTime"`
}
```

### 文件存储

#### 存储策略

- **本地存储**: 文件存储在服务器本地文件系统
- **路径结构**: `/uploads/{year}/{month}/{day}/{uuid}.{ext}`
- **文件命名**: 使用 UUID 避免冲突

#### 存储配置

```go
type StorageConfig struct {
    UploadDir   string // 上传目录
    MaxFileSize int64  // 最大文件大小
    AllowedExts []string // 允许的文件扩展名
}
```

### 配额管理

#### 配额计算

```go
func (s *service) GetQuota(ctx context.Context, userID uint) (*Quota, error) {
    // 1. 获取用户总配额（从配置或数据库）
    totalQuota := s.config.DefaultQuota

    // 2. 计算已使用配额
    var usedQuota int64
    s.db.Model(&CloudFile{}).
        Where("user_id = ?", userID).
        Select("COALESCE(SUM(file_size), 0)").
        Scan(&usedQuota)

    // 3. 计算可用配额
    availableQuota := totalQuota - usedQuota

    return &Quota{
        TotalQuota:     totalQuota,
        UsedQuota:      usedQuota,
        AvailableQuota: availableQuota,
    }, nil
}
```

#### 配额检查

上传文件前检查配额：

```go
func (s *service) CheckQuota(ctx context.Context, userID uint, fileSize int64) error {
    quota, err := s.GetQuota(ctx, userID)
    if err != nil {
        return err
    }

    if quota.AvailableQuota < fileSize {
        return errors.New("配额不足")
    }

    return nil
}
```

### 安全机制

#### 1. 权限验证

- 所有 API 需要 JWT 认证
- 用户只能访问自己的文件和文件夹
- 分享链接可以公开访问（需要密码验证）

#### 2. 文件类型验证

```go
var AllowedFileTypes = []string{
    ".jpg", ".jpeg", ".png", ".gif",
    ".pdf", ".doc", ".docx",
    ".xls", ".xlsx",
    ".txt", ".md",
    ".zip", ".rar",
}

func ValidateFileType(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    for _, allowed := range AllowedFileTypes {
        if ext == allowed {
            return true
        }
    }
    return false
}
```

#### 3. 文件大小限制

```go
const MaxFileSize = 100 * 1024 * 1024 // 100MB

func ValidateFileSize(size int64) error {
    if size > MaxFileSize {
        return errors.New("文件大小超过限制")
    }
    return nil
}
```

---

## 💡 最佳实践

### 1. 错误处理

始终检查 API 响应的 `code` 字段：

```javascript
const response = await fetch('/api/v1/cloud/files', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
})

const result = await response.json()

if (result.code !== 201) {
  console.error('上传失败:', result.message)
  return
}

console.log('上传成功:', result.data)
```

### 2. 大文件上传

对于大文件（>10MB），建议使用分片上传。参见 [云盘高级上传功能文档](./cloud-upload-advanced.md)。

### 3. 批量操作

使用批量 API 而不是循环调用单个 API：

```javascript
// ✅ 推荐：批量删除
await fetch('/api/v1/cloud/files/batch-delete', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    fileIds: [1, 2, 3, 4, 5]
  })
})

// ❌ 不推荐：循环删除
for (const id of [1, 2, 3, 4, 5]) {
  await fetch(`/api/v1/cloud/files/${id}`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` }
  })
}
```

### 4. 配额检查

上传前检查配额：

```javascript
// 1. 获取配额信息
const quotaResponse = await fetch('/api/v1/cloud/quota', {
  headers: { 'Authorization': `Bearer ${token}` }
})
const quota = await quotaResponse.json()

// 2. 检查是否有足够空间
if (quota.data.availableQuota < file.size) {
  alert('存储空间不足')
  return
}

// 3. 上传文件
const formData = new FormData()
formData.append('file', file)
await fetch('/api/v1/cloud/files', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: formData
})
```

### 5. 分享链接安全

- 设置合理的过期时间
- 对敏感文件使用密码保护
- 定期清理过期分享

```javascript
// 创建7天有效、带密码的分享
await fetch('/api/v1/cloud/shares', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    itemId: fileId,
    itemType: 'file',
    expireDays: 7,
    password: 'secure123'
  })
})
```

---

## 📚 相关文档

- [云盘高级上传功能](./cloud-upload-advanced.md) - 大文件上传、分片上传、断点续传
- [云盘故障排查](./cloud-troubleshooting.md) - 常见问题和解决方案
- [文件上传 API](./file-upload-api.md) - 通用文件上传接口

---

**文档版本**: v2.0
**最后更新**: 2026-01-28
**维护者**: Sky Team

> 本文档整合自: cloud-api.md, cloud-api-quickstart.md, cloud-service-implementation.md
