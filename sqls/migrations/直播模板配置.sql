-- 直播模板配置：回调模板、录制模板与域名绑定规则（本地为准，保存后同步到公司的腾讯云账号）
-- 新增「模板管理」菜单（直播中心，与直播域名同一权限目录）
-- 幂等：可重复执行
-- 执行后刷新元数据缓存：POST /api/v1/metadata/refresh（或删除 Redis 中 metadata:* 键）
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `live_callback_template` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属公司',
  `NAME` varchar(255) NOT NULL COMMENT '模板名称（仅中文、英文、数字、_、-）',
  `DESCRIPTION` varchar(1024) DEFAULT NULL COMMENT '描述',
  `STREAM_BEGIN_URL` varchar(500) DEFAULT NULL COMMENT '开播回调',
  `STREAM_END_URL` varchar(500) DEFAULT NULL COMMENT '断流回调',
  `RECORD_URL` varchar(500) DEFAULT NULL COMMENT '录制文件回调',
  `RECORD_STATUS_URL` varchar(500) DEFAULT NULL COMMENT '录制状态回调',
  `RECORD_EXCEPTION_URL` varchar(500) DEFAULT NULL COMMENT '录制异常回调',
  `RECORD_EXCEPTION_LEVELS` varchar(100) DEFAULT NULL COMMENT '录制异常级别，逗号分隔：error,alarm',
  `PUSH_EXCEPTION_URL` varchar(500) DEFAULT NULL COMMENT '推流异常回调',
  `CALLBACK_KEY` varchar(500) DEFAULT NULL COMMENT '回调密钥（加密存储）',
  `CLOUD_TEMPLATE_ID` bigint NOT NULL DEFAULT 0 COMMENT '腾讯云模板ID，0 为未同步',
  `SOURCE` varchar(20) NOT NULL DEFAULT 'local' COMMENT '来源：local 本系统创建，import 从腾讯云导入，standard 标准配置',
  `SYNC_STATUS` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '同步状态：pending/synced/failed/drift',
  `SYNC_ERROR` varchar(500) DEFAULT NULL COMMENT '同步失败原因',
  `SYNC_TIME` datetime DEFAULT NULL COMMENT '最后同步时间',
  `CREATE_BY` varchar(255) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_BY` varchar(255) DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  `IS_ACTIVE` char(1) DEFAULT 'Y',
  PRIMARY KEY (`ID`),
  KEY `IDX_LCT_COMPANY` (`SYS_COMPANY_ID`),
  KEY `IDX_LCT_CLOUD` (`SYS_COMPANY_ID`, `CLOUD_TEMPLATE_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='直播回调模板';

CREATE TABLE IF NOT EXISTS `live_record_template` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属公司',
  `NAME` varchar(255) NOT NULL COMMENT '模板名称',
  `DESCRIPTION` varchar(1024) DEFAULT NULL COMMENT '描述',
  `COS_STORE` char(1) NOT NULL DEFAULT 'Y' COMMENT 'Y 存储至 COS；N 云点播 VOD（仅导入）',
  `COS_BUCKET` varchar(255) DEFAULT NULL COMMENT 'COS 桶名（不含 -appid）',
  `COS_REGION` varchar(50) DEFAULT NULL COMMENT 'COS 地域',
  `COS_PATH` varchar(500) DEFAULT NULL COMMENT '存储路径，以 / 开头并含时间占位符',
  `RECORD_INTERVAL` bigint NOT NULL DEFAULT 1800 COMMENT 'MP4 分段时长（秒，60-43200）',
  `REMOVE_WATERMARK` char(1) NOT NULL DEFAULT 'Y' COMMENT 'Y 录制原始流',
  `FORMATS` varchar(50) DEFAULT 'mp4' COMMENT '录制格式，逗号分隔',
  `CLOUD_TEMPLATE_ID` bigint NOT NULL DEFAULT 0 COMMENT '腾讯云模板ID，0 为未同步',
  `SOURCE` varchar(20) NOT NULL DEFAULT 'local' COMMENT '来源：local/import/standard',
  `SYNC_STATUS` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '同步状态：pending/synced/failed/drift',
  `SYNC_ERROR` varchar(500) DEFAULT NULL COMMENT '同步失败原因',
  `SYNC_TIME` datetime DEFAULT NULL COMMENT '最后同步时间',
  `CREATE_BY` varchar(255) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_BY` varchar(255) DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  `IS_ACTIVE` char(1) DEFAULT 'Y',
  PRIMARY KEY (`ID`),
  KEY `IDX_LRT_COMPANY` (`SYS_COMPANY_ID`),
  KEY `IDX_LRT_CLOUD` (`SYS_COMPANY_ID`, `CLOUD_TEMPLATE_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='直播录制模板';

CREATE TABLE IF NOT EXISTS `live_template_rule` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属公司',
  `KIND` varchar(20) NOT NULL COMMENT 'callback 回调规则 / record 录制规则',
  `DOMAIN_NAME` varchar(128) NOT NULL COMMENT '推流域名',
  `APP_NAME` varchar(64) NOT NULL DEFAULT '' COMMENT '推流路径 AppName',
  `STREAM_NAME` varchar(128) NOT NULL DEFAULT '' COMMENT '流名称（仅录制规则，空为全部流）',
  `TEMPLATE_ID` int unsigned NOT NULL COMMENT '本地模板ID',
  `SYNC_STATUS` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '同步状态：pending/synced/failed/drift',
  `SYNC_ERROR` varchar(500) DEFAULT NULL COMMENT '同步失败原因',
  `SYNC_TIME` datetime DEFAULT NULL COMMENT '最后同步时间',
  `CREATE_BY` varchar(255) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_BY` varchar(255) DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_LTR_RULE` (`SYS_COMPANY_ID`, `KIND`, `DOMAIN_NAME`, `APP_NAME`, `STREAM_NAME`),
  KEY `IDX_LTR_TEMPLATE` (`KIND`, `TEMPLATE_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='直播模板绑定规则';

-- 菜单：直播中心 →「模板管理」（权限目录与直播域名相同）
INSERT INTO `sys_table` (
  `NAME`, `DISPLAY_NAME`, `ENAME`, `MASK`, `SYS_TABLECATEGORY_ID`, `ORDERNO`, `URL`, `IS_MENU`, `ICO_IMG`,
  `SYS_DIRECTORY_ID`, `IS_BIG`, `DESCRIPTION`,
  `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
)
SELECT 'LIVE_TEMPLATE', '模板管理', 'Live Templates', 'AMDQ', d.SYS_TABLECATEGORY_ID, 15, '/live/templates', 'Y', 'icon-file',
       d.SYS_DIRECTORY_ID, 'N', '直播回调、录制模板管理（同步到腾讯云）',
       0, 'system', NOW(), 'system', NOW(), 'Y'
FROM `sys_table` d
WHERE d.NAME = 'LIVE_DOMAIN'
  AND NOT EXISTS (SELECT 1 FROM `sys_table` t WHERE t.NAME = 'LIVE_TEMPLATE');
