-- 更新 sys_subsystem 的英文名称
UPDATE sys_subsystem SET EName = 'System Management' WHERE ID = 1;
UPDATE sys_subsystem SET EName = 'Video Live' WHERE ID = 7;
UPDATE sys_subsystem SET EName = 'Cloud Storage' WHERE ID = 10;

-- 更新 sys_table_category 的英文名称
UPDATE sys_table_category SET EName = 'Metadata' WHERE ID = 1;
UPDATE sys_table_category SET EName = 'Live Center' WHERE ID = 2;
UPDATE sys_table_category SET EName = 'User Rights' WHERE ID = 3;
UPDATE sys_table_category SET EName = 'Cloud Disk' WHERE ID = 4;

-- 更新 sys_table 的英文名称
UPDATE sys_table SET EName = 'System Columns' WHERE ID = 3;
UPDATE sys_table SET EName = 'Company' WHERE ID = 4;
UPDATE sys_table SET EName = 'Data Dictionary' WHERE ID = 5;
UPDATE sys_table SET EName = 'Permission Groups' WHERE ID = 11;
UPDATE sys_table SET EName = 'Sequence Generator' WHERE ID = 18;
UPDATE sys_table SET EName = 'Subsystem' WHERE ID = 19;
UPDATE sys_table SET EName = 'System Form' WHERE ID = 20;
UPDATE sys_table SET EName = 'Table Category' WHERE ID = 21;
UPDATE sys_table SET EName = 'System User' WHERE ID = 25;
UPDATE sys_table SET EName = 'Live Domain' WHERE ID = 35;
UPDATE sys_table SET EName = 'Cloud Item' WHERE ID = 37;
UPDATE sys_table SET EName = 'Live Stream' WHERE ID = 38;
UPDATE sys_table SET EName = 'Social Distribution' WHERE ID = 39;
UPDATE sys_table SET EName = 'Live Recording' WHERE ID = 41;
UPDATE sys_table SET EName = 'Live Room' WHERE ID = 54;

-- 查询验证
SELECT ID, NAME, EName FROM sys_subsystem;
SELECT ID, NAME, EName FROM sys_table_category;
SELECT ID, NAME, DISPLAY_NAME, EName FROM sys_table WHERE IS_MENU = 'Y';
