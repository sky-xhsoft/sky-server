-- 修正自动任务相关字段的 MASK：原值为表级格式 "AMQ"，按字段 10 位格式解析后全为 0，
-- 导致字段在列表/表单中不显示、也无法编辑（页面无法停用自动任务）。
-- 字段 MASK 位序：新增可见、新增可编辑、编辑可见、编辑可编辑、列表可见、列表可编辑、导入、导出、打印、保留
UPDATE sys_column SET MASK = '1111111111' WHERE SYS_TABLE_ID = 83 AND DB_NAME = 'IS_ACTIVE' AND MASK = 'AMQ';
-- 上次/下次运行时间由调度器计算：只读显示
UPDATE sys_column SET MASK = '1010100110' WHERE SYS_TABLE_ID = 83 AND DB_NAME IN ('LAST_RUN_AT', 'NEXT_RUN_AT') AND MASK = 'AMQ';
-- 执行日志状态：只读显示
UPDATE sys_column SET MASK = '1010100110' WHERE SYS_TABLE_ID = 84 AND DB_NAME = 'STATUS' AND MASK = 'AMQ';
