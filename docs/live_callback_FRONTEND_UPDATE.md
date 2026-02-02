# 高光切片前端页面更新 - 支持新数据格式

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
