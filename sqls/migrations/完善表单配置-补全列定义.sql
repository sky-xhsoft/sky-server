-- ============================================================
-- 表单配置完善 — 为所有缺失列的表补充 sys_column 元数据
-- 日期: 2026-05-13
-- ============================================================

-- ============================================================
-- cloud_folder 列定义 (17 columns)
-- ============================================================
SET @cf_id = (SELECT ID FROM sys_table WHERE NAME='cloud_folder');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@cf_id, 'ID', 'ID', 'ID', 'text', 'bigint', 20, 'N', 1, '0000000000', 'pk', 'Y'),
(@cf_id, 'NAME', '文件夹名称', '文件夹名称', 'text', 'varchar', 255, 'N', 8, '1111111111', 'byPage', 'Y'),
(@cf_id, 'PARENT_ID', '父文件夹ID', '父文件夹ID', 'text', 'bigint', 20, 'Y', 9, '1111111111', 'byPage', 'Y'),
(@cf_id, 'PATH', '路径', '路径', 'text', 'varchar', 1000, 'N', 10, '1100000000', 'byPage', 'Y'),
(@cf_id, 'OWNER_ID', '所有者', '所有者', 'fk', 'bigint', 20, 'N', 11, '1111111111', 'fk', 'Y'),
(@cf_id, 'IS_PUBLIC', '是否公开', '是否公开', 'select', 'char', 1, 'Y', 12, '1111111111', 'select', 'Y'),
(@cf_id, 'SHARE_CODE', '分享码', '分享码', 'text', 'varchar', 50, 'Y', 13, '1100000000', 'byPage', 'Y'),
(@cf_id, 'SHARE_EXPIRE', '分享过期时间', '分享过期时间', 'datetime', 'datetime', 0, 'Y', 14, '1111111111', 'byPage', 'Y'),
(@cf_id, 'DESCRIPTION', '描述', '描述', 'textarea', 'varchar', 500, 'Y', 15, '1111111111', 'byPage', 'Y'),
(@cf_id, 'FILE_COUNT', '文件数量', '文件数量', 'text', 'int', 11, 'Y', 16, '1000000000', 'byPage', 'Y'),
(@cf_id, 'TOTAL_SIZE', '总大小', '总大小', 'text', 'bigint', 20, 'Y', 17, '1000000000', 'byPage', 'Y'),
(@cf_id, 'SYS_COMPANY_ID', '公司ID', '公司ID', 'text', 'bigint', 20, 'Y', 18, '0000000000', 'object', 'Y'),
(@cf_id, 'CREATE_BY', '创建人', '创建人', 'text', 'varchar', 80, 'Y', 19, '0000000000', 'createBy', 'Y'),
(@cf_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'datetime', 0, 'Y', 20, '0000000000', 'sysdate', 'Y'),
(@cf_id, 'UPDATE_BY', '修改人', '修改人', 'text', 'varchar', 80, 'Y', 21, '0000000000', 'operator', 'Y'),
(@cf_id, 'UPDATE_TIME', '修改时间', '修改时间', 'datetime', 'datetime', 0, 'Y', 22, '0000000000', 'sysdate', 'Y'),
(@cf_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'Y', 23, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- cloud_file 列定义 (25 columns)
-- ============================================================
SET @cfile_id = (SELECT ID FROM sys_table WHERE NAME='cloud_file');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@cfile_id, 'ID', 'ID', 'ID', 'text', 'bigint', 20, 'N', 1, '0000000000', 'pk', 'Y'),
(@cfile_id, 'FILE_NAME', '文件名', '文件名', 'text', 'varchar', 255, 'N', 8, '1111111111', 'byPage', 'Y'),
(@cfile_id, 'FOLDER_ID', '文件夹ID', '文件夹ID', 'fk', 'bigint', 20, 'Y', 9, '1111111111', 'fk', 'Y'),
(@cfile_id, 'PATH', '路径', '路径', 'text', 'varchar', 1000, 'N', 10, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'STORAGE_TYPE', '存储类型', '存储类型', 'select', 'varchar', 20, 'N', 11, '1111111111', 'select', 'Y'),
(@cfile_id, 'STORAGE_PATH', '存储路径', '存储路径', 'text', 'varchar', 500, 'N', 12, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'FILE_SIZE', '文件大小', '文件大小', 'text', 'bigint', 20, 'N', 13, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'FILE_TYPE', '文件类型', '文件类型', 'text', 'varchar', 100, 'Y', 14, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'FILE_EXT', '扩展名', '扩展名', 'text', 'varchar', 20, 'Y', 15, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'MD5', 'MD5值', 'MD5值', 'text', 'varchar', 32, 'Y', 16, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'OWNER_ID', '所有者', '所有者', 'fk', 'bigint', 20, 'N', 17, '1111111111', 'fk', 'Y'),
(@cfile_id, 'IS_PUBLIC', '是否公开', '是否公开', 'select', 'char', 1, 'Y', 18, '1111111111', 'select', 'Y'),
(@cfile_id, 'SHARE_CODE', '分享码', '分享码', 'text', 'varchar', 50, 'Y', 19, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'SHARE_EXPIRE', '分享过期时间', '分享过期时间', 'datetime', 'datetime', 0, 'Y', 20, '1111111111', 'byPage', 'Y'),
(@cfile_id, 'ACCESS_URL', '访问URL', '访问URL', 'text', 'varchar', 500, 'Y', 21, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'THUMBNAIL', '缩略图', '缩略图', 'text', 'varchar', 500, 'Y', 22, '1100000000', 'byPage', 'Y'),
(@cfile_id, 'DOWNLOAD_COUNT', '下载次数', '下载次数', 'text', 'int', 11, 'Y', 23, '1000000000', 'byPage', 'Y'),
(@cfile_id, 'TAGS', '标签', '标签', 'text', 'varchar', 500, 'Y', 24, '1111111111', 'byPage', 'Y'),
(@cfile_id, 'DESCRIPTION', '描述', '描述', 'textarea', 'varchar', 500, 'Y', 25, '1111111111', 'byPage', 'Y'),
(@cfile_id, 'SYS_COMPANY_ID', '公司ID', '公司ID', 'text', 'bigint', 20, 'Y', 26, '0000000000', 'object', 'Y'),
(@cfile_id, 'CREATE_BY', '创建人', '创建人', 'text', 'varchar', 80, 'Y', 27, '0000000000', 'createBy', 'Y'),
(@cfile_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'datetime', 0, 'Y', 28, '0000000000', 'sysdate', 'Y'),
(@cfile_id, 'UPDATE_BY', '修改人', '修改人', 'text', 'varchar', 80, 'Y', 29, '0000000000', 'operator', 'Y'),
(@cfile_id, 'UPDATE_TIME', '修改时间', '修改时间', 'datetime', 'datetime', 0, 'Y', 30, '0000000000', 'sysdate', 'Y'),
(@cfile_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'Y', 31, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- cloud_chunk_record 列定义 (11 columns)
-- ============================================================
SET @ccr_id = (SELECT ID FROM sys_table WHERE NAME='cloud_chunk_record');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@ccr_id, 'ID', 'ID', 'ID', 'text', 'bigint', 20, 'N', 1, '0000000000', 'pk', 'Y'),
(@ccr_id, 'SESSION_ID', '会话ID', '会话ID', 'fk', 'bigint', 20, 'N', 8, '1111111111', 'fk', 'Y'),
(@ccr_id, 'CHUNK_INDEX', '分片序号', '分片序号', 'text', 'int', 11, 'N', 9, '1111111111', 'byPage', 'Y'),
(@ccr_id, 'CHUNK_SIZE', '分片大小', '分片大小', 'text', 'int', 11, 'N', 10, '1111111111', 'byPage', 'Y'),
(@ccr_id, 'CHUNK_MD5', '分片MD5', '分片MD5', 'text', 'varchar', 32, 'N', 11, '1100000000', 'byPage', 'Y'),
(@ccr_id, 'CHUNK_PATH', '分片路径', '分片路径', 'text', 'varchar', 500, 'Y', 12, '1100000000', 'byPage', 'Y'),
(@ccr_id, 'UPLOADED', '是否已上传', '是否已上传', 'select', 'tinyint', 1, 'N', 13, '1111111111', 'select', 'Y'),
(@ccr_id, 'UPLOAD_TIME', '上传时间', '上传时间', 'datetime', 'timestamp', 0, 'Y', 14, '1000000000', 'sysdate', 'Y'),
(@ccr_id, 'RETRY_COUNT', '重试次数', '重试次数', 'text', 'int', 11, 'N', 14, '1000000000', 'byPage', 'Y'),
(@ccr_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'timestamp', 0, 'N', 15, '0000000000', 'sysdate', 'Y'),
(@ccr_id, 'UPDATE_TIME', '更新时间', '更新时间', 'datetime', 'timestamp', 0, 'N', 16, '0000000000', 'sysdate', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- live_callback_event 列定义 (15 columns)
-- ============================================================
SET @lce_id = (SELECT ID FROM sys_table WHERE NAME='live_callback_event');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@lce_id, 'ID', 'ID', 'ID', 'text', 'bigint', 20, 'N', 1, '0000000000', 'pk', 'Y'),
(@lce_id, 'EVENT_TYPE', '事件类型', '事件类型', 'select', 'varchar', 50, 'N', 8, '1111111111', 'select', 'Y'),
(@lce_id, 'EVENT_TIME', '事件时间', '事件时间', 'datetime', 'bigint', 20, 'N', 9, '1111111111', 'byPage', 'Y'),
(@lce_id, 'DOMAIN_NAME', '域名', '域名', 'text', 'varchar', 255, 'Y', 10, '1100000000', 'byPage', 'Y'),
(@lce_id, 'APP_NAME', '应用名', '应用名', 'text', 'varchar', 255, 'Y', 11, '1100000000', 'byPage', 'Y'),
(@lce_id, 'STREAM_NAME', '流名称', '流名称', 'text', 'varchar', 255, 'Y', 12, '1100000000', 'byPage', 'Y'),
(@lce_id, 'STREAM_ID', '流ID', '流ID', 'text', 'varchar', 255, 'Y', 13, '1100000000', 'byPage', 'Y'),
(@lce_id, 'ROOM_NAME', '直播间名称', '直播间名称', 'text', 'varchar', 255, 'Y', 14, '1100000000', 'byPage', 'Y'),
(@lce_id, 'CLIENT_IP', '客户端IP', '客户端IP', 'text', 'varchar', 50, 'Y', 15, '1100000000', 'byPage', 'Y'),
(@lce_id, 'EVENT_DATA', '事件数据', '事件数据', 'json', 'text', 0, 'Y', 16, '1100000000', 'byPage', 'Y'),
(@lce_id, 'SIGN', '签名', '签名', 'text', 'varchar', 255, 'Y', 17, '1000000000', 'byPage', 'Y'),
(@lce_id, 'T_VALUE', '时间戳值', '时间戳值', 'text', 'bigint', 20, 'N', 18, '1000000000', 'byPage', 'Y'),
(@lce_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'datetime', 0, 'N', 19, '0000000000', 'sysdate', 'Y'),
(@lce_id, 'SYS_COMPANY_ID', '公司ID', '公司ID', 'text', 'bigint', 20, 'N', 20, '0000000000', 'object', 'Y'),
(@lce_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'Y', 21, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- live_push_domain_config 列定义 (已废弃删除, 配置改从 live:domains:list 缓存获取推流域名)
-- ============================================================

-- ============================================================
-- cloud_storage_config 列定义 (21 columns)
-- ============================================================
SET @csc_id = (SELECT ID FROM sys_table WHERE NAME='cloud_storage_config');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@csc_id, 'ID', 'ID', 'ID', 'text', 'bigint', 20, 'N', 1, '0000000000', 'pk', 'Y'),
(@csc_id, 'SYS_COMPANY_ID', '公司', '公司', 'fk', 'int', 11, 'N', 8, '1111111111', 'fk', 'Y'),
(@csc_id, 'STORAGE_TYPE', '存储类型', '存储类型', 'select', 'varchar', 20, 'N', 9, '1111111111', 'select', 'Y'),
(@csc_id, 'LOCAL_BASE_PATH', '本地路径', '本地路径', 'text', 'varchar', 500, 'Y', 10, '1111111111', 'byPage', 'Y'),
(@csc_id, 'LOCAL_BASE_URL', '本地URL', '本地URL', 'text', 'varchar', 500, 'Y', 11, '1111111111', 'byPage', 'Y'),
(@csc_id, 'ALIYUN_OSS_ENDPOINT', 'OSS Endpoint', 'OSS Endpoint', 'text', 'varchar', 255, 'Y', 12, '1111111111', 'byPage', 'Y'),
(@csc_id, 'ALIYUN_OSS_ACCESS_KEY_ID', 'OSS Key ID', 'OSS Key ID', 'text', 'varchar', 255, 'Y', 13, '1111111111', 'byPage', 'Y'),
(@csc_id, 'ALIYUN_OSS_ACCESS_KEY_SECRET', 'OSS Key Secret', 'OSS Key Secret', 'text', 'varchar', 255, 'Y', 14, '1111111111', 'password', 'Y'),
(@csc_id, 'ALIYUN_OSS_BUCKET_NAME', 'OSS Bucket', 'OSS Bucket', 'text', 'varchar', 255, 'Y', 15, '1111111111', 'byPage', 'Y'),
(@csc_id, 'ALIYUN_OSS_CDN_DOMAIN', 'OSS CDN域名', 'OSS CDN域名', 'text', 'varchar', 500, 'Y', 16, '1111111111', 'byPage', 'Y'),
(@csc_id, 'TENCENT_COS_BUCKET_URL', 'COS Bucket URL', 'COS Bucket URL', 'text', 'varchar', 500, 'Y', 17, '1111111111', 'byPage', 'Y'),
(@csc_id, 'TENCENT_COS_SECRET_ID', 'COS Secret ID', 'COS Secret ID', 'text', 'varchar', 255, 'Y', 18, '1111111111', 'byPage', 'Y'),
(@csc_id, 'TENCENT_COS_SECRET_KEY', 'COS Secret Key', 'COS Secret Key', 'text', 'varchar', 255, 'Y', 19, '1111111111', 'password', 'Y'),
(@csc_id, 'TENCENT_COS_BUCKET_NAME', 'COS Bucket', 'COS Bucket', 'text', 'varchar', 255, 'Y', 20, '1111111111', 'byPage', 'Y'),
(@csc_id, 'TENCENT_COS_REGION', 'COS 区域', 'COS 区域', 'text', 'varchar', 50, 'Y', 21, '1111111111', 'byPage', 'Y'),
(@csc_id, 'TENCENT_COS_CDN_DOMAIN', 'COS CDN域名', 'COS CDN域名', 'text', 'varchar', 500, 'Y', 22, '1111111111', 'byPage', 'Y'),
(@csc_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'N', 23, '1111111111', 'select', 'Y'),
(@csc_id, 'CREATE_BY', '创建人', '创建人', 'text', 'varchar', 80, 'Y', 24, '0000000000', 'createBy', 'Y'),
(@csc_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'datetime', 0, 'N', 25, '0000000000', 'sysdate', 'Y'),
(@csc_id, 'UPDATE_BY', '修改人', '修改人', 'text', 'varchar', 80, 'Y', 26, '0000000000', 'operator', 'Y'),
(@csc_id, 'UPDATE_TIME', '修改时间', '修改时间', 'datetime', 'datetime', 0, 'N', 27, '0000000000', 'sysdate', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- sys_company_conf 补全列定义 (原有基础列 + migration新增列)
-- ============================================================
SET @scc_id = (SELECT ID FROM sys_table WHERE NAME='sys_company_conf');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
-- 原有基础列 (ORDERNO 8-22)
(@scc_id, 'SECRET_ID', 'Secret ID', 'Secret ID', 'text', 'varchar', 255, 'Y', 8, '1111111111', 'byPage', 'Y'),
(@scc_id, 'SECRET_KEY', 'Secret Key', 'Secret Key', 'text', 'varchar', 255, 'Y', 9, '1111111111', 'password', 'Y'),
(@scc_id, 'REGION', '区域', '区域', 'text', 'varchar', 255, 'Y', 10, '1111111111', 'byPage', 'Y'),
(@scc_id, 'STORAGE_TYPE', '存储类型', '存储类型', 'select', 'varchar', 20, 'Y', 11, '1111111111', 'select', 'Y'),
(@scc_id, 'LOCAL_BASE_PATH', '本地路径', '本地路径', 'text', 'varchar', 500, 'Y', 12, '1111111111', 'byPage', 'Y'),
(@scc_id, 'LOCAL_BASE_URL', '本地URL', '本地URL', 'text', 'varchar', 500, 'Y', 13, '1111111111', 'byPage', 'Y'),
(@scc_id, 'ALIYUN_OSS_ENDPOINT', 'OSS Endpoint', 'OSS Endpoint', 'text', 'varchar', 255, 'Y', 14, '1111111111', 'byPage', 'Y'),
(@scc_id, 'ALIYUN_OSS_ACCESS_KEY_ID', 'OSS Key ID', 'OSS Key ID', 'text', 'varchar', 255, 'Y', 15, '1111111111', 'byPage', 'Y'),
(@scc_id, 'ALIYUN_OSS_ACCESS_KEY_SECRET', 'OSS Key Secret', 'OSS Key Secret', 'text', 'varchar', 255, 'Y', 16, '1111111111', 'password', 'Y'),
(@scc_id, 'ALIYUN_OSS_BUCKET_NAME', 'OSS Bucket', 'OSS Bucket', 'text', 'varchar', 255, 'Y', 17, '1111111111', 'byPage', 'Y'),
(@scc_id, 'ALIYUN_OSS_CDN_DOMAIN', 'OSS CDN域名', 'OSS CDN域名', 'text', 'varchar', 500, 'Y', 18, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_COS_BUCKET_URL', 'COS URL', 'COS URL', 'text', 'varchar', 500, 'Y', 19, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_COS_SECRET_ID', 'COS Secret ID', 'COS Secret ID', 'text', 'varchar', 255, 'Y', 20, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_COS_SECRET_KEY', 'COS Secret Key', 'COS Secret Key', 'text', 'varchar', 255, 'Y', 21, '1111111111', 'password', 'Y'),
(@scc_id, 'TENCENT_COS_BUCKET_NAME', 'COS Bucket', 'COS Bucket', 'text', 'varchar', 255, 'Y', 22, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_COS_REGION', 'COS 区域', 'COS 区域', 'text', 'varchar', 50, 'Y', 23, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_COS_CDN_DOMAIN', 'COS CDN域名', 'COS CDN域名', 'text', 'varchar', 500, 'Y', 24, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_CLOUD_SECRET_ID', '腾讯云Secret ID', '腾讯云Secret ID', 'text', 'varchar', 255, 'Y', 25, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_CLOUD_SECRET_KEY', '腾讯云Secret Key', '腾讯云Secret Key', 'text', 'varchar', 255, 'Y', 26, '1111111111', 'password', 'Y'),
(@scc_id, 'TENCENT_CLOUD_REGION', '腾讯云区域', '腾讯云区域', 'text', 'varchar', 255, 'Y', 27, '1111111111', 'byPage', 'Y'),
(@scc_id, 'TENCENT_CLOUD_CALLBACK_KEY', '回调密钥', '回调密钥', 'text', 'varchar', 255, 'Y', 28, '1111111111', 'password', 'Y'),
-- migration 新增列
(@scc_id, 'DEFAULT_LANGUAGE', '默认语言', '默认语言', 'select', 'varchar', 50, 'N', 29, '1111111111', 'select', 'Y'),
(@scc_id, 'TIMEZONE', '时区', '时区', 'text', 'varchar', 50, 'N', 30, '1111111111', 'byPage', 'Y'),
(@scc_id, 'CURRENCY', '货币', '货币', 'text', 'varchar', 10, 'N', 31, '1111111111', 'byPage', 'Y'),
(@scc_id, 'MAX_STORAGE_SIZE', '最大存储', '最大存储', 'text', 'bigint', 20, 'N', 32, '1111111111', 'byPage', 'Y'),
(@scc_id, 'MAX_FILE_SIZE', '最大文件', '最大文件', 'text', 'bigint', 20, 'N', 33, '1111111111', 'byPage', 'Y'),
(@scc_id, 'MAX_FILES', '最大文件数', '最大文件数', 'text', 'int', 11, 'N', 34, '1111111111', 'byPage', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- sys_user_session 列定义 (14 columns)
-- ============================================================
SET @sus_id = (SELECT ID FROM sys_table WHERE NAME='sys_user_session');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@sus_id, 'ID', 'ID', 'ID', 'text', 'int', 11, 'N', 1, '0000000000', 'pk', 'Y'),
(@sus_id, 'USER_ID', '用户', '用户', 'fk', 'int', 11, 'N', 8, '1111111111', 'fk', 'Y'),
(@sus_id, 'COMPANY_ID', '公司', '公司', 'fk', 'int', 11, 'N', 9, '1111111111', 'fk', 'Y'),
(@sus_id, 'TOKEN', '访问令牌', '访问令牌', 'text', 'varchar', 500, 'Y', 10, '1000000000', 'byPage', 'Y'),
(@sus_id, 'REFRESH_TOKEN', '刷新令牌', '刷新令牌', 'text', 'varchar', 500, 'Y', 11, '1000000000', 'byPage', 'Y'),
(@sus_id, 'CLIENT_TYPE', '客户端类型', '客户端类型', 'select', 'varchar', 20, 'Y', 12, '1111111111', 'select', 'Y'),
(@sus_id, 'DEVICE_ID', '设备ID', '设备ID', 'text', 'varchar', 500, 'Y', 13, '1100000000', 'byPage', 'Y'),
(@sus_id, 'DEVICE_NAME', '设备名称', '设备名称', 'text', 'varchar', 255, 'Y', 14, '1100000000', 'byPage', 'Y'),
(@sus_id, 'IP_ADDRESS', 'IP地址', 'IP地址', 'text', 'varchar', 50, 'Y', 15, '1100000000', 'byPage', 'Y'),
(@sus_id, 'USER_AGENT', 'User Agent', 'User Agent', 'text', 'varchar', 500, 'Y', 16, '1100000000', 'byPage', 'Y'),
(@sus_id, 'LOGIN_TIME', '登录时间', '登录时间', 'datetime', 'datetime', 0, 'N', 17, '1100000000', 'sysdate', 'Y'),
(@sus_id, 'LAST_ACTIVE_TIME', '最后活跃', '最后活跃', 'datetime', 'datetime', 0, 'Y', 18, '1000000000', 'sysdate', 'Y'),
(@sus_id, 'EXPIRE_TIME', '过期时间', '过期时间', 'datetime', 'datetime', 0, 'Y', 19, '1000000000', 'byPage', 'Y'),
(@sus_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'Y', 20, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- sys_file 列定义 (23 columns)
-- ============================================================
SET @sf_id = (SELECT ID FROM sys_table WHERE NAME='sys_file');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@sf_id, 'ID', 'ID', 'ID', 'text', 'int', 11, 'N', 1, '0000000000', 'pk', 'Y'),
(@sf_id, 'FILE_NAME', '文件名', '文件名', 'text', 'varchar', 255, 'N', 8, '1111111111', 'byPage', 'Y'),
(@sf_id, 'STORAGE_NAME', '存储名', '存储名', 'text', 'varchar', 255, 'N', 9, '1111111111', 'byPage', 'Y'),
(@sf_id, 'FILE_PATH', '文件路径', '文件路径', 'text', 'varchar', 500, 'N', 10, '1100000000', 'byPage', 'Y'),
(@sf_id, 'FILE_SIZE', '文件大小', '文件大小', 'text', 'bigint', 20, 'N', 11, '1100000000', 'byPage', 'Y'),
(@sf_id, 'FILE_TYPE', '文件类型', '文件类型', 'text', 'varchar', 100, 'Y', 12, '1100000000', 'byPage', 'Y'),
(@sf_id, 'FILE_EXT', '扩展名', '扩展名', 'text', 'varchar', 20, 'Y', 13, '1100000000', 'byPage', 'Y'),
(@sf_id, 'STORAGE_TYPE', '存储类型', '存储类型', 'select', 'varchar', 20, 'N', 14, '1111111111', 'select', 'Y'),
(@sf_id, 'BUCKET_NAME', 'Bucket名称', 'Bucket名称', 'text', 'varchar', 100, 'Y', 15, '1111111111', 'byPage', 'Y'),
(@sf_id, 'ACCESS_URL', '访问URL', '访问URL', 'text', 'varchar', 500, 'Y', 16, '1100000000', 'byPage', 'Y'),
(@sf_id, 'THUMBNAIL_URL', '缩略图URL', '缩略图URL', 'text', 'varchar', 500, 'Y', 17, '1100000000', 'byPage', 'Y'),
(@sf_id, 'MD5', 'MD5值', 'MD5值', 'text', 'varchar', 32, 'Y', 18, '1100000000', 'byPage', 'Y'),
(@sf_id, 'UPLOAD_IP', '上传者IP', '上传者IP', 'text', 'varchar', 50, 'Y', 19, '1000000000', 'byPage', 'Y'),
(@sf_id, 'DOWNLOAD_COUNT', '下载次数', '下载次数', 'text', 'int', 11, 'N', 20, '1000000000', 'byPage', 'Y'),
(@sf_id, 'CATEGORY', '分类', '分类', 'text', 'varchar', 50, 'Y', 21, '1111111111', 'byPage', 'Y'),
(@sf_id, 'DESCRIPTION', '描述', '描述', 'textarea', 'varchar', 500, 'Y', 22, '1111111111', 'byPage', 'Y'),
(@sf_id, 'EXPIRE_TIME', '过期时间', '过期时间', 'datetime', 'datetime', 0, 'Y', 23, '1111111111', 'byPage', 'Y'),
(@sf_id, 'SYS_COMPANY_ID', '公司ID', '公司ID', 'text', 'int', 11, 'Y', 24, '0000000000', 'object', 'Y'),
(@sf_id, 'CREATE_BY', '创建人', '创建人', 'text', 'varchar', 80, 'Y', 25, '0000000000', 'createBy', 'Y'),
(@sf_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'datetime', 0, 'N', 26, '0000000000', 'sysdate', 'Y'),
(@sf_id, 'UPDATE_BY', '修改人', '修改人', 'text', 'varchar', 80, 'Y', 27, '0000000000', 'operator', 'Y'),
(@sf_id, 'UPDATE_TIME', '修改时间', '修改时间', 'datetime', 'datetime', 0, 'N', 28, '0000000000', 'sysdate', 'Y'),
(@sf_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'N', 29, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);

-- ============================================================
-- 最后: 更新 pull_stream_task 的列定义 (已有部分列，补全剩余)
-- ============================================================
SET @pst_id = (SELECT ID FROM sys_table WHERE NAME='pull_stream_task');
INSERT IGNORE INTO sys_column (SYS_TABLE_ID, DB_NAME, FULL_NAME, DISPLAY_NAME, DISPLAY_TYPE, COL_TYPE, COL_LENGTH, NULL_ABLE, ORDERNO, MASK, SET_VALUE_TYPE, IS_ACTIVE) VALUES
(@pst_id, 'ID', 'ID', 'ID', 'text', 'bigint', 20, 'N', 1, '0000000000', 'pk', 'Y'),
(@pst_id, 'TASK_ID', '任务ID', '任务ID', 'text', 'varchar', 255, 'N', 2, '1000000000', 'byPage', 'Y'),
(@pst_id, 'COMMENT', '备注(含区域)', '备注', 'text', 'varchar', 255, 'Y', 10, '1111111111', 'byPage', 'Y'),
(@pst_id, 'REGION', '区域', '区域', 'text', 'varchar', 255, 'N', 9, '1111111111', 'byPage', 'Y'),
(@pst_id, 'SOURCE_TYPE', '源类型', '源类型', 'text', 'varchar', 255, 'N', 10, '1111111111', 'byPage', 'Y'),
(@pst_id, 'SOURCE_URL', '源流地址', '源流地址', 'text', 'varchar', 255, 'N', 11, '1111111111', 'byPage', 'Y'),
(@pst_id, 'TARGET_URL', '目标地址', '目标地址', 'text', 'varchar', 255, 'N', 12, '1111111111', 'byPage', 'Y'),
(@pst_id, 'STATUS', '状态', '状态', 'select', 'varchar', 255, 'Y', 13, '1111111111', 'select', 'Y'),
(@pst_id, 'OPERATOR', '操作人', '操作人', 'text', 'varchar', 255, 'N', 14, '1000000000', 'operator', 'Y'),
(@pst_id, 'ROOM_ID', '直播间ID', '直播间ID', 'fk', 'varchar', 255, 'Y', 16, '1111111111', 'fk', 'Y'),
(@pst_id, 'ROOM_NAME', '直播间名称', '直播间名称', 'text', 'varchar', 255, 'Y', 17, '1100000000', 'byPage', 'Y'),
(@pst_id, 'CREATE_BY', '创建人', '创建人', 'text', 'varchar', 255, 'Y', 18, '0000000000', 'createBy', 'Y'),
(@pst_id, 'CREATE_TIME', '创建时间', '创建时间', 'datetime', 'datetime', 0, 'Y', 19, '0000000000', 'sysdate', 'Y'),
(@pst_id, 'UPDATE_BY', '修改人', '修改人', 'text', 'varchar', 255, 'Y', 20, '0000000000', 'operator', 'Y'),
(@pst_id, 'UPDATE_TIME', '修改时间', '修改时间', 'datetime', 'datetime', 0, 'Y', 21, '0000000000', 'sysdate', 'Y'),
(@pst_id, 'IS_ACTIVE', '是否有效', '是否有效', 'select', 'char', 1, 'Y', 22, '1111111111', 'select', 'Y')
ON DUPLICATE KEY UPDATE DISPLAY_NAME=VALUES(DISPLAY_NAME), ORDERNO=VALUES(ORDERNO);
