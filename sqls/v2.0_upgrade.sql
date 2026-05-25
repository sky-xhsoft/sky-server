-- ============================================================================
-- Sky-Server v2.0 升级脚本
-- 版本: v1.0.2 -> v2.0
-- 日期: 2026-05-22
-- 说明: 此脚本为 v2.0 重构配套的最小化数据库升级。
--       v2.0 主要变更在代码架构层面（Wire DI、领域模型、高级搜索），
--       数据库变更需求极少，依赖已有的 SET_VALUE_TYPE='fk' 和 IS_QUERY 字段。
-- ============================================================================

-- ============================================================================
-- 1. 扩展 PROPS 字段容量（支持高级搜索配置存储）
-- ============================================================================
ALTER TABLE `sys_table`
    MODIFY COLUMN `PROPS` TEXT DEFAULT NULL COMMENT '扩展属性（JSON格式），可存储搜索配置';

ALTER TABLE `sys_column`
    MODIFY COLUMN `PROPS` TEXT DEFAULT NULL COMMENT '扩展属性（JSON格式），可存储字段级搜索配置';

-- ============================================================================
-- 2. 确保外键相关字段已存在（v1.0.2 已包含）
--    无需操作：SET_VALUE_TYPE / REF_TABLE_ID / REF_COLUMN_ID / REF_ON_DELETE 已存在
-- ============================================================================

-- ============================================================================
-- 3. 确保搜索相关字段已存在（v1.0.2 已包含）
--    无需操作：IS_QUERY 字段已存在
--    -- 标记字段为可搜索的示例：
--    -- UPDATE sys_column SET IS_QUERY = 'Y' WHERE ...;
-- ============================================================================

-- ============================================================================
-- 4. 高级搜索配置示例（通过 PROPS JSON 配置，无需新增字段）
-- ============================================================================

-- 示例1：为表配置默认搜索行为
-- UPDATE sys_table SET PROPS = JSON_OBJECT(
--     'searchFields', JSON_ARRAY('NAME', 'DESCRIPTION', 'CREATE_TIME'),
--     'defaultOperator', 'contains',
--     'maxConditions', 10
-- ) WHERE NAME = 'your_table';

-- 示例2：为字段配置支持的搜索运算符
-- UPDATE sys_column SET PROPS = JSON_OBJECT(
--     'operators', JSON_ARRAY('eq', 'ne', 'contains', 'startsWith', 'endsWith', 'between'),
--     'defaultOperator', 'contains'
-- ) WHERE DB_NAME = 'NAME' AND SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'your_table');

-- ============================================================================
-- 5. 外键配置示例（无需新增字段，使用已有 SET_VALUE_TYPE='fk'）
-- ============================================================================

-- 示例：配置外键字段
-- UPDATE sys_column SET
--     SET_VALUE_TYPE = 'fk',
--     REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user'),
--     REF_COLUMN_ID = (SELECT ID FROM sys_column WHERE DB_NAME = 'NAME' AND SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user')),
--     REF_ON_DELETE = 'restrict'
-- WHERE DB_NAME = 'USER_ID' AND SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'your_table');

-- ============================================================================
-- 升级完成
-- 版本信息: 通过 PROPS 字段记录
-- ============================================================================
