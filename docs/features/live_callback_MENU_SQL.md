# 直播回调事件页面 - 菜单配置 SQL

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
