# 直播录制列表当前页统计显示为0问题调试

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
