-- ==========================================
-- sys_company 表添加 DOMAIN 字段迁移脚本
-- ==========================================
-- 用途：为现有数据库的 sys_company 表添加域名字段，支持多租户域名识别
-- 日期：2026-01-14
-- ==========================================

-- 1. 添加 DOMAIN 字段
ALTER TABLE `sys_company`
ADD COLUMN `DOMAIN` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '公司域名（用于多租户识别）' AFTER `NAME`;

-- 2. 为 DOMAIN 字段创建唯一索引
CREATE UNIQUE INDEX `idx_domain` ON `sys_company`(`DOMAIN` ASC) USING BTREE;

-- ==========================================
-- 使用说明
-- ==========================================

/*
字段说明：
- DOMAIN: 公司域名，用于多租户识别
- 允许为 NULL（可选配置）
- 唯一索引确保每个域名只能绑定一个公司

使用示例：

1. 为现有公司配置域名：
UPDATE sys_company SET DOMAIN = 'company1.example.com' WHERE ID = 1;
UPDATE sys_company SET DOMAIN = 'company2.example.com' WHERE ID = 2;

2. 域名格式支持：
   - 完整域名：company.example.com
   - 子域名：app.company.com
   - 本地域名：localhost:8080 (开发环境)
   - IP + 端口：192.168.1.100:8080

3. 多租户识别流程：
   - 用户访问 http://company1.example.com/api/xxx
   - 系统自动从请求 Host 头提取 "company1.example.com"
   - 查询 sys_company 表找到对应的公司 ID
   - 将公司 ID 设置到请求上下文中
   - 后续所有查询自动添加 SYS_COMPANY_ID 过滤

4. 注意事项：
   - DOMAIN 为 NULL 表示不使用域名识别（使用其他方式）
   - 确保 DNS 已正确配置指向服务器
   - 生产环境建议使用 HTTPS
   - 可以为同一公司配置多个域名（需要在代码层面支持）
*/

-- ==========================================
-- 验证脚本
-- ==========================================

-- 查看表结构
DESC sys_company;

-- 查看索引
SHOW INDEX FROM sys_company;

-- 查看现有公司及域名配置
SELECT ID, NAME, DOMAIN, IS_ACTIVE FROM sys_company;
