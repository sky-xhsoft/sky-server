-- 更新 audit_log 表配置，隐藏系统字段
-- 使用场景：审计日志表单不需要显示系统字段（创建人、创建时间等），因为这些信息已经在日志内容中体现

UPDATE `sys_table`
SET `PROPS` = JSON_SET(
    COALESCE(`PROPS`, '{}'),
    '$.hideSystemFields',
    true
)
WHERE `NAME` = 'audit_log';

-- 验证更新结果
SELECT `ID`, `NAME`, `DISPLAY_NAME`, `PROPS`
FROM `sys_table`
WHERE `NAME` = 'audit_log';
