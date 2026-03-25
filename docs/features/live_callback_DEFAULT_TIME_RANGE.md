# 前端页面默认查询时间范围更新

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
