# 云盘服务 API 文档

本文档提供云盘服务的完整 REST API 接口说明,供前端开发使用。

## 基础信息

- **Base URL**: `/api/v1/cloud`
- **认证方式**: JWT Bearer Token
- **请求头**: `Authorization: Bearer {token}`

## 目录

1. [文件夹管理](#文件夹管理)
2. [文件管理](#文件管理)
3. [分享管理](#分享管理)
4. [配额管理](#配额管理)

---

## 文件夹管理

### 1. 创建文件夹

创建新文件夹。

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

---

### 2. 列出文件夹

列出指定父文件夹下的所有子文件夹。

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
    },
    {
      "ID": 2,
      "FolderName": "照片",
      "ParentID": 0,
      "UserID": 100,
      "CreateTime": "2026-01-13T10:05:00Z"
    }
  ]
}
```

---

### 3. 获取文件夹树

获取用户的完整文件夹树结构。

**请求**

```http
GET /api/v1/cloud/folders/tree
Authorization: Bearer {token}
```

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
      "Children": [
        {
          "ID": 3,
          "FolderName": "工作文档",
          "ParentID": 1,
          "UserID": 100,
          "Children": []
        }
      ]
    },
    {
      "ID": 2,
      "FolderName": "照片",
      "ParentID": 0,
      "UserID": 100,
      "Children": []
    }
  ]
}
```

---

### 4. 删除文件夹

删除指定文件夹（包括其中的所有文件和子文件夹）。

**请求**

```http
DELETE /api/v1/cloud/folders/{id}
Authorization: Bearer {token}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件夹ID |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "删除文件夹成功"
  }
}
```

---

### 5. 重命名文件夹

重命名指定文件夹。

**请求**

```http
PUT /api/v1/cloud/folders/{id}/rename
Content-Type: application/json
Authorization: Bearer {token}

{
  "newName": "新文件夹名称"
}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件夹ID |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "重命名文件夹成功"
  }
}
```

---

## 文件管理

### 1. 上传文件

上传文件到指定文件夹。

**请求**

```http
POST /api/v1/cloud/files/upload
Content-Type: multipart/form-data
Authorization: Bearer {token}

file: [文件二进制数据]
folderId: 0  // 可选，默认为0（根目录）
```

**表单参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 要上传的文件 |
| folderId | int | 否 | 目标文件夹ID，默认为0 |

**响应**

```json
{
  "code": 201,
  "message": "success",
  "data": {
    "ID": 100,
    "FileName": "report.pdf",
    "FileSize": 1024000,
    "FileType": "application/pdf",
    "FolderID": 0,
    "UserID": 100,
    "StoragePath": "/uploads/cloud/xxx.pdf",
    "CreateTime": "2026-01-13T10:00:00Z"
  }
}
```

---

### 2. 下载文件

下载指定文件。

**请求**

```http
GET /api/v1/cloud/files/{id}/download
Authorization: Bearer {token}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应**

文件二进制流，浏览器将自动下载文件。

**响应头**

```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="report.pdf"
```

---

### 3. 删除文件

删除指定文件。

**请求**

```http
DELETE /api/v1/cloud/files/{id}
Authorization: Bearer {token}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "删除文件成功"
  }
}
```

---

### 4. 移动文件

移动文件到指定文件夹。

**请求**

```http
PUT /api/v1/cloud/files/{id}/move
Content-Type: application/json
Authorization: Bearer {token}

{
  "targetFolderId": 5  // 目标文件夹ID
}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "移动文件成功"
  }
}
```

---

### 5. 重命名文件

重命名指定文件。

**请求**

```http
PUT /api/v1/cloud/files/{id}/rename
Content-Type: application/json
Authorization: Bearer {token}

{
  "newName": "新文件名.pdf"
}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文件ID |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "重命名文件成功"
  }
}
```

---

### 6. 列出文件

列出指定文件夹下的所有文件。

**请求**

```http
GET /api/v1/cloud/files?folderId=0
Authorization: Bearer {token}
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| folderId | int | 否 | 文件夹ID，默认为0（根目录） |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "ID": 100,
      "FileName": "report.pdf",
      "FileSize": 1024000,
      "FileType": "application/pdf",
      "FolderID": 0,
      "UserID": 100,
      "CreateTime": "2026-01-13T10:00:00Z"
    },
    {
      "ID": 101,
      "FileName": "photo.jpg",
      "FileSize": 2048000,
      "FileType": "image/jpeg",
      "FolderID": 0,
      "UserID": 100,
      "CreateTime": "2026-01-13T10:05:00Z"
    }
  ]
}
```

---

## 分享管理

### 1. 创建分享

创建文件分享链接。

**请求**

```http
POST /api/v1/cloud/shares
Content-Type: application/json
Authorization: Bearer {token}

{
  "fileId": 100,          // 文件ID
  "expireDays": 7,        // 过期天数，0表示永久
  "password": "abc123"    // 访问密码（可选）
}
```

**响应**

```json
{
  "code": 201,
  "message": "success",
  "data": {
    "ID": 1,
    "FileID": 100,
    "ShareCode": "xY9zK2",
    "Password": "abc123",
    "ExpireTime": "2026-01-20T10:00:00Z",
    "UserID": 100,
    "CreateTime": "2026-01-13T10:00:00Z"
  }
}
```

**分享链接格式**: `https://your-domain.com/share/{shareCode}`

---

### 2. 获取分享信息

根据分享码获取分享信息（不需要认证）。

**请求**

```http
GET /api/v1/cloud/shares/{code}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| code | string | 分享码 |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "FileName": "report.pdf",
    "FileSize": 1024000,
    "FileType": "application/pdf",
    "HasPassword": true,
    "ExpireTime": "2026-01-20T10:00:00Z"
  }
}
```

---

### 3. 访问分享

访问分享文件（需要密码验证）。

**请求**

```http
POST /api/v1/cloud/shares/{code}/access
Content-Type: application/json

{
  "password": "abc123"  // 如果分享有密码
}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| code | string | 分享码 |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "ID": 100,
    "FileName": "report.pdf",
    "FileSize": 1024000,
    "FileType": "application/pdf",
    "StoragePath": "/uploads/cloud/xxx.pdf",
    "CreateTime": "2026-01-13T10:00:00Z"
  }
}
```

---

### 4. 取消分享

取消指定的文件分享。

**请求**

```http
DELETE /api/v1/cloud/shares/{id}
Authorization: Bearer {token}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分享ID |

**响应**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "取消分享成功"
  }
}
```

---

## 配额管理

### 1. 获取用户配额

获取当前用户的存储配额和使用情况。

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
    "UserID": 100,
    "TotalQuota": 10737418240,      // 总配额（字节），10GB
    "UsedQuota": 1073741824,        // 已使用（字节），1GB
    "AvailableQuota": 9663676416,   // 可用配额（字节），9GB
    "UsagePercentage": 10.0,        // 使用百分比
    "FileCount": 25                 // 文件数量
  }
}
```

---

## 错误响应

所有 API 在发生错误时都会返回统一格式的错误响应：

```json
{
  "code": 400,
  "message": "error message",
  "data": null
}
```

**常见错误码**

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权（未登录或 token 无效） |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 前端使用示例

### JavaScript/Axios 示例

```javascript
// 1. 创建文件夹
async function createFolder(parentId, folderName) {
    const response = await axios.post('/api/v1/cloud/folders', {
        parentId: parentId,
        folderName: folderName
    }, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.data;
}

// 2. 上传文件
async function uploadFile(file, folderId = 0) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('folderId', folderId);

    const response = await axios.post('/api/v1/cloud/files/upload', formData, {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'multipart/form-data'
        }
    });
    return response.data;
}

// 3. 列出文件
async function listFiles(folderId = 0) {
    const response = await axios.get('/api/v1/cloud/files', {
        params: { folderId: folderId },
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.data;
}

// 4. 下载文件
function downloadFile(fileId, fileName) {
    window.open(`/api/v1/cloud/files/${fileId}/download`, '_blank');
}

// 5. 创建分享
async function createShare(fileId, expireDays = 7, password = '') {
    const response = await axios.post('/api/v1/cloud/shares', {
        fileId: fileId,
        expireDays: expireDays,
        password: password
    }, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.data;
}

// 6. 获取配额信息
async function getQuota() {
    const response = await axios.get('/api/v1/cloud/quota', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    return response.data;
}
```

### React Hook 示例

```jsx
import { useState, useEffect } from 'react';
import axios from 'axios';

// 自定义 Hook：获取文件列表
function useFiles(folderId) {
    const [files, setFiles] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchFiles = async () => {
            try {
                setLoading(true);
                const response = await axios.get('/api/v1/cloud/files', {
                    params: { folderId },
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                setFiles(response.data.data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchFiles();
    }, [folderId]);

    return { files, loading, error };
}

// 使用示例
function FileList({ folderId }) {
    const { files, loading, error } = useFiles(folderId);

    if (loading) return <div>加载中...</div>;
    if (error) return <div>错误: {error}</div>;

    return (
        <ul>
            {files.map(file => (
                <li key={file.ID}>{file.FileName} ({file.FileSize} bytes)</li>
            ))}
        </ul>
    );
}
```

---

## Swagger 文档

本项目已集成 Swagger 文档，可以在浏览器中访问可交互的 API 文档：

```
http://localhost:9090/swagger/index.html
```

在 Swagger UI 中，你可以：
- 查看所有 API 接口详细信息
- 直接在浏览器中测试 API
- 查看请求/响应示例
- 导出 OpenAPI 规范文件

---

## 注意事项

1. **文件大小限制**: 上传文件大小受服务器配置限制，默认为配置文件中的 `file.maxFileSize`
2. **文件类型限制**: 允许上传的文件类型由配置文件中的 `file.allowedExts` 控制
3. **配额限制**: 用户上传文件时会检查配额，超出配额将无法上传
4. **分享过期**: 分享链接会在设置的过期时间后自动失效
5. **删除操作**: 删除文件夹会级联删除其中的所有文件和子文件夹，操作不可逆
6. **认证要求**: 除了获取分享信息和访问分享，其他所有接口都需要 JWT 认证

---

## 更新日志

### v1.0.0 (2026-01-13)
- 初始版本发布
- 支持文件夹管理（创建、列表、树形结构、删除、重命名）
- 支持文件管理（上传、下载、删除、移动、重命名、列表）
- 支持分享管理（创建、获取信息、访问、取消）
- 支持配额管理（查询使用情况）
