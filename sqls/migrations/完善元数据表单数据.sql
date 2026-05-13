-- ============================================================
-- 元数据完善: 缺失表的 sys_table/sys_column 数据补全
-- 日期: 2026-05-13
-- ============================================================

-- ============================================================
-- Phase 1: 为12个缺失物理表添加 sys_table 元数据条目
-- ============================================================

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_share', '云盘分享', 'Cloud Share', 'N', 1, 'Y', 'AMDQE')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='云盘分享', ENAME='Cloud Share';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_quota', '云盘配额', 'Cloud Quota', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='云盘配额', ENAME='Cloud Quota';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_upload_session', '分片上传会话', 'Upload Session', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='分片上传会话', ENAME='Upload Session';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_chunk_record', '分片上传记录', 'Chunk Record', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='分片上传记录', ENAME='Chunk Record';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_folder', '云盘文件夹', 'Cloud Folder', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='云盘文件夹', ENAME='Cloud Folder';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_file', '云盘文件', 'Cloud File', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='云盘文件', ENAME='Cloud File';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('live_callback_event', '直播回调事件', 'Live Callback Event', 'N', 1, 'Y', 'Q')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='直播回调事件', ENAME='Live Callback Event';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('pull_stream_task', '拉流任务', 'Pull Stream Task', 'N', 1, 'Y', 'AMDQV')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='拉流任务', ENAME='Pull Stream Task';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('live_push_domain_config', '推流域名配置', 'Push Domain Config', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='推流域名配置', ENAME='Push Domain Config';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('cloud_storage_config', '存储配置', 'Storage Config', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='存储配置', ENAME='Storage Config';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('sys_language', '语言管理', 'Language', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='语言管理', ENAME='Language';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('sys_translation', '翻译管理', 'Translation', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='翻译管理', ENAME='Translation';

INSERT INTO sys_table (NAME, DISPLAY_NAME, ENAME, IS_BIG, ORDERNO, IS_ACTIVE, MASK)
VALUES ('sys_company_conf', '公司配置', 'Company Config', 'N', 1, 'Y', 'AMDQ')
ON DUPLICATE KEY UPDATE DISPLAY_NAME='公司配置', ENAME='Company Config';

-- ============================================================
-- Phase 2: cloud_item 缺少的 21 个业务列元数据
-- ============================================================

-- 先获取 cloud_item 的 SYS_TABLE_ID
SET @cloud_item_id = (SELECT ID FROM sys_table WHERE NAME='cloud_item');

-- cloud_item 业务字段 (所有字段 ORDERNO 从 8 开始，因为1-7是标准审计字段)
INSERT INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, ENAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, IS_QUERY, ORDERNO, MASK, SET_VALUE_TYPE, MODIFI_ABLE, IS_ACTIVE)
VALUES
(@cloud_item_id, 'NAME', '名称', '名称', 'Name', 'text', 'varchar', 500, 'N', 'Y', 8, '1111111111', 'byPage', 'Y', 'Y'),
(@cloud_item_id, 'ITEM_TYPE', '项目类型', '项目类型', 'Item Type', 'select', 'varchar', 20, 'N', 'Y', 9, '1111111111', 'select', 'Y', 'Y'),
(@cloud_item_id, 'PARENT_ID', '父文件夹ID', '父文件夹ID', 'Parent ID', 'text', 'int', 11, 'Y', 'N', 10, '1111111111', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'PATH', '路径', '路径', 'Path', 'text', 'varchar', 1000, 'N', 'N', 11, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'OWNER_ID', '所有者', '所有者', 'Owner', 'fk', 'int', 11, 'N', 'N', 12, '1111111111', 'fk', 'N', 'Y'),
(@cloud_item_id, 'STORAGE_TYPE', '存储类型', '存储类型', 'Storage Type', 'select', 'varchar', 20, 'Y', 'N', 13, '1111111111', 'byPage', 'Y', 'Y'),
(@cloud_item_id, 'STORAGE_PATH', '存储路径', '存储路径', 'Storage Path', 'text', 'varchar', 1000, 'Y', 'N', 14, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'FILE_SIZE', '文件大小', '文件大小', 'File Size', 'text', 'bigint', 20, 'Y', 'N', 15, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'FILE_TYPE', '文件类型', '文件类型', 'File Type', 'text', 'varchar', 100, 'Y', 'N', 16, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'FILE_EXT', '文件扩展名', '文件扩展名', 'File Ext', 'text', 'varchar', 20, 'Y', 'N', 17, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'MD5', 'MD5值', 'MD5值', 'MD5', 'text', 'varchar', 64, 'Y', 'N', 18, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'ACCESS_URL', '访问URL', '访问URL', 'Access URL', 'text', 'varchar', 2000, 'Y', 'N', 19, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'THUMBNAIL', '缩略图', '缩略图', 'Thumbnail', 'text', 'varchar', 2000, 'Y', 'N', 20, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'DOWNLOAD_COUNT', '下载次数', '下载次数', 'Download Count', 'text', 'int', 11, 'Y', 'N', 21, '1000000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'TAGS', '标签', '标签', 'Tags', 'text', 'varchar', 500, 'Y', 'N', 22, '1111111111', 'byPage', 'Y', 'Y'),
(@cloud_item_id, 'FILE_COUNT', '文件数量', '文件数量', 'File Count', 'text', 'int', 11, 'Y', 'N', 23, '1000000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'TOTAL_SIZE', '总大小', '总大小', 'Total Size', 'text', 'bigint', 20, 'Y', 'N', 24, '1000000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'IS_PUBLIC', '是否公开', '是否公开', 'Is Public', 'select', 'char', 1, 'Y', 'N', 25, '1111111111', 'select', 'Y', 'Y'),
(@cloud_item_id, 'SHARE_CODE', '分享码', '分享码', 'Share Code', 'text', 'varchar', 32, 'Y', 'N', 26, '1100000000', 'byPage', 'N', 'Y'),
(@cloud_item_id, 'SHARE_EXPIRE', '分享过期时间', '分享过期时间', 'Share Expire', 'datetime', 'datetime', 0, 'Y', 'N', 27, '1100000000', 'byPage', 'Y', 'Y'),
(@cloud_item_id, 'DESCRIPTION', '描述', '描述', 'Description', 'textarea', 'varchar', 2000, 'Y', 'N', 28, '1111111111', 'byPage', 'Y', 'Y');

-- ============================================================
-- Phase 3: 全局修复 IS_BIG=NULL → 'N'
-- ============================================================
UPDATE sys_table SET IS_BIG = 'N' WHERE IS_BIG IS NULL;

-- ============================================================
-- Phase 4: 全局修复 ORDERNO 缺失
-- ============================================================
UPDATE sys_table SET ORDERNO = 1 WHERE ORDERNO IS NULL;

-- ============================================================
-- Phase 5: 修复 sys_company_conf 的 MASK
-- ============================================================
UPDATE sys_table SET MASK = 'AMDQ' WHERE NAME = 'SYS_COMPANY_CONF' AND (MASK IS NULL OR MASK = '');

-- ============================================================
-- Phase 6: 修复 live_room 所有列的 MASK=NULL
-- ============================================================
SET @live_room_tid = (SELECT ID FROM sys_table WHERE NAME='live_room');
UPDATE sys_column SET MASK = '1111111111' WHERE SYS_TABLE_ID = @live_room_tid AND MASK IS NULL AND SET_VALUE_TYPE IN ('byPage', 'select', 'createBy', 'operator');
UPDATE sys_column SET MASK = '1000000000' WHERE SYS_TABLE_ID = @live_room_tid AND MASK IS NULL AND SET_VALUE_TYPE IN ('pk', 'sysdate');

-- ============================================================
-- Phase 7: 修复 select 字段缺少 SYS_DICT_ID (YESNO字典ID=1)
-- ============================================================
UPDATE sys_column 
SET SYS_DICT_ID = '1' 
WHERE SET_VALUE_TYPE = 'select' 
  AND DB_NAME = 'IS_ACTIVE' 
  AND (SYS_DICT_ID IS NULL OR SYS_DICT_ID = '');

-- ============================================================
-- Phase 8: 修复 FK 字段缺失 REF_TABLE_ID (按实际关联表补充)
-- ============================================================

-- sys_action.SYS_TABLE_ID → sys_table
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_table')
WHERE c.DB_NAME = 'SYS_TABLE_ID' AND t.NAME = 'sys_action' AND (c.REF_TABLE_ID IS NULL);

-- sys_table_cmd.SYS_TABLE_ID → sys_table
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_table')
WHERE c.DB_NAME = 'SYS_TABLE_ID' AND t.NAME = 'sys_table_cmd' AND (c.REF_TABLE_ID IS NULL);

-- sys_table_ref.SYS_TABLE_ID → sys_table
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_table')
WHERE c.DB_NAME IN ('SYS_TABLE_ID', 'REF_TABLE_ID') AND t.NAME = 'sys_table_ref' AND (c.REF_TABLE_ID IS NULL);

-- sys_user_groups.SYS_USER_ID → sys_user
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user')
WHERE c.DB_NAME = 'SYS_USER_ID' AND t.NAME = 'sys_user_groups' AND (c.REF_TABLE_ID IS NULL);

-- sys_user_groups.SYS_DIRECTORY_ID → sys_directory
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_directory')
WHERE c.DB_NAME = 'SYS_DIRECTORY_ID' AND t.NAME = 'sys_user_groups' AND (c.REF_TABLE_ID IS NULL);

-- sys_group_prem.SYS_DIRECTORY_ID → sys_directory
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_directory')
WHERE c.DB_NAME = 'SYS_DIRECTORY_ID' AND t.NAME = 'sys_group_prem' AND (c.REF_TABLE_ID IS NULL);

-- sys_group_prem.SYS_GROUPS_ID → sys_groups
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_groups')
WHERE c.DB_NAME = 'SYS_GROUPS_ID' AND t.NAME = 'sys_group_prem' AND (c.REF_TABLE_ID IS NULL);

-- sys_user_session.COMPANY_ID → sys_company
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_company')
WHERE c.DB_NAME = 'COMPANY_ID' AND t.NAME = 'sys_user_session' AND (c.REF_TABLE_ID IS NULL);

-- sys_directory.SYS_TABLE_ID → sys_table
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_table')
WHERE c.DB_NAME = 'SYS_TABLE_ID' AND t.NAME = 'sys_directory' AND (c.REF_TABLE_ID IS NULL);

-- sys_directory.SYS_TABLE_CATEGORY_ID → sys_table_category
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_table_category')
WHERE c.DB_NAME = 'SYS_TABLE_CATEGORY_ID' AND t.NAME = 'sys_directory' AND (c.REF_TABLE_ID IS NULL);

-- sys_message.SENDER_ID → sys_user
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user')
WHERE c.DB_NAME = 'SENDER_ID' AND t.NAME = 'sys_message' AND (c.REF_TABLE_ID IS NULL);

-- sys_user_message.MESSAGE_ID → sys_message
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_message')
WHERE c.DB_NAME = 'MESSAGE_ID' AND t.NAME = 'sys_user_message' AND (c.REF_TABLE_ID IS NULL);

-- sys_notification_log.USER_ID → sys_user
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user')
WHERE c.DB_NAME = 'USER_ID' AND t.NAME = 'sys_notification_log' AND (c.REF_TABLE_ID IS NULL);

-- audit_log.USER_ID → sys_user
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user')
WHERE c.DB_NAME = 'USER_ID' AND t.NAME = 'audit_log' AND (c.REF_TABLE_ID IS NULL);

-- cloud_item.OWNER_ID → sys_user
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'sys_user')
WHERE c.DB_NAME = 'OWNER_ID' AND t.NAME = 'cloud_item' AND (c.REF_TABLE_ID IS NULL);

-- Workflow tables
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.REF_TABLE_ID = (SELECT ID FROM sys_table WHERE NAME = 'wf_definition')
WHERE c.DB_NAME = 'WF_DEFINITION_ID' AND t.NAME IN ('wf_node','wf_transition','wf_instance','wf_task') AND (c.REF_TABLE_ID IS NULL);

-- ============================================================
-- Phase 9: 补全 sys_message 缺失的列
-- ============================================================
SET @msg_id = (SELECT ID FROM sys_table WHERE NAME='sys_message');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@msg_id, 'IS_ACTIVE', '是否启用', '是否启用', 'select', 'char', 1, 'Y', 13, '1111111111', 'select', 'Y'),
(@msg_id, 'MESSAGE_TYPE', '消息类型', '消息类型', 'select', 'varchar', 50, 'N', 8, '1111111111', 'select', 'Y'),
(@msg_id, 'PRIORITY', '优先级', '优先级', 'select', 'varchar', 20, 'Y', 9, '1111111111', 'select', 'Y'),
(@msg_id, 'STATUS', '状态', '状态', 'select', 'varchar', 20, 'N', 10, '1111111111', 'select', 'Y'),
(@msg_id, 'TARGET_TYPE', '目标类型', '目标类型', 'select', 'varchar', 50, 'N', 11, '1111111111', 'select', 'Y'),
(@msg_id, 'LINK_TYPE', '链接类型', '链接类型', 'text', 'varchar', 50, 'Y', 12, '1111111111', 'byPage', 'Y');

-- ============================================================
-- Phase 10: cloud_share 列定义
-- ============================================================
SET @cs_id = (SELECT ID FROM sys_table WHERE NAME='cloud_share');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@cs_id, 'SHARE_CODE', '分享码', '分享码', 'text', 'varchar', 50, 'N', 8, '1100000000', 'byPage', 'Y'),
(@cs_id, 'RESOURCE_TYPE', '资源类型', '资源类型', 'select', 'varchar', 20, 'N', 9, '1111111111', 'select', 'Y'),
(@cs_id, 'RESOURCE_ID', '资源ID', '资源ID', 'text', 'bigint', 20, 'N', 10, '1111111111', 'byPage', 'Y'),
(@cs_id, 'SHARER_ID', '分享者', '分享者', 'fk', 'bigint', 20, 'N', 11, '1111111111', 'fk', 'Y'),
(@cs_id, 'SHARE_TYPE', '分享类型', '分享类型', 'select', 'varchar', 20, 'N', 12, '1111111111', 'select', 'Y'),
(@cs_id, 'PASSWORD', '访问密码', '访问密码', 'text', 'varchar', 50, 'Y', 13, '1111111111', 'password', 'Y'),
(@cs_id, 'EXPIRE_TIME', '过期时间', '过期时间', 'datetime', 'datetime', 0, 'Y', 14, '1111111111', 'byPage', 'Y'),
(@cs_id, 'MAX_DOWNLOADS', '最大下载次数', '最大下载次数', 'text', 'int', 11, 'Y', 15, '1111111111', 'byPage', 'Y'),
(@cs_id, 'DOWNLOAD_COUNT', '下载次数', '下载次数', 'text', 'int', 11, 'Y', 16, '1000000000', 'operator', 'Y'),
(@cs_id, 'VIEW_COUNT', '查看次数', '查看次数', 'text', 'int', 11, 'Y', 17, '1000000000', 'operator', 'Y'),
(@cs_id, 'STATUS', '状态', '状态', 'select', 'varchar', 20, 'Y', 18, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME);

-- ============================================================
-- Phase 11: cloud_quota 列定义
-- ============================================================
SET @cq_id = (SELECT ID FROM sys_table WHERE NAME='cloud_quota');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@cq_id, 'USER_ID', '用户', '用户', 'fk', 'bigint', 20, 'N', 8, '1111111111', 'fk', 'Y'),
(@cq_id, 'TOTAL_QUOTA', '总配额(字节)', '总配额', 'text', 'bigint', 20, 'N', 9, '1111111111', 'byPage', 'Y'),
(@cq_id, 'USED_SPACE', '已用空间(字节)', '已用空间', 'text', 'bigint', 20, 'Y', 10, '1000000000', 'operator', 'Y'),
(@cq_id, 'FILE_COUNT', '文件数量', '文件数量', 'text', 'int', 11, 'Y', 11, '1000000000', 'operator', 'Y'),
(@cq_id, 'FOLDER_COUNT', '文件夹数量', '文件夹数量', 'text', 'int', 11, 'Y', 12, '1000000000', 'operator', 'Y'),
(@cq_id, 'MAX_FILE_SIZE', '单文件最大(字节)', '最大文件', 'text', 'bigint', 20, 'Y', 13, '1111111111', 'byPage', 'Y'),
(@cq_id, 'QUOTA_TYPE', '配额类型', '配额类型', 'select', 'varchar', 20, 'Y', 14, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME);

-- ============================================================
-- Phase 12: pull_stream_task 列定义
-- ============================================================
SET @pt_id = (SELECT ID FROM sys_table WHERE NAME='pull_stream_task');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@pt_id, 'TASK_ID', '任务ID', '任务ID', 'text', 'varchar', 100, 'N', 8, '1000000000', 'pk', 'Y'),
(@pt_id, 'TASK_NAME', '任务名称', '任务名称', 'text', 'varchar', 255, 'N', 9, '1111111111', 'byPage', 'Y'),
(@pt_id, 'SOURCE_URL', '源流地址', '源流地址', 'text', 'varchar', 500, 'N', 10, '1111111111', 'byPage', 'Y'),
(@pt_id, 'PUSH_DOMAIN', '推流域名', '推流域名', 'text', 'varchar', 255, 'Y', 11, '1111111111', 'byPage', 'Y'),
(@pt_id, 'STREAM_KEY', '推流密钥', '推流密钥', 'text', 'varchar', 255, 'Y', 12, '1111111111', 'password', 'Y'),
(@pt_id, 'STATUS', '状态', '状态', 'select', 'varchar', 50, 'N', 13, '1111111111', 'select', 'Y'),
(@pt_id, 'START_TIME', '开始时间', '开始时间', 'datetime', 'datetime', 0, 'Y', 14, '1111111111', 'byPage', 'Y'),
(@pt_id, 'END_TIME', '结束时间', '结束时间', 'datetime', 'datetime', 0, 'Y', 15, '1111111111', 'byPage', 'Y'),
(@pt_id, 'ROOM_ID', '直播间ID', '直播间ID', 'fk', 'bigint', 20, 'Y', 16, '1111111111', 'fk', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME);

-- ============================================================
-- Phase 13: sys_language + sys_translation 列定义
-- ============================================================
SET @sl_id = (SELECT ID FROM sys_table WHERE NAME='sys_language');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@sl_id, 'CODE', '语言代码', '语言代码', 'text', 'varchar', 20, 'N', 8, '1111111111', 'byPage', 'Y'),
(@sl_id, 'NAME', '语言名称', '语言名称', 'text', 'varchar', 100, 'N', 9, '1111111111', 'byPage', 'Y'),
(@sl_id, 'IS_DEFAULT', '默认语言', '默认语言', 'select', 'char', 1, 'Y', 10, '1111111111', 'select', 'Y'),
(@sl_id, 'IS_ACTIVE', '启用', '启用', 'select', 'char', 1, 'Y', 11, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME);

SET @st_id = (SELECT ID FROM sys_table WHERE NAME='sys_translation');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@st_id, 'KEY', '键名', '键名', 'text', 'varchar', 255, 'N', 8, '1111111111', 'byPage', 'Y'),
(@st_id, 'LANGUAGE', '语言', '语言', 'select', 'varchar', 20, 'N', 9, '1111111111', 'select', 'Y'),
(@st_id, 'VALUE', '翻译值', '翻译值', 'textarea', 'text', 0, 'N', 10, '1111111111', 'byPage', 'Y'),
(@st_id, 'MODULE', '所属模块', '所属模块', 'text', 'varchar', 100, 'Y', 11, '1111111111', 'byPage', 'Y'),
(@st_id, 'IS_ACTIVE', '启用', '启用', 'select', 'char', 1, 'Y', 12, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME);

-- ============================================================
-- Phase 14: cloud_upload_session + cloud_chunk_record 列定义
-- ============================================================
SET @us_id = (SELECT ID FROM sys_table WHERE NAME='cloud_upload_session');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE)
VALUES
(@us_id, 'SESSION_ID', '会话ID', '会话ID', 'text', 'varchar', 100, 'N', 8, '1000000000', 'pk', 'Y'),
(@us_id, 'FILE_NAME', '文件名', '文件名', 'text', 'varchar', 500, 'N', 9, '1111111111', 'byPage', 'Y'),
(@us_id, 'FILE_SIZE', '文件大小', '文件大小', 'text', 'bigint', 20, 'N', 10, '1111111111', 'byPage', 'Y'),
(@us_id, 'FILE_MD5', '文件MD5', '文件MD5', 'text', 'varchar', 64, 'Y', 11, '1100000000', 'byPage', 'Y'),
(@us_id, 'CHUNK_SIZE', '分片大小', '分片大小', 'text', 'int', 11, 'N', 12, '1111111111', 'byPage', 'Y'),
(@us_id, 'TOTAL_CHUNKS', '总分片数', '总分片数', 'text', 'int', 11, 'N', 13, '1000000000', 'byPage', 'Y'),
(@us_id, 'UPLOADED_CHUNKS', '已传分片', '已传分片', 'text', 'int', 11, 'Y', 14, '1000000000', 'operator', 'Y'),
(@us_id, 'FOLDER_ID', '目标文件夹', '目标文件夹', 'fk', 'bigint', 20, 'Y', 15, '1111111111', 'fk', 'Y'),
(@us_id, 'STATUS', '状态', '状态', 'select', 'varchar', 50, 'N', 16, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME);

