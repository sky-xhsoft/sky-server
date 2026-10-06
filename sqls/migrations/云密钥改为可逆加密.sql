-- 云密钥字段由 password（bcrypt 单向哈希，无法还原）改为 encrypt（AES-256-GCM 可逆加密）
-- 配合代码：SET_VALUE_TYPE=encrypt 写入时加密、查询时不返回；实体字段使用 serializer:encrypted 透明解密
-- 存量明文执行：go run ./cmd/encrypt-secrets
-- 执行后需刷新缓存（元数据在 Redis 中缓存 24h）：POST /api/v1/metadata/refresh 与 POST /api/v1/dicts/refresh

-- 1. 赋值策略字典新增「加密」
INSERT INTO sys_dict_item (SYS_COMPANY_ID, IS_ACTIVE, SYS_DICT_ID, DISPLAY_NAME, VALUE, ORDERNO, IS_DEFAULT_VALUE)
SELECT 0, 'Y', d.ID, '加密(可还原)', 'encrypt', 10, 'N'
FROM sys_dict d
WHERE d.NAME = 'set_value_type'
  AND NOT EXISTS (SELECT 1 FROM sys_dict_item i WHERE i.SYS_DICT_ID = d.ID AND i.VALUE = 'encrypt');

-- 2. 云密钥字段改为 encrypt
UPDATE sys_column c
JOIN sys_table t ON t.ID = c.SYS_TABLE_ID
SET c.SET_VALUE_TYPE = 'encrypt'
WHERE (t.NAME = 'CLOUD_STORAGE_CONFIG' AND c.DB_NAME IN ('ALIYUN_OSS_ACCESS_KEY_SECRET', 'TENCENT_COS_SECRET_KEY'))
   OR (t.NAME = 'SYS_COMPANY_CONF' AND c.DB_NAME IN ('TENCENT_CLOUD_SECRET_KEY', 'TENCENT_CLOUD_CALLBACK_KEY'));
