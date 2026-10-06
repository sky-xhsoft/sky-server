-- 会话表 TOKEN / REFRESH_TOKEN 改为只保存 SHA-256 摘要（与 jwt.TokenHash 一致：小写十六进制）
-- 幂等：只处理仍是 JWT 原文（eyJ 开头）的记录

UPDATE sys_user_session SET TOKEN = SHA2(TOKEN, 256) WHERE TOKEN LIKE 'eyJ%';
UPDATE sys_user_session SET REFRESH_TOKEN = SHA2(REFRESH_TOKEN, 256) WHERE REFRESH_TOKEN LIKE 'eyJ%';
