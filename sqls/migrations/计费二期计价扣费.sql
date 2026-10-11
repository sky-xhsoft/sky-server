-- 直播计费二期：价格、套餐、资源包、账户、费用明细、余额流水、月账单、赠送 / 调账复核
-- 金额以「厘」（0.001 元）整数存储；单价以 0.0001 元为单位
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `bill_price` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL DEFAULT 0 COMMENT '0 为默认价，否则为该公司的专属价',
  `METER` varchar(32) NOT NULL,
  `UNIT_PRICE` bigint NOT NULL COMMENT '0.0001 元 / 计价单位（存储为 GB·月）',
  `EFFECTIVE_DATE` char(10) NOT NULL,
  `REMARK` varchar(255) DEFAULT NULL,
  `CREATE_BY` varchar(80) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_BP_LOOKUP` (`SYS_COMPANY_ID`, `METER`, `EFFECTIVE_DATE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：单价';

CREATE TABLE IF NOT EXISTS `bill_plan` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `NAME` varchar(64) NOT NULL,
  `MONTHLY_FEE` bigint NOT NULL COMMENT '厘',
  `QUOTA_MINUTES` decimal(20,6) NOT NULL DEFAULT 0,
  `QUOTA_FLUX_GB` decimal(20,6) NOT NULL DEFAULT 0,
  `QUOTA_STORAGE_GB` decimal(20,6) NOT NULL DEFAULT 0 COMMENT '日均存储额度',
  `STATUS` varchar(10) NOT NULL DEFAULT 'on' COMMENT 'on 在售 / off 下架',
  `SORT_NO` int NOT NULL DEFAULT 0,
  `DESCRIPTION` varchar(500) DEFAULT NULL,
  `UPDATE_BY` varchar(80) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：套餐';

CREATE TABLE IF NOT EXISTS `bill_pack_def` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `NAME` varchar(64) NOT NULL,
  `METER` varchar(32) NOT NULL COMMENT 'live_duration / play_flux',
  `QUANTITY` decimal(20,6) NOT NULL,
  `VALID_MONTHS` int NOT NULL,
  `PRICE` bigint NOT NULL COMMENT '厘',
  `STATUS` varchar(10) NOT NULL DEFAULT 'on',
  `SORT_NO` int NOT NULL DEFAULT 0,
  `DESCRIPTION` varchar(500) DEFAULT NULL,
  `UPDATE_BY` varchar(80) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：资源包定义';

CREATE TABLE IF NOT EXISTS `bill_account` (
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `BALANCE` bigint NOT NULL DEFAULT 0 COMMENT '厘，可为负',
  `GIFT_BALANCE` bigint NOT NULL DEFAULT 0 COMMENT '余额中来自赠送的部分',
  `CREDIT_LIMIT` bigint NOT NULL DEFAULT 0 COMMENT '授信额度，厘',
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：公司账户';

CREATE TABLE IF NOT EXISTS `bill_subscription` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `PLAN_ID` bigint unsigned NOT NULL,
  `PLAN_NAME` varchar(64) NOT NULL,
  `START_DATE` char(10) NOT NULL,
  `END_DATE` char(10) NOT NULL COMMENT '含当天',
  `QUOTA_MINUTES` decimal(20,6) NOT NULL,
  `QUOTA_FLUX_GB` decimal(20,6) NOT NULL,
  `QUOTA_STORAGE_GB` decimal(20,6) NOT NULL,
  `USED_MINUTES` decimal(20,6) NOT NULL DEFAULT 0,
  `USED_FLUX_GB` decimal(20,6) NOT NULL DEFAULT 0,
  `FEE` bigint NOT NULL COMMENT '本周期已收套餐费（含升级补差），厘',
  `AUTO_RENEW` tinyint(1) NOT NULL DEFAULT 1,
  `STATUS` varchar(10) NOT NULL DEFAULT 'active' COMMENT 'active / ended',
  `CREATE_BY` varchar(80) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_BS_COMPANY` (`SYS_COMPANY_ID`, `START_DATE`),
  KEY `IDX_BS_DUE` (`STATUS`, `END_DATE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：套餐周期';

CREATE TABLE IF NOT EXISTS `bill_resource_pack` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `PACK_DEF_ID` bigint unsigned NOT NULL,
  `NAME` varchar(64) NOT NULL,
  `METER` varchar(32) NOT NULL,
  `TOTAL` decimal(20,6) NOT NULL,
  `REMAINING` decimal(20,6) NOT NULL,
  `VALID_FROM` char(10) NOT NULL,
  `VALID_TO` char(10) NOT NULL,
  `SOURCE` varchar(20) NOT NULL COMMENT 'purchase 购买 / grant 平台发放',
  `PRICE` bigint NOT NULL DEFAULT 0,
  `CREATE_BY` varchar(80) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_BRP_COMPANY` (`SYS_COMPANY_ID`, `METER`, `VALID_TO`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：公司资源包';

CREATE TABLE IF NOT EXISTS `bill_charge` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `USAGE_DATE` char(10) NOT NULL,
  `METER` varchar(32) NOT NULL,
  `USAGE_QTY` decimal(20,6) NOT NULL,
  `PLAN_OFFSET` decimal(20,6) NOT NULL DEFAULT 0,
  `PACK_OFFSET` decimal(20,6) NOT NULL DEFAULT 0,
  `PACK_DETAIL` varchar(1000) DEFAULT NULL,
  `BILLABLE` decimal(20,6) NOT NULL,
  `UNIT_PRICE` bigint NOT NULL,
  `AMOUNT` bigint NOT NULL COMMENT '应扣，厘',
  `CHARGED` bigint NOT NULL DEFAULT 0 COMMENT '已扣，厘',
  `MODE` varchar(10) NOT NULL COMMENT 'trial 试算 / charge 扣费',
  `SUBSCRIPTION_ID` bigint unsigned NOT NULL DEFAULT 0,
  `UPDATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_BC_DAY` (`SYS_COMPANY_ID`, `USAGE_DATE`, `METER`),
  KEY `IDX_BC_DATE` (`USAGE_DATE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：费用明细';

CREATE TABLE IF NOT EXISTS `bill_transaction` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `PERIOD` char(7) NOT NULL COMMENT '计入的账期',
  `TYPE` varchar(20) NOT NULL,
  `METER` varchar(32) NOT NULL DEFAULT '',
  `AMOUNT` bigint NOT NULL COMMENT '厘，正为收入',
  `GIFT_AMOUNT` bigint NOT NULL DEFAULT 0,
  `BALANCE_AFTER` bigint NOT NULL,
  `REF_TYPE` varchar(32) DEFAULT NULL,
  `REF_ID` bigint unsigned NOT NULL DEFAULT 0,
  `USAGE_DATE` char(10) NOT NULL DEFAULT '',
  `VOUCHER` varchar(128) DEFAULT NULL,
  `REMARK` varchar(500) DEFAULT NULL,
  `OPERATOR` varchar(80) DEFAULT NULL,
  `CREATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_BT_COMPANY` (`SYS_COMPANY_ID`, `PERIOD`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：余额流水';

CREATE TABLE IF NOT EXISTS `bill_statement` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `MONTH` char(7) NOT NULL,
  `OPENING` bigint NOT NULL,
  `USAGE_DURATION` bigint NOT NULL DEFAULT 0,
  `USAGE_FLUX` bigint NOT NULL DEFAULT 0,
  `USAGE_STORAGE` bigint NOT NULL DEFAULT 0,
  `USAGE_ADJUST` bigint NOT NULL DEFAULT 0 COMMENT '以往月份补扣（正）/ 退还（负）',
  `PLAN_FEE` bigint NOT NULL DEFAULT 0,
  `PACK_FEE` bigint NOT NULL DEFAULT 0,
  `RECHARGE` bigint NOT NULL DEFAULT 0,
  `GIFT` bigint NOT NULL DEFAULT 0,
  `ADJUST` bigint NOT NULL DEFAULT 0,
  `CLOSING` bigint NOT NULL,
  `DETAIL` mediumtext,
  `CREATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_BST_MONTH` (`SYS_COMPANY_ID`, `MONTH`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：月账单';

CREATE TABLE IF NOT EXISTS `bill_adjust_request` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` bigint unsigned NOT NULL,
  `TYPE` varchar(20) NOT NULL COMMENT 'gift / adjust',
  `AMOUNT` bigint NOT NULL,
  `REASON` varchar(500) NOT NULL,
  `STATUS` varchar(10) NOT NULL DEFAULT 'pending',
  `REQUEST_BY` varchar(80) NOT NULL,
  `REQUEST_BY_ID` bigint unsigned NOT NULL,
  `REVIEW_BY` varchar(80) DEFAULT NULL,
  `REVIEW_TIME` datetime DEFAULT NULL,
  `REVIEW_REMARK` varchar(500) DEFAULT NULL,
  `TXN_ID` bigint unsigned NOT NULL DEFAULT 0,
  `CREATE_TIME` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `IDX_BAR_COMPANY` (`SYS_COMPANY_ID`),
  KEY `IDX_BAR_STATUS` (`STATUS`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计费：赠送 / 调账复核';

-- 费用中心改用独立权限目录「直播费用」：公司管理员默认可见，其他用户需在权限组中授权
INSERT INTO `sys_directory` (`SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`, `NAME`, `DISPLAY_NAME`, `SYS_TABLE_CATEGORY_ID`, `SYS_TABLE_ID`)
SELECT 0, 'system', NOW(), 'system', NOW(), 'Y', 'LIVE_BILLING_LIST', '直播费用', t.SYS_TABLECATEGORY_ID, t.ID
FROM `sys_table` t
WHERE t.NAME = 'LIVE_BILLING'
  AND NOT EXISTS (SELECT 1 FROM `sys_directory` d WHERE d.NAME = 'LIVE_BILLING_LIST');
UPDATE `sys_table` t JOIN `sys_directory` d ON d.NAME = 'LIVE_BILLING_LIST'
SET t.SYS_DIRECTORY_ID = d.ID, t.MASK = 'AMQ', t.DESCRIPTION = '本公司费用、账单、余额流水、套餐与资源包'
WHERE t.NAME = 'LIVE_BILLING';

-- 系统中心 →「计费运营」→「计费管理」（仅平台公司可用）
INSERT INTO `sys_table_category` (`SYS_COMPANY_ID`, `IS_ACTIVE`, `SYS_SUBSYSTEM_ID`, `NAME`, `ENAME`, `ORDERNO`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`)
SELECT 0, 'Y', c.SYS_SUBSYSTEM_ID, '计费运营', 'Billing', 30, 'system', NOW(), 'system', NOW()
FROM `sys_table_category` c
WHERE c.NAME = '用户权限'
  AND NOT EXISTS (SELECT 1 FROM `sys_table_category` x WHERE x.NAME = '计费运营')
LIMIT 1;

INSERT INTO `sys_table` (
  `NAME`, `DISPLAY_NAME`, `ENAME`, `MASK`, `SYS_TABLECATEGORY_ID`, `ORDERNO`, `URL`, `IS_MENU`, `ICO_IMG`,
  `IS_BIG`, `DESCRIPTION`, `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
)
SELECT 'BILLING_ADMIN', '计费管理', 'Billing Admin', 'AMQ', c.ID, 10, '/system/billing', 'Y', 'icon-bar-chart',
       'N', '价格、套餐、资源包、公司账户、扣费设置与复核', 0, 'system', NOW(), 'system', NOW(), 'Y'
FROM `sys_table_category` c
WHERE c.NAME = '计费运营'
  AND NOT EXISTS (SELECT 1 FROM `sys_table` t WHERE t.NAME = 'BILLING_ADMIN');

INSERT INTO `sys_directory` (`SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`, `NAME`, `DISPLAY_NAME`, `SYS_TABLE_CATEGORY_ID`, `SYS_TABLE_ID`)
SELECT 0, 'system', NOW(), 'system', NOW(), 'Y', 'BILLING_ADMIN_LIST', '计费管理', t.SYS_TABLECATEGORY_ID, t.ID
FROM `sys_table` t
WHERE t.NAME = 'BILLING_ADMIN'
  AND NOT EXISTS (SELECT 1 FROM `sys_directory` d WHERE d.NAME = 'BILLING_ADMIN_LIST');
UPDATE `sys_table` t JOIN `sys_directory` d ON d.NAME = 'BILLING_ADMIN_LIST'
SET t.SYS_DIRECTORY_ID = d.ID
WHERE t.NAME = 'BILLING_ADMIN';
