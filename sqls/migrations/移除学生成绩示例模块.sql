-- 移除测试用的「学生成绩」示例模块：数据表、表单、字段、表关联、专用数据字典与菜单分类。
-- 按名称匹配，并限定在「学生成绩」分类下；数据字典只在没有其他字段使用时删除。可重复执行
SET NAMES utf8mb4;

-- 待删除的表单（限定分类，避免误删其他分类中同名的表单）
DROP TEMPORARY TABLE IF EXISTS tmp_demo_tables;
CREATE TEMPORARY TABLE tmp_demo_tables (ID int unsigned PRIMARY KEY);
INSERT INTO tmp_demo_tables (ID)
SELECT t.ID FROM `sys_table` t JOIN `sys_table_category` c ON c.ID = t.SYS_TABLECATEGORY_ID
WHERE c.NAME = '学生成绩' AND t.NAME IN ('TEACHER', 'STU_CLASS', 'STUDENT', 'COURSE', 'SCORE');

-- 表单的权限、动作、命令钩子、SQL、流程（本地均为空，其他环境若有一并删除）
DELETE gp FROM `sys_group_prem` gp JOIN `sys_directory` d ON d.ID = gp.SYS_DIRECTORY_ID WHERE d.SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `sys_directory` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `sys_action` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `sys_table_cmd` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `sys_table_sql` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `wf_instance` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `wf_definition` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);

-- 表关联、字段、表单
-- MySQL 同一语句中不能两次引用临时表，按两个方向分别删除
DELETE FROM `sys_table_ref` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `sys_table_ref` WHERE REF_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DROP TEMPORARY TABLE IF EXISTS tmp_demo_dicts;
CREATE TEMPORARY TABLE tmp_demo_dicts (ID int unsigned PRIMARY KEY);
INSERT IGNORE INTO tmp_demo_dicts (ID)
SELECT DISTINCT SYS_DICT_ID FROM `sys_column` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables) AND SYS_DICT_ID > 0;
DELETE FROM `sys_column` WHERE SYS_TABLE_ID IN (SELECT ID FROM tmp_demo_tables);
DELETE FROM `sys_table` WHERE ID IN (SELECT ID FROM tmp_demo_tables);

-- 只被这些字段使用的数据字典（性别、科目、考试类型）
DELETE FROM tmp_demo_dicts WHERE ID IN (SELECT SYS_DICT_ID FROM `sys_column` WHERE SYS_DICT_ID > 0);
DELETE FROM `sys_dict_item` WHERE SYS_DICT_ID IN (SELECT ID FROM tmp_demo_dicts);
DELETE FROM `sys_dict` WHERE ID IN (SELECT ID FROM tmp_demo_dicts);

-- 已没有表单的「学生成绩」菜单分类
DELETE c FROM `sys_table_category` c
WHERE c.NAME = '学生成绩' AND NOT EXISTS (SELECT 1 FROM `sys_table` t WHERE t.SYS_TABLECATEGORY_ID = c.ID);

DROP TEMPORARY TABLE IF EXISTS tmp_demo_tables;
DROP TEMPORARY TABLE IF EXISTS tmp_demo_dicts;

-- 数据表（先删引用其他表的成绩、学生）
DROP TABLE IF EXISTS `score`;
DROP TABLE IF EXISTS `student`;
DROP TABLE IF EXISTS `course`;
DROP TABLE IF EXISTS `stu_class`;
DROP TABLE IF EXISTS `teacher`;
