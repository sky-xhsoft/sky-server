-- 云盘存储空间改为按公司计算：未配置上限的公司默认 100GB。
-- 公司 1 已用约 128GB，单独配置为 1000GB（与原 admin 个人配额一致）；已配置过上限的不覆盖。
UPDATE sys_company_conf
SET MAX_STORAGE_SIZE = 1073741824000, UPDATE_TIME = NOW()
WHERE SYS_COMPANY_ID = 1 AND IS_ACTIVE = 'Y' AND (MAX_STORAGE_SIZE IS NULL OR MAX_STORAGE_SIZE = 0);
