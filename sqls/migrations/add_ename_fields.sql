-- 添加 ENAME 字段到相关表
-- sys_table: 英文显示名称
-- sys_column: 英文显示名称
-- sys_subsystem: 英文显示名称
-- sys_table_category: 英文显示名称

-- 为 sys_table 添加 ENAME 字段
ALTER TABLE `sys_table`
ADD COLUMN `ENAME` VARCHAR(255) NULL COMMENT '英文显示名称' AFTER `DISPLAY_NAME`;

-- 为 sys_column 添加 ENAME 字段
ALTER TABLE `sys_column`
ADD COLUMN `ENAME` VARCHAR(255) NULL COMMENT '英文显示名称' AFTER `DISPLAY_NAME`;

-- 为 sys_subsystem 添加 ENAME 字段
ALTER TABLE `sys_subsystem`
ADD COLUMN `ENAME` VARCHAR(255) NULL COMMENT '英文显示名称' AFTER `NAME`;

-- 为 sys_table_category 添加 ENAME 字段
ALTER TABLE `sys_table_category`
ADD COLUMN `ENAME` VARCHAR(255) NULL COMMENT '英文显示名称' AFTER `NAME`;
