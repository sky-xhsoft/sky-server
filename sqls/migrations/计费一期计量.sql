-- 直播计费一期：只计量不扣费（直播时长、播放流量、存储空间）
SET NAMES utf8mb4;

-- 推流段落：一次推流到断流，按直播间记录
CREATE TABLE IF NOT EXISTS `bill_live_session` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `ROOM_ID` bigint unsigned NOT NULL,
  `STREAM_NAME` varchar(128) NOT NULL,
  `PUSH_DOMAIN` varchar(255) DEFAULT NULL,
  `START_TIME` datetime NOT NULL,
  `END_TIME` datetime DEFAULT NULL,
  `LAST_SEEN_TIME` datetime NOT NULL COMMENT '巡检最后确认在线的时间',
  `SECONDS` bigint NOT NULL DEFAULT 0,
  `END_SOURCE` varchar(20) DEFAULT NULL COMMENT 'callback 断流回调 / patrol 巡检 / cap 超过最长时长',
  `PUSH_EVENT_ID` bigint NOT NULL COMMENT '推流回调事件 ID；由断流回调补出的段落为负的断流事件 ID',
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_BLS_PUSH_EVENT` (`PUSH_EVENT_ID`),
  KEY `IDX_BLS_COMPANY_START` (`SYS_COMPANY_ID`, `START_TIME`),
  KEY `IDX_BLS_ROOM` (`ROOM_ID`),
  KEY `IDX_BLS_OPEN` (`END_TIME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：推流段落';

-- 日用量：每个公司、直播间（存储为 0）、计费点、细项一行
CREATE TABLE IF NOT EXISTS `bill_usage_daily` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `USAGE_DATE` char(10) NOT NULL,
  `METER` varchar(32) NOT NULL COMMENT 'live_duration 直播时长 / play_flux 播放流量 / storage 存储空间',
  `ITEM` varchar(32) NOT NULL DEFAULT '' COMMENT '存储细项：record 录制 / clip 切片 / disk 云盘',
  `ROOM_ID` bigint unsigned NOT NULL DEFAULT 0,
  `QUANTITY` decimal(20,6) NOT NULL COMMENT '计费用量：分钟 / GB',
  `UNIT` varchar(16) NOT NULL,
  `RAW_QUANTITY` decimal(20,6) NOT NULL DEFAULT 0 COMMENT '时长：秒；流量：分摊前的监控流量 GB；存储：字节',
  `SOURCE_KEY` varchar(191) NOT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_BUD_SOURCE` (`SOURCE_KEY`),
  KEY `IDX_BUD_COMPANY_DATE` (`SYS_COMPANY_ID`, `USAGE_DATE`),
  KEY `IDX_BUD_DATE_METER` (`USAGE_DATE`, `METER`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：日用量';

-- 播放流量对账：腾讯云账号 + 播放域名 + 日期
CREATE TABLE IF NOT EXISTS `bill_flux_reconcile` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `USAGE_DATE` char(10) NOT NULL,
  `ACCOUNT` varchar(64) NOT NULL COMMENT 'global 平台账号 / tc:<密钥 ID 摘要> 公司自有账号',
  `PLAY_DOMAIN` varchar(255) NOT NULL,
  `MONITOR_MB` decimal(20,3) NOT NULL COMMENT '按流监控流量之和',
  `BILLING_MB` decimal(20,3) NOT NULL COMMENT '域名计费流量',
  `UNATTRIBUTED_MB` decimal(20,3) NOT NULL COMMENT '找不到直播间的流（分摊后）',
  `STREAM_COUNT` int NOT NULL,
  `STATUS` varchar(20) NOT NULL COMMENT 'ok / diff 差异超过 3% / error 拉取失败',
  `MESSAGE` varchar(500) DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_BFR_DAY` (`USAGE_DATE`, `ACCOUNT`, `PLAY_DOMAIN`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：播放流量对账';

-- 计量任务状态（回调处理游标、每日计量完成标记）
CREATE TABLE IF NOT EXISTS `bill_meter_state` (
  `STATE_KEY` varchar(64) NOT NULL,
  `STATE_VALUE` varchar(255) NOT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`STATE_KEY`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：计量任务状态';

-- 菜单：直播中心 →「费用中心」（一期只有用量；权限目录与直播间管理相同）
INSERT INTO `sys_table` (
  `NAME`, `DISPLAY_NAME`, `ENAME`, `MASK`, `SYS_TABLECATEGORY_ID`, `ORDERNO`, `URL`, `IS_MENU`, `ICO_IMG`,
  `SYS_DIRECTORY_ID`, `IS_BIG`, `DESCRIPTION`,
  `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
)
SELECT 'LIVE_BILLING', '费用中心', 'Billing', 'Q', d.SYS_TABLECATEGORY_ID, 70, '/live/billing', 'Y', 'icon-bar-chart',
       d.SYS_DIRECTORY_ID, 'N', '本公司直播时长、播放流量、存储空间用量',
       0, 'system', NOW(), 'system', NOW(), 'Y'
FROM `sys_table` d
WHERE d.NAME = 'LIVE_ROOM'
  AND NOT EXISTS (SELECT 1 FROM `sys_table` t WHERE t.NAME = 'LIVE_BILLING');
