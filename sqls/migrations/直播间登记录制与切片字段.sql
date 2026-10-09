-- 直播间登记「是否录制」「是否切片」元数据字段
-- 物理表 live_room 已有 IS_RECORD / IS_CLIP 列，但未登记到 sys_column，
-- 通用 CRUD 会丢弃未登记字段，导致直播间表单中的两个开关从未保存（始终为默认值 Y）。
-- 腾讯云录制/切片回调按这两个字段决定是否处理，需先登记。
-- 幂等：已登记则跳过。按表名定位，不依赖各环境的表 ID。
-- 执行后刷新元数据缓存：POST /api/v1/metadata/refresh（或删除 Redis 中 metadata:* 键）

INSERT INTO sys_column (SYS_COMPANY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME,
                        DISPLAY_NAME, ENAME, MASK, ORDERNO, DB_NAME, COL_TYPE, COL_LENGTH, SYS_TABLE_ID,
                        NULL_ABLE, IS_QUERY, FULL_NAME, MODIFI_ABLE, SET_VALUE_TYPE, SYS_DICT_ID,
                        DEFAULT_VALUE, DISPLAY_TYPE, DESCRIPTION)
SELECT 0, 'Y', 'system', NOW(), 'system', NOW(),
       '是否录制', 'IS_RECORD', '1111111111', 95, 'IS_RECORD', 'char', 1, t.ID,
       'Y', 'N', 'LIVE_ROOM.IS_RECORD', 'Y', 'byPage', d.ID,
       'Y', 'select', '直播全程录制；关闭后忽略腾讯云录制相关回调'
FROM sys_table t
JOIN sys_dict d ON d.NAME = 'YESNO'
WHERE t.NAME = 'LIVE_ROOM'
  AND NOT EXISTS (SELECT 1 FROM sys_column c WHERE c.SYS_TABLE_ID = t.ID AND c.DB_NAME = 'IS_RECORD');

INSERT INTO sys_column (SYS_COMPANY_ID, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME,
                        DISPLAY_NAME, ENAME, MASK, ORDERNO, DB_NAME, COL_TYPE, COL_LENGTH, SYS_TABLE_ID,
                        NULL_ABLE, IS_QUERY, FULL_NAME, MODIFI_ABLE, SET_VALUE_TYPE, SYS_DICT_ID,
                        DEFAULT_VALUE, DISPLAY_TYPE, DESCRIPTION)
SELECT 0, 'Y', 'system', NOW(), 'system', NOW(),
       '是否切片', 'IS_CLIP', '1111111111', 96, 'IS_CLIP', 'char', 1, t.ID,
       'Y', 'N', 'LIVE_ROOM.IS_CLIP', 'Y', 'byPage', d.ID,
       'Y', 'select', '直播内容智能切片；关闭后忽略腾讯云高光切片回调'
FROM sys_table t
JOIN sys_dict d ON d.NAME = 'YESNO'
WHERE t.NAME = 'LIVE_ROOM'
  AND NOT EXISTS (SELECT 1 FROM sys_column c WHERE c.SYS_TABLE_ID = t.ID AND c.DB_NAME = 'IS_CLIP');
