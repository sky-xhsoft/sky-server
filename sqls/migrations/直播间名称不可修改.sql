-- 直播间名称创建后不可修改（与 live_room_service.UpdateRoom 一致）：
-- 关闭 LIVE_ROOM.ROOM_NAME 的「修改可修改」（第 4 位）和「列表可修改」（第 6 位），通用数据接口 PUT 也不能改名
-- 执行后需删除 Redis 元数据缓存 metadata:columns:<LIVE_ROOM 表 ID>
UPDATE sys_column c
JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
SET c.MASK = CONCAT(SUBSTRING(c.MASK, 1, 3), '0', SUBSTRING(c.MASK, 5, 1), '0', SUBSTRING(c.MASK, 7)), c.UPDATE_TIME = NOW()
WHERE t.NAME = 'LIVE_ROOM' AND c.DB_NAME = 'ROOM_NAME' AND LENGTH(c.MASK) = 10;
