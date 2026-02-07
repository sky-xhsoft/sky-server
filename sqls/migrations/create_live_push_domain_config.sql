-- ==========================================
-- 直播推流域名配置表创建脚本
-- ==========================================
-- 用途：存储推流域名及其对应的密钥配置，支持多域名管理
-- 日期：2026-02-05
-- ==========================================

-- 创建直播推流域名配置表
CREATE TABLE IF NOT EXISTS `live_push_domain_config` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `DOMAIN_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '推流域名',
  `STREAM_KEY` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '推流密钥',
  `APP_NAME` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'live' COMMENT '应用名称',
  `PLAY_DOMAIN` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联的播放域名',
  `IS_DEFAULT` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为默认域名（0-否，1-是）',
  `IS_ACTIVE` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用（0-禁用，1-启用）',
  `REMARK` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注说明',
  `CREATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `CREATED_BY` bigint NULL DEFAULT NULL COMMENT '创建人ID',
  `UPDATED_BY` bigint NULL DEFAULT NULL COMMENT '更新人ID',
  PRIMARY KEY (`ID`) USING BTREE,
  INDEX `idx_company_id`(`SYS_COMPANY_ID` ASC) USING BTREE,
  INDEX `idx_domain_name`(`DOMAIN_NAME` ASC) USING BTREE,
  UNIQUE INDEX `uk_company_domain`(`SYS_COMPANY_ID`, `DOMAIN_NAME`) USING BTREE COMMENT '同一公司下域名唯一'
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '直播推流域名配置表' ROW_FORMAT = Dynamic;

-- ==========================================
-- 使用说明
-- ==========================================

/*
字段说明：
- ID: 主键ID
- SYS_COMPANY_ID: 公司ID，关联sys_company表
- DOMAIN_NAME: 推流域名，如 push.example.com
- STREAM_KEY: 推流密钥，用于生成鉴权签名
- APP_NAME: 应用名称，默认为live
- PLAY_DOMAIN: 关联的播放域名，可选
- IS_DEFAULT: 是否为默认域名，每个公司只能有一个默认域名
- IS_ACTIVE: 是否启用
- REMARK: 备注说明

使用示例：

1. 添加推流域名配置：
INSERT INTO live_push_domain_config (SYS_COMPANY_ID, DOMAIN_NAME, STREAM_KEY, APP_NAME, PLAY_DOMAIN, IS_DEFAULT, REMARK)
VALUES (1, 'push1.example.com', 'your-stream-key-1', 'live', 'play1.example.com', 1, '主推流域名');

INSERT INTO live_push_domain_config (SYS_COMPANY_ID, DOMAIN_NAME, STREAM_KEY, APP_NAME, PLAY_DOMAIN, IS_DEFAULT, REMARK)
VALUES (1, 'push2.example.com', 'your-stream-key-2', 'live', 'play2.example.com', 0, '备用推流域名');

2. 查询公司的推流域名配置：
SELECT * FROM live_push_domain_config WHERE SYS_COMPANY_ID = 1 AND IS_ACTIVE = 1;

3. 获取默认推流域名：
SELECT * FROM live_push_domain_config WHERE SYS_COMPANY_ID = 1 AND IS_DEFAULT = 1 AND IS_ACTIVE = 1;

4. 更新默认域名（先取消其他默认，再设置新默认）：
UPDATE live_push_domain_config SET IS_DEFAULT = 0 WHERE SYS_COMPANY_ID = 1;
UPDATE live_push_domain_config SET IS_DEFAULT = 1 WHERE ID = 2;

5. 禁用域名：
UPDATE live_push_domain_config SET IS_ACTIVE = 0 WHERE ID = 1;

注意事项：
- 每个公司可以配置多个推流域名
- 同一公司下域名名称必须唯一
- 建议每个公司只设置一个默认域名
- 密钥信息敏感，建议加密存储（可后续优化）
*/

-- ==========================================
-- 验证脚本
-- ==========================================

-- 查看表结构
DESC live_push_domain_config;

-- 查看索引
SHOW INDEX FROM live_push_domain_config;

-- 查看现有配置
SELECT ID, SYS_COMPANY_ID, DOMAIN_NAME, APP_NAME, IS_DEFAULT, IS_ACTIVE, CREATED_AT
FROM live_push_domain_config;
