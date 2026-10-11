-- 菜单：直播中心 →「同传字幕」（权限目录与社媒分发相同，与同传接口的表权限一致）
SET NAMES utf8mb4;

INSERT INTO `sys_table` (
  `NAME`, `DISPLAY_NAME`, `ENAME`, `MASK`, `SYS_TABLECATEGORY_ID`, `ORDERNO`, `URL`, `IS_MENU`, `ICO_IMG`,
  `SYS_DIRECTORY_ID`, `IS_BIG`, `DESCRIPTION`,
  `SYS_COMPANY_ID`, `CREATE_BY`, `CREATE_TIME`, `UPDATE_BY`, `UPDATE_TIME`, `IS_ACTIVE`
)
SELECT 'LIVE_SPEECH', '同传字幕', 'Live Interpretation', 'AMDQ', d.SYS_TABLECATEGORY_ID, 45, '/live/speech-tasks', 'Y', 'icon-language',
       d.SYS_DIRECTORY_ID, 'N', '声影同传、实时同传与原文字幕任务（火山引擎）',
       0, 'system', NOW(), 'system', NOW(), 'Y'
FROM `sys_table` d
WHERE d.NAME = 'LIVE_PULL'
  AND NOT EXISTS (SELECT 1 FROM `sys_table` t WHERE t.NAME = 'LIVE_SPEECH');
