-- ============================================================================
-- 数据库性能优化 - 添加缺失索引
-- 为经常查询的字段添加适当的索引以提升查询性能
-- ============================================================================

-- sys_user 表 - 用户名登录查询
ALTER TABLE `sys_user` ADD INDEX IF NOT EXISTS `idx_username` (`USERNAME`);
ALTER TABLE `sys_user` ADD INDEX IF NOT EXISTS `idx_company_id` (`SYS_COMPANY_ID`);
ALTER TABLE `sys_user` ADD INDEX IF NOT EXISTS `idx_is_active` (`IS_ACTIVE`);

-- sys_column 表 - 按表ID查询字段
ALTER TABLE `sys_column` ADD INDEX IF NOT EXISTS `idx_sys_table_id` (`SYS_TABLE_ID`);

-- sys_table_ref 表 - 按主表或关联表查询
ALTER TABLE `sys_table_ref` ADD INDEX IF NOT EXISTS `idx_sys_table_id` (`SYS_TABLE_ID`);
ALTER TABLE `sys_table_ref` ADD INDEX IF NOT EXISTS `idx_ref_table_id` (`REF_TABLE_ID`);

-- sys_table_cmd 表 - 按表ID查询命令
ALTER TABLE `sys_table_cmd` ADD INDEX IF NOT EXISTS `idx_sys_table_id` (`SYS_TABLE_ID`);

-- sys_group_prem 表 - 按权限组或目录查询
ALTER TABLE `sys_group_prem` ADD INDEX IF NOT EXISTS `idx_groups_id` (`SYS_GROUPS_ID`);
ALTER TABLE `sys_group_prem` ADD INDEX IF NOT EXISTS `idx_directory_id` (`SYS_DIRECTORY_ID`);

-- sys_user_groups 表 - 按用户ID查询
ALTER TABLE `sys_user_groups` ADD INDEX IF NOT EXISTS `idx_user_id` (`SYS_USER_ID`);
ALTER TABLE `sys_user_groups` ADD INDEX IF NOT EXISTS `idx_directory_id` (`SYS_DIRECTORY_ID`);

-- sys_action 表 - 按表ID查询动作
ALTER TABLE `sys_action` ADD INDEX IF NOT EXISTS `idx_sys_table_id` (`SYS_TABLE_ID`);

-- sys_dict_item 表 - 按字典ID查询
ALTER TABLE `sys_dict_item` ADD INDEX IF NOT EXISTS `idx_dict_id` (`SYS_DICT_ID`);

-- sys_directory 表 - 按父目录ID查询子目录
ALTER TABLE `sys_directory` ADD INDEX IF NOT EXISTS `idx_parent_id` (`PARENT_ID`);
ALTER TABLE `sys_directory` ADD INDEX IF NOT EXISTS `idx_sys_table_id` (`SYS_TABLE_ID`);

-- sys_user_session 表优化 - 按用户和状态查询
ALTER TABLE `sys_user_session` ADD INDEX IF NOT EXISTS `idx_status_expire` (`STATUS`, `EXPIRES_AT`);

-- sys_file 表优化 - 按业务类型查询
ALTER TABLE `sys_file` ADD INDEX IF NOT EXISTS `idx_biz_type_id` (`BIZ_TYPE`, `BIZ_ID`);

-- sys_message 表优化 - 按接收人查询
ALTER TABLE `sys_message` ADD INDEX IF NOT EXISTS `idx_recipient` (`RECIPIENT_ID`, `IS_READ`);

SELECT '索引优化完成！' AS message;