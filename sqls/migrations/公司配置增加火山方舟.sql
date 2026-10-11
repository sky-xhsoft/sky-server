-- 公司配置增加火山方舟大模型（字幕同传词库智能提取）：API Key（加密存储）与推理接入点
-- 须在部署新版本前执行，否则读取公司配置会因缺少列而失败
-- 幂等：可重复执行
SET @has_key := (SELECT COUNT(*) FROM information_schema.COLUMNS
                 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_company_conf' AND COLUMN_NAME = 'ARK_API_KEY');
SET @ddl := IF(@has_key = 0,
               'ALTER TABLE sys_company_conf ADD COLUMN ARK_API_KEY VARCHAR(500) NULL COMMENT ''火山方舟 API Key（加密存储）''',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_model := (SELECT COUNT(*) FROM information_schema.COLUMNS
                   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_company_conf' AND COLUMN_NAME = 'ARK_MODEL');
SET @ddl := IF(@has_model = 0,
               'ALTER TABLE sys_company_conf ADD COLUMN ARK_MODEL VARCHAR(255) NULL COMMENT ''火山方舟推理接入点 ID（ep-…）或模型名''',
               'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
