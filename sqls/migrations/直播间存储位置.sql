-- 直播间「存储位置」：开启录制/切片时，录制文件与高光切片保存到该云盘文件夹下的「直播间名称/直播录制|直播切片」
-- 替代配置文件中写死的 cloud.defaultRoomParentID（该配置项已不再使用）
-- 存储位置须为直播间创建人自己的有效文件夹，否则保存到云盘根目录
-- 幂等：列已存在、字段已登记则跳过
-- 执行后刷新元数据缓存：POST /api/v1/metadata/refresh（或删除 Redis 中 metadata:* 键）

-- 1. 物理列
SET @col_exists := (SELECT COUNT(*) FROM information_schema.COLUMNS
                    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND COLUMN_NAME = 'STORAGE_FOLDER_ID');
SET @ddl := IF(@col_exists = 0,
               'ALTER TABLE live_room ADD COLUMN STORAGE_FOLDER_ID INT UNSIGNED NULL COMMENT ''录制/切片存储位置（云盘文件夹ID，空为根目录）'' AFTER IS_CLIP',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 元数据登记（通用 CRUD 只保存已登记字段）
INSERT INTO sys_column (SYS_COMPANY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME,
                        DISPLAY_NAME, ENAME, MASK, ORDERNO, DB_NAME, COL_TYPE, COL_LENGTH, SYS_TABLE_ID,
                        NULL_ABLE, IS_QUERY, FULL_NAME, MODIFI_ABLE, SET_VALUE_TYPE,
                        DISPLAY_TYPE, DESCRIPTION)
SELECT 0, 'Y', 'system', NOW(), 'system', NOW(),
       '存储位置', 'STORAGE_FOLDER_ID', '1111111111', 97, 'STORAGE_FOLDER_ID', 'int', 10, t.ID,
       'Y', 'N', 'LIVE_ROOM.STORAGE_FOLDER_ID', 'Y', 'byPage',
       'text', '录制/切片保存的云盘文件夹ID，空为云盘根目录'
FROM sys_table t
WHERE t.NAME = 'LIVE_ROOM'
  AND NOT EXISTS (SELECT 1 FROM sys_column c WHERE c.SYS_TABLE_ID = t.ID AND c.DB_NAME = 'STORAGE_FOLDER_ID');
