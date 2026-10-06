-- 分享访问密码改为 AES-256-GCM 可逆加密（分享者需在「我的分享」中查看提取码，故不用单向哈希）
-- 依赖：云密钥改为可逆加密.sql（新增 encrypt 字典项）
-- 存量明文执行：go run ./cmd/encrypt-secrets（须在本脚本之后，密文长度超过原 50 字符）
-- 执行后需刷新缓存（元数据在 Redis 中缓存 24h）：POST /api/v1/metadata/refresh 与 POST /api/v1/dicts/refresh

-- 1. 加宽列以容纳密文（enc:v1: 前缀 + base64）
ALTER TABLE cloud_share MODIFY COLUMN PASSWORD VARCHAR(255) NULL COMMENT '访问密码（加密存储）';

-- 2. 元数据由 password（bcrypt，与云盘服务的明文比对不兼容）改为 encrypt
UPDATE sys_column c
JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
SET c.SET_VALUE_TYPE = 'encrypt'
WHERE t.NAME = 'CLOUD_SHARE' AND c.DB_NAME = 'PASSWORD';
