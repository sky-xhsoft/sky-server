# 直播回调事件页面 - 菜单配置说明

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
