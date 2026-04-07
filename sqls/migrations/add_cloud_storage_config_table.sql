-- ================================================================
-- Cloud Storage Configuration Table Migration
-- 云盘存储配置表迁移脚本
-- ================================================================
-- Usage:
--   mysql -u root -p skyserver < sqls/migrations/add_cloud_storage_config_table.sql
-- ================================================================

USE skyserver;

-- ================= 第1步: 创建cloud_storage_config表 =================

-- 检查表是否已存在
SET @table_exists = 0;
SELECT COUNT(*) INTO @table_exists
FROM information_schema.tables
WHERE table_schema = DATABASE()
  AND table_name = 'cloud_storage_config';

-- 只有在表不存在时才创建
SET @sql = IF(@table_exists = 0,
    'CREATE TABLE `cloud_storage_config` (
        `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT ''主键ID'',
        `SYS_COMPANY_ID` INT UNSIGNED NOT NULL COMMENT ''公司ID'',
        `STORAGE_TYPE` VARCHAR(20) NOT NULL DEFAULT ''local'' COMMENT ''存储类型: local, aliyunOSS, tencentCOS'',
        `LOCAL_BASE_PATH` VARCHAR(500) DEFAULT ''uploads'' COMMENT ''本地存储基础路径'',
        `LOCAL_BASE_URL` VARCHAR(500) DEFAULT ''/files'' COMMENT ''本地存储基础URL'',
        `ALIYUN_OSS_ENDPOINT` VARCHAR(255) COMMENT ''阿里云OSS Endpoint'',
        `ALIYUN_OSS_ACCESS_KEY_ID` VARCHAR(255) COMMENT ''阿里云OSS AccessKeyID'',
        `ALIYUN_OSS_ACCESS_KEY_SECRET` VARCHAR(255) COMMENT ''阿里云OSS AccessKeySecret'',
        `ALIYUN_OSS_BUCKET_NAME` VARCHAR(255) COMMENT ''阿里云OSS Bucket名称'',
        `ALIYUN_OSS_CDN_DOMAIN` VARCHAR(500) COMMENT ''阿里云OSS CDN域名'',
        `TENCENT_COS_BUCKET_URL` VARCHAR(500) COMMENT ''腾讯云COS Bucket URL'',
        `TENCENT_COS_SECRET_ID` VARCHAR(255) COMMENT ''腾讯云COS SecretID'',
        `TENCENT_COS_SECRET_KEY` VARCHAR(255) COMMENT ''腾讯云COS SecretKey'',
        `TENCENT_COS_BUCKET_NAME` VARCHAR(255) COMMENT ''腾讯云COS Bucket名称'',
        `TENCENT_COS_REGION` VARCHAR(50) COMMENT ''腾讯云COS 区域'',
        `TENCENT_COS_CDN_DOMAIN` VARCHAR(500) COMMENT ''腾讯云COS CDN域名'',
        `IS_ACTIVE` CHAR(1) NOT NULL DEFAULT ''Y'' COMMENT ''是否有效(Y/N)'',
        `CREATE_BY` VARCHAR(80) COMMENT ''创建人'',
        `CREATE_TIME` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
        `UPDATE_BY` VARCHAR(80) COMMENT ''更新人'',
        `UPDATE_TIME` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
        UNIQUE KEY `idx_company` (`SYS_COMPANY_ID`),
        INDEX `idx_storage_type` (`STORAGE_TYPE`),
        INDEX `idx_active` (`IS_ACTIVE`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''云盘存储配置表''',
    'SELECT ''Table cloud_storage_config already exists'' AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ================= 第2步: 为默认公司插入初始配置 =================

-- 检查默认配置是否已存在
SET @default_exists = 0;
SELECT COUNT(*) INTO @default_exists
FROM cloud_storage_config
WHERE SYS_COMPANY_ID = 1;

-- 插入默认配置
SET @sql_insert = IF(@default_exists = 0,
    'INSERT INTO cloud_storage_config (
        SYS_COMPANY_ID,
        STORAGE_TYPE,
        LOCAL_BASE_PATH,
        LOCAL_BASE_URL,
        IS_ACTIVE,
        CREATE_BY,
        CREATE_TIME
    ) VALUES (
        1,
        ''local'',
        ''uploads'',
        ''/files'',
        ''Y'',
        ''system'',
        NOW()
    )',
    'SELECT ''Default config already exists'' AS message'
);

PREPARE stmt_insert FROM @sql_insert;
EXECUTE stmt_insert;
DEALLOCATE PREPARE stmt_insert;

-- ================= 第3步: 扩展sys_company_conf表（可选方案） =================

-- 检查 sys_company_conf 表是否存在
SET @sys_conf_exists = 0;
SELECT COUNT(*) INTO @sys_conf_exists
FROM information_schema.tables
WHERE table_schema = DATABASE()
  AND table_name = 'sys_company_conf';

-- 检查是否已经有存储配置列
SET @has_storage_cols = 0;
SELECT COUNT(*) INTO @has_storage_cols
FROM information_schema.columns
WHERE table_schema = DATABASE()
  AND table_name = 'sys_company_conf'
  AND column_name = 'STORAGE_TYPE';

-- 如果 sys_company_conf 表存在但没有存储配置列，添加它们
SET @sql_alter = IF(@sys_conf_exists = 1 AND @has_storage_cols = 0,
    'ALTER TABLE `sys_company_conf`
     ADD COLUMN `STORAGE_TYPE` VARCHAR(20) DEFAULT ''local'' COMMENT ''存储类型: local, aliyunOSS, tencentCOS'',
     ADD COLUMN `LOCAL_BASE_PATH` VARCHAR(500) DEFAULT ''uploads'' COMMENT ''本地存储基础路径'',
     ADD COLUMN `LOCAL_BASE_URL` VARCHAR(500) DEFAULT ''/files'' COMMENT ''本地存储基础URL'',
     ADD COLUMN `ALIYUN_OSS_ENDPOINT` VARCHAR(255) COMMENT ''阿里云OSS Endpoint'',
     ADD COLUMN `ALIYUN_OSS_ACCESS_KEY_ID` VARCHAR(255) COMMENT ''阿里云OSS AccessKeyID'',
     ADD COLUMN `ALIYUN_OSS_ACCESS_KEY_SECRET` VARCHAR(255) COMMENT ''阿里云OSS AccessKeySecret'',
     ADD COLUMN `ALIYUN_OSS_BUCKET_NAME` VARCHAR(255) COMMENT ''阿里云OSS Bucket名称'',
     ADD COLUMN `ALIYUN_OSS_CDN_DOMAIN` VARCHAR(500) COMMENT ''阿里云OSS CDN域名'',
     ADD COLUMN `TENCENT_COS_BUCKET_URL` VARCHAR(500) COMMENT ''腾讯云COS Bucket URL'',
     ADD COLUMN `TENCENT_COS_SECRET_ID` VARCHAR(255) COMMENT ''腾讯云COS SecretID'',
     ADD COLUMN `TENCENT_COS_SECRET_KEY` VARCHAR(255) COMMENT ''腾讯云COS SecretKey'',
     ADD COLUMN `TENCENT_COS_BUCKET_NAME` VARCHAR(255) COMMENT ''腾讯云COS Bucket名称'',
     ADD COLUMN `TENCENT_COS_REGION` VARCHAR(50) COMMENT ''腾讯云COS 区域'',
     ADD COLUMN `TENCENT_COS_CDN_DOMAIN` VARCHAR(500) COMMENT ''腾讯云COS CDN域名''',
    'SELECT ''No need to alter sys_company_conf table'' AS message'
);

PREPARE stmt_alter FROM @sql_alter;
EXECUTE stmt_alter;
DEALLOCATE PREPARE stmt_alter;

-- ================= 完成提示 =================

SELECT 'Cloud storage config migration completed!' AS message;
SELECT 'Use cloud_storage_config table for per-company storage configuration.' AS info;
SELECT 'Default company (ID=1) config set to local storage.' AS default_config;

