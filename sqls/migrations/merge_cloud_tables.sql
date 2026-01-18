-- ==========================================
-- 云盘表合并迁移脚本
-- 将 cloud_file 和 cloud_folder 合并为 cloud_item
-- ==========================================

-- 1. 创建新的统一表 cloud_item
DROP TABLE IF EXISTS `cloud_item`;
CREATE TABLE `cloud_item` (
  `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `ITEM_TYPE` VARCHAR(20) NOT NULL COMMENT '项目类型: file, folder',
  `NAME` VARCHAR(255) NOT NULL COMMENT '名称（文件名或文件夹名）',
  `PARENT_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '父文件夹ID',
  `PATH` VARCHAR(1000) NOT NULL COMMENT '完整路径',
  `OWNER_ID` BIGINT UNSIGNED NOT NULL COMMENT '所有者ID',

  -- 文件专用字段（文件夹时为NULL）
  `STORAGE_TYPE` VARCHAR(20) DEFAULT NULL COMMENT '存储类型: local, oss（仅文件）',
  `STORAGE_PATH` VARCHAR(500) DEFAULT NULL COMMENT '存储路径（仅文件）',
  `FILE_SIZE` BIGINT DEFAULT NULL COMMENT '文件大小（字节，仅文件）',
  `FILE_TYPE` VARCHAR(100) DEFAULT NULL COMMENT '文件MIME类型（仅文件）',
  `FILE_EXT` VARCHAR(20) DEFAULT NULL COMMENT '文件扩展名（仅文件）',
  `MD5` VARCHAR(32) DEFAULT NULL COMMENT 'MD5值（仅文件）',
  `ACCESS_URL` VARCHAR(500) DEFAULT NULL COMMENT '访问URL（仅文件）',
  `THUMBNAIL` VARCHAR(500) DEFAULT NULL COMMENT '缩略图URL（仅文件）',
  `DOWNLOAD_COUNT` INT DEFAULT 0 COMMENT '下载次数（仅文件）',
  `TAGS` VARCHAR(500) DEFAULT NULL COMMENT '标签（逗号分隔，仅文件）',

  -- 文件夹专用字段（文件时为NULL）
  `FILE_COUNT` INT DEFAULT 0 COMMENT '文件数量（仅文件夹）',
  `TOTAL_SIZE` BIGINT DEFAULT 0 COMMENT '总大小（字节，仅文件夹）',

  -- 共用字段
  `IS_PUBLIC` CHAR(1) DEFAULT 'N' COMMENT '是否公开 Y/N',
  `SHARE_CODE` VARCHAR(50) DEFAULT NULL COMMENT '分享码',
  `SHARE_EXPIRE` DATETIME DEFAULT NULL COMMENT '分享过期时间',
  `DESCRIPTION` VARCHAR(500) DEFAULT NULL COMMENT '描述',

  -- 系统字段
  `SYS_COMPANY_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` CHAR(1) DEFAULT 'Y' COMMENT '是否有效 Y/N',

  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_share_code` (`SHARE_CODE`),
  KEY `idx_parent_type` (`PARENT_ID`, `ITEM_TYPE`),
  KEY `idx_owner_type` (`OWNER_ID`, `ITEM_TYPE`),
  KEY `idx_type` (`ITEM_TYPE`),
  KEY `idx_md5` (`MD5`),
  KEY `idx_path` (`PATH`(255)),
  KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘项目表（文件+文件夹统一）';

-- 2. 数据迁移：从 cloud_folder 迁移数据
INSERT INTO `cloud_item` (
  `ID`,
  `ITEM_TYPE`,
  `NAME`,
  `PARENT_ID`,
  `PATH`,
  `OWNER_ID`,
  `FILE_COUNT`,
  `TOTAL_SIZE`,
  `IS_PUBLIC`,
  `SHARE_CODE`,
  `SHARE_EXPIRE`,
  `DESCRIPTION`,
  `SYS_COMPANY_ID`,
  `CREATE_BY`,
  `CREATE_TIME`,
  `UPDATE_BY`,
  `UPDATE_TIME`,
  `IS_ACTIVE`
)
SELECT
  `ID`,
  'folder' AS `ITEM_TYPE`,
  `NAME`,
  `PARENT_ID`,
  `PATH`,
  `OWNER_ID`,
  `FILE_COUNT`,
  `TOTAL_SIZE`,
  `IS_PUBLIC`,
  `SHARE_CODE`,
  `SHARE_EXPIRE`,
  `DESCRIPTION`,
  `SYS_COMPANY_ID`,
  `CREATE_BY`,
  `CREATE_TIME`,
  `UPDATE_BY`,
  `UPDATE_TIME`,
  `IS_ACTIVE`
FROM `cloud_folder`
WHERE `IS_ACTIVE` = 'Y';

-- 3. 创建临时映射表，记录文件旧ID到新ID的映射
CREATE TEMPORARY TABLE `file_id_mapping` (
  `old_id` BIGINT UNSIGNED NOT NULL,
  `new_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`old_id`)
);

-- 4. 数据迁移：从 cloud_file 迁移数据（使用 AUTO_INCREMENT 自动分配新ID）
INSERT INTO `cloud_item` (
  `ITEM_TYPE`,
  `NAME`,
  `PARENT_ID`,
  `PATH`,
  `OWNER_ID`,
  `STORAGE_TYPE`,
  `STORAGE_PATH`,
  `FILE_SIZE`,
  `FILE_TYPE`,
  `FILE_EXT`,
  `MD5`,
  `ACCESS_URL`,
  `THUMBNAIL`,
  `DOWNLOAD_COUNT`,
  `TAGS`,
  `IS_PUBLIC`,
  `SHARE_CODE`,
  `SHARE_EXPIRE`,
  `DESCRIPTION`,
  `SYS_COMPANY_ID`,
  `CREATE_BY`,
  `CREATE_TIME`,
  `UPDATE_BY`,
  `UPDATE_TIME`,
  `IS_ACTIVE`
)
SELECT
  'file' AS `ITEM_TYPE`,
  `FILE_NAME` AS `NAME`,
  `FOLDER_ID` AS `PARENT_ID`,
  `PATH`,
  `OWNER_ID`,
  `STORAGE_TYPE`,
  `STORAGE_PATH`,
  `FILE_SIZE`,
  `FILE_TYPE`,
  `FILE_EXT`,
  `MD5`,
  `ACCESS_URL`,
  `THUMBNAIL`,
  `DOWNLOAD_COUNT`,
  `TAGS`,
  `IS_PUBLIC`,
  `SHARE_CODE`,
  `SHARE_EXPIRE`,
  `DESCRIPTION`,
  `SYS_COMPANY_ID`,
  `CREATE_BY`,
  `CREATE_TIME`,
  `UPDATE_BY`,
  `UPDATE_TIME`,
  `IS_ACTIVE`
FROM `cloud_file`
WHERE `IS_ACTIVE` = 'Y'
ORDER BY `ID`;

-- 5. 记录文件ID映射（通过CREATE_TIME和OWNER_ID关联，确保唯一性）
INSERT INTO `file_id_mapping` (`old_id`, `new_id`)
SELECT cf.ID, ci.ID
FROM `cloud_file` cf
JOIN `cloud_item` ci ON 
  ci.NAME = cf.FILE_NAME AND 
  ci.OWNER_ID = cf.OWNER_ID AND
  ci.CREATE_TIME = cf.CREATE_TIME AND
  ci.ITEM_TYPE = 'file'
WHERE cf.IS_ACTIVE = 'Y';

-- 6. 更新 cloud_share 表的引用（使用映射表）
UPDATE `cloud_share` cs
JOIN `file_id_mapping` fim ON cs.RESOURCE_ID = fim.old_id
SET cs.RESOURCE_ID = fim.new_id
WHERE cs.RESOURCE_TYPE = 'file';

-- -- 7. 更新 cloud_upload_session 表的文件夹引用
-- UPDATE `cloud_upload_session` 
-- SET `FOLDER_ID` = NULL 
-- WHERE `FOLDER_ID` IS NOT NULL AND `FOLDER_ID` NOT IN (
--   SELECT ID FROM `cloud_item` WHERE ITEM_TYPE = 'folder'
-- );

-- 8. 备份旧表（不删除，以防需要回滚）
RENAME TABLE `cloud_file` TO `cloud_file_backup`;
RENAME TABLE `cloud_folder` TO `cloud_folder_backup`;

-- ==========================================
-- 验证迁移结果
-- ==========================================

-- 检查数据量
SELECT 'cloud_item 总数' AS 说明, COUNT(*) AS 数量 FROM `cloud_item`
UNION ALL
SELECT '文件夹数量', COUNT(*) FROM `cloud_item` WHERE `ITEM_TYPE` = 'folder'
UNION ALL
SELECT '文件数量', COUNT(*) FROM `cloud_item` WHERE `ITEM_TYPE` = 'file'
UNION ALL
SELECT '原 cloud_folder 数量', COUNT(*) FROM `cloud_folder_backup`
UNION ALL
SELECT '原 cloud_file 数量', COUNT(*) FROM `cloud_file_backup`;

-- ==========================================
-- 回滚脚本（如果需要）
-- ==========================================

/*
-- 恢复旧表
RENAME TABLE `cloud_file_backup` TO `cloud_file`;
RENAME TABLE `cloud_folder_backup` TO `cloud_folder`;

-- 删除新表
DROP TABLE IF EXISTS `cloud_item`;
*/
