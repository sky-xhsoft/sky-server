# 分片上传和断点续传功能实现总结

## 📋 实施概览

**实施时间**: 2026-01-15
**功能**: 云盘分片上传和断点续传
**状态**: ✅ 实现完成

---

## 🎯 实现的功能

### 核心功能

- ✅ **分片上传**: 支持将大文件分片上传，默认 5MB/片
- ✅ **断点续传**: 上传中断后可从断点继续，无需重新上传
- ✅ **秒传**: 基于 MD5 检测重复文件，相同文件无需上传
- ✅ **进度跟踪**: 实时查询上传进度和已上传分片列表
- ✅ **会话管理**: 24小时有效期，过期自动清理
- ✅ **MD5校验**: 分片级别和文件级别的完整性校验
- ✅ **自动清理**: 完成或取消后自动清理临时文件

---

## 📁 新增文件清单

### 1. 数据库文件
- `sqls/cloud_multipart_upload.sql` - 数据库表结构（2个新表）

### 2. 实体模型
- `internal/model/entity/cloud_disk.go` - 添加了 2 个新实体
  - `CloudUploadSession` - 上传会话表
  - `CloudChunkRecord` - 分片记录表

### 3. Service 层
- `internal/service/cloud/multipart_upload_service.go` - 分片上传服务（~700行）

### 4. Handler 层
- `api/handler/multipart_upload_handler.go` - 分片上传处理器（~200行）

### 5. 文档
- `docs/multipart-upload-implementation-summary.md` - 实现总结（本文档）
- `docs/multipart-upload-api.md` - API文档
- `docs/multipart-upload-frontend-guide.md` - 前端集成指南

---

## 🗄️ 数据库设计

### 表1: cloud_upload_session（上传会话表）

```sql
CREATE TABLE `cloud_upload_session` (
  `ID` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `FILE_ID` VARCHAR(64) NOT NULL,        -- 文件MD5
  `USER_ID` BIGINT UNSIGNED NOT NULL,    -- 用户ID
  `FILE_NAME` VARCHAR(255) NOT NULL,     -- 文件名
  `FILE_SIZE` BIGINT NOT NULL,           -- 文件大小
  `FILE_TYPE` VARCHAR(100),              -- MIME类型
  `FOLDER_ID` BIGINT UNSIGNED,           -- 目标文件夹
  `CHUNK_SIZE` INT DEFAULT 5242880,      -- 分片大小（5MB）
  `TOTAL_CHUNKS` INT NOT NULL,           -- 总分片数
  `UPLOADED_CHUNKS` TEXT,                -- 已上传分片（JSON数组）
  `STATUS` VARCHAR(20) DEFAULT 'uploading', -- uploading/paused/completed/failed
  `STORAGE_TYPE` VARCHAR(20) DEFAULT 'local',
  `STORAGE_PATH` VARCHAR(500),           -- 临时存储路径
  `EXPIRE_TIME` TIMESTAMP NOT NULL,      -- 过期时间
  -- 标准字段...
  INDEX `idx_file_id` (`FILE_ID`),
  INDEX `idx_user_id` (`USER_ID`),
  INDEX `idx_status` (`STATUS`)
);
```

### 表2: cloud_chunk_record（分片记录表）

```sql
CREATE TABLE `cloud_chunk_record` (
  `ID` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `SESSION_ID` BIGINT UNSIGNED NOT NULL, -- 会话ID
  `CHUNK_INDEX` INT NOT NULL,            -- 分片索引
  `CHUNK_SIZE` INT NOT NULL,             -- 分片大小
  `CHUNK_MD5` VARCHAR(32) NOT NULL,      -- 分片MD5
  `CHUNK_PATH` VARCHAR(500),             -- 分片路径
  `UPLOADED` TINYINT(1) DEFAULT 0,       -- 是否已上传
  `UPLOAD_TIME` TIMESTAMP NULL,          -- 上传时间
  `RETRY_COUNT` INT DEFAULT 0,           -- 重试次数
  INDEX `idx_session_id` (`SESSION_ID`),
  UNIQUE KEY `uk_session_chunk` (`SESSION_ID`, `CHUNK_INDEX`)
);
```

---

## 🔌 API 设计

### 1. 初始化上传

```http
POST /api/v1/cloud/files/multipart/init
Content-Type: application/json

{
  "fileName": "video.mp4",
  "fileSize": 104857600,
  "fileMd5": "abc123...",
  "fileType": "video/mp4",
  "chunkSize": 5242880,
  "folderId": 1,
  "storageType": "local"
}

Response 200:
{
  "code": 200,
  "data": {
    "sessionId": 123,
    "fileId": "abc123...",
    "fileName": "video.mp4",
    "fileSize": 104857600,
    "chunkSize": 5242880,
    "totalChunks": 20,
    "uploadedChunks": [],  // 断点续传时返回已上传的分片
    "status": "uploading",
    "expireTime": "2026-01-16T10:00:00Z"
  }
}
```

### 2. 上传分片

```http
POST /api/v1/cloud/files/multipart/upload
Content-Type: multipart/form-data

sessionId=123
chunkIndex=0
chunkMd5=xyz789...
chunkData=<binary>

Response 200:
{
  "code": 200,
  "data": {
    "message": "分片上传成功",
    "chunkIndex": 0,
    "uploaded": true
  }
}
```

### 3. 查询状态

```http
GET /api/v1/cloud/files/multipart/status?sessionId=123

Response 200:
{
  "code": 200,
  "data": {
    "sessionId": 123,
    "fileId": "abc123...",
    "fileName": "video.mp4",
    "fileSize": 104857600,
    "totalChunks": 20,
    "uploadedChunks": [0, 1, 2, 5, 6],
    "status": "uploading",
    "progress": 0.25,
    "expireTime": "2026-01-16T10:00:00Z"
  }
}
```

### 4. 完成上传

```http
POST /api/v1/cloud/files/multipart/complete
Content-Type: application/json

{
  "sessionId": 123
}

Response 200:
{
  "code": 200,
  "data": {
    "id": 100,
    "fileName": "video.mp4",
    "fileSize": 104857600,
    "md5": "abc123...",
    "accessUrl": "https://...",
    ...
  }
}
```

### 5. 取消上传

```http
DELETE /api/v1/cloud/files/multipart/123

Response 200:
{
  "code": 200,
  "data": {
    "message": "取消上传成功"
  }
}
```

### 6. 恢复上传（断点续传）

```http
POST /api/v1/cloud/files/multipart/resume
Content-Type: application/json

{
  "fileMd5": "abc123..."
}

Response 200:
{
  "code": 200,
  "data": {
    "sessionId": 123,
    "fileId": "abc123...",
    "uploadedChunks": [0, 1, 2, 5, 6],  // 已上传的分片
    ...
  }
}
```

---

## 🔄 工作流程

### 正常上传流程

```
1. [前端] 计算文件MD5
2. [前端] 调用 /init 初始化上传会话
3. [后端] 返回 sessionId 和分片信息
4. [前端] 循环上传每个分片到 /upload
5. [后端] 验证MD5，保存分片，更新进度
6. [前端] 所有分片上传完成后调用 /complete
7. [后端] 合并分片，验证文件MD5，创建文件记录
8. [后端] 异步清理临时文件
9. [前端] 显示上传成功
```

### 断点续传流程

```
1. [前端] 上传中断（网络断开、页面刷新、浏览器关闭）
2. [前端] 保存 sessionId 到 localStorage
3. [前端] 重新打开页面，检测到未完成的上传
4. [前端] 调用 /status 查询已上传的分片
5. [后端] 返回已上传分片列表 [0,1,2,5,6]
6. [前端] 跳过已上传的分片，继续上传剩余分片
7. [前端] 完成后调用 /complete
```

### 秒传流程

```
1. [前端] 计算文件MD5
2. [前端] 调用 /init 初始化
3. [后端] 检查是否存在相同MD5的会话
4. [后端] 如果存在，返回已有会话信息
5. [前端] 上传缺失的分片
6. [前端] 调用 /complete
7. [后端] 检查是否存在相同MD5的文件
8. [后端] 如果存在，直接复制文件记录（秒传）
9. [后端] 无需实际合并分片
```

---

## 🎨 前端集成示例

### JavaScript 分片上传类

```javascript
class ChunkedFileUploader {
    constructor(file, options = {}) {
        this.file = file;
        this.chunkSize = options.chunkSize || 5 * 1024 * 1024; // 5MB
        this.totalChunks = Math.ceil(file.size / this.chunkSize);
        this.sessionId = null;
        this.uploadedChunks = [];
        this.fileMD5 = null;

        // 回调函数
        this.onProgress = options.onProgress;
        this.onComplete = options.onComplete;
        this.onError = options.onError;
    }

    // 计算文件MD5
    async calculateMD5() {
        // 使用 spark-md5 库
        return new Promise((resolve, reject) => {
            const blobSlice = File.prototype.slice;
            const chunkSize = 2097152; // 2MB
            const chunks = Math.ceil(this.file.size / chunkSize);
            let currentChunk = 0;
            const spark = new SparkMD5.ArrayBuffer();
            const fileReader = new FileReader();

            fileReader.onload = (e) => {
                spark.append(e.target.result);
                currentChunk++;

                if (currentChunk < chunks) {
                    loadNext();
                } else {
                    resolve(spark.end());
                }
            };

            fileReader.onerror = reject;

            function loadNext() {
                const start = currentChunk * chunkSize;
                const end = Math.min(start + chunkSize, this.file.size);
                fileReader.readAsArrayBuffer(blobSlice.call(this.file, start, end));
            }

            loadNext.call(this);
        });
    }

    // 初始化上传
    async init() {
        this.fileMD5 = await this.calculateMD5();

        const response = await fetch('/api/v1/cloud/files/multipart/init', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                fileName: this.file.name,
                fileSize: this.file.size,
                fileMd5: this.fileMD5,
                fileType: this.file.type,
                chunkSize: this.chunkSize,
                storageType: 'local'
            })
        });

        const data = await response.json();
        this.sessionId = data.data.sessionId;
        this.uploadedChunks = data.data.uploadedChunks || [];

        return this.sessionId;
    }

    // 上传单个分片
    async uploadChunk(chunkIndex) {
        const start = chunkIndex * this.chunkSize;
        const end = Math.min(start + this.chunkSize, this.file.size);
        const chunk = this.file.slice(start, end);

        // 计算分片MD5
        const chunkMD5 = await this.calculateChunkMD5(chunk);

        const formData = new FormData();
        formData.append('sessionId', this.sessionId);
        formData.append('chunkIndex', chunkIndex);
        formData.append('chunkMd5', chunkMD5);
        formData.append('chunkData', chunk);

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
    }

    // 开始上传
    async start() {
        if (!this.sessionId) {
            await this.init();
        }

        for (let i = 0; i < this.totalChunks; i++) {
            // 跳过已上传的分片（断点续传）
            if (this.uploadedChunks.includes(i)) {
                continue;
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
                if (this.onError) {
                    this.onError(error, i);
                }
                throw error;
            }
        }

        // 完成上传
        await this.complete();
    }

    // 完成上传
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

        if (this.onComplete) {
            this.onComplete(data.data);
        }

        return data.data;
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
const uploader = new ChunkedFileUploader(file, {
    onProgress: (progress, uploaded, total) => {
        console.log(`进度: ${(progress * 100).toFixed(2)}%`);
        updateProgressBar(progress);
    },
    onComplete: (file) => {
        console.log('上传完成', file);
        alert('上传成功！');
    },
    onError: (error, chunkIndex) => {
        console.error(`分片 ${chunkIndex} 上传失败`, error);
    }
});

uploader.start();
```

---

## ⚙️ 集成步骤

### 0. 配置参数

在 `configs/config.yaml` 中配置分片上传参数：

```yaml
# 分片上传配置
multipartUpload:
  chunkSize: 5242880  # 默认分片大小：5MB（字节）
  sessionExpireHours: 24  # 上传会话过期时间（小时）
  cleanupInterval: 3600  # 清理任务执行间隔（秒）
```

在 `internal/config/config.go` 中添加配置结构：

```go
// MultipartUploadConfig 分片上传配置
type MultipartUploadConfig struct {
    ChunkSize          int `mapstructure:"chunkSize"`          // 默认分片大小（字节）
    SessionExpireHours int `mapstructure:"sessionExpireHours"` // 会话过期时间（小时）
    CleanupInterval    int `mapstructure:"cleanupInterval"`    // 清理任务执行间隔（秒）
}
```

### 1. 运行数据库迁移

```bash
mysql -u root -p sky_server < sqls/cloud_multipart_upload.sql
```

### 2. 更新路由配置

在 `api/router/router.go` 或 `cmd/server/main.go` 中注册路由：

```go
// 创建服务（传递配置参数）
multipartService := cloud.NewMultipartUploadService(
    db,
    storageInstance,
    cloudService,
    cfg.MultipartUpload.ChunkSize,          // 默认分片大小
    cfg.MultipartUpload.SessionExpireHours, // 会话过期时间
)
multipartHandler := handler.NewMultipartUploadHandler(multipartService)

// 注册路由
cloudGroup := r.Group("/api/v1/cloud/files/multipart")
cloudGroup.Use(authMiddleware) // 添加认证中间件
{
    cloudGroup.POST("/init", multipartHandler.InitUpload)
    cloudGroup.POST("/upload", multipartHandler.UploadChunk)
    cloudGroup.GET("/status", multipartHandler.GetUploadStatus)
    cloudGroup.POST("/complete", multipartHandler.CompleteUpload)
    cloudGroup.DELETE("/:sessionId", multipartHandler.AbortUpload)
    cloudGroup.POST("/resume", multipartHandler.ResumeUpload)
}
```

### 3. 定时清理任务（已自动集成）

定时清理任务已在 `main.go` 中自动集成，无需手动添加：

```go
// 10. 启动定时清理任务
go func() {
    ticker := time.NewTicker(time.Duration(cfg.MultipartUpload.CleanupInterval) * time.Second)
    defer ticker.Stop()

    logger.Info("分片上传清理任务已启动",
        zap.Int("intervalSeconds", cfg.MultipartUpload.CleanupInterval))

    for range ticker.C {
        ctx := context.Background()
        if err := multipartService.CleanupExpiredSessions(ctx); err != nil {
            logger.Error("清理过期会话失败", zap.Error(err))
        } else {
            logger.Info("过期会话清理完成")
        }
    }
}()
```

清理任务会根据配置文件中的 `multipartUpload.cleanupInterval` 参数定期执行（默认每小时）。

---

## 🧪 测试

### 单元测试

```bash
go test ./internal/service/cloud/... -v
```

### 手动测试

#### 1. 准备测试文件

```bash
# 创建 100MB 测试文件
dd if=/dev/zero of=test_100mb.bin bs=1M count=100
```

#### 2. 初始化上传

```bash
curl -X POST http://localhost:9090/api/v1/cloud/files/multipart/init \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "fileName": "test_100mb.bin",
    "fileSize": 104857600,
    "fileMd5": "FILE_MD5_HERE",
    "chunkSize": 5242880
  }'
```

#### 3. 上传分片

```bash
curl -X POST http://localhost:9090/api/v1/cloud/files/multipart/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "sessionId=1" \
  -F "chunkIndex=0" \
  -F "chunkMd5=CHUNK_MD5" \
  -F "chunkData=@chunk_0.bin"
```

#### 4. 完成上传

```bash
curl -X POST http://localhost:9090/api/v1/cloud/files/multipart/complete \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sessionId": 1}'
```

---

## 📊 性能优化

### 1. 分片大小选择

| 文件大小 | 推荐分片大小 | 说明 |
|---------|------------|------|
| < 100MB | 2-5MB | 减少请求次数 |
| 100MB-1GB | 5-10MB | 平衡性能和可靠性 |
| > 1GB | 10-20MB | 提高传输效率 |

### 2. 并发上传

```javascript
// 同时上传多个分片
async startParallel(concurrency = 3) {
    const chunks = Array.from({length: this.totalChunks}, (_, i) => i)
        .filter(i => !this.uploadedChunks.includes(i));

    for (let i = 0; i < chunks.length; i += concurrency) {
        const batch = chunks.slice(i, i + concurrency);
        await Promise.all(batch.map(index => this.uploadChunk(index)));
    }

    await this.complete();
}
```

### 3. 数据库索引

确保以下字段有索引：
- `cloud_upload_session.FILE_ID`
- `cloud_upload_session.USER_ID`
- `cloud_upload_session.STATUS`
- `cloud_chunk_record.SESSION_ID`

---

## 🔒 安全建议

1. **认证授权**: 所有API都需要认证
2. **配额检查**: 初始化时检查用户配额
3. **MD5校验**: 分片和文件级别都要校验
4. **过期清理**: 定期清理过期会话和临时文件
5. **并发限制**: 限制单用户并发上传数
6. **文件大小限制**: 设置合理的单文件大小上限

---

## 📝 常见问题

### Q1: 上传中断后如何恢复？

A: 前端保存 `sessionId`，重新打开页面时调用 `/status` 获取已上传分片，然后继续上传未完成的分片。

### Q2: 如何实现秒传？

A: 系统会在 `CompleteUpload` 时检查是否存在相同 MD5 的文件，如果存在则直接复制文件记录。

### Q3: 过期会话如何清理？

A: 可以通过定时任务调用 `CleanupExpiredSessions` 清理，或使用数据库存储过程。

### Q4: 支持哪些存储类型？

A: 目前支持本地存储（local）和阿里云OSS（oss）。

### Q5: 分片大小可以自定义吗？

A: 可以，在初始化时指定 `chunkSize`，默认 5MB。

---

## 🎉 总结

分片上传和断点续传功能已完整实现：

- ✅ 完整的分片上传流程
- ✅ 可靠的断点续传机制
- ✅ 智能的秒传功能
- ✅ 完善的MD5校验
- ✅ 自动清理机制
- ✅ 详细的文档和示例

可以支持 **任意大小** 的文件上传，有效解决了大文件上传的痛点！

---

**实现完成时间**: 2026-01-15
**总代码量**: ~1500行
**新增表**: 2个
**新增API**: 6个
