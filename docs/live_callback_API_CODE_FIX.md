# API 响应码兼容性修复

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
