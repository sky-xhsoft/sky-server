-- 录制列表字段为空问题 - 数据库检查脚本

-- 1. 检查录制事件总数
SELECT '录制事件总数' as 检查项, COUNT(*) as 数量
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';

-- 2. 检查字段为空的情况
SELECT
    '字段空值统计' as 检查项,
    COUNT(*) as 总数,
    SUM(CASE WHEN DOMAIN_NAME IS NULL OR DOMAIN_NAME = '' THEN 1 ELSE 0 END) as 域名为空,
    SUM(CASE WHEN APP_NAME IS NULL OR APP_NAME = '' THEN 1 ELSE 0 END) as 应用名为空,
    SUM(CASE WHEN STREAM_NAME IS NULL OR STREAM_NAME = '' THEN 1 ELSE 0 END) as 流名称为空,
    SUM(CASE WHEN STREAM_ID IS NULL OR STREAM_ID = '' THEN 1 ELSE 0 END) as 流ID为空
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';

-- 3. 查看最近5条录制事件的详细数据
SELECT
    ID,
    EVENT_TYPE as 事件类型,
    DOMAIN_NAME as 域名,
    APP_NAME as 应用名,
    STREAM_NAME as 流名称,
    STREAM_ID as 流ID,
    LEFT(EVENT_DATA, 100) as 事件数据前100字符,
    CREATE_TIME as 创建时间
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 5;

-- 4. 查看完整的 EVENT_DATA（JSON格式）
SELECT
    ID,
    EVENT_DATA as 完整事件数据
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file'
ORDER BY CREATE_TIME DESC
LIMIT 1;

-- 5. 检查是否有字段不为空的数据
SELECT
    '有效数据统计' as 检查项,
    COUNT(*) as 总数,
    SUM(CASE WHEN DOMAIN_NAME IS NOT NULL AND DOMAIN_NAME != '' THEN 1 ELSE 0 END) as 有域名,
    SUM(CASE WHEN APP_NAME IS NOT NULL AND APP_NAME != '' THEN 1 ELSE 0 END) as 有应用名,
    SUM(CASE WHEN STREAM_NAME IS NOT NULL AND STREAM_NAME != '' THEN 1 ELSE 0 END) as 有流名称
FROM live_callback_event
WHERE EVENT_TYPE = 'recording_file';

-- 6. 如果需要清空测试数据，取消下面的注释
-- DELETE FROM live_callback_event WHERE EVENT_TYPE = 'recording_file';
-- SELECT '已清空录制事件数据' as 操作结果;
