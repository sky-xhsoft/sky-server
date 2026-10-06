-- 停用与物理表不一致的元数据字段（物理表无此列，通用查询报 Unknown column 导致 500）
-- 只停用不删除，便于追溯；如需恢复把 IS_ACTIVE 改回 'Y' 并补建物理列
-- 执行后刷新元数据缓存：POST /api/v1/metadata/refresh

UPDATE sys_column c
JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
SET c.IS_ACTIVE = 'N'
WHERE (t.NAME = 'SYS_TRANSLATION' AND c.DB_NAME = 'DISPLAY_NAME')
   OR (t.NAME = 'SYS_LANGUAGE' AND c.DB_NAME IN ('CODE_PREFIX', 'DISPLAY_NAME'))
   OR (t.NAME = 'AUDIT_LOG' AND (c.DB_NAME LIKE 'test\_col\_%' OR c.DB_NAME LIKE 'del\_%'));
