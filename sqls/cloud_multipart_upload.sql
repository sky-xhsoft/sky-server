-- =====================================================
-- 云盘分片上传表结构
-- 功能：支持大文件分片上传和断点续传
-- 创建时间：2026-01-15
-- =====================================================

-- 上传会话表
CREATE TABLE IF NOT EXISTS `cloud_upload_session` (
  `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `FILE_ID` VARCHAR(64) NOT NULL COMMENT '文件唯一标识（MD5）',
  `USER_ID` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `FILE_NAME` VARCHAR(255) NOT NULL COMMENT '文件名',
  `FILE_SIZE` BIGINT NOT NULL COMMENT '文件总大小（字节）',
  `FILE_TYPE` VARCHAR(100) COMMENT '文件MIME类型',
  `FOLDER_ID` BIGINT UNSIGNED COMMENT '目标文件夹ID',
  `CHUNK_SIZE` INT NOT NULL DEFAULT 5242880 COMMENT '分片大小（默认5MB）',
  `TOTAL_CHUNKS` INT NOT NULL COMMENT '总分片数',
  `UPLOADED_CHUNKS` TEXT COMMENT '已上传的分片索引（JSON数组）',
  `STATUS` VARCHAR(20) NOT NULL DEFAULT 'uploading' COMMENT '状态：uploading,paused,completed,failed',
  `STORAGE_TYPE` VARCHAR(20) NOT NULL DEFAULT 'local' COMMENT '存储类型：local,oss',
  `STORAGE_PATH` VARCHAR(500) COMMENT '临时存储路径',
  `EXPIRE_TIME` TIMESTAMP NOT NULL COMMENT '过期时间（默认24小时）',
  `ERROR_MESSAGE` TEXT COMMENT '错误信息',

  -- 标准字段
  `SYS_COMPANY_ID` BIGINT UNSIGNED COMMENT '公司ID',
  `CREATE_BY` VARCHAR(50) NOT NULL COMMENT '创建人',
  `CREATE_TIME` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` VARCHAR(50) NOT NULL COMMENT '修改人',
  `UPDATE_TIME` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `IS_ACTIVE` CHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效（Y/N）',

  INDEX `idx_file_id` (`FILE_ID`),
  INDEX `idx_user_id` (`USER_ID`),
  INDEX `idx_status` (`STATUS`),
  INDEX `idx_expire_time` (`EXPIRE_TIME`),
  INDEX `idx_create_time` (`CREATE_TIME`),

  FOREIGN KEY (`USER_ID`) REFERENCES `sys_user`(`ID`),
  FOREIGN KEY (`FOLDER_ID`) REFERENCES `cloud_folder`(`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云盘上传会话表';

-- 分片记录表
CREATE TABLE IF NOT EXISTS `cloud_chunk_record` (
  `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `SESSION_ID` BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
  `CHUNK_INDEX` INT NOT NULL COMMENT '分片索引（从0开始）',
  `CHUNK_SIZE` INT NOT NULL COMMENT '分片大小（字节）',
  `CHUNK_MD5` VARCHAR(32) NOT NULL COMMENT '分片MD5',
  `CHUNK_PATH` VARCHAR(500) COMMENT '分片存储路径',
  `UPLOADED` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已上传（0=否，1=是）',
  `UPLOAD_TIME` TIMESTAMP NULL COMMENT '上传时间',
  `RETRY_COUNT` INT NOT NULL DEFAULT 0 COMMENT '重试次数',

  -- 标准字段
  `CREATE_TIME` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_TIME` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',

  INDEX `idx_session_id` (`SESSION_ID`),
  INDEX `idx_chunk_index` (`CHUNK_INDEX`),
  INDEX `idx_uploaded` (`UPLOADED`),
  UNIQUE KEY `uk_session_chunk` (`SESSION_ID`, `CHUNK_INDEX`),

  FOREIGN KEY (`SESSION_ID`) REFERENCES `cloud_upload_session`(`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云盘分片记录表';

-- 索引优化
ALTER TABLE `cloud_file` ADD INDEX `idx_md5` (`MD5`) COMMENT 'MD5索引，用于秒传';

-- 清理过期会话的定时任务（可选，也可以通过应用层定时任务实现）
-- 创建存储过程清理过期会话
DELIMITER $$

CREATE PROCEDURE IF NOT EXISTS `cleanup_expired_sessions`()
BEGIN
    -- 删除过期的会话（同时会级联删除分片记录）
    DELETE FROM `cloud_upload_session`
    WHERE `EXPIRE_TIME` < NOW()
      AND `STATUS` IN ('uploading', 'paused', 'failed')
      AND `IS_ACTIVE` = 'Y';

    -- 记录清理日志
    SELECT CONCAT('清理了 ', ROW_COUNT(), ' 个过期会话') AS result;
END$$

DELIMITER ;

-- 创建事件（需要启用事件调度器）
-- SET GLOBAL event_scheduler = ON;
-- CREATE EVENT IF NOT EXISTS `cleanup_expired_sessions_event`
-- ON SCHEDULE EVERY 1 HOUR
-- DO CALL cleanup_expired_sessions();

-- 插入示例数据（测试用）
-- INSERT INTO `cloud_upload_session` (
--   `FILE_ID`, `USER_ID`, `FILE_NAME`, `FILE_SIZE`, `FILE_TYPE`,
--   `CHUNK_SIZE`, `TOTAL_CHUNKS`, `UPLOADED_CHUNKS`, `STATUS`,
--   `STORAGE_TYPE`, `STORAGE_PATH`, `EXPIRE_TIME`,
--   `CREATE_BY`, `UPDATE_BY`
-- ) VALUES (
--   'abc123def456', 1, 'test_video.mp4', 104857600, 'video/mp4',
--   5242880, 20, '[]', 'uploading',
--   'local', 'cloud/temp/1/abc123def456', DATE_ADD(NOW(), INTERVAL 24 HOUR),
--   'system', 'system'
-- );
