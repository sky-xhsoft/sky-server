# 直播切片列表增加开始时间和结束时间列

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
