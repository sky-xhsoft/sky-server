-- 清空并重建 sys_directory，与 sys_table 1:1 对应（只含 IS_MENU='Y' 的表）
DELETE FROM sys_directory;

-- 为每个 IS_MENU='Y' 的表创建一条目录记录
INSERT INTO sys_directory (NAME, DISPLAY_NAME, SYS_TABLE_ID, SYS_TABLE_CATEGORY_ID, IS_ACTIVE)
SELECT 
  t.NAME,
  t.DISPLAY_NAME,
  t.ID,
  t.SYS_TABLECATEGORY_ID,
  'Y'
FROM sys_table t
WHERE t.IS_ACTIVE = 'Y' 
  AND t.IS_MENU = 'Y'
  AND t.SYS_TABLECATEGORY_ID IS NOT NULL;
