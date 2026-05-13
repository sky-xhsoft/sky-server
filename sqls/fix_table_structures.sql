-- ==========================================
-- 修复 sys_table 表单查询问题
-- 问题：CRUD引擎对所有表自动注入标准字段，部分表缺失这些列
-- ==========================================

-- ==========================================
-- 1. AUDIT_LOG - 添加缺失的标准字段
-- ==========================================
ALTER TABLE `audit_log` 
  ADD COLUMN IF NOT EXISTS `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人' AFTER `SYS_COMPANY_ID`,
  ADD COLUMN IF NOT EXISTS `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人' AFTER `CREATE_BY`,
  ADD COLUMN IF NOT EXISTS `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' AFTER `UPDATE_BY`;

-- ==========================================
-- 2. cloud_chunk_record - 添加缺失的标准字段
-- ==========================================
ALTER TABLE `cloud_chunk_record` 
  ADD COLUMN IF NOT EXISTS `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID' AFTER `ID`,
  ADD COLUMN IF NOT EXISTS `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人' AFTER `UPDATE_TIME`,
  ADD COLUMN IF NOT EXISTS `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人' AFTER `CREATE_BY`,
  ADD COLUMN IF NOT EXISTS `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效' AFTER `UPDATE_BY`;

-- ==========================================
-- 3. cloud_file - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `cloud_file` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `FILE_NAME` varchar(255) NOT NULL COMMENT '文件名',
  `FOLDER_ID` bigint DEFAULT NULL COMMENT '文件夹ID',
  `PATH` varchar(1000) NOT NULL COMMENT '完整路径',
  `STORAGE_TYPE` varchar(20) NOT NULL DEFAULT 'local' COMMENT '存储类型',
  `STORAGE_PATH` varchar(500) NOT NULL COMMENT '存储路径',
  `FILE_SIZE` bigint NOT NULL DEFAULT 0 COMMENT '文件大小',
  `FILE_TYPE` varchar(100) DEFAULT NULL COMMENT '文件MIME类型',
  `FILE_EXT` varchar(20) DEFAULT NULL COMMENT '文件扩展名',
  `MD5` varchar(32) DEFAULT NULL COMMENT 'MD5值',
  `OWNER_ID` bigint NOT NULL COMMENT '所有者ID',
  `IS_PUBLIC` char(1) DEFAULT 'N' COMMENT '是否公开',
  `SHARE_CODE` varchar(50) DEFAULT NULL COMMENT '分享码',
  `SHARE_EXPIRE` datetime DEFAULT NULL COMMENT '分享过期时间',
  `ACCESS_URL` varchar(500) DEFAULT NULL COMMENT '访问URL',
  `THUMBNAIL` varchar(500) DEFAULT NULL COMMENT '缩略图',
  `DOWNLOAD_COUNT` int DEFAULT 0 COMMENT '下载次数',
  `TAGS` varchar(500) DEFAULT NULL COMMENT '标签',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_folder_id` (`FOLDER_ID`),
  KEY `idx_owner_id` (`OWNER_ID`),
  KEY `idx_md5` (`MD5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云盘文件表';

-- ==========================================
-- 4. cloud_folder - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `cloud_folder` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `NAME` varchar(255) NOT NULL COMMENT '文件夹名称',
  `PARENT_ID` bigint DEFAULT NULL COMMENT '父文件夹ID',
  `PATH` varchar(1000) NOT NULL COMMENT '完整路径',
  `OWNER_ID` bigint NOT NULL COMMENT '所有者ID',
  `IS_PUBLIC` char(1) DEFAULT 'N' COMMENT '是否公开',
  `SHARE_CODE` varchar(50) DEFAULT NULL COMMENT '分享码',
  `SHARE_EXPIRE` datetime DEFAULT NULL COMMENT '分享过期时间',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `FILE_COUNT` int DEFAULT 0 COMMENT '文件数量',
  `TOTAL_SIZE` bigint DEFAULT 0 COMMENT '总大小',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_parent_id` (`PARENT_ID`),
  KEY `idx_owner_id` (`OWNER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云盘文件夹表';

-- ==========================================
-- 5. live_callback_event - 添加缺失的标准字段
-- ==========================================
ALTER TABLE `live_callback_event` 
  ADD COLUMN IF NOT EXISTS `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人' AFTER `IS_ACTIVE`,
  ADD COLUMN IF NOT EXISTS `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人' AFTER `CREATE_BY`,
  ADD COLUMN IF NOT EXISTS `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' AFTER `UPDATE_BY`;

-- ==========================================
-- 6. pull_stream_task - 添加缺失的标准字段
-- ==========================================
ALTER TABLE `pull_stream_task` 
  ADD COLUMN IF NOT EXISTS `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID' AFTER `ID`;

-- ==========================================
-- 7. SYS_COMPANY_CONF - 添加缺失的列
-- ==========================================
ALTER TABLE `sys_company_conf` 
  ADD COLUMN IF NOT EXISTS `TIMEZONE` varchar(50) DEFAULT 'Asia/Shanghai' COMMENT '时区' AFTER `CURRENCY`,
  ADD COLUMN IF NOT EXISTS `CURRENCY` varchar(10) DEFAULT 'CNY' COMMENT '货币' AFTER `TIMEZONE`,
  ADD COLUMN IF NOT EXISTS `MAX_STORAGE_SIZE` bigint DEFAULT 0 COMMENT '最大存储大小' AFTER `CURRENCY`,
  ADD COLUMN IF NOT EXISTS `MAX_FILE_SIZE` bigint DEFAULT 0 COMMENT '最大文件大小' AFTER `MAX_STORAGE_SIZE`,
  ADD COLUMN IF NOT EXISTS `MAX_FILES` int DEFAULT 0 COMMENT '最大文件数' AFTER `MAX_FILE_SIZE`;

-- ==========================================
-- 8. SYS_USER_SESSION - 添加缺失的标准字段
-- ==========================================
ALTER TABLE `sys_user_session` 
  ADD COLUMN IF NOT EXISTS `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人' AFTER `IS_ACTIVE`,
  ADD COLUMN IF NOT EXISTS `CREATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间' AFTER `CREATE_BY`,
  ADD COLUMN IF NOT EXISTS `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人' AFTER `CREATE_TIME`,
  ADD COLUMN IF NOT EXISTS `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' AFTER `UPDATE_BY`;

-- ==========================================
-- 9. sys_language - 创建物理表（存在迁移SQL但未执行）
-- ==========================================
CREATE TABLE IF NOT EXISTS `sys_language` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `CODE` varchar(50) NOT NULL COMMENT '语言代码',
  `NAME` varchar(255) NOT NULL COMMENT '语言名称',
  `IS_DEFAULT` char(1) DEFAULT 'N' COMMENT '是否默认',
  `IS_ACTIVE` char(1) DEFAULT 'Y' COMMENT '是否有效',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='语言配置表';

-- ==========================================
-- 10. sys_translation - 创建物理表（存在迁移SQL但未执行）
-- ==========================================
CREATE TABLE IF NOT EXISTS `sys_translation` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `KEY_NAME` varchar(255) NOT NULL COMMENT '翻译键',
  `LANGUAGE` varchar(50) NOT NULL COMMENT '语言代码',
  `VALUE` text COMMENT '翻译值',
  `MODULE` varchar(50) DEFAULT NULL COMMENT '所属模块',
  `IS_ACTIVE` char(1) DEFAULT 'Y' COMMENT '是否有效',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='翻译配置表';

-- ==========================================
-- 11. live_push_domain_config - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_push_domain_config` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `DOMAIN_NAME` varchar(255) NOT NULL COMMENT '推流域名',
  `STREAM_KEY` varchar(255) NOT NULL COMMENT '推流密钥',
  `APP_NAME` varchar(100) NOT NULL DEFAULT 'live' COMMENT '应用名称',
  `PLAY_DOMAIN` varchar(255) DEFAULT NULL COMMENT '关联的播放域名',
  `IS_DEFAULT` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为默认域名',
  `IS_ACTIVE` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `REMARK` varchar(500) DEFAULT NULL COMMENT '备注说明',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`ID`),
  INDEX `idx_company_id` (`SYS_COMPANY_ID`),
  UNIQUE INDEX `uk_company_domain` (`SYS_COMPANY_ID`, `DOMAIN_NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播推流域名配置表';

-- ==========================================
-- 12. live_domain - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_domain` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `DOMAIN_NAME` varchar(255) NOT NULL COMMENT '域名',
  `DOMAIN_TYPE` varchar(50) NOT NULL DEFAULT 'push' COMMENT '域名类型',
  `CNAME_STATUS` varchar(50) DEFAULT 'unknown' COMMENT 'CNAME状态',
  `STATUS` varchar(50) NOT NULL DEFAULT 'draft' COMMENT '状态',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播域名表';

-- ==========================================
-- 13. live_stream - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_stream` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `DOMAIN_ID` bigint DEFAULT NULL COMMENT '域名ID',
  `ROOM_ID` bigint DEFAULT NULL COMMENT '直播间ID',
  `STREAM_NAME` varchar(255) DEFAULT NULL COMMENT '流名称',
  `STREAM_ID` varchar(255) DEFAULT NULL COMMENT '流ID',
  `PUSH_URL` varchar(500) DEFAULT NULL COMMENT '推流地址',
  `PLAY_URL` varchar(500) DEFAULT NULL COMMENT '播放地址',
  `STATUS` varchar(50) NOT NULL DEFAULT 'inactive' COMMENT '状态',
  `BITRATE` int DEFAULT 0 COMMENT '码率',
  `FRAME_RATE` int DEFAULT 0 COMMENT '帧率',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播流表';

-- ==========================================
-- 14. live_pull - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_pull` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `SOURCE_URL` varchar(500) DEFAULT NULL COMMENT '源流地址',
  `TARGET_PLATFORM` varchar(100) DEFAULT NULL COMMENT '目标平台',
  `TARGET_URL` varchar(500) DEFAULT NULL COMMENT '目标推流地址',
  `STATUS` varchar(50) NOT NULL DEFAULT 'draft' COMMENT '状态',
  `START_TIME` datetime DEFAULT NULL COMMENT '开始时间',
  `END_TIME` datetime DEFAULT NULL COMMENT '结束时间',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拉流分发表';

-- ==========================================
-- 15. live_cut - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_cut` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `ROOM_ID` bigint DEFAULT NULL COMMENT '直播间ID',
  `CLIP_NAME` varchar(255) NOT NULL COMMENT '片段名称',
  `CLIP_URL` varchar(500) DEFAULT NULL COMMENT '片段URL',
  `START_TIME` bigint DEFAULT 0 COMMENT '开始时间戳',
  `END_TIME` bigint DEFAULT 0 COMMENT '结束时间戳',
  `DURATION` int DEFAULT 0 COMMENT '时长',
  `THUMBNAIL` varchar(500) DEFAULT NULL COMMENT '缩略图',
  `FILE_SIZE` bigint DEFAULT 0 COMMENT '文件大小',
  `STATUS` varchar(50) NOT NULL DEFAULT 'processing' COMMENT '状态',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_room_id` (`ROOM_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播精彩片段表';

-- ==========================================
-- 16. live_recode - 创建物理表
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_recode` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL DEFAULT 1 COMMENT '公司ID',
  `ROOM_ID` bigint DEFAULT NULL COMMENT '直播间ID',
  `STREAM_ID` varchar(255) DEFAULT NULL COMMENT '流ID',
  `RECORD_URL` varchar(500) DEFAULT NULL COMMENT '录制文件URL',
  `RECORD_TYPE` varchar(50) DEFAULT 'flv' COMMENT '录制格式',
  `FILE_SIZE` bigint DEFAULT 0 COMMENT '文件大小',
  `DURATION` int DEFAULT 0 COMMENT '时长',
  `START_TIME` bigint DEFAULT 0 COMMENT '开始时间戳',
  `END_TIME` bigint DEFAULT 0 COMMENT '结束时间戳',
  `STATUS` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_room_id` (`ROOM_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播录制表';

-- ==========================================
-- 验证脚本
-- ==========================================
SELECT 'Tables created and columns added successfully' AS result;
SHOW TABLES LIKE 'cloud_file';
SHOW TABLES LIKE 'cloud_folder';
SHOW TABLES LIKE 'sys_language';
SHOW TABLES LIKE 'sys_translation';
SHOW TABLES LIKE 'live_domain';
SHOW TABLES LIKE 'live_stream';
SHOW TABLES LIKE 'live_pull';
SHOW TABLES LIKE 'live_cut';
SHOW TABLES LIKE 'live_recode';
SHOW TABLES LIKE 'live_push_domain_config';
