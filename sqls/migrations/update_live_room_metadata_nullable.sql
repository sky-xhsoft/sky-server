-- 更新直播间表字段的元数据配置
-- 将前端删除的字段标记为可空

USE `skyserver`;

SET NAMES utf8mb4;

-- 获取表ID
SET @live_room_table_id = (SELECT ID FROM sys_table WHERE NAME = 'live_room');

-- 将以下字段标记为可空（NULL_ABLE = 'Y'）
UPDATE `sys_column`
SET `NULL_ABLE` = 'Y'
WHERE `SYS_TABLE_ID` = @live_room_table_id
  AND `DB_NAME` IN (
    'ROOM_TYPE',
    'ROOM_STAGE',
    'DISPLAY_MODE',
    'END_TIME',
    'VIEWING_METHOD',
    'VIEWING_PASSWORD',
    'VIEWING_PRICE',
    'PLAYBACK_METHOD',
    'PLAYBACK_VALIDITY',
    'PLAYBACK_START_TIME',
    'PLAYBACK_END_TIME',
    'STREAM_NAME',
    'PUSH_URL',
    'PLAY_URL',
    'DESCRIPTION',
    'PROPS'
  );

-- 确保保留的必填字段仍然为 NOT NULL（NULL_ABLE = 'N'）
UPDATE `sys_column`
SET `NULL_ABLE` = 'N'
WHERE `SYS_TABLE_ID` = @live_room_table_id
  AND `DB_NAME` IN (
    'ID',
    'ROOM_NAME',
    'BROADCAST_FORMAT',
    'START_TIME'
  );
