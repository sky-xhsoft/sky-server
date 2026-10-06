-- 清理残留表（不可恢复，请确认后手动执行；执行前建议先备份）
-- 1) cloud_file / cloud_folder：旧版云盘表，已被 cloud_item 取代，均为 0 行；代码中 CloudFile/CloudFolder 仅作数据传输结构使用
-- 2) _bak_lpdc_sys_column / _bak_lpdc_sys_table：元数据调整时留下的备份表（15 行 / 1 行），无主键、无代码引用
-- 执行后刷新元数据缓存：POST /api/v1/metadata/refresh

-- 先停用旧表元数据（同时停用其字段），避免启动自检与通用查询引用已删除的表
UPDATE sys_column SET IS_ACTIVE = 'N' WHERE SYS_TABLE_ID IN (SELECT ID FROM (SELECT ID FROM sys_table WHERE NAME IN ('CLOUD_FILE', 'CLOUD_FOLDER')) t);
UPDATE sys_table SET IS_ACTIVE = 'N' WHERE NAME IN ('CLOUD_FILE', 'CLOUD_FOLDER');

DROP TABLE IF EXISTS cloud_file;
DROP TABLE IF EXISTS cloud_folder;
DROP TABLE IF EXISTS _bak_lpdc_sys_column;
DROP TABLE IF EXISTS _bak_lpdc_sys_table;
