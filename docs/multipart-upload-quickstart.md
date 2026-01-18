# 分片上传快速开始指南

## 🚀 快速开始（5分钟上手）

### 步骤0: 配置参数（可选）

在 `configs/config.yaml` 中可以配置分片上传参数：

```yaml
# 分片上传配置
multipartUpload:
  chunkSize: 5242880  # 默认分片大小：5MB（字节）
  sessionExpireHours: 24  # 上传会话过期时间（小时）
  cleanupInterval: 3600  # 清理任务执行间隔（秒），默认1小时
```

**参数说明：**
- `chunkSize`: 默认分片大小（字节）
  - 5MB (5242880) - 适合普通网络环境
  - 10MB (10485760) - 适合良好网络环境
  - 20MB (20971520) - 适合高速网络环境
- `sessionExpireHours`: 上传会话过期时间（小时），超过此时间的未完成会话将被清理
- `cleanupInterval`: 自动清理过期会话的执行间隔（秒）

### 步骤1: 运行数据库迁移

```bash
cd F:\work\golang\src\github.com\sky-xhsoft\sky-server

# 创建数据库表
mysql -u root -p sky_server < sqls/cloud_multipart_upload.sql
```

### 步骤2: 在 main.go 中注册服务和路由

在 `cmd/server/main.go` 中添加以下代码：

```go
// 在初始化云盘服务的位置
cloudService := cloud.NewService(db, storageInstance)

// 👇 添加：创建分片上传服务
multipartService := cloud.NewMultipartUploadService(db, storageInstance, cloudService)

// 在初始化 handler 的位置
cloudHandler := handler.NewCloudHandler(cloudService)

// 👇 添加：创建分片上传 handler
multipartHandler := handler.NewMultipartUploadHandler(multipartService)

// 在注册云盘路由的位置
cloudGroup := v1.Group("/cloud")
cloudGroup.Use(authMiddleware)
{
    // 原有的云盘路由...
    cloudGroup.POST("/files/upload", cloudHandler.UploadFile)
    cloudGroup.GET("/files/:id/download", cloudHandler.DownloadFile)
    // ...

    // 👇 添加：分片上传路由
    multipart := cloudGroup.Group("/files/multipart")
    {
        multipart.POST("/init", multipartHandler.InitUpload)
        multipart.POST("/upload", multipartHandler.UploadChunk)
        multipart.GET("/status", multipartHandler.GetUploadStatus)
        multipart.POST("/complete", multipartHandler.CompleteUpload)
        multipart.DELETE("/:sessionId", multipartHandler.AbortUpload)
        multipart.POST("/resume", multipartHandler.ResumeUpload)
    }
}
```

### 步骤3: 编译和运行

```bash
# 编译
go build -o bin/sky-server.exe cmd/server/main.go

# 运行
./bin/sky-server.exe
```

**注意**：服务器启动后会自动启动定时清理任务，根据配置的 `cleanupInterval` 定期清理过期会话。

### 步骤4: 测试 API

#### 测试工具准备

```bash
# 安装 httpie（可选，也可以用 curl 或 Postman）
pip install httpie
```

#### 1. 初始化上传

```bash
http POST http://localhost:9090/api/v1/cloud/files/multipart/init \
  Authorization:"Bearer YOUR_TOKEN" \
  fileName="test.mp4" \
  fileSize:=104857600 \
  fileMd5="abc123def456..." \
  fileType="video/mp4" \
  chunkSize:=5242880

# 响应
{
    "code": 200,
    "data": {
        "sessionId": 1,
        "fileId": "abc123def456...",
        "totalChunks": 20,
        "uploadedChunks": [],
        "status": "uploading"
    }
}
```

#### 2. 上传分片

```bash
# 使用 curl 上传第一个分片
curl -X POST http://localhost:9090/api/v1/cloud/files/multipart/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "sessionId=1" \
  -F "chunkIndex=0" \
  -F "chunkMd5=xyz789..." \
  -F "chunkData=@chunk_0.bin"

# 响应
{
    "code": 200,
    "data": {
        "message": "分片上传成功",
        "chunkIndex": 0,
        "uploaded": true
    }
}
```

#### 3. 查询状态

```bash
http GET http://localhost:9090/api/v1/cloud/files/multipart/status \
  Authorization:"Bearer YOUR_TOKEN" \
  sessionId==1

# 响应
{
    "code": 200,
    "data": {
        "sessionId": 1,
        "totalChunks": 20,
        "uploadedChunks": [0, 1, 2, 3],
        "progress": 0.2,
        "status": "uploading"
    }
}
```

#### 4. 完成上传

```bash
http POST http://localhost:9090/api/v1/cloud/files/multipart/complete \
  Authorization:"Bearer YOUR_TOKEN" \
  sessionId:=1

# 响应
{
    "code": 200,
    "data": {
        "id": 100,
        "fileName": "test.mp4",
        "fileSize": 104857600,
        "md5": "abc123def456...",
        "accessUrl": "http://localhost:9090/files/cloud/...",
        ...
    }
}
```

---

## 🎨 前端集成（HTML + JavaScript）

创建一个简单的测试页面 `upload_test.html`：

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>分片上传测试</title>
    <script src="https://cdn.jsdelivr.net/npm/spark-md5@3.0.2/spark-md5.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .progress-bar { width: 100%; height: 30px; background: #f0f0f0; border-radius: 5px; overflow: hidden; margin: 20px 0; }
        .progress-fill { height: 100%; background: #4CAF50; transition: width 0.3s; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .status.info { background: #e3f2fd; color: #1976d2; }
        .status.success { background: #e8f5e9; color: #388e3c; }
        .status.error { background: #ffebee; color: #c62828; }
        button { padding: 10px 20px; margin: 5px; cursor: pointer; }
        #chunksList { max-height: 200px; overflow-y: auto; border: 1px solid #ddd; padding: 10px; margin: 10px 0; }
    </style>
</head>
<body>
    <h1>📤 分片上传测试</h1>

    <div>
        <input type="file" id="fileInput" />
        <button onclick="startUpload()">开始上传</button>
        <button onclick="pauseUpload()">暂停</button>
        <button onclick="resumeUpload()">继续</button>
        <button onclick="cancelUpload()">取消</button>
    </div>

    <div class="progress-bar">
        <div class="progress-fill" id="progressBar" style="width: 0%"></div>
    </div>

    <div id="status" class="status info">请选择文件...</div>

    <div>
        <h3>已上传分片:</h3>
        <div id="chunksList"></div>
    </div>

    <script>
        const API_BASE = 'http://localhost:9090/api/v1/cloud/files/multipart';
        const TOKEN = 'YOUR_TOKEN'; // 替换为实际的 token

        let uploader = null;
        let paused = false;

        class ChunkedFileUploader {
            constructor(file) {
                this.file = file;
                this.chunkSize = 5 * 1024 * 1024; // 5MB
                this.totalChunks = Math.ceil(file.size / this.chunkSize);
                this.sessionId = null;
                this.uploadedChunks = [];
                this.fileMD5 = null;
            }

            // 计算文件MD5
            async calculateMD5() {
                return new Promise((resolve, reject) => {
                    const blobSlice = File.prototype.slice || File.prototype.mozSlice || File.prototype.webkitSlice;
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

                    const loadNext = () => {
                        const start = currentChunk * chunkSize;
                        const end = Math.min(start + chunkSize, this.file.size);
                        fileReader.readAsArrayBuffer(blobSlice.call(this.file, start, end));
                    };

                    loadNext();
                });
            }

            // 初始化上传
            async init() {
                updateStatus('正在计算文件MD5...', 'info');
                this.fileMD5 = await this.calculateMD5();

                updateStatus('正在初始化上传...', 'info');

                const response = await fetch(`${API_BASE}/init`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${TOKEN}`
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
                if (data.code !== 200) {
                    throw new Error(data.message || '初始化失败');
                }

                this.sessionId = data.data.sessionId;
                this.uploadedChunks = data.data.uploadedChunks || [];

                updateStatus(`初始化成功！会话ID: ${this.sessionId}`, 'success');
                updateChunksList(this.uploadedChunks, this.totalChunks);

                // 保存到 localStorage（断点续传）
                localStorage.setItem('uploadSession', JSON.stringify({
                    sessionId: this.sessionId,
                    fileName: this.file.name,
                    fileMD5: this.fileMD5
                }));

                return this.sessionId;
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

            // 上传单个分片
            async uploadChunk(chunkIndex) {
                const start = chunkIndex * this.chunkSize;
                const end = Math.min(start + this.chunkSize, this.file.size);
                const chunk = this.file.slice(start, end);

                const chunkMD5 = await this.calculateChunkMD5(chunk);

                const formData = new FormData();
                formData.append('sessionId', this.sessionId);
                formData.append('chunkIndex', chunkIndex);
                formData.append('chunkMd5', chunkMD5);
                formData.append('chunkData', chunk);

                const response = await fetch(`${API_BASE}/upload`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${TOKEN}`
                    },
                    body: formData
                });

                const data = await response.json();
                if (data.code !== 200) {
                    throw new Error(data.message || `上传分片 ${chunkIndex} 失败`);
                }
            }

            // 开始上传
            async start() {
                if (!this.sessionId) {
                    await this.init();
                }

                for (let i = 0; i < this.totalChunks; i++) {
                    if (paused) {
                        updateStatus('上传已暂停', 'info');
                        return;
                    }

                    // 跳过已上传的分片（断点续传）
                    if (this.uploadedChunks.includes(i)) {
                        console.log(`分片 ${i} 已上传，跳过`);
                        continue;
                    }

                    try {
                        await this.uploadChunk(i);
                        this.uploadedChunks.push(i);

                        // 更新进度
                        const progress = (this.uploadedChunks.length / this.totalChunks * 100).toFixed(2);
                        updateProgress(progress);
                        updateStatus(`上传中... ${progress}% (${this.uploadedChunks.length}/${this.totalChunks})`, 'info');
                        updateChunksList(this.uploadedChunks, this.totalChunks);
                    } catch (error) {
                        updateStatus(`上传失败: ${error.message}`, 'error');
                        throw error;
                    }
                }

                // 完成上传
                await this.complete();
            }

            // 完成上传
            async complete() {
                updateStatus('正在合并分片...', 'info');

                const response = await fetch(`${API_BASE}/complete`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${TOKEN}`
                    },
                    body: JSON.stringify({
                        sessionId: this.sessionId
                    })
                });

                const data = await response.json();
                if (data.code !== 200) {
                    throw new Error(data.message || '完成上传失败');
                }

                updateStatus('✅ 上传完成！', 'success');
                updateProgress(100);

                // 清理 localStorage
                localStorage.removeItem('uploadSession');

                console.log('文件信息:', data.data);
            }
        }

        // 开始上传
        async function startUpload() {
            const fileInput = document.getElementById('fileInput');
            if (!fileInput.files[0]) {
                alert('请选择文件');
                return;
            }

            uploader = new ChunkedFileUploader(fileInput.files[0]);
            paused = false;

            try {
                await uploader.start();
            } catch (error) {
                console.error('上传失败:', error);
            }
        }

        // 暂停上传
        function pauseUpload() {
            paused = true;
            updateStatus('上传已暂停', 'info');
        }

        // 继续上传
        async function resumeUpload() {
            if (!uploader) {
                // 从 localStorage 恢复
                const session = JSON.parse(localStorage.getItem('uploadSession'));
                if (!session) {
                    alert('没有可恢复的上传');
                    return;
                }

                // 重新选择文件
                const fileInput = document.getElementById('fileInput');
                if (!fileInput.files[0]) {
                    alert('请重新选择文件');
                    return;
                }

                uploader = new ChunkedFileUploader(fileInput.files[0]);
                uploader.sessionId = session.sessionId;
                uploader.fileMD5 = session.fileMD5;

                // 获取已上传的分片
                const response = await fetch(`${API_BASE}/status?sessionId=${session.sessionId}`, {
                    headers: {
                        'Authorization': `Bearer ${TOKEN}`
                    }
                });

                const data = await response.json();
                if (data.code === 200) {
                    uploader.uploadedChunks = data.data.uploadedChunks || [];
                    updateChunksList(uploader.uploadedChunks, uploader.totalChunks);
                }
            }

            paused = false;
            updateStatus('继续上传...', 'info');
            await uploader.start();
        }

        // 取消上传
        async function cancelUpload() {
            if (!uploader || !uploader.sessionId) {
                alert('没有正在进行的上传');
                return;
            }

            const response = await fetch(`${API_BASE}/${uploader.sessionId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${TOKEN}`
                }
            });

            const data = await response.json();
            if (data.code === 200) {
                updateStatus('上传已取消', 'info');
                updateProgress(0);
                uploader = null;
                localStorage.removeItem('uploadSession');
            }
        }

        // UI 更新函数
        function updateProgress(percent) {
            document.getElementById('progressBar').style.width = percent + '%';
        }

        function updateStatus(message, type) {
            const statusEl = document.getElementById('status');
            statusEl.textContent = message;
            statusEl.className = 'status ' + type;
        }

        function updateChunksList(uploaded, total) {
            const listEl = document.getElementById('chunksList');
            listEl.innerHTML = `已上传: ${uploaded.length}/${total}<br>分片索引: [${uploaded.join(', ')}]`;
        }

        // 页面加载时检查是否有未完成的上传
        window.addEventListener('load', () => {
            const session = localStorage.getItem('uploadSession');
            if (session) {
                const data = JSON.parse(session);
                updateStatus(`检测到未完成的上传: ${data.fileName}`, 'info');
            }
        });
    </script>
</body>
</html>
```

打开 `upload_test.html` 即可测试分片上传功能！

---

## ✅ 验证清单

- [ ] 数据库表已创建
- [ ] 服务和 Handler 已注册
- [ ] 路由已配置
- [ ] 服务器可以正常启动
- [ ] 可以初始化上传会话
- [ ] 可以上传分片
- [ ] 可以查询上传状态
- [ ] 可以完成上传
- [ ] 暂停后可以继续上传（断点续传）
- [ ] 定时清理任务正常运行

---

## 🎉 完成！

恭喜！你已经成功集成了分片上传和断点续传功能。

**下一步**:
- 查看 `docs/multipart-upload-implementation-summary.md` 了解详细信息
- 根据实际需求调整分片大小和过期时间
- 添加前端 UI 界面
- 配置生产环境的存储和性能优化

有任何问题请参考文档或查看源码注释！
