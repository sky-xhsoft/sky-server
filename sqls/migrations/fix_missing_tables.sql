-- ==========================================
-- 修复缺失的物理表 和 sys_column 不一致问题
-- 生成时间: 2026-05-13
-- ==========================================

-- ==========================================
-- 1. 创建 live_push_domain_config 表（推流域名配置）
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_push_domain_config` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `DOMAIN_NAME` varchar(255) NOT NULL COMMENT '推流域名',
  `STREAM_KEY` varchar(255) NOT NULL COMMENT '推流密钥',
  `APP_NAME` varchar(100) NOT NULL DEFAULT 'live' COMMENT '应用名称',
  `PLAY_DOMAIN` varchar(255) DEFAULT NULL COMMENT '关联的播放域名',
  `IS_DEFAULT` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为默认域名',
  `IS_ACTIVE` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `REMARK` varchar(500) DEFAULT NULL COMMENT '备注说明',
  `CREATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `CREATED_BY` bigint DEFAULT NULL COMMENT '创建人ID',
  `UPDATED_BY` bigint DEFAULT NULL COMMENT '更新人ID',
  PRIMARY KEY (`ID`),
  INDEX `idx_company_id` (`SYS_COMPANY_ID`),
  INDEX `idx_domain_name` (`DOMAIN_NAME`),
  UNIQUE INDEX `uk_company_domain` (`SYS_COMPANY_ID`, `DOMAIN_NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播推流域名配置表';

-- ==========================================
-- 2. 创建 live_domain 表（直播域名管理）
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_domain` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `DOMAIN_NAME` varchar(255) NOT NULL COMMENT '域名',
  `DOMAIN_TYPE` varchar(50) NOT NULL DEFAULT 'push' COMMENT '域名类型: push(推流), play(播放)',
  `CNAME_STATUS` varchar(50) DEFAULT 'unknown' COMMENT 'CNAME状态',
  `STATUS` varchar(50) NOT NULL DEFAULT 'draft' COMMENT '状态: draft, enabled, disabled',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_domain_type` (`DOMAIN_TYPE`),
  KEY `idx_status` (`STATUS`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播域名表';

-- ==========================================
-- 3. 创建 live_stream 表（直播流管理）
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_stream` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `DOMAIN_ID` bigint DEFAULT NULL COMMENT '域名ID',
  `ROOM_ID` bigint DEFAULT NULL COMMENT '直播间ID',
  `STREAM_NAME` varchar(255) DEFAULT NULL COMMENT '流名称',
  `STREAM_ID` varchar(255) DEFAULT NULL COMMENT '流ID',
  `PUSH_URL` varchar(500) DEFAULT NULL COMMENT '推流地址',
  `PLAY_URL` varchar(500) DEFAULT NULL COMMENT '播放地址',
  `STATUS` varchar(50) NOT NULL DEFAULT 'inactive' COMMENT '状态: active, inactive, banned',
  `BITRATE` int DEFAULT 0 COMMENT '码率(kbps)',
  `FRAME_RATE` int DEFAULT 0 COMMENT '帧率(fps)',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_domain_id` (`DOMAIN_ID`),
  KEY `idx_room_id` (`ROOM_ID`),
  KEY `idx_stream_name` (`STREAM_NAME`),
  KEY `idx_status` (`STATUS`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播流表';

-- ==========================================
-- 4. 创建 live_pull 表（拉流分发）
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_pull` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `TASK_ID` varchar(255) DEFAULT NULL COMMENT '拉流任务ID',
  `SOURCE_URL` varchar(500) DEFAULT NULL COMMENT '源流地址',
  `TARGET_PLATFORM` varchar(100) DEFAULT NULL COMMENT '目标平台: bilibili, douyin, kuaishou等',
  `TARGET_URL` varchar(500) DEFAULT NULL COMMENT '目标推流地址',
  `TARGET_KEY` varchar(255) DEFAULT NULL COMMENT '目标推流密钥',
  `STATUS` varchar(50) NOT NULL DEFAULT 'draft' COMMENT '状态: draft, pulling, stopped, error',
  `START_TIME` datetime DEFAULT NULL COMMENT '开始时间',
  `END_TIME` datetime DEFAULT NULL COMMENT '结束时间',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_status` (`STATUS`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拉流分发表';

-- ==========================================
-- 5. 创建 live_cut 表（直播精彩片段/切片）
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_cut` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `ROOM_ID` bigint DEFAULT NULL COMMENT '直播间ID',
  `STREAM_ID` varchar(255) DEFAULT NULL COMMENT '流ID',
  `CLIP_NAME` varchar(255) NOT NULL COMMENT '片段名称',
  `CLIP_URL` varchar(500) DEFAULT NULL COMMENT '片段URL',
  `START_TIME` bigint DEFAULT 0 COMMENT '开始时间戳(秒)',
  `END_TIME` bigint DEFAULT 0 COMMENT '结束时间戳(秒)',
  `DURATION` int DEFAULT 0 COMMENT '时长(秒)',
  `THUMBNAIL` varchar(500) DEFAULT NULL COMMENT '缩略图',
  `FILE_SIZE` bigint DEFAULT 0 COMMENT '文件大小(字节)',
  `STATUS` varchar(50) NOT NULL DEFAULT 'processing' COMMENT '状态: processing, done, failed',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_room_id` (`ROOM_ID`),
  KEY `idx_status` (`STATUS`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播精彩片段表';

-- ==========================================
-- 6. 创建 live_recode 表（直播录制管理）
-- ==========================================
CREATE TABLE IF NOT EXISTS `live_recode` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `ROOM_ID` bigint DEFAULT NULL COMMENT '直播间ID',
  `STREAM_ID` varchar(255) DEFAULT NULL COMMENT '流ID',
  `RECORD_URL` varchar(500) DEFAULT NULL COMMENT '录制文件URL',
  `RECORD_TYPE` varchar(50) DEFAULT 'flv' COMMENT '录制格式: flv, mp4, hls',
  `FILE_SIZE` bigint DEFAULT 0 COMMENT '文件大小(字节)',
  `DURATION` int DEFAULT 0 COMMENT '时长(秒)',
  `START_TIME` bigint DEFAULT 0 COMMENT '开始时间戳(秒)',
  `END_TIME` bigint DEFAULT 0 COMMENT '结束时间戳(秒)',
  `STATUS` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '状态: pending, recording, done, failed',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_room_id` (`ROOM_ID`),
  KEY `idx_status` (`STATUS`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='直播录制表';
