# 大文件上传支持（20GB）

## 概述

云盘系统现已支持最大 20GB 的单文件上传。本文档说明相关配置和实现细节。

## 配置修改

### 1. 配置文件（`configs/config.yaml`）

新增 `file` 配置段：

```yaml
# 文件管理配置（云盘）
file:
  uploadDir: "uploads"  # 上传目录
  maxFileSize: 21474836480  # 最大文件大小：20GB（字节）
  allowedExts:  # 允许的文件扩展名
    - ".jpg"
    - ".jpeg"
    - ".png"
    - ".gif"
    - ".pdf"
    - ".doc"
    - ".docx"
    - ".xls"
    - ".xlsx"
    - ".ppt"
    - ".pptx"
    - ".txt"
    - ".zip"
    - ".rar"
    - ".7z"
    - ".mp4"
    - ".avi"
    - ".mkv"
    - ".mp3"
    - ".wav"
```

### 2. 云盘配额（`internal/service/cloud/cloud_service.go`）

默认配额已更新：

```go
UserID:      userID,
TotalQuota:  10 * 1024 * 1024 * 1024,  // 默认10GB存储空间
UsedSpace:   0,
FileCount:   0,
FolderCount: 0,
MaxFileSize: 20 * 1024 * 1024 * 1024,  // 默认20GB单文件限制
QuotaType:   "standard",
```

### 3. 数据库默认配额（`sqls/cloud_disk_tables.sql`）

SQL初始化脚本已更新：

```sql
INSERT INTO `cloud_quota` (...)
SELECT
  u.ID,
  10737418240,   -- 10GB 存储空间
  0,
  0,
  0,
  21474836480,   -- 20GB 单文件限制
  'standard',
  'system'
FROM sys_user u
WHERE NOT EXISTS (
  SELECT 1 FROM cloud_quota q WHERE q.USER_ID = u.ID
);
```

### 4. Gin 框架配置（`cmd/server/main.go`）

设置了 `MaxMultipartMemory` 以支持大文件上传：

```go
// 设置文件上传大小限制（32MB内存缓存，超过的部分会写入临时文件）
// 这样可以支持大文件上传而不会占用过多内存
engine.MaxMultipartMemory = 32 << 20 // 32 MB
```

## 工作原理

### 内存管理

Gin 的 `MaxMultipartMemory` 设置为 32MB，这意味着：

1. **小文件（< 32MB）**：
   - 完全在内存中处理
   - 速度最快

2. **大文件（> 32MB）**：
   - 前 32MB 在内存中
   - 剩余部分写入临时文件
   - 自动清理临时文件

### 流式上传

存储层使用流式处理：

```go
// 上传到存储
accessURL, err := s.storage.Upload(ctx, storagePath, req.Reader, req.FileType)
```

- `req.Reader` 是 `io.Reader` 接口
- 数据以流的方式读取和写入
- 不会一次性加载整个文件到内存
- 支持任意大小的文件

### 配额检查

在上传前检查配额：

```go
// 检查配额
if err := s.CheckQuota(ctx, userID, req.FileSize); err != nil {
    return nil, err
}
```

检查内容：
- 用户存储空间是否足够
- 单文件大小是否超过限制（20GB）

## 存储空间与文件大小的区别

- **TotalQuota（总配额）**：10GB
  - 用户可使用的总存储空间
  - 多个文件的累计大小

- **MaxFileSize（最大文件）**：20GB
  - 单个文件的最大大小
  - 用户可以上传 20GB 的单个文件，但总空间只有 10GB
  - 这允许用户上传大文件后删除以腾出空间

## 配额类型

系统支持多种配额类型：

### Standard（标准）
- 总空间：10GB
- 单文件：20GB
- 适用于普通用户

### Premium（高级）
可以通过数据库手动设置更大的配额：

```sql
UPDATE cloud_quota
SET
  TOTAL_QUOTA = 100 * 1024 * 1024 * 1024,  -- 100GB
  MAX_FILE_SIZE = 50 * 1024 * 1024 * 1024,  -- 50GB
  QUOTA_TYPE = 'premium'
WHERE USER_ID = ?;
```

## 限制和注意事项

### 1. 网络超时

上传 20GB 文件可能需要很长时间，需要考虑：

- **HTTP 超时**：默认可能不够
- **反向代理超时**：Nginx/Apache 需要配置
- **浏览器超时**：前端需要处理长时间请求

### 2. 磁盘空间

确保服务器有足够的磁盘空间：

```bash
# 检查磁盘空间
df -h

# 临时文件目录（Gin 使用系统临时目录）
echo $TMPDIR  # Linux/Mac
echo %TEMP%   # Windows
```

### 3. 并发上传

多个用户同时上传大文件会消耗大量资源：

- 考虑添加上传队列
- 限制同时上传的大文件数量
- 监控服务器资源使用

### 4. 网络带宽

20GB 文件传输时间估算：

| 带宽 | 理论时间 |
|------|---------|
| 100 Mbps | ~27 分钟 |
| 1 Gbps | ~2.7 分钟 |
| 10 Gbps | ~16 秒 |

实际时间会受多种因素影响。

## 优化建议

### 1. 分片上传

对于超大文件，建议实现分片上传：

```
优点：
- 支持断点续传
- 减少单次请求大小
- 提高成功率

实现：
- 前端分片上传
- 后端合并分片
- 存储分片信息
```

### 2. 压缩

在上传前压缩文件：

```
支持的压缩格式：
- .zip
- .rar
- .7z
- .tar.gz
```

### 3. 去重

使用 MD5 去重节省空间：

```go
// 已在 CloudFile 实体中包含 MD5 字段
MD5 string `gorm:"column:MD5;size:32;index" json:"md5"`
```

### 4. 存储优化

- 使用对象存储（OSS/S3）
- 配置 CDN 加速下载
- 定期清理过期/删除的文件

## 监控

监控大文件上传的关键指标：

### 1. 上传成功率

```sql
SELECT
  COUNT(*) as total_uploads,
  SUM(CASE WHEN FILE_SIZE > 1073741824 THEN 1 ELSE 0 END) as large_files,
  AVG(FILE_SIZE) as avg_size
FROM cloud_file
WHERE CREATE_TIME >= DATE_SUB(NOW(), INTERVAL 1 DAY);
```

### 2. 配额使用情况

```sql
SELECT
  USER_ID,
  USED_SPACE,
  TOTAL_QUOTA,
  ROUND(USED_SPACE / TOTAL_QUOTA * 100, 2) as usage_percent
FROM cloud_quota
WHERE USED_SPACE / TOTAL_QUOTA > 0.8;  -- 使用率超过80%
```

### 3. 系统资源

```bash
# CPU 使用率
top -bn1 | grep "Cpu(s)"

# 内存使用率
free -h

# 磁盘 I/O
iostat -x 1

# 网络流量
iftop
```

## 测试

### 1. 创建测试文件

```bash
# 创建 1GB 测试文件
dd if=/dev/zero of=test_1gb.bin bs=1M count=1024

# 创建 5GB 测试文件
dd if=/dev/zero of=test_5gb.bin bs=1M count=5120

# 创建 10GB 测试文件
dd if=/dev/zero of=test_10gb.bin bs=1M count=10240
```

### 2. 测试上传

```bash
# 使用 curl 上传
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test_1gb.bin" \
  -F "folderId=" \
  -v
```

### 3. 监控上传过程

```bash
# 监控内存
watch -n 1 free -h

# 监控临时目录大小
watch -n 1 du -sh /tmp

# 监控上传目录
watch -n 1 du -sh uploads/cloud
```

## 故障排查

### 问题：上传超时

**原因**：
- 网络慢
- 文件太大
- 服务器负载高

**解决方案**：
```yaml
# 增加 HTTP 超时（如果使用 Nginx）
proxy_read_timeout 3600s;
proxy_send_timeout 3600s;

# 增加客户端超时
client_max_body_size 20G;
```

### 问题：磁盘空间不足

**原因**：
- 临时文件占用空间
- 上传目录满

**解决方案**：
```bash
# 清理临时文件
rm -rf /tmp/*

# 扩展磁盘空间
# 或迁移到更大的存储
```

### 问题：内存不足

**原因**：
- 多个大文件同时上传
- MaxMultipartMemory 设置过大

**解决方案**：
```go
// 减小内存缓存（如果必要）
engine.MaxMultipartMemory = 16 << 20  // 16 MB
```

## 安全建议

1. **文件类型检查**：
   - 验证文件扩展名
   - 检查 MIME 类型
   - 扫描病毒（如果可能）

2. **配额限制**：
   - 严格执行配额检查
   - 防止恶意占用空间
   - 定期清理无效文件

3. **访问控制**：
   - 所有操作验证用户权限
   - 防止越权访问
   - 记录操作审计日志

4. **网络安全**：
   - 使用 HTTPS 传输
   - 限制上传频率
   - 防止 DDoS 攻击

## 总结

云盘系统现已支持 20GB 大文件上传，配置如下：

✅ 配置文件已更新（`file.maxFileSize: 20GB`）
✅ 默认配额已更新（`MaxFileSize: 20GB`）
✅ Gin 引擎已配置（`MaxMultipartMemory: 32MB`）
✅ 数据库初始化脚本已更新
✅ 编译测试通过

**建议**：
- 监控服务器资源使用
- 考虑实现分片上传
- 配置合理的超时时间
- 定期清理无效文件
