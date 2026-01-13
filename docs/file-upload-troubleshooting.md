# 文件上传故障排查

## 问题：文件大小超过限制（最大100MB）

### 错误信息
```json
{
  "code": 500,
  "message": "上传文件失败: [10001] 文件大小超过限制（最大100MB）",
  "timestamp": "2026-01-13T12:20:00+08:00"
}
```

### 原因分析

这个错误来自 `file service`，而不是 `cloud service`。虽然配置文件已经设置为 20GB，但服务器可能：

1. **还没有重启**，仍在使用旧配置
2. **配置文件路径不正确**
3. **使用了默认配置**（当 `cfg.File.MaxFileSize == 0` 时）

### 解决方案

#### 步骤 1：确认配置文件正确

检查 `configs/config.yaml` 是否包含以下配置：

```yaml
# 文件管理配置（云盘）
file:
  uploadDir: "uploads"
  maxFileSize: 21474836480  # 20GB
  allowedExts:
    - ".jpg"
    - ".jpeg"
    - ".png"
    - ".gif"
    # ... 其他扩展名
```

**注意**：
- ✅ 使用 `maxFileSize`（驼峰命名）
- ❌ 不要使用 `max_file_size`（下划线）
- 数值是**字节**：`21474836480` = 20GB

#### 步骤 2：重启服务器

**Windows**:
```bash
# 停止当前服务器（Ctrl+C）

# 重新编译
go build ./cmd/server

# 启动服务器
.\sky-server.exe
```

**Linux/Mac**:
```bash
# 停止当前服务器
kill -9 $(pgrep sky-server)

# 重新编译
go build ./cmd/server

# 启动服务器
./sky-server
```

#### 步骤 3：验证配置加载

启动服务器后，检查日志中是否有配置信息：

```bash
# 查看启动日志
tail -f logs/app.log

# 或在控制台查看输出
```

预期应该看到类似信息：
```
{"level":"info","time":"...","msg":"Services initialized"}
```

#### 步骤 4：测试上传

使用小文件测试（避免浪费时间）：

```bash
# 创建 150MB 测试文件（超过旧的 100MB 限制）
dd if=/dev/zero of=test_150mb.bin bs=1M count=150

# 上传测试
curl -X POST http://localhost:9090/api/v1/cloud/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test_150mb.bin" \
  -F "folderId=" \
  -v
```

**预期结果**：
- ✅ 如果配置正确：上传成功
- ❌ 如果仍然失败：继续下面的排查

### 高级排查

#### 检查 1：确认配置文件被读取

在 `cmd/server/main.go` 的第 62 行之后添加调试日志：

```go
// 1. 加载配置
cfg, err := config.Load()
if err != nil {
    fmt.Printf("Failed to load config: %v\n", err)
    os.Exit(1)
}

// 添加调试日志
fmt.Printf("DEBUG: File MaxFileSize = %d bytes (%.2f GB)\n",
    cfg.File.MaxFileSize,
    float64(cfg.File.MaxFileSize)/(1024*1024*1024))
```

重新编译运行，应该看到：
```
DEBUG: File MaxFileSize = 21474836480 bytes (20.00 GB)
```

如果看到：
```
DEBUG: File MaxFileSize = 0 bytes (0.00 GB)
```

这说明配置文件没有正确读取。

#### 检查 2：确认配置文件路径

在 `internal/config/config.go` 的 `Load()` 函数中，检查配置文件搜索路径：

```go
func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")  // 当前目录的 configs 文件夹
    viper.AddConfigPath(".")          // 当前目录

    // ... 读取配置
}
```

**确保**：
- 运行服务器时，当前工作目录包含 `configs/config.yaml`
- 或者在当前目录有 `config.yaml`

**验证方法**：
```bash
# 进入项目根目录
cd F:\work\golang\src\github.com\sky-xhsoft\sky-server

# 检查配置文件是否存在
ls configs/config.yaml  # Linux/Mac
dir configs\config.yaml # Windows

# 从项目根目录启动服务器
go run ./cmd/server
```

#### 检查 3：YAML 语法问题

确认 YAML 格式正确：

```yaml
# 正确 ✅
file:
  uploadDir: "uploads"
  maxFileSize: 21474836480

# 错误 ❌（缩进不对）
file:
uploadDir: "uploads"
maxFileSize: 21474836480

# 错误 ❌（使用了 Tab 而不是空格）
file:
→uploadDir: "uploads"  # Tab 字符
→maxFileSize: 21474836480
```

**验证工具**：
```bash
# 在线验证
# 复制 config.yaml 内容到: https://www.yamllint.com/

# 或使用 Python 验证
python -c "import yaml; yaml.safe_load(open('configs/config.yaml'))"
```

### 常见错误

#### 错误 1：配置未生效

**症状**：启动服务器后仍然显示 100MB 限制

**解决**：
1. 确认修改的是正确的配置文件
2. 确认服务器已完全重启（不是热重载）
3. 检查是否有多个配置文件

```bash
# 查找所有 config.yaml 文件
find . -name "config.yaml"  # Linux/Mac
dir /s config.yaml          # Windows

# 确保只修改了正在使用的那个
```

#### 错误 2：配置文件权限问题

**症状**：服务器启动失败或配置读取失败

**解决**：
```bash
# Linux/Mac
chmod 644 configs/config.yaml

# Windows
# 右键 -> 属性 -> 安全 -> 编辑权限
```

#### 错误 3：环境变量覆盖

**症状**：配置文件正确但仍使用旧值

**原因**：环境变量可能覆盖了配置文件

**检查**：
```bash
# 检查环境变量
env | grep -i file  # Linux/Mac
set | findstr /i file  # Windows

# 如果发现相关环境变量，清除它
unset FILE_MAXFILESIZE  # Linux/Mac
set FILE_MAXFILESIZE=   # Windows
```

### 验证成功的标志

配置正确加载后，您应该能够：

1. ✅ 上传 150MB 文件（超过旧的 100MB 限制）
2. ✅ 上传 1GB 文件
3. ✅ 上传最大 20GB 文件
4. ❌ 上传超过 20GB 文件会收到新的错误：
   ```json
   {
     "message": "文件大小超过限制（最大20480MB）"
   }
   ```

### 性能建议

上传大文件时的最佳实践：

1. **分片上传**：
   - 将大文件分成多个小块
   - 逐个上传
   - 服务器端合并

2. **断点续传**：
   - 记录上传进度
   - 失败时从断点继续

3. **进度显示**：
   - 显示上传百分比
   - 估算剩余时间

4. **网络优化**：
   - 使用 HTTP/2
   - 启用压缩（如果适用）
   - 考虑 CDN 加速

### 快速检查清单

在报告问题前，请确认：

- [ ] 配置文件 `configs/config.yaml` 存在
- [ ] 配置文件包含 `file.maxFileSize: 21474836480`
- [ ] YAML 语法正确（使用空格缩进，不是 Tab）
- [ ] 服务器已完全重启
- [ ] 从项目根目录启动服务器
- [ ] 没有环境变量覆盖配置
- [ ] 测试文件大小在合理范围内（先测试 150MB）

### 获取帮助

如果问题仍未解决，请提供：

1. **配置文件内容**：
   ```bash
   cat configs/config.yaml
   ```

2. **服务器启动日志**：
   ```bash
   # 启动服务器并保存日志
   go run ./cmd/server 2>&1 | tee server.log
   ```

3. **测试命令和响应**：
   ```bash
   curl -X POST ... -v > response.log 2>&1
   ```

4. **环境信息**：
   - 操作系统
   - Go 版本
   - 项目版本/分支

### 相关文档

- [large-file-upload.md](./large-file-upload.md) - 大文件上传详细文档
- [cloud-service-implementation.md](./cloud-service-implementation.md) - 云盘服务实现
- [config.example.yaml](../configs/config.example.yaml) - 配置文件示例
