# 直播回调事件功能文档

> 整合自 25 个独立文档

---

## API_CODE_FIX

## 问题描述

前端页面（高光切片列表和录制列表）显示为空，即使后端已经返回了数据。

## 问题原因

后端 API 返回的响应码是 `200`，但前端代码只检查 `res.code === 0`，导致数据无法正确解析和显示。

**后端实际返回：**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

**前端检查：**
```typescript
if (res.code === 0) {  // ❌ 永远不会匹配
  // 解析数据...
}
```

## 修复方案

修改前端代码，同时支持 `code === 200` 和 `code === 0` 两种响应格式。

### 修复文件

**1. LiveHighlightClips.vue**

**修改位置：** 第 308 行

**修改前：**
```typescript
if (res.code === 0) {
  // 解析事件数据
  tableData.value = res.data.list.map((item: any) => {
```

**修改后：**
```typescript
if (res.code === 200 || res.code === 0) {
  // 解析事件数据
  tableData.value = res.data.list.map((item: any) => {
```

**2. LiveRecordings.vue**

**修改位置：** 第 433 行、第 378 行、第 390 行

**修改前：**
```typescript
// 主数据加载
if (res.code === 0) {
  tableData.value = res.data.list

// 统计数据 - 总数
if (totalRes.code === 0) {
  stats.total = totalRes.data.total

// 统计数据 - 今日
if (todayRes.code === 0) {
  stats.today = todayRes.data.total
```

**修改后：**
```typescript
// 主数据加载
if (res.code === 200 || res.code === 0) {
  tableData.value = res.data.list

// 统计数据 - 总数
if (totalRes.code === 200 || totalRes.code === 0) {
  stats.total = totalRes.data.total

// 统计数据 - 今日
if (todayRes.code === 200 || todayRes.code === 0) {
  stats.today = todayRes.data.total
```

## 兼容性说明

使用 `||` 运算符同时支持两种响应码格式：

```typescript
if (res.code === 200 || res.code === 0) {
  // 处理数据
}
```

这样可以兼容：
- ✅ 标准 HTTP 状态码格式（`code: 200`）
- ✅ 自定义业务状态码格式（`code: 0`）

## 测试验证

### 1. 高光切片列表测试

**步骤：**
1. 打开浏览器开发者工具（F12）
2. 访问高光切片列表页面
3. 查看 Network 标签中的 API 响应
4. 检查表格是否显示数据

**预期结果：**
- ✅ API 返回 `code: 200`
- ✅ 前端正确解析数据
- ✅ 表格显示标题、关键词等信息
- ✅ 数据条数与后端返回一致

### 2. 录制列表测试

**步骤：**
1. 访问录制列表页面
2. 检查统计卡片和表格数据

**预期结果：**
- ✅ 统计卡片显示正确的数字
- ✅ 表格显示录制文件列表
- ✅ 所有字段正常显示

### 3. 控制台检查

打开浏览器控制台（F12 -> Console），检查是否有错误信息：

```javascript
// 应该没有以下错误：
// ❌ Cannot read property 'list' of undefined
// ❌ Cannot read property 'total' of undefined
```

## 刷新步骤

修改完成后，请按以下步骤刷新：

1. **清除浏览器缓存**
   - Chrome: Ctrl + Shift + Delete
   - 或者使用硬刷新：Ctrl + F5

2. **重新加载页面**
   - 访问：`http://localhost:8080/#/live/highlight-clips`
   - 或者：`http://localhost:8080/#/live/recordings`

3. **检查数据显示**
   - 表格应该显示数据
   - 统计卡片应该显示正确的数字

## 后续建议

### 统一响应码格式

建议后端统一使用一种响应码格式，避免混淆：

**选项 1：使用 HTTP 状态码（推荐）**
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

**选项 2：使用业务状态码**
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 创建响应拦截器

在前端创建统一的响应拦截器，自动处理不同的响应码格式：

```typescript
// src/utils/request.ts
api.interceptors.response.use(
  (response) => {
    const { code, data, message } = response.data

    // 统一处理成功响应
    if (code === 200 || code === 0) {
      return response.data
    }

    // 处理错误响应
    Message.error(message || '请求失败')
    return Promise.reject(new Error(message))
  },
  (error) => {
    Message.error('网络错误')
    return Promise.reject(error)
  }
)
```

## 相关文件

- `sky-web/src/pages/LiveHighlightClips.vue` - 高光切片列表页面
- `sky-web/src/pages/LiveRecordings.vue` - 录制列表页面
- `sky-web/src/api/live.ts` - API 接口定义

## 更新日期

2026-02-02

---

## CONFIG

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

---

## DEFAULT_TIME_RANGE

## 更新内容

已将以下两个页面的默认查询时间范围设置为**最近7天**：

1. **LiveHighlightClips.vue** - 直播高光切片列表
2. **LiveRecordings.vue** - 直播录制列表

## 修改详情

### 修改前
```typescript
// 搜索表单
const searchForm = reactive({
  streamName: '',
  domainName: '',
  appName: '',
  timeRange: []  // 空数组，不设置默认时间范围
})
```

### 修改后
```typescript
// 搜索表单 - 默认查询最近7天数据
const searchForm = reactive({
  streamName: '',
  domainName: '',
  appName: '',
  timeRange: [
    dayjs().subtract(7, 'day').startOf('day').toDate(),  // 7天前的00:00:00
    dayjs().endOf('day').toDate()                         // 今天的23:59:59
  ]
})
```

## 功能说明

### 默认时间范围
- **开始时间**: 7天前的 00:00:00
- **结束时间**: 今天的 23:59:59

### 示例
如果今天是 2026-02-02，则默认查询时间范围为：
- 开始时间: `2026-01-26 00:00:00`
- 结束时间: `2026-02-02 23:59:59`

## 用户体验改进

### 改进前
- 页面加载时不设置时间范围
- 查询所有历史数据
- 数据量大时加载慢
- 用户需要手动设置时间范围

### 改进后
- ✅ 页面加载时自动设置最近7天
- ✅ 只查询最近7天的数据
- ✅ 加载速度更快
- ✅ 符合用户常用场景
- ✅ 用户仍可手动修改时间范围

## 时间范围选择器显示

页面加载后，时间范围选择器会自动显示：
```
[2026-01-26 00:00:00] 至 [2026-02-02 23:59:59]
```

用户可以：
1. **保持默认** - 直接点击"查询"按钮，查询最近7天数据
2. **修改范围** - 点击时间选择器，选择其他时间范围
3. **清除范围** - 点击清除按钮，查询所有数据
4. **重置** - 点击"重置"按钮，恢复到默认的7天范围

## 重置功能

点击"重置"按钮时，会恢复到默认的7天时间范围：

```typescript
const handleReset = () => {
  searchForm.streamName = ''
  searchForm.domainName = ''
  searchForm.appName = ''
  searchForm.timeRange = [
    dayjs().subtract(7, 'day').startOf('day').toDate(),
    dayjs().endOf('day').toDate()
  ]
  pagination.current = 1
  loadData()
}
```

## 后端查询参数

当设置了时间范围后，前端会将时间转换为后端需要的格式：

```typescript
if (searchForm.timeRange && searchForm.timeRange.length === 2) {
  params.startTime = dayjs(searchForm.timeRange[0]).format('YYYY-MM-DD HH:mm:ss')
  params.endTime = dayjs(searchForm.timeRange[1]).format('YYYY-MM-DD HH:mm:ss')
}
```

发送到后端的参数示例：
```
GET /api/v1/live/callback/events?eventType=highlight&startTime=2026-01-26%2000:00:00&endTime=2026-02-02%2023:59:59&pageNum=1&pageSize=20
```

## 性能优化

### 数据库查询优化
通过设置默认时间范围，后端查询会使用时间索引：

```sql
SELECT * FROM live_callback_event
WHERE EVENT_TYPE = 'highlight'
  AND CREATE_TIME >= '2026-01-26 00:00:00'
  AND CREATE_TIME <= '2026-02-02 23:59:59'
ORDER BY CREATE_TIME DESC
LIMIT 20 OFFSET 0;
```

### 性能提升
- ✅ 减少查询数据量
- ✅ 利用时间索引加速查询
- ✅ 降低数据库负载
- ✅ 提升页面加载速度

## 其他时间范围选项

如果需要其他默认时间范围，可以修改代码：

### 最近24小时
```typescript
timeRange: [
  dayjs().subtract(1, 'day').toDate(),
  dayjs().toDate()
]
```

### 最近30天
```typescript
timeRange: [
  dayjs().subtract(30, 'day').startOf('day').toDate(),
  dayjs().endOf('day').toDate()
]
```

### 本月
```typescript
timeRange: [
  dayjs().startOf('month').toDate(),
  dayjs().endOf('day').toDate()
]
```

### 本周
```typescript
timeRange: [
  dayjs().startOf('week').toDate(),
  dayjs().endOf('day').toDate()
]
```

## 测试验证

### 1. 页面加载测试
- 打开高光切片页面
- 检查时间范围选择器是否显示最近7天
- 检查表格数据是否正确加载

### 2. 查询测试
- 点击"查询"按钮
- 验证返回的数据是否在7天范围内
- 检查分页是否正常

### 3. 重置测试
- 修改时间范围为其他值
- 点击"重置"按钮
- 验证时间范围是否恢复到默认的7天

### 4. 清除测试
- 点击时间范围的清除按钮
- 点击"查询"按钮
- 验证是否查询所有数据

## 注意事项

1. **时区问题**
   - dayjs 使用本地时区
   - 确保服务器和客户端时区一致
   - 或在后端统一转换为 UTC

2. **数据为空**
   - 如果最近7天没有数据，表格会显示"暂无数据"
   - 用户可以扩大时间范围查询历史数据

3. **性能考虑**
   - 7天是一个平衡点，既能满足常用需求，又不会查询过多数据
   - 如果数据量特别大，可以考虑缩短为3天或1天

## 相关文件

- `sky-web/src/pages/LiveHighlightClips.vue` - 高光切片页面
- `sky-web/src/pages/LiveRecordings.vue` - 录制列表页面

## 更新日期

2026-02-02

---

## FRONTEND_UPDATE

## 更新内容

更新了高光切片列表页面（LiveHighlightClips.vue），以支持腾讯云返回的新数据格式，包括标题、摘要、关键词等AI生成的内容。

## 修改文件

- `sky-web/src/pages/LiveHighlightClips.vue`

## 主要改动

### 1. 数据解析逻辑更新

**修改前：**
```typescript
tableData.value = res.data.list.map((item: any) => {
  const eventData = JSON.parse(item.eventData)
  return {
    id: item.id,
    streamId: item.streamId,
    streamName: item.streamName,
    domainName: item.domainName,
    appName: item.appName,
    clipUrl: eventData.clip_url,
    startTime: eventData.start_time != null ? eventData.start_time * 1000 : null,
    endTime: eventData.end_time != null ? eventData.end_time * 1000 : null,
    score: eventData.score,
    eventTime: item.eventTime != null ? item.eventTime * 1000 : null,
    createTime: item.createTime
  }
})
```

**修改后：**
```typescript
tableData.value = res.data.list.map((item: any) => {
  const eventData = JSON.parse(item.eventData)
  return {
    id: item.id,
    streamId: item.streamId,
    streamName: item.streamName,
    domainName: item.domainName,
    appName: item.appName,
    clipUrl: eventData.video_store_url || eventData.clip_url,  // 优先使用新字段
    startTime: (eventData.begin_time || eventData.start_time) != null
      ? (eventData.begin_time || eventData.start_time) * 1000
      : null,
    endTime: eventData.end_time != null ? eventData.end_time * 1000 : null,
    title: eventData.title || '',              // 新增：标题
    summary: eventData.summary || '',          // 新增：摘要
    keyWords: eventData.key_words || [],       // 新增：关键词
    coverUrl: eventData.cov_img_store_url || '', // 新增：封面图
    score: eventData.score,                    // 保留：兼容旧数据
    eventTime: item.eventTime != null ? item.eventTime * 1000 : null,
    createTime: item.createTime
  }
})
```

**改进点：**
- ✅ 使用 `video_store_url` 或 `clip_url`（兼容新旧格式）
- ✅ 使用 `begin_time` 或 `start_time`（兼容新旧格式）
- ✅ 添加 `title`、`summary`、`keyWords`、`coverUrl` 字段
- ✅ 保留 `score` 字段以兼容旧数据

### 2. 表格列更新

**修改前：**
```vue
<a-table-column title="精彩度评分" :width="120">
  <template #cell="{ record }">
    <a-tag :color="getScoreColor(record.score)">
      {{ record.score }}
    </a-tag>
  </template>
</a-table-column>
```

**修改后：**
```vue
<a-table-column title="标题" :width="250">
  <template #cell="{ record }">
    <a-tooltip :content="record.title" v-if="record.title">
      <div style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
        {{ record.title }}
      </div>
    </a-tooltip>
    <span v-else style="color: #999;">-</span>
  </template>
</a-table-column>

<a-table-column title="关键词" :width="200">
  <template #cell="{ record }">
    <a-space wrap v-if="record.keyWords && record.keyWords.length > 0">
      <a-tag v-for="(keyword, index) in record.keyWords.slice(0, 3)" :key="index" size="small">
        {{ keyword }}
      </a-tag>
      <a-tag v-if="record.keyWords.length > 3" size="small">+{{ record.keyWords.length - 3 }}</a-tag>
    </a-space>
    <span v-else style="color: #999;">-</span>
  </template>
</a-table-column>
```

**改进点：**
- ✅ 将"精彩度评分"列替换为"标题"列
- ✅ 添加"关键词"列，最多显示3个关键词
- ✅ 标题过长时显示省略号，鼠标悬停显示完整内容
- ✅ 空值时显示 `-`

### 3. 预览对话框更新

**修改前：**
```vue
<a-descriptions :column="2" bordered>
  <a-descriptions-item label="流名称">
    {{ currentClip?.streamName }}
  </a-descriptions-item>
  <a-descriptions-item label="精彩度评分">
    <a-tag :color="getScoreColor(currentClip?.score)">
      {{ currentClip?.score }}
    </a-tag>
  </a-descriptions-item>
  <!-- ... -->
</a-descriptions>
```

**修改后：**
```vue
<a-descriptions :column="2" bordered>
  <a-descriptions-item label="标题" :span="2" v-if="currentClip?.title">
    {{ currentClip.title }}
  </a-descriptions-item>
  <a-descriptions-item label="摘要" :span="2" v-if="currentClip?.summary">
    {{ currentClip.summary }}
  </a-descriptions-item>
  <a-descriptions-item label="关键词" :span="2" v-if="currentClip?.keyWords && currentClip.keyWords.length > 0">
    <a-space wrap>
      <a-tag v-for="(keyword, index) in currentClip.keyWords" :key="index">
        {{ keyword }}
      </a-tag>
    </a-space>
  </a-descriptions-item>
  <a-descriptions-item label="流名称">
    {{ currentClip?.streamName || '-' }}
  </a-descriptions-item>
  <!-- ... -->
</a-descriptions>
```

**改进点：**
- ✅ 优先显示标题、摘要、关键词
- ✅ 使用 `v-if` 条件渲染，只在有数据时显示
- ✅ 标题和摘要占据整行（`:span="2"`）

### 4. 详情对话框更新

**修改前：**
```vue
<a-descriptions-item label="精彩度评分">
  <a-tag :color="getScoreColor(currentClip.score)">
    {{ currentClip.score }}
  </a-tag>
</a-descriptions-item>
```

**修改后：**
```vue
<a-descriptions-item label="标题" v-if="currentClip.title">
  {{ currentClip.title }}
</a-descriptions-item>
<a-descriptions-item label="摘要" v-if="currentClip.summary">
  {{ currentClip.summary }}
</a-descriptions-item>
<a-descriptions-item label="关键词" v-if="currentClip.keyWords && currentClip.keyWords.length > 0">
  <a-space wrap>
    <a-tag v-for="(keyword, index) in currentClip.keyWords" :key="index">
      {{ keyword }}
    </a-tag>
  </a-space>
</a-descriptions-item>
<a-descriptions-item label="封面图URL" v-if="currentClip.coverUrl">
  <a-link :href="currentClip.coverUrl" target="_blank">
    {{ currentClip.coverUrl }}
  </a-link>
</a-descriptions-item>
```

**改进点：**
- ✅ 添加标题、摘要、关键词、封面图的显示
- ✅ 移除精彩度评分（新数据格式中没有此字段）
- ✅ 使用条件渲染，只在有数据时显示

## 字段映射关系

| 后端字段 | 前端字段 | 说明 | 显示位置 |
|---------|---------|------|---------|
| `eventData.video_store_url` | `clipUrl` | 视频URL | 表格、预览、详情 |
| `eventData.begin_time` | `startTime` | 开始时间 | 表格、预览、详情 |
| `eventData.end_time` | `endTime` | 结束时间 | 表格、预览、详情 |
| `eventData.title` | `title` | 标题 | 表格、预览、详情 |
| `eventData.summary` | `summary` | 摘要 | 预览、详情 |
| `eventData.key_words` | `keyWords` | 关键词数组 | 表格、预览、详情 |
| `eventData.cov_img_store_url` | `coverUrl` | 封面图URL | 详情 |

## 兼容性处理

代码中使用了 `||` 运算符来兼容新旧字段名：

```typescript
// 优先使用新字段，如果不存在则使用旧字段
clipUrl: eventData.video_store_url || eventData.clip_url
startTime: (eventData.begin_time || eventData.start_time) != null
  ? (eventData.begin_time || eventData.start_time) * 1000
  : null
```

这样可以同时支持：
- ✅ 新格式数据（`video_store_url`、`begin_time`）
- ✅ 旧格式数据（`clip_url`、`start_time`）

## 界面效果

### 表格列表

| ID | 流名称 | 域名 | 应用名称 | 标题 | 关键词 | 切片时长 | 生成时间 | 操作 |
|----|--------|------|---------|------|--------|---------|---------|------|
| 28 | - | upload.skyzhou.cn | live | 独家采访：刷新吉尼斯纪录技术揭秘 | 吉尼斯纪录、技术平台、智能悬架 | 0分16秒 | 2026-02-02 18:10:39 | 预览 下载 详情 |
| 27 | - | upload.skyzhou.cn | live | 重要通知：30秒移动绕桩挑战正式启动 | 倒计时、挑战规则、移动绕桩 | 0分19秒 | 2026-02-02 18:09:32 | 预览 下载 详情 |

### 预览对话框

```
┌─────────────────────────────────────────┐
│ 高光切片预览                              │
├─────────────────────────────────────────┤
│ [视频播放器]                              │
│                                          │
│ 标题：独家采访：刷新吉尼斯纪录技术揭秘      │
│                                          │
│ 摘要：官方宣布成功创造新的吉尼斯世界纪录，  │
│ 并独家解读e-zero XS平台智能空气悬架与CDC  │
│ 系统如何保障极限操控。                     │
│                                          │
│ 关键词：[吉尼斯纪录] [技术平台]            │
│         [智能悬架] [独家采访]             │
│                                          │
│ 流名称：-                                 │
│ 切片时长：0分16秒                          │
│ 开始时间：2026-02-02 18:10:39             │
│ 结束时间：2026-02-02 18:10:55             │
└─────────────────────────────────────────┘
```

## 测试验证

### 1. 数据显示测试

**步骤：**
1. 打开高光切片列表页面
2. 查询数据
3. 检查表格中是否显示标题和关键词

**预期结果：**
- ✅ 表格中显示标题列
- ✅ 表格中显示关键词列（最多3个）
- ✅ 标题过长时显示省略号
- ✅ 鼠标悬停显示完整标题

### 2. 预览功能测试

**步骤：**
1. 点击某条记录的"预览"按钮
2. 检查预览对话框中的信息

**预期结果：**
- ✅ 视频正常播放
- ✅ 显示标题、摘要、关键词
- ✅ 显示开始时间、结束时间、时长

### 3. 详情功能测试

**步骤：**
1. 点击某条记录的"详情"按钮
2. 检查详情对话框中的信息

**预期结果：**
- ✅ 显示完整的标题和摘要
- ✅ 显示所有关键词
- ✅ 显示视频URL和封面图URL（如果有）
- ✅ 显示所有时间字段

### 4. 兼容性测试

**步骤：**
1. 使用旧格式数据测试（只有 `clip_url`、`start_time`）
2. 使用新格式数据测试（有 `video_store_url`、`begin_time`）

**预期结果：**
- ✅ 两种格式的数据都能正常显示
- ✅ 新格式优先使用新字段
- ✅ 旧格式回退到旧字段

## 注意事项

1. **空值处理**：所有新字段都使用了 `|| ''` 或 `|| []` 来提供默认值，避免 undefined 错误

2. **条件渲染**：使用 `v-if` 来条件渲染，只在有数据时显示相关字段

3. **关键词显示**：表格中最多显示3个关键词，超过的显示 `+N`

4. **标题省略**：表格中的标题使用 `text-overflow: ellipsis` 来处理过长的文本

5. **兼容性**：代码同时支持新旧两种数据格式

## 后续优化建议

### 1. 添加搜索功能

可以添加按标题或关键词搜索的功能：

```vue
<a-form-item label="标题">
  <a-input
    v-model="searchForm.title"
    placeholder="请输入标题"
    style="width: 200px"
    allow-clear
  />
</a-form-item>

<a-form-item label="关键词">
  <a-input
    v-model="searchForm.keyword"
    placeholder="请输入关键词"
    style="width: 200px"
    allow-clear
  />
</a-form-item>
```

### 2. 添加封面图预览

在表格中添加封面图列：

```vue
<a-table-column title="封面图" :width="100">
  <template #cell="{ record }">
    <img
      v-if="record.coverUrl"
      :src="record.coverUrl"
      style="width: 60px; height: 40px; object-fit: cover; cursor: pointer;"
      @click="handlePreviewCover(record)"
    />
    <span v-else style="color: #999;">-</span>
  </template>
</a-table-column>
```

### 3. 添加批量操作

添加批量下载、批量删除等功能：

```vue
<a-button @click="handleBatchDownload" :disabled="selectedKeys.length === 0">
  批量下载
</a-button>
```

## 相关文件

- `sky-web/src/pages/LiveHighlightClips.vue` - 高光切片列表页面
- `sky-server/api/handler/live_callback_handler.go` - 后端结构体定义
- `sky-server/api/handler/live_callback_handler_ext.go` - 后端处理逻辑

## 更新日期

2026-02-02

---

## HIGHLIGHT_ADD_TIME_COLUMNS

## 需求描述

在直播切片列表页面的表格中增加"开始时间"和"结束时间"两列。

## 实现方案

### 1. 添加列定义

在"应用"列和"时长"列之间添加两个新列：

```vue
<a-table-column title="开始时间" :width="160">
  <template #cell="{ record }">
    {{ formatDateTime(record.startTime) }}
  </template>
</a-table-column>
<a-table-column title="结束时间" :width="160">
  <template #cell="{ record }">
    {{ formatDateTime(record.endTime) }}
  </template>
</a-table-column>
```

### 2. 调整表格滚动宽度

由于增加了两列（每列 160px），需要调整表格的水平滚动宽度：

```vue
<a-table
  :scroll="{ x: 1720 }"
>
```

**计算**：
- 原宽度：1400px
- 新增列：160px × 2 = 320px
- 新宽度：1400px + 320px = 1720px

## 修改的文件

- `sky-web/src/pages/LiveHighlightClips.vue`
  - 第107-121行：添加开始时间和结束时间列
  - 第78行：调整滚动宽度从 1400 到 1720

## 列顺序

修改后的列顺序：

1. ID (60px)
2. 标题 (280px)
3. 关键词 (280px)
4. 域名 (160px)
5. 应用 (80px)
6. **开始时间 (160px)** ← 新增
7. **结束时间 (160px)** ← 新增
8. 时长 (90px)
9. 生成时间 (160px)
10. 操作 (160px, fixed right)

**总宽度**：1720px

## 数据来源

开始时间和结束时间的数据来自：

```typescript
{
  startTime: eventData.begin_time || eventData.start_time,
  endTime: eventData.end_time || eventData.end_time
}
```

这些字段在数据解析时已经处理：
- 从 `eventData` 中提取 `begin_time` 或 `start_time`
- 从 `eventData` 中提取 `end_time`
- 转换为毫秒级时间戳（乘以 1000）

## 时间格式化

使用现有的 `formatDateTime` 函数：

```typescript
const formatDateTime = (timestamp: number | string) => {
  if (timestamp == null || isNaN(Number(timestamp))) return '-'
  const date = dayjs(timestamp)
  if (!date.isValid()) return '-'
  return date.format('YYYY-MM-DD HH:mm:ss')
}
```

**格式**：`YYYY-MM-DD HH:mm:ss`（如：2026-02-02 19:20:17）

## 时长计算

"时长"列继续使用现有的 `formatDuration` 函数：

```typescript
const formatDuration = (startTime: number, endTime: number) => {
  if (startTime == null || endTime == null) return '-'
  const durationMs = endTime - startTime
  if (durationMs <= 0) return '-'

  const seconds = Math.floor(durationMs / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  const remainingMinutes = minutes % 60
  const remainingSeconds = seconds % 60

  if (hours > 0) {
    return `${hours}小时${remainingMinutes}分${remainingSeconds}秒`
  } else if (minutes > 0) {
    return `${minutes}分${remainingSeconds}秒`
  } else {
    return `${seconds}秒`
  }
}
```

## 显示效果

### 示例数据

| 开始时间 | 结束时间 | 时长 |
|---------|---------|------|
| 2026-02-02 19:20:17 | 2026-02-02 19:21:13 | 56秒 |
| 2026-02-02 18:30:00 | 2026-02-02 19:00:00 | 30分0秒 |
| 2026-02-02 17:00:00 | 2026-02-02 19:30:00 | 2小时30分0秒 |

### 空值处理

如果开始时间或结束时间为空，显示 `-`：

| 开始时间 | 结束时间 | 时长 |
|---------|---------|------|
| - | - | - |

## 响应式设计

表格设置了水平滚动：
- 在小屏幕上，用户可以左右滚动查看所有列
- "操作"列固定在右侧（`fixed="right"`）
- 表格使用 `size="small"` 以节省空间

## 与录制列表的对比

### 录制列表的时间列

- 录制时间（startTime）
- 生成时间（eventTime）

### 切片列表的时间列（修改后）

- 开始时间（startTime）
- 结束时间（endTime）
- 时长（计算值）
- 生成时间（eventTime）

**区别**：
- 录制列表只有一个时间点（录制开始时间）
- 切片列表有时间段（开始时间 + 结束时间）

## 测试验证

### 测试步骤

1. 打开直播切片列表页面
2. 查看表格列
3. 验证新增的"开始时间"和"结束时间"列是否显示
4. 验证时间格式是否正确
5. 验证时长计算是否正确

### 预期结果

- ✅ 表格中显示"开始时间"和"结束时间"两列
- ✅ 时间格式为 `YYYY-MM-DD HH:mm:ss`
- ✅ 时长 = 结束时间 - 开始时间
- ✅ 空值显示为 `-`
- ✅ 表格可以水平滚动
- ✅ 操作列固定在右侧

## 相关功能

### 详情对话框

详情对话框中已经包含了开始时间和结束时间：

```vue
<a-descriptions-item label="开始时间">
  {{ formatDateTime(currentClip.startTime) }}
</a-descriptions-item>
<a-descriptions-item label="结束时间">
  {{ formatDateTime(currentClip.endTime) }}
</a-descriptions-item>
<a-descriptions-item label="切片时长">
  {{ formatDuration(currentClip.startTime, currentClip.endTime) }}
</a-descriptions-item>
```

### 预览对话框

预览对话框中也显示了这些信息：

```vue
<a-descriptions-item label="切片时长">
  {{ formatDuration(currentClip?.startTime, currentClip?.endTime) }}
</a-descriptions-item>
<a-descriptions-item label="开始时间">
  {{ formatDateTime(currentClip?.startTime) }}
</a-descriptions-item>
<a-descriptions-item label="结束时间">
  {{ formatDateTime(currentClip?.endTime) }}
</a-descriptions-item>
```

## 总结

通过添加"开始时间"和"结束时间"两列，用户可以在列表中直接查看切片的时间范围，无需打开详情对话框。这提升了用户体验，使信息更加直观。

修改内容：
- ✅ 添加两个新列
- ✅ 调整表格滚动宽度
- ✅ 使用现有的时间格式化函数
- ✅ 保持与详情对话框的一致性

---

## HIGHLIGHT_PREVIEW_AUDIO_FIX

## 问题描述

在直播切片列表页面，点击"预览"按钮打开预览对话框后，关闭对话框时，视频的声音仍然在播放。

## 问题原因

与录制列表相同，当对话框关闭时：
- 对话框的 DOM 元素被隐藏
- `<video>` 元素没有被销毁
- 视频播放状态没有被重置
- 音频继续在后台播放

## 解决方案

### 1. 添加视频元素引用

在 `<video>` 元素上添加 ref 引用：

```vue
<video
  ref="videoRef"
  :src="currentClip.clipUrl"
  controls
  crossorigin="anonymous"
  preload="metadata"
  style="width: 100%; max-height: 500px; background: #000;"
  @error="handleVideoError"
  @loadstart="handleVideoLoadStart"
  @canplay="handleVideoCanPlay"
>
  您的浏览器不支持视频播放
</video>
```

### 2. 在 script 中声明 ref

```typescript
// 视频元素引用
const videoRef = ref<HTMLVideoElement | null>(null)
```

### 3. 监听对话框关闭事件

在 `<a-modal>` 上添加关闭事件监听：

```vue
<a-modal
  v-model:visible="previewVisible"
  title="高光切片预览"
  :width="800"
  :footer="false"
  @close="handlePreviewClose"
  @cancel="handlePreviewClose"
>
```

### 4. 实现关闭处理函数

```typescript
// 关闭预览对话框
const handlePreviewClose = () => {
  // 暂停视频播放
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

## 修改的文件

- `sky-web/src/pages/LiveHighlightClips.vue`
  - 第133-157行：添加 ref 引用和关闭事件监听
  - 第298行：声明视频元素 ref
  - 第407-416行：实现关闭处理函数

## 与录制列表的区别

### 相同点
- 都是视频播放器关闭后音频持续播放的问题
- 解决方案相同：添加 ref 引用和关闭事件处理

### 不同点
- **录制列表**：支持视频和音频两种格式（video 和 audio 元素）
- **切片列表**：只有视频格式（只有 video 元素）

因此切片列表的修复更简单，只需要处理 video 元素。

## 测试验证

### 测试步骤

1. 打开直播切片列表页面
2. 点击任意一条记录的"预览"按钮
3. 等待视频开始播放
4. 关闭预览对话框（点击 X 按钮、取消按钮或遮罩层）
5. 验证音频是否停止播放

### 预期结果

- ✅ 关闭对话框后，视频立即停止播放
- ✅ 再次打开预览时，从头开始播放
- ✅ 不会有任何残留的声音

## 已修复的页面

现在两个页面都已修复：

1. ✅ **直播录制列表** (`LiveRecordings.vue`)
   - 支持视频和音频格式
   - 处理 video 和 audio 两种元素

2. ✅ **直播切片列表** (`LiveHighlightClips.vue`)
   - 只支持视频格式
   - 处理 video 元素

## 代码对比

### 录制列表（支持视频和音频）

```typescript
const videoRef = ref<HTMLVideoElement | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)

const handlePreviewClose = () => {
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  if (audioRef.value) {
    audioRef.value.pause()
    audioRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

### 切片列表（只支持视频）

```typescript
const videoRef = ref<HTMLVideoElement | null>(null)

const handlePreviewClose = () => {
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

## 扩展功能

切片列表已经实现了一些额外的视频处理功能：

### 1. 视频加载状态处理

```typescript
const handleVideoLoadStart = () => {
  console.log('视频开始加载')
}

const handleVideoCanPlay = () => {
  console.log('视频可以播放')
}
```

### 2. 视频错误处理

```typescript
const handleVideoError = (event: Event) => {
  const video = event.target as HTMLVideoElement
  let errorMsg = '视频加载失败'
  if (video.error) {
    switch (video.error.code) {
      case 1: // MEDIA_ERR_ABORTED
        errorMsg = '视频加载被中止'
        break
      case 2: // MEDIA_ERR_NETWORK
        errorMsg = '网络错误，无法加载视频'
        break
      case 3: // MEDIA_ERR_DECODE
        errorMsg = '视频解码失败'
        break
      case 4: // MEDIA_ERR_SRC_NOT_SUPPORTED
        errorMsg = '视频格式不支持或视频URL无效'
        break
    }
  }
  console.error('视频播放错误:', errorMsg, video.error)
  Message.error(errorMsg)
}
```

### 3. CORS 支持

```vue
<video
  crossorigin="anonymous"
  preload="metadata"
>
```

这些功能在录制列表中也可以考虑添加。

## 相关文档

- 录制列表修复文档：`live_callback_RECORDING_PREVIEW_AUDIO_FIX.md`
- 本文档：切片列表修复

## 总结

通过添加视频元素引用和关闭事件处理，我们成功解决了直播切片列表预览对话框关闭后音频持续播放的问题。

现在两个列表页面都能正确处理视频播放：
- ✅ 关闭对话框时立即停止播放
- ✅ 重置播放进度
- ✅ 释放媒体资源
- ✅ 提供良好的用户体验

---

## HIGHLIGHT_RAW_BODY

## 更新内容

优化了高光切片（Highlight）事件的处理逻辑，在保存事件数据时同时保留原始 HTTP 请求体和客户端 IP 地址。

## 修改文件

- `sky-server/api/handler/live_callback_handler_ext.go`

## 修改详情

### 1. 添加必要的导入

**修改前：**
```go
import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)
```

**修改后：**
```go
import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)
```

**说明：**
- 添加 `bytes` 包：用于创建字节缓冲区
- 添加 `io` 包：用于读取和恢复请求体

### 2. 读取并保留原始请求体

**修改前：**
```go
func (h *LiveCallbackHandler) HandleHighlight(c *gin.Context) {
	var event HighlightEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析高光切片事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}
	// ...
}
```

**修改后：**
```go
func (h *LiveCallbackHandler) HandleHighlight(c *gin.Context) {
	// 读取原始请求体
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("读取高光切片事件请求体失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "failed to read request body"})
		return
	}
	// 恢复请求体，以便后续的 ShouldBindJSON 可以读取
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var event HighlightEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析高光切片事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}
	// ...
}
```

**说明：**
1. 使用 `io.ReadAll()` 读取原始请求体
2. 使用 `io.NopCloser()` 和 `bytes.NewBuffer()` 恢复请求体
3. 这样既保留了原始数据，又不影响后续的 JSON 解析

### 3. 保存客户端 IP 地址

**修改前：**
```go
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "highlight",
	EventTime:    event.EventTime,
	DomainName:   event.PushDomain,
	AppName:      event.AppName,
	StreamName:   event.StreamName,
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

**修改后：**
```go
// 保存原始请求体和解析后的事件数据
eventData, _ := json.Marshal(event)
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "highlight",
	EventTime:    event.EventTime,
	DomainName:   event.PushDomain,
	AppName:      event.AppName,
	StreamName:   event.StreamName,
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	ClientIP:     c.ClientIP(),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

**说明：**
- 添加 `ClientIP` 字段，使用 `c.ClientIP()` 获取客户端 IP 地址
- 添加注释说明保存的是原始请求体和解析后的事件数据

## 功能说明

### 1. 原始请求体保留

**为什么需要保留原始请求体？**

- **调试和排查**：当事件处理出现问题时，可以查看原始请求内容
- **审计追踪**：记录完整的请求历史，用于审计和合规
- **数据重放**：可以使用原始请求体重新处理事件
- **数据对比**：对比原始数据和解析后的数据，发现潜在问题

**实现方式：**

```go
// 1. 读取原始请求体
rawBody, err := io.ReadAll(c.Request.Body)

// 2. 恢复请求体（因为 Body 是一次性读取的流）
c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

// 3. 正常解析 JSON
var event HighlightEvent
c.ShouldBindJSON(&event)
```

**注意事项：**
- `c.Request.Body` 是一个 `io.ReadCloser`，只能读取一次
- 读取后必须恢复，否则后续的 `ShouldBindJSON` 会失败
- `io.NopCloser` 创建一个不需要关闭的 ReadCloser

### 2. 客户端 IP 地址记录

**为什么需要记录客户端 IP？**

- **安全审计**：追踪请求来源，识别异常访问
- **问题定位**：当某个地区或网络出现问题时，可以快速定位
- **统计分析**：分析不同地区的事件分布

**实现方式：**

```go
ClientIP: c.ClientIP()
```

**`c.ClientIP()` 的工作原理：**
1. 优先从 `X-Forwarded-For` 头获取（如果有代理）
2. 其次从 `X-Real-IP` 头获取
3. 最后从 `RemoteAddr` 获取

## 数据库字段

确保 `live_callback_event` 表包含以下字段：

```sql
CREATE TABLE `live_callback_event` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `event_type` varchar(50) NOT NULL COMMENT '事件类型',
  `event_time` bigint DEFAULT NULL COMMENT '事件时间戳',
  `domain_name` varchar(255) DEFAULT NULL COMMENT '域名',
  `app_name` varchar(255) DEFAULT NULL COMMENT '应用名称',
  `stream_name` varchar(255) DEFAULT NULL COMMENT '流名称',
  `stream_id` varchar(255) DEFAULT NULL COMMENT '流ID',
  `event_data` text COMMENT '事件数据JSON',
  `client_ip` varchar(50) DEFAULT NULL COMMENT '客户端IP地址',
  `sign` varchar(255) DEFAULT NULL COMMENT '签名',
  `t_value` bigint DEFAULT NULL COMMENT 'T值',
  `sys_company_id` bigint NOT NULL COMMENT '公司ID',
  `is_active` char(1) DEFAULT 'Y' COMMENT '是否有效',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_stream_id` (`stream_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播回调事件表';
```

## 使用示例

### 查询高光切片事件

```sql
-- 查询所有高光切片事件
SELECT
  id,
  stream_id,
  stream_name,
  event_data,
  client_ip,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
ORDER BY create_time DESC;

-- 查询特定流的高光切片
SELECT
  id,
  JSON_EXTRACT(event_data, '$.clip_url') as clip_url,
  JSON_EXTRACT(event_data, '$.score') as score,
  JSON_EXTRACT(event_data, '$.start_time') as start_time,
  JSON_EXTRACT(event_data, '$.end_time') as end_time,
  client_ip,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
  AND stream_id = '0202'
ORDER BY create_time DESC;

-- 统计不同IP的请求数量
SELECT
  client_ip,
  COUNT(*) as request_count
FROM live_callback_event
WHERE event_type = 'highlight'
GROUP BY client_ip
ORDER BY request_count DESC;
```

### API 查询示例

```bash
# 查询高光切片列表
curl "http://localhost:3000/api/v1/live/callback/events?eventType=highlight&pageNum=1&pageSize=20"

# 查询特定流的高光切片
curl "http://localhost:3000/api/v1/live/callback/events?eventType=highlight&streamId=0202&pageNum=1&pageSize=20"

# 查询特定时间范围的高光切片
curl "http://localhost:3000/api/v1/live/callback/events?eventType=highlight&startTime=2026-01-26%2000:00:00&endTime=2026-02-02%2023:59:59&pageNum=1&pageSize=20"
```

## 测试验证

### 1. 发送测试请求

```bash
curl -X POST http://localhost:3000/api/v1/live/callback/highlight \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 349,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": 1738483200,
    "sign": "test_sign",
    "event_time": 1738483200,
    "clip_url": "http://example.com/clip.mp4",
    "start_time": 1738483100,
    "end_time": 1738483200,
    "score": 8.5,
    "stream_param": "test_param",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

### 2. 验证数据保存

```sql
-- 查询最新的高光切片事件
SELECT
  id,
  stream_id,
  event_data,
  client_ip,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
ORDER BY id DESC
LIMIT 1;
```

**预期结果：**
- `event_data` 字段包含完整的 JSON 数据
- `client_ip` 字段包含请求的客户端 IP 地址
- 所有字段都正确保存

### 3. 验证前端显示

1. 打开高光切片列表页面：`http://localhost:8080/#/live/highlight-clips`
2. 查询数据
3. 验证表格中显示的数据是否正确

## 后续优化建议

### 1. 添加原始请求体字段

如果需要保存完整的原始请求体（而不仅仅是解析后的 JSON），可以在数据库中添加 `raw_body` 字段：

```sql
ALTER TABLE live_callback_event
ADD COLUMN raw_body TEXT COMMENT '原始请求体' AFTER event_data;
```

然后在代码中保存：

```go
callbackEvent := &entity.LiveCallbackEvent{
	// ...
	EventData:    string(eventData),
	RawBody:      string(rawBody),  // 保存原始请求体
	ClientIP:     c.ClientIP(),
	// ...
}
```

### 2. 添加请求头记录

如果需要记录请求头信息（如 User-Agent、Referer 等），可以添加 `request_headers` 字段：

```sql
ALTER TABLE live_callback_event
ADD COLUMN request_headers TEXT COMMENT '请求头JSON' AFTER raw_body;
```

然后在代码中保存：

```go
// 获取请求头
headers := make(map[string]string)
for key, values := range c.Request.Header {
	if len(values) > 0 {
		headers[key] = values[0]
	}
}
headersJSON, _ := json.Marshal(headers)

callbackEvent := &entity.LiveCallbackEvent{
	// ...
	RequestHeaders: string(headersJSON),
	// ...
}
```

### 3. 添加响应时间记录

记录请求处理时间，用于性能分析：

```go
startTime := time.Now()

// ... 处理逻辑 ...

processingTime := time.Since(startTime).Milliseconds()

callbackEvent := &entity.LiveCallbackEvent{
	// ...
	ProcessingTime: processingTime,
	// ...
}
```

## 相关文件

- `sky-server/api/handler/live_callback_handler_ext.go` - 高光切片事件处理器
- `sky-server/internal/model/entity/live_callback_event.go` - 回调事件实体
- `sky-web/src/pages/LiveHighlightClips.vue` - 高光切片列表页面

## 更新日期

2026-02-02

---

## HIGHLIGHT_STRUCTURE_FIX

## 问题描述

原有的 `HighlightEvent` 结构体定义与腾讯云实际返回的数据格式不匹配，导致：
1. 无法正确解析请求数据
2. 所有字段都是空值
3. 数据无法保存到数据库

## 实际数据格式

腾讯云实际返回的高光切片事件格式：

```json
{
  "appid": 1301212747,
  "domain": "upload.skyzhou.cn",
  "event_type": 349,
  "items": [
    {
      "begin_time": 1770026056,
      "cov_img_store_url": "",
      "end_time": 1770026081,
      "key_words": ["全球标准", "技术规范", "新法规", "高性能"],
      "summary": "官方解读Hima高性能、多功能、全地形能力的技术规范，标志全球皮卡行业新标准。",
      "title": "新法规解读：Hima技术规范与全球标准",
      "video_store_url": "http://video-1301212747.cos.ap-nanjing.myqcloud.com/..."
    }
  ],
  "path": "live",
  "stream_id": "0202"
}
```

**关键特点：**
1. 包含 `items` 数组，一个请求可能包含多个高光切片
2. 使用 `begin_time` 和 `end_time` 而不是 `start_time` 和 `end_time`
3. 包含 `title`、`summary`、`key_words` 等 AI 生成的内容
4. 使用 `video_store_url` 而不是 `clip_url`
5. 包含 `appid`、`domain`、`path` 等顶层字段

## 修复方案

### 1. 更新结构体定义

**文件：** `sky-server/api/handler/live_callback_handler.go`

**修改前：**
```go
// HighlightEvent 高光切片事件通知
type HighlightEvent struct {
	EventType   int      `json:"event_type"`   // 事件类型：325-高光切片
	StreamID    string   `json:"stream_id"`    // 流ID
	ChannelID   string   `json:"channel_id"`   // 频道ID
	T           int64    `json:"t"`            // 过期时间
	Sign        string   `json:"sign"`         // 签名
	EventTime   int64    `json:"event_time"`   // 事件时间
	ClipURL     string   `json:"clip_url"`     // 切片URL
	StartTime   int64    `json:"start_time"`   // 开始时间
	EndTime     int64    `json:"end_time"`     // 结束时间
	Score       float64  `json:"score"`        // 精彩度评分
	StreamParam string   `json:"stream_param"` // 推流参数
	PushDomain  string   `json:"push_domain"`  // 推流域名
	AppName     string   `json:"app_name"`     // 应用名称
	StreamName  string   `json:"stream_name"`  // 流名称
}
```

**修改后：**
```go
// HighlightEvent 高光切片事件通知（实际格式）
type HighlightEvent struct {
	AppID     int64           `json:"appid"`      // 应用ID
	Domain    string          `json:"domain"`     // 域名
	EventType int             `json:"event_type"` // 事件类型：349-高光切片
	Items     []HighlightItem `json:"items"`      // 高光切片列表
	Path      string          `json:"path"`       // 路径（如：live）
	StreamID  string          `json:"stream_id"`  // 流ID
	T         int64           `json:"t"`          // 过期时间（可选）
	Sign      string          `json:"sign"`       // 签名（可选）
}

// HighlightItem 单个高光切片信息
type HighlightItem struct {
	BeginTime       int64    `json:"begin_time"`         // 开始时间（Unix时间戳）
	EndTime         int64    `json:"end_time"`           // 结束时间（Unix时间戳）
	CovImgStoreURL  string   `json:"cov_img_store_url"`  // 封面图片存储URL
	VideoStoreURL   string   `json:"video_store_url"`    // 视频存储URL
	Title           string   `json:"title"`              // 标题
	Summary         string   `json:"summary"`            // 摘要
	KeyWords        []string `json:"key_words"`          // 关键词列表
}
```

### 2. 更新处理逻辑

**文件：** `sky-server/api/handler/live_callback_handler_ext.go`

**关键改动：**

1. **循环处理多个切片**：
```go
// 为每个高光切片创建一条记录
for i, item := range event.Items {
	// 处理每个切片...
}
```

2. **构建兼容的数据结构**：
```go
itemData := map[string]interface{}{
	"event_type":        event.EventType,
	"stream_id":         event.StreamID,
	"appid":             event.AppID,
	"domain":            event.Domain,
	"path":              event.Path,
	"begin_time":        item.BeginTime,
	"end_time":          item.EndTime,
	"video_store_url":   item.VideoStoreURL,
	"cov_img_store_url": item.CovImgStoreURL,
	"title":             item.Title,
	"summary":           item.Summary,
	"key_words":         item.KeyWords,
	// 兼容旧字段名
	"clip_url":          item.VideoStoreURL,
	"start_time":        item.BeginTime,
	"end_time":          item.EndTime,
}
```

3. **保存到数据库**：
```go
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "highlight",
	EventTime:    item.BeginTime,
	DomainName:   event.Domain,
	AppName:      event.Path,
	StreamName:   "",
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	ClientIP:     c.ClientIP(),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

## 数据映射关系

| 腾讯云字段 | 数据库字段 | 说明 |
|-----------|-----------|------|
| `appid` | `event_data.appid` | 应用ID |
| `domain` | `domain_name` | 域名 |
| `event_type` | `event_data.event_type` | 事件类型（349） |
| `path` | `app_name` | 路径（如：live） |
| `stream_id` | `stream_id` | 流ID |
| `items[].begin_time` | `event_time` | 开始时间 |
| `items[].end_time` | `event_data.end_time` | 结束时间 |
| `items[].video_store_url` | `event_data.video_store_url` | 视频URL |
| `items[].title` | `event_data.title` | 标题 |
| `items[].summary` | `event_data.summary` | 摘要 |
| `items[].key_words` | `event_data.key_words` | 关键词 |

## 兼容性处理

为了保持前端兼容性，在 `event_data` 中同时保存新旧字段名：

```json
{
  "begin_time": 1770026056,
  "end_time": 1770026081,
  "video_store_url": "http://...",
  "start_time": 1770026056,    // 兼容旧字段名
  "end_time": 1770026081,      // 兼容旧字段名
  "clip_url": "http://..."     // 兼容旧字段名
}
```

这样前端可以使用 `start_time` 或 `begin_time`，`clip_url` 或 `video_store_url`。

## 前端更新建议

虽然后端已经做了兼容处理，但建议前端也更新字段名以匹配实际格式：

**文件：** `sky-web/src/pages/LiveHighlightClips.vue`

```typescript
// 当前使用（兼容）
startTime: eventData.start_time != null ? eventData.start_time * 1000 : null
clipUrl: eventData.clip_url

// 建议更新为
startTime: eventData.begin_time != null ? eventData.begin_time * 1000 : null
clipUrl: eventData.video_store_url

// 或者使用兼容写法
startTime: (eventData.begin_time || eventData.start_time) != null
  ? (eventData.begin_time || eventData.start_time) * 1000
  : null
clipUrl: eventData.video_store_url || eventData.clip_url
```

## 新增字段展示

前端可以展示更多 AI 生成的内容：

1. **标题（title）**：高光切片的标题
2. **摘要（summary）**：内容摘要
3. **关键词（key_words）**：关键词列表
4. **封面图（cov_img_store_url）**：封面图片URL

**示例：**
```vue
<template>
  <a-table-column title="标题" :width="200">
    <template #cell="{ record }">
      {{ record.title || '-' }}
    </template>
  </a-table-column>

  <a-table-column title="摘要" :width="300">
    <template #cell="{ record }">
      <a-tooltip :content="record.summary">
        <div class="text-ellipsis">{{ record.summary || '-' }}</div>
      </a-tooltip>
    </template>
  </a-table-column>

  <a-table-column title="关键词" :width="200">
    <template #cell="{ record }">
      <a-space wrap>
        <a-tag v-for="keyword in record.keyWords" :key="keyword">
          {{ keyword }}
        </a-tag>
      </a-space>
    </template>
  </a-table-column>
</template>

<script setup lang="ts">
const tableData = ref([])

const loadData = async () => {
  // ...
  tableData.value = res.data.list.map((item: any) => {
    const eventData = JSON.parse(item.eventData)
    return {
      id: item.id,
      streamId: item.streamId,
      streamName: item.streamName,
      domainName: item.domainName,
      appName: item.appName,
      clipUrl: eventData.video_store_url || eventData.clip_url,
      startTime: (eventData.begin_time || eventData.start_time) != null
        ? (eventData.begin_time || eventData.start_time) * 1000
        : null,
      endTime: (eventData.end_time) != null
        ? eventData.end_time * 1000
        : null,
      title: eventData.title,           // 新增
      summary: eventData.summary,       // 新增
      keyWords: eventData.key_words,    // 新增
      coverUrl: eventData.cov_img_store_url, // 新增
      eventTime: item.eventTime != null ? item.eventTime * 1000 : null,
      createTime: item.createTime
    }
  })
}
</script>
```

## 测试验证

### 1. 发送测试请求

使用您提供的实际数据格式：

```bash
curl -X POST http://localhost:3000/api/v1/live/callback/highlight \
  -H "Content-Type: application/json" \
  -d '{
    "appid": 1301212747,
    "domain": "upload.skyzhou.cn",
    "event_type": 349,
    "items": [
      {
        "begin_time": 1770026056,
        "cov_img_store_url": "",
        "end_time": 1770026081,
        "key_words": ["全球标准", "技术规范", "新法规", "高性能"],
        "summary": "官方解读Hima高性能、多功能、全地形能力的技术规范，标志全球皮卡行业新标准。",
        "title": "新法规解读：Hima技术规范与全球标准",
        "video_store_url": "http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/0202/LIVE_0228577F359F9989378520BF9280AA1274-1770025101592/2026-02-02_17-59-14_954866.mp4"
      }
    ],
    "path": "live",
    "stream_id": "0202"
  }'
```

### 2. 验证数据保存

```sql
SELECT
  id,
  stream_id,
  domain_name,
  app_name,
  event_time,
  JSON_EXTRACT(event_data, '$.title') as title,
  JSON_EXTRACT(event_data, '$.summary') as summary,
  JSON_EXTRACT(event_data, '$.video_store_url') as video_url,
  JSON_EXTRACT(event_data, '$.key_words') as keywords,
  create_time
FROM live_callback_event
WHERE event_type = 'highlight'
ORDER BY id DESC
LIMIT 1;
```

**预期结果：**
- `stream_id` = "0202"
- `domain_name` = "upload.skyzhou.cn"
- `app_name` = "live"
- `event_time` = 1770026056
- `title` = "新法规解读：Hima技术规范与全球标准"
- `summary` = "官方解读Hima高性能..."
- `video_url` = "http://video-1301212747.cos..."
- `keywords` = ["全球标准", "技术规范", "新法规", "高性能"]

### 3. 验证前端显示

1. 打开高光切片列表页面
2. 查询数据
3. 验证表格中显示的数据是否正确
4. 检查时间、URL、标题等字段

## 日志输出

修改后的日志会显示更详细的信息：

```
INFO  收到高光切片事件  streamId=0202  domain=upload.skyzhou.cn  itemCount=1
INFO  保存高光切片成功  itemIndex=0  title=新法规解读：Hima技术规范与全球标准  videoUrl=http://video-1301212747.cos...
```

## 注意事项

1. **多个切片处理**：一个请求可能包含多个高光切片，每个切片会创建一条独立的数据库记录

2. **时间戳单位**：`begin_time` 和 `end_time` 是 Unix 时间戳（秒），前端需要乘以 1000 转换为毫秒

3. **签名验证**：签名字段是可选的，只有在提供了签名时才进行验证

4. **字段兼容性**：`event_data` 中同时保存新旧字段名，确保前端兼容

5. **空值处理**：使用 `!= null` 检查，避免将 `0` 当作空值过滤掉

## 相关文件

- `sky-server/api/handler/live_callback_handler.go` - 结构体定义
- `sky-server/api/handler/live_callback_handler_ext.go` - 处理逻辑
- `sky-web/src/pages/LiveHighlightClips.vue` - 前端页面

## 更新日期

2026-02-02

---

## INVALID_DATE_FIX

## 问题描述

直播录制列表和高光切片列表页面存在以下问题：

1. **统计数据显示 "Invalid Date"**
   - "当前页文件大小" 显示 "Invalid Date"
   - "当前页总时长" 显示 "Invalid Date"

2. **表格中时间字段显示 "Invalid Date"**
   - 当数据为空时显示 "Invalid Date"

3. **有数据但不显示**
   - 后端返回了数据，但前端表格为空
   - 时间戳为 `0` 的数据被过滤掉

## 问题原因

### 原因一：时间戳处理不当

当后端返回的时间戳字段（`start_time`、`end_time`、`eventTime`）为 `null` 或 `undefined` 时：

```typescript
// 问题代码
startTime: eventData.start_time * 1000  // undefined * 1000 = NaN
```

- `undefined * 1000` 会得到 `NaN`
- `dayjs(NaN)` 会创建一个无效的日期对象
- 格式化后显示为 "Invalid Date"

### 原因二：a-statistic 组件的 value 属性

`a-statistic` 组件的 `:value` 属性接收格式化后的字符串（如 "0 B" 或 "0秒"）时，组件内部可能尝试将其解析为日期或数字，导致显示 "Invalid Date"。

```vue
<!-- 问题代码 -->
<a-statistic
  title="当前页文件大小"
  :value="formatFileSize(stats.totalSize)"
/>
```

### 原因三：错误的 falsy 值检查

**最关键的问题**：代码使用了 `!value` 或 `value ? ... : null` 来检查值是否存在，但这会将 `0` 也当作 falsy 值过滤掉。

```typescript
// 问题代码
if (!timestamp) return '-'  // 0 会被当作 false
startTime: eventData.start_time ? eventData.start_time * 1000 : null  // 0 会返回 null
```

在 JavaScript 中，以下值都是 falsy：
- `false`
- `0`
- `''` (空字符串)
- `null`
- `undefined`
- `NaN`

但 `0` 是一个有效的时间戳（1970-01-01 00:00:00 UTC），应该被保留！

**实际案例**：
```json
{
  "eventTime": 0,
  "start_time": 0,
  "end_time": 0,
  "score": 0
}
```

这些 `0` 值都是有效数据，但被错误地过滤掉了。

## 修复方案

### 修复一：使用严格的 null 检查

**修复前：**
```typescript
// 错误：将 0 当作 falsy 值
if (!timestamp) return '-'
startTime: eventData.start_time ? eventData.start_time * 1000 : null
```

**修复后：**
```typescript
// 正确：只检查 null 和 undefined
if (timestamp == null) return '-'
startTime: eventData.start_time != null ? eventData.start_time * 1000 : null
```

**关键点：**
- `== null` 会同时匹配 `null` 和 `undefined`，但不会匹配 `0`
- `!= null` 会保留 `0`、`false`、`''` 等有效值

### 修复二：增强 formatDateTime 函数

**修复前：**
```typescript
const formatDateTime = (timestamp: number | string) => {
  if (!timestamp) return '-'  // 问题：0 会返回 '-'
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}
```

**修复后：**
```typescript
const formatDateTime = (timestamp: number | string) => {
  if (timestamp == null || isNaN(Number(timestamp))) return '-'
  const date = dayjs(timestamp)
  if (!date.isValid()) return '-'
  return date.format('YYYY-MM-DD HH:mm:ss')
}
```

**改进点：**
1. 使用 `== null` 代替 `!timestamp`，保留 `0` 值
2. 增加 `isNaN()` 检查，过滤掉 `NaN` 值
3. 增加 `date.isValid()` 检查，确保日期对象有效

### 修复三：安全的时间戳转换

**修复前：**
```typescript
return {
  startTime: eventData.start_time ? eventData.start_time * 1000 : null,
  endTime: eventData.end_time ? eventData.end_time * 1000 : null,
  eventTime: item.eventTime ? item.eventTime * 1000 : null,
}
```

**修复后：**
```typescript
return {
  startTime: eventData.start_time != null ? eventData.start_time * 1000 : null,
  endTime: eventData.end_time != null ? eventData.end_time * 1000 : null,
  eventTime: item.eventTime != null ? item.eventTime * 1000 : null,
}
```

**改进点：**
- 使用 `!= null` 代替 `? ... :`，保留 `0` 值
- `0 != null` 返回 `true`，所以 `0 * 1000 = 0` 会被保留

### 修复四：formatDuration 函数

**修复前：**
```typescript
const formatDuration = (seconds: number) => {
  if (!seconds) return '-'  // 问题：0 会返回 '-'
  // ...
}
```

**修复后：**
```typescript
const formatDuration = (seconds: number) => {
  if (seconds == null) return '-'
  if (seconds === 0) return '0秒'  // 明确处理 0 的情况
  // ...
}
```

### 修复五：使用 a-statistic 的 value 插槽

**修复前：**
```vue
<a-statistic
  title="当前页文件大小"
  :value="formatFileSize(stats.totalSize)"
>
  <template #prefix>
    <icon-storage />
  </template>
</a-statistic>
```

**修复后：**
```vue
<a-statistic title="当前页文件大小" :value="0">
  <template #prefix>
    <icon-storage />
  </template>
  <template #value>
    {{ formatFileSize(stats.totalSize) }}
  </template>
</a-statistic>
```

**改进点：**
1. `:value` 设置为数字 `0`（组件需要的默认值）
2. 使用 `#value` 插槽自定义显示内容
3. 避免组件尝试解析格式化后的字符串

## 修复文件

### 1. LiveRecordings.vue

**修复内容：**
- ✅ 增强 `formatDateTime` 函数（使用 `== null`）
- ✅ 增强 `formatDuration` 函数（明确处理 `0` 值）
- ✅ 修复时间戳转换逻辑（使用 `!= null`）
- ✅ 修复 "当前页文件大小" 统计卡片
- ✅ 修复 "当前页总时长" 统计卡片

**关键代码位置：**
- `formatDateTime` 函数：第 555-561 行
- `formatDuration` 函数：第 572-587 行
- 数据解析：第 434-455 行
- 统计卡片：第 38-57 行

### 2. LiveHighlightClips.vue

**修复内容：**
- ✅ 增强 `formatDateTime` 函数（使用 `== null`）
- ✅ 增强 `formatDuration` 函数（使用 `== null`）
- ✅ 修复时间戳转换逻辑（使用 `!= null`）

**关键代码位置：**
- `formatDateTime` 函数：第 351-357 行
- `formatDuration` 函数：第 359-366 行
- 数据解析：第 275-290 行

## JavaScript 中的 falsy 值陷阱

### Falsy 值列表

在 JavaScript 中，以下值在布尔上下文中会被转换为 `false`：

```javascript
false      // 布尔值 false
0          // 数字 0
-0         // 负零
0n         // BigInt 零
''         // 空字符串
null       // null
undefined  // undefined
NaN        // Not a Number
```

### 常见错误

```javascript
// ❌ 错误：会过滤掉 0
if (!value) {
  return '-'
}

// ❌ 错误：0 会返回 null
const result = value ? value * 1000 : null

// ✅ 正确：只过滤 null 和 undefined
if (value == null) {
  return '-'
}

// ✅ 正确：保留 0
const result = value != null ? value * 1000 : null
```

### == null vs === null

```javascript
// == null 会匹配 null 和 undefined
null == null       // true
undefined == null  // true
0 == null          // false
'' == null         // false

// === null 只匹配 null
null === null      // true
undefined === null // false
```

**推荐使用 `== null` 和 `!= null`**，因为它们可以同时检查 `null` 和 `undefined`，而不会误判 `0` 等有效值。

## 测试验证

### 1. 时间戳为 0 的数据测试

**测试数据：**
```json
{
  "eventTime": 0,
  "start_time": 0,
  "end_time": 0
}
```

**预期结果：**
- 时间显示为 "1970-01-01 08:00:00"（北京时间）
- 数据正常显示在表格中
- 不会被过滤掉

### 2. 空数据测试

**步骤：**
1. 打开直播录制列表页面
2. 设置一个没有数据的时间范围
3. 点击"查询"按钮

**预期结果：**
- 统计卡片显示 "0 B" 和 "0秒"，而不是 "Invalid Date"
- 表格显示"暂无录制数据"

### 3. 部分数据缺失测试

**测试数据：**
```json
{
  "eventTime": null,
  "start_time": 1234567890,
  "end_time": null
}
```

**预期结果：**
- 有效的时间戳正常显示
- 缺失的时间戳显示为 "-"
- 不会出现 "Invalid Date"

### 4. 正常数据测试

**步骤：**
1. 查询有完整数据的录制记录
2. 检查所有统计数据和表格字段

**预期结果：**
- 所有统计数据正确显示
- 文件大小显示为 "XX MB" 格式
- 时长显示为 "XX小时XX分" 格式
- 时间显示为正确的日期时间格式

## 技术要点

### 1. Falsy 值检查的最佳实践

```typescript
// ❌ 不推荐：会过滤掉 0、false、''
if (!value) { }

// ✅ 推荐：只过滤 null 和 undefined
if (value == null) { }

// ✅ 推荐：明确检查类型
if (typeof value === 'undefined') { }
if (value === null) { }

// ✅ 推荐：明确处理 0
if (value == null) return '-'
if (value === 0) return '0秒'
```

### 2. 时间戳 0 的含义

```typescript
// 0 是一个有效的 Unix 时间戳
new Date(0)  // 1970-01-01 00:00:00 UTC
dayjs(0)     // 1970-01-01 08:00:00 (北京时间)

// 在某些场景下，0 可能表示：
// 1. 真实的 1970-01-01 时间
// 2. 默认值或初始值
// 3. 测试数据

// 无论哪种情况，都应该保留并正确显示
```

### 3. NaN 的特殊性

```typescript
// NaN 不是 falsy 值
!NaN  // false
NaN == false  // false
NaN == true   // false

// 必须使用 isNaN() 检查
isNaN(NaN)  // true
isNaN(undefined * 1000)  // true
isNaN(0)    // false
```

### 4. dayjs 的日期验证

```typescript
// dayjs 可以接受无效值，但会创建无效日期对象
const invalidDate = dayjs(NaN)
invalidDate.format('YYYY-MM-DD')  // "Invalid Date"

// 应该使用 isValid() 检查
invalidDate.isValid()  // false

// 0 是有效的时间戳
const validDate = dayjs(0)
validDate.isValid()  // true
validDate.format('YYYY-MM-DD HH:mm:ss')  // "1970-01-01 08:00:00"
```

### 5. a-statistic 组件的 value 属性

```typescript
// a-statistic 的 :value 属性期望接收数字
// 如果传入字符串，组件可能尝试解析它
:value="123"           // ✅ 正确
:value="'123'"         // ⚠️ 可能被解析
:value="'0 B'"         // ❌ 可能显示 "Invalid Date"

// 解决方案：使用 #value 插槽
<template #value>
  {{ formatFileSize(size) }}
</template>
```

## 相关文件

- `sky-web/src/pages/LiveRecordings.vue` - 直播录制列表页面
- `sky-web/src/pages/LiveHighlightClips.vue` - 直播高光切片列表页面

## 更新日期

2026-02-02

---

## MENU_SQL

## 前端代码已更新

已在 `BasicLayout.vue` 中添加了以下支持：

1. ✅ 导入 `LiveHighlightClips` 和 `LiveRecordings` 组件
2. ✅ 在 `componentRegistry` 中注册组件
3. ✅ 在 `loadComponent` 函数中添加路由处理

## 数据库菜单配置

### 方式一：添加到现有直播管理菜单下

假设您已经有一个"直播管理"的父菜单，执行以下SQL：

```sql
-- 1. 查找直播管理的父菜单ID（根据实际情况调整）
SELECT ID, NAME, DISPLAY_NAME FROM sys_subsystem WHERE NAME LIKE '%live%' OR DISPLAY_NAME LIKE '%直播%';

-- 假设直播管理的ID是 100（请根据实际查询结果替换）

-- 2. 添加高光切片菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_highlight_clips',
  '高光切片',
  100,  -- 替换为实际的直播管理菜单ID
  '/live/highlight-clips',
  'icon-video',
  10,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);

-- 3. 添加录制列表菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_recordings',
  '录制列表',
  100,  -- 替换为实际的直播管理菜单ID
  '/live/recordings',
  'icon-record',
  11,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);
```

### 方式二：创建新的直播管理菜单（如果不存在）

如果还没有直播管理的父菜单，先创建父菜单：

```sql
-- 1. 创建直播管理父菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_management',
  '直播管理',
  NULL,  -- 顶级菜单
  NULL,  -- 父菜单没有URL
  'icon-live-broadcast',
  50,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);

-- 2. 获取刚创建的父菜单ID
SET @parent_id = LAST_INSERT_ID();

-- 3. 添加直播域名管理子菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_domains',
  '域名管理',
  @parent_id,
  '/live/domains',
  'icon-link',
  1,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);

-- 4. 添加直播流管理子菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_streams',
  '流管理',
  @parent_id,
  '/live/streams',
  'icon-play-circle',
  2,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);

-- 5. 添加拉流转推子菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_pull_stream',
  '拉流转推',
  @parent_id,
  '/live/pull-stream',
  'icon-swap',
  3,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);

-- 6. 添加高光切片子菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_highlight_clips',
  '高光切片',
  @parent_id,
  '/live/highlight-clips',
  'icon-video',
  4,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);

-- 7. 添加录制列表子菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME,
  UPDATE_BY,
  UPDATE_TIME
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_recordings',
  '录制列表',
  @parent_id,
  '/live/recordings',
  'icon-record',
  5,
  'Y',
  1,
  'admin',
  NOW(),
  'admin',
  NOW()
);
```

### 方式三：简化版（仅添加两个新菜单）

如果您只想快速添加这两个菜单，不关心父菜单结构：

```sql
-- 查找合适的父菜单ID
SELECT ID, NAME, DISPLAY_NAME, PARENT_ID
FROM sys_subsystem
WHERE IS_ACTIVE = 'Y'
ORDER BY ORDER_NO;

-- 假设找到的父菜单ID是 100，执行以下SQL：

-- 添加高光切片
INSERT INTO sys_subsystem (
  NAME, DISPLAY_NAME, PARENT_ID, URL, ICON, ORDER_NO, IS_ACTIVE, SYS_COMPANY_ID
) VALUES (
  'live_highlight_clips', '高光切片', 100, '/live/highlight-clips', 'icon-video', 10, 'Y', 1
);

-- 添加录制列表
INSERT INTO sys_subsystem (
  NAME, DISPLAY_NAME, PARENT_ID, URL, ICON, ORDER_NO, IS_ACTIVE, SYS_COMPANY_ID
) VALUES (
  'live_recordings', '录制列表', 100, '/live/recordings', 'icon-record', 11, 'Y', 1
);
```

## 验证配置

### 1. 检查菜单是否添加成功

```sql
SELECT
  s.ID,
  s.NAME,
  s.DISPLAY_NAME,
  s.URL,
  s.PARENT_ID,
  p.DISPLAY_NAME as PARENT_NAME,
  s.ORDER_NO,
  s.IS_ACTIVE
FROM sys_subsystem s
LEFT JOIN sys_subsystem p ON s.PARENT_ID = p.ID
WHERE s.NAME IN ('live_highlight_clips', 'live_recordings')
ORDER BY s.ORDER_NO;
```

### 2. 查看完整的直播菜单树

```sql
SELECT
  s.ID,
  s.NAME,
  s.DISPLAY_NAME,
  s.URL,
  s.PARENT_ID,
  s.ORDER_NO,
  s.IS_ACTIVE
FROM sys_subsystem s
WHERE s.NAME LIKE 'live%' OR s.DISPLAY_NAME LIKE '%直播%'
ORDER BY s.PARENT_ID, s.ORDER_NO;
```

## 权限配置（可选）

如果系统有权限控制，需要配置相应的权限：

```sql
-- 1. 添加高光切片查看权限
INSERT INTO sys_directory (
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME
) VALUES (
  'live_highlight_clips_view',
  '查看高光切片',
  (SELECT ID FROM sys_directory WHERE NAME = 'live_management'),
  10,
  'Y',
  1,
  'admin',
  NOW()
);

-- 2. 添加录制列表查看权限
INSERT INTO sys_directory (
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID,
  CREATE_BY,
  CREATE_TIME
) VALUES (
  'live_recordings_view',
  '查看录制列表',
  (SELECT ID FROM sys_directory WHERE NAME = 'live_management'),
  11,
  'Y',
  1,
  'admin',
  NOW()
);

-- 3. 为管理员组授权
INSERT INTO sys_group_prem (
  GROUP_ID,
  DIRECTORY_ID,
  PREM_VALUE,
  SYS_COMPANY_ID
) VALUES (
  (SELECT ID FROM sys_groups WHERE NAME = 'admin'),
  (SELECT ID FROM sys_directory WHERE NAME = 'live_highlight_clips_view'),
  7,  -- 7 = 读写权限
  1
);

INSERT INTO sys_group_prem (
  GROUP_ID,
  DIRECTORY_ID,
  PREM_VALUE,
  SYS_COMPANY_ID
) VALUES (
  (SELECT ID FROM sys_groups WHERE NAME = 'admin'),
  (SELECT ID FROM sys_directory WHERE NAME = 'live_recordings_view'),
  7,  -- 7 = 读写权限
  1
);
```

## 刷新菜单

配置完成后，需要刷新菜单缓存：

### 方法一：重新登录
退出系统后重新登录，菜单会自动重新加载。

### 方法二：清除缓存
如果系统有缓存管理功能，清除菜单缓存。

### 方法三：重启服务
重启后端服务，清除所有缓存。

## 测试步骤

1. **执行SQL** - 在数据库中执行上述SQL语句
2. **刷新菜单** - 重新登录或清除缓存
3. **查看菜单** - 在左侧菜单栏应该能看到"高光切片"和"录制列表"
4. **点击菜单** - 点击菜单项，页面应该正常加载
5. **测试功能** - 测试筛选、预览、下载等功能

## 常见问题

### Q1: 菜单添加成功但点击没反应

**原因**：
- 前端代码未更新
- 浏览器缓存未清除
- URL路径配置错误

**解决方法**：
1. 确认 `BasicLayout.vue` 已更新
2. 清除浏览器缓存（Ctrl+Shift+Delete）
3. 检查数据库中的URL是否正确：`/live/highlight-clips` 和 `/live/recordings`

### Q2: 菜单不显示

**原因**：
- 菜单未激活（IS_ACTIVE = 'N'）
- 父菜单ID错误
- 权限不足

**解决方法**：
1. 检查 `IS_ACTIVE` 字段是否为 'Y'
2. 检查 `PARENT_ID` 是否正确
3. 检查用户权限配置

### Q3: 页面显示空白

**原因**：
- 组件导入路径错误
- 组件未在 componentRegistry 中注册
- 路由路径不匹配

**解决方法**：
1. 检查浏览器控制台是否有错误
2. 确认组件文件存在：`src/pages/LiveHighlightClips.vue` 和 `src/pages/LiveRecordings.vue`
3. 确认 URL 路径与 switch case 中的路径一致

## 完整的菜单结构示例

```
直播管理 (live_management)
├── 域名管理 (live_domains) - /live/domains
├── 流管理 (live_streams) - /live/streams
├── 拉流转推 (live_pull_stream) - /live/pull-stream
├── 高光切片 (live_highlight_clips) - /live/highlight-clips
└── 录制列表 (live_recordings) - /live/recordings
```

## 图标说明

可用的图标（Arco Design Icons）：
- `icon-video` - 视频图标（推荐用于高光切片）
- `icon-record` - 录制图标（推荐用于录制列表）
- `icon-play-circle` - 播放图标
- `icon-live-broadcast` - 直播图标
- `icon-link` - 链接图标
- `icon-swap` - 交换图标

更多图标请参考：https://arco.design/vue/component/icon

## 相关文档

- [页面功能说明](./live_callback_PAGES.md)
- [回调功能完整文档](./live_callback_README_FULL.md)
- [回调配置说明](./live_callback_CONFIG.md)

---

## PAGES

## 新增页面

已创建以下两个新页面：

1. **LiveHighlightClips.vue** - 直播高光切片列表
2. **LiveRecordings.vue** - 直播录制列表

## 页面路径

```
src/pages/LiveHighlightClips.vue
src/pages/LiveRecordings.vue
```

## 菜单配置

### 方式一：通过数据库配置（推荐）

在 `sys_subsystem` 表中添加菜单项：

```sql
-- 1. 添加直播高光切片菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_highlight_clips',
  '高光切片',
  (SELECT ID FROM sys_subsystem WHERE NAME = 'live_management'),  -- 假设直播管理的ID
  '/live/highlight-clips',
  'icon-video',
  10,
  'Y',
  1
);

-- 2. 添加直播录制列表菜单
INSERT INTO sys_subsystem (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  URL,
  ICON,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_subsystem s),
  'live_recordings',
  '录制列表',
  (SELECT ID FROM sys_subsystem WHERE NAME = 'live_management'),  -- 假设直播管理的ID
  '/live/recordings',
  'icon-record',
  11,
  'Y',
  1
);
```

### 方式二：手动配置路由

如果项目使用前端路由配置，需要在路由文件中添加：

```typescript
// router/index.ts 或相应的路由配置文件

{
  path: '/live',
  name: 'Live',
  meta: { title: '直播管理' },
  children: [
    // ... 其他直播相关路由
    {
      path: 'highlight-clips',
      name: 'LiveHighlightClips',
      component: () => import('@/pages/LiveHighlightClips.vue'),
      meta: {
        title: '高光切片',
        icon: 'icon-video',
        requiresAuth: true
      }
    },
    {
      path: 'recordings',
      name: 'LiveRecordings',
      component: () => import('@/pages/LiveRecordings.vue'),
      meta: {
        title: '录制列表',
        icon: 'icon-record',
        requiresAuth: true
      }
    }
  ]
}
```

## 页面功能说明

### 1. 直播高光切片列表 (LiveHighlightClips.vue)

**功能特性**：
- ✅ 查询高光切片列表
- ✅ 按流名称、域名、应用名称筛选
- ✅ 按时间范围筛选
- ✅ 显示精彩度评分（带颜色标识）
- ✅ 视频预览功能
- ✅ 下载切片文件
- ✅ 查看详细信息
- ✅ 分页显示

**数据来源**：
- 从回调事件表查询 `event_type = 'highlight'` 的记录
- 解析 `event_data` 字段获取切片详情

**评分颜色规则**：
- 9分以上：红色（非常精彩）
- 8-9分：橙红色（很精彩）
- 7-8分：橙色（精彩）
- 6-7分：金色（较精彩）
- 6分以下：蓝色（一般）

### 2. 直播录制列表 (LiveRecordings.vue)

**功能特性**：
- ✅ 查询录制文件列表
- ✅ 按流名称、域名、应用名称筛选
- ✅ 按文件格式筛选（FLV/MP4/HLS/AAC）
- ✅ 按时间范围筛选
- ✅ 显示统计数据（总录制数、今日录制、总大小、总时长）
- ✅ 视频/音频预览功能
- ✅ 下载录制文件
- ✅ 查看详细信息
- ✅ 删除录制记录
- ✅ 导出数据
- ✅ 分页显示

**数据来源**：
- 从回调事件表查询 `event_type = 'recording_file'` 的记录
- 解析 `event_data` 字段获取录制文件详情

**文件格式支持**：
- FLV：蓝色标签
- MP4：绿色标签
- HLS：橙色标签
- AAC：紫色标签

## API 接口

两个页面都使用统一的回调事件查询接口：

```typescript
GET /api/v1/live/callback/events

参数：
- eventType: string (highlight 或 recording_file)
- streamName: string (可选)
- domainName: string (可选)
- appName: string (可选)
- startTime: string (可选，格式：YYYY-MM-DD HH:mm:ss)
- endTime: string (可选，格式：YYYY-MM-DD HH:mm:ss)
- pageNum: number (默认1)
- pageSize: number (默认20)
```

## 权限配置

如果系统有权限控制，需要配置相应的权限：

```sql
-- 添加高光切片查看权限
INSERT INTO sys_directory (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  TABLE_ID,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_directory d),
  'live_highlight_clips_view',
  '查看高光切片',
  (SELECT ID FROM sys_directory WHERE NAME = 'live_management'),
  NULL,
  10,
  'Y',
  1
);

-- 添加录制列表查看权限
INSERT INTO sys_directory (
  ID,
  NAME,
  DISPLAY_NAME,
  PARENT_ID,
  TABLE_ID,
  ORDER_NO,
  IS_ACTIVE,
  SYS_COMPANY_ID
) VALUES (
  (SELECT IFNULL(MAX(ID), 0) + 1 FROM sys_directory d),
  'live_recordings_view',
  '查看录制列表',
  (SELECT ID FROM sys_directory WHERE NAME = 'live_management'),
  NULL,
  11,
  'Y',
  1
);
```

## 使用说明

### 1. 访问页面

配置完成后，用户可以通过以下方式访问：

- 通过左侧菜单：**直播管理** > **高光切片**
- 通过左侧菜单：**直播管理** > **录制列表**
- 直接访问URL：
  - `/live/highlight-clips`
  - `/live/recordings`

### 2. 数据要求

页面正常显示需要：
- ✅ 后端回调接口已配置并运行
- ✅ 腾讯云已配置回调URL
- ✅ 有实际的回调事件数据

### 3. 测试数据

如果需要测试，可以使用测试脚本发送模拟回调：

```bash
# 测试高光切片回调
curl -X POST "http://localhost:9090/api/v1/live/callback/highlight" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 325,
    "stream_id": "test_001",
    "t": '$(date +%s)',
    "sign": "test",
    "event_time": '$(date +%s)',
    "clip_url": "https://example.com/highlight.mp4",
    "start_time": '$(date +%s)',
    "end_time": '$(($(date +%s) + 30))',
    "score": 9.5,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test"
  }'

# 测试录制文件回调
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_001",
    "t": '$(date +%s)',
    "sign": "test",
    "event_time": '$(date +%s)',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test"
  }'
```

## 常见问题

### Q1: 页面显示"暂无数据"

**原因**：
- 没有配置腾讯云回调URL
- 回调事件还未触发
- 数据库中没有对应的事件记录

**解决方法**：
1. 检查腾讯云控制台回调配置
2. 使用测试脚本发送模拟数据
3. 检查数据库 `live_callback_event` 表

### Q2: 视频无法预览

**原因**：
- 视频URL无效或已过期
- 浏览器不支持该视频格式
- 跨域问题

**解决方法**：
1. 检查视频URL是否可访问
2. 尝试使用下载功能
3. 配置CORS允许视频资源访问

### Q3: 统计数据不准确

**原因**：
- 统计数据基于当前页数据计算
- 需要查询全部数据才能得到准确统计

**解决方法**：
- 可以添加专门的统计接口
- 或在后端实现聚合统计

## 扩展功能建议

### 1. 批量操作
- 批量下载
- 批量删除
- 批量导出

### 2. 高级筛选
- 按评分范围筛选（高光切片）
- 按文件大小范围筛选（录制列表）
- 按时长范围筛选

### 3. 数据分析
- 高光切片趋势图
- 录制文件统计图表
- 存储空间使用分析

### 4. 自动化处理
- 自动转码
- 自动上传到CDN
- 自动生成缩略图

## 相关文档

- [直播回调事件完整文档](./live_callback_README_FULL.md)
- [回调配置说明](./live_callback_CONFIG.md)
- [回调测试文档](./live_callback_TEST.md)

---

## README

## 概述

本模块实现了腾讯云直播回调事件的接收和处理功能，支持以下3种事件类型：

1. **推断流事件通知** - 推流和断流事件
2. **录制文件事件通知** - 录制文件生成事件
3. **录制状态事件通知** - 录制状态变更事件

## 功能特性

- ✅ 接收腾讯云直播回调事件
- ✅ 签名验证（可选）
- ✅ 事件数据持久化存储
- ✅ 事件查询接口
- ✅ 支持多种事件类型
- ✅ 详细的事件日志记录

## 数据库表结构

### live_callback_event 表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| ID | bigint | 主键ID |
| EVENT_TYPE | varchar(50) | 事件类型 |
| EVENT_TIME | bigint | 事件时间戳（秒） |
| DOMAIN_NAME | varchar(255) | 推流域名 |
| APP_NAME | varchar(255) | 应用名称 |
| STREAM_NAME | varchar(255) | 流名称 |
| STREAM_ID | varchar(255) | 流ID |
| CLIENT_IP | varchar(50) | 客户端IP |
| EVENT_DATA | text | 事件详细数据（JSON格式） |
| SIGN | varchar(255) | 签名 |
| T_VALUE | bigint | 签名过期时间 |
| CREATE_TIME | datetime | 创建时间 |
| SYS_COMPANY_ID | bigint | 公司ID |
| IS_ACTIVE | char(1) | 是否有效 |

## API 接口

### 1. 推流事件回调

**接口地址**: `POST /api/v1/live/callback/push`

**说明**: 接收腾讯云推流事件通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 1,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "sequence": "seq_001",
  "node": "192.168.1.1",
  "user_ip": "1.2.3.4",
  "stream_param": "",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

**响应示例**:
```json
{
  "code": 0
}
```

### 2. 断流事件回调

**接口地址**: `POST /api/v1/live/callback/disconnect`

**说明**: 接收腾讯云断流事件通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 0,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "sequence": "seq_001",
  "node": "192.168.1.1",
  "user_ip": "1.2.3.4",
  "stream_param": "",
  "duration": 3600,
  "reason": "normal",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 3. 录制文件事件回调

**接口地址**: `POST /api/v1/live/callback/recording-file`

**说明**: 接收腾讯云录制文件生成通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 100,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "video_url": "https://example.com/record.flv",
  "file_size": 1024000,
  "duration": 3600,
  "file_format": "flv",
  "start_time": 1234567890,
  "end_time": 1234571490,
  "stream_param": "",
  "video_id": "video_001",
  "record_file_id": "file_001",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 4. 录制状态事件回调

**接口地址**: `POST /api/v1/live/callback/recording-status`

**说明**: 接收腾讯云录制状态变更通知（公开接口，不需要认证）

**请求示例**:
```json
{
  "event_type": 200,
  "stream_id": "test_stream_001",
  "channel_id": "test_channel",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "status": 1,
  "stream_param": "",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 5. 查询回调事件列表

**接口地址**: `GET /api/v1/live/callback/events`

**说明**: 查询直播回调事件记录（需要认证）

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| eventType | string | 否 | 事件类型 |
| streamId | string | 否 | 流ID |
| domainName | string | 否 | 域名 |
| appName | string | 否 | 应用名称 |
| streamName | string | 否 | 流名称 |
| startTime | string | 否 | 开始时间 |
| endTime | string | 否 | 结束时间 |
| pageNum | int | 否 | 页码（默认1） |
| pageSize | int | 否 | 每页大小（默认20） |

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "eventType": "push_stream",
        "eventTime": 1234567890,
        "domainName": "push.example.com",
        "appName": "live",
        "streamName": "test_stream",
        "streamId": "test_stream_001",
        "clientIp": "1.2.3.4",
        "eventData": "{...}",
        "createTime": "2026-02-02T10:00:00Z"
      }
    ],
    "total": 100,
    "pageNum": 1,
    "pageSize": 20
  }
}
```

## 配置说明

### 1. 数据库迁移

执行以下SQL创建回调事件表：

```bash
mysql -u root -p your_database < sqls/migrations/create_live_callback_event.sql
```

### 2. 配置文件

在 `configs/config.yaml` 中添加回调密钥配置：

```yaml
tencentCloud:
  secretID: "your_secret_id"
  secretKey: "your_secret_key"
  region: "ap-guangzhou"
  callbackKey: "your_callback_key"  # 回调密钥，用于验证签名
  live:
    enabled: true
    pushDomain: "push.example.com"
    playDomain: "play.example.com"
    appName: "live"
    streamKey: "your_stream_key"
    callbackURL: "https://your-domain.com/api/v1/live/callback"
```

### 3. 腾讯云控制台配置

在腾讯云直播控制台配置回调URL：

1. 登录 [腾讯云直播控制台](https://console.cloud.tencent.com/live)
2. 进入【功能配置】>【事件中心】>【直播回调】
3. 配置回调URL：
   - 推流回调URL: `https://your-domain.com/api/v1/live/callback/push`
   - 断流回调URL: `https://your-domain.com/api/v1/live/callback/disconnect`
   - 录制回调URL: `https://your-domain.com/api/v1/live/callback/recording-file`
   - 录制状态回调URL: `https://your-domain.com/api/v1/live/callback/recording-status`
4. 配置回调密钥（与config.yaml中的callbackKey保持一致）

## 签名验证

回调签名验证算法：

```
sign = MD5(callbackKey + t)
```

其中：
- `callbackKey`: 配置文件中的回调密钥
- `t`: 签名过期时间戳（秒）

如果不需要签名验证，可以将 `callbackKey` 配置为空字符串。

## 事件类型说明

| 事件类型 | event_type | 说明 |
|----------|-----------|------|
| push_stream | 1 | 推流事件 |
| disconnect_stream | 0 | 断流事件 |
| recording_file | 100 | 录制文件生成 |
| recording_status | 200 | 录制状态变更 |

## 日志记录

所有回调事件都会记录详细日志，包括：
- 事件类型
- 流信息（域名、应用名、流名称）
- 客户端IP
- 事件详细数据

日志级别：
- INFO: 正常事件接收
- WARN: 签名验证失败
- ERROR: 数据解析失败、数据库保存失败

## 故障排查

### 1. 回调未收到

- 检查腾讯云控制台回调URL配置是否正确
- 检查服务器防火墙是否开放对应端口
- 检查服务器日志是否有错误信息

### 2. 签名验证失败

- 检查 `callbackKey` 配置是否与腾讯云控制台一致
- 检查时间戳是否过期（默认10分钟有效期）

### 3. 数据未保存

- 检查数据库连接是否正常
- 检查 `live_callback_event` 表是否已创建
- 查看服务器日志中的错误信息

## 参考文档

- [腾讯云直播回调事件通知](https://cloud.tencent.com/document/product/267/32744)
- [腾讯云直播API文档](https://cloud.tencent.com/document/product/267/20456)

---

## README_FULL

## 概述

本模块实现了腾讯云直播回调事件的接收和处理功能，支持以下17种事件类型：

### 基础事件
1. **推流事件通知** - 推流开始事件
2. **断流事件通知** - 推流结束事件

### 录制相关事件
3. **录制文件事件通知** - 录制文件生成事件
4. **录制状态事件通知** - 录制状态变更事件
5. **录制异常事件通知** - 录制过程异常事件

### 截图和审核事件
6. **截图事件通知** - 直播截图生成事件
7. **画面审核事件通知** - 视频内容审核事件
8. **音频审核事件通知** - 音频内容审核事件

### 质检和评测事件
9. **质检事件通知** - 直播质量检测事件
10. **评测阈值事件通知** - 质量指标超阈值事件
11. **评测平均分事件通知** - 质量平均分统计事件

### AI功能事件
12. **智能擦除事件通知** - 智能擦除任务完成事件
13. **直播字幕事件通知** - 实时字幕生成事件
14. **直播摘要事件通知** - 直播内容摘要事件
15. **高光切片事件通知** - 精彩片段识别事件

### 异常和监控事件
16. **推流异常事件通知** - 推流过程异常事件
17. **拉流转推事件通知** - 拉流转推状态变更事件
18. **监播事件通知** - 直播监控告警事件

## 功能特性

- ✅ 支持18种腾讯云直播回调事件
- ✅ 签名验证（可选）
- ✅ 事件数据持久化存储
- ✅ 事件查询接口
- ✅ 详细的事件日志记录
- ✅ 统一的错误处理

## API 接口列表

### 基础事件

#### 1. 推流事件回调
**接口**: `POST /api/v1/live/callback/push`

#### 2. 断流事件回调
**接口**: `POST /api/v1/live/callback/disconnect`

### 录制相关

#### 3. 录制文件事件回调
**接口**: `POST /api/v1/live/callback/recording-file`

#### 4. 录制状态事件回调
**接口**: `POST /api/v1/live/callback/recording-status`

#### 5. 录制异常事件回调
**接口**: `POST /api/v1/live/callback/record-exception`

### 截图和审核

#### 6. 截图事件回调
**接口**: `POST /api/v1/live/callback/screenshot`

**请求示例**:
```json
{
  "event_type": 200,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "pic_url": "https://example.com/screenshot.jpg",
  "width": 1920,
  "height": 1080,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 7. 画面审核事件回调
**接口**: `POST /api/v1/live/callback/video-audit`

**请求示例**:
```json
{
  "event_type": 317,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "confidence": 95,
  "label": "porn",
  "suggestion": "block",
  "screenshot_url": "https://example.com/audit.jpg",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 8. 音频审核事件回调
**接口**: `POST /api/v1/live/callback/audio-audit`

**请求示例**:
```json
{
  "event_type": 318,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "confidence": 90,
  "label": "abuse",
  "suggestion": "review",
  "audio_text": "转写的音频文本内容",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 质检和评测

#### 9. 质检事件回调
**接口**: `POST /api/v1/live/callback/quality-inspection`

**请求示例**:
```json
{
  "event_type": 319,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "diagnose_type": "video_freeze",
  "level": "warning",
  "description": "检测到视频卡顿",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 10. 评测阈值事件回调
**接口**: `POST /api/v1/live/callback/quality-threshold`

**请求示例**:
```json
{
  "event_type": 320,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "metric_type": "bitrate",
  "threshold": 2000.0,
  "current_value": 1500.0,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 11. 评测平均分事件回调
**接口**: `POST /api/v1/live/callback/quality-average`

**请求示例**:
```json
{
  "event_type": 321,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "score": 85.5,
  "duration": 3600,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### AI功能

#### 12. 智能擦除事件回调
**接口**: `POST /api/v1/live/callback/smart-erase`

**请求示例**:
```json
{
  "event_type": 322,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "task_id": "task_001",
  "status": "success",
  "output_url": "https://example.com/erased.mp4",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 13. 直播字幕事件回调
**接口**: `POST /api/v1/live/callback/subtitle`

**请求示例**:
```json
{
  "event_type": 323,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "text": "这是实时生成的字幕内容",
  "language": "zh-CN",
  "start_time": 1234567890,
  "end_time": 1234567895,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 14. 直播摘要事件回调
**接口**: `POST /api/v1/live/callback/summary`

**请求示例**:
```json
{
  "event_type": 324,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "summary": "本次直播的主要内容摘要",
  "keywords": ["关键词1", "关键词2", "关键词3"],
  "duration": 3600,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 15. 高光切片事件回调
**接口**: `POST /api/v1/live/callback/highlight`

**请求示例**:
```json
{
  "event_type": 325,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "clip_url": "https://example.com/highlight.mp4",
  "start_time": 1234567890,
  "end_time": 1234567920,
  "score": 9.5,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 异常和监控

#### 16. 推流异常事件回调
**接口**: `POST /api/v1/live/callback/push-exception`

**请求示例**:
```json
{
  "event_type": 326,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "error_code": 1001,
  "error_msg": "推流连接超时",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 17. 拉流转推事件回调
**接口**: `POST /api/v1/live/callback/pull-stream`

**请求示例**:
```json
{
  "event_type": 328,
  "task_id": "task_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "status": "start",
  "source_url": "rtmp://source.example.com/live/stream",
  "target_url": "rtmp://target.example.com/live/stream",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

#### 18. 监播事件回调
**接口**: `POST /api/v1/live/callback/monitor`

**请求示例**:
```json
{
  "event_type": 329,
  "stream_id": "test_stream_001",
  "t": 1234567890,
  "sign": "abc123...",
  "event_time": 1234567890,
  "alert_type": "black_screen",
  "alert_level": "error",
  "description": "检测到黑屏",
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 查询接口

#### 查询回调事件列表
**接口**: `GET /api/v1/live/callback/events`（需要认证）

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| eventType | string | 否 | 事件类型 |
| streamId | string | 否 | 流ID |
| domainName | string | 否 | 域名 |
| appName | string | 否 | 应用名称 |
| streamName | string | 否 | 流名称 |
| startTime | string | 否 | 开始时间 |
| endTime | string | 否 | 结束时间 |
| pageNum | int | 否 | 页码（默认1） |
| pageSize | int | 否 | 每页大小（默认20） |

## 事件类型映射表

| 事件类型 | event_type | 数据库存储名称 | 说明 |
|----------|-----------|---------------|------|
| 推流事件 | 1 | push_stream | 推流开始 |
| 断流事件 | 0 | disconnect_stream | 推流结束 |
| 录制文件 | 100 | recording_file | 录制文件生成 |
| 录制状态 | 200 | recording_status | 录制状态变更 |
| 截图事件 | 200 | screenshot | 截图生成 |
| 画面审核 | 317 | video_audit | 视频内容审核 |
| 音频审核 | 318 | audio_audit | 音频内容审核 |
| 质检事件 | 319 | quality_inspection | 质量检测 |
| 评测阈值 | 320 | quality_threshold | 阈值告警 |
| 评测平均分 | 321 | quality_average | 平均分统计 |
| 智能擦除 | 322 | smart_erase | 智能擦除完成 |
| 直播字幕 | 323 | subtitle | 实时字幕 |
| 直播摘要 | 324 | summary | 内容摘要 |
| 高光切片 | 325 | highlight | 精彩片段 |
| 推流异常 | 326 | push_exception | 推流异常 |
| 录制异常 | 327 | record_exception | 录制异常 |
| 拉流转推 | 328 | pull_stream | 拉流转推状态 |
| 监播事件 | 329 | monitor | 监控告警 |

## 配置说明

### 1. 数据库迁移

```bash
mysql -u root -p your_database < sqls/migrations/create_live_callback_event.sql
```

### 2. 配置文件

在 `configs/config.yaml` 中配置：

```yaml
tencentCloud:
  secretID: "your_secret_id"
  secretKey: "your_secret_key"
  region: "ap-guangzhou"
  callbackKey: "your_callback_key"  # 回调密钥
  live:
    enabled: true
    pushDomain: "push.example.com"
    playDomain: "play.example.com"
    appName: "live"
    streamKey: "your_stream_key"
```

### 3. 腾讯云控制台配置

在腾讯云直播控制台配置各类回调URL：

**基础事件**:
- 推流回调: `https://your-domain.com/api/v1/live/callback/push`
- 断流回调: `https://your-domain.com/api/v1/live/callback/disconnect`

**录制相关**:
- 录制文件: `https://your-domain.com/api/v1/live/callback/recording-file`
- 录制状态: `https://your-domain.com/api/v1/live/callback/recording-status`
- 录制异常: `https://your-domain.com/api/v1/live/callback/record-exception`

**截图和审核**:
- 截图: `https://your-domain.com/api/v1/live/callback/screenshot`
- 画面审核: `https://your-domain.com/api/v1/live/callback/video-audit`
- 音频审核: `https://your-domain.com/api/v1/live/callback/audio-audit`

**质检和评测**:
- 质检: `https://your-domain.com/api/v1/live/callback/quality-inspection`
- 评测阈值: `https://your-domain.com/api/v1/live/callback/quality-threshold`
- 评测平均分: `https://your-domain.com/api/v1/live/callback/quality-average`

**AI功能**:
- 智能擦除: `https://your-domain.com/api/v1/live/callback/smart-erase`
- 直播字幕: `https://your-domain.com/api/v1/live/callback/subtitle`
- 直播摘要: `https://your-domain.com/api/v1/live/callback/summary`
- 高光切片: `https://your-domain.com/api/v1/live/callback/highlight`

**异常和监控**:
- 推流异常: `https://your-domain.com/api/v1/live/callback/push-exception`
- 拉流转推: `https://your-domain.com/api/v1/live/callback/pull-stream`
- 监播: `https://your-domain.com/api/v1/live/callback/monitor`

## 使用场景

### 1. 内容审核
通过画面审核和音频审核事件，实时监控直播内容，自动识别违规内容。

### 2. 质量监控
通过质检、评测阈值和评测平均分事件，实时监控直播质量，及时发现和处理质量问题。

### 3. 智能运营
通过字幕、摘要和高光切片事件，自动生成直播内容的文字记录、摘要和精彩片段，提升运营效率。

### 4. 异常告警
通过推流异常、录制异常和监播事件，及时发现和处理直播过程中的各类异常。

### 5. 数据分析
通过查询接口，分析直播数据，生成运营报表。

## 故障排查

### 1. 回调未收到
- 检查腾讯云控制台回调URL配置
- 检查服务器防火墙设置
- 查看服务器日志

### 2. 签名验证失败
- 检查 `callbackKey` 配置
- 检查时间戳是否过期

### 3. 数据未保存
- 检查数据库连接
- 检查表是否已创建
- 查看错误日志

## 参考文档

- [腾讯云直播回调事件通知](https://cloud.tencent.com/document/product/267/32744)
- [腾讯云直播API文档](https://cloud.tencent.com/document/product/267/20456)

---

## RECORDING_ACTUAL_FIELDS_FIX

## 问题根源

腾讯云直播录制回调的**实际字段名**与**官方文档**不一致！

### 官方文档中的字段名
```json
{
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 实际回调中的字段名
```json
{
  "app": "upload.skyzhou.cn",
  "appname": "live",
  "stream_id": "0202"
}
```

**注意**：
- `push_domain` → 实际是 `app`
- `app_name` → 实际是 `appname`
- `stream_name` → 实际不存在，只有 `stream_id`

## 实际回调数据示例

```json
{
  "app" : "upload.skyzhou.cn",
  "appid" : 1301212747,
  "appname" : "live",
  "callback_ext" : "{\"video_codec\":\"h264\",\"session_id\":\"1807121258070784645\",\"resolution\":\"1920x1080\"}",
  "channel_id" : "0202",
  "duration" : 56,
  "end_time" : 1770031273,
  "end_time_usec" : 880554,
  "event_type" : 100,
  "file_format" : "mp4",
  "file_id" : "1301212747_d782d6455d5e42f78181e0e475e14cfc",
  "file_size" : 43911385,
  "media_start_time" : 33,
  "record_bps" : 0,
  "record_file_id" : "1301212747_d782d6455d5e42f78181e0e475e14cfc",
  "record_temp_id" : "1622658",
  "start_time" : 1770031217,
  "start_time_usec" : 650373,
  "stream_id" : "0202",
  "stream_param" : "txSecret=5d00f4090fdb09f6666cc9b278f47097&txTime=69830DEC",
  "task_id" : "1807121258070784645",
  "video_id" : "1301212747_11554a2e8cbb4326a0fe858f9418f5eb",
  "video_url" : "http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/0202/1807121258070784645-6d97196c6eec4b66b62355185cf34146/2026-02-02-19-20-17.mp4"
}
```

## 解决方案

### 1. 更新 RecordingFileEvent 结构体

添加实际的字段名，同时保持向后兼容：

```go
type RecordingFileEvent struct {
	EventType    int    `json:"event_type"`
	StreamID     string `json:"stream_id"`
	// ... 其他字段 ...

	// 实际的腾讯云字段（优先使用）
	App     string `json:"app"`     // 推流域名（实际字段）
	AppName string `json:"appname"` // 应用名称（实际字段）

	// 文档中的字段（向后兼容）
	PushDomain     string `json:"push_domain"`  // 推流域名（文档字段）
	AppNameCompat  string `json:"app_name"`     // 应用名称（文档字段）
	StreamName     string `json:"stream_name"`  // 流名称（文档字段）
}
```

### 2. 添加辅助函数

创建辅助函数来获取字段值，优先使用实际字段：

```go
// getDomainName 获取推流域名（优先使用实际字段 app）
func getDomainName(event RecordingFileEvent) string {
	if event.App != "" {
		return event.App
	}
	return event.PushDomain
}

// getAppName 获取应用名称（优先使用实际字段 appname）
func getAppName(event RecordingFileEvent) string {
	if event.AppName != "" {
		return event.AppName
	}
	return event.AppNameCompat
}

// getStreamName 获取流名称（优先使用 stream_name，回退到 stream_id）
func getStreamName(event RecordingFileEvent) string {
	if event.StreamName != "" {
		return event.StreamName
	}
	return event.StreamID
}
```

### 3. 更新保存逻辑

使用辅助函数来获取字段值：

```go
callbackEvent := &entity.LiveCallbackEvent{
	EventType:    "recording_file",
	EventTime:    event.EventTime,
	DomainName:   getDomainName(event),  // 优先使用 app
	AppName:      getAppName(event),     // 优先使用 appname
	StreamName:   getStreamName(event),  // 优先使用 stream_name，回退到 stream_id
	StreamID:     event.StreamID,
	EventData:    string(eventData),
	Sign:         event.Sign,
	TValue:       event.T,
	SysCompanyID: 1,
	IsActive:     "Y",
}
```

## 修改的文件

- `sky-server/api/handler/live_callback_handler.go`
  - 第69-95行：更新 `RecordingFileEvent` 结构体
  - 第520-545行：更新保存逻辑
  - 第1010-1033行：添加辅助函数

## 字段映射关系

| 数据库字段 | 实际回调字段 | 文档字段 | 说明 |
|-----------|-------------|---------|------|
| DOMAIN_NAME | `app` | `push_domain` | 推流域名 |
| APP_NAME | `appname` | `app_name` | 应用名称 |
| STREAM_NAME | - | `stream_name` | 流名称（实际回调中不存在，使用 stream_id） |
| STREAM_ID | `stream_id` | `stream_id` | 流ID |

## 向后兼容性

这个修复保持了向后兼容：
- ✅ 如果回调使用实际字段（`app`、`appname`），会正确解析
- ✅ 如果回调使用文档字段（`push_domain`、`app_name`），也能正确解析
- ✅ 如果没有 `stream_name`，会使用 `stream_id` 作为流名称

## 测试验证

### 1. 使用实际数据测试

```bash
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "0202",
    "channel_id": "0202",
    "t": 1234567890,
    "sign": "test",
    "event_time": 1234567890,
    "video_url": "http://example.com/video.mp4",
    "file_size": 43911385,
    "duration": 56,
    "file_format": "mp4",
    "start_time": 1234567834,
    "end_time": 1234567890,
    "app": "upload.skyzhou.cn",
    "appname": "live"
  }'
```

### 2. 检查数据库

```sql
SELECT
    ID,
    DOMAIN_NAME,
    APP_NAME,
    STREAM_NAME,
    STREAM_ID,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 1;
```

**预期结果**：
- `DOMAIN_NAME` = "upload.skyzhou.cn"
- `APP_NAME` = "live"
- `STREAM_NAME` = "0202" (使用 stream_id)
- `STREAM_ID` = "0202"

### 3. 检查前端显示

刷新录制列表页面，应该能看到：
- 域名：upload.skyzhou.cn
- 应用名：live
- 流名称：0202

## 其他发现的字段

实际回调中还包含以下额外字段（可以考虑后续添加）：

- `appid`: 应用ID
- `callback_ext`: 回调扩展信息（JSON字符串）
  - `video_codec`: 视频编码
  - `session_id`: 会话ID
  - `resolution`: 分辨率
- `end_time_usec`: 结束时间微秒
- `start_time_usec`: 开始时间微秒
- `media_start_time`: 媒体开始时间
- `record_bps`: 录制码率
- `record_temp_id`: 录制临时ID
- `task_id`: 任务ID
- `file_id`: 文件ID

## 总结

这个问题的根本原因是**腾讯云的实际回调数据与官方文档不一致**。通过添加实际字段名并保持向后兼容，我们解决了字段为空的问题。

修复后，系统能够：
1. 正确解析腾讯云的实际回调数据
2. 保持对文档字段的向后兼容
3. 在前端正确显示域名、应用名和流名称

---

## RECORDING_EMPTY_DATA_FIX

## 问题描述

用户报告录制列表页面显示多个空列，包括：
- 流名称
- 域名
- 应用名称
- 文件格式
- 文件大小
- 录制时长
- 录制时间

## 问题原因

### 1. 数据库中的事件数据不完整

通过控制台日志发现，数据库中存储的 `eventData` 只包含最基本的字段：
```json
{
  "event_type": 100,
  "stream_id": "0202"
}
```

而前端期望的完整字段包括：
```json
{
  "event_type": 100,
  "stream_id": "test_stream_001",
  "video_url": "https://example.com/record.flv",
  "file_size": 1024000,
  "duration": 3600,
  "file_format": "flv",
  "start_time": 1234567890,
  "end_time": 1234571490,
  "push_domain": "push.example.com",
  "app_name": "live",
  "stream_name": "test_stream"
}
```

### 2. 可能的原因

1. **测试数据不完整**：数据库中的数据是通过简化的测试脚本创建的，没有包含所有字段
2. **腾讯云回调未配置**：实际的腾讯云录制回调还未配置或触发
3. **回调数据格式变化**：腾讯云的实际回调格式可能与文档不一致

## 解决方案

### 前端修复（已完成）

更新 `LiveRecordings.vue` 以更好地处理缺失数据：

1. **增强字段提取逻辑**：
   - 支持多种可能的字段名（snake_case 和 camelCase）
   - 从 `eventData` 和 `item` 的多个位置尝试获取数据
   - 使用空值合并运算符 `??` 处理 0 值

```typescript
// 从 eventData 中提取字段，支持多种可能的字段名
const videoUrl = eventData.video_url || eventData.videoUrl || ''
const fileSize = eventData.file_size ?? eventData.fileSize ?? 0
const duration = eventData.duration ?? 0
const fileFormat = eventData.file_format || eventData.fileFormat || ''

// 如果 eventData 中没有这些字段，尝试从 item 的其他字段获取
const streamName = item.streamName || eventData.stream_name || eventData.streamName || ''
const domainName = item.domainName || eventData.push_domain || eventData.pushDomain || ''
const appName = item.appName || eventData.app_name || eventData.appName || ''
```

2. **改进空值显示**：
   - 文件大小为 0 或 null 时显示 "-" 而不是 "0 B"
   - 时长为 0 或 null 时显示 "-" 而不是 "0秒"
   - 空字符串字段显示 "-"
   - 文件格式为空时显示 "-" 而不是空的标签

3. **添加调试日志**：
```typescript
console.log('Available eventData keys:', Object.keys(eventData))
```

### 后端验证（需要检查）

需要验证后端是否正确保存了完整的事件数据：

1. **检查 `HandleRecordingFile` 函数**：
   - 确认是否正确解析了所有字段
   - 确认 `json.Marshal(event)` 是否包含所有字段

2. **检查数据库表结构**：
   - `live_callback_event` 表的 `event_data` 字段是否足够大
   - 是否有字段长度限制导致数据被截断

3. **检查实际回调数据**：
   - 查看后端日志，确认腾讯云发送的实际数据格式
   - 对比文档和实际数据的差异

## 测试方法

### 1. 使用完整测试数据

创建包含所有字段的测试数据：

```bash
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_001",
    "t": '$(date +%s)',
    "sign": "test",
    "event_time": '$(date +%s)',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "start_time": '$(date +%s)',
    "end_time": '$(date +%s)',
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream",
    "record_file_id": "file_001",
    "video_id": "video_001"
  }'
```

### 2. 检查数据库

```sql
SELECT id, stream_id, stream_name, domain_name, app_name,
       LENGTH(event_data) as data_length,
       event_data
FROM live_callback_event
WHERE event_type = 'recording_file'
ORDER BY id DESC
LIMIT 5;
```

### 3. 查看前端控制台

打开浏览器控制台，查看以下日志：
- `Available eventData keys:` - 显示实际可用的字段
- `Parsed recording eventData:` - 显示解析后的完整数据

## 预期结果

修复后，即使数据不完整，页面也应该：
1. 显示 "-" 而不是空白单元格
2. 正确显示可用的字段
3. 不会因为缺失字段而报错

## 后续改进建议

1. **后端数据验证**：
   - 在保存事件前验证必需字段
   - 记录警告日志如果字段缺失

2. **前端数据提示**：
   - 在列表顶部显示数据完整性提示
   - 如果大部分记录缺少关键字段，提示用户检查配置

3. **配置检查工具**：
   - 添加腾讯云回调配置检查功能
   - 提供测试回调的工具

## 相关文件

- `sky-web/src/pages/LiveRecordings.vue` - 录制列表页面（已修复）
- `sky-server/api/handler/live_callback_handler.go` - 录制回调处理器
- `sky-server/docs/live_callback_TEST.md` - 测试文档
- `sky-server/docs/live_callback_README.md` - API 文档

---

## RECORDING_EMPTY_FIELDS_DEBUG

## 问题描述

直播录制列表页面中，以下字段显示为空：
- 流名称 (streamName)
- 域名 (domainName)
- 应用名称 (appName)

## 代码检查结果

### 1. 后端保存逻辑 ✓ 正确

`live_callback_handler.go:533-545`：
```go
callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    event.EventTime,
    DomainName:   event.PushDomain,      // ✓ 正确映射
    AppName:      event.AppName,          // ✓ 正确映射
    StreamName:   event.StreamName,       // ✓ 正确映射
    StreamID:     event.StreamID,
    EventData:    string(eventData),
    Sign:         event.Sign,
    TValue:       event.T,
    SysCompanyID: 1,
    IsActive:     "Y",
}
```

### 2. 数据库表结构 ✓ 正确

`create_live_callback_event.sql`：
```sql
`DOMAIN_NAME` varchar(255) DEFAULT NULL COMMENT '推流域名',
`APP_NAME` varchar(255) DEFAULT NULL COMMENT '应用名称',
`STREAM_NAME` varchar(255) DEFAULT NULL COMMENT '流名称',
```

### 3. 实体定义 ✓ 正确

`live_callback_event.go:12-14`：
```go
DomainName   string    `gorm:"column:DOMAIN_NAME;type:varchar(255);index" json:"domainName"`
AppName      string    `gorm:"column:APP_NAME;type:varchar(255);index" json:"appName"`
StreamName   string    `gorm:"column:STREAM_NAME;type:varchar(255);index" json:"streamName"`
```

### 4. 查询接口 ✓ 正确

`live_callback_handler.go:685-696`：
```go
if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(size).Find(&events).Error; err != nil {
    // ...
}

utils.Success(c, gin.H{
    "list":     events,  // 直接返回实体数组
    "total":    total,
    "pageNum":  page,
    "pageSize": size,
})
```

### 5. 前端解析逻辑 ✓ 正确

`LiveRecordings.vue:442-444`：
```typescript
const streamName = item.streamName || eventData.stream_name || eventData.streamName || ''
const domainName = item.domainName || eventData.push_domain || eventData.pushDomain || ''
const appName = item.appName || eventData.app_name || eventData.appName || ''
```

## 问题根源

代码逻辑都是正确的，问题很可能是：

### 1. 数据库中的数据本身就是空的

可能的原因：
- 测试数据是手动创建的，只包含了 `event_type` 和 `stream_id` 等最基本字段
- 实际的腾讯云回调还没有触发，或者回调数据不完整
- 早期的测试数据没有包含这些字段

### 2. GORM 查询时的字段选择问题

虽然不太可能，但可以尝试显式指定要查询的字段。

## 解决方案

### 方案1：检查数据库数据（推荐）

执行以下 SQL 查询，检查数据库中的实际数据：

```sql
-- 查看最近的录制事件数据
SELECT
    ID,
    EVENT_TYPE,
    DOMAIN_NAME,
    APP_NAME,
    STREAM_NAME,
    STREAM_ID,
    EVENT_DATA,
    CREATE_TIME
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 5;
```

**预期结果**：
- 如果 `DOMAIN_NAME`、`APP_NAME`、`STREAM_NAME` 列都是 NULL 或空字符串，说明数据本身就是空的
- 如果这些列有值，但前端显示为空，说明是查询或序列化问题

### 方案2：使用完整的测试数据

如果数据库中的数据确实是空的，使用以下完整的测试数据：

```bash
#!/bin/bash

# 发送完整的录制文件回调测试数据
curl -X POST "http://localhost:9090/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '$(date +%s)',
    "sign": "test_sign",
    "event_time": '$(date +%s)',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "start_time": '$(date +%s)',
    "end_time": '$(date +%s)',
    "stream_param": "",
    "video_id": "video_001",
    "record_file_id": "file_001",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

**关键字段**：
- `push_domain` → 保存到 `DOMAIN_NAME`
- `app_name` → 保存到 `APP_NAME`
- `stream_name` → 保存到 `STREAM_NAME`

### 方案3：修改后端查询（如果方案1和2都不行）

如果确认数据库有数据但查询不到，可以尝试显式指定查询字段：

```go
// 在 QueryCallbackEvents 函数中
if err := query.
    Select("ID, EVENT_TYPE, EVENT_TIME, DOMAIN_NAME, APP_NAME, STREAM_NAME, STREAM_ID, CLIENT_IP, EVENT_DATA, SIGN, T_VALUE, CREATE_TIME, SYS_COMPANY_ID, IS_ACTIVE").
    Order("CREATE_TIME DESC").
    Offset(offset).
    Limit(size).
    Find(&events).Error; err != nil {
    // ...
}
```

### 方案4：添加调试日志

在后端查询接口中添加日志，查看实际返回的数据：

```go
// 在 QueryCallbackEvents 函数中，返回之前添加
logger.Info("查询到的事件数据",
    zap.Int("count", len(events)),
    zap.Any("firstEvent", events[0])) // 打印第一条数据

utils.Success(c, gin.H{
    "list":     events,
    "total":    total,
    "pageNum":  page,
    "pageSize": size,
})
```

## 验证步骤

### 1. 检查数据库

```sql
-- 检查是否有数据
SELECT COUNT(*) FROM live_callback_event WHERE EVENT_TYPE = 'recording_file';

-- 检查字段是否为空
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN DOMAIN_NAME IS NULL OR DOMAIN_NAME = '' THEN 1 ELSE 0 END) as empty_domain,
    SUM(CASE WHEN APP_NAME IS NULL OR APP_NAME = '' THEN 1 ELSE 0 END) as empty_app,
    SUM(CASE WHEN STREAM_NAME IS NULL OR STREAM_NAME = '' THEN 1 ELSE 0 END) as empty_stream
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';
```

### 2. 发送测试数据

使用上面的 curl 命令发送完整的测试数据。

### 3. 检查前端控制台

打开浏览器控制台，查看以下日志：
- `Processing recording item:` - 查看 `item.domainName`、`item.appName`、`item.streamName` 的值
- `Parsed recording eventData:` - 查看 `eventData` 中的字段

### 4. 检查后端日志

查看后端日志中的 "收到录制文件事件" 日志，确认：
- `domain`: 是否有值
- `app`: 是否有值
- `stream`: 是否有值

## 最可能的原因

根据之前的调试信息（控制台显示 `eventData` 只有 `event_type` 和 `stream_id`），最可能的原因是：

**数据库中的数据本身就是不完整的测试数据**

解决方法：
1. 清空现有的测试数据
2. 使用完整的测试脚本重新创建数据
3. 或者配置腾讯云回调，等待真实的回调数据

## 清空测试数据

```sql
-- 删除所有录制事件测试数据
DELETE FROM live_callback_event WHERE EVENT_TYPE = 'recording_file';

-- 或者只删除字段为空的数据
DELETE FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
  AND (DOMAIN_NAME IS NULL OR DOMAIN_NAME = '')
  AND (APP_NAME IS NULL OR APP_NAME = '')
  AND (STREAM_NAME IS NULL OR STREAM_NAME = '');
```

## 总结

1. **首先**：执行 SQL 查询检查数据库中的实际数据
2. **如果数据为空**：使用完整的测试脚本重新创建数据
3. **如果数据不为空**：添加后端日志，检查查询和序列化过程
4. **最后**：检查前端控制台日志，确认数据传输是否正确

---

## RECORDING_EVENT_TIME_DEBUG

## 问题描述

直播录制列表页面中，"生成时间"列不显示数据。

## 问题分析

### 1. 前端代码检查

**表格列定义** (LiveRecordings.vue:177-181)：
```vue
<a-table-column title="生成时间" :width="180">
  <template #cell="{ record }">
    {{ formatDateTime(record.eventTime) }}
  </template>
</a-table-column>
```

**数据解析** (LiveRecordings.vue:468)：
```typescript
eventTime: item.eventTime != null ? item.eventTime * 1000 : null
```

**时间格式化函数** (LiveRecordings.vue:589-594)：
```typescript
const formatDateTime = (timestamp: number | string) => {
  if (timestamp == null || isNaN(Number(timestamp))) return '-'
  const date = dayjs(timestamp)
  if (!date.isValid()) return '-'
  return date.format('YYYY-MM-DD HH:mm:ss')
}
```

### 2. 后端数据结构

**实体定义** (live_callback_event.go:11)：
```go
EventTime int64 `gorm:"column:EVENT_TIME;not null" json:"eventTime"` // 事件时间戳（秒）
```

**腾讯云回调数据**：
```json
{
  "event_time": 1770031273,
  ...
}
```

### 3. 可能的原因

1. **后端 eventTime 字段为 0 或 null**
   - 保存时没有正确设置 eventTime
   - 数据库中的值为 0

2. **时间戳单位问题**
   - 后端存储的是秒级时间戳
   - 前端需要毫秒级时间戳（乘以 1000）
   - 如果时间戳太小或太大，dayjs 可能无法正确解析

3. **字段名映射问题**
   - JSON 序列化时字段名不匹配

## 调试步骤

### 已添加的调试日志

在 `LiveRecordings.vue` 中添加了以下调试日志：

```typescript
console.log('Processing recording item:', item)
console.log('item.eventTime:', item.eventTime, 'type:', typeof item.eventTime)
const eventData = JSON.parse(item.eventData)
console.log('Parsed recording eventData:', eventData)
console.log('Available eventData keys:', Object.keys(eventData))

// ...

const processedEventTime = item.eventTime != null ? item.eventTime * 1000 : null
console.log('Processed eventTime:', processedEventTime, 'from:', item.eventTime)
```

### 如何查看调试信息

1. 打开浏览器开发者工具（F12）
2. 切换到 Console 标签
3. 刷新录制列表页面
4. 查看控制台输出

### 需要检查的信息

1. **item.eventTime 的值**：
   - 是否为 null 或 undefined？
   - 是否为 0？
   - 数值是否合理（应该是秒级时间戳，如 1770031273）？

2. **processedEventTime 的值**：
   - 乘以 1000 后的值是否正确？
   - 是否为 null？

3. **eventData 中的 event_time**：
   - 检查原始回调数据中是否有 event_time 字段
   - 值是否正确？

## 可能的解决方案

### 方案1：后端保存时设置 eventTime

如果 `item.eventTime` 为 0 或 null，需要检查后端保存逻辑：

**检查点** (live_callback_handler.go:533-545)：
```go
callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    event.EventTime,  // 确保这个值不为 0
    DomainName:   getDomainName(event),
    AppName:      getAppName(event),
    StreamName:   getStreamName(event),
    StreamID:     event.StreamID,
    EventData:    string(eventData),
    Sign:         event.Sign,
    TValue:       event.T,
    SysCompanyID: 1,
    IsActive:     "Y",
}
```

**可能的问题**：
- 腾讯云回调中 `event_time` 字段为 0
- 结构体解析失败，`event.EventTime` 为 0

**解决方法**：
如果 `event.EventTime` 为 0，使用当前时间：
```go
eventTime := event.EventTime
if eventTime == 0 {
    eventTime = time.Now().Unix()
}

callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    eventTime,
    // ...
}
```

### 方案2：前端使用 createTime 作为备选

如果 eventTime 不可用，使用 createTime：

```typescript
const processedEventTime = item.eventTime != null && item.eventTime > 0
  ? item.eventTime * 1000
  : (item.createTime ? new Date(item.createTime).getTime() : null)
```

### 方案3：从 eventData 中提取 event_time

如果后端的 eventTime 字段没有正确保存，可以从 eventData 中提取：

```typescript
const eventTime = item.eventTime || eventData.event_time || eventData.eventTime || null
const processedEventTime = eventTime != null && eventTime > 0
  ? eventTime * 1000
  : null
```

## 数据库检查

### SQL 查询

检查数据库中的实际数据：

```sql
-- 查看最近的录制事件
SELECT
    ID,
    EVENT_TYPE,
    EVENT_TIME,
    CREATE_TIME,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 5;

-- 检查 EVENT_TIME 为 0 的记录
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN EVENT_TIME = 0 THEN 1 ELSE 0 END) as zero_event_time,
    SUM(CASE WHEN EVENT_TIME IS NULL THEN 1 ELSE 0 END) as null_event_time
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';
```

### 预期结果

- `EVENT_TIME` 应该是秒级时间戳（如 1770031273）
- `EVENT_TIME` 不应该为 0 或 NULL
- `CREATE_TIME` 应该是 datetime 格式

## 时间戳验证

### 验证时间戳是否合理

```javascript
// 在浏览器控制台中验证
const timestamp = 1770031273 // 从数据库或日志中获取
console.log('原始时间戳（秒）:', timestamp)
console.log('转换为毫秒:', timestamp * 1000)
console.log('格式化时间:', new Date(timestamp * 1000).toLocaleString())
```

**预期输出**：
```
原始时间戳（秒）: 1770031273
转换为毫秒: 1770031273000
格式化时间: 2026/2/2 19:21:13
```

如果时间不合理（如显示为 1970年），说明时间戳有问题。

## 下一步

1. **查看控制台日志**：
   - 打开页面，查看 `item.eventTime` 的实际值
   - 确认是否为 0、null 或其他异常值

2. **检查数据库**：
   - 运行 SQL 查询，查看 `EVENT_TIME` 字段的实际值
   - 确认是否有数据

3. **根据调试结果选择解决方案**：
   - 如果 eventTime 为 0：修改后端保存逻辑
   - 如果 eventTime 为 null：使用 createTime 作为备选
   - 如果时间戳格式错误：调整转换逻辑

## 临时解决方案

在确定根本原因之前，可以使用 createTime 作为备选：

```typescript
eventTime: item.eventTime != null && item.eventTime > 0
  ? item.eventTime * 1000
  : (item.createTime ? new Date(item.createTime).getTime() : null)
```

这样至少能显示创建时间，而不是空白。

## 相关文件

- `sky-web/src/pages/LiveRecordings.vue` - 前端列表页面
- `sky-server/internal/model/entity/live_callback_event.go` - 后端实体定义
- `sky-server/api/handler/live_callback_handler.go` - 后端回调处理器

---

## RECORDING_EVENT_TIME_FIX

## 问题描述

直播录制列表页面中，"生成时间"列不显示数据，显示为空白。

## 问题原因

腾讯云录制回调中的 `event_time` 字段可能为 0 或不存在，导致：
1. 后端保存时 `EventTime` 字段为 0
2. 前端显示时，`formatDateTime(0)` 返回无效日期
3. 或者前端判断为 null/0 后显示 "-"

## 解决方案

### 后端修复：使用当前时间作为备选

在保存录制事件时，如果 `event_time` 为 0，使用当前时间：

```go
// 如果 event_time 为 0，使用当前时间
eventTime := event.EventTime
if eventTime == 0 {
    eventTime = time.Now().Unix()
    logger.Warn("录制文件事件的 event_time 为 0，使用当前时间",
        zap.String("streamId", event.StreamID),
        zap.Int64("currentTime", eventTime))
}

// 保存事件到数据库
callbackEvent := &entity.LiveCallbackEvent{
    EventType:    "recording_file",
    EventTime:    eventTime,  // 使用处理后的时间
    // ...
}
```

### 添加日志

在日志中添加 `eventTime` 字段，方便调试：

```go
logger.Info("收到录制文件事件",
    zap.String("streamId", event.StreamID),
    zap.String("domain", getDomainName(event)),
    zap.String("app", getAppName(event)),
    zap.String("stream", getStreamName(event)),
    zap.String("videoUrl", event.VideoURL),
    zap.Int64("fileSize", event.FileSize),
    zap.Int64("duration", event.Duration),
    zap.String("format", event.FileFormat),
    zap.Int64("eventTime", event.EventTime))  // 添加 eventTime 日志
```

## 修改的文件

- `sky-server/api/handler/live_callback_handler.go`
  - 第527-551行：添加 eventTime 日志和备选逻辑

## 为什么 event_time 可能为 0？

### 可能的原因

1. **腾讯云回调数据不完整**：
   - 某些情况下，腾讯云可能不发送 `event_time` 字段
   - 或者发送的值为 0

2. **JSON 解析问题**：
   - 如果字段不存在，Go 会将 int64 类型默认为 0
   - 如果字段名不匹配，也会导致解析为 0

3. **测试数据问题**：
   - 手动创建的测试数据可能没有包含 `event_time` 字段

## 腾讯云实际回调数据分析

根据用户提供的实际回调数据：

```json
{
  "app" : "upload.skyzhou.cn",
  "appid" : 1301212747,
  "appname" : "live",
  "callback_ext" : "{...}",
  "channel_id" : "0202",
  "duration" : 56,
  "end_time" : 1770031273,
  "end_time_usec" : 880554,
  "event_type" : 100,
  "file_format" : "mp4",
  "file_id" : "1301212747_d782d6455d5e42f78181e0e475e14cfc",
  "file_size" : 43911385,
  "media_start_time" : 33,
  "record_bps" : 0,
  "record_file_id" : "1301212747_d782d6455d5e42f78181e0e475e14cfc",
  "record_temp_id" : "1622658",
  "start_time" : 1770031217,
  "start_time_usec" : 650373,
  "stream_id" : "0202",
  "stream_param" : "txSecret=5d00f4090fdb09f6666cc9b278f47097&txTime=69830DEC",
  "task_id" : "1807121258070784645",
  "video_id" : "1301212747_11554a2e8cbb4326a0fe858f9418f5eb",
  "video_url" : "http://video-1301212747.cos.ap-nanjing.myqcloud.com/..."
}
```

**注意**：这个数据中**没有 `event_time` 字段**！

### 可用的时间字段

实际回调数据中有以下时间字段：
- `start_time`: 1770031217（录制开始时间，秒）
- `end_time`: 1770031273（录制结束时间，秒）
- `start_time_usec`: 650373（开始时间微秒部分）
- `end_time_usec`: 880554（结束时间微秒部分）

但是**没有 `event_time` 字段**！

## 更好的解决方案

既然腾讯云的实际回调中没有 `event_time` 字段，我们应该使用其他时间字段作为备选：

### 方案1：使用 end_time 作为生成时间

录制文件生成时间应该接近录制结束时间，所以可以使用 `end_time`：

```go
// 确定事件时间：优先使用 event_time，如果为 0 则使用 end_time，最后使用当前时间
eventTime := event.EventTime
if eventTime == 0 {
    if event.EndTime > 0 {
        eventTime = event.EndTime
        logger.Info("event_time 为 0，使用 end_time 作为生成时间",
            zap.String("streamId", event.StreamID),
            zap.Int64("endTime", event.EndTime))
    } else {
        eventTime = time.Now().Unix()
        logger.Warn("event_time 和 end_time 都为 0，使用当前时间",
            zap.String("streamId", event.StreamID),
            zap.Int64("currentTime", eventTime))
    }
}
```

### 方案2：添加 EndTime 到结构体（如果还没有）

确保 `RecordingFileEvent` 结构体中有 `EndTime` 字段（已经有了，第81行）。

## 实施更好的方案

让我更新代码以使用 `end_time` 作为备选：

```go
// 确定事件时间
eventTime := event.EventTime
if eventTime == 0 {
    // 优先使用录制结束时间
    if event.EndTime > 0 {
        eventTime = event.EndTime
        logger.Info("event_time 为 0，使用 end_time 作为生成时间",
            zap.String("streamId", event.StreamID),
            zap.Int64("endTime", event.EndTime))
    } else {
        // 最后使用当前时间
        eventTime = time.Now().Unix()
        logger.Warn("event_time 和 end_time 都为 0，使用当前时间",
            zap.String("streamId", event.StreamID),
            zap.Int64("currentTime", eventTime))
    }
}
```

## 前端调试

前端已经添加了调试日志（LiveRecordings.vue）：

```typescript
console.log('item.eventTime:', item.eventTime, 'type:', typeof item.eventTime)
console.log('Processed eventTime:', processedEventTime, 'from:', item.eventTime)
```

查看控制台输出可以确认：
1. 后端是否正确保存了 eventTime
2. 前端是否正确接收和处理了 eventTime

## 测试验证

### 1. 检查后端日志

重启后端服务后，当收到录制回调时，查看日志：

```
INFO  收到录制文件事件
  streamId: 0202
  domain: upload.skyzhou.cn
  app: live
  stream: 0202
  videoUrl: http://video-1301212747.cos.ap-nanjing.myqcloud.com/...
  fileSize: 43911385
  duration: 56
  format: mp4
  eventTime: 0  <-- 如果为 0，说明回调中没有 event_time
```

如果看到 `eventTime: 0`，然后应该看到：

```
INFO  event_time 为 0，使用 end_time 作为生成时间
  streamId: 0202
  endTime: 1770031273
```

### 2. 检查数据库

```sql
SELECT
    ID,
    EVENT_TYPE,
    EVENT_TIME,
    CREATE_TIME,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 1;
```

**预期结果**：
- `EVENT_TIME` 应该有值（不为 0）
- 如果原始回调没有 `event_time`，应该等于 `end_time` 的值

### 3. 检查前端显示

刷新录制列表页面，"生成时间"列应该显示时间，而不是空白或 "-"。

## 总结

### 问题根源

腾讯云的实际录制回调数据中**没有 `event_time` 字段**，导致后端保存时 `EventTime` 为 0，前端无法显示。

### 解决方案

1. **当前修复**：如果 `event_time` 为 0，使用当前时间
2. **更好的方案**：如果 `event_time` 为 0，使用 `end_time`（录制结束时间）

### 优先级

```
event_time > end_time > 当前时间
```

这样可以确保：
- 如果有 `event_time`，使用它（最准确）
- 如果没有，使用 `end_time`（接近生成时间）
- 如果都没有，使用当前时间（保底方案）

## 下一步

建议实施"更好的方案"，使用 `end_time` 作为第一备选，而不是直接使用当前时间。这样生成时间会更准确。

---

## RECORDING_MULTIPLE_REQUESTS_FIX

## 问题描述

用户报告：点击一次"查询"按钮，会发出多个API请求。

## 问题分析

### 第一版问题（3个请求）

在 `LiveRecordings.vue` 中，点击查询时会调用 `loadData()` 函数：

```typescript
// loadData() 函数
const loadData = async () => {
  // 1. 第一个请求：获取列表数据
  const res = await queryCallbackEvents(params)

  // 处理数据...
  pagination.total = res.data.data.total  // 已经获取了总数

  // 2. 调用 loadStats() 函数
  loadStats()
}

// loadStats() 函数
const loadStats = async () => {
  // 3. 第二个请求：获取总数统计（重复！）
  const totalRes = await queryCallbackEvents({
    ...statsParams,
    pageNum: 1,
    pageSize: 1
  })
  stats.total = totalRes.data.data.total

  // 4. 第三个请求：获取今日数据统计
  const todayRes = await queryCallbackEvents({
    eventType: 'recording_file',
    startTime: today,
    pageNum: 1,
    pageSize: 1
  })
  stats.today = todayRes.data.data.total
}
```

### 第一版问题根源

1. **重复请求总数**：
   - 第1个请求（列表查询）已经返回了 `total` 字段
   - 第2个请求（统计查询）又重复请求了相同的总数
   - 这是完全不必要的重复

2. **每次查询都刷新今日统计**：
   - "今日录制"是一个全局统计，不受用户筛选条件影响
   - 每次查询都重新请求今日统计是不必要的

## 解决方案

### 第一次优化（减少到2个请求）

消除重复的总数请求：

```typescript
// loadData() 函数
const loadData = async () => {
  // 1. 第一个请求：获取列表数据
  const res = await queryCallbackEvents(params)

  // 处理数据...
  pagination.total = res.data.data.total
  stats.total = res.data.data.total  // 直接使用列表查询返回的总数

  // 2. 调用 loadStats() 获取其他统计数据
  loadStats()
}

// loadStats() 函数（优化后）
const loadStats = async () => {
  // 2. 第二个请求：只获取今日数据统计
  const todayRes = await queryCallbackEvents({
    eventType: 'recording_file',
    startTime: today,
    pageNum: 1,
    pageSize: 1
  })
  stats.today = todayRes.data.data.total

  // 计算当前页的文件大小和时长（不需要请求）
  // ...
}
```

**效果**：从3个请求减少到2个请求

### 第二次优化（减少到1个请求）

将"今日录制"统计改为只在页面初始加载时请求一次：

```typescript
// 加载统计数据（只在页面初始加载时调用）
const loadStats = async () => {
  try {
    // 只加载今日数据统计
    const today = dayjs().startOf('day').format('YYYY-MM-DD HH:mm:ss')
    const todayRes = await queryCallbackEvents({
      eventType: 'recording_file',
      startTime: today,
      pageNum: 1,
      pageSize: 1
    })
    if (todayRes.data?.code === 200 || todayRes.data?.code === 0) {
      stats.today = todayRes.data.data.total
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 计算当前页统计（文件大小和时长）
const calculatePageStats = () => {
  let totalSize = 0
  let totalDuration = 0
  tableData.value.forEach((item: any) => {
    totalSize += item.fileSize || 0
    totalDuration += item.duration || 0
  })
  stats.totalSize = totalSize
  stats.totalDuration = totalDuration
}

// loadData() 函数
const loadData = async () => {
  // 1. 唯一的请求：获取列表数据
  const res = await queryCallbackEvents(params)

  // 处理数据...
  pagination.total = res.data.data.total
  stats.total = res.data.data.total

  // 计算当前页统计（不需要请求）
  calculatePageStats()
}

// 页面加载时
onMounted(() => {
  // 先加载今日统计（只加载一次）
  loadStats()
  // 然后加载列表数据
  loadData()
})
```

**效果**：
- **页面初始加载**：2个请求（今日统计 + 列表数据）
- **点击查询按钮**：1个请求（只有列表数据）
- **点击刷新按钮**：1个请求（只有列表数据）
- **切换分页**：1个请求（只有列表数据）

### 优化理由

"今日录制"是一个全局统计，特点是：
1. **不受筛选条件影响**：无论用户如何筛选，今日录制数都是固定的
2. **变化频率低**：在用户浏览页面的短时间内，今日录制数不会频繁变化
3. **实时性要求不高**：用户不需要每次查询都看到最新的今日录制数

因此，只在页面初始加载时请求一次是合理的。

## 修改的文件

- `sky-web/src/pages/LiveRecordings.vue`
  - 第362-390行：拆分 `loadStats()` 和 `calculatePageStats()` 函数
  - 第500-503行：在 `loadData()` 中调用 `calculatePageStats()` 而不是 `loadStats()`
  - 第622-627行：在 `onMounted()` 中先调用 `loadStats()`，再调用 `loadData()`

## 测试验证

### 测试步骤

1. 打开浏览器开发者工具的 Network 面板
2. 访问录制列表页面
3. 观察初始加载的请求数量
4. 点击"查询"按钮
5. 观察查询时的请求数量

### 预期结果

- **页面初始加载**：2个请求
  - 第1个：获取今日统计
  - 第2个：获取列表数据
- **点击查询按钮**：1个请求
  - 只有列表数据查询
- **点击刷新按钮**：1个请求
  - 只有列表数据查询
- **切换分页**：1个请求
  - 只有列表数据查询

### 验证要点

1. 页面数据显示正常
2. 统计卡片显示正确：
   - 总录制数（筛选后）- 随查询更新
   - 今日录制 - 保持初始值
   - 当前页文件大小 - 随查询更新
   - 当前页总时长 - 随查询更新
3. 筛选功能正常工作
4. 分页功能正常工作

## 进一步优化建议

### 1. 定时刷新今日统计

如果需要保持"今日录制"统计的实时性，可以添加定时刷新：

```typescript
// 每5分钟刷新一次今日统计
let statsTimer: number | null = null

onMounted(() => {
  loadStats()
  loadData()

  // 设置定时器
  statsTimer = setInterval(() => {
    loadStats()
  }, 5 * 60 * 1000) // 5分钟
})

onUnmounted(() => {
  // 清理定时器
  if (statsTimer) {
    clearInterval(statsTimer)
  }
})
```

### 2. 手动刷新今日统计

在页面上添加一个刷新按钮，让用户可以手动刷新今日统计：

```vue
<a-card class="stat-card">
  <a-statistic title="今日录制" :value="stats.today">
    <template #prefix>
      <icon-calendar />
    </template>
    <template #suffix>
      <a-button size="mini" type="text" @click="loadStats">
        <icon-refresh />
      </a-button>
    </template>
  </a-statistic>
</a-card>
```

### 3. 后端聚合接口

如果需要多个统计维度，考虑在后端提供一个聚合接口：

```typescript
// 一次请求获取所有统计数据
GET /api/v1/live/callback/events/stats?eventType=recording_file
// 返回：
{
  "total": 1000,
  "today": 50,
  "thisWeek": 200,
  "thisMonth": 800
}
```

## 相关问题

### Q: 为什么不在每次查询时都刷新"今日录制"？

A: 因为：
1. "今日录制"不受用户筛选条件影响，是全局统计
2. 在用户浏览页面的短时间内，这个数字不会频繁变化
3. 减少不必要的请求可以提升性能和用户体验

### Q: 如果用户长时间停留在页面，"今日录制"会过时吗？

A: 会的。如果需要保持实时性，可以：
1. 添加定时刷新（如每5分钟）
2. 添加手动刷新按钮
3. 在用户点击"刷新"按钮时也刷新统计

### Q: 能否完全避免"今日统计"请求？

A: 可以考虑：

1. **前端计算**：如果用户选择的时间范围包含今天，可以从列表数据中筛选
   - 优点：不需要额外请求
   - 缺点：只能统计当前页的数据，不准确

2. **后端优化**：在主查询响应中附带今日统计
   - 优点：只需1个请求
   - 缺点：需要修改后端接口

3. **移除统计**：如果"今日录制"不重要，可以直接移除
   - 优点：最简单
   - 缺点：失去了有用的统计信息

## 总结

通过两次优化：
1. **第一次**：消除重复请求，从3个减少到2个
2. **第二次**：将"今日统计"改为只在初始加载时请求，查询时从2个减少到1个

最终效果：
- 页面初始加载：2个请求
- 用户查询操作：1个请求（减少50%）
- 性能提升明显，用户体验更好

---

## RECORDING_PREVIEW_AUDIO_FIX

## 问题描述

在直播录制列表页面，点击"预览"按钮打开预览对话框后，关闭对话框时，视频/音频的声音仍然在播放。

## 问题原因

当使用 `v-model:visible` 控制对话框显示/隐藏时，对话框关闭后：
- 对话框的 DOM 元素被隐藏（`display: none`）
- 但是 `<video>` 和 `<audio>` 元素并没有被销毁
- 媒体元素的播放状态没有被重置
- 因此音频/视频继续在后台播放

## 解决方案

### 1. 添加媒体元素引用

为 video 和 audio 元素添加 ref 引用，以便在 JavaScript 中控制它们：

```vue
<video
  ref="videoRef"
  :src="currentRecord.videoUrl"
  controls
>
  您的浏览器不支持视频播放
</video>

<audio
  ref="audioRef"
  :src="currentRecord.videoUrl"
  controls
>
  您的浏览器不支持音频播放
</audio>
```

### 2. 在 script 中声明 ref

```typescript
// 媒体元素引用
const videoRef = ref<HTMLVideoElement | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)
```

### 3. 监听对话框关闭事件

在 `<a-modal>` 上添加关闭事件监听：

```vue
<a-modal
  v-model:visible="previewVisible"
  title="录制文件预览"
  :width="900"
  :footer="false"
  @close="handlePreviewClose"
  @cancel="handlePreviewClose"
>
```

**注意**：
- `@close` - 监听对话框关闭事件（点击 X 按钮或按 ESC 键）
- `@cancel` - 监听取消事件（点击取消按钮或遮罩层）

### 4. 实现关闭处理函数

```typescript
// 关闭预览对话框
const handlePreviewClose = () => {
  // 暂停视频播放
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  // 暂停音频播放
  if (audioRef.value) {
    audioRef.value.pause()
    audioRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

**功能说明**：
- `pause()` - 暂停播放
- `currentTime = 0` - 重置播放进度到开始位置
- `previewVisible.value = false` - 关闭对话框

## 修改的文件

- `sky-web/src/pages/LiveRecordings.vue`
  - 第199-221行：添加 ref 引用和关闭事件监听
  - 第365-367行：声明媒体元素 ref
  - 第533-548行：实现关闭处理函数

## 为什么需要重置 currentTime？

重置播放进度有以下好处：
1. **用户体验更好**：下次打开预览时，从头开始播放
2. **避免混淆**：用户不会看到上次播放到一半的进度
3. **释放资源**：某些浏览器在 currentTime 为 0 时会释放部分缓存

## 其他可能的解决方案

### 方案1：使用 v-if 代替 v-show

```vue
<video
  v-if="previewVisible && currentRecord && isVideoFormat(currentRecord.fileFormat)"
  :src="currentRecord.videoUrl"
  controls
>
```

**优点**：
- 对话框关闭时，video 元素会被完全销毁
- 不需要手动调用 pause()

**缺点**：
- 每次打开对话框都会重新创建元素
- 可能导致轻微的性能开销

### 方案2：使用 watch 监听 previewVisible

```typescript
watch(previewVisible, (newVal) => {
  if (!newVal) {
    // 对话框关闭时暂停播放
    if (videoRef.value) {
      videoRef.value.pause()
      videoRef.value.currentTime = 0
    }
    if (audioRef.value) {
      audioRef.value.pause()
      audioRef.value.currentTime = 0
    }
  }
})
```

**优点**：
- 不需要在模板中添加事件监听
- 逻辑更集中

**缺点**：
- 需要额外的 watch 监听器
- 代码稍微复杂一些

## 测试验证

### 测试步骤

1. 打开录制列表页面
2. 点击任意一条记录的"预览"按钮
3. 等待视频/音频开始播放
4. 关闭预览对话框（点击 X 按钮、取消按钮或遮罩层）
5. 验证音频是否停止播放

### 预期结果

- ✅ 关闭对话框后，音频/视频立即停止播放
- ✅ 再次打开预览时，从头开始播放
- ✅ 不会有任何残留的声音

### 测试不同的关闭方式

1. **点击 X 按钮** - 应该停止播放
2. **点击遮罩层** - 应该停止播放
3. **按 ESC 键** - 应该停止播放
4. **点击取消按钮**（如果有）- 应该停止播放

## 相关问题

### Q: 为什么不在 onUnmounted 中处理？

A: `onUnmounted` 只在组件销毁时调用，而对话框关闭时组件并没有销毁，只是隐藏了。

### Q: 是否需要处理 detailVisible（详情对话框）？

A: 详情对话框中没有视频/音频播放，所以不需要处理。

### Q: 如果用户快速打开/关闭对话框会有问题吗？

A: 不会。每次关闭都会暂停播放并重置进度，不会有累积效应。

## 扩展功能建议

### 1. 自动播放

如果需要打开对话框时自动播放：

```typescript
const handlePreview = (record: any) => {
  currentRecord.value = record
  previewVisible.value = true

  // 等待 DOM 更新后自动播放
  nextTick(() => {
    if (videoRef.value) {
      videoRef.value.play()
    }
    if (audioRef.value) {
      audioRef.value.play()
    }
  })
}
```

### 2. 播放错误处理

添加错误处理：

```vue
<video
  ref="videoRef"
  :src="currentRecord.videoUrl"
  controls
  @error="handleMediaError"
>
```

```typescript
const handleMediaError = (event: Event) => {
  const media = event.target as HTMLMediaElement
  console.error('媒体加载失败:', media.error)
  Message.error('视频加载失败，请检查文件是否存在')
}
```

### 3. 播放状态提示

显示加载状态：

```vue
<a-spin :loading="mediaLoading" tip="加载中...">
  <video
    ref="videoRef"
    :src="currentRecord.videoUrl"
    controls
    @loadstart="mediaLoading = true"
    @canplay="mediaLoading = false"
  >
  </video>
</a-spin>
```

## 总结

通过添加媒体元素引用和关闭事件处理，我们成功解决了预览对话框关闭后音频持续播放的问题。这个修复确保了：

1. ✅ 用户体验良好 - 关闭对话框后立即停止播放
2. ✅ 资源管理正确 - 及时释放媒体资源
3. ✅ 代码简洁清晰 - 使用标准的 HTML5 媒体 API
4. ✅ 兼容性好 - 适用于所有现代浏览器

---

## RECORDING_STATS_ZERO_DEBUG

## 问题描述

直播录制列表页面顶部的统计卡片中：
- "当前页文件大小" 显示为 0
- "当前页总时长" 显示为 0

## 问题分析

### 统计卡片代码

```vue
<a-card class="stat-card">
  <a-statistic title="当前页文件大小" :value="0">
    <template #prefix>
      <icon-storage />
    </template>
    <template #value>
      {{ formatFileSize(stats.totalSize) }}
    </template>
  </a-statistic>
</a-card>

<a-card class="stat-card">
  <a-statistic title="当前页总时长" :value="0">
    <template #prefix>
      <icon-clock-circle />
    </template>
    <template #value>
      {{ formatTotalDuration(stats.totalDuration) }}
    </template>
  </a-statistic>
</a-card>
```

**注意**：`:value="0"` 是硬编码的，但实际显示使用的是 `#value` 模板中的格式化函数。

### 计算逻辑

```typescript
// 计算当前页统计（文件大小和时长）
const calculatePageStats = () => {
  let totalSize = 0
  let totalDuration = 0
  tableData.value.forEach((item: any) => {
    totalSize += item.fileSize || 0
    totalDuration += item.duration || 0
  })
  stats.totalSize = totalSize
  stats.totalDuration = totalDuration
}
```

### 数据提取

```typescript
const fileSize = eventData.file_size ?? eventData.fileSize ?? 0
const duration = eventData.duration ?? 0
```

## 可能的原因

### 1. eventData 中没有这些字段

根据用户之前提供的腾讯云实际回调数据：

```json
{
  "file_size" : 43911385,
  "duration" : 56,
  ...
}
```

数据中是有这些字段的，所以问题可能不在这里。

### 2. 数据类型问题

- `file_size` 可能是字符串而不是数字
- `duration` 可能是字符串而不是数字

### 3. 字段名不匹配

- 实际字段名可能不是 `file_size` 和 `duration`
- 可能是其他变体

### 4. 数据为 null 或 undefined

- 如果字段不存在，会被设置为 0
- 累加时 0 + 0 = 0

## 调试步骤

### 已添加的调试日志

#### 1. 数据提取日志

```typescript
console.log('Extracted fileSize:', fileSize, 'from eventData.file_size:', eventData.file_size)
console.log('Extracted duration:', duration, 'from eventData.duration:', eventData.duration)
```

#### 2. 计算统计日志

```typescript
console.log('Calculating page stats, tableData length:', tableData.value.length)
tableData.value.forEach((item: any) => {
  console.log('Item fileSize:', item.fileSize, 'duration:', item.duration)
  totalSize += item.fileSize || 0
  totalDuration += item.duration || 0
})
console.log('Total size:', totalSize, 'Total duration:', totalDuration)
```

### 如何查看调试信息

1. 打开浏览器开发者工具（F12）
2. 切换到 Console 标签
3. 刷新录制列表页面
4. 查看控制台输出

### 需要检查的信息

1. **eventData 中的原始值**：
   - `eventData.file_size` 是什么？
   - `eventData.duration` 是什么？
   - 是数字还是字符串？

2. **提取后的值**：
   - `fileSize` 是什么？
   - `duration` 是什么？

3. **tableData 中的值**：
   - 每个 item 的 `fileSize` 和 `duration` 是什么？
   - 是否为 0、null 或 undefined？

4. **计算结果**：
   - `totalSize` 是什么？
   - `totalDuration` 是什么？

## 可能的解决方案

### 方案1：确保数值类型

如果字段是字符串，需要转换为数字：

```typescript
const fileSize = Number(eventData.file_size || eventData.fileSize || 0)
const duration = Number(eventData.duration || 0)
```

### 方案2：检查字段名

如果实际字段名不同，添加更多备选：

```typescript
const fileSize = eventData.file_size
  ?? eventData.fileSize
  ?? eventData.size
  ?? eventData.file_length
  ?? 0

const duration = eventData.duration
  ?? eventData.time
  ?? eventData.length
  ?? 0
```

### 方案3：从其他字段计算

如果 `duration` 不存在，可以从开始和结束时间计算：

```typescript
let duration = eventData.duration ?? 0
if (duration === 0 && startTime != null && endTime != null) {
  duration = endTime - startTime
}
```

## 格式化函数检查

### formatFileSize

```typescript
const formatFileSize = (bytes: number) => {
  if (bytes == null || bytes === 0) return '-'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}
```

**问题**：如果 `bytes` 为 0，返回 `-`，这是正确的。

### formatTotalDuration

```typescript
const formatTotalDuration = (seconds: number) => {
  if (seconds == null || seconds === 0) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  if (hours > 0) {
    return `${hours}小时${minutes}分`
  } else if (minutes > 0) {
    return `${minutes}分`
  } else {
    return `${seconds}秒`
  }
}
```

**问题**：如果 `seconds` 为 0，返回 `-`，这是正确的。

## 预期结果

如果数据正确，应该显示：

### 示例1：单个文件

- 文件大小：43911385 字节 = 41.88 MB
- 时长：56 秒

**显示**：
- 当前页文件大小：41.88 MB
- 当前页总时长：56秒

### 示例2：多个文件

假设有3个文件：
- 文件1：10 MB，60秒
- 文件2：20 MB，120秒
- 文件3：30 MB，180秒

**显示**：
- 当前页文件大小：60.00 MB
- 当前页总时长：6分0秒

## 临时解决方案

如果调试后发现数据确实为0，可以暂时隐藏这两个统计卡片，或者显示提示信息：

```vue
<a-card class="stat-card">
  <a-statistic title="当前页文件大小" :value="0">
    <template #prefix>
      <icon-storage />
    </template>
    <template #value>
      <span v-if="stats.totalSize > 0">
        {{ formatFileSize(stats.totalSize) }}
      </span>
      <span v-else style="color: #999;">
        暂无数据
      </span>
    </template>
  </a-statistic>
</a-card>
```

## 数据库检查

如果前端数据为0，检查数据库中的原始数据：

```sql
SELECT
    ID,
    EVENT_TYPE,
    EVENT_DATA
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 1;
```

查看 `EVENT_DATA` 字段中是否包含 `file_size` 和 `duration`。

## 后端检查

检查后端保存时是否正确解析了这些字段：

```go
logger.Info("收到录制文件事件",
    zap.String("streamId", event.StreamID),
    zap.Int64("fileSize", event.FileSize),    // 检查这个值
    zap.Int64("duration", event.Duration),    // 检查这个值
    zap.String("format", event.FileFormat))
```

## 总结

通过添加详细的调试日志，我们可以追踪数据从后端到前端的整个流程：

1. **后端接收** → 检查后端日志中的 fileSize 和 duration
2. **数据库存储** → 检查 EVENT_DATA 中的 file_size 和 duration
3. **前端接收** → 检查 eventData 中的原始值
4. **数据提取** → 检查提取后的 fileSize 和 duration
5. **统计计算** → 检查 totalSize 和 totalDuration
6. **页面显示** → 检查最终显示的值

找到数据为0的环节，就能确定问题所在并修复。

## 下一步

请查看浏览器控制台的日志输出，并告诉我：
1. `eventData.file_size` 和 `eventData.duration` 的值是什么？
2. 提取后的 `fileSize` 和 `duration` 是什么？
3. 计算后的 `totalSize` 和 `totalDuration` 是什么？

根据这些信息，我可以提供具体的修复方案。

---

## STATS_FIX

## 问题描述

直播录制列表页面的统计数据显示异常，主要问题：

1. **总录制数** - 显示的是所有数据，没有考虑筛选条件
2. **总文件大小** - 只计算当前页数据，不是全部数据
3. **总时长** - 只计算当前页数据，不是全部数据
4. **统计不准确** - 用户筛选后，统计数据没有相应更新

## 修复方案

### 1. 统计数据分类

将统计数据分为两类：

#### 全局统计（基于筛选条件）
- **总录制数（筛选后）** - 根据当前筛选条件统计总数
- **今日录制** - 今天的录制总数（不受筛选影响）

#### 当前页统计
- **当前页文件大小** - 当前页面显示的文件总大小
- **当前页总时长** - 当前页面显示的录制总时长

### 2. 修复内容

#### 修复前
```typescript
// 加载统计数据
const loadStats = async () => {
  // 总数查询不考虑筛选条件
  const totalRes = await queryCallbackEvents({
    eventType: 'recording_file',
    pageNum: 1,
    pageSize: 1
  })

  // 只计算当前页数据
  tableData.value.forEach((item: any) => {
    totalSize += item.fileSize || 0
    totalDuration += item.duration || 0
  })
}
```

#### 修复后
```typescript
// 加载统计数据
const loadStats = async () => {
  // 构建统计查询参数（与当前筛选条件一致）
  const statsParams: any = {
    eventType: 'recording_file'
  }

  // 应用当前的筛选条件
  if (searchForm.streamName) {
    statsParams.streamName = searchForm.streamName
  }
  if (searchForm.domainName) {
    statsParams.domainName = searchForm.domainName
  }
  if (searchForm.appName) {
    statsParams.appName = searchForm.appName
  }
  if (searchForm.timeRange && searchForm.timeRange.length === 2) {
    statsParams.startTime = dayjs(searchForm.timeRange[0]).format('YYYY-MM-DD HH:mm:ss')
    statsParams.endTime = dayjs(searchForm.timeRange[1]).format('YYYY-MM-DD HH:mm:ss')
  }

  // 加载总数（基于当前筛选条件）
  const totalRes = await queryCallbackEvents({
    ...statsParams,
    pageNum: 1,
    pageSize: 1
  })

  // 计算当前页数据
  tableData.value.forEach((item: any) => {
    totalSize += item.fileSize || 0
    totalDuration += item.duration || 0
  })
}
```

### 3. UI 更新

#### 统计卡片标题更新

| 修复前 | 修复后 | 说明 |
|--------|--------|------|
| 总录制数 | 总录制数（筛选后） | 明确说明是基于筛选条件的统计 |
| 今日录制 | 今日录制 | 保持不变 |
| 总文件大小 | 当前页文件大小 | 明确说明只统计当前页 |
| 总时长 | 当前页总时长 | 明确说明只统计当前页 |

#### 添加图标

为每个统计卡片添加图标，提升视觉效果：

- 📄 总录制数 - `IconFile`
- 📅 今日录制 - `IconCalendar`
- 💾 当前页文件大小 - `IconStorage`
- ⏱️ 当前页总时长 - `IconClockCircle`

## 统计逻辑说明

### 总录制数（筛选后）

**计算方式**：
- 根据当前的筛选条件（流名称、域名、应用名称、时间范围）
- 查询符合条件的录制总数
- 用户修改筛选条件后，统计数据会相应更新

**示例**：
```
筛选条件：时间范围 = 最近7天
总录制数（筛选后）= 最近7天的录制总数
```

### 今日录制

**计算方式**：
- 固定查询今天（00:00:00 至当前时间）的录制数
- 不受用户筛选条件影响
- 始终显示今天的录制总数

**示例**：
```
今日录制 = 今天 00:00:00 至现在的录制总数
```

### 当前页文件大小

**计算方式**：
- 累加当前页面显示的所有录制文件大小
- 只统计当前页（如第1页的20条数据）
- 翻页后会重新计算

**示例**：
```
当前页显示 20 条录制
文件大小分别为：100MB, 200MB, 150MB, ...
当前页文件大小 = 所有文件大小之和
```

### 当前页总时长

**计算方式**：
- 累加当前页面显示的所有录制时长
- 只统计当前页（如第1页的20条数据）
- 翻页后会重新计算

**示例**：
```
当前页显示 20 条录制
时长分别为：3600秒, 7200秒, 1800秒, ...
当前页总时长 = 所有时长之和
```

## 为什么不统计全部数据的大小和时长？

### 性能考虑

如果要统计全部数据的文件大小和时长，需要：

1. **查询所有数据** - 可能有成千上万条记录
2. **解析所有 JSON** - 每条记录的 `eventData` 都需要解析
3. **累加计算** - 遍历所有数据进行累加
4. **响应时间长** - 用户等待时间过长

### 替代方案

如果确实需要全部数据的统计，可以考虑：

#### 方案一：后端聚合查询

在后端添加专门的统计接口：

```go
// 统计接口
GET /api/v1/live/callback/events/stats

// 返回数据
{
  "total": 1000,
  "totalSize": 10737418240,  // 10GB
  "totalDuration": 3600000,   // 1000小时
  "today": 50
}
```

#### 方案二：数据库聚合

使用数据库的聚合函数：

```sql
SELECT
  COUNT(*) as total,
  SUM(JSON_EXTRACT(EVENT_DATA, '$.file_size')) as total_size,
  SUM(JSON_EXTRACT(EVENT_DATA, '$.duration')) as total_duration
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
  AND CREATE_TIME >= '2026-01-26 00:00:00'
  AND CREATE_TIME <= '2026-02-02 23:59:59';
```

#### 方案三：定时统计

使用定时任务定期计算统计数据，存储到缓存或数据库：

```go
// 每小时执行一次
func calculateStats() {
  // 计算统计数据
  stats := calculateRecordingStats()

  // 存储到 Redis
  redis.Set("recording_stats", stats, 1*time.Hour)
}
```

## 用户体验优化

### 1. 清晰的标题

通过标题明确告知用户统计的范围：
- "总录制数（筛选后）" - 用户知道这是基于筛选条件的
- "当前页文件大小" - 用户知道这只是当前页的统计

### 2. 图标提示

添加图标让统计卡片更直观：
- 📄 文件图标 - 录制数量
- 📅 日历图标 - 今日数据
- 💾 存储图标 - 文件大小
- ⏱️ 时钟图标 - 时长

### 3. 实时更新

统计数据会在以下情况下更新：
- 页面首次加载
- 用户点击"查询"按钮
- 用户点击"重置"按钮
- 用户翻页

## 测试验证

### 1. 筛选条件测试

```
步骤：
1. 打开录制列表页面
2. 记录"总录制数（筛选后）"的值
3. 修改时间范围为"最近1天"
4. 点击"查询"按钮
5. 验证"总录制数（筛选后）"是否减少

预期结果：
总录制数应该减少，因为时间范围缩小了
```

### 2. 当前页统计测试

```
步骤：
1. 打开录制列表页面
2. 记录"当前页文件大小"和"当前页总时长"
3. 点击"下一页"
4. 验证统计数据是否变化

预期结果：
当前页的统计数据应该重新计算
```

### 3. 今日录制测试

```
步骤：
1. 打开录制列表页面
2. 记录"今日录制"的值
3. 修改时间范围为"最近30天"
4. 点击"查询"按钮
5. 验证"今日录制"是否保持不变

预期结果：
今日录制不受筛选条件影响，应该保持不变
```

## 相关文件

- `sky-web/src/pages/LiveRecordings.vue` - 录制列表页面

## 更新日期

2026-02-02

---

## TEST

本文档提供了测试各类回调事件的curl命令示例。

## 环境变量

```bash
export BASE_URL="http://localhost:9090"
export CALLBACK_KEY="your_callback_key"
export TIMESTAMP=$(date +%s)
```

## 1. 推流事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/push" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 1,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "sequence": "seq_001",
    "node": "192.168.1.1",
    "user_ip": "1.2.3.4",
    "stream_param": "",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 2. 断流事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/disconnect" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 0,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "duration": 3600,
    "reason": "normal",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 3. 录制文件事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 4. 截图事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/screenshot" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 200,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "pic_url": "https://example.com/screenshot.jpg",
    "width": 1920,
    "height": 1080,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 5. 画面审核事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/video-audit" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 317,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "confidence": 95,
    "label": "normal",
    "suggestion": "pass",
    "screenshot_url": "https://example.com/audit.jpg",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 6. 音频审核事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/audio-audit" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 318,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "confidence": 90,
    "label": "normal",
    "suggestion": "pass",
    "audio_text": "测试音频内容",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 7. 质检事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/quality-inspection" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 319,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "diagnose_type": "video_freeze",
    "level": "warning",
    "description": "检测到视频卡顿",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 8. 评测阈值事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/quality-threshold" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 320,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "metric_type": "bitrate",
    "threshold": 2000.0,
    "current_value": 1500.0,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 9. 评测平均分事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/quality-average" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 321,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "score": 85.5,
    "duration": 3600,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 10. 智能擦除事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/smart-erase" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 322,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "task_id": "task_001",
    "status": "success",
    "output_url": "https://example.com/erased.mp4",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 11. 直播字幕事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/subtitle" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 323,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "text": "这是实时生成的字幕内容",
    "language": "zh-CN",
    "start_time": '${TIMESTAMP}',
    "end_time": '$((TIMESTAMP + 5))',
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 12. 直播摘要事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/summary" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 324,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "summary": "本次直播的主要内容摘要",
    "keywords": ["关键词1", "关键词2", "关键词3"],
    "duration": 3600,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 13. 高光切片事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/highlight" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 325,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "clip_url": "https://example.com/highlight.mp4",
    "start_time": '${TIMESTAMP}',
    "end_time": '$((TIMESTAMP + 30))',
    "score": 9.5,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 14. 推流异常事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/push-exception" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 326,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "error_code": 1001,
    "error_msg": "推流连接超时",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 15. 录制异常事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/record-exception" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 327,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "error_code": 2001,
    "error_msg": "录制存储空间不足",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 16. 拉流转推事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/pull-stream" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 328,
    "task_id": "task_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "status": "start",
    "source_url": "rtmp://source.example.com/live/stream",
    "target_url": "rtmp://target.example.com/live/stream",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 17. 监播事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/monitor" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 329,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "alert_type": "black_screen",
    "alert_level": "error",
    "description": "检测到黑屏",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 查询回调事件

```bash
# 需要先获取JWT token
export TOKEN="your_jwt_token"

# 查询所有事件
curl -X GET "${BASE_URL}/api/v1/live/callback/events" \
  -H "Authorization: Bearer ${TOKEN}"

# 查询特定类型的事件
curl -X GET "${BASE_URL}/api/v1/live/callback/events?eventType=push_stream" \
  -H "Authorization: Bearer ${TOKEN}"

# 查询特定流的事件
curl -X GET "${BASE_URL}/api/v1/live/callback/events?streamId=test_stream_001" \
  -H "Authorization: Bearer ${TOKEN}"

# 分页查询
curl -X GET "${BASE_URL}/api/v1/live/callback/events?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer ${TOKEN}"
```

## 批量测试脚本

创建一个测试脚本 `test_callbacks.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:9090"
TIMESTAMP=$(date +%s)

echo "测试推流事件..."
curl -s -X POST "${BASE_URL}/api/v1/live/callback/push" \
  -H "Content-Type: application/json" \
  -d "{\"event_type\":1,\"stream_id\":\"test_001\",\"t\":${TIMESTAMP},\"sign\":\"test\",\"event_time\":${TIMESTAMP},\"push_domain\":\"push.example.com\",\"app_name\":\"live\",\"stream_name\":\"test\"}" \
  | jq .

echo "测试断流事件..."
curl -s -X POST "${BASE_URL}/api/v1/live/callback/disconnect" \
  -H "Content-Type: application/json" \
  -d "{\"event_type\":0,\"stream_id\":\"test_001\",\"t\":${TIMESTAMP},\"sign\":\"test\",\"event_time\":${TIMESTAMP},\"duration\":3600,\"push_domain\":\"push.example.com\",\"app_name\":\"live\",\"stream_name\":\"test\"}" \
  | jq .

echo "测试截图事件..."
curl -s -X POST "${BASE_URL}/api/v1/live/callback/screenshot" \
  -H "Content-Type: application/json" \
  -d "{\"event_type\":200,\"stream_id\":\"test_001\",\"t\":${TIMESTAMP},\"sign\":\"test\",\"event_time\":${TIMESTAMP},\"pic_url\":\"https://example.com/pic.jpg\",\"width\":1920,\"height\":1080,\"push_domain\":\"push.example.com\",\"app_name\":\"live\",\"stream_name\":\"test\"}" \
  | jq .

echo "所有测试完成！"
```

运行测试：

```bash
chmod +x test_callbacks.sh
./test_callbacks.sh
```

---

## UI_COMPACT

## 优化内容

对高光切片列表页面进行了紧凑化优化，提升信息密度和用户体验。

## 修改文件

- `sky-web/src/pages/LiveHighlightClips.vue`

## 主要改动

### 1. 表格优化

**调整列宽和顺序：**

| 修改前 | 修改后 | 说明 |
|--------|--------|------|
| ID: 80px | ID: 60px | 缩小ID列宽度 |
| 流名称: 150px | 移除 | 流名称通常为空，移除此列 |
| 域名: 180px | 域名: 160px | 缩小域名列宽度 |
| 应用名称: 120px | 应用: 80px | 缩短标题和列宽 |
| 标题: 250px | 标题: 280px | 增加标题列宽度（更重要） |
| 关键词: 200px | 关键词: 220px | 略微增加关键词列宽 |
| 切片时长: 120px | 时长: 90px | 缩短标题和列宽 |
| 生成时间: 180px | 生成时间: 160px | 缩小时间列宽度 |
| 操作: 200px | 操作: 160px | 缩小操作列宽度 |

**添加表格属性：**
```vue
<a-table
  size="small"              <!-- 紧凑尺寸 -->
  :scroll="{ x: 1400 }"     <!-- 横向滚动 -->
>
```

### 2. 搜索栏优化

**表单尺寸：**
```vue
<!-- 修改前 -->
<a-form :model="searchForm" layout="inline">
  <a-input style="width: 200px" />
  <a-range-picker style="width: 380px" />
  <a-button type="primary">查询</a-button>
</a-form>

<!-- 修改后 -->
<a-form :model="searchForm" layout="inline" size="small">
  <a-input style="width: 160px" />
  <a-range-picker style="width: 340px" />
  <a-button type="primary" size="small">查询</a-button>
</a-form>
```

**改进点：**
- ✅ 表单尺寸改为 `small`
- ✅ 输入框宽度从 200px 缩小到 160px
- ✅ 应用名称输入框缩小到 120px
- ✅ 时间选择器从 380px 缩小到 340px
- ✅ 按钮尺寸改为 `small`

### 3. 间距优化

**页面间距：**

| 元素 | 修改前 | 修改后 | 减少 |
|------|--------|--------|------|
| 页面内边距 | 20px | 12px | -40% |
| 标题下边距 | 20px | 12px | -40% |
| 搜索栏内边距 | 20px | 12px 16px | -40% |
| 搜索栏下边距 | 16px | 12px | -25% |
| 预览信息上边距 | 20px | 16px | -20% |

**表格单元格间距：**
```less
.arco-table-th,
.arco-table-td {
  padding: 8px 12px;  // 从默认的 12px 16px 缩小
}
```

### 4. 字体和标签优化

**字体大小：**
```less
// 表格字体
.arco-table {
  font-size: 13px;  // 从默认的 14px 缩小
}

// 标签字体
.arco-tag {
  font-size: 12px;  // 从默认的 13px 缩小
  padding: 0 6px;   // 从默认的 0 8px 缩小
  line-height: 20px; // 从默认的 22px 缩小
}

// 链接字体
.arco-link {
  font-size: 13px;  // 从默认的 14px 缩小
}
```

**标题字体：**
```less
h2 {
  font-size: 18px;  // 从 20px 缩小
}
```

### 5. 操作按钮优化

**按钮间距：**
```vue
<!-- 修改前 -->
<a-space>
  <a-link>预览</a-link>
  <a-link>下载</a-link>
  <a-link>详情</a-link>
</a-space>

<!-- 修改后 -->
<a-space size="mini">
  <a-link>预览</a-link>
  <a-link>下载</a-link>
  <a-link>详情</a-link>
</a-space>
```

**关键词标签间距：**
```vue
<a-space wrap size="mini">
  <a-tag size="small">关键词</a-tag>
</a-space>
```

### 6. 表头样式优化

```less
.arco-table-th {
  background-color: #f7f8fa;  // 浅灰色背景
  font-weight: 600;           // 加粗字体
}
```

## 优化效果对比

### 修改前
```
┌─────────────────────────────────────────────────────────────────┐
│ 直播高光切片                                          [刷新]      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│ 流名称: [____] 域名: [____] 应用: [____] 时间: [________]       │
│ [查询] [重置]                                                    │
│                                                                  │
├──┬────┬────┬────┬────────┬────────┬────┬────────┬──────────┤
│ID│流名│域名│应用│  标题  │ 关键词 │时长│生成时间│  操作    │
├──┼────┼────┼────┼────────┼────────┼────┼────────┼──────────┤
│  │    │    │    │        │        │    │        │          │
│  │    │    │    │        │        │    │        │          │
└──┴────┴────┴────┴────────┴────────┴────┴────────┴──────────┘
```

### 修改后
```
┌───────────────────────────────────────────────────────────────┐
│ 直播高光切片                                      [刷新]        │
├───────────────────────────────────────────────────────────────┤
│ 流名称:[__] 域名:[__] 应用:[_] 时间:[______] [查询][重置]    │
├─┬──────────┬──────────┬────┬──┬───┬────────┬────────────┤
│ID│  标题    │  关键词  │域名│应│时│生成时间│   操作     │
├─┼──────────┼──────────┼────┼──┼───┼────────┼────────────┤
│ │          │          │    │  │  │        │            │
│ │          │          │    │  │  │        │            │
└─┴──────────┴──────────┴────┴──┴───┴────────┴────────────┘
```

## 优化收益

### 空间利用率提升

- **页面内边距减少 40%**：从 20px 减少到 12px
- **搜索栏高度减少约 20%**：通过 small 尺寸和紧凑间距
- **表格行高减少约 15%**：通过 small 尺寸和紧凑单元格
- **整体可视区域增加约 25%**：可以显示更多数据行

### 信息密度提升

- **移除冗余列**：移除通常为空的"流名称"列
- **优化列宽**：根据内容重要性调整列宽
- **标题列增加**：从 250px 增加到 280px，显示更多标题内容
- **关键词列增加**：从 200px 增加到 220px，显示更多关键词

### 用户体验提升

- **减少滚动**：一屏可以显示更多数据
- **快速扫描**：紧凑布局便于快速浏览
- **重点突出**：标题和关键词列更宽，更容易识别
- **操作便捷**：操作按钮仍然清晰可点击

## 响应式支持

添加了横向滚动支持，确保在小屏幕上也能正常显示：

```vue
<a-table :scroll="{ x: 1400 }">
```

当屏幕宽度小于 1400px 时，表格会出现横向滚动条。

## 兼容性说明

所有修改都使用 Arco Design 的标准属性和样式，完全兼容：

- ✅ Arco Design Vue 组件库
- ✅ 现代浏览器（Chrome, Firefox, Safari, Edge）
- ✅ 响应式布局
- ✅ 深色模式（如果启用）

## 后续优化建议

### 1. 添加列配置功能

允许用户自定义显示哪些列：

```vue
<a-button @click="showColumnSettings">
  <icon-settings />
  列设置
</a-button>

<a-modal v-model:visible="columnSettingsVisible">
  <a-checkbox-group v-model="visibleColumns">
    <a-checkbox value="id">ID</a-checkbox>
    <a-checkbox value="title">标题</a-checkbox>
    <a-checkbox value="keywords">关键词</a-checkbox>
    <!-- ... -->
  </a-checkbox-group>
</a-modal>
```

### 2. 添加密度切换

允许用户在紧凑、标准、宽松三种密度间切换：

```vue
<a-radio-group v-model="tableDensity" size="small">
  <a-radio value="small">紧凑</a-radio>
  <a-radio value="medium">标准</a-radio>
  <a-radio value="large">宽松</a-radio>
</a-radio-group>

<a-table :size="tableDensity">
```

### 3. 添加快捷筛选

在表头添加快捷筛选按钮：

```vue
<a-table-column title="域名">
  <template #title>
    域名
    <a-dropdown>
      <icon-filter />
      <template #content>
        <a-doption>全部</a-doption>
        <a-doption>upload.skyzhou.cn</a-doption>
      </template>
    </a-dropdown>
  </template>
</a-table-column>
```

### 4. 添加批量操作

支持批量下载、批量删除等操作：

```vue
<a-table
  :row-selection="{
    type: 'checkbox',
    showCheckedAll: true
  }"
  @select="handleSelect"
>
```

## 相关文件

- `sky-web/src/pages/LiveHighlightClips.vue` - 高光切片列表页面

## 更新日期

2026-02-02

---

## VIDEO_403_FIX

## 问题描述

高光切片预览时，视频无法播放，控制台显示：
```
Failed to load resource: the server responded with a status of 403 (Forbidden)
Video error code: 4
Video error message: MEDIA_ELEMENT_ERROR: Format error
```

## 问题原因

腾讯云 COS 存储桶配置为私有访问，视频 URL 需要签名才能访问。

**视频URL示例：**
```
http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/0202/LIVE_0228577F359F9989378520BF9280AA1274-1770025101592/2026-02-02_18-12-45_1998749.mp4
```

## 解决方案

### 方案 1：配置 COS Bucket 为公共读（推荐 - 最简单）

如果视频内容是公开的，可以直接配置 COS bucket 权限。

#### 步骤：

1. **登录腾讯云控制台**
   - 访问：https://console.cloud.tencent.com/cos
   - 找到 bucket：`video-1301212747`

2. **配置存储桶权限**
   - 进入"权限管理" -> "存储桶访问权限"
   - 设置为"公有读私有写"
   - 或添加策略：
     ```json
     {
       "version": "2.0",
       "statement": [
         {
           "effect": "allow",
           "principal": {
             "qcs": ["qcs::cam::anyone:anyone"]
           },
           "action": [
             "name/cos:GetObject",
             "name/cos:HeadObject"
           ],
           "resource": [
             "qcs::cos:ap-nanjing:uid/1301212747:video-1301212747/*"
           ]
         }
       ]
     }
     ```

3. **配置 CORS 规则**
   - 进入"安全管理" -> "跨域访问CORS设置"
   - 添加规则：
     ```
     来源 Origin: *
     操作 Methods: GET, HEAD
     Allow-Headers: *
     Expose-Headers: ETag
     超时 Max-Age: 3600
     ```

4. **保存并测试**
   - 保存配置后，等待几分钟生效
   - 在浏览器新标签页直接访问视频URL，应该能播放

### 方案 2：后端生成签名 URL（推荐 - 安全）

如果视频需要保持私有，由后端生成带签名的临时访问 URL。

#### 后端实现：

1. **安装 COS SDK**
   ```bash
   go get -u github.com/tencentyun/cos-go-sdk-v5
   ```

2. **配置文件添加 COS 凭证**
   ```yaml
   # config.yaml
   tencentCloud:
     secretId: "your-secret-id"
     secretKey: "your-secret-key"
     cosRegion: "ap-nanjing"
   ```

3. **实现签名接口**
   ```go
   // 生成预签名URL（有效期1小时）
   func (h *LiveCallbackHandler) GetVideoSignedURL(c *gin.Context) {
       videoURL := c.Query("videoUrl")

       // 使用COS SDK生成预签名URL
       presignedURL, err := h.cosClient.Object.GetPresignedURL(
           context.Background(),
           http.MethodGet,
           videoURL,
           h.secretId,
           h.secretKey,
           time.Hour,
           nil,
       )

       c.JSON(200, gin.H{
           "code": 200,
           "data": gin.H{
               "signedUrl": presignedURL.String(),
           },
       })
   }
   ```

4. **添加路由**
   ```go
   // live_callback_router.go
   public.GET("/video/signed-url", handler.GetVideoSignedURL)
   ```

#### 前端实现：

```typescript
// src/api/live.ts
export const getVideoSignedURL = (videoUrl: string) => {
  return api.get('/live/video/signed-url', {
    params: { videoUrl }
  })
}

// LiveHighlightClips.vue
const handlePreview = async (record: any) => {
  try {
    // 获取签名URL
    const res = await getVideoSignedURL(record.clipUrl)
    if (res.data?.code === 200) {
      currentClip.value = {
        ...record,
        clipUrl: res.data.data.signedUrl  // 使用签名URL
      }
      previewVisible.value = true
    }
  } catch (error) {
    Message.error('获取视频签名失败')
  }
}
```

### 方案 3：使用腾讯云 CDN

如果视频访问量大，建议配置 CDN 加速。

#### 步骤：

1. **开通 CDN 服务**
   - 在腾讯云控制台开通 CDN
   - 添加加速域名

2. **配置回源**
   - 回源地址：`video-1301212747.cos.ap-nanjing.myqcloud.com`
   - 回源协议：HTTP/HTTPS

3. **配置鉴权**
   - 如果需要防盗链，配置 URL 鉴权
   - 设置鉴权密钥和有效期

4. **更新视频URL**
   - 将 COS URL 替换为 CDN URL
   - 例如：`https://video-cdn.yourdomain.com/...`

## 临时解决方案

在配置 COS 权限之前，可以使用以下临时方案：

### 方法 1：直接在新窗口打开

```vue
<!-- 修改预览对话框 -->
<a-modal
  v-model:visible="previewVisible"
  title="高光切片预览"
  :width="800"
  :footer="false"
>
  <div class="preview-container">
    <a-alert type="warning" style="margin-bottom: 16px;">
      由于视频权限限制，请点击下方链接在新窗口中播放
    </a-alert>
    <a-button type="primary" @click="openVideoInNewTab">
      在新窗口中播放视频
    </a-button>
    <!-- 其他信息 -->
  </div>
</a-modal>

<script>
const openVideoInNewTab = () => {
  if (currentClip.value?.clipUrl) {
    window.open(currentClip.value.clipUrl, '_blank')
  }
}
</script>
```

### 方法 2：使用下载链接

```vue
<a-button type="primary" @click="handleDownload(currentClip)">
  下载视频
</a-button>
```

## 推荐方案

根据您的使用场景选择：

1. **公开视频** → 使用方案 1（配置 COS 为公共读）
   - ✅ 最简单，无需修改代码
   - ✅ 性能最好
   - ❌ 视频完全公开

2. **私有视频** → 使用方案 2（后端签名）
   - ✅ 安全可控
   - ✅ 可以设置访问有效期
   - ❌ 需要后端支持

3. **高访问量** → 使用方案 3（CDN）
   - ✅ 访问速度快
   - ✅ 降低 COS 成本
   - ❌ 配置较复杂

## 测试验证

配置完成后，测试步骤：

1. **直接访问测试**
   - 在浏览器新标签页打开视频URL
   - 应该能直接播放或下载

2. **前端预览测试**
   - 刷新页面
   - 点击"预览"按钮
   - 视频应该能正常播放

3. **控制台检查**
   - 打开开发者工具
   - Network 标签应该显示视频请求成功（200 OK）
   - 不应该有 403 错误

## 相关文档

- [腾讯云 COS 权限管理](https://cloud.tencent.com/document/product/436/12470)
- [腾讯云 COS 预签名 URL](https://cloud.tencent.com/document/product/436/35059)
- [腾讯云 CDN 配置](https://cloud.tencent.com/document/product/228/41867)

## 更新日期

2026-02-02

---

