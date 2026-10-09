-- 直播间录制、高光切片分别设置存储位置（云盘文件夹）
-- 录制保存到「录制存储位置/直播间名称/直播录制」，切片保存到「切片存储位置/直播间名称/直播切片」
-- 替代「直播间存储位置.sql」中的单一 STORAGE_FOLDER_ID：原值复制到两个新列后删除旧列与其元数据
-- 存储位置须为直播间创建人自己的有效文件夹，否则保存到云盘根目录
-- 幂等：可重复执行
-- 执行后刷新元数据缓存：POST /api/v1/metadata/refresh（或删除 Redis 中 metadata:* 键）

-- 1. 新增物理列
SET @has_record := (SELECT COUNT(*) FROM information_schema.COLUMNS
                    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND COLUMN_NAME = 'RECORD_FOLDER_ID');
SET @ddl := IF(@has_record = 0,
               'ALTER TABLE live_room ADD COLUMN RECORD_FOLDER_ID INT UNSIGNED NULL COMMENT ''录制存储位置（云盘文件夹ID，空为根目录）'' AFTER IS_CLIP',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_clip := (SELECT COUNT(*) FROM information_schema.COLUMNS
                  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND COLUMN_NAME = 'CLIP_FOLDER_ID');
SET @ddl := IF(@has_clip = 0,
               'ALTER TABLE live_room ADD COLUMN CLIP_FOLDER_ID INT UNSIGNED NULL COMMENT ''切片存储位置（云盘文件夹ID，空为根目录）'' AFTER RECORD_FOLDER_ID',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 原单一存储位置复制到两个新列，然后删除旧列
SET @has_old := (SELECT COUNT(*) FROM information_schema.COLUMNS
                 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND COLUMN_NAME = 'STORAGE_FOLDER_ID');
SET @dml := IF(@has_old = 1,
               'UPDATE live_room SET RECORD_FOLDER_ID = IFNULL(RECORD_FOLDER_ID, STORAGE_FOLDER_ID), CLIP_FOLDER_ID = IFNULL(CLIP_FOLDER_ID, STORAGE_FOLDER_ID) WHERE STORAGE_FOLDER_ID IS NOT NULL',
               'SELECT 1');
PREPARE stmt FROM @dml;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DELETE c FROM sys_column c
JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
WHERE t.NAME = 'LIVE_ROOM' AND c.DB_NAME = 'STORAGE_FOLDER_ID';

SET @ddl := IF(@has_old = 1, 'ALTER TABLE live_room DROP COLUMN STORAGE_FOLDER_ID', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. 元数据登记（通用 CRUD 只保存已登记字段）
INSERT INTO sys_column (SYS_COMPANY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME,
                        DISPLAY_NAME, ENAME, MASK, ORDERNO, DB_NAME, COL_TYPE, COL_LENGTH, SYS_TABLE_ID,
                        NULL_ABLE, IS_QUERY, FULL_NAME, MODIFI_ABLE, SET_VALUE_TYPE, DISPLAY_TYPE, DESCRIPTION)
SELECT 0, 'Y', 'system', NOW(), 'system', NOW(),
       '录制存储位置', 'RECORD_FOLDER_ID', '1111111111', 97, 'RECORD_FOLDER_ID', 'int', 10, t.ID,
       'Y', 'N', 'LIVE_ROOM.RECORD_FOLDER_ID', 'Y', 'byPage', 'text', '录制文件保存的云盘文件夹ID，空为云盘根目录'
FROM sys_table t
WHERE t.NAME = 'LIVE_ROOM'
  AND NOT EXISTS (SELECT 1 FROM sys_column c WHERE c.SYS_TABLE_ID = t.ID AND c.DB_NAME = 'RECORD_FOLDER_ID');

INSERT INTO sys_column (SYS_COMPANY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME,
                        DISPLAY_NAME, ENAME, MASK, ORDERNO, DB_NAME, COL_TYPE, COL_LENGTH, SYS_TABLE_ID,
                        NULL_ABLE, IS_QUERY, FULL_NAME, MODIFI_ABLE, SET_VALUE_TYPE, DISPLAY_TYPE, DESCRIPTION)
SELECT 0, 'Y', 'system', NOW(), 'system', NOW(),
       '切片存储位置', 'CLIP_FOLDER_ID', '1111111111', 98, 'CLIP_FOLDER_ID', 'int', 10, t.ID,
       'Y', 'N', 'LIVE_ROOM.CLIP_FOLDER_ID', 'Y', 'byPage', 'text', '高光切片保存的云盘文件夹ID，空为云盘根目录'
FROM sys_table t
WHERE t.NAME = 'LIVE_ROOM'
  AND NOT EXISTS (SELECT 1 FROM sys_column c WHERE c.SYS_TABLE_ID = t.ID AND c.DB_NAME = 'CLIP_FOLDER_ID');
