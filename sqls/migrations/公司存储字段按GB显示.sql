-- 公司配置「最大存储」按 GB、「最大文件」按 MB 输入/显示；库中仍按字节保存
-- 依赖前端 NumberField / DynamicTable 对 PROPS.unit、unitScale 的支持
UPDATE sys_column
SET DISPLAY_TYPE = 'number', PROPS = '{"unit":"GB","unitScale":1073741824,"min":0}', UPDATE_TIME = NOW()
WHERE FULL_NAME = 'SYS_COMPANY_CONF.MAX_STORAGE_SIZE' AND IS_ACTIVE = 'Y';

UPDATE sys_column
SET DISPLAY_TYPE = 'number', PROPS = '{"unit":"MB","unitScale":1048576,"min":0}', UPDATE_TIME = NOW()
WHERE FULL_NAME = 'SYS_COMPANY_CONF.MAX_FILE_SIZE' AND IS_ACTIVE = 'Y';
