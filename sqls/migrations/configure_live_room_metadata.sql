-- 配置直播间元数据

-- 1. 在sys_table中注册live_room表（如果不存在则插入，如果存在则更新）
INSERT INTO `sys_table` (
  `NAME`, `DISPLAY_NAME`, `MASK`, `IS_MENU`, `DESCRIPTION`, `ORDERNO`,
  `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
) VALUES (
  'live_room', '直播间管理', 'AMDQSUVE', 'Y', '直播间管理模块', 50,
  1, 'system', NOW(), 'system', NOW(), 'Y'
)
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '直播间管理',
  `MASK` = 'AMDQSUVE',
  `IS_MENU` = 'Y',
  `DESCRIPTION` = '直播间管理模块',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();

-- 获取表ID（无论是新插入还是已存在）
SET @live_room_table_id = (SELECT ID FROM sys_table WHERE NAME = 'live_room');

-- 2. 删除可能存在的旧字段配置
DELETE FROM sys_column WHERE SYS_TABLE_ID = @live_room_table_id;

-- 3. 配置sys_column字段
INSERT INTO `sys_column` (
  `SYS_TABLE_ID`, `DB_NAME`, `DISPLAY_NAME`, `DISPLAY_TYPE`, `SET_VALUE_TYPE`,
  `NULL_ABLE`, `ORDERNO`, `IS_QUERY`, `COL_TYPE`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`,
  `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
) VALUES
-- 主键
(@live_room_table_id, 'ID', 'ID', 'number', 'pk', 'N', 10, 'Y', 'bigint', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 直播间名称
(@live_room_table_id, 'ROOM_NAME', '直播间名称', 'text', 'byPage', 'N', 20, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 直播间类型
(@live_room_table_id, 'ROOM_TYPE', '直播间类型', 'select', 'byPage', 'N', 30, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 播出形式
(@live_room_table_id, 'BROADCAST_FORMAT', '播出形式', 'select', 'byPage', 'N', 40, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 直播间阶段
(@live_room_table_id, 'ROOM_STAGE', '直播间阶段', 'select', 'byPage', 'N', 50, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 显示方式
(@live_room_table_id, 'DISPLAY_MODE', '显示方式', 'select', 'byPage', 'Y', 60, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 开始时间
(@live_room_table_id, 'START_TIME', '开始时间', 'datetime', 'byPage', 'Y', 70, 'Y', 'datetime', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 结束时间
(@live_room_table_id, 'END_TIME', '结束时间', 'datetime', 'byPage', 'Y', 80, 'N', 'datetime', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 直播间封面
(@live_room_table_id, 'COVER_IMAGE', '直播间封面', 'image', 'byPage', 'Y', 90, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 观看方式
(@live_room_table_id, 'VIEWING_METHOD', '观看方式', 'select', 'byPage', 'N', 100, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 观看密码
(@live_room_table_id, 'VIEWING_PASSWORD', '观看密码', 'password', 'byPage', 'Y', 110, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 观看价格
(@live_room_table_id, 'VIEWING_PRICE', '观看价格', 'number', 'byPage', 'Y', 120, 'N', 'decimal', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 回放方式
(@live_room_table_id, 'PLAYBACK_METHOD', '回放方式', 'select', 'byPage', 'N', 130, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 回放有效期
(@live_room_table_id, 'PLAYBACK_VALIDITY', '回放有效期', 'select', 'byPage', 'Y', 140, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 回放开始时间
(@live_room_table_id, 'PLAYBACK_START_TIME', '回放开始时间', 'time', 'byPage', 'Y', 150, 'N', 'time', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 回放结束时间
(@live_room_table_id, 'PLAYBACK_END_TIME', '回放结束时间', 'time', 'byPage', 'Y', 160, 'N', 'time', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 流名称
(@live_room_table_id, 'STREAM_NAME', '流名称', 'text', 'byPage', 'Y', 170, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 推流地址
(@live_room_table_id, 'PUSH_URL', '推流地址', 'text', 'byPage', 'Y', 180, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 播放地址
(@live_room_table_id, 'PLAY_URL', '播放地址', 'text', 'byPage', 'Y', 190, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 状态
(@live_room_table_id, 'STATUS', '状态', 'select', 'byPage', 'N', 200, 'Y', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 观看人数
(@live_room_table_id, 'VIEWER_COUNT', '观看人数', 'number', 'ignore', 'N', 210, 'N', 'int', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 峰值观看人数
(@live_room_table_id, 'PEAK_VIEWER_COUNT', '峰值观看人数', 'number', 'ignore', 'N', 220, 'N', 'int', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 直播时长
(@live_room_table_id, 'DURATION', '直播时长(秒)', 'number', 'ignore', 'N', 230, 'N', 'int', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 描述
(@live_room_table_id, 'DESCRIPTION', '直播间描述', 'textarea', 'byPage', 'Y', 240, 'N', 'text', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 扩展属性
(@live_room_table_id, 'PROPS', '扩展属性', 'json', 'byPage', 'Y', 250, 'N', 'text', 1, 'system', NOW(), 'system', NOW(), 'Y'),
-- 标准字段
(@live_room_table_id, 'SYS_COMPANY_ID', '公司ID', 'number', 'ignore', 'N', 255, 'N', 'bigint', 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@live_room_table_id, 'CREATE_BY', '创建人', 'text', 'createBy', 'Y', 260, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@live_room_table_id, 'CREATE_TIME', '创建时间', 'datetime', 'sysdate', 'N', 270, 'Y', 'datetime', 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@live_room_table_id, 'UPDATE_BY', '更新人', 'text', 'operator', 'Y', 280, 'N', 'varchar', 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@live_room_table_id, 'UPDATE_TIME', '更新时间', 'datetime', 'sysdate', 'N', 290, 'N', 'datetime', 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@live_room_table_id, 'IS_ACTIVE', '是否有效', 'select', 'byPage', 'N', 300, 'Y', 'char', 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 3. 创建数据字典（使用 INSERT ... ON DUPLICATE KEY UPDATE 避免重复）
-- 直播间类型字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('live_room_type', '直播间类型', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '直播间类型',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_room_type_id = (SELECT ID FROM sys_dict WHERE NAME = 'live_room_type');

-- 删除旧的字典项
DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_room_type_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_room_type_id, 'video', '视频直播', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_room_type_id, 'image', '图片直播', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_room_type_id, 'vr', 'VR直播', 30, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_room_type_id, 'audio', '语音直播', 40, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_room_type_id, 'graphic', '图文直播', 50, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 播出形式字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('broadcast_format', '播出形式', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '播出形式',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_broadcast_format_id = (SELECT ID FROM sys_dict WHERE NAME = 'broadcast_format');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_broadcast_format_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_broadcast_format_id, 'live', '直播', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_broadcast_format_id, 'vod', '点播/录播', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_broadcast_format_id, 'pseudo', '伪直播', 30, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 直播间阶段字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('room_stage', '直播间阶段', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '直播间阶段',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_room_stage_id = (SELECT ID FROM sys_dict WHERE NAME = 'room_stage');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_room_stage_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_room_stage_id, 'formal', '正式直播', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_room_stage_id, 'test', '测试直播', 20, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 显示方式字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('display_mode', '显示方式', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '显示方式',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_display_mode_id = (SELECT ID FROM sys_dict WHERE NAME = 'display_mode');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_display_mode_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_display_mode_id, 'landscape', '横屏', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_display_mode_id, 'portrait', '竖屏', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_display_mode_id, 'three_screen', '三分屏', 30, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 观看方式字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('viewing_method', '观看方式', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '观看方式',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_viewing_method_id = (SELECT ID FROM sys_dict WHERE NAME = 'viewing_method');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_viewing_method_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_viewing_method_id, 'public', '公开', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_viewing_method_id, 'encrypted', '加密', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_viewing_method_id, 'paid', '付费', 30, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_viewing_method_id, 'ticket', '购票进入', 40, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_viewing_method_id, 'enterprise', '企业成员观看', 50, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_viewing_method_id, 'custom', '自建成员观看', 60, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 回放方式字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('playback_method', '回放方式', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '回放方式',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_playback_method_id = (SELECT ID FROM sys_dict WHERE NAME = 'playback_method');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_playback_method_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_playback_method_id, 'post_end', '结束后回放', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_playback_method_id, 'real_time', '实时回放', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_playback_method_id, 'no_playback', '结束后不回放', 30, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 回放有效期字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('playback_validity', '回放有效期', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '回放有效期',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_playback_validity_id = (SELECT ID FROM sys_dict WHERE NAME = 'playback_validity');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_playback_validity_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_playback_validity_id, 'unlimited', '无限制', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_playback_validity_id, 'all_day', '全天', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_playback_validity_id, 'partial', '部分时段', 30, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 状态字典
INSERT INTO `sys_dict` (`NAME`, `DISPLAY_NAME`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES ('live_room_status', '直播间状态', 1, 'system', NOW(), 'system', NOW(), 'Y')
ON DUPLICATE KEY UPDATE
  `DISPLAY_NAME` = '直播间状态',
  `UPDATE_BY` = 'system',
  `UPDATE_TIME` = NOW();
SET @dict_status_id = (SELECT ID FROM sys_dict WHERE NAME = 'live_room_status');

DELETE FROM sys_dict_item WHERE SYS_DICT_ID = @dict_status_id;

INSERT INTO `sys_dict_item` (`SYS_DICT_ID`, `VALUE`, `DISPLAY_NAME`, `ORDERNO`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`)
VALUES
(@dict_status_id, 'draft', '草稿', 10, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_status_id, 'scheduled', '已排期', 20, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_status_id, 'live', '直播中', 30, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_status_id, 'ended', '已结束', 40, 1, 'system', NOW(), 'system', NOW(), 'Y'),
(@dict_status_id, 'archived', '已归档', 50, 1, 'system', NOW(), 'system', NOW(), 'Y');

-- 4. 关联字段与数据字典
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_room_type_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'ROOM_TYPE';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_broadcast_format_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'BROADCAST_FORMAT';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_room_stage_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'ROOM_STAGE';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_display_mode_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'DISPLAY_MODE';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_viewing_method_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'VIEWING_METHOD';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_playback_method_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'PLAYBACK_METHOD';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_playback_validity_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'PLAYBACK_VALIDITY';
UPDATE `sys_column` SET `SYS_DICT_ID` = @dict_status_id WHERE `SYS_TABLE_ID` = @live_room_table_id AND `DB_NAME` = 'STATUS';

-- 5. 添加菜单项到视频直播模块
-- 首先查找视频直播的目录ID
SET @live_directory_id = (SELECT ID FROM sys_directory WHERE NAME = 'live_stream' LIMIT 1);

-- 删除可能存在的旧菜单项
DELETE FROM sys_directory WHERE NAME = 'live_room_management';

-- 如果找到了视频直播目录，则添加直播间管理菜单
INSERT INTO `sys_directory` (
  `NAME`, `DISPLAY_NAME`, `PARENT_ID`, `ORDERNO`, `URL`, `SYS_TABLE_ID`,
  `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
)
SELECT
  'live_room_management', '直播间管理', @live_directory_id, 60,
  CONCAT('/metadata/list?tableId=', @live_room_table_id), @live_room_table_id,
  1, 'system', NOW(), 'system', NOW(), 'Y'
WHERE @live_directory_id IS NOT NULL;
