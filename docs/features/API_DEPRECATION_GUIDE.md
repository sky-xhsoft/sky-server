# 云盘API废弃与迁移指南

## 📅 重要日期

- **废弃日期**: 2026-01-17
- **计划移除日期**: 2026-06-01 (6个月后)
- **迁移截止日期**: 2026-05-31

## 🎯 废弃原因

为了简化API设计和提高一致性，我们将云盘的文件和文件夹管理合并为统一的接口。新的 `/cloud/items` 接口使用统一的数据模型，减少了代码重复，提高了可维护性。

## ⚠️ 废弃接口列表

以下接口已被标记为废弃，但仍然可用（向后兼容）：

### 文件夹管理接口

| 废弃接口 | 新接口 | 说明 |
|---------|--------|------|
| `POST /api/v1/cloud/folders` | `POST /api/v1/cloud/items` | 创建文件夹 |
| `GET /api/v1/cloud/folders` | `GET /api/v1/cloud/items` | 获取文件夹列表 |
| `GET /api/v1/cloud/folders/tree` | `GET /api/v1/cloud/items` | 获取文件夹树 |
| `GET /api/v1/cloud/folders/content` | `GET /api/v1/cloud/items?parentId={id}` | 获取文件夹内容 |
| `DELETE /api/v1/cloud/folders/:id` | `DELETE /api/v1/cloud/items/:id` | 删除文件夹 |
| `PUT /api/v1/cloud/folders/:id/rename` | `PUT /api/v1/cloud/items/:id/rename` | 重命名文件夹 |

### 文件管理接口

| 废弃接口 | 新接口 | 说明 |
|---------|--------|------|
| `GET /api/v1/cloud/files` | `GET /api/v1/cloud/items` | 获取文件列表 |
| `DELETE /api/v1/cloud/files/:id` | `DELETE /api/v1/cloud/items/:id` | 删除文件 |
| `PUT /api/v1/cloud/files/:id/move` | `PUT /api/v1/cloud/items/:id/move` | 移动文件 |
| `PUT /api/v1/cloud/files/:id/rename` | `PUT /api/v1/cloud/items/:id/rename` | 重命名文件 |

### 批量操作接口

| 废弃接口 | 新接口 | 说明 |
|---------|--------|------|
| `POST /api/v1/cloud/batch/delete` | `POST /api/v1/cloud/items/batch/delete` | 批量删除 |
| `POST /api/v1/cloud/batch/move` | `POST /api/v1/cloud/items/batch/move` | 批量移动 |

### 保留接口（不废弃）

以下接口**不受影响**，继续使用：

- `POST /api/v1/cloud/files/upload` - 文件上传
- `GET /api/v1/cloud/files/:id/download` - 文件下载
- `POST /api/v1/cloud/files/multipart/*` - 分片上传
- `GET /api/v1/cloud/quota` - 配额查询
- `/api/v1/cloud/shares/*` - 分享管理

## 🔄 迁移指南

### 1. 识别废弃API调用

废弃的API会在响应头中包含以下信息：

```http
X-API-Deprecated: true
X-API-Deprecated-Date: 2026-01-17
X-API-New-Endpoint: POST /api/v1/cloud/items
Warning: 299 - "This API endpoint is deprecated and will be removed in a future version. Please use POST /api/v1/cloud/items instead."
```

### 2. 迁移示例

#### 示例 1: 获取文件夹列表

**旧方式（废弃）**:
```javascript
// 分别获取文件夹和文件
const folders = await axios.get('/api/v1/cloud/folders', {
  params: { parentId: 0 }
});
const files = await axios.get('/api/v1/cloud/files', {
  params: { folderId: 0 }
});
```

**新方式（推荐）**:
```javascript
// 一次性获取文件夹和文件
const { folders, files } = await axios.get('/api/v1/cloud/items', {
  params: { parentId: 0 } // parentId 可选，不传则获取根目录
});
```

#### 示例 2: 创建文件夹

**旧方式（废弃）**:
```javascript
await axios.post('/api/v1/cloud/folders', {
  folderName: '新文件夹',
  parentId: 10
});
```

**新方式（推荐）**:
```javascript
await axios.post('/api/v1/cloud/items', {
  itemType: 'folder',  // 必须指定类型
  name: '新文件夹',
  parentId: 10
});
```

#### 示例 3: 删除文件或文件夹

**旧方式（废弃）**:
```javascript
// 删除文件夹
await axios.delete(`/api/v1/cloud/folders/${folderId}`);

// 删除文件
await axios.delete(`/api/v1/cloud/files/${fileId}`);
```

**新方式（推荐）**:
```javascript
// 统一的删除接口，自动识别类型
await axios.delete(`/api/v1/cloud/items/${itemId}`);
```

#### 示例 4: 批量删除

**旧方式（废弃）**:
```javascript
await axios.post('/api/v1/cloud/batch/delete', {
  fileIds: [1, 2, 3],
  folderIds: [10, 11]
});
```

**新方式（推荐）**:
```javascript
await axios.post('/api/v1/cloud/items/batch/delete', {
  itemIds: [1, 2, 3, 10, 11]  // 统一的ID数组
});

// 向后兼容：旧格式仍然支持
await axios.post('/api/v1/cloud/items/batch/delete', {
  fileIds: [1, 2, 3],
  folderIds: [10, 11]
});
```

### 3. 响应格式变化

#### 获取列表

**新接口返回格式**:
```json
{
  "code": 200,
  "data": {
    "folders": [
      {
        "id": 1,
        "itemType": "folder",
        "name": "我的文件夹",
        "parentId": null,
        "fileCount": 5,
        "totalSize": 1024000
      }
    ],
    "files": [
      {
        "id": 100,
        "itemType": "file",
        "name": "document.pdf",
        "parentId": 1,
        "fileSize": 204800,
        "fileType": "application/pdf"
      }
    ]
  }
}
```

#### CloudItem 字段说明

```typescript
interface CloudItem {
  id: number;
  itemType: 'file' | 'folder';  // 类型标识
  name: string;
  parentId: number | null;
  path: string;

  // 文件专用字段
  fileSize?: number;
  fileType?: string;
  fileExt?: string;
  storagePath?: string;

  // 文件夹专用字段
  fileCount?: number;
  totalSize?: number;

  // 共用字段
  isPublic: 'Y' | 'N';
  createTime: string;
  updateTime: string;
}
```

## 📝 前端代码更新清单

### Vue/React 项目

1. ✅ 更新 API 调用文件（如 `api/cloud.ts`）
2. ✅ 更新 store/state 管理（合并 folders 和 files 的处理逻辑）
3. ✅ 更新组件中的API调用
4. ✅ 测试所有云盘相关功能

### 更新 TypeScript 类型

```typescript
// 推荐：使用新的统一类型
import { CloudItem } from '@/types/cloud';

// 如果需要向后兼容
import { Folder, FileItem } from '@/types/cloud-legacy';
```

## ⚡ 性能优势

使用新接口的优势：

1. **减少请求次数**: 一次请求同时获取文件夹和文件，减少 50% 的 API 调用
2. **统一数据模型**: 前后端使用相同的数据结构，减少转换开销
3. **简化代码**: 减少重复的处理逻辑

## 🧪 测试建议

1. **并行运行**: 在迁移期间，新旧接口可以并行运行
2. **逐步迁移**: 建议按模块逐步迁移，而不是一次性全部替换
3. **监控废弃警告**: 在开发环境中监控 `Warning` 响应头，识别需要迁移的调用

### 测试脚本

```bash
# 检测项目中使用废弃API的位置
grep -r "cloud/folders" src/
grep -r "cloud/files" src/ | grep -v "upload\|download"
grep -r "cloud/batch" src/
```

## 🚨 常见问题

### Q1: 旧接口什么时候会被完全移除？

A: 计划在 2026-06-01 移除，届时旧接口将返回 410 Gone 状态码。

### Q2: 如果无法在截止日期前完成迁移怎么办？

A: 请联系技术支持团队，我们可以根据实际情况延长过渡期。

### Q3: 新接口是否完全向后兼容？

A: 是的，批量操作接口支持新旧两种请求格式。但建议尽快迁移到新格式。

### Q4: 文件上传和下载接口有变化吗？

A: 没有，上传和下载接口保持不变。

## 📞 技术支持

如有问题，请联系：

- **技术支持邮箱**: support@xhsoft.com
- **开发者文档**: https://docs.xhsoft.com/api/cloud
- **GitHub Issues**: https://github.com/sky-xhsoft/sky-server/issues

## 📚 相关文档

- [CloudItem API 完整文档](./CLOUD_API.md)
- [数据库迁移说明](../sqls/migrations/merge_cloud_tables.sql)
- [前端迁移示例](./FRONTEND_MIGRATION_EXAMPLES.md)

---

**最后更新**: 2026-01-17
**版本**: 1.0.0
