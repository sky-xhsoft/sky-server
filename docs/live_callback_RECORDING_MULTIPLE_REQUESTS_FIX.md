# 录制列表多次请求问题修复

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
