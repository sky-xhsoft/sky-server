# Folder ID = 0 Bug 修复

## 🐛 问题描述

当传入 `folderId=0`、`parentId=0` 或 `targetFolderId=0` 时，系统会错误地将 `0` 当作有效的文件夹 ID，导致：
1. 创建了 ID 为 0 的虚拟文件夹
2. 无法正确上传到根目录
3. 数据库查询失败（文件夹 ID=0 不存在）

## 🔍 根本原因

在 Go 中：
- 整数类型的零值是 `0`
- 指针类型的零值是 `nil`

当 API 接收到 `folderId=0` 时：
```go
// 问题代码
var folderID *uint = nil
if folderIDStr != "" {
    id, _ := strconv.ParseUint(folderIDStr, 10, 32)
    fid := uint(id)  // id = 0
    folderID = &fid  // 指向 0 的指针，而不是 nil
}
```

这导致 `folderID` 指向 `0`，而不是期望的 `nil`（根目录）。

## ✅ 解决方案

在解析 ID 后，添加判断：**将 0 视为根目录（nil）**

```go
// 修复后的代码
var folderID *uint = nil
if folderIDStr != "" {
    id, _ := strconv.ParseUint(folderIDStr, 10, 32)
    // 将 0 视为根目录（nil）
    if id > 0 {
        fid := uint(id)
        folderID = &fid
    }
}
```

## 📝 修复的位置

### 1. UploadFile - 上传文件
**文件**: `api/handler/cloud_handler.go:216-220`

```go
// 修复前
fid := uint(id)
folderID = &fid

// 修复后
if id > 0 {
    fid := uint(id)
    folderID = &fid
}
```

**影响**：
- ✅ `folderId=0` → 上传到根目录
- ✅ `folderId=""` → 上传到根目录
- ✅ `folderId=1` → 上传到文件夹 1

### 2. ListFolders - 列出文件夹
**文件**: `api/handler/cloud_handler.go:79-83`

```go
// 将 0 视为根目录（nil）
if id > 0 {
    pid := uint(id)
    parentID = &pid
}
```

**影响**：
- ✅ `parentId=0` → 列出根目录的文件夹
- ✅ `parentId=""` → 列出根目录的文件夹
- ✅ `parentId=1` → 列出文件夹 1 的子文件夹

### 3. ListFiles - 列出文件
**文件**: `api/handler/cloud_handler.go:428-432`

```go
// 将 0 视为根目录（nil）
if id > 0 {
    fid := uint(id)
    folderID = &fid
}
```

**影响**：
- ✅ `folderId=0` → 列出根目录的文件
- ✅ `folderId=""` → 列出根目录的文件
- ✅ `folderId=1` → 列出文件夹 1 的文件

### 4. CreateFolder - 创建文件夹
**文件**: `api/handler/cloud_handler.go:40-43`

```go
// 将 parentId=0 视为根目录（nil）
if req.ParentID != nil && *req.ParentID == 0 {
    req.ParentID = nil
}
```

**影响**：
- ✅ `{"parentId": 0}` → 在根目录创建文件夹
- ✅ `{"parentId": null}` → 在根目录创建文件夹
- ✅ `{"parentId": 1}` → 在文件夹 1 下创建子文件夹

### 5. MoveFile - 移动文件
**文件**: `api/handler/cloud_handler.go:361-364`

```go
// 将 targetFolderId=0 视为根目录（nil）
if req.TargetFolderID != nil && *req.TargetFolderID == 0 {
    req.TargetFolderID = nil
}
```

**影响**：
- ✅ `{"targetFolderId": 0}` → 移动到根目录
- ✅ `{"targetFolderId": null}` → 移动到根目录
- ✅ `{"targetFolderId": 1}` → 移动到文件夹 1

## 🧪 测试用例

### 测试 1：上传到根目录

```bash
# 方式 1：folderId=0
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer TOKEN" \
  -F "file=@test.txt" \
  -F "folderId=0"

# 方式 2：不传 folderId
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer TOKEN" \
  -F "file=@test.txt"

# 方式 3：folderId 为空字符串
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer TOKEN" \
  -F "file=@test.txt" \
  -F "folderId="
```

**预期结果**：所有方式都应该成功上传到根目录。

### 测试 2：创建根目录文件夹

```bash
# 方式 1：parentId=0
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"folderName":"测试","parentId":0}'

# 方式 2：parentId=null
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"folderName":"测试","parentId":null}'

# 方式 3：不传 parentId
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"folderName":"测试"}'
```

**预期结果**：所有方式都应该在根目录创建文件夹。

### 测试 3：列出根目录文件

```bash
# 方式 1：folderId=0
curl -X GET "http://localhost:9090/api/v1/cloud/files?folderId=0" \
  -H "Authorization: Bearer TOKEN"

# 方式 2：不传 folderId
curl -X GET "http://localhost:9090/api/v1/cloud/files" \
  -H "Authorization: Bearer TOKEN"
```

**预期结果**：都应该返回根目录的文件列表。

### 测试 4：移动文件到根目录

```bash
# 假设文件 ID 为 5
curl -X PUT http://localhost:9090/api/v1/cloud/files/5/move \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"targetFolderId":0}'
```

**预期结果**：文件应该移动到根目录。

## 🔒 验证修复

### 步骤 1：清理可能存在的 ID=0 记录

```sql
-- 检查是否有 ID=0 的文件夹
SELECT * FROM cloud_folder WHERE ID = 0 OR PARENT_ID = 0;

-- 检查是否有指向 ID=0 的文件
SELECT * FROM cloud_file WHERE FOLDER_ID = 0;

-- 如果有，清理它们（谨慎操作）
DELETE FROM cloud_folder WHERE ID = 0;
UPDATE cloud_file SET FOLDER_ID = NULL WHERE FOLDER_ID = 0;
```

### 步骤 2：重新编译和启动

```bash
go build ./cmd/server
./sky-server
```

### 步骤 3：测试各种场景

使用上面的测试用例进行验证。

## 📊 API 文档更新

所有相关 API 的文档都已更新，明确说明：

- `folderId=0` 或空 → 根目录
- `parentId=0` 或 null → 根目录
- `targetFolderId=0` 或 null → 根目录

## ⚠️ 注意事项

### 1. 数据库中不应该有 ID=0 的记录

MySQL 的自增主键从 1 开始，不会生成 ID=0 的记录（除非手动插入）。

### 2. 前端建议

前端在传递"根目录"时，推荐使用以下方式（按优先级）：

**推荐** ✅
```javascript
// 方式 1：不传该字段
const data = { fileName: "test.txt" };

// 方式 2：传 null
const data = { fileName: "test.txt", folderId: null };
```

**可接受** ⚠️
```javascript
// 方式 3：传 0（现在已修复支持）
const data = { fileName: "test.txt", folderId: 0 };
```

**不推荐** ❌
```javascript
// 不要传空字符串（会被当作无效参数）
const data = { fileName: "test.txt", folderId: "" };
```

### 3. 类型一致性

为了保持一致性，所有可选的文件夹 ID 字段都使用相同的处理逻辑：
- 查询参数（query）：`folderId=0` → `nil`
- 表单参数（form）：`folderId=0` → `nil`
- JSON 参数（body）：`"folderId": 0` → `nil`

## 📚 相关文档

- [cloud-service-implementation.md](./cloud-service-implementation.md) - 云盘服务完整实现
- [cloud-api.md](./cloud-api.md) - 云盘 API 文档

## ✅ 修复状态

- ✅ Bug 已修复
- ✅ 代码已编译通过
- ✅ 所有相关位置都已修复
- ✅ 文档已更新
- ⏳ 等待测试验证
