-- ========================================
-- ⚠️ 已废弃 - 无需数据库配置 ⚠️
-- ========================================
--
-- 从 v3.0 开始，插件系统采用自动发现机制，无需数据库配置！
--
-- 只需：
-- 1. 将插件放入 plugins/runtime/插件名/ 目录
-- 2. 系统会自动扫描、编译、加载
--
-- 本文件保留仅供参考，实际使用中不需要执行这些 SQL
--
-- ========================================

-- ========================================
-- 方式 1: 使用 sys_action 表配置
-- ========================================
-- 适用场景：通用动作、独立功能插件

-- 示例 1: 订单通知插件（使用 sys_action）
INSERT INTO sys_action (
  NAME,
  ACTION_TYPE,        -- 设置为 'go' 表示热加载插件
  CONTENT,            -- 插件源码路径（相对 plugins/runtime/）
  DISPLAY_NAME,
  SYS_TABLE_ID,       -- 关联的表ID（可选）
  IS_ACTIVE
) VALUES (
  'order_notify_plugin',
  'go',                          -- ACTION_TYPE = 'go'
  'order_notify',                -- 源码路径：plugins/runtime/order_notify/
  '订单通知插件',
  (SELECT ID FROM sys_table WHERE DB_NAME = 'order' LIMIT 1),
  'Y'
);

-- 示例 2: 用户欢迎插件（使用 sys_action）
INSERT INTO sys_action (
  NAME,
  ACTION_TYPE,
  CONTENT,
  DISPLAY_NAME,
  SYS_TABLE_ID,
  IS_ACTIVE
) VALUES (
  'user_welcome',
  'go',
  'user_welcome',
  '用户欢迎插件',
  (SELECT ID FROM sys_table WHERE DB_NAME = 'sys_user' LIMIT 1),
  'Y'
);

-- ========================================
-- 方式 2: 使用 sys_table_cmd 表配置
-- ========================================
-- 适用场景：表级别钩子、before/after 操作

-- 示例 3: 订单创建后通知（使用 sys_table_cmd）
INSERT INTO sys_table_cmd (
  SYS_TABLE_ID,       -- 关联的表ID
  ACTION,             -- 动作类型：A=create, M=update, D=delete, Q=query
  ACTION_NAME,
  EVENT,              -- begin=before, end=after
  CONTENT,            -- 插件源码路径
  CONTENT_TYPE,       -- 设置为 'go' 表示热加载插件
  ORDERNO,            -- 优先级（数字越小越先执行）
  IS_ACTIVE
) VALUES (
  (SELECT ID FROM sys_table WHERE DB_NAME = 'order' LIMIT 1),
  'A',                -- A = create（创建）
  '订单创建后通知',
  'end',              -- end = after（之后）
  'order_after_create',  -- 源码路径
  'go',               -- CONTENT_TYPE = 'go'
  50,                 -- 优先级
  'Y'
);

-- 示例 4: 订单更新前验证（使用 sys_table_cmd）
INSERT INTO sys_table_cmd (
  SYS_TABLE_ID,
  ACTION,
  ACTION_NAME,
  EVENT,
  CONTENT,
  CONTENT_TYPE,
  ORDERNO,
  IS_ACTIVE
) VALUES (
  (SELECT ID FROM sys_table WHERE DB_NAME = 'order' LIMIT 1),
  'M',                -- M = update（更新）
  '订单更新前验证',
  'begin',            -- begin = before（之前）
  'order_before_update',
  'go',
  10,                 -- 更高优先级（先执行）
  'Y'
);

-- 示例 5: 用户删除前检查（使用 sys_table_cmd）
INSERT INTO sys_table_cmd (
  SYS_TABLE_ID,
  ACTION,
  ACTION_NAME,
  EVENT,
  CONTENT,
  CONTENT_TYPE,
  ORDERNO,
  IS_ACTIVE
) VALUES (
  (SELECT ID FROM sys_table WHERE DB_NAME = 'sys_user' LIMIT 1),
  'D',                -- D = delete（删除）
  '用户删除前检查',
  'begin',            -- begin = before（之前）
  'user_before_delete',
  'go',
  5,                  -- 高优先级（先执行）
  'Y'
);

-- ========================================
-- 动作代码说明
-- ========================================
-- ACTION 字段：
--   A = create   (创建)
--   M = update   (更新)
--   D = delete   (删除)
--   Q = query    (查询)
--   S = submit   (提交)
--   U = unsubmit (反提交)
--   V = void     (作废)
--   I = import   (导入)
--   E = export   (导出)

-- EVENT 字段：
--   begin = before (之前)
--   end = after    (之后)

-- ========================================
-- 查询已配置的热加载插件
-- ========================================

-- 查询 sys_action 中的插件
SELECT
  ID,
  NAME as '插件名称',
  CONTENT as '源码路径',
  DISPLAY_NAME as '描述',
  IS_ACTIVE as '状态'
FROM sys_action
WHERE ACTION_TYPE = 'go';

-- 查询 sys_table_cmd 中的插件
SELECT
  c.ID,
  t.DB_NAME as '表名',
  c.ACTION as '动作',
  c.EVENT as '时机',
  c.CONTENT as '源码路径',
  c.ACTION_NAME as '描述',
  c.ORDERNO as '优先级',
  c.IS_ACTIVE as '状态'
FROM sys_table_cmd c
LEFT JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
WHERE c.CONTENT_TYPE = 'go';

-- ========================================
-- 禁用/启用插件
-- ========================================

-- 禁用插件（sys_action）
UPDATE sys_action
SET IS_ACTIVE = 'N'
WHERE NAME = 'order_notify_plugin';

-- 启用插件（sys_action）
UPDATE sys_action
SET IS_ACTIVE = 'Y'
WHERE NAME = 'order_notify_plugin';

-- 禁用插件（sys_table_cmd）
UPDATE sys_table_cmd
SET IS_ACTIVE = 'N'
WHERE CONTENT = 'order_after_create';

-- ========================================
-- 完整示例：创建一个热加载插件
-- ========================================

-- 步骤 1: 创建插件源码（命令行）
-- mkdir -p plugins/runtime/order_notify
-- cp plugins/runtime/TEMPLATE.go plugins/runtime/order_notify/plugin.go
-- vim plugins/runtime/order_notify/plugin.go

-- 步骤 2: 配置到数据库（选择一种方式）

-- 方式 A: 使用 sys_table_cmd（推荐）
INSERT INTO sys_table_cmd (
  SYS_TABLE_ID,
  ACTION,
  ACTION_NAME,
  EVENT,
  CONTENT,
  CONTENT_TYPE,
  ORDERNO,
  IS_ACTIVE
) VALUES (
  (SELECT ID FROM sys_table WHERE DB_NAME = 'order' LIMIT 1),
  'A',
  '订单通知',
  'end',
  'order_notify',     -- 必须与源码目录名一致
  'go',
  50,
  'Y'
);

-- 方式 B: 使用 sys_action
INSERT INTO sys_action (
  NAME,
  ACTION_TYPE,
  CONTENT,
  DISPLAY_NAME,
  IS_ACTIVE
) VALUES (
  'order_notify',
  'go',
  'order_notify',     -- 必须与源码目录名一致
  '订单通知插件',
  'Y'
);

-- 步骤 3: 重启服务器或等待自动加载
-- 系统会自动：
-- 1. 扫描数据库配置
-- 2. 编译源码 → .so 文件
-- 3. 加载到运行时
-- 4. 监听文件变化（支持热重载）

-- 步骤 4: 测试热重载
-- vim plugins/runtime/order_notify/plugin.go  （修改代码）
-- 保存后系统会自动重新编译和加载！
