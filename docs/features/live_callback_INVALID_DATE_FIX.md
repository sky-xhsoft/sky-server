# 直播录制页面 "Invalid Date" 和数据不显示问题修复

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
