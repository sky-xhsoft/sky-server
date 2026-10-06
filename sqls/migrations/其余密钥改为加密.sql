-- 其余服务密钥改为 encrypt（AES-256-GCM 可逆加密）：通用 SecretKey、SMTP 密码
-- 依赖：云密钥改为可逆加密.sql（新增 encrypt 字典项）
-- 火山引擎 VOLCENGINE_ACCESS_KEY_SECRET 未登记元数据，仅通过实体 serializer:encrypted 加密
-- 存量明文执行：go run ./cmd/encrypt-secrets
-- 执行后需刷新缓存（元数据在 Redis 中缓存 24h）：POST /api/v1/metadata/refresh 与 POST /api/v1/dicts/refresh

UPDATE sys_column c
JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
SET c.SET_VALUE_TYPE = 'encrypt'
WHERE (t.NAME = 'SYS_COMPANY_CONF' AND c.DB_NAME = 'SECRET_KEY')
   OR (t.NAME = 'SYS_EMAIL_CONFIG' AND c.DB_NAME = 'SMTP_PASSWORD');
