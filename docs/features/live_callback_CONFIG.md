# 腾讯云直播回调配置说明

## callbackKey 配置项说明

### 什么是 callbackKey？

`callbackKey` 是腾讯云直播回调的安全密钥，用于验证回调请求的真实性，防止恶意请求伪造回调数据。

### 配置位置

在 `configs/config.yaml` 文件中的 `tencentCloud` 部分：

```yaml
tencentCloud:
  secretID: "your-tencent-secret-id"
  secretKey: "your-tencent-secret-key"
  region: "ap-guangzhou"
  callbackKey: "your-callback-key"  # 回调密钥
```

### 配置方式

#### 方式一：不使用签名验证（开发环境）

如果您在开发环境中测试，或者不需要签名验证，可以将 `callbackKey` 设置为空字符串：

```yaml
tencentCloud:
  callbackKey: ""
```

**特点**：
- ✅ 配置简单，无需在腾讯云控制台配置
- ✅ 适合开发测试环境
- ⚠️ 不验证签名，存在安全风险
- ⚠️ 任何人都可以伪造回调请求

#### 方式二：使用签名验证（生产环境推荐）

在生产环境中，强烈建议启用签名验证：

```yaml
tencentCloud:
  callbackKey: "abc123xyz789"  # 设置一个复杂的密钥
```

**特点**：
- ✅ 验证回调请求的真实性
- ✅ 防止恶意请求伪造数据
- ✅ 生产环境必备
- ⚠️ 需要在腾讯云控制台配置相同的密钥

### 腾讯云控制台配置步骤

如果启用了签名验证（callbackKey 不为空），需要在腾讯云控制台进行相应配置：

#### 1. 登录腾讯云控制台

访问：https://console.cloud.tencent.com/live

#### 2. 进入回调配置

导航路径：**功能配置** > **事件中心** > **直播回调**

#### 3. 配置回调密钥

在回调配置页面：
- 找到"回调密钥"配置项
- 输入与 `config.yaml` 中 `callbackKey` 相同的值
- 保存配置

#### 4. 配置回调URL

为每种事件类型配置对应的回调URL：

**基础事件**：
- 推流回调URL: `https://your-domain.com/api/v1/live/callback/push`
- 断流回调URL: `https://your-domain.com/api/v1/live/callback/disconnect`

**录制相关**：
- 录制回调URL: `https://your-domain.com/api/v1/live/callback/recording-file`
- 录制状态回调URL: `https://your-domain.com/api/v1/live/callback/recording-status`

**其他事件**：
- 截图回调URL: `https://your-domain.com/api/v1/live/callback/screenshot`
- 审核回调URL: `https://your-domain.com/api/v1/live/callback/video-audit`
- ... 等等（详见完整文档）

### 签名验证原理

#### 签名生成算法

腾讯云使用以下算法生成签名：

```
sign = MD5(callbackKey + t)
```

其中：
- `callbackKey`: 配置的回调密钥
- `t`: 签名过期时间戳（Unix时间戳，秒）
- `MD5`: MD5哈希算法（32位小写）

#### 验证流程

1. 腾讯云发送回调请求时，在请求体中包含 `t` 和 `sign` 字段
2. 服务器接收到回调请求后：
   - 检查 `t` 是否过期（当前时间 > t 则过期）
   - 使用本地 `callbackKey` 和请求中的 `t` 计算签名
   - 比对计算出的签名与请求中的 `sign` 是否一致
3. 验证通过则处理回调，否则拒绝

#### 示例

假设：
- `callbackKey` = "abc123"
- `t` = 1234567890

计算签名：
```
sign = MD5("abc123" + "1234567890")
     = MD5("abc1231234567890")
     = "e10adc3949ba59abbe56e057f20f883e"
```

### 配置示例

#### 开发环境配置

```yaml
# configs/config.yaml
tencentCloud:
  secretID: "AKIDxxxxxxxxxxxxx"
  secretKey: "xxxxxxxxxxxxxxxx"
  region: "ap-guangzhou"
  callbackKey: ""  # 不验证签名
  live:
    enabled: true
    pushDomain: "push.example.com"
    playDomain: "play.example.com"
```

#### 生产环境配置

```yaml
# configs/config.yaml
tencentCloud:
  secretID: "AKIDxxxxxxxxxxxxx"
  secretKey: "xxxxxxxxxxxxxxxx"
  region: "ap-guangzhou"
  callbackKey: "your-strong-callback-key-here"  # 使用强密钥
  live:
    enabled: true
    pushDomain: "push.example.com"
    playDomain: "play.example.com"
```

### 密钥生成建议

生成一个安全的 `callbackKey`：

#### 方法一：使用随机字符串生成器

```bash
# Linux/Mac
openssl rand -hex 16

# 输出示例：a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

#### 方法二：使用在线工具

访问：https://www.random.org/strings/

设置：
- 字符串数量：1
- 字符串长度：32
- 字符类型：字母+数字

#### 方法三：手动创建

创建一个包含大小写字母、数字的32位字符串，例如：
```
Abc123Xyz789Def456Uvw012Ghi345Jk
```

### 安全建议

1. **使用强密钥**
   - 长度至少16位，建议32位
   - 包含大小写字母、数字
   - 避免使用简单密码如 "123456"

2. **定期更换**
   - 建议每3-6个月更换一次
   - 更换时同步更新腾讯云控制台配置

3. **保密存储**
   - 不要将密钥提交到代码仓库
   - 使用环境变量或密钥管理服务
   - 限制配置文件的访问权限

4. **监控日志**
   - 定期检查签名验证失败的日志
   - 发现异常立即更换密钥

### 故障排查

#### 问题1：签名验证失败

**现象**：日志中出现 "签名验证失败" 警告

**原因**：
- 本地 `callbackKey` 与腾讯云控制台配置不一致
- 时间戳过期（服务器时间不准确）

**解决方法**：
1. 检查 `config.yaml` 中的 `callbackKey`
2. 检查腾讯云控制台的回调密钥配置
3. 确保两者完全一致（区分大小写）
4. 检查服务器时间是否准确

#### 问题2：回调未收到

**现象**：腾讯云发送回调，但服务器未收到

**原因**：
- 回调URL配置错误
- 服务器防火墙阻止
- 服务未启动

**解决方法**：
1. 检查腾讯云控制台的回调URL配置
2. 确保URL可以从公网访问
3. 检查服务器防火墙设置
4. 查看服务器日志

#### 问题3：时间戳过期

**现象**：日志中出现 "回调签名已过期"

**原因**：
- 服务器时间不准确
- 网络延迟过大

**解决方法**：
1. 同步服务器时间：
   ```bash
   # Linux
   ntpdate ntp.aliyun.com

   # 或使用 systemd-timesyncd
   timedatectl set-ntp true
   ```
2. 检查服务器时区设置
3. 如果网络延迟大，可以适当放宽时间验证

### 测试验证

#### 测试签名验证

使用以下命令测试签名验证：

```bash
# 设置变量
CALLBACK_KEY="your-callback-key"
TIMESTAMP=$(date +%s)
SIGN=$(echo -n "${CALLBACK_KEY}${TIMESTAMP}" | md5sum | cut -d' ' -f1)

# 发送测试请求
curl -X POST "http://localhost:9090/api/v1/live/callback/push" \
  -H "Content-Type: application/json" \
  -d "{
    \"event_type\": 1,
    \"stream_id\": \"test_001\",
    \"t\": ${TIMESTAMP},
    \"sign\": \"${SIGN}\",
    \"event_time\": ${TIMESTAMP},
    \"push_domain\": \"push.example.com\",
    \"app_name\": \"live\",
    \"stream_name\": \"test\"
  }"
```

**预期结果**：
- 如果签名正确，返回 `{"code": 0}`
- 如果签名错误，返回 `{"code": 1, "msg": "invalid signature"}`

### 参考文档

- [腾讯云直播回调配置](https://cloud.tencent.com/document/product/267/20388)
- [腾讯云直播回调事件](https://cloud.tencent.com/document/product/267/32744)
- [腾讯云直播回调签名](https://cloud.tencent.com/document/product/267/32744#.E4.BA.8B.E4.BB.B6.E6.B6.88.E6.81.AF.E9.80.9A.E7.9F.A5.E7.AD.BE.E5.90.8D)
