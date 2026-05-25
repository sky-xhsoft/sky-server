
-- 配置 sys_table_ref 为 sys_table 的明细表

-- ============================================================================
-- 第一部分：确保 sys_table 和 sys_table_ref 配置正确
-- ============================================================================

-- 1. 配置 sys_table（系统表）表
UPDATE sys_table SET
  DISPLAY_NAME = '系统表',
  MASK = 'AMDQ',  -- A:新增, M:修改, D:删除, Q:查询
  IS_MENU = 'Y',   -- 显示在菜单中
  DESCRIPTION = '系统表定义，用于配置系统中的表'
WHERE NAME = 'sys_table';

-- 2. 配置 sys_table_ref（表关联关系）表
UPDATE sys_table SET
  DISPLAY_NAME = '表关联关系',
  MASK = 'AMDQ',  -- A:新增, M:修改, D:删除, Q:查询
  IS_MENU = 'N',  -- 不单独显示在菜单中，作为sys_table的子表
  DESCRIPTION = '表关联关系，定义表之间的关联关系'
WHERE NAME = 'sys_table_ref';

-- ============================================================================
-- 第二部分：配置 sys_table_ref 的 sys_column 字段
-- ============================================================================

-- 获取 sys_table_ref 的表ID
SET @table_sys_table_ref_id = (SELECT ID FROM sys_table WHERE NAME = 'sys_table_ref');
SET @table_sys_table_id = (SELECT ID FROM sys_table WHERE NAME = 'sys_table');

-- 更新 sys_table_ref 的 SYS_TABLE_ID 字段，设置为关联 sys_table 的外键
UPDATE sys_column 
SET 
  DISPLAY_TYPE = 'fk',
  REF_TABLE_ID = @table_sys_table_id,
  REF_COLUMN_ID = (SELECT ID FROM sys_column WHERE DB_NAME = 'ID' AND SYS_TABLE_ID = @table_sys_table_id)
WHERE SYS_TABLE_ID = @table_sys_table_ref_id 
  AND DB_NAME = 'SYS_TABLE_ID';

-- ============================================================================
-- 第三部分：配置 sys_table_ref 父子表关系
-- ============================================================================

-- 删除可能存在的旧配置
DELETE FROM sys_table_ref 
WHERE SYS_TABLE_ID = @table_sys_table_id 
  AND REF_TABLE_ID = @table_sys_table_ref_id;

-- 配置 sys_table 的子表：sys_table_ref
INSERT INTO sys_table_ref (
  SYS_TABLE_ID, REF_TABLE_ID, ASSOCTYPE, EDIT_TYPE, ORDERNO, DISPLAY_NAME
) VALUES (
  @table_sys_table_id,
  @table_sys_table_ref_id,
  'n',  -- 1对多关系
  'Y',  -- 标准内嵌编辑
  10,
  '表关联关系'
);

-- ============================================================================
-- 完成
-- ============================================================================

SELECT 'sys_table_ref 配置为 sys_table 的明细表完成！' AS message;

