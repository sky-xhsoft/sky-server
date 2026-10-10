-- 公司云盘文件数不做限制：后端不再校验 MAX_FILES，停用该字段的元数据（表单、列表不再显示）
-- 只停用不删除物理列，便于恢复；执行后刷新元数据缓存：POST /api/v1/metadata/refresh
UPDATE sys_column
SET IS_ACTIVE = 'N', UPDATE_TIME = NOW()
WHERE FULL_NAME = 'SYS_COMPANY_CONF.MAX_FILES';
