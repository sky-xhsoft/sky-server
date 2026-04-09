-- =================================================================
-- Sky-Server 开发环境数据库清理脚本
-- =================================================================
-- 清理所有日志数据，保留元数据配置和基础数据
--
-- Usage:
--   mysql -u root -p skyserver < sqls/cleanup_dev_logs.sql
-- =================================================================

USE skyserver;

-- 禁用外键检查
SET FOREIGN_KEY_CHECKS = 0;

-- =================================================================
-- 清理日志和运行时数据
-- =================================================================

-- 审计日志 - 完全清理
TRUNCATE TABLE `audit_log`;

-- 通知日志 - 完全清理
TRUNCATE TABLE `sys_notification_log`;

-- 用户会话 - 完全清理
TRUNCATE TABLE `sys_user_session`;

-- 用户消息 - 完全清理
TRUNCATE TABLE `sys_user_message`;

-- 云盘分片上传记录 - 完全清理
TRUNCATE TABLE `cloud_chunk_record`;

-- 云盘上传会话 - 完全清理
TRUNCATE TABLE `cloud_upload_session`;

-- 云盘分享记录 - 保留基础数据，只清理过期的（可选）
-- DELETE FROM `cloud_share` WHERE expire_time < NOW();

-- 直播回调事件 - 完全清理
TRUNCATE TABLE `live_callback_event`;

-- 拉流任务 - 完全清理
TRUNCATE TABLE `pull_stream_task`;

-- 工作流实例 - 完全清理
TRUNCATE TABLE `wf_instance`;

-- 工作流任务 - 完全清理
TRUNCATE TABLE `wf_task`;

-- 工作流转换记录 - 完全清理（如果有）
-- TRUNCATE TABLE `wf_transition`;

-- =================================================================
-- 启用外键检查
-- =================================================================
SET FOREIGN_KEY_CHECKS = 1;

-- =================================================================
-- 显示清理后的数据统计
-- =================================================================
SELECT '清理完成！当前数据统计：' AS info;

SELECT
  'audit_log' as table_name, COUNT(*) as row_count FROM audit_log
UNION ALL
SELECT 'sys_notification_log', COUNT(*) FROM sys_notification_log
UNION ALL
SELECT 'sys_user_session', COUNT(*) FROM sys_user_session
UNION ALL
SELECT 'cloud_upload_session', COUNT(*) FROM cloud_upload_session
UNION ALL
SELECT 'live_callback_event', COUNT(*) FROM live_callback_event;

SELECT '========================================' AS separator;
SELECT '清理成功！数据库已重置为干净状态。' AS result;
