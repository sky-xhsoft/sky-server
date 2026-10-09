-- live_cut 补充 STREAM_ID 列（实体已定义，物理表缺失，写入切片记录会报 Unknown column）
-- 腾讯云高光切片回调写入切片列表时记录流 ID
-- 幂等：列已存在则跳过

SET @has_col := (SELECT COUNT(*) FROM information_schema.COLUMNS
                 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'live_cut' AND COLUMN_NAME = 'STREAM_ID');
SET @ddl := IF(@has_col = 0,
               'ALTER TABLE live_cut ADD COLUMN STREAM_ID VARCHAR(255) NULL COMMENT ''流ID'' AFTER ROOM_ID',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
