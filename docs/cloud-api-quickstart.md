# 云盘 API 快速开始

## API 端点列表

### 文件夹管理
- `POST /api/v1/cloud/folders` - 创建文件夹
- `GET /api/v1/cloud/folders` - 列出文件夹
- `GET /api/v1/cloud/folders/tree` - 获取文件夹树
- `DELETE /api/v1/cloud/folders/:id` - 删除文件夹
- `PUT /api/v1/cloud/folders/:id/rename` - 重命名文件夹

### 文件管理
- `POST /api/v1/cloud/files/upload` - 上传文件
- `GET /api/v1/cloud/files/:id/download` - 下载文件
- `GET /api/v1/cloud/files` - 列出文件（支持分页）
- `DELETE /api/v1/cloud/files/:id` - 删除文件
- `PUT /api/v1/cloud/files/:id/move` - 移动文件
- `PUT /api/v1/cloud/files/:id/rename` - 重命名文件

### 分享管理
- `POST /api/v1/cloud/shares` - 创建分享
- `GET /api/v1/cloud/shares/:code` - 获取分享信息
- `POST /api/v1/cloud/shares/:code/access` - 访问分享
- `DELETE /api/v1/cloud/shares/:id` - 取消分享

### 配额管理
- `GET /api/v1/cloud/quota` - 获取用户配额

## 快速示例

### 1. 创建文件夹

```bash
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "folderName": "我的文档",
    "parentId": null,
    "description": "存放工作文档"
  }'
```

### 2. 上传文件

```bash
curl -X POST http://localhost:9090/api/v1/cloud/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/file.pdf" \
  -F "folderId=1"
```

### 3. 列出文件

```bash
curl -X GET "http://localhost:9090/api/v1/cloud/files?folderId=1&page=1&pageSize=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. 下载文件

```bash
curl -X GET http://localhost:9090/api/v1/cloud/files/100/download \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -O
```

### 5. 创建分享

```bash
curl -X POST http://localhost:9090/api/v1/cloud/shares \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resourceType": "file",
    "resourceId": 100,
    "shareType": "password",
    "password": "abc123",
    "expireDays": 7
  }'
```

### 6. 获取配额

```bash
curl -X GET http://localhost:9090/api/v1/cloud/quota \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## JavaScript 示例

```javascript
// 上传文件
const uploadFile = async (file, folderId = null) => {
  const formData = new FormData();
  formData.append('file', file);
  if (folderId) {
    formData.append('folderId', folderId);
  }

  const response = await fetch('/api/v1/cloud/files/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  return await response.json();
};

// 列出文件
const listFiles = async (folderId = null, page = 1, pageSize = 20) => {
  const params = new URLSearchParams({
    page: page,
    pageSize: pageSize
  });
  if (folderId !== null) {
    params.append('folderId', folderId);
  }

  const response = await fetch(`/api/v1/cloud/files?${params}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });

  return await response.json();
};

// 创建分享
const createShare = async (fileId, options = {}) => {
  const response = await fetch('/api/v1/cloud/shares', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      resourceType: 'file',
      resourceId: fileId,
      shareType: options.password ? 'password' : 'public',
      password: options.password || '',
      expireDays: options.expireDays || 7,
      maxDownloads: options.maxDownloads || 0
    })
  });

  return await response.json();
};
```

## 完整文档

详细的 API 文档请查看:
- [完整 API 文档](./cloud-api.md)
- [Swagger UI](http://localhost:9090/swagger/index.html)

## 注意事项

1. 所有接口（除了获取分享信息）都需要 JWT 认证
2. 文件上传使用 `multipart/form-data` 格式
3. 分页接口默认每页 20 条，最大 100 条
4. 分享支持三种类型：`public`（公开）、`password`（密码保护）、`private`（私有）
5. 资源类型支持：`file`（文件）和 `folder`（文件夹）
