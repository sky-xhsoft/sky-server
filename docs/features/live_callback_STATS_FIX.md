# 直播录制页面统计数据修复说明

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
