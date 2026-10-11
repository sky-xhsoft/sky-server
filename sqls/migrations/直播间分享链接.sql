-- 直播间分享链接：任何人打开链接即可观看，无需登录
-- SHARE_CODE：分享码（链接 /live-share/{code}），重新生成后旧链接失效；SHARE_ENABLED：是否开启分享
-- 两列只由分享接口读写，不登记到元数据（通用 CRUD 不读写）
-- 幂等：列、索引已存在则跳过

SET @col_exists := (SELECT COUNT(*) FROM information_schema.COLUMNS
                    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND COLUMN_NAME = 'SHARE_CODE');
SET @ddl := IF(@col_exists = 0,
               'ALTER TABLE live_room ADD COLUMN SHARE_CODE VARCHAR(64) NULL COMMENT ''分享码（免登录观看链接）''',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM information_schema.COLUMNS
                    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND COLUMN_NAME = 'SHARE_ENABLED');
SET @ddl := IF(@col_exists = 0,
               'ALTER TABLE live_room ADD COLUMN SHARE_ENABLED CHAR(1) NOT NULL DEFAULT ''N'' COMMENT ''是否开启分享：Y/N''',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (SELECT COUNT(*) FROM information_schema.STATISTICS
                    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_room' AND INDEX_NAME = 'UK_LIVE_ROOM_SHARE_CODE');
SET @ddl := IF(@idx_exists = 0,
               'ALTER TABLE live_room ADD UNIQUE INDEX UK_LIVE_ROOM_SHARE_CODE (SHARE_CODE)',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
