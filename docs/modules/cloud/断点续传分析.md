# Cloud 模块断点续传功能分析报告

## 📋 执行总结

**分析时间**: 2026-01-15
**模块**: Cloud（云盘）
**版本**: Phase 13+
**结论**: ❌ **不支持断点续传**

---

## 🔍 详细分析

### 一、当前实现情况

#### 1. 上传实现方式

**Handler层** (`api/handler/cloud_handler.go:209-262`):
```go
func (h *CloudHandler) UploadFile(c *gin.Context) {
    // 获取上传的文件
    fileHeader, err := c.FormFile("file")
    if err != nil {
        utils.BadRequest(c, "未找到上传文件")
        return
    }

    // 打开文件
    file, err := fileHeader.Open()
    if err != nil {
        utils.InternalError(c, "打开文件失败: " + err.Error())
        return
    }
    defer file.Close()

    // 构造上传请求（一次性传输整个文件）
    uploadReq := &cloud.UploadFileRequest{
        FileName:    fileHeader.Filename,
        FolderID:    folderID,
        FileSize:    fileHeader.Size,
        FileType:    fileHeader.Header.Get("Content-Type"),
        Reader:      file,              // ⚠️ 整个文件的 Reader
        StorageType: "local",
    }

    uploadedFile, err := h.cloudService.UploadFile(c.Request.Context(), uploadReq, userID.(uint))
    // ...
}
```

**Service层** (`internal/service/cloud/cloud_service.go:253-314`):
```go
func (s *service) UploadFile(ctx context.Context, req *UploadFileRequest, userID uint) (*entity.CloudFile, error) {
    // 1. 检查配额
    if err := s.CheckQuota(ctx, userID, req.FileSize); err != nil {
        return nil, err
    }

    // 2. 构建存储路径（UUID）
    ext := filepath.Ext(req.FileName)
    storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
    dateDir := time.Now().Format("2006/01/02")
    storagePath := fmt.Sprintf("cloud/%d/%s/%s", userID, dateDir, storageName)

    // 3. 上传到存储（⚠️ 一次性完整上传）
    accessURL, err := s.storage.Upload(ctx, storagePath, req.Reader, req.FileType)
    if err != nil {
        return nil, err
    }

    // 4. 创建文件记录
    file := &entity.CloudFile{
        FileName:    req.FileName,
        StoragePath: storagePath,
        FileSize:    req.FileSize,
        // ...
    }

    if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
        s.storage.Delete(ctx, storagePath) // ⚠️ 失败时删除，不保留已上传的部分
        return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
    }

    return file, nil
}
```

**Storage层** (`internal/pkg/storage/local_storage.go:40-67`):
```go
func (s *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
    fullPath := filepath.Join(s.basePath, path)

    // 创建文件
    file, err := os.Create(fullPath)
    if err != nil {
        return "", errors.Wrap(errors.ErrInternal, "创建文件失败", err)
    }
    defer file.Close()

    // ⚠️ 一次性写入整个文件，使用 io.Copy
    if _, err := io.Copy(file, reader); err != nil {
        os.Remove(fullPath) // ⚠️ 失败时删除文件，不保留已写入的部分
        return "", errors.Wrap(errors.ErrInternal, "写入文件失败", err)
    }

    url := fmt.Sprintf("%s/%s", s.baseURL, path)
    return url, nil
}
```

**OSS层** (`internal/pkg/storage/aliyun_oss.go:54-74`):
```go
func (s *AliyunOSS) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
    options := []oss.Option{
        oss.ContentType(contentType),
    }

    // ⚠️ 使用 PutObject 一次性上传整个文件
    err := s.bucket.PutObject(path, reader, options...)
    if err != nil {
        return "", errors.Wrap(errors.ErrInternal, "上传文件到OSS失败", err)
    }

    url, err := s.GetURL(ctx, path, 0)
    if err != nil {
        return "", err
    }

    return url, nil
}
```

#### 2. 下载实现方式

**Download Handler** (`api/handler/cloud_handler.go:264-304`):
```go
func (h *CloudHandler) DownloadFile(c *gin.Context) {
    reader, fileInfo, err := h.cloudService.DownloadFile(c.Request.Context(), uint(id), userID.(uint))
    if err != nil {
        utils.InternalError(c, "下载文件失败: " + err.Error())
        return
    }
    defer reader.Close()

    // ⚠️ 设置响应头，不支持 Range 请求
    c.Header("Content-Disposition", "attachment; filename=" + fileInfo.FileName)
    c.Header("Content-Type", fileInfo.FileType)
    c.Header("Content-Length", strconv.FormatInt(fileInfo.FileSize, 10))

    // ⚠️ 流式传输整个文件
    if _, err := io.Copy(c.Writer, reader); err != nil {
        utils.InternalError(c, "传输文件失败: " + err.Error())
        return
    }
}
```

**Download Service** (`internal/service/cloud/cloud_service.go:513-537`):
```go
func (s *service) DownloadFile(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, *entity.CloudFile, error) {
    file, err := s.getFileByID(ctx, fileID)
    if err != nil {
        return nil, nil, err
    }

    // 检查权限
    if file.OwnerID != userID {
        return nil, nil, errors.New(errors.ErrPermissionDenied, "无权限下载此文件")
    }

    // ⚠️ 从存储中下载完整文件
    reader, err := s.storage.Download(ctx, file.StoragePath)
    if err != nil {
        return nil, nil, err
    }

    // 更新下载次数
    s.db.WithContext(ctx).Model(&entity.CloudFile{}).
        Where("ID = ?", fileID).
        Update("DOWNLOAD_COUNT", gorm.Expr("DOWNLOAD_COUNT + 1"))

    return reader, file, nil
}
```

---

### 二、缺失的功能

#### ❌ 1. **分片上传（Multipart Upload）**

**现状**:
- 只支持单次完整上传
- 使用 `c.FormFile()` 获取整个文件
- 没有分片管理机制

**需要**:
```go
// 需要实现的接口
type ChunkUpload struct {
    FileID       string  // 文件唯一标识
    ChunkIndex   int     // 分片索引
    TotalChunks  int     // 总分片数
    ChunkSize    int64   // 分片大小
    ChunkData    []byte  // 分片数据
    ChunkMD5     string  // 分片MD5
}

// 需要的API
POST /api/v1/cloud/files/multipart/init     // 初始化分片上传
POST /api/v1/cloud/files/multipart/upload   // 上传单个分片
POST /api/v1/cloud/files/multipart/complete // 完成上传并合并
POST /api/v1/cloud/files/multipart/abort    // 取消上传
GET  /api/v1/cloud/files/multipart/status   // 查询上传状态
```

#### ❌ 2. **断点续传（Resume Upload）**

**现状**:
- 没有上传进度记录
- 上传失败后完全删除已上传的数据
- 没有分片状态跟踪

**需要**:
```go
// 需要的数据表
type UploadSession struct {
    ID           uint      // 会话ID
    UserID       uint      // 用户ID
    FileID       string    // 文件唯一标识（MD5或UUID）
    FileName     string    // 文件名
    FileSize     int64     // 文件总大小
    ChunkSize    int64     // 分片大小
    TotalChunks  int       // 总分片数
    UploadedChunks []int   // 已上传的分片索引列表
    Status       string    // 状态: uploading, paused, completed, failed
    ExpireTime   time.Time // 过期时间
    StoragePath  string    // 临时存储路径
}

// 分片记录
type ChunkRecord struct {
    SessionID   uint      // 会话ID
    ChunkIndex  int       // 分片索引
    ChunkMD5    string    // 分片MD5
    ChunkPath   string    // 分片存储路径
    Uploaded    bool      // 是否已上传
    UploadTime  time.Time // 上传时间
}
```

#### ❌ 3. **Range 请求支持**

**现状**:
```go
// 当前下载实现不支持 Range 请求
c.Header("Content-Disposition", "attachment; filename=" + fileInfo.FileName)
c.Header("Content-Type", fileInfo.FileType)
c.Header("Content-Length", strconv.FormatInt(fileInfo.FileSize, 10))
// ⚠️ 缺少 Accept-Ranges 头
// ⚠️ 不处理 Range 请求头

io.Copy(c.Writer, reader) // ⚠️ 始终传输整个文件
```

**需要**:
```go
// 需要支持 HTTP Range 请求
func (h *CloudHandler) DownloadFile(c *gin.Context) {
    // 1. 解析 Range 请求头
    rangeHeader := c.GetHeader("Range")

    // 2. 设置响应头
    c.Header("Accept-Ranges", "bytes")

    if rangeHeader != "" {
        // 3. 解析 Range（如: bytes=0-1023）
        start, end := parseRange(rangeHeader, fileSize)

        // 4. 设置 206 Partial Content 响应
        c.Status(http.StatusPartialContent)
        c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
        c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))

        // 5. 读取指定范围的数据
        reader.Seek(start, io.SeekStart)
        io.CopyN(c.Writer, reader, end-start+1)
    } else {
        // 6. 完整文件下载
        c.Status(http.StatusOK)
        c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
        io.Copy(c.Writer, reader)
    }
}
```

#### ❌ 4. **秒传功能（文件去重）**

**现状**:
```go
// CloudFile 实体中有 MD5 字段，但未使用
type CloudFile struct {
    // ...
    MD5 string `gorm:"column:MD5;size:32;index" json:"md5"` // ⚠️ 未使用
    // ...
}
```

**需要**:
```go
// 秒传功能实现
func (s *service) QuickUpload(ctx context.Context, fileMD5 string, fileName string, userID uint) (*entity.CloudFile, error) {
    // 1. 查询是否存在相同MD5的文件
    var existingFile entity.CloudFile
    err := s.db.WithContext(ctx).
        Where("MD5 = ? AND IS_ACTIVE = ?", fileMD5, "Y").
        First(&existingFile).Error

    if err == nil {
        // 2. 存在相同文件，复制记录（秒传）
        newFile := &entity.CloudFile{
            FileName:    fileName,
            StoragePath: existingFile.StoragePath, // 共享存储路径
            FileSize:    existingFile.FileSize,
            FileType:    existingFile.FileType,
            MD5:         fileMD5,
            OwnerID:     userID,
            // ...
        }

        // 3. 创建新记录，不实际上传文件
        s.db.WithContext(ctx).Create(newFile)

        // 4. 更新配额（空间已占用，只增加文件计数）
        s.UpdateQuota(ctx, userID, 0, 1)

        return newFile, nil
    }

    // 5. 不存在，需要正常上传
    return nil, errors.New(errors.ErrResourceNotFound, "文件不存在，需要上传")
}
```

#### ❌ 5. **上传进度跟踪**

**现状**:
- 没有进度回调机制
- 前端无法获取实时上传进度
- 使用 `io.Copy` 无法跟踪进度

**需要**:
```go
// 进度跟踪器
type ProgressTracker struct {
    FileID       string
    TotalSize    int64
    UploadedSize int64
    Progress     float64
    Speed        int64  // 字节/秒
    StartTime    time.Time
    EstimatedTime int64 // 预计剩余时间（秒）
}

// 带进度的 Reader
type ProgressReader struct {
    reader   io.Reader
    total    int64
    current  int64
    callback func(current, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    pr.current += int64(n)

    // 回调进度
    if pr.callback != nil {
        pr.callback(pr.current, pr.total)
    }

    return n, err
}
```

#### ❌ 6. **OSS 分片上传支持**

**现状**:
```go
// 只使用了 PutObject（普通上传）
err := s.bucket.PutObject(path, reader, options...)
```

**阿里云OSS SDK 支持但未使用**:
```go
// OSS SDK 提供的分片上传API（未使用）
- InitiateMultipartUpload()  // 初始化分片上传
- UploadPart()               // 上传分片
- CompleteMultipartUpload()  // 完成分片上传
- AbortMultipartUpload()     // 取消分片上传
- ListMultipartUploads()     // 列出未完成的分片上传
- ListUploadedParts()        // 列出已上传的分片
```

**需要实现**:
```go
func (s *AliyunOSS) MultipartUpload(ctx context.Context, path string, reader io.Reader, fileSize int64) (string, error) {
    // 1. 初始化分片上传
    imur, err := s.bucket.InitiateMultipartUpload(path)
    if err != nil {
        return "", err
    }

    // 2. 分片上传
    chunkSize := int64(5 * 1024 * 1024) // 5MB per chunk
    var parts []oss.UploadPart

    for partNum := 1; ; partNum++ {
        chunk := make([]byte, chunkSize)
        n, err := io.ReadFull(reader, chunk)

        if n > 0 {
            part, err := s.bucket.UploadPart(imur, bytes.NewReader(chunk[:n]), int64(n), partNum)
            if err != nil {
                s.bucket.AbortMultipartUpload(imur) // 取消上传
                return "", err
            }
            parts = append(parts, part)
        }

        if err == io.EOF || err == io.ErrUnexpectedEOF {
            break
        }
        if err != nil {
            s.bucket.AbortMultipartUpload(imur)
            return "", err
        }
    }

    // 3. 完成分片上传
    _, err = s.bucket.CompleteMultipartUpload(imur, parts)
    if err != nil {
        return "", err
    }

    return s.GetURL(ctx, path, 0)
}
```

---

### 三、大文件支持分析

#### ✅ 已实现的功能

1. **流式传输**:
   ```go
   // 使用 io.Reader/io.Writer 接口，支持流式处理
   io.Copy(file, reader)  // 不会一次性加载到内存
   ```

2. **内存管理**:
   ```go
   // Gin 配置了内存缓存限制（cmd/server/main.go）
   engine.MaxMultipartMemory = 32 << 20 // 32 MB

   // 小文件（< 32MB）：完全在内存中处理
   // 大文件（> 32MB）：超出部分写入临时文件
   ```

3. **配额控制**:
   ```go
   // 支持最大单文件 20GB
   MaxFileSize: 20 * 1024 * 1024 * 1024  // 20GB
   ```

#### ⚠️ 存在的问题

1. **网络超时**:
   - 上传 20GB 文件可能需要很长时间
   - HTTP 超时设置可能不够
   - 需要配置反向代理超时（Nginx/Apache）

2. **失败重试**:
   - 上传失败后必须从头开始
   - 浪费已传输的数据和时间
   - 对于大文件非常不友好

3. **并发控制**:
   - 没有上传队列机制
   - 多用户同时上传大文件会消耗大量资源

4. **网络稳定性**:
   - 网络中断会导致上传失败
   - 没有自动重试机制

---

### 四、文档说明

#### 1. `docs/large-file-upload.md`

**内容摘要**:
- ✅ 说明了支持 20GB 大文件上传
- ✅ 配置了 `MaxMultipartMemory = 32MB`
- ✅ 使用流式上传（`io.Copy`）
- ⚠️ **承认了限制**：

```markdown
### 1. 分片上传

对于超大文件，建议实现分片上传：

优点：
- 支持断点续传        ⚠️ 建议实现但未实现
- 减少单次请求大小
- 提高成功率

实现：
- 前端分片上传        ⚠️ 未实现
- 后端合并分片        ⚠️ 未实现
- 存储分片信息        ⚠️ 未实现
```

#### 2. `docs/Phase13-云盘功能设计总结.md`

**后续工作建议**:
```markdown
### 5. 性能优化
- 缓存文件树结构（Redis）
- 大文件分片上传           ⚠️ 未实现
- 断点续传                 ⚠️ 未实现
- CDN预热
- 缩略图异步生成
```

---

## 📊 功能对比表

| 功能 | 当前状态 | 说明 |
|------|---------|------|
| **上传功能** |
| 单次完整上传 | ✅ 支持 | 使用 `c.FormFile()` 和 `io.Copy` |
| 流式上传 | ✅ 支持 | 不会一次性加载到内存 |
| 分片上传 | ❌ 不支持 | 需要实现 multipart upload |
| 断点续传 | ❌ 不支持 | 需要上传会话管理 |
| 秒传（MD5去重） | ❌ 不支持 | MD5 字段未使用 |
| 上传进度跟踪 | ❌ 不支持 | 没有进度回调机制 |
| 并发上传控制 | ❌ 不支持 | 没有队列机制 |
| **下载功能** |
| 完整下载 | ✅ 支持 | 使用 `io.Copy` |
| Range 请求 | ❌ 不支持 | 不支持 HTTP Range 头 |
| 断点下载 | ❌ 不支持 | 没有 Range 支持 |
| **存储支持** |
| 本地存储 | ✅ 支持 | 完整实现 |
| 阿里云 OSS | ⚠️ 部分支持 | 仅支持普通上传，未使用分片API |
| OSS 分片上传 | ❌ 不支持 | 未使用 SDK 的 multipart API |
| **配额和限制** |
| 单文件大小限制 | ✅ 支持 | 最大 20GB |
| 用户配额限制 | ✅ 支持 | 总空间 10GB（可配置） |
| 网络超时处理 | ⚠️ 需配置 | 需要配置反向代理超时 |
| **数据管理** |
| 文件记录管理 | ✅ 支持 | 完整的 CRUD |
| 上传会话管理 | ❌ 不支持 | 没有会话表 |
| 分片记录管理 | ❌ 不支持 | 没有分片表 |
| MD5 去重 | ❌ 不支持 | 字段存在但未使用 |

---

## 🎯 断点续传实现建议

### 方案一：基于数据库的断点续传（推荐）

#### 1. 数据库设计

```sql
-- 上传会话表
CREATE TABLE `cloud_upload_session` (
  `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `FILE_ID` VARCHAR(64) NOT NULL COMMENT '文件唯一标识（MD5）',
  `USER_ID` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `FILE_NAME` VARCHAR(255) NOT NULL COMMENT '文件名',
  `FILE_SIZE` BIGINT NOT NULL COMMENT '文件总大小',
  `FILE_TYPE` VARCHAR(100) COMMENT '文件类型',
  `CHUNK_SIZE` INT NOT NULL DEFAULT 5242880 COMMENT '分片大小（默认5MB）',
  `TOTAL_CHUNKS` INT NOT NULL COMMENT '总分片数',
  `UPLOADED_CHUNKS` TEXT COMMENT '已上传的分片索引（JSON数组）',
  `STATUS` VARCHAR(20) NOT NULL DEFAULT 'uploading' COMMENT '状态：uploading,paused,completed,failed',
  `STORAGE_PATH` VARCHAR(500) COMMENT '临时存储路径',
  `EXPIRE_TIME` TIMESTAMP NOT NULL COMMENT '过期时间',
  `CREATE_BY` VARCHAR(50),
  `CREATE_TIME` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `UPDATE_BY` VARCHAR(50),
  `UPDATE_TIME` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_file_id` (`FILE_ID`),
  INDEX `idx_user_id` (`USER_ID`),
  INDEX `idx_status` (`STATUS`),
  INDEX `idx_expire_time` (`EXPIRE_TIME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云盘上传会话表';

-- 分片记录表
CREATE TABLE `cloud_chunk_record` (
  `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `SESSION_ID` BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
  `CHUNK_INDEX` INT NOT NULL COMMENT '分片索引',
  `CHUNK_SIZE` INT NOT NULL COMMENT '分片大小',
  `CHUNK_MD5` VARCHAR(32) NOT NULL COMMENT '分片MD5',
  `CHUNK_PATH` VARCHAR(500) COMMENT '分片存储路径',
  `UPLOADED` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已上传',
  `UPLOAD_TIME` TIMESTAMP COMMENT '上传时间',
  `CREATE_TIME` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_session_id` (`SESSION_ID`),
  INDEX `idx_chunk_index` (`CHUNK_INDEX`),
  UNIQUE KEY `uk_session_chunk` (`SESSION_ID`, `CHUNK_INDEX`),
  FOREIGN KEY (`SESSION_ID`) REFERENCES `cloud_upload_session`(`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云盘分片记录表';
```

#### 2. API 设计

```go
// 1. 初始化上传会话
POST /api/v1/cloud/files/multipart/init
Request:
{
    "fileName": "large_file.mp4",
    "fileSize": 5368709120,      // 5GB
    "fileMD5": "abc123...",
    "chunkSize": 5242880,        // 5MB
    "folderId": 1
}
Response:
{
    "sessionId": "uuid-xxxx",
    "fileId": "abc123...",
    "uploadedChunks": [],         // 已上传的分片（断点续传时返回）
    "uploadUrl": "/api/v1/cloud/files/multipart/upload"
}

// 2. 上传单个分片
POST /api/v1/cloud/files/multipart/upload
Request (multipart/form-data):
- sessionId: uuid-xxxx
- chunkIndex: 0
- chunkData: binary
- chunkMD5: xyz789...
Response:
{
    "success": true,
    "chunkIndex": 0,
    "uploaded": true
}

// 3. 查询上传状态（用于断点续传）
GET /api/v1/cloud/files/multipart/status?sessionId=uuid-xxxx
Response:
{
    "sessionId": "uuid-xxxx",
    "status": "uploading",
    "totalChunks": 1024,
    "uploadedChunks": [0, 1, 2, 5, 6, 7],  // 已上传的分片索引
    "progress": 0.68,                       // 上传进度
    "expireTime": "2026-01-16T10:00:00Z"
}

// 4. 完成上传（合并分片）
POST /api/v1/cloud/files/multipart/complete
Request:
{
    "sessionId": "uuid-xxxx",
    "fileMD5": "abc123..."  // 用于验证完整性
}
Response:
{
    "success": true,
    "file": {
        "id": 100,
        "fileName": "large_file.mp4",
        "fileSize": 5368709120,
        "accessUrl": "https://..."
    }
}

// 5. 取消上传
DELETE /api/v1/cloud/files/multipart/{sessionId}
Response:
{
    "success": true,
    "message": "上传已取消，临时文件已清理"
}
```

#### 3. Service 实现

```go
// multipart_upload_service.go

type MultipartUploadService interface {
    // InitUpload 初始化上传会话
    InitUpload(ctx context.Context, req *InitUploadRequest, userID uint) (*UploadSession, error)

    // UploadChunk 上传单个分片
    UploadChunk(ctx context.Context, req *UploadChunkRequest, userID uint) error

    // GetUploadStatus 获取上传状态
    GetUploadStatus(ctx context.Context, sessionID string, userID uint) (*UploadStatus, error)

    // CompleteUpload 完成上传（合并分片）
    CompleteUpload(ctx context.Context, sessionID string, userID uint) (*entity.CloudFile, error)

    // AbortUpload 取消上传
    AbortUpload(ctx context.Context, sessionID string, userID uint) error

    // ResumeUpload 恢复上传（断点续传）
    ResumeUpload(ctx context.Context, fileMD5 string, userID uint) (*UploadSession, error)
}

// InitUpload 初始化上传会话
func (s *service) InitUpload(ctx context.Context, req *InitUploadRequest, userID uint) (*UploadSession, error) {
    // 1. 检查配额
    if err := s.CheckQuota(ctx, userID, req.FileSize); err != nil {
        return nil, err
    }

    // 2. 检查是否存在未完成的会话（断点续传）
    var existingSession entity.UploadSession
    err := s.db.WithContext(ctx).
        Where("FILE_ID = ? AND USER_ID = ? AND STATUS IN (?, ?)",
              req.FileMD5, userID, "uploading", "paused").
        First(&existingSession).Error

    if err == nil {
        // 存在未完成的会话，返回已上传的分片信息
        var uploadedChunks []int
        json.Unmarshal([]byte(existingSession.UploadedChunks), &uploadedChunks)

        return &UploadSession{
            SessionID:      existingSession.ID,
            FileID:         existingSession.FileID,
            TotalChunks:    existingSession.TotalChunks,
            UploadedChunks: uploadedChunks,
        }, nil
    }

    // 3. 创建新的上传会话
    totalChunks := int(math.Ceil(float64(req.FileSize) / float64(req.ChunkSize)))

    session := &entity.UploadSession{
        FileID:         req.FileMD5,
        UserID:         userID,
        FileName:       req.FileName,
        FileSize:       req.FileSize,
        FileType:       req.FileType,
        ChunkSize:      req.ChunkSize,
        TotalChunks:    totalChunks,
        UploadedChunks: "[]",
        Status:         "uploading",
        StoragePath:    fmt.Sprintf("cloud/temp/%d/%s", userID, req.FileMD5),
        ExpireTime:     time.Now().Add(24 * time.Hour), // 24小时过期
    }

    if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
        return nil, errors.Wrap(errors.ErrDatabase, "创建上传会话失败", err)
    }

    return &UploadSession{
        SessionID:      session.ID,
        FileID:         session.FileID,
        TotalChunks:    totalChunks,
        UploadedChunks: []int{},
    }, nil
}

// UploadChunk 上传单个分片
func (s *service) UploadChunk(ctx context.Context, req *UploadChunkRequest, userID uint) error {
    // 1. 获取上传会话
    var session entity.UploadSession
    if err := s.db.WithContext(ctx).
        Where("ID = ? AND USER_ID = ?", req.SessionID, userID).
        First(&session).Error; err != nil {
        return errors.New(errors.ErrResourceNotFound, "上传会话不存在")
    }

    // 2. 检查会话状态
    if session.Status != "uploading" && session.Status != "paused" {
        return errors.New(errors.ErrInvalidParam, "上传会话状态无效")
    }

    // 3. 检查会话是否过期
    if session.ExpireTime.Before(time.Now()) {
        return errors.New(errors.ErrInvalidParam, "上传会话已过期")
    }

    // 4. 检查分片是否已上传
    var chunkRecord entity.ChunkRecord
    err := s.db.WithContext(ctx).
        Where("SESSION_ID = ? AND CHUNK_INDEX = ?", session.ID, req.ChunkIndex).
        First(&chunkRecord).Error

    if err == nil && chunkRecord.Uploaded {
        return nil // 分片已上传，跳过
    }

    // 5. 验证分片MD5
    actualMD5 := calculateMD5(req.ChunkData)
    if actualMD5 != req.ChunkMD5 {
        return errors.New(errors.ErrInvalidParam, "分片MD5校验失败")
    }

    // 6. 保存分片到临时目录
    chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, req.ChunkIndex)
    if err := s.storage.Upload(ctx, chunkPath, bytes.NewReader(req.ChunkData), "application/octet-stream"); err != nil {
        return errors.Wrap(errors.ErrInternal, "保存分片失败", err)
    }

    // 7. 更新分片记录
    chunkRecord = entity.ChunkRecord{
        SessionID:  session.ID,
        ChunkIndex: req.ChunkIndex,
        ChunkSize:  len(req.ChunkData),
        ChunkMD5:   req.ChunkMD5,
        ChunkPath:  chunkPath,
        Uploaded:   true,
        UploadTime: time.Now(),
    }

    if err := s.db.WithContext(ctx).Create(&chunkRecord).Error; err != nil {
        return errors.Wrap(errors.ErrDatabase, "创建分片记录失败", err)
    }

    // 8. 更新上传会话的已上传分片列表
    var uploadedChunks []int
    json.Unmarshal([]byte(session.UploadedChunks), &uploadedChunks)
    uploadedChunks = append(uploadedChunks, req.ChunkIndex)
    sort.Ints(uploadedChunks)

    uploadedChunksJSON, _ := json.Marshal(uploadedChunks)
    s.db.WithContext(ctx).Model(&entity.UploadSession{}).
        Where("ID = ?", session.ID).
        Update("UPLOADED_CHUNKS", string(uploadedChunksJSON))

    return nil
}

// CompleteUpload 完成上传（合并分片）
func (s *service) CompleteUpload(ctx context.Context, sessionID string, userID uint) (*entity.CloudFile, error) {
    // 1. 获取上传会话
    var session entity.UploadSession
    if err := s.db.WithContext(ctx).
        Where("ID = ? AND USER_ID = ?", sessionID, userID).
        First(&session).Error; err != nil {
        return nil, errors.New(errors.ErrResourceNotFound, "上传会话不存在")
    }

    // 2. 检查所有分片是否已上传
    var uploadedChunks []int
    json.Unmarshal([]byte(session.UploadedChunks), &uploadedChunks)

    if len(uploadedChunks) != session.TotalChunks {
        return nil, errors.New(errors.ErrInvalidParam,
            fmt.Sprintf("分片未完全上传：已上传 %d/%d", len(uploadedChunks), session.TotalChunks))
    }

    // 3. 合并分片
    finalPath := fmt.Sprintf("cloud/%d/%s/%s%s",
        userID,
        time.Now().Format("2006/01/02"),
        uuid.New().String(),
        filepath.Ext(session.FileName))

    // 创建最终文件
    finalFile, err := os.Create(filepath.Join(s.basePath, finalPath))
    if err != nil {
        return nil, errors.Wrap(errors.ErrInternal, "创建最终文件失败", err)
    }
    defer finalFile.Close()

    // 按顺序合并所有分片
    for i := 0; i < session.TotalChunks; i++ {
        chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, i)

        // 读取分片
        chunkReader, err := s.storage.Download(ctx, chunkPath)
        if err != nil {
            return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("读取分片 %d 失败", i), err)
        }

        // 写入最终文件
        if _, err := io.Copy(finalFile, chunkReader); err != nil {
            chunkReader.Close()
            return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("合并分片 %d 失败", i), err)
        }
        chunkReader.Close()
    }

    // 4. 验证文件完整性（MD5）
    finalFile.Seek(0, io.SeekStart)
    actualMD5 := calculateFileMD5(finalFile)
    if actualMD5 != session.FileID {
        os.Remove(filepath.Join(s.basePath, finalPath))
        return nil, errors.New(errors.ErrInvalidParam, "文件MD5校验失败")
    }

    // 5. 创建文件记录
    file := &entity.CloudFile{
        FileName:    session.FileName,
        StoragePath: finalPath,
        FileSize:    session.FileSize,
        FileType:    session.FileType,
        MD5:         session.FileID,
        OwnerID:     userID,
        // ...
    }

    if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
        return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
    }

    // 6. 更新配额
    s.UpdateQuota(ctx, userID, session.FileSize, 1)

    // 7. 更新会话状态为已完成
    s.db.WithContext(ctx).Model(&entity.UploadSession{}).
        Where("ID = ?", session.ID).
        Update("STATUS", "completed")

    // 8. 异步清理临时文件
    go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

    return file, nil
}

// cleanupChunks 清理临时分片文件
func (s *service) cleanupChunks(ctx context.Context, sessionID uint, storagePath string) {
    // 1. 删除所有分片文件
    objects, _ := s.storage.ListObjects(ctx, storagePath, 0)
    for _, obj := range objects {
        s.storage.Delete(ctx, obj.Key)
    }

    // 2. 删除分片记录
    s.db.WithContext(ctx).
        Where("SESSION_ID = ?", sessionID).
        Delete(&entity.ChunkRecord{})
}
```

#### 4. 前端实现示例

```javascript
// 文件上传类
class ChunkedFileUploader {
    constructor(file, chunkSize = 5 * 1024 * 1024) {
        this.file = file;
        this.chunkSize = chunkSize;
        this.totalChunks = Math.ceil(file.size / chunkSize);
        this.uploadedChunks = [];
        this.sessionId = null;
        this.paused = false;
        this.onProgress = null;
        this.onComplete = null;
        this.onError = null;
    }

    // 初始化上传
    async init() {
        // 计算文件MD5
        const fileMD5 = await this.calculateMD5();

        // 初始化上传会话
        const response = await fetch('/api/v1/cloud/files/multipart/init', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                fileName: this.file.name,
                fileSize: this.file.size,
                fileMD5: fileMD5,
                chunkSize: this.chunkSize,
                folderId: this.folderId
            })
        });

        const data = await response.json();
        this.sessionId = data.sessionId;
        this.uploadedChunks = data.uploadedChunks || [];

        console.log(`初始化成功，会话ID: ${this.sessionId}`);
        console.log(`已上传分片: ${this.uploadedChunks.length}/${this.totalChunks}`);

        return this.sessionId;
    }

    // 开始上传
    async start() {
        if (!this.sessionId) {
            await this.init();
        }

        // 上传所有未完成的分片
        for (let i = 0; i < this.totalChunks; i++) {
            // 检查是否已上传
            if (this.uploadedChunks.includes(i)) {
                console.log(`分片 ${i} 已上传，跳过`);
                continue;
            }

            // 检查是否暂停
            if (this.paused) {
                console.log('上传已暂停');
                return;
            }

            try {
                await this.uploadChunk(i);
                this.uploadedChunks.push(i);

                // 触发进度回调
                const progress = this.uploadedChunks.length / this.totalChunks;
                if (this.onProgress) {
                    this.onProgress(progress, this.uploadedChunks.length, this.totalChunks);
                }
            } catch (error) {
                console.error(`上传分片 ${i} 失败:`, error);

                // 触发错误回调
                if (this.onError) {
                    this.onError(error, i);
                }

                // 可以选择重试
                // await this.uploadChunk(i);
                return;
            }
        }

        // 所有分片上传完成，合并文件
        await this.complete();
    }

    // 上传单个分片
    async uploadChunk(chunkIndex) {
        const start = chunkIndex * this.chunkSize;
        const end = Math.min(start + this.chunkSize, this.file.size);
        const chunk = this.file.slice(start, end);

        // 计算分片MD5
        const chunkMD5 = await this.calculateChunkMD5(chunk);

        // 构造FormData
        const formData = new FormData();
        formData.append('sessionId', this.sessionId);
        formData.append('chunkIndex', chunkIndex);
        formData.append('chunkData', chunk);
        formData.append('chunkMD5', chunkMD5);

        // 上传分片
        const response = await fetch('/api/v1/cloud/files/multipart/upload', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        });

        if (!response.ok) {
            throw new Error(`上传分片失败: ${response.statusText}`);
        }

        console.log(`分片 ${chunkIndex} 上传成功`);
    }

    // 完成上传（合并分片）
    async complete() {
        const response = await fetch('/api/v1/cloud/files/multipart/complete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                sessionId: this.sessionId
            })
        });

        const data = await response.json();

        console.log('上传完成:', data);

        // 触发完成回调
        if (this.onComplete) {
            this.onComplete(data.file);
        }

        return data.file;
    }

    // 暂停上传
    pause() {
        this.paused = true;
        console.log('上传已暂停');
    }

    // 恢复上传（断点续传）
    async resume() {
        // 获取上传状态
        const response = await fetch(`/api/v1/cloud/files/multipart/status?sessionId=${this.sessionId}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();
        this.uploadedChunks = data.uploadedChunks;
        this.paused = false;

        console.log(`恢复上传，已完成: ${this.uploadedChunks.length}/${this.totalChunks}`);

        // 继续上传
        await this.start();
    }

    // 取消上传
    async cancel() {
        const response = await fetch(`/api/v1/cloud/files/multipart/${this.sessionId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        console.log('上传已取消');
    }

    // 计算文件MD5
    async calculateMD5() {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => {
                const spark = new SparkMD5.ArrayBuffer();
                spark.append(e.target.result);
                resolve(spark.end());
            };
            reader.onerror = reject;
            reader.readAsArrayBuffer(this.file);
        });
    }

    // 计算分片MD5
    async calculateChunkMD5(chunk) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => {
                const spark = new SparkMD5.ArrayBuffer();
                spark.append(e.target.result);
                resolve(spark.end());
            };
            reader.onerror = reject;
            reader.readAsArrayBuffer(chunk);
        });
    }
}

// 使用示例
async function uploadFile(file) {
    const uploader = new ChunkedFileUploader(file);

    // 设置进度回调
    uploader.onProgress = (progress, uploaded, total) => {
        console.log(`上传进度: ${(progress * 100).toFixed(2)}%`);
        console.log(`已上传分片: ${uploaded}/${total}`);

        // 更新UI
        updateProgressBar(progress);
    };

    // 设置完成回调
    uploader.onComplete = (file) => {
        console.log('文件上传完成:', file);
        alert('上传成功！');
    };

    // 设置错误回调
    uploader.onError = (error, chunkIndex) => {
        console.error(`分片 ${chunkIndex} 上传失败:`, error);

        // 可以选择暂停上传，等待用户重试
        uploader.pause();

        if (confirm('上传失败，是否重试？')) {
            uploader.resume(); // 断点续传
        }
    };

    // 开始上传
    await uploader.start();
}

// 页面刷新时保存上传状态
window.addEventListener('beforeunload', (e) => {
    if (uploader && !uploader.paused && uploader.uploadedChunks.length < uploader.totalChunks) {
        // 保存会话ID到 localStorage
        localStorage.setItem('uploadSessionId', uploader.sessionId);

        e.preventDefault();
        e.returnValue = '文件正在上传，确定要离开吗？';
    }
});

// 页面加载时恢复上传
window.addEventListener('load', async () => {
    const sessionId = localStorage.getItem('uploadSessionId');
    if (sessionId) {
        if (confirm('检测到未完成的上传，是否继续？')) {
            const uploader = new ChunkedFileUploader(file);
            uploader.sessionId = sessionId;
            await uploader.resume(); // 断点续传

            localStorage.removeItem('uploadSessionId');
        }
    }
});
```

---

### 方案二：基于 OSS 的断点续传（适合云存储）

如果使用阿里云 OSS，可以利用 OSS SDK 的分片上传功能：

```go
// OSS 分片上传实现
func (s *AliyunOSS) MultipartUpload(ctx context.Context, path string, reader io.Reader, fileSize int64) (string, error) {
    // 1. 初始化分片上传
    imur, err := s.bucket.InitiateMultipartUpload(path)
    if err != nil {
        return "", errors.Wrap(errors.ErrInternal, "初始化分片上传失败", err)
    }

    // 2. 分片上传
    chunkSize := int64(5 * 1024 * 1024) // 5MB per chunk
    var parts []oss.UploadPart
    buffer := make([]byte, chunkSize)

    for partNum := 1; ; partNum++ {
        // 读取分片
        n, err := io.ReadFull(reader, buffer)

        if n > 0 {
            // 上传分片
            part, uploadErr := s.bucket.UploadPart(
                imur,
                bytes.NewReader(buffer[:n]),
                int64(n),
                partNum,
            )

            if uploadErr != nil {
                // 上传失败，取消分片上传
                s.bucket.AbortMultipartUpload(imur)
                return "", errors.Wrap(errors.ErrInternal, "上传分片失败", uploadErr)
            }

            parts = append(parts, part)
        }

        // 检查是否读取完毕
        if err == io.EOF || err == io.ErrUnexpectedEOF {
            break
        }

        if err != nil {
            s.bucket.AbortMultipartUpload(imur)
            return "", errors.Wrap(errors.ErrInternal, "读取文件失败", err)
        }
    }

    // 3. 完成分片上传
    _, err = s.bucket.CompleteMultipartUpload(imur, parts)
    if err != nil {
        return "", errors.Wrap(errors.ErrInternal, "完成分片上传失败", err)
    }

    // 4. 返回文件URL
    return s.GetURL(ctx, path, 0)
}

// 列出未完成的分片上传任务（用于恢复上传）
func (s *AliyunOSS) ListMultipartUploads(ctx context.Context, prefix string) ([]oss.UncompletedUpload, error) {
    lmu, err := s.bucket.ListMultipartUploads(oss.Prefix(prefix))
    if err != nil {
        return nil, errors.Wrap(errors.ErrInternal, "列举未完成的分片上传失败", err)
    }
    return lmu.Uploads, nil
}

// 列出已上传的分片（用于断点续传）
func (s *AliyunOSS) ListUploadedParts(ctx context.Context, uploadID string, objectKey string) ([]oss.UploadPart, error) {
    imur := oss.InitiateMultipartUploadResult{
        UploadID: uploadID,
        Key:      objectKey,
        Bucket:   s.bucketName,
    }

    lp, err := s.bucket.ListUploadedParts(imur)
    if err != nil {
        return nil, errors.Wrap(errors.ErrInternal, "列举已上传的分片失败", err)
    }

    return lp.UploadedParts, nil
}

// 恢复分片上传（断点续传）
func (s *AliyunOSS) ResumeMultipartUpload(ctx context.Context, uploadID string, objectKey string, reader io.Reader, fileSize int64) (string, error) {
    // 1. 获取已上传的分片
    uploadedParts, err := s.ListUploadedParts(ctx, uploadID, objectKey)
    if err != nil {
        return "", err
    }

    // 2. 计算已上传的字节数
    uploadedSize := int64(0)
    for _, part := range uploadedParts {
        uploadedSize += part.Size
    }

    // 3. 跳过已上传的部分
    if seeker, ok := reader.(io.Seeker); ok {
        _, err := seeker.Seek(uploadedSize, io.SeekStart)
        if err != nil {
            return "", errors.Wrap(errors.ErrInternal, "定位文件失败", err)
        }
    }

    // 4. 继续上传剩余分片
    imur := oss.InitiateMultipartUploadResult{
        UploadID: uploadID,
        Key:      objectKey,
        Bucket:   s.bucketName,
    }

    chunkSize := int64(5 * 1024 * 1024)
    buffer := make([]byte, chunkSize)
    partNum := len(uploadedParts) + 1

    for {
        n, err := io.ReadFull(reader, buffer)

        if n > 0 {
            part, uploadErr := s.bucket.UploadPart(
                imur,
                bytes.NewReader(buffer[:n]),
                int64(n),
                partNum,
            )

            if uploadErr != nil {
                return "", errors.Wrap(errors.ErrInternal, "上传分片失败", uploadErr)
            }

            uploadedParts = append(uploadedParts, part)
            partNum++
        }

        if err == io.EOF || err == io.ErrUnexpectedEOF {
            break
        }

        if err != nil {
            return "", errors.Wrap(errors.ErrInternal, "读取文件失败", err)
        }
    }

    // 5. 完成分片上传
    _, err = s.bucket.CompleteMultipartUpload(imur, uploadedParts)
    if err != nil {
        return "", errors.Wrap(errors.ErrInternal, "完成分片上传失败", err)
    }

    return s.GetURL(ctx, objectKey, 0)
}
```

---

## 📝 总结

### 当前状态

| 项目 | 状态 |
|------|------|
| **上传** | ✅ 支持单次完整上传（最大20GB） |
| | ✅ 支持流式处理（不会一次性加载到内存） |
| | ❌ 不支持分片上传 |
| | ❌ 不支持断点续传 |
| | ❌ 不支持上传进度跟踪 |
| **下载** | ✅ 支持完整下载 |
| | ❌ 不支持 HTTP Range 请求 |
| | ❌ 不支持断点下载 |
| **存储** | ✅ 本地存储完整实现 |
| | ⚠️ OSS 仅使用普通上传，未使用分片API |

### 实现建议

**优先级排序**:

1. **高优先级**（推荐先实现）:
   - ✅ 分片上传基础功能
   - ✅ 断点续传（基于会话管理）
   - ✅ 上传进度跟踪
   - ✅ MD5 去重（秒传）

2. **中优先级**:
   - ✅ HTTP Range 请求支持
   - ✅ OSS 分片上传集成
   - ✅ 并发上传控制

3. **低优先级**:
   - ✅ 自动重试机制
   - ✅ 上传队列管理
   - ✅ 过期会话清理

### 工作量评估

| 任务 | 工作量 | 说明 |
|------|--------|------|
| 数据库设计 | 0.5天 | 2个新表 |
| 后端API实现 | 2-3天 | 5个新接口 + Service |
| 存储层改造 | 1-2天 | 支持分片上传 |
| 前端实现 | 2-3天 | 分片上传 + 断点续传UI |
| 测试调试 | 1-2天 | 功能测试 + 边界测试 |
| **总计** | **7-11天** | 1-2周 |

---

## 🔗 相关文档

- `docs/large-file-upload.md` - 大文件上传说明
- `docs/Phase13-云盘功能设计总结.md` - 云盘功能总结
- `internal/service/cloud/cloud_service.go` - 云盘服务实现
- `internal/pkg/storage/` - 存储层实现
- 阿里云OSS SDK文档: https://help.aliyun.com/document_detail/32144.html

---

**分析完成时间**: 2026-01-15
**分析人员**: Claude Code Assistant
**结论**: ❌ 当前不支持断点续传，但提供了完整的实现方案
