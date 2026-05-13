-- ============================================================
-- 数据字典补全
-- 日期: 2026-05-13
-- ============================================================

-- ============================================================
-- Phase 1: 新建业务字典 (sys_dict)
-- ============================================================

INSERT IGNORE INTO sys_dict (ID, NAME, DISPLAY_NAME, IS_ACTIVE) VALUES
(14, 'cloud_item_type', '云盘项目类型', 'Y'),
(15, 'storage_type', '存储类型', 'Y'),
(16, 'quota_type', '配额类型', 'Y'),
(17, 'share_type', '分享类型', 'Y'),
(18, 'share_status', '分享状态', 'Y'),
(19, 'upload_status', '上传状态', 'Y'),
(20, 'pull_task_status', '拉流任务状态', 'Y'),
(21, 'message_type', '消息类型', 'Y'),
(22, 'message_status', '消息状态', 'Y'),
(23, 'callback_event_type', '回调事件类型', 'Y'),
(24, 'priority_level', '优先级', 'Y'),
(25, 'action_status', '操作状态', 'Y'),
(26, 'notify_status', '通知状态', 'Y'),
(27, 'language_code', '语言代码', 'Y');

-- ============================================================
-- Phase 2: 填充字典项 (sys_dict_item)
-- ============================================================

-- 14: cloud_item_type
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(14, 'file', '文件', 1, 'Y'),
(14, 'folder', '文件夹', 2, 'Y');

-- 15: storage_type
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(15, 'local', '本地存储', 1, 'Y'),
(15, 'aliyunOSS', '阿里云OSS', 2, 'Y'),
(15, 'tencentCOS', '腾讯云COS', 3, 'Y');

-- 16: quota_type
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(16, 'standard', '标准版', 1, 'Y'),
(16, 'premium', '高级版', 2, 'Y');

-- 17: share_type
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(17, 'public', '公开分享', 1, 'Y'),
(17, 'password', '加密分享', 2, 'Y'),
(17, 'private', '私有分享', 3, 'Y');

-- 18: share_status
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(18, 'active', '有效', 1, 'Y'),
(18, 'expired', '已过期', 2, 'Y'),
(18, 'disabled', '已禁用', 3, 'Y');

-- 19: upload_status
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(19, 'uploading', '上传中', 1, 'Y'),
(19, 'paused', '已暂停', 2, 'Y'),
(19, 'completed', '已完成', 3, 'Y'),
(19, 'failed', '上传失败', 4, 'Y');

-- 20: pull_task_status
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(20, 'enable', '启用', 1, 'Y'),
(20, 'pause', '暂停', 2, 'Y'),
(20, 'running', '运行中', 3, 'Y'),
(20, 'stopped', '已停止', 4, 'Y');

-- 21: message_type
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(21, 'system', '系统消息', 1, 'Y'),
(21, 'notify', '通知消息', 2, 'Y'),
(21, 'warning', '警告消息', 3, 'Y'),
(21, 'error', '错误消息', 4, 'Y');

-- 22: message_status
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(22, 'active', '有效', 1, 'Y'),
(22, 'expired', '已过期', 2, 'Y'),
(22, 'deleted', '已删除', 3, 'Y');

-- 23: callback_event_type
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(23, 'push_stream', '推流事件', 1, 'Y'),
(23, 'disconnect_stream', '断流事件', 2, 'Y'),
(23, 'recording_file', '录制文件', 3, 'Y'),
(23, 'recording_status', '录制状态', 4, 'Y');

-- 24: priority_level
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(24, 'low', '普通', 1, 'Y'),
(24, 'medium', '重要', 2, 'Y'),
(24, 'high', '紧急', 3, 'Y');

-- 25: action_status
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(25, 'success', '成功', 1, 'Y'),
(25, 'failure', '失败', 2, 'Y');

-- 26: notify_status
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(26, 'pending', '待发送', 1, 'Y'),
(26, 'sent', '已发送', 2, 'Y'),
(26, 'failed', '发送失败', 3, 'Y'),
(26, 'read', '已读', 4, 'Y');

-- 27: language_code
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_DEFAULT_VALUE, IS_ACTIVE) VALUES
(27, 'zh-CN', '简体中文', 1, 'Y', 'Y'),
(27, 'en-US', 'English', 2, 'N', 'Y');

-- ============================================================
-- Phase 3: 更新 sys_column 的 SYS_DICT_ID 引用
-- ============================================================

-- cloud_item.ITEM_TYPE → dict 14
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '14'
WHERE c.DB_NAME = 'ITEM_TYPE' AND t.NAME = 'cloud_item' AND (c.SYS_DICT_ID IS NULL OR c.SYS_DICT_ID = '');

-- cloud_item.STORAGE_TYPE → dict 15
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '15'
WHERE c.DB_NAME = 'STORAGE_TYPE' AND t.NAME = 'cloud_item';

-- cloud_share.SHARE_TYPE → dict 17
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '17'
WHERE c.DB_NAME = 'SHARE_TYPE' AND t.NAME = 'cloud_share';

-- cloud_share.STATUS → dict 18
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '18'
WHERE c.DB_NAME = 'STATUS' AND t.NAME = 'cloud_share';

-- cloud_quota.QUOTA_TYPE → dict 16
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '16'
WHERE c.DB_NAME = 'QUOTA_TYPE' AND t.NAME = 'cloud_quota';

-- cloud_upload_session.STATUS → dict 19
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '19'
WHERE c.DB_NAME = 'STATUS' AND t.NAME = 'cloud_upload_session';

-- pull_stream_task.STATUS → dict 20
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '20'
WHERE c.DB_NAME = 'STATUS' AND t.NAME = 'pull_stream_task';

-- sys_message.MESSAGE_TYPE → dict 21
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '21'
WHERE c.DB_NAME = 'MESSAGE_TYPE' AND t.NAME = 'sys_message';

-- sys_message.STATUS → dict 22
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '22'
WHERE c.DB_NAME = 'STATUS' AND t.NAME = 'sys_message';

-- sys_message.PRIORITY → dict 24
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '24'
WHERE c.DB_NAME = 'PRIORITY' AND t.NAME = 'sys_message';

-- live_callback_event.EVENT_TYPE → dict 23
UPDATE sys_column c JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
SET c.SYS_DICT_ID = '23'
WHERE c.DB_NAME IN ('EVENT_TYPE', 'TYPE') AND t.NAME = 'live_callback_event';

-- sync_column.COL_TYPE → needs dict (varchar/int/decimal/date/datetime/char)
INSERT IGNORE INTO sys_dict (ID, NAME, DISPLAY_NAME, IS_ACTIVE) VALUES
(28, 'col_type', '字段类型', 'Y');

INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(28, 'varchar', 'varchar', 1, 'Y'),
(28, 'int', 'int', 2, 'Y'),
(28, 'bigint', 'bigint', 3, 'Y'),
(28, 'decimal', 'decimal', 4, 'Y'),
(28, 'date', 'date', 5, 'Y'),
(28, 'datetime', 'datetime', 6, 'Y'),
(28, 'char', 'char', 7, 'Y'),
(28, 'text', 'text', 8, 'Y');

-- sys_column.COL_TYPE → dict 28
UPDATE sys_column SET SYS_DICT_ID = '28' WHERE DB_NAME = 'COL_TYPE' AND (SYS_DICT_ID IS NULL OR SYS_DICT_ID = '');

-- sync_column.DISPLAY_TYPE → needs dict
INSERT IGNORE INTO sys_dict (ID, NAME, DISPLAY_NAME, IS_ACTIVE) VALUES
(29, 'display_type', '显示控件类型', 'Y');

INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(29, 'text', '单行文本', 1, 'Y'),
(29, 'textarea', '多行文本', 2, 'Y'),
(29, 'number', '数字', 3, 'Y'),
(29, 'select', '下拉选择', 4, 'Y'),
(29, 'date', '日期', 5, 'Y'),
(29, 'datetime', '日期时间', 6, 'Y'),
(29, 'file', '文件上传', 7, 'Y'),
(29, 'image', '图片上传', 8, 'Y'),
(29, 'checkbox', '多选', 9, 'Y'),
(29, 'radio', '单选', 10, 'Y'),
(29, 'json', 'JSON编辑器', 11, 'Y'),
(29, 'clob', '富文本', 12, 'Y'),
(29, 'blank', '空白占位', 13, 'Y'),
(29, 'button', '按钮', 14, 'Y'),
(29, 'hr', '分割线', 15, 'Y');

-- sys_column.DISPLAY_TYPE → dict 29
UPDATE sys_column SET SYS_DICT_ID = '29' WHERE DB_NAME = 'DISPLAY_TYPE' AND (SYS_DICT_ID IS NULL OR SYS_DICT_ID = '');

-- ============================================================
-- Phase 4: 修复前端 LiveRoomList 的 viewing_method 映射不一致
-- ============================================================
-- 当前 dict 10 items: public, encrypted, paid, ticket, enterprise, custom
-- 前端使用: public, password, pay (错误的键名)
-- 修复: 为 dict 10 添加前端兼容的别名项

INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(10, 'password', '加密(兼容)', 99, 'Y'),
(10, 'pay', '付费(兼容)', 99, 'Y');

-- ============================================================
-- Phase 5: 清理测试数据
-- ============================================================
UPDATE sys_dict SET IS_ACTIVE = 'N' WHERE NAME IN ('abc', 'asdf');
UPDATE sys_dict_item SET IS_ACTIVE = 'N' WHERE SYS_DICT_ID IN (4, 5);

-- ============================================================
-- Phase 6: 修复 SetValueType 列引用
-- ============================================================
INSERT IGNORE INTO sys_dict (ID, NAME, DISPLAY_NAME, IS_ACTIVE) VALUES
(30, 'set_value_type', '赋值策略', 'Y');

INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(30, 'pk', '主键', 1, 'Y'),
(30, 'docno', '单据编号', 2, 'Y'),
(30, 'fk', '外键关联', 3, 'Y'),
(30, 'byPage', '页面输入', 4, 'Y'),
(30, 'select', '字典选择', 5, 'Y'),
(30, 'createBy', '创建人', 6, 'Y'),
(30, 'sysdate', '系统时间', 7, 'Y'),
(30, 'operator', '操作人', 8, 'Y'),
(30, 'password', '密码', 9, 'Y'),
(30, 'ignore', '忽略', 10, 'Y');

-- ============================================================
-- Phase 7: 补全 display_type 字典缺失项 (dict 29)
-- ============================================================
INSERT IGNORE INTO sys_dict_item (SYS_DICT_ID, VALUE, DISPLAY_NAME, ORDERNO, IS_ACTIVE) VALUES
(29, 'number', '数字输入', 3, 'Y'),
(29, 'time', '时间', 6, 'Y'),
(29, 'switch', '开关', 9, 'Y'),
(29, 'color', '颜色选择', 14, 'Y'),
(29, 'fk', '外键选择', 3, 'Y'),
(29, 'richtext', '富文本', 12, 'Y'),
(29, 'xml', 'XML编辑器', 13, 'Y'),
(29, 'permissionbit', '权限位', 14, 'Y');

-- ============================================================
-- Phase 8: 更新 sys_column 的 COL_TYPE SYS_DICT_ID 引用
-- ============================================================
UPDATE sys_column SET SYS_DICT_ID = '28' WHERE DB_NAME = 'COL_TYPE' AND SYS_DICT_ID IS NULL;
UPDATE sys_column SET SYS_DICT_ID = '29' WHERE DB_NAME = 'DISPLAY_TYPE' AND SYS_DICT_ID IS NULL;
UPDATE sys_column SET SYS_DICT_ID = '30' WHERE DB_NAME = 'SET_VALUE_TYPE' AND SYS_DICT_ID IS NULL AND SET_VALUE_TYPE IS NOT NULL;
