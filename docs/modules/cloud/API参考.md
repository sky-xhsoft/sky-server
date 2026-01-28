# 云盘 API 参考

## 概述

云盘系统提供完整的文件管理 API，包括文件夹管理、文件上传下载、分享管理和配额管理。

**基础URL**: `http://localhost:9090/api/v1/cloud`

**认证方式**: Bearer Token

```http
Authorization: Bearer YOUR_TOKEN
```

---

## 文件夹管理

### 1. 创建文件夹

```http
POST /api/v1/cloud/folders
Content-Type: application/json

{
  "parentId": 0,
  "folderName": "我的文档"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "folderName": "我的文档",
    "parentId": 0,
    "path": "/我的文档",
    "createTime": "2026-01-28T10:00:00Z"
  }
}
```

### 2. 获取文件夹列表

```http
GET /api/v1/cloud/folders?parentId=0
```

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "folderName": "我的文档",
      "parentId": 0,
      "fileCount": 5,
      "subfolderCount": 2,
      "createTime": "2026-01-28T10:00:00Z"
    }
  ]
}
```

### 3. 重命名文件夹

```http
PUT /api/v1/cloud/folders/1
Content-Type: application/json

{
  "folderName": "工作文档"
}
```

### 4. 删除文件夹

```http
DELETE /api/v1/cloud/folders/1
```

---

## 文件管理

### 1. 上传文件

```http
POST /api/v1/cloud/files
Content-Type: multipart/form-data

file: <binary>
folderId: 1
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 100,
    "fileName": "document.pdf",
    "fileSize": 1048576,
    "fileType": "application/pdf",
    "md5": "abc123...",
    "accessUrl": "http://localhost:9090/files/cloud/...",
    "createTime": "2026-01-28T10:00:00Z"
  }
}
```

### 2. 获取文件列表

```http
GET /api/v1/cloud/files?folderId=1&page=1&pageSize=20
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "total": 50,
    "page": 1,
    "pageSize": 20,
    "items": [
      {
        "id": 100,
        "fileName": "document.pdf",
        "fileSize": 1048576,
        "fileType": "application/pdf",
        "downloadCount": 5,
        "createTime": "2026-01-28T10:00:00Z"
      }
    ]
  }
}
```

### 3. 下载文件

```http
GET /api/v1/cloud/files/100/download
```

**响应**: 文件二进制流

### 4. 重命名文件

```http
PUT /api/v1/cloud/files/100
Content-Type: application/json

{
  "fileName": "new-document.pdf"
}
```

### 5. 删除文件

```http
DELETE /api/v1/cloud/files/100
```

### 6. 移动文件

```http
PUT /api/v1/cloud/files/100/move
Content-Type: application/json

{
  "targetFolderId": 2
}
```

---

## 分享管理

### 1. 创建分享链接

```http
POST /api/v1/cloud/shares
Content-Type: application/json

{
  "fileId": 100,
  "shareType": "link",
  "expireDays": 7,
  "password": "1234",
  "allowDownload": true
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "shareCode": "abc123",
    "shareUrl": "http://localhost:9090/share/abc123",
    "expireTime": "2026-02-04T10:00:00Z",
    "password": "1234"
  }
}
```

### 2. 获取分享列表

```http
GET /api/v1/cloud/shares?page=1&pageSize=20
```

### 3. 访问分享链接

```http
GET /api/v1/cloud/shares/abc123
```

**如果有密码**:
```http
POST /api/v1/cloud/shares/abc123/verify
Content-Type: application/json

{
  "password": "1234"
}
```

### 4. 下载分享文件

```http
GET /api/v1/cloud/shares/abc123/download
```

### 5. 取消分享

```http
DELETE /api/v1/cloud/shares/1
```

---

## 配额管理

### 1. 获取配额信息

```http
GET /api/v1/cloud/quota
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "totalQuota": 10737418240,
    "usedSpace": 5368709120,
    "fileCount": 150,
    "folderCount": 20,
    "maxFileSize": 21474836480,
    "quotaType": "standard",
    "usagePercent": 50.0
  }
}
```

### 2. 更新配额（管理员）

```http
PUT /api/v1/cloud/quota/user/123
Content-Type: application/json

{
  "totalQuota": 107374182400,
  "maxFileSize": 53687091200,
  "quotaType": "premium"
}
```

---

## 搜索

### 1. 搜索文件

```http
GET /api/v1/cloud/search?keyword=document&page=1&pageSize=20
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "total": 10,
    "items": [
      {
        "id": 100,
        "fileName": "document.pdf",
        "fileSize": 1048576,
        "folderPath": "/我的文档/工作",
        "createTime": "2026-01-28T10:00:00Z"
      }
    ]
  }
}
```

---

## 批量操作

### 1. 批量删除文件

```http
POST /api/v1/cloud/files/batch/delete
Content-Type: application/json

{
  "fileIds": [100, 101, 102]
}
```

### 2. 批量移动文件

```http
POST /api/v1/cloud/files/batch/move
Content-Type: application/json

{
  "fileIds": [100, 101, 102],
  "targetFolderId": 2
}
```

### 3. 批量下载（打包）

```http
POST /api/v1/cloud/files/batch/download
Content-Type: application/json

{
  "fileIds": [100, 101, 102]
}
```

**响应**: ZIP 文件流

---

## 错误码

| 错误码 | 说明 |
|-------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 413 | 文件过大 |
| 507 | 存储空间不足 |
| 500 | 服务器内部错误 |

**错误响应格式**:
```json
{
  "code": 400,
  "message": "文件名不能为空",
  "data": null
}
```

---

## 快速开始示例

### cURL 示例

```bash
# 1. 创建文件夹
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"parentId": 0, "folderName": "我的文档"}'

# 2. 上传文件
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@document.pdf" \
  -F "folderId=1"

# 3. 下载文件
curl -X GET http://localhost:9090/api/v1/cloud/files/100/download \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o document.pdf

# 4. 创建分享
curl -X POST http://localhost:9090/api/v1/cloud/shares \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"fileId": 100, "shareType": "link", "expireDays": 7}'
```

### JavaScript 示例

```javascript
const API_BASE = 'http://localhost:9090/api/v1/cloud';
const token = 'YOUR_TOKEN';

// 上传文件
async function uploadFile(file, folderId) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('folderId', folderId);

  const response = await fetch(`${API_BASE}/files`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  return await response.json();
}

// 获取文件列表
async function getFiles(folderId, page = 1) {
  const response = await fetch(
    `${API_BASE}/files?folderId=${folderId}&page=${page}&pageSize=20`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );

  return await response.json();
}

// 创建分享
async function createShare(fileId, expireDays = 7) {
  const response = await fetch(`${API_BASE}/shares`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      fileId,
      shareType: 'link',
      expireDays,
      allowDownload: true
    })
  });

  return await response.json();
}
```

---

## 相关文档

- `上传功能指南.md` - 上传功能详细指南
- `分片上传.md` - 分片上传和断点续传
- `resumable-upload-analysis.md` - 断点续传功能分析

---

**最后更新**: 2026-01-28
