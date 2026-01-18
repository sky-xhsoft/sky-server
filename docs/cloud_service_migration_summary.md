# CloudService Migration Summary

## 概述

成功将 `cloud_service.go` 中的所有 `CloudFile` 引用迁移到 `CloudItem`，解决了 "Table 'skyserver.cloud_file' doesn't exist" 错误。

## 迁移日期

2026-01-17

## 文件修改

### 1. `internal/service/cloud/cloud_service.go`

#### 接口定义修改
- `UploadFile` 返回类型: `*entity.CloudFile` → `*entity.CloudItem`
- `DownloadFile` 返回类型: `*entity.CloudFile` → `*entity.CloudItem`
- `ListFiles` 返回类型: `[]*entity.CloudFile` → `[]*entity.CloudItem`

#### 方法迁移详情

**1. UploadFile 方法**
- 创建 `CloudItem` 结构体而非 `CloudFile`
- 添加 `ItemType: "file"` 标识
- 字段映射:
  - `FileName` → `Name`
  - `FolderID` → `ParentID`
- 文件专属字段使用指针:
  - `StorageType: &storageType`
  - `StoragePath: &storagePath`
  - `FileSize: &fileSize`
  - `FileType: &fileType`
  - `FileExt: &ext`
  - `AccessURL: &accessURL`

**2. getFileByID 辅助方法**
- 使用 `CloudItem` 表
- 添加 `ITEM_TYPE = "file"` 过滤条件
- WHERE 条件: `ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?`

**3. DownloadFile 方法**
- 使用指针解引用访问 `StoragePath`: `*file.StoragePath`
- 添加空指针检查
- 更新下载计数查询添加 `ITEM_TYPE = "file"` 过滤

**4. DeleteFile 方法**
- 软删除查询添加 `ITEM_TYPE = "file"` 过滤
- 配额更新处理指针: `*file.FileSize`
- 异步删除物理文件添加指针检查和解引用

**5. MoveFile 方法**
- 检查同名文件时使用:
  - `PARENT_ID` 替代 `FOLDER_ID`
  - `NAME` 替代 `FILE_NAME`
  - 添加 `ITEM_TYPE = "file"` 过滤
- 更新查询添加 `ITEM_TYPE = "file"` 过滤

**6. RenameFile 方法**
- 检查同名文件使用 `PARENT_ID` 和 `NAME`
- 添加 `ITEM_TYPE = "file"` 过滤
- 更新字段从 `FILE_NAME` 改为 `NAME`

**7. ListFiles 方法**
- Model 使用 `CloudItem`
- WHERE 条件添加 `ITEM_TYPE = "file"`
- `FOLDER_ID` 改为 `PARENT_ID`

**8. DeleteFolder 方法**
- 检查文件时使用 `CloudItem` 表
- WHERE 条件: `PARENT_ID = ? AND ITEM_TYPE = "file" AND IS_ACTIVE = ?`

**9. GetShareInfo 方法**
- 加载文件资源时使用 `CloudItem`
- 添加 `ITEM_TYPE = "file"` 过滤

**10. updateChildrenPaths 方法**
- SQL 更新从 `cloud_file` 改为 `cloud_item`

### 2. `api/handler/cloud_handler.go`

#### 字段访问修改
- `fileInfo.FileName` → `fileInfo.Name`
- `fileInfo.FileType` → `*fileInfo.FileType` (指针解引用)
- `fileInfo.FileSize` → `*fileInfo.FileSize` (指针解引用)

#### 类型定义修改
- `filesResult.files` 类型: `[]*entity.CloudFile` → `[]*entity.CloudItem`
- 响应结构中 `Files` 字段类型: `[]*entity.CloudFile` → `[]*entity.CloudItem`

## 关键迁移模式

### 字段名映射
```
CloudFile          →  CloudItem
-----------------     -----------------
FileName           →  Name
FolderID           →  ParentID
FILE_NAME (SQL)    →  NAME
FOLDER_ID (SQL)    →  PARENT_ID
```

### 查询条件模式
```go
// 旧模式
Model(&entity.CloudFile{}).
Where("FOLDER_ID = ? AND IS_ACTIVE = ?", folderID, "Y")

// 新模式
Model(&entity.CloudItem{}).
Where("PARENT_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", folderID, "file", "Y")
```

### 指针字段访问模式
```go
// CloudItem 中文件专属字段是指针类型
// 创建时
storageType := req.StorageType
item := &entity.CloudItem{
    StorageType: &storageType,  // 传递指针
}

// 访问时
if file.StoragePath != nil {
    path := *file.StoragePath   // 解引用
}
```

## 编译验证

```bash
go build -o sky-server.exe ./cmd/server
# ✅ 编译成功，无错误
```

## 服务器启动验证

```bash
./sky-server.exe
# ✅ 服务器成功启动
# ✅ 所有路由注册成功
# ✅ 健康检查端点响应正常
```

## 测试建议

1. **上传文件测试**: 验证新 CloudItem 记录正确创建
2. **下载文件测试**: 验证指针字段正确解引用
3. **列表文件测试**: 验证 ITEM_TYPE 过滤正常工作
4. **移动/重命名测试**: 验证 PARENT_ID 和 NAME 字段正确使用
5. **删除文件测试**: 验证软删除和配额更新正常
6. **文件夹操作测试**: 验证文件检查使用 CloudItem 表

## 影响范围

- ✅ 所有文件操作 API 使用统一的 `cloud_item` 表
- ✅ 数据模型统一，便于后续维护
- ✅ 支持未来扩展（文件夹也可以迁移到 CloudItem）

## 备份

迁移前的备份文件:
- `cloud_service.go.bak_before_migration`
- `cloud_service.go.bak2`

## 下一步

可以考虑将其他使用 CloudFile 的代码也迁移到 CloudItem，实现完全统一。
