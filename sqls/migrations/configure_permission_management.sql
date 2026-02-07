-- ============================================================================
-- 权限管理模块配置脚本
-- 功能：配置sys_groups、sys_group_prem、sys_user_groups表的元数据
-- 使用配置驱动CRUD框架，通过MetadataListView和MetadataFormView自动生成UI
-- ============================================================================

-- ============================================================================
-- 第一部分：配置sys_table表元数据
-- ============================================================================

-- 1. 配置sys_groups（权限组）表
UPDATE sys_table SET
  DISPLAY_NAME = '权限组管理',
  MASK = 'AMDQE',  -- A:新增, M:修改, D:删除, Q:查询, E:导出
  IS_MENU = 'Y',   -- 显示在菜单中
  DESCRIPTION = '权限组管理，用于定义不同的权限组并分配权限'
WHERE NAME = 'sys_groups';

-- 2. 配置sys_group_prem（权限组明细）表
UPDATE sys_table SET
  DISPLAY_NAME = '权限明细',
  MASK = 'AMDQ',  -- A:新增, M:修改, D:删除, Q:查询
  IS_MENU = 'N',  -- 不单独显示在菜单中，作为sys_groups的子表
  DESCRIPTION = '权限组明细，定义权限组对各个目录的访问权限'
WHERE NAME = 'sys_group_prem';

-- 3. 配置sys_user_groups（用户权限组关联）表
UPDATE sys_table SET
  DISPLAY_NAME = '用户权限组',
  MASK = 'AMDQ',  -- A:新增, M:修改, D:删除, Q:查询
  IS_MENU = 'N',  -- 不单独显示在菜单中，作为sys_user的子表
  DESCRIPTION = '用户权限组关联，定义用户所属的权限组'
WHERE NAME = 'sys_user_groups';

-- ============================================================================
-- 第二部分：配置sys_column字段元数据
-- ============================================================================

-- ============================================================================
-- 2.1 配置sys_groups表字段
-- ============================================================================

-- 删除可能存在的旧配置
DELETE FROM sys_column WHERE SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_groups');

-- ID字段（主键）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  'ID', 'ID', 'text', 'pk',
  'N', 1, 'N', 'int', '1111111111'
);

-- NAME字段（权限组名称）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK, REG_EXPRESSION, ERR_MSG
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  'NAME', '权限组名称', 'text', 'byPage',
  'N', 10, 'Y', 'varchar', '1111111111', '^.{2,50}$', '权限组名称长度必须在2-50个字符之间'
);

-- DESCRIPTION字段（描述）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  'DESCRIPTION', '描述', 'textarea', 'byPage',
  'Y', 20, 'N', 'varchar', '1111111111'
);

-- SGRADE字段（安全等级）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK, DEFAULT_VALUE
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  'SGRADE', '安全等级', 'number', 'byPage',
  'Y', 30, 'Y', 'int', '1111111111', '0'
);

-- 系统字段（CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, IS_ACTIVE）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES
  ((SELECT ID FROM sys_table WHERE NAME='sys_groups'), 'CREATE_BY', '创建人', 'text', 'createBy', 'Y', 1001, 'N', 'varchar', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_groups'), 'CREATE_TIME', '创建时间', 'datetime', 'sysdate', 'Y', 1002, 'N', 'datetime', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_groups'), 'UPDATE_BY', '更新人', 'text', 'operator', 'Y', 1003, 'N', 'varchar', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_groups'), 'UPDATE_TIME', '更新时间', 'datetime', 'sysdate', 'Y', 1004, 'N', 'datetime', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_groups'), 'IS_ACTIVE', '状态', 'check', 'byPage', 'Y', 1005, 'Y', 'varchar', '1111111111');

-- ============================================================================
-- 2.2 配置sys_group_prem表字段
-- ============================================================================

-- 删除可能存在的旧配置
DELETE FROM sys_column WHERE SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_group_prem');

-- ID字段（主键）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_group_prem'),
  'ID', 'ID', 'text', 'pk',
  'N', 1, 'N', 'int', '1111111111'
);

-- SYS_GROUPS_ID字段（外键：权限组）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK,
  REF_TABLE_ID, REF_COLUMN_ID
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_group_prem'),
  'SYS_GROUPS_ID', '权限组', 'fk', 'byPage',
  'N', 10, 'Y', 'int', '1111111111',
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  (SELECT ID FROM sys_column WHERE DB_NAME='ID' AND SYS_TABLE_ID=(SELECT ID FROM sys_table WHERE NAME='sys_groups'))
);

-- SYS_DIRECTORY_ID字段（外键：目录）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK,
  REF_TABLE_ID, REF_COLUMN_ID
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_group_prem'),
  'SYS_DIRECTORY_ID', '目录', 'fk', 'byPage',
  'N', 20, 'Y', 'int', '1111111111',
  (SELECT ID FROM sys_table WHERE NAME='sys_directory'),
  (SELECT ID FROM sys_column WHERE DB_NAME='ID' AND SYS_TABLE_ID=(SELECT ID FROM sys_table WHERE NAME='sys_directory'))
);

-- PERMISSION字段（权限位，使用自定义渲染器）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK, DEFAULT_VALUE, DESCRIPTION
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_group_prem'),
  'PERMISSION', '权限', 'permission_bits', 'byPage',
  'N', 30, 'N', 'int', '1111111111', '0', '权限位：1=读取,2=创建,4=更新,8=删除,16=提交,32=反提交,64=导出,128=导入'
);

-- FILTER_OBJ字段（数据过滤JSON）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK, DESCRIPTION
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_group_prem'),
  'FILTER_OBJ', '数据过滤', 'json', 'byPage',
  'Y', 40, 'N', 'text', '1111111111', '可选的数据过滤条件（JSON格式）'
);

-- 系统字段
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES
  ((SELECT ID FROM sys_table WHERE NAME='sys_group_prem'), 'CREATE_BY', '创建人', 'text', 'createBy', 'Y', 1001, 'N', 'varchar', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_group_prem'), 'CREATE_TIME', '创建时间', 'datetime', 'sysdate', 'Y', 1002, 'N', 'datetime', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_group_prem'), 'UPDATE_BY', '更新人', 'text', 'operator', 'Y', 1003, 'N', 'varchar', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_group_prem'), 'UPDATE_TIME', '更新时间', 'datetime', 'sysdate', 'Y', 1004, 'N', 'datetime', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_group_prem'), 'IS_ACTIVE', '状态', 'check', 'byPage', 'Y', 1005, 'N', 'varchar', '1111111111');

-- ============================================================================
-- 2.3 配置sys_user_groups表字段
-- ============================================================================

-- 删除可能存在的旧配置
DELETE FROM sys_column WHERE SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_user_groups');

-- ID字段（主键）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_user_groups'),
  'ID', 'ID', 'text', 'pk',
  'N', 1, 'N', 'int', '1111111111'
);

-- SYS_USER_ID字段（外键：用户）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK,
  REF_TABLE_ID, REF_COLUMN_ID
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_user_groups'),
  'SYS_USER_ID', '用户', 'fk', 'byPage',
  'N', 10, 'Y', 'int', '1111111111',
  (SELECT ID FROM sys_table WHERE NAME='sys_user'),
  (SELECT ID FROM sys_column WHERE DB_NAME='ID' AND SYS_TABLE_ID=(SELECT ID FROM sys_table WHERE NAME='sys_user'))
);

-- SYS_DIRECTORY_ID字段（外键：权限组目录）
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK,
  REF_TABLE_ID, REF_COLUMN_ID
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_user_groups'),
  'SYS_DIRECTORY_ID', '权限组', 'fk', 'byPage',
  'N', 20, 'Y', 'int', '1111111111',
  (SELECT ID FROM sys_table WHERE NAME='sys_directory'),
  (SELECT ID FROM sys_column WHERE DB_NAME='ID' AND SYS_TABLE_ID=(SELECT ID FROM sys_table WHERE NAME='sys_directory'))
);

-- 系统字段
INSERT INTO sys_column (
  SYS_TABLE_ID, DB_NAME, DISPLAY_NAME, DISPLAY_TYPE, SET_VALUE_TYPE,
  NULL_ABLE, ORDERNO, IS_QUERY, COL_TYPE, MASK
) VALUES
  ((SELECT ID FROM sys_table WHERE NAME='sys_user_groups'), 'CREATE_BY', '创建人', 'text', 'createBy', 'Y', 1001, 'N', 'varchar', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_user_groups'), 'CREATE_TIME', '创建时间', 'datetime', 'sysdate', 'Y', 1002, 'N', 'datetime', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_user_groups'), 'UPDATE_BY', '更新人', 'text', 'operator', 'Y', 1003, 'N', 'varchar', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_user_groups'), 'UPDATE_TIME', '更新时间', 'datetime', 'sysdate', 'Y', 1004, 'N', 'datetime', '1000000000'),
  ((SELECT ID FROM sys_table WHERE NAME='sys_user_groups'), 'IS_ACTIVE', '状态', 'check', 'byPage', 'Y', 1005, 'N', 'varchar', '1111111111');

-- ============================================================================
-- 第三部分：配置sys_table_ref父子表关系
-- ============================================================================

-- 删除可能存在的旧配置
DELETE FROM sys_table_ref WHERE SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_groups') AND REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_group_prem');
DELETE FROM sys_table_ref WHERE SYS_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_user') AND REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME='sys_user_groups');

-- 配置sys_groups的子表：sys_group_prem
INSERT INTO sys_table_ref (
  SYS_TABLE_ID, REF_TABLE_ID, ASSOCTYPE, EDIT_TYPE, ORDERNO, DISPLAY_NAME
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  (SELECT ID FROM sys_table WHERE NAME='sys_group_prem'),
  'n',  -- 1对多关系
  'Y',  -- 标准内嵌编辑
  10,
  '权限明细'
);

-- 配置sys_user的子表：sys_user_groups
INSERT INTO sys_table_ref (
  SYS_TABLE_ID, REF_TABLE_ID, ASSOCTYPE, EDIT_TYPE, ORDERNO, DISPLAY_NAME
) VALUES (
  (SELECT ID FROM sys_table WHERE NAME='sys_user'),
  (SELECT ID FROM sys_table WHERE NAME='sys_user_groups'),
  'n',  -- 1对多关系
  'Y',  -- 标准内嵌编辑
  20,
  '用户权限组'
);

-- ============================================================================
-- 第四部分：配置菜单项（sys_directory）
-- ============================================================================

-- 删除可能存在的旧菜单配置
DELETE FROM sys_directory WHERE NAME IN ('permission_management', 'permission_groups', 'user_management_perm');

-- 创建权限管理父菜单
INSERT INTO sys_directory (
  NAME, DISPLAY_NAME, PARENT_ID, ORDERNO, URL, DESCRIPTION, IS_ACTIVE
) VALUES (
  'permission_management', '权限管理', NULL, 100, NULL, '权限管理模块', 'Y'
);

-- 创建权限组管理子菜单
INSERT INTO sys_directory (
  NAME, DISPLAY_NAME, PARENT_ID, ORDERNO, URL, SYS_TABLE_ID, DESCRIPTION, IS_ACTIVE
) VALUES (
  'permission_groups', '权限组管理',
  (SELECT ID FROM sys_directory WHERE NAME='permission_management'),
  10,
  CONCAT('/metadata/list?tableId=', (SELECT ID FROM sys_table WHERE NAME='sys_groups')),
  (SELECT ID FROM sys_table WHERE NAME='sys_groups'),
  '管理权限组，配置权限组的权限明细',
  'Y'
);

-- 创建用户管理子菜单（如果不存在）
INSERT INTO sys_directory (
  NAME, DISPLAY_NAME, PARENT_ID, ORDERNO, URL, SYS_TABLE_ID, DESCRIPTION, IS_ACTIVE
)
SELECT
  'user_management_perm', '用户管理',
  (SELECT ID FROM sys_directory WHERE NAME='permission_management'),
  20,
  CONCAT('/metadata/list?tableId=', (SELECT ID FROM sys_table WHERE NAME='sys_user')),
  (SELECT ID FROM sys_table WHERE NAME='sys_user'),
  '管理用户，分配用户权限组',
  'Y'
WHERE NOT EXISTS (
  SELECT 1 FROM sys_directory WHERE NAME='user_management_perm'
);

-- ============================================================================
-- 完成
-- ============================================================================

SELECT '权限管理模块配置完成！' AS message;
SELECT '请访问以下URL查看：' AS message;
SELECT CONCAT('/metadata/list?tableId=', ID) AS url, DISPLAY_NAME
FROM sys_table
WHERE NAME IN ('sys_groups', 'sys_user');
