-- 审计日志补全所属公司：审计中间件原先未记录公司（SYS_COMPANY_ID 全为 0），
-- 「最新动态」按公司隔离后，按操作人所在公司补全历史记录；无操作人的记录保持 0（不在最新动态展示）
UPDATE audit_log a
JOIN sys_user u ON u.ID = a.USER_ID
SET a.SYS_COMPANY_ID = u.SYS_COMPANY_ID
WHERE (a.SYS_COMPANY_ID IS NULL OR a.SYS_COMPANY_ID = 0) AND a.USER_ID > 0 AND u.SYS_COMPANY_ID > 0;
