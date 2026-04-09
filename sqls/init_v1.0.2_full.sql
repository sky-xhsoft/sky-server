-- MySQL dump 10.13  Distrib 8.0.37, for Win64 (x86_64)
--
-- Host: localhost    Database: skyserver
-- ------------------------------------------------------
-- Server version	8.0.37

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `audit_log`
--

DROP TABLE IF EXISTS `audit_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `audit_log` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `USER_ID` int unsigned DEFAULT NULL COMMENT '操作用户ID',
  `USERNAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '操作用户名',
  `ACTION` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '操作类型(login,logout,create,update,delete等)',
  `RESOURCE` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '资源类型(user,table,action,workflow等)',
  `RESOURCE_ID` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '资源ID',
  `RESOURCE_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '资源名称',
  `METHOD` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'HTTP方法',
  `PATH` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '请求路径',
  `IP` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '客户端IP',
  `USER_AGENT` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '用户代理',
  `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '操作状态(success,failure)',
  `ERROR_MESSAGE` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '错误信息',
  `REQUEST_BODY` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '请求体',
  `RESPONSE_BODY` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '响应体',
  `OLD_VALUE` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '修改前的值(JSON)',
  `NEW_VALUE` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '修改后的值(JSON)',
  `DURATION` bigint DEFAULT NULL COMMENT '执行时长(毫秒)',
  `TAGS` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '标签(用于分类和搜索)',
  `CREATED_AT` datetime DEFAULT NULL COMMENT '创建时间',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_audit_user` (`USER_ID`) USING BTREE,
  KEY `idx_audit_action` (`ACTION`) USING BTREE,
  KEY `idx_audit_resource` (`RESOURCE`) USING BTREE,
  KEY `idx_audit_resource_id` (`RESOURCE_ID`) USING BTREE,
  KEY `idx_audit_status` (`STATUS`) USING BTREE,
  KEY `idx_audit_created` (`CREATED_AT`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='审计日志';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `audit_log`
--

LOCK TABLES `audit_log` WRITE;
/*!40000 ALTER TABLE `audit_log` DISABLE KEYS */;
/*!40000 ALTER TABLE `audit_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cloud_chunk_record`
--

DROP TABLE IF EXISTS `cloud_chunk_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cloud_chunk_record` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SESSION_ID` bigint unsigned NOT NULL COMMENT '会话ID',
  `CHUNK_INDEX` int NOT NULL COMMENT '分片索引（从0开始）',
  `CHUNK_SIZE` int NOT NULL COMMENT '分片大小（字节）',
  `CHUNK_MD5` varchar(32) NOT NULL COMMENT '分片MD5',
  `CHUNK_PATH` varchar(500) DEFAULT NULL COMMENT '分片存储路径',
  `UPLOADED` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已上传（0=否，1=是）',
  `UPLOAD_TIME` timestamp NULL DEFAULT NULL COMMENT '上传时间',
  `RETRY_COUNT` int NOT NULL DEFAULT '0' COMMENT '重试次数',
  `CREATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `ETAG` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_session_chunk` (`SESSION_ID`,`CHUNK_INDEX`),
  KEY `idx_session_id` (`SESSION_ID`),
  KEY `idx_chunk_index` (`CHUNK_INDEX`),
  KEY `idx_uploaded` (`UPLOADED`),
  CONSTRAINT `cloud_chunk_record_ibfk_1` FOREIGN KEY (`SESSION_ID`) REFERENCES `cloud_upload_session` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='云盘分片记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cloud_chunk_record`
--

LOCK TABLES `cloud_chunk_record` WRITE;
/*!40000 ALTER TABLE `cloud_chunk_record` DISABLE KEYS */;
/*!40000 ALTER TABLE `cloud_chunk_record` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cloud_item`
--

DROP TABLE IF EXISTS `cloud_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cloud_item` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '涓婚敭ID',
  `ITEM_TYPE` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '椤圭洰绫诲瀷: file, folder',
  `NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '鍚嶇О锛堟枃浠跺悕鎴栨枃浠跺す鍚嶏級',
  `PARENT_ID` bigint unsigned DEFAULT NULL COMMENT '鐖舵枃浠跺すID',
  `PATH` varchar(1000) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '瀹屾暣璺?緞',
  `OWNER_ID` bigint unsigned NOT NULL COMMENT '鎵?湁鑰匢D',
  `STORAGE_TYPE` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '瀛樺偍绫诲瀷: local, oss锛堜粎鏂囦欢锛',
  `STORAGE_PATH` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '瀛樺偍璺?緞锛堜粎鏂囦欢锛',
  `FILE_SIZE` bigint DEFAULT NULL COMMENT '鏂囦欢澶у皬锛堝瓧鑺傦紝浠呮枃浠讹級',
  `FILE_TYPE` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鏂囦欢MIME绫诲瀷锛堜粎鏂囦欢锛',
  `FILE_EXT` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鏂囦欢鎵╁睍鍚嶏紙浠呮枃浠讹級',
  `MD5` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MD5鍊硷紙浠呮枃浠讹級',
  `ACCESS_URL` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '璁块棶URL锛堜粎鏂囦欢锛',
  `THUMBNAIL` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '缂╃暐鍥綰RL锛堜粎鏂囦欢锛',
  `DOWNLOAD_COUNT` int DEFAULT '0' COMMENT '涓嬭浇娆℃暟锛堜粎鏂囦欢锛',
  `TAGS` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鏍囩?锛堥?鍙峰垎闅旓紝浠呮枃浠讹級',
  `FILE_COUNT` int DEFAULT '0' COMMENT '鏂囦欢鏁伴噺锛堜粎鏂囦欢澶癸級',
  `TOTAL_SIZE` bigint DEFAULT '0' COMMENT '鎬诲ぇ灏忥紙瀛楄妭锛屼粎鏂囦欢澶癸級',
  `IS_PUBLIC` char(1) COLLATE utf8mb4_unicode_ci DEFAULT 'N' COMMENT '鏄?惁鍏?紑 Y/N',
  `SHARE_CODE` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鍒嗕韩鐮',
  `SHARE_EXPIRE` datetime DEFAULT NULL COMMENT '鍒嗕韩杩囨湡鏃堕棿',
  `DESCRIPTION` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鎻忚堪',
  `SYS_COMPANY_ID` bigint unsigned DEFAULT NULL COMMENT '鍏?徃ID',
  `CREATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鍒涘缓浜',
  `CREATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '鍒涘缓鏃堕棿',
  `UPDATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '鏇存柊浜',
  `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '鏇存柊鏃堕棿',
  `IS_ACTIVE` char(1) COLLATE utf8mb4_unicode_ci DEFAULT 'Y' COMMENT '鏄?惁鏈夋晥 Y/N',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_share_code` (`SHARE_CODE`),
  KEY `idx_parent_type` (`PARENT_ID`,`ITEM_TYPE`),
  KEY `idx_owner_type` (`OWNER_ID`,`ITEM_TYPE`),
  KEY `idx_type` (`ITEM_TYPE`),
  KEY `idx_md5` (`MD5`),
  KEY `idx_path` (`PATH`(255)),
  KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB AUTO_INCREMENT=480 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='浜戠洏椤圭洰琛?紙鏂囦欢+鏂囦欢澶圭粺涓?級';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cloud_item`
--

LOCK TABLES `cloud_item` WRITE;
/*!40000 ALTER TABLE `cloud_item` DISABLE KEYS */;
INSERT INTO `cloud_item` VALUES (389,'folder','0408',NULL,'/0408',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,0,'N',NULL,NULL,'',1,'admin','2026-04-08 04:21:44','admin','2026-04-08 04:21:44','Y'),(390,'folder','直播录制',389,'/0408/直播录制',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,0,'N',NULL,NULL,'',1,'admin','2026-04-08 04:21:44','admin','2026-04-08 04:21:44','Y'),(391,'folder','直播切片',389,'/0408/直播切片',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,0,'N',NULL,NULL,'',1,'admin','2026-04-08 04:21:44','admin','2026-04-08 04:21:44','Y'),(392,'file','录制_1775593653.mp4',390,'/0408/直播录制/录制_1775593653.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1784886158828627185-193775602cb94938a94a657ba49d30b0/2026-04-08-04-23-17.mp4',198111611,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1784886158828627185-193775602cb94938a94a657ba49d30b0/2026-04-08-04-23-17.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:27:35','admin','2026-04-08 04:40:11','Y'),(393,'file','高光_1775593407_.mp4',391,'/0408/直播切片/高光_1775593407_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_022AA7767CB49E7E30F21D326ACAC1EC53-1775593396902/26-04-08_04-28-39_9267.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_022AA7767CB49E7E30F21D326ACAC1EC53-1775593396902/26-04-08_04-28-39_9267.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:28:42','admin','2026-04-08 04:40:12','Y'),(394,'file','高光_1775593493_.mp4',391,'/0408/直播切片/高光_1775593493_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_022AA7767CB49E7E30F21D326ACAC1EC53-1775593396902/26-04-08_04-28-39_95218.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_022AA7767CB49E7E30F21D326ACAC1EC53-1775593396902/26-04-08_04-28-39_95218.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:28:42','admin','2026-04-08 04:40:16','Y'),(395,'file','高光_1775594603_.mp4',391,'/0408/直播切片/高光_1775594603_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-37_47109.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-37_47109.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:48:39','admin','2026-04-08 04:48:39','Y'),(396,'file','高光_1775594580_.mp4',391,'/0408/直播切片/高光_1775594580_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-37_24744.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-37_24744.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:48:39','admin','2026-04-08 04:48:39','Y'),(397,'file','高光_1775594692_.mp4',391,'/0408/直播切片/高光_1775594692_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-38_136466.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-38_136466.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:48:39','admin','2026-04-08 04:48:39','Y'),(398,'file','高光_1775594628_.mp4',391,'/0408/直播切片/高光_1775594628_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-38_72283.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-48-38_72283.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:48:39','admin','2026-04-08 04:48:39','Y'),(399,'file','高光_1775594973_.mp4',391,'/0408/直播切片/高光_1775594973_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-52-43_419202.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-52-43_419202.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:52:45','admin','2026-04-08 04:52:45','Y'),(400,'file','录制_1775595180.mp4',390,'/0408/直播录制/录制_1775595180.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1782915834287990316-71de6ee15391437b8a1791a3d9d48478/2026-04-08-04-42-35.mp4',483810579,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1782915834287990316-71de6ee15391437b8a1791a3d9d48478/2026-04-08-04-42-35.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:53:03','admin','2026-04-08 04:53:03','Y'),(401,'file','高光_1775595094_.mp4',391,'/0408/直播切片/高光_1775595094_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-53-50_539686.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_020893F4CABBB7AC4FAD3370440DA0EB82-1775594554432/26-04-08_04-53-50_539686.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 04:53:52','admin','2026-04-08 04:53:52','Y'),(402,'file','高光_1775595357_.mp4',391,'/0408/直播切片/高光_1775595357_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-00-40_81051.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-00-40_81051.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 05:01:00','admin','2026-04-08 05:01:00','Y'),(403,'file','录制_1775595727.mp4',390,'/0408/直播录制/录制_1775595727.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1780664034658776192-49c63d26ac344767974d93f4e0c0aa24/2026-04-08-04-54-36.mp4',349767200,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1780664034658776192-49c63d26ac344767974d93f4e0c0aa24/2026-04-08-04-54-36.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 05:02:09','admin','2026-04-08 05:02:09','Y'),(404,'file','高光_1775595633_.mp4',391,'/0408/直播切片/高光_1775595633_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-03-06_359040.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-03-06_359040.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 05:03:08','admin','2026-04-08 05:03:08','Y'),(405,'file','高光_1775595515_.mp4',391,'/0408/直播切片/高光_1775595515_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-03-06_240538.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-03-06_240538.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 05:03:08','admin','2026-04-08 05:03:08','Y'),(406,'file','高光_1775595557_.mp4',391,'/0408/直播切片/高光_1775595557_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-03-06_282914.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_023FC50E692EE1FE78D9766B26CC4445DF-1775595275124/26-04-08_05-03-06_282914.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 05:03:08','admin','2026-04-08 05:03:08','Y'),(407,'file','高光_1775642258_.mp4',391,'/0408/直播切片/高光_1775642258_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-03-01_28938.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-03-01_28938.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:03:02','admin','2026-04-08 18:03:02','Y'),(408,'folder','瑞虎上市',NULL,'/瑞虎上市',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,0,'N',NULL,NULL,'',1,'admin','2026-04-08 18:06:18','admin','2026-04-08 18:06:18','Y'),(409,'folder','直播录制',408,'/瑞虎上市/直播录制',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,0,'N',NULL,NULL,'',1,'admin','2026-04-08 18:06:18','admin','2026-04-08 18:06:18','Y'),(410,'folder','直播切片',408,'/瑞虎上市/直播切片',1,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,0,'N',NULL,NULL,'',1,'admin','2026-04-08 18:06:18','admin','2026-04-08 18:06:18','Y'),(411,'file','录制_1775642793.mp4',390,'/0408/直播录制/录制_1775642793.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1788545346026988950-327254c939e94e23a01d6ab673014d25/2026-04-08-17-57-10.mp4',436181684,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/11/1788545346026988950-327254c939e94e23a01d6ab673014d25/2026-04-08-17-57-10.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:06:34','admin','2026-04-08 18:06:34','Y'),(412,'file','高光_1775642712_.mp4',391,'/0408/直播切片/高光_1775642712_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-12_482883.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-12_482883.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:07:13','admin','2026-04-08 18:07:13','Y'),(413,'file','高光_1775642471_.mp4',391,'/0408/直播切片/高光_1775642471_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-52_242266.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-52_242266.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:07:54','admin','2026-04-08 18:07:54','Y'),(414,'file','高光_1775642493_.mp4',391,'/0408/直播切片/高光_1775642493_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-53_265156.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-53_265156.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:07:54','admin','2026-04-08 18:07:54','Y'),(415,'file','高光_1775642549_.mp4',391,'/0408/直播切片/高光_1775642549_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-53_320936.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/11/LIVE_02F29391C0EC6600F605EA184B3B477DAB-1775642229521/26-04-08_18-07-53_320936.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:07:54','admin','2026-04-08 18:07:54','Y'),(416,'file','高光_1775643398_.mp4',410,'/瑞虎上市/直播切片/高光_1775643398_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-18-31_484316.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-18-31_484316.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:18:32','admin','2026-04-09 17:59:31','N'),(417,'file','高光_1775643498_.mp4',410,'/瑞虎上市/直播切片/高光_1775643498_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-22-40_583920.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-22-40_583920.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:22:41','admin','2026-04-09 17:59:31','N'),(418,'file','高光_1775643533_.mp4',410,'/瑞虎上市/直播切片/高光_1775643533_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-22-40_618767.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-22-40_618767.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:22:41','admin','2026-04-09 17:59:31','N'),(419,'file','高光_1775643665_.mp4',410,'/瑞虎上市/直播切片/高光_1775643665_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-22-40_750891.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-22-40_750891.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:22:41','admin','2026-04-09 17:59:31','N'),(420,'file','高光_1775643693_.mp4',410,'/瑞虎上市/直播切片/高光_1775643693_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-26-34_779038.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-26-34_779038.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:26:35','admin','2026-04-09 17:59:31','N'),(421,'file','高光_1775643916_.mp4',410,'/瑞虎上市/直播切片/高光_1775643916_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-26-34_1002207.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-26-34_1002207.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:26:35','admin','2026-04-09 17:59:31','N'),(422,'file','高光_1775643953_.mp4',410,'/瑞虎上市/直播切片/高光_1775643953_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-30-26_1038852.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-30-26_1038852.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:30:28','admin','2026-04-09 17:59:31','N'),(423,'file','高光_1775644100_.mp4',410,'/瑞虎上市/直播切片/高光_1775644100_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-30-26_1185449.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-30-26_1185449.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:30:28','admin','2026-04-09 17:59:31','N'),(424,'file','高光_1775644338_.mp4',410,'/瑞虎上市/直播切片/高光_1775644338_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-34-31_1424159.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-34-31_1424159.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:34:31','admin','2026-04-09 17:59:31','N'),(425,'file','高光_1775644540_.mp4',410,'/瑞虎上市/直播切片/高光_1775644540_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-39-11_1625425.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-39-11_1625425.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:39:12','admin','2026-04-09 17:59:31','N'),(426,'file','高光_1775644698_.mp4',410,'/瑞虎上市/直播切片/高光_1775644698_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-42-45_1784483.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-42-45_1784483.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:42:47','admin','2026-04-09 17:59:31','N'),(427,'file','高光_1775644667_.mp4',410,'/瑞虎上市/直播切片/高光_1775644667_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-42-45_1753621.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-42-45_1753621.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:42:47','admin','2026-04-09 17:59:31','N'),(428,'file','高光_1775644654_.mp4',410,'/瑞虎上市/直播切片/高光_1775644654_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-42-45_1739964.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-42-45_1739964.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:42:47','admin','2026-04-09 17:59:31','N'),(429,'file','高光_1775644893_.mp4',410,'/瑞虎上市/直播切片/高光_1775644893_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-40_1978914.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-40_1978914.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:46:41','admin','2026-04-09 17:59:44','N'),(430,'file','高光_1775645032_.mp4',410,'/瑞虎上市/直播切片/高光_1775645032_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-40_2118285.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-40_2118285.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:46:41','admin','2026-04-09 17:59:44','N'),(431,'file','高光_1775645048_.mp4',410,'/瑞虎上市/直播切片/高光_1775645048_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-41_2134028.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-41_2134028.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:46:41','admin','2026-04-09 17:59:31','N'),(432,'file','高光_1775644836_.mp4',410,'/瑞虎上市/直播切片/高光_1775644836_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-41_1921638.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-46-41_1921638.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:46:41','admin','2026-04-09 17:59:44','N'),(433,'file','高光_1775645157_.mp4',410,'/瑞虎上市/直播切片/高光_1775645157_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-50-41_2242580.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-50-41_2242580.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:50:43','admin','2026-04-09 17:59:44','N'),(434,'file','高光_1775645280_.mp4',410,'/瑞虎上市/直播切片/高光_1775645280_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-50-41_2365807.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-50-41_2365807.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:50:43','admin','2026-04-09 17:59:44','N'),(435,'file','高光_1775645135_.mp4',410,'/瑞虎上市/直播切片/高光_1775645135_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-50-41_2220244.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-50-41_2220244.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:50:43','admin','2026-04-09 17:59:44','N'),(436,'file','高光_1775645320_.mp4',410,'/瑞虎上市/直播切片/高光_1775645320_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-54-37_2406198.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-54-37_2406198.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:54:38','admin','2026-04-09 17:59:44','N'),(437,'file','高光_1775645566_.mp4',410,'/瑞虎上市/直播切片/高光_1775645566_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-54-37_2652301.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-54-37_2652301.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:54:38','admin','2026-04-09 17:59:44','N'),(438,'file','高光_1775645372_.mp4',410,'/瑞虎上市/直播切片/高光_1775645372_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-54-38_2458336.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-54-38_2458336.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:54:38','admin','2026-04-09 17:59:44','N'),(439,'file','高光_1775645824_.mp4',410,'/瑞虎上市/直播切片/高光_1775645824_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-58-45_2909421.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-58-45_2909421.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:58:47','admin','2026-04-09 17:59:44','N'),(440,'file','高光_1775645622_.mp4',410,'/瑞虎上市/直播切片/高光_1775645622_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-58-45_2707964.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-58-45_2707964.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:58:47','admin','2026-04-09 17:59:44','N'),(441,'file','高光_1775645631_.mp4',410,'/瑞虎上市/直播切片/高光_1775645631_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-58-46_2716110.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_18-58-46_2716110.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 18:58:47','admin','2026-04-09 17:59:44','N'),(442,'file','高光_1775645884_.mp4',410,'/瑞虎上市/直播切片/高光_1775645884_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-02-50_2968926.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-02-50_2968926.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:02:52','admin','2026-04-09 17:59:44','N'),(443,'file','高光_1775645994_.mp4',410,'/瑞虎上市/直播切片/高光_1775645994_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-02-50_3079045.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-02-50_3079045.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:02:52','admin','2026-04-09 17:59:44','N'),(444,'file','高光_1775645850_.mp4',410,'/瑞虎上市/直播切片/高光_1775645850_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-02-50_2935622.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-02-50_2935622.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:02:52','admin','2026-04-09 17:59:44','N'),(445,'file','高光_1775646323_.mp4',410,'/瑞虎上市/直播切片/高光_1775646323_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-07-51_3408639.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-07-51_3408639.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:07:55','admin','2026-04-09 17:59:44','N'),(446,'file','高光_1775646510_.mp4',410,'/瑞虎上市/直播切片/高光_1775646510_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-11-15_3596317.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-11-15_3596317.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:11:14','admin','2026-04-09 17:59:44','N'),(447,'file','高光_1775646488_.mp4',410,'/瑞虎上市/直播切片/高光_1775646488_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-11-15_3573789.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-11-15_3573789.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:11:14','admin','2026-04-09 17:59:44','N'),(448,'file','高光_1775646569_.mp4',410,'/瑞虎上市/直播切片/高光_1775646569_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-14-30_3654024.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-14-30_3654024.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:14:31','admin','2026-04-09 17:59:44','N'),(449,'file','高光_1775646781_.mp4',410,'/瑞虎上市/直播切片/高光_1775646781_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-14-30_3866206.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-14-30_3866206.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:14:31','admin','2026-04-09 17:59:44','N'),(450,'file','高光_1775646610_.mp4',410,'/瑞虎上市/直播切片/高光_1775646610_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-14-30_3695310.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-14-30_3695310.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:14:31','admin','2026-04-09 17:59:44','N'),(451,'file','高光_1775646836_.mp4',410,'/瑞虎上市/直播切片/高光_1775646836_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-18-15_3921830.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-18-15_3921830.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:18:15','admin','2026-04-09 17:59:44','N'),(452,'file','高光_1775647047_.mp4',410,'/瑞虎上市/直播切片/高光_1775647047_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-22-45_4132681.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-22-45_4132681.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:22:47','admin','2026-04-09 17:59:44','N'),(453,'file','高光_1775647207_.mp4',410,'/瑞虎上市/直播切片/高光_1775647207_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-22-45_4292294.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-22-45_4292294.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:22:47','admin','2026-04-09 17:59:44','N'),(454,'file','高光_1775647111_.mp4',410,'/瑞虎上市/直播切片/高光_1775647111_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-22-45_4196914.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-22-45_4196914.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:22:47','admin','2026-04-09 17:59:44','N'),(455,'file','高光_1775647299_.mp4',410,'/瑞虎上市/直播切片/高光_1775647299_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-26-17_4384582.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-26-17_4384582.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:26:17','admin','2026-04-09 17:59:44','N'),(456,'file','高光_1775647554_.mp4',410,'/瑞虎上市/直播切片/高光_1775647554_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-30-17_4639167.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-30-17_4639167.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:30:17','admin','2026-04-09 17:59:44','N'),(457,'file','高光_1775647470_.mp4',410,'/瑞虎上市/直播切片/高光_1775647470_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-30-17_4555609.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_025E2F600D0B400C23243931CEC8C2D6E2-1775642915001/26-04-08_19-30-17_4555609.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 19:30:17','admin','2026-04-09 17:59:44','N'),(458,'file','录制_1775651145.mp4',409,'/瑞虎上市/直播录制/录制_1775651145.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/12/1787137971318988996-df424b6978ab4a92b25f947ce382aa26/2026-04-08-18-08-35.mp4',6284459865,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/12/1787137971318988996-df424b6978ab4a92b25f947ce382aa26/2026-04-08-18-08-35.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 20:25:45','admin','2026-04-08 20:25:45','Y'),(459,'file','垫片.m4v',NULL,'/垫片.m4v',1,'local','cloud/1/2026/04/08/1954cbf9-0d86-49be-96c2-a078367d1831.m4v',368580484,'video/mp4','.m4v',NULL,'/files/cloud/1/2026/04/08/1954cbf9-0d86-49be-96c2-a078367d1831.m4v',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-08 22:17:17','admin','2026-04-08 22:17:21','N'),(460,'file','同传音频版本.mp4',NULL,'',1,'oss','cloud/1/2026/04/08/1ec93ff97-c4a056070eef6a6d90e0e877d9276a61.mp4',8264089495,'video/mp4','.mp4','c4a056070eef6a6d90e0e877d9276a61','https://zhibo-1301212747.cos.ap-nanjing.myqcloud.com/cloud/1/2026/04/08/1ec93ff97-c4a056070eef6a6d90e0e877d9276a61.mp4',NULL,2,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 00:27:55','admin','2026-04-09 02:57:29','N'),(461,'file','all.mp4',NULL,'',1,'oss','cloud/1/2026/04/09/12526efba-6b57995aee52a0fc3f09a1fc42d9e838.mp4',4918276026,'video/mp4','.mp4','6b57995aee52a0fc3f09a1fc42d9e838','https://zhibo-1301212747.cos.ap-nanjing.myqcloud.com/cloud/1/2026/04/09/12526efba-6b57995aee52a0fc3f09a1fc42d9e838.mp4',NULL,2,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 01:46:30','admin','2026-04-09 02:11:55','N'),(462,'file','all.mp4',NULL,'',1,'oss','cloud/1/2026/04/09/12526efba-6b57995aee52a0fc3f09a1fc42d9e838.mp4',4918276026,'video/mp4','.mp4','6b57995aee52a0fc3f09a1fc42d9e838','https://zhibo-1301212747.cos.ap-nanjing.myqcloud.com/cloud/1/2026/04/09/12526efba-6b57995aee52a0fc3f09a1fc42d9e838.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 02:48:10','admin','2026-04-09 02:57:29','N'),(463,'file','无垫片.mp4',NULL,'',1,'oss','cloud/1/2026/04/09/11f17fac1-79a37d98726b59620aa7180f80e5d993.mp4',4816632513,'video/mp4','.mp4','79a37d98726b59620aa7180f80e5d993','https://zhibo-1301212747.cos.ap-nanjing.myqcloud.com/cloud/1/2026/04/09/11f17fac1-79a37d98726b59620aa7180f80e5d993.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 02:50:23','admin','2026-04-09 02:57:29','N'),(467,'file','垫片.m4v',NULL,'/垫片.m4v',1,'oss','cloud/1/2026/04/09/e164a7fd-bd8c-4858-8fe2-c4171ae6f3f1.m4v',368580484,'video/mp4','.m4v',NULL,'https://zhibo-1301212747.cos.ap-nanjing.myqcloud.com/cloud/1/2026/04/09/e164a7fd-bd8c-4858-8fe2-c4171ae6f3f1.m4v',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 16:13:31','admin','2026-04-09 16:14:24','N'),(468,'file','高光_1775727440_.mp4',410,'/瑞虎上市/直播切片/高光_1775727440_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-42-44_27902.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-42-44_27902.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:42:46','admin','2026-04-09 17:59:44','N'),(469,'file','高光_1775727416_.mp4',410,'/瑞虎上市/直播切片/高光_1775727416_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-42-44_3943.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-42-44_3943.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:42:46','admin','2026-04-09 17:59:44','N'),(470,'file','录制_1775727802.mp4',409,'/瑞虎上市/直播录制/录制_1775727802.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/12/1782915868299585578-09f31f32dac94f8392efc9cd5f3ddbf3/2026-04-09-17-36-52.mp4',301687388,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/12/1782915868299585578-09f31f32dac94f8392efc9cd5f3ddbf3/2026-04-09-17-36-52.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:43:24','admin','2026-04-09 17:43:24','Y'),(471,'file','高光_1775727651_.mp4',410,'/瑞虎上市/直播切片/高光_1775727651_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-44-14_239699.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-44-14_239699.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:44:16','admin','2026-04-09 17:59:44','N'),(472,'file','高光_1775727678_.mp4',410,'/瑞虎上市/直播切片/高光_1775727678_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-44-14_266709.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-44-14_266709.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:44:16','admin','2026-04-09 17:59:44','N'),(473,'file','高光_1775727733_.mp4',410,'/瑞虎上市/直播切片/高光_1775727733_.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-44-14_322489.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_0244C3385A5DB176E00108B991C4EDC9B0-1775727412217/26-04-09_17-44-14_322489.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:44:16','admin','2026-04-09 17:59:44','N'),(474,'file','高光_1775728362_奇瑞发布16电池战略及全生命周期绿色政策.mp4',410,'/瑞虎上市/直播切片/高光_1775728362_奇瑞发布16电池战略及全生命周期绿色政策.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_17-53-29_260533.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_17-53-29_260533.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:53:30','admin','2026-04-09 17:59:22','N'),(475,'file','高光_1775728564_奇瑞电池技术蓝图发布.mp4',410,'/瑞虎上市/直播切片/高光_1775728564_奇瑞电池技术蓝图发布.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_17-58-30_462466.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_17-58-30_462466.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:58:32','admin','2026-04-09 17:59:44','N'),(476,'file','高光_1775728664_奇瑞新能源国家使命与法规节点.mp4',410,'/瑞虎上市/直播切片/高光_1775728664_奇瑞新能源国家使命与法规节点.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_17-58-30_562616.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_17-58-30_562616.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 17:58:34','admin','2026-04-09 17:59:44','N'),(477,'file','录制_1775728812.mp4',409,'/瑞虎上市/直播录制/录制_1775728812.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/12/1785167668289916625-9447d53fa42649c7ae6d6db2a47eaa84/2026-04-09-17-48-22.mp4',549299987,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/live/origin/upload.skyzhou.cn/live/12/1785167668289916625-9447d53fa42649c7ae6d6db2a47eaa84/2026-04-09-17-48-22.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 18:00:13','admin','2026-04-09 18:00:13','Y'),(478,'file','高光_1775728759_奇瑞发布2030年电池投资与研发生态政策.mp4',410,'/瑞虎上市/直播切片/高光_1775728759_奇瑞发布2030年电池投资与研发生态政策.mp4',1,'oss','http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_18-00-28_657116.mp4',0,'video/mp4','.mp4',NULL,'http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/12/LIVE_02C401BFBBE1C9186BECB4A8D0B652FBD8-1775728102177/2026-04-09_18-00-28_657116.mp4',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 18:00:29','admin','2026-04-09 18:00:29','Y'),(479,'file','垫片.m4v',NULL,'/垫片.m4v',1,'oss','cloud/1/2026/04/09/c2bf171d-4e57-4985-9312-ad9b75c138e2.m4v',368580484,'video/mp4','.m4v',NULL,'https://zigebo-pro-1401565388.cos.ap-nanjing.myqcloud.com/cloud/1/2026/04/09/c2bf171d-4e57-4985-9312-ad9b75c138e2.m4v',NULL,0,NULL,0,0,'N',NULL,NULL,'',0,'admin','2026-04-09 21:05:20','admin','2026-04-09 21:36:38','N');
/*!40000 ALTER TABLE `cloud_item` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cloud_quota`
--

DROP TABLE IF EXISTS `cloud_quota`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cloud_quota` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `USER_ID` bigint unsigned NOT NULL COMMENT '用户ID',
  `TOTAL_QUOTA` bigint NOT NULL COMMENT '总配额（字节）',
  `USED_SPACE` bigint DEFAULT '0' COMMENT '已用空间（字节）',
  `FILE_COUNT` int DEFAULT '0' COMMENT '文件数量',
  `FOLDER_COUNT` int DEFAULT '0' COMMENT '文件夹数量',
  `MAX_FILE_SIZE` bigint DEFAULT '0' COMMENT '单文件最大大小（字节）',
  `QUOTA_TYPE` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'standard' COMMENT '配额类型: standard, premium',
  `SYS_COMPANY_ID` bigint unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) COLLATE utf8mb4_unicode_ci DEFAULT 'Y' COMMENT '是否有效 Y/N',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_user_id` (`USER_ID`),
  KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘配额表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cloud_quota`
--

LOCK TABLES `cloud_quota` WRITE;
/*!40000 ALTER TABLE `cloud_quota` DISABLE KEYS */;
INSERT INTO `cloud_quota` VALUES (1,1,107374182400,8603318314,25,6,21474836480,'standard',1,'system','2026-01-13 03:15:27','system','2026-04-09 21:36:38','Y');
/*!40000 ALTER TABLE `cloud_quota` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cloud_share`
--

DROP TABLE IF EXISTS `cloud_share`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cloud_share` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SHARE_CODE` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分享码',
  `RESOURCE_TYPE` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型: file, folder',
  `RESOURCE_ID` bigint unsigned NOT NULL COMMENT '资源ID',
  `SHARER_ID` bigint unsigned NOT NULL COMMENT '分享者ID',
  `SHARE_TYPE` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分享类型: public, password, private',
  `PASSWORD` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访问密码',
  `EXPIRE_TIME` datetime DEFAULT NULL COMMENT '过期时间',
  `MAX_DOWNLOADS` int DEFAULT '0' COMMENT '最大下载次数（0=无限制）',
  `DOWNLOAD_COUNT` int DEFAULT '0' COMMENT '已下载次数',
  `VIEW_COUNT` int DEFAULT '0' COMMENT '查看次数',
  `STATUS` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '状态: active, expired, disabled',
  `SYS_COMPANY_ID` bigint unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) COLLATE utf8mb4_unicode_ci DEFAULT 'Y' COMMENT '是否有效 Y/N',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_share_code` (`SHARE_CODE`),
  KEY `idx_resource_id` (`RESOURCE_ID`),
  KEY `idx_sharer_id` (`SHARER_ID`),
  KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘分享记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cloud_share`
--

LOCK TABLES `cloud_share` WRITE;
/*!40000 ALTER TABLE `cloud_share` DISABLE KEYS */;
INSERT INTO `cloud_share` VALUES (1,'lD3cm2ha','folder',128,1,'password','123456','2026-01-29 23:20:10',0,21,21,'disabled',0,'admin','2026-01-28 23:20:10','user_1','2026-01-30 03:13:37','N'),(2,'XSAMXfQm','folder',138,1,'password','11111111',NULL,0,0,0,'disabled',0,'admin','2026-01-28 23:45:37','user_1','2026-01-28 23:48:21','N'),(3,'ZSQO5tJW','folder',138,1,'password','123456','2026-01-30 10:54:58',0,12,12,'disabled',0,'admin','2026-01-29 10:54:58','user_1','2026-01-30 15:27:14','N'),(4,'Ircqmlll','folder',138,1,'password','123456','2026-01-31 03:13:32',0,7,7,'disabled',0,'admin','2026-01-30 03:13:32','user_1','2026-01-30 15:27:12','N'),(5,'Dze6Lz0K','folder',158,1,'password','xtcik0','2026-01-31 17:30:39',0,1,1,'disabled',0,'admin','2026-01-30 17:30:39','user_1','2026-02-27 15:50:36','N'),(6,'kr6OHRFw','file',162,1,'password','ugy0yk','2026-02-12 17:19:06',0,0,0,'disabled',0,'admin','2026-02-05 17:19:06','user_1','2026-02-27 15:50:33','N'),(7,'mTeNsf2s','folder',166,1,'password','mfpudq','2026-03-06 15:30:41',0,0,0,'disabled',0,'admin','2026-02-27 15:30:41','user_1','2026-03-15 04:32:39','N'),(8,'DhQJjv8C','folder',166,1,'password','wbcs49','2026-03-06 15:31:37',0,0,0,'disabled',0,'admin','2026-02-27 15:31:37','user_1','2026-03-15 04:32:36','N'),(9,'YEG4Cfym','folder',234,1,'password','m4mjk9','2026-03-25 18:52:14',0,0,0,'disabled',0,'admin','2026-03-18 18:52:14','user_1','2026-04-07 23:54:12','N'),(10,'Ere549iG','folder',236,1,'password','2gnx24','2026-03-25 19:25:34',0,30,30,'disabled',0,'admin','2026-03-18 19:25:34','user_1','2026-04-07 23:54:09','N'),(11,'CeiehLjk','folder',227,1,'password','12345','2026-04-01 13:02:48',0,0,0,'disabled',0,'admin','2026-03-25 13:02:48','user_1','2026-03-25 13:03:24','N'),(12,'gjEDnVYz','folder',227,1,'public','','2026-04-01 13:03:08',0,0,0,'disabled',0,'admin','2026-03-25 13:03:08','user_1','2026-04-07 23:54:07','N'),(13,'FOCPQjO2','file',356,1,'password','1234','2026-04-01 13:54:00',0,1,1,'disabled',0,'admin','2026-03-25 13:54:00','user_1','2026-04-07 23:54:05','N'),(14,'9deHZbJl','file',366,1,'password','test123','2026-04-10 03:32:35',10,0,0,'disabled',0,'admin','2026-04-03 03:32:35','user_1','2026-04-03 03:32:35','N'),(15,'9FsSjQww','file',369,1,'password','test123','2026-04-10 03:33:48',10,0,0,'disabled',0,'admin','2026-04-03 03:33:48','user_1','2026-04-03 03:33:48','N'),(16,'55mmyLDy','file',372,1,'password','test123','2026-04-10 03:36:38',10,0,0,'disabled',0,'admin','2026-04-03 03:36:38','user_1','2026-04-03 03:36:38','N'),(17,'jZsyTHvR','file',375,1,'password','test123','2026-04-10 03:37:09',10,0,0,'disabled',0,'admin','2026-04-03 03:37:09','user_1','2026-04-03 03:37:09','N'),(18,'f2ygOizt','file',378,1,'password','test123','2026-04-10 03:42:24',10,0,0,'disabled',0,'admin','2026-04-03 03:42:24','user_1','2026-04-03 03:42:24','N'),(19,'VF8YLf75','file',381,1,'password','test123','2026-04-10 03:44:39',10,0,0,'disabled',0,'admin','2026-04-03 03:44:39','user_1','2026-04-03 03:44:39','N'),(20,'4F55Y8TL','folder',408,1,'password','0jk4xt','2026-04-15 18:26:21',0,123,123,'active',0,'admin','2026-04-08 18:26:21','admin','2026-04-10 01:34:54','Y');
/*!40000 ALTER TABLE `cloud_share` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cloud_storage_config`
--

DROP TABLE IF EXISTS `cloud_storage_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cloud_storage_config` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `SYS_COMPANY_ID` int unsigned NOT NULL COMMENT '公司ID',
  `STORAGE_TYPE` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'local' COMMENT '存储类型: local, aliyunOSS, tencentCOS',
  `LOCAL_BASE_PATH` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT 'uploads' COMMENT '本地存储基础路径',
  `LOCAL_BASE_URL` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '/files' COMMENT '本地存储基础URL',
  `ALIYUN_OSS_ENDPOINT` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '阿里云OSS Endpoint',
  `ALIYUN_OSS_ACCESS_KEY_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '阿里云OSS AccessKeyID',
  `ALIYUN_OSS_ACCESS_KEY_SECRET` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '阿里云OSS AccessKeySecret',
  `ALIYUN_OSS_BUCKET_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '阿里云OSS Bucket名称',
  `ALIYUN_OSS_CDN_DOMAIN` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '阿里云OSS CDN域名',
  `TENCENT_COS_BUCKET_URL` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '腾讯云COS Bucket URL',
  `TENCENT_COS_SECRET_ID` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '腾讯云COS SecretID',
  `TENCENT_COS_SECRET_KEY` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '腾讯云COS SecretKey',
  `TENCENT_COS_BUCKET_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '腾讯云COS Bucket名称',
  `TENCENT_COS_REGION` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '腾讯云COS 区域',
  `TENCENT_COS_CDN_DOMAIN` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '腾讯云COS CDN域名',
  `IS_ACTIVE` char(1) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `CREATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idx_company` (`SYS_COMPANY_ID`),
  KEY `idx_storage_type` (`STORAGE_TYPE`),
  KEY `idx_active` (`IS_ACTIVE`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘存储配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cloud_storage_config`
--

LOCK TABLES `cloud_storage_config` WRITE;
/*!40000 ALTER TABLE `cloud_storage_config` DISABLE KEYS */;
INSERT INTO `cloud_storage_config` VALUES (1,1,'local','uploads','/files',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'Y','system','2026-04-08 01:51:27',NULL,'2026-04-08 01:51:27');
/*!40000 ALTER TABLE `cloud_storage_config` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cloud_upload_session`
--

DROP TABLE IF EXISTS `cloud_upload_session`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cloud_upload_session` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `FILE_ID` varchar(64) NOT NULL COMMENT '文件唯一标识（MD5）',
  `USER_ID` bigint unsigned NOT NULL COMMENT '用户ID',
  `FILE_NAME` varchar(255) NOT NULL COMMENT '文件名',
  `FILE_SIZE` bigint NOT NULL COMMENT '文件总大小（字节）',
  `FILE_TYPE` varchar(100) DEFAULT NULL COMMENT '文件MIME类型',
  `FOLDER_ID` bigint unsigned DEFAULT NULL COMMENT '目标文件夹ID',
  `CHUNK_SIZE` int NOT NULL DEFAULT '5242880' COMMENT '分片大小（默认5MB）',
  `TOTAL_CHUNKS` int NOT NULL COMMENT '总分片数',
  `UPLOADED_CHUNKS` text COMMENT '已上传的分片索引（JSON数组）',
  `STATUS` varchar(20) NOT NULL DEFAULT 'uploading' COMMENT '状态：uploading,paused,completed,failed',
  `STORAGE_TYPE` varchar(20) NOT NULL DEFAULT 'local' COMMENT '存储类型：local,oss',
  `STORAGE_PATH` varchar(500) DEFAULT NULL COMMENT '临时存储路径',
  `EXPIRE_TIME` timestamp NOT NULL COMMENT '过期时间（默认24小时）',
  `ERROR_MESSAGE` text COMMENT '错误信息',
  `SYS_COMPANY_ID` bigint unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(50) NOT NULL COMMENT '创建人',
  `CREATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(50) NOT NULL COMMENT '修改人',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效（Y/N）',
  `STORAGE_UPLOAD_ID` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `idx_file_id` (`FILE_ID`),
  KEY `idx_status` (`STATUS`),
  KEY `idx_expire_time` (`EXPIRE_TIME`),
  KEY `idx_create_time` (`CREATE_TIME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='云盘上传会话表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cloud_upload_session`
--

LOCK TABLES `cloud_upload_session` WRITE;
/*!40000 ALTER TABLE `cloud_upload_session` DISABLE KEYS */;
/*!40000 ALTER TABLE `cloud_upload_session` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `live_callback_event`
--

DROP TABLE IF EXISTS `live_callback_event`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `live_callback_event` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `EVENT_TYPE` varchar(50) NOT NULL COMMENT '事件类型：push_stream, disconnect_stream, recording_file, recording_status',
  `EVENT_TIME` bigint NOT NULL COMMENT '事件时间戳（秒）',
  `DOMAIN_NAME` varchar(255) DEFAULT NULL COMMENT '推流域名',
  `APP_NAME` varchar(255) DEFAULT NULL COMMENT '应用名称',
  `STREAM_NAME` varchar(255) DEFAULT NULL COMMENT '流名称',
  `STREAM_ID` varchar(255) DEFAULT NULL COMMENT '流ID',
  `CLIENT_IP` varchar(50) DEFAULT NULL COMMENT '客户端IP',
  `EVENT_DATA` text COMMENT '事件详细数据（JSON格式）',
  `SIGN` varchar(255) DEFAULT NULL COMMENT '签名',
  `T_VALUE` bigint NOT NULL COMMENT '签名过期时间',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `SYS_COMPANY_ID` bigint NOT NULL COMMENT '公司ID',
  `IS_ACTIVE` char(1) DEFAULT 'Y' COMMENT '是否有效',
  `ROOM_NAME` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `idx_event_type` (`EVENT_TYPE`),
  KEY `idx_domain_name` (`DOMAIN_NAME`),
  KEY `idx_app_name` (`APP_NAME`),
  KEY `idx_stream_name` (`STREAM_NAME`),
  KEY `idx_stream_id` (`STREAM_ID`),
  KEY `idx_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_create_time` (`CREATE_TIME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='直播回调事件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `live_callback_event`
--

LOCK TABLES `live_callback_event` WRITE;
/*!40000 ALTER TABLE `live_callback_event` DISABLE KEYS */;
/*!40000 ALTER TABLE `live_callback_event` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `live_room`
--

DROP TABLE IF EXISTS `live_room`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `live_room` (
  `ID` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` bigint unsigned NOT NULL COMMENT '公司ID',
  `ROOM_NAME` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '直播间名称',
  `ROOM_TYPE` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '直播间类型：video(视频直播), image(图片直播), vr(VR直播), audio(语音直播), graphic(图文直播)',
  `BROADCAST_FORMAT` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '播出形式：live(直播), vod(点播/录播), pseudo(伪直播)',
  `ROOM_STAGE` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '直播间阶段：formal(正式直播), test(测试直播)',
  `DISPLAY_MODE` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '显示方式：landscape(横屏), portrait(竖屏), three_screen(三分屏)',
  `START_TIME` datetime DEFAULT NULL COMMENT '开始时间',
  `END_TIME` datetime DEFAULT NULL COMMENT '结束时间',
  `COVER_IMAGE` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '直播间封面',
  `VIEWING_METHOD` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'public' COMMENT '观看方式：public(公开), encrypted(加密), paid(付费), ticket(购票进入), enterprise(企业成员观看), custom(自建成员观看)',
  `VIEWING_PASSWORD` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '观看密码（加密观看时使用）',
  `VIEWING_PRICE` decimal(10,2) DEFAULT NULL COMMENT '观看价格（付费观看时使用）',
  `PLAYBACK_METHOD` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'post_end' COMMENT '回放方式：post_end(结束后回放), real_time(实时回放), no_playback(结束后不回放)',
  `PLAYBACK_VALIDITY` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'unlimited' COMMENT '回放有效期：unlimited(无限制), all_day(全天), partial(部分时段)',
  `PLAYBACK_START_TIME` time DEFAULT NULL COMMENT '回放开始时间（部分时段时使用）',
  `PLAYBACK_END_TIME` time DEFAULT NULL COMMENT '回放结束时间（部分时段时使用）',
  `STREAM_NAME` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '流名称',
  `PUSH_URL` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '推流地址',
  `PLAY_URL` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '播放地址',
  `STATUS` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'draft' COMMENT '状态：draft(草稿), scheduled(已排期), live(直播中), ended(已结束), archived(已归档)',
  `VIEWER_COUNT` int DEFAULT '0' COMMENT '观看人数',
  `PEAK_VIEWER_COUNT` int DEFAULT '0' COMMENT '峰值观看人数',
  `DURATION` int DEFAULT '0' COMMENT '直播时长（秒）',
  `DESCRIPTION` text COLLATE utf8mb4_unicode_ci COMMENT '直播间描述',
  `PROPS` text COLLATE utf8mb4_unicode_ci COMMENT '扩展属性（JSON）',
  `CREATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效：Y(有效), N(无效)',
  PRIMARY KEY (`ID`),
  KEY `idx_sys_company_id` (`SYS_COMPANY_ID`),
  KEY `idx_room_type` (`ROOM_TYPE`),
  KEY `idx_stream_name` (`STREAM_NAME`),
  KEY `idx_status` (`STATUS`)
) ENGINE=InnoDB AUTO_INCREMENT=42 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='直播间表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `live_room`
--

LOCK TABLES `live_room` WRITE;
/*!40000 ALTER TABLE `live_room` DISABLE KEYS */;
INSERT INTO `live_room` VALUES (10,1,'dfddfd',NULL,'live',NULL,NULL,'2026-04-08 00:00:00',NULL,'','public',NULL,NULL,'post_end','unlimited',NULL,NULL,NULL,NULL,NULL,'draft',0,0,0,NULL,NULL,'admin','2026-04-08 04:16:15','admin','2026-04-08 04:21:37','N'),(11,1,'0408',NULL,'live',NULL,NULL,'2026-04-08 00:00:00',NULL,'','public',NULL,NULL,'post_end','unlimited',NULL,NULL,NULL,NULL,NULL,'ended',0,0,0,NULL,NULL,'admin','2026-04-08 04:21:44',NULL,'2026-04-08 18:06:32','Y'),(12,1,'瑞虎上市',NULL,'live',NULL,NULL,'2026-04-08 00:00:01',NULL,'','public',NULL,NULL,'post_end','unlimited',NULL,NULL,NULL,NULL,NULL,'ended',0,0,0,NULL,NULL,'admin','2026-04-08 18:06:18',NULL,'2026-04-09 18:00:12','Y'),(13,1,'test',NULL,'live',NULL,NULL,'2026-04-09 00:00:00',NULL,'','public',NULL,NULL,'post_end','unlimited',NULL,NULL,NULL,NULL,NULL,'draft',0,0,0,NULL,NULL,'admin','2026-04-09 14:35:56','admin','2026-04-09 14:36:02','N');
/*!40000 ALTER TABLE `live_room` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `pull_stream_task`
--

DROP TABLE IF EXISTS `pull_stream_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pull_stream_task` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `TASK_ID` varchar(255) NOT NULL COMMENT '任务ID',
  `COMMENT` varchar(255) DEFAULT NULL COMMENT '任务备注',
  `REGION` varchar(255) NOT NULL COMMENT '地域',
  `SOURCE_TYPE` varchar(255) NOT NULL COMMENT '内容类型',
  `SOURCE_URL` varchar(255) NOT NULL COMMENT '直播源地址',
  `TARGET_URL` varchar(255) NOT NULL COMMENT '目标地址',
  `START_TIME` datetime NOT NULL COMMENT '开始时间',
  `END_TIME` datetime NOT NULL COMMENT '结束时间',
  `STATUS` varchar(255) DEFAULT 'enable' COMMENT '任务状态：enable-启用，pause-暂停',
  `OPERATOR` varchar(255) NOT NULL COMMENT '操作者',
  `CREATE_BY` varchar(255) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(255) DEFAULT NULL COMMENT '修改人',
  `UPDATE_TIME` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `IS_ACTIVE` char(1) DEFAULT 'Y' COMMENT '是否激活：Y-激活，N-禁用',
  `ROOM_ID` varchar(255) DEFAULT NULL COMMENT '直播间ID',
  `ROOM_NAME` varchar(255) DEFAULT NULL COMMENT '直播间名称',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UK_TASK_ID` (`TASK_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='拉流任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `pull_stream_task`
--

LOCK TABLES `pull_stream_task` WRITE;
/*!40000 ALTER TABLE `pull_stream_task` DISABLE KEYS */;
/*!40000 ALTER TABLE `pull_stream_task` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_action`
--

DROP TABLE IF EXISTS `sys_action`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_action` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '动作名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示描述',
  `DISPLAY_TYPE` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示样式(list_button:列表栏按钮,list_menu_item:列表栏菜单,obj_button:单对象界面按钮,obj_menu_item:单对象界面菜单,tab_button:单对象标签页按钮)',
  `ORDERNO` int DEFAULT NULL COMMENT '排序',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '所属表单',
  `FILTER` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示条件',
  `ACTION_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '动作类型(url:URL,sp:存储过程,job:任务程序,js:JavaScript,bsh: OS Shell,py:Python,)',
  `CONTENT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '动作内容',
  `SCRIPTS` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '脚本(javascript将直接部署到页面上)',
  `URLTARGET` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'URL目标页(_blank or div id 去哪里显示url内容)',
  `SAVE_OBJ` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '保存修改(针对ObjButton/ObjMenuItem/TabButton)',
  `COMMENTS` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '提醒 (如果有内容，针对Button和MenuItem, not ListXXX and TreeNode)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='动作定义';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_action`
--

LOCK TABLES `sys_action` WRITE;
/*!40000 ALTER TABLE `sys_action` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_action` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_column`
--

DROP TABLE IF EXISTS `sys_column`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_column` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT '0' COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `MASK` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '字段读写规则',
  `ORDERNO` int DEFAULT NULL COMMENT '序号',
  `DB_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字段名称',
  `COL_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字段类型(varchar,datetime,int,decimal,float,char,datenumber,date)',
  `COL_LENGTH` int DEFAULT NULL COMMENT '字段长度',
  `COL_PRECISION` int DEFAULT NULL COMMENT '字段精度',
  `SYS_TABLE_ID` int unsigned DEFAULT NULL COMMENT '所属表单',
  `IS_DK` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'N' COMMENT '显示键(DK)',
  `IS_AK` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '输入键(AK)',
  `NULL_ABLE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'Y' COMMENT '空值(Y: 是,N: 否)',
  `IS_UPPERCASE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'N' COMMENT '是否大写(Y:是,N:否)',
  `IS_QUERY` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'N' COMMENT '是否查询条件',
  `SUBMETHOD` varchar(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '统计方法(sum:求和)',
  `FULL_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '字段全名',
  `MODIFI_ABLE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '允许界面修改',
  `SET_VALUE_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '赋值方式(pk:pk,docno:单据编号,createBy:创建人,byPage:界面输入,select:下拉选项,fk:外键关联,sysdate:操作时间,operator:操作用户,ignore:忽略)',
  `REF_TABLE_ID` int unsigned DEFAULT NULL COMMENT '关联表id',
  `REF_COLUMN_ID` int unsigned DEFAULT NULL COMMENT '关联字段id',
  `REF_ON_DELETE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '外键删除动作(noAction:无动作)',
  `SEQ` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '单据编号生成器',
  `SYS_DICT_ID` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '数据字典',
  `DEFAULT_VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '默认值',
  `REG_EXPRESSION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '输入校验正则',
  `ERR_MSG` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '正则校验失败提醒',
  `FILTER` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '字段过滤器(sql)',
  `DISPLAY_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示控件(blank,button,hr,check,file,image,select,text,textarea,date,datetime)',
  `DISPLAY_COLS` int DEFAULT NULL COMMENT '显示列数',
  `DISPLAY_ROWS` int DEFAULT NULL COMMENT '显示行数',
  `PROPS` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '扩展属性',
  `IS_SHOW_TITLE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'Y' COMMENT '是否显示备注(Y:是,N:否)',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  `SHOW_COLUMN_ID` int DEFAULT NULL COMMENT '级联显示字段',
  `SHOW_COLUMN_VAL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '级联显示条件',
  `HR_COLUMN_ID` int DEFAULT NULL COMMENT '关联HR折叠字段',
  `SGRADE` int DEFAULT NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE KEY `idx_cloumn_full_name` (`FULL_NAME`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=766 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='系统表字段';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_column`
--

LOCK TABLES `sys_column` WRITE;
/*!40000 ALTER TABLE `sys_column` DISABLE KEYS */;
INSERT INTO `sys_column` VALUES (1,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,1,'Y','Y','N','N','N',NULL,'AUDIT_LOG.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(2,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','操作用户ID','1111111111',20,'USER_ID','int',NULL,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.USER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','操作用户ID',NULL,NULL,NULL,NULL),(3,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','操作用户名','1111111111',30,'USERNAME','varchar',80,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.USERNAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','操作用户名',NULL,NULL,NULL,NULL),(4,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','操作类型(login,logout,create,update,delete等)','1111111111',40,'ACTION','varchar',50,NULL,1,'N',NULL,'N','N','N',NULL,'AUDIT_LOG.ACTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','操作类型(login,logout,create,update,delete等)',NULL,NULL,NULL,NULL),(5,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','资源类型(user,table,action,workflow等)','1111111111',50,'RESOURCE','varchar',100,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.RESOURCE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','资源类型(user,table,action,workflow等)',NULL,NULL,NULL,NULL),(6,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','资源ID','1111111111',60,'RESOURCE_ID','varchar',100,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.RESOURCE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','资源ID',NULL,NULL,NULL,NULL),(7,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','资源名称','1111111111',70,'RESOURCE_NAME','varchar',255,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.RESOURCE_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','资源名称',NULL,NULL,NULL,NULL),(8,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','HTTP方法','1111111111',80,'METHOD','varchar',10,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.METHOD','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','HTTP方法',NULL,NULL,NULL,NULL),(9,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','请求路径','1111111111',90,'PATH','varchar',500,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.PATH','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','请求路径',NULL,NULL,NULL,NULL),(10,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','客户端IP','1111111111',100,'IP','varchar',50,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.IP','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','客户端IP',NULL,NULL,NULL,NULL),(11,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','用户代理','1111111111',110,'USER_AGENT','varchar',500,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.USER_AGENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','用户代理',NULL,NULL,NULL,NULL),(12,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','操作状态(success,failure)','1111111111',120,'STATUS','varchar',20,NULL,1,'N',NULL,'N','N','N',NULL,'AUDIT_LOG.STATUS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','操作状态(success,failure)',NULL,NULL,NULL,NULL),(13,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','错误信息','1111111111',130,'ERROR_MESSAGE','varchar',2000,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.ERROR_MESSAGE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','错误信息',NULL,NULL,NULL,NULL),(14,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','请求体','1111111111',140,'REQUEST_BODY','text',65535,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.REQUEST_BODY','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','请求体',NULL,NULL,NULL,NULL),(15,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','响应体','1111111111',150,'RESPONSE_BODY','text',65535,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.RESPONSE_BODY','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','响应体',NULL,NULL,NULL,NULL),(16,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','修改前的值(JSON)','1111111111',160,'OLD_VALUE','text',65535,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.OLD_VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','修改前的值(JSON)',NULL,NULL,NULL,NULL),(17,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','修改后的值(JSON)','1111111111',170,'NEW_VALUE','text',65535,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.NEW_VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','修改后的值(JSON)',NULL,NULL,NULL,NULL),(18,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','执行时长(毫秒)','1111111111',180,'DURATION','int',NULL,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.DURATION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','执行时长(毫秒)',NULL,NULL,NULL,NULL),(19,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','标签(用于分类和搜索)','1111111111',190,'TAGS','varchar',500,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.TAGS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','标签(用于分类和搜索)',NULL,NULL,NULL,NULL),(20,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','1111111111',200,'CREATED_AT','datetime',NULL,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.CREATED_AT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(21,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',210,'SYS_COMPANY_ID','int',NULL,NULL,1,'N',NULL,'Y','N','N',NULL,'AUDIT_LOG.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(22,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,2,'Y','Y','N','N','N',NULL,'SYS_ACTION.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(23,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(24,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(25,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(26,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(27,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(28,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,2,'N',NULL,'N','N','N',NULL,'SYS_ACTION.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(29,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','动作名称','1111111111',80,'NAME','varchar',80,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','动作名称',NULL,NULL,NULL,NULL),(30,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示描述','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示描述',NULL,NULL,NULL,NULL),(31,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示样式(list_button:列表栏按钮,list_menu_item:列表栏菜单,obj_button:单对象界面按钮,obj_menu_item:单对象界面菜单,tab_button:单对象标签页按钮)','1111111111',100,'DISPLAY_TYPE','varchar',80,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.DISPLAY_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示样式(list_button:列表栏按钮,list_menu_item:列表栏菜单,obj_button:单对象界面按钮,obj_menu_item:单对象界面菜单,tab_button:单对象标签页按钮)',NULL,NULL,NULL,NULL),(32,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','排序','1111111111',110,'ORDERNO','int',NULL,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','排序',NULL,NULL,NULL,NULL),(33,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属表单','1111111111',120,'SYS_TABLE_ID','int',NULL,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属表单',NULL,NULL,NULL,NULL),(34,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示条件','1111111111',130,'FILTER','varchar',255,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.FILTER','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示条件',NULL,NULL,NULL,NULL),(35,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','动作类型(url:URL,sp:存储过程,job:任务程序,js:JavaScript,bsh: OS Shell,py:Python,)','1111111111',140,'ACTION_TYPE','varchar',255,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.ACTION_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','动作类型(url:URL,sp:存储过程,job:任务程序,js:JavaScript,bsh: OS Shell,py:Python,)',NULL,NULL,NULL,NULL),(36,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','动作内容','1111111111',150,'CONTENT','varchar',255,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.CONTENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','动作内容',NULL,NULL,NULL,NULL),(37,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','脚本(javascript将直接部署到页面上)','1111111111',160,'SCRIPTS','varchar',2000,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.SCRIPTS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','脚本(javascript将直接部署到页面上)',NULL,NULL,NULL,NULL),(38,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','URL目标页(_blank or div id 去哪里显示url内容)','1111111111',170,'URLTARGET','varchar',255,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.URLTARGET','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','URL目标页(_blank or div id 去哪里显示url内容)',NULL,NULL,NULL,NULL),(39,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','保存修改(针对ObjButton/ObjMenuItem/TabButton)','1111111111',180,'SAVE_OBJ','varchar',80,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.SAVE_OBJ','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','保存修改(针对ObjButton/ObjMenuItem/TabButton)',NULL,NULL,NULL,NULL),(40,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','提醒 (如果有内容，针对Button和MenuItem, not ListXXX and TreeNode)','1111111111',190,'COMMENTS','varchar',255,NULL,2,'N',NULL,'Y','N','N',NULL,'SYS_ACTION.COMMENTS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','提醒 (如果有内容，针对Button和MenuItem, not ListXXX and TreeNode)',NULL,NULL,NULL,NULL),(41,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,3,'Y','Y','N','N','N',NULL,'SYS_COLUMN.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(42,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,'0',NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(43,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(44,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(45,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(46,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(47,1,'system','2026-01-12 20:52:14','admin','2026-03-14 00:58:59','Y','可用','0000000000',1040,'IS_ACTIVE','char',1,NULL,3,'N',NULL,'N','N','N',NULL,'SYS_COLUMN.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(48,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',80,'DISPLAY_NAME','varchar',255,NULL,3,'N',NULL,'N','N','Y',NULL,'SYS_COLUMN.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(49,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段读写规则','1111111111',90,'MASK','varchar',10,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.MASK','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段读写规则',NULL,NULL,NULL,NULL),(50,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','序号','1111111111',100,'ORDERNO','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','序号',NULL,NULL,NULL,NULL),(51,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段名称','1111111111',110,'DB_NAME','varchar',255,NULL,3,'N',NULL,'N','N','Y',NULL,'SYS_COLUMN.DB_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段名称',NULL,NULL,NULL,NULL),(52,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段类型(varchar,datetime,int,decimal,float,char,datenumber,date)','1111111111',120,'COL_TYPE','varchar',255,NULL,3,'N',NULL,'N','N','N',NULL,'SYS_COLUMN.COL_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段类型(varchar,datetime,int,decimal,float,char,datenumber,date)',NULL,NULL,NULL,NULL),(53,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段长度','1111111111',130,'COL_LENGTH','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.COL_LENGTH','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段长度',NULL,NULL,NULL,NULL),(54,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:20:49','Y','字段精度','1111001111',140,'COL_PRECISION','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.COL_PRECISION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段精度',NULL,NULL,NULL,NULL),(55,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属表单','1111111111',150,'SYS_TABLE_ID','int',NULL,NULL,3,'N',NULL,'N','Y','Y',NULL,'SYS_COLUMN.SYS_TABLE_ID','Y','fk',20,314,'',NULL,NULL,'',NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属表单',NULL,NULL,NULL,NULL),(56,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:46:34','Y','显示键(DK)','1111111111',160,'IS_DK','char',1,NULL,3,'N',NULL,'N','N','N',NULL,'SYS_COLUMN.IS_DK','Y','select',NULL,NULL,NULL,'','1','N',NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','显示键(DK)',NULL,NULL,NULL,NULL),(57,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:54:56','Y','输入键(AK)','1111111111',170,'IS_AK','char',1,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.IS_AK','Y','select',NULL,NULL,NULL,'','1',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','输入键(AK)',NULL,NULL,NULL,NULL),(58,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','空值(Y: 是,N: 否)','1111111111',180,'NULL_ABLE','char',1,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.NULL_ABLE','Y','byPage',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','空值(Y: 是,N: 否)',NULL,NULL,NULL,NULL),(59,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否大写(Y:是,N:否)','1111111111',190,'IS_UPPERCASE','char',1,NULL,3,'N',NULL,'N','N','N',NULL,'SYS_COLUMN.IS_UPPERCASE','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否大写(Y:是,N:否)',NULL,NULL,NULL,NULL),(60,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否查询条件','1111111111',200,'IS_QUERY','char',1,NULL,3,'N',NULL,'N','N','N',NULL,'SYS_COLUMN.IS_QUERY','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否查询条件',NULL,NULL,NULL,NULL),(61,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','统计方法(sum:求和)','1111111111',210,'SUBMETHOD','varchar',3,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SUBMETHOD','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','统计方法(sum:求和)',NULL,NULL,NULL,NULL),(62,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段全名','1111111111',220,'FULL_NAME','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.FULL_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段全名',NULL,NULL,NULL,NULL),(63,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','允许界面修改','1111111111',230,'MODIFI_ABLE','char',1,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.MODIFI_ABLE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','允许界面修改',NULL,NULL,NULL,NULL),(64,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','赋值方式(pk:pk,docno:单据编号,createBy:创建人,byPage:界面输入,select:下拉选项,fk:外键关联,sysdate:操作时间,operator:操作用户,ignore:忽略)','1111111111',240,'SET_VALUE_TYPE','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SET_VALUE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','赋值方式(pk:pk,docno:单据编号,createBy:创建人,byPage:界面输入,select:下拉选项,fk:外键关联,sysdate:操作时间,operator:操作用户,ignore:忽略)',NULL,NULL,NULL,NULL),(65,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联表id','1111111111',250,'REF_TABLE_ID','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.REF_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联表id',NULL,NULL,NULL,NULL),(66,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联字段id','1111111111',260,'REF_COLUMN_ID','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.REF_COLUMN_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联字段id',NULL,NULL,NULL,NULL),(67,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','外键删除动作(noAction:无动作)','1111111111',270,'REF_ON_DELETE','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.REF_ON_DELETE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','外键删除动作(noAction:无动作)',NULL,NULL,NULL,NULL),(68,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','单据编号生成器','1111111111',280,'SEQ','varchar',80,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SEQ','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','单据编号生成器',NULL,NULL,NULL,NULL),(69,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','数据字典','1111111111',290,'SYS_DICT_ID','varchar',80,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SYS_DICT_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','数据字典',NULL,NULL,NULL,NULL),(70,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','默认值','1111111111',300,'DEFAULT_VALUE','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.DEFAULT_VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','默认值',NULL,NULL,NULL,NULL),(71,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','输入校验正则','1111111111',310,'REG_EXPRESSION','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.REG_EXPRESSION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','输入校验正则',NULL,NULL,NULL,NULL),(72,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','正则校验失败提醒','1111111111',320,'ERR_MSG','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.ERR_MSG','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','正则校验失败提醒',NULL,NULL,NULL,NULL),(73,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段过滤器(sql)','1111111111',330,'FILTER','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.FILTER','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段过滤器(sql)',NULL,NULL,NULL,NULL),(74,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示控件(blank,button,hr,check,file,image,select,text,textarea,date,datetime)','1111111111',340,'DISPLAY_TYPE','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.DISPLAY_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示控件(blank,button,hr,check,file,image,select,text,textarea,date,datetime)',NULL,NULL,NULL,NULL),(75,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示列数','1111111111',350,'DISPLAY_COLS','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.DISPLAY_COLS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示列数',NULL,NULL,NULL,NULL),(76,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示行数','1111111111',360,'DISPLAY_ROWS','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.DISPLAY_ROWS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示行数',NULL,NULL,NULL,NULL),(77,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:29:19','Y','扩展属性','1111111111',370,'PROPS','varchar',2000,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.PROPS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'json',2,4,NULL,'Y','扩展属性',NULL,NULL,NULL,NULL),(78,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否显示备注(Y:是,N:否)','1111111111',380,'IS_SHOW_TITLE','char',1,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.IS_SHOW_TITLE','Y','byPage',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否显示备注(Y:是,N:否)',NULL,NULL,NULL,NULL),(79,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',390,'DESCRIPTION','varchar',2000,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(80,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','级联显示字段','1111111111',400,'SHOW_COLUMN_ID','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SHOW_COLUMN_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','级联显示字段',NULL,NULL,NULL,NULL),(81,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','级联显示条件','1111111111',410,'SHOW_COLUMN_VAL','varchar',255,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SHOW_COLUMN_VAL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','级联显示条件',NULL,NULL,NULL,NULL),(82,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联HR折叠字段','1111111111',420,'HR_COLUMN_ID','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.HR_COLUMN_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联HR折叠字段',NULL,NULL,NULL,NULL),(83,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段访问级别','1111111111',430,'SGRADE','int',NULL,NULL,3,'N',NULL,'Y','N','N',NULL,'SYS_COLUMN.SGRADE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段访问级别',NULL,NULL,NULL,NULL),(84,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,4,'N','N','N','N','N',NULL,'SYS_COMPANY.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(86,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,4,'N','N','Y','N','N',NULL,'SYS_COMPANY.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(87,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,4,'N','N','Y','N','N',NULL,'SYS_COMPANY.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(88,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,4,'N','N','Y','N','N',NULL,'SYS_COMPANY.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(89,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,4,'N','N','Y','N','N',NULL,'SYS_COMPANY.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(90,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:50:18','Y','可用','1111111111',1050,'IS_ACTIVE','char',1,NULL,4,'N','N','N','N','N',NULL,'SYS_COMPANY.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(91,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司名称','1111111111',80,'NAME','varchar',255,NULL,4,'Y','Y','N','N','Y',NULL,'SYS_COMPANY.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(92,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,5,'N','N','N','N','N',NULL,'SYS_DICT.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(93,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(94,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(95,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(96,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(97,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,5,'N',NULL,'N','N','N',NULL,'SYS_DICT.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(98,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,5,'N',NULL,'N','N','N',NULL,'SYS_DICT.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(99,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字典名称','1111111111',80,'NAME','varchar',255,NULL,5,'Y','Y','N','N','Y',NULL,'SYS_DICT.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字典名称',NULL,NULL,NULL,NULL),(100,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,5,'N',NULL,'N','N','Y',NULL,'SYS_DICT.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(101,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段类型(0: String, 1: int)','1111111111',100,'TYPE','int',NULL,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段类型(0: String, 1: int)',NULL,NULL,NULL,NULL),(102,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',110,'DESCRIPTION','varchar',2000,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(103,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','默认值','0010100000',120,'DEFAULT_VALUE','varchar',255,NULL,5,'N',NULL,'Y','N','N',NULL,'SYS_DICT.DEFAULT_VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','默认值',NULL,NULL,NULL,NULL),(104,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,6,'Y','Y','N','N','N',NULL,'SYS_DICT_ITEM.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(105,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(106,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(107,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(108,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(109,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(110,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,6,'N',NULL,'N','N','N',NULL,'SYS_DICT_ITEM.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(111,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属字典','0000000000',80,'SYS_DICT_ID','int',NULL,NULL,6,'N',NULL,'N','N','N',NULL,'SYS_DICT_ITEM.SYS_DICT_ID','Y','fk',5,99,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属字典',NULL,NULL,NULL,NULL),(112,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,6,'N',NULL,'N','N','N',NULL,'SYS_DICT_ITEM.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(113,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字典值','1111111111',100,'VALUE','varchar',255,NULL,6,'N',NULL,'N','N','N',NULL,'SYS_DICT_ITEM.VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字典值',NULL,NULL,NULL,NULL),(114,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','排序','1111111111',110,'ORDERNO','int',NULL,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','排序',NULL,NULL,NULL,NULL),(115,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','css','1111111111',120,'CSSCLASS','varchar',255,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.CSSCLASS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','css',NULL,NULL,NULL,NULL),(116,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',130,'DESCRIPTION','varchar',2000,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(117,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否默认值(Y:是,N:否)','1111111111',140,'IS_DEFAULT_VALUE','char',1,NULL,6,'N',NULL,'Y','N','N',NULL,'SYS_DICT_ITEM.IS_DEFAULT_VALUE','Y','select',NULL,NULL,NULL,NULL,'1','N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否默认值(Y:是,N:否)',NULL,NULL,NULL,NULL),(118,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,7,'Y','Y','N','N','N',NULL,'SYS_DIRECTORY.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(119,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(120,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(121,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(122,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(123,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(124,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,7,'N',NULL,'N','N','N',NULL,'SYS_DIRECTORY.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(125,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1111111111',80,'NAME','varchar',255,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(126,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(127,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属表类别','1111111111',100,'SYS_TABLE_CATEGORY_ID','int',NULL,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.SYS_TABLE_CATEGORY_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属表类别',NULL,NULL,NULL,NULL),(128,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','位置','1111111111',110,'URL','varchar',255,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','位置',NULL,NULL,NULL,NULL),(129,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','对应表','1111111111',120,'SYS_TABLE_ID','int',NULL,NULL,7,'N',NULL,'Y','N','N',NULL,'SYS_DIRECTORY.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','对应表',NULL,NULL,NULL,NULL),(130,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主键','0000000000',10,'ID','int',NULL,NULL,8,'Y','Y','N','N','N',NULL,'SYS_EMAIL_CONFIG.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(131,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,8,'N',NULL,'Y','N','N',NULL,'SYS_EMAIL_CONFIG.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(132,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,8,'N',NULL,'Y','N','N',NULL,'SYS_EMAIL_CONFIG.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(133,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(134,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,8,'N',NULL,'Y','N','N',NULL,'SYS_EMAIL_CONFIG.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(135,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(136,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(137,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SMTP服务器地址','1111111111',80,'SMTP_HOST','varchar',100,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.SMTP_HOST','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','SMTP服务器地址',NULL,NULL,NULL,NULL),(138,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SMTP端口','1111111111',90,'SMTP_PORT','int',NULL,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.SMTP_PORT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','SMTP端口',NULL,NULL,NULL,NULL),(139,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SMTP用户名','1111111111',100,'SMTP_USER','varchar',100,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.SMTP_USER','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','SMTP用户名',NULL,NULL,NULL,NULL),(140,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SMTP密码（加密存储）','1111111111',110,'SMTP_PASSWORD','varchar',255,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.SMTP_PASSWORD','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','SMTP密码（加密存储）',NULL,NULL,NULL,NULL),(141,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','发件人邮箱','1111111111',120,'FROM_EMAIL','varchar',100,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.FROM_EMAIL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','发件人邮箱',NULL,NULL,NULL,NULL),(142,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','发件人名称','1111111111',130,'FROM_NAME','varchar',100,NULL,8,'N',NULL,'Y','N','N',NULL,'SYS_EMAIL_CONFIG.FROM_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','发件人名称',NULL,NULL,NULL,NULL),(143,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否使用TLS Y/N','1111111111',140,'USE_TLS','char',1,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.USE_TLS','Y','byPage',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否使用TLS Y/N',NULL,NULL,NULL,NULL),(144,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否默认配置 Y/N','1111111111',150,'IS_DEFAULT','char',1,NULL,8,'N',NULL,'N','N','N',NULL,'SYS_EMAIL_CONFIG.IS_DEFAULT','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否默认配置 Y/N',NULL,NULL,NULL,NULL),(145,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','描述','1111111111',160,'DESCRIPTION','varchar',500,NULL,8,'N',NULL,'Y','N','N',NULL,'SYS_EMAIL_CONFIG.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','描述',NULL,NULL,NULL,NULL),(146,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主键','0000000000',10,'ID','int',NULL,NULL,9,'Y','Y','N','N','N',NULL,'SYS_FILE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(147,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(148,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(149,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(150,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(151,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(152,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(153,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','原始文件名','1111111111',80,'FILE_NAME','varchar',255,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.FILE_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','原始文件名',NULL,NULL,NULL,NULL),(154,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','存储文件名（唯一）','1111111111',90,'STORAGE_NAME','varchar',255,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.STORAGE_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','存储文件名（唯一）',NULL,NULL,NULL,NULL),(155,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件路径','1111111111',100,'FILE_PATH','varchar',500,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.FILE_PATH','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件路径',NULL,NULL,NULL,NULL),(156,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件大小（字节）','1111111111',110,'FILE_SIZE','int',NULL,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.FILE_SIZE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件大小（字节）',NULL,NULL,NULL,NULL),(157,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件类型/MIME类型','1111111111',120,'FILE_TYPE','varchar',100,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.FILE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件类型/MIME类型',NULL,NULL,NULL,NULL),(158,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件扩展名','1111111111',130,'FILE_EXT','varchar',20,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.FILE_EXT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件扩展名',NULL,NULL,NULL,NULL),(159,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','存储类型：local, oss, s3','1111111111',140,'STORAGE_TYPE','varchar',20,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.STORAGE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,'local',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','存储类型：local, oss, s3',NULL,NULL,NULL,NULL),(160,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','存储桶名称（云存储）','1111111111',150,'BUCKET_NAME','varchar',100,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.BUCKET_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','存储桶名称（云存储）',NULL,NULL,NULL,NULL),(161,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','访问URL','1111111111',160,'ACCESS_URL','varchar',500,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.ACCESS_URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','访问URL',NULL,NULL,NULL,NULL),(162,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','缩略图URL','1111111111',170,'THUMBNAIL_URL','varchar',500,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.THUMBNAIL_URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','缩略图URL',NULL,NULL,NULL,NULL),(163,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件MD5值','1111111111',180,'MD5','varchar',32,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.MD5','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件MD5值',NULL,NULL,NULL,NULL),(164,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','上传IP','1111111111',190,'UPLOAD_IP','varchar',50,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.UPLOAD_IP','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','上传IP',NULL,NULL,NULL,NULL),(165,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','下载次数','1111111111',200,'DOWNLOAD_COUNT','int',NULL,NULL,9,'N',NULL,'N','N','N',NULL,'SYS_FILE.DOWNLOAD_COUNT','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','下载次数',NULL,NULL,NULL,NULL),(166,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件分类','1111111111',210,'CATEGORY','varchar',50,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.CATEGORY','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件分类',NULL,NULL,NULL,NULL),(167,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','文件描述','1111111111',220,'DESCRIPTION','varchar',500,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','文件描述',NULL,NULL,NULL,NULL),(168,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','过期时间','1111111111',230,'EXPIRE_TIME','datetime',NULL,NULL,9,'N',NULL,'Y','N','N',NULL,'SYS_FILE.EXPIRE_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','过期时间',NULL,NULL,NULL,NULL),(190,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主键','0000000000',10,'ID','int',NULL,NULL,12,'Y','Y','N','N','N',NULL,'SYS_MESSAGE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(191,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(192,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(193,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(194,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(195,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(196,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(197,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息标题','1111111111',80,'TITLE','varchar',255,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.TITLE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','消息标题',NULL,NULL,NULL,NULL),(198,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息内容','1111111111',90,'CONTENT','text',65535,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.CONTENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','消息内容',NULL,NULL,NULL,NULL),(199,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息类型: system, workflow, business, notice','1111111111',100,'MESSAGE_TYPE','varchar',50,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.MESSAGE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','消息类型: system, workflow, business, notice',NULL,NULL,NULL,NULL),(200,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','优先级: 0=普通, 1=重要, 2=紧急','1111111111',110,'PRIORITY','int',NULL,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.PRIORITY','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','优先级: 0=普通, 1=重要, 2=紧急',NULL,NULL,NULL,NULL),(201,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息分类','1111111111',120,'CATEGORY','varchar',50,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.CATEGORY','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','消息分类',NULL,NULL,NULL,NULL),(202,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','发送者ID（系统消息为NULL）','1111111111',130,'SENDER_ID','int',NULL,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.SENDER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','发送者ID（系统消息为NULL）',NULL,NULL,NULL,NULL),(203,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','发送者姓名','1111111111',140,'SENDER_NAME','varchar',100,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.SENDER_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','发送者姓名',NULL,NULL,NULL,NULL),(204,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','目标类型: user, role, group, all','1111111111',150,'TARGET_TYPE','varchar',20,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.TARGET_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,'user',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','目标类型: user, role, group, all',NULL,NULL,NULL,NULL),(205,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','目标ID列表（逗号分隔）','1111111111',160,'TARGET_IDS','varchar',1000,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.TARGET_IDS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','目标ID列表（逗号分隔）',NULL,NULL,NULL,NULL),(206,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联URL','1111111111',170,'LINK_URL','varchar',500,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.LINK_URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','关联URL',NULL,NULL,NULL,NULL),(207,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','链接类型: internal, external','1111111111',180,'LINK_TYPE','varchar',50,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.LINK_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','链接类型: internal, external',NULL,NULL,NULL,NULL),(208,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息参数（JSON）','1111111111',190,'PARAMS','text',65535,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.PARAMS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','消息参数（JSON）',NULL,NULL,NULL,NULL),(209,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息模板ID','1111111111',200,'TEMPLATE_ID','int',NULL,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.TEMPLATE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','消息模板ID',NULL,NULL,NULL,NULL),(210,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','已读人数','1111111111',210,'READ_COUNT','int',NULL,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.READ_COUNT','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','已读人数',NULL,NULL,NULL,NULL),(211,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','总接收人数','1111111111',220,'TOTAL_COUNT','int',NULL,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.TOTAL_COUNT','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','总接收人数',NULL,NULL,NULL,NULL),(212,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','过期时间','1111111111',230,'EXPIRE_TIME','datetime',NULL,NULL,12,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE.EXPIRE_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','过期时间',NULL,NULL,NULL,NULL),(213,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','状态: active, expired, deleted','1111111111',240,'STATUS','varchar',20,NULL,12,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE.STATUS','Y','byPage',NULL,NULL,NULL,NULL,NULL,'active',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','状态: active, expired, deleted',NULL,NULL,NULL,NULL),(214,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主键','0000000000',10,'ID','int',NULL,NULL,13,'Y','Y','N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(215,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,13,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE_TEMPLATE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(216,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,13,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE_TEMPLATE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(217,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(218,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,13,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE_TEMPLATE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(219,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(220,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(221,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','模板代码','1111111111',80,'CODE','varchar',50,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.CODE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','模板代码',NULL,NULL,NULL,NULL),(222,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','模板名称','1111111111',90,'NAME','varchar',100,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','模板名称',NULL,NULL,NULL,NULL),(223,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息类型','1111111111',100,'MESSAGE_TYPE','varchar',50,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.MESSAGE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','消息类型',NULL,NULL,NULL,NULL),(224,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','标题模板','1111111111',110,'TITLE','varchar',255,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.TITLE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','标题模板',NULL,NULL,NULL,NULL),(225,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','内容模板','1111111111',120,'CONTENT','text',65535,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.CONTENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','内容模板',NULL,NULL,NULL,NULL),(226,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','变量列表（逗号分隔）','1111111111',130,'VARIABLES','varchar',500,NULL,13,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE_TEMPLATE.VARIABLES','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','变量列表（逗号分隔）',NULL,NULL,NULL,NULL),(227,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','描述','1111111111',140,'DESCRIPTION','varchar',500,NULL,13,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE_TEMPLATE.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','描述',NULL,NULL,NULL,NULL),(228,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否启用 Y/N','1111111111',150,'IS_ENABLED','char',1,NULL,13,'N',NULL,'N','N','N',NULL,'SYS_MESSAGE_TEMPLATE.IS_ENABLED','Y','byPage',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否启用 Y/N',NULL,NULL,NULL,NULL),(229,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','分类','1111111111',160,'CATEGORY','varchar',50,NULL,13,'N',NULL,'Y','N','N',NULL,'SYS_MESSAGE_TEMPLATE.CATEGORY','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','分类',NULL,NULL,NULL,NULL),(230,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,14,'Y','Y','N','N','N',NULL,'SYS_MODEL.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(231,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,14,'N',NULL,'Y','N','N',NULL,'SYS_MODEL.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(232,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,14,'N',NULL,'Y','N','N',NULL,'SYS_MODEL.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(233,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,14,'N',NULL,'Y','N','N',NULL,'SYS_MODEL.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(234,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,14,'N',NULL,'Y','N','N',NULL,'SYS_MODEL.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(235,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,14,'N',NULL,'Y','N','N',NULL,'SYS_MODEL.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(236,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,14,'N',NULL,'N','N','N',NULL,'SYS_MODEL.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(237,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主键','0000000000',10,'ID','int',NULL,NULL,15,'Y','Y','N','N','N',NULL,'SYS_NOTIFICATION_LOG.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(238,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(239,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(240,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,15,'N',NULL,'N','N','N',NULL,'SYS_NOTIFICATION_LOG.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(241,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(242,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,15,'N',NULL,'N','N','N',NULL,'SYS_NOTIFICATION_LOG.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(243,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,15,'N',NULL,'N','N','N',NULL,'SYS_NOTIFICATION_LOG.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(244,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息ID','1111111111',80,'MESSAGE_ID','int',NULL,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.MESSAGE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','消息ID',NULL,NULL,NULL,NULL),(245,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','接收用户ID','1111111111',90,'USER_ID','int',NULL,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.USER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','接收用户ID',NULL,NULL,NULL,NULL),(246,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','通知类型: websocket, email, sms','1111111111',100,'NOTIFY_TYPE','varchar',20,NULL,15,'N',NULL,'N','N','N',NULL,'SYS_NOTIFICATION_LOG.NOTIFY_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','通知类型: websocket, email, sms',NULL,NULL,NULL,NULL),(247,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','状态: pending, sent, failed, read','1111111111',110,'STATUS','varchar',20,NULL,15,'N',NULL,'N','N','N',NULL,'SYS_NOTIFICATION_LOG.STATUS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','状态: pending, sent, failed, read',NULL,NULL,NULL,NULL),(248,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','发送时间','1111111111',120,'SENT_TIME','datetime',NULL,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.SENT_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','发送时间',NULL,NULL,NULL,NULL),(249,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','读取时间','1111111111',130,'READ_TIME','datetime',NULL,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.READ_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','读取时间',NULL,NULL,NULL,NULL),(250,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','错误信息','1111111111',140,'ERROR_MESSAGE','varchar',500,NULL,15,'N',NULL,'Y','N','N',NULL,'SYS_NOTIFICATION_LOG.ERROR_MESSAGE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','错误信息',NULL,NULL,NULL,NULL),(251,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','重试次数','1111111111',150,'RETRY_COUNT','int',NULL,NULL,15,'N',NULL,'N','N','N',NULL,'SYS_NOTIFICATION_LOG.RETRY_COUNT','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','重试次数',NULL,NULL,NULL,NULL),(252,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,16,'Y','Y','N','N','N',NULL,'SYS_OBJUICONF.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(253,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(254,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(255,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(256,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(257,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(258,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,16,'N',NULL,'N','N','N',NULL,'SYS_OBJUICONF.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(259,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1111111111',80,'NAME','varchar',255,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(260,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(261,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','tableid参数名','1111111111',100,'TABLE_PARAM_NAME','varchar',255,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.TABLE_PARAM_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','tableid参数名',NULL,NULL,NULL,NULL),(262,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','id参数名','1111111111',110,'PK_PARAM_NAME','varchar',255,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.PK_PARAM_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','id参数名',NULL,NULL,NULL,NULL),(263,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','CSS类','1111111111',120,'CSS_CLASS','varchar',255,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.CSS_CLASS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','CSS类',NULL,NULL,NULL,NULL),(264,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','每行字段个数','1111111111',130,'COLS','int',NULL,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.COLS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','每行字段个数',NULL,NULL,NULL,NULL),(265,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','缺省动作','1111111111',140,'DEFAULT_ACTION','varchar',10,NULL,16,'N',NULL,'Y','N','N',NULL,'SYS_OBJUICONF.DEFAULT_ACTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','缺省动作',NULL,NULL,NULL,NULL),(266,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,17,'Y','Y','N','N','N',NULL,'SYS_PARAM.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(267,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(268,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(269,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(270,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(271,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(272,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,17,'N',NULL,'N','N','N',NULL,'SYS_PARAM.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(273,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1111111111',80,'NAME','varchar',255,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(274,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','默认值','1111111111',90,'DEFAULT_VALUE','varchar',255,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.DEFAULT_VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','默认值',NULL,NULL,NULL,NULL),(275,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','当前值','1111111111',100,'VALUE','varchar',255,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','当前值',NULL,NULL,NULL,NULL),(276,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','值类型','1111111111',110,'VALUE_TYPE','char',3,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.VALUE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','值类型',NULL,NULL,NULL,NULL),(277,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','值列表','1111111111',120,'VALUE_LIST','varchar',255,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.VALUE_LIST','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','值列表',NULL,NULL,NULL,NULL),(278,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',130,'DESCRIPTION','varchar',255,NULL,17,'N',NULL,'Y','N','N',NULL,'SYS_PARAM.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(279,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,18,'Y','Y','N','N','N',NULL,'SYS_SEQ.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(280,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(281,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(282,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(283,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(284,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(285,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,18,'N',NULL,'N','N','N',NULL,'SYS_SEQ.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(286,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1111111111',80,'NAME','varchar',255,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(287,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(288,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','格式','1111111111',100,'VFORMAT','varchar',255,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.VFORMAT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','格式',NULL,NULL,NULL,NULL),(289,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','递增','1111111111',110,'INCRE','int',NULL,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.INCRE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','递增',NULL,NULL,NULL,NULL),(290,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','循环方式','1111111111',120,'CYCLETYPE','char',1,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.CYCLETYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','循环方式',NULL,NULL,NULL,NULL),(291,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','前缀','1111111111',130,'PREFIX','varchar',10,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.PREFIX','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','前缀',NULL,NULL,NULL,NULL),(292,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','后缀','1111111111',140,'SUFFIX','varchar',10,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.SUFFIX','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','后缀',NULL,NULL,NULL,NULL),(293,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','当前周期值','1111111111',150,'CUR_DATE','varchar',20,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.CUR_DATE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','当前周期值',NULL,NULL,NULL,NULL),(294,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','当前流水号','1111111111',160,'CUR_NUM','int',NULL,NULL,18,'N',NULL,'Y','N','N',NULL,'SYS_SEQ.CUR_NUM','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','当前流水号',NULL,NULL,NULL,NULL),(295,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,19,'Y','Y','N','N','N',NULL,'SYS_SUBSYSTEM.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(296,1,'system','2026-01-12 20:52:14','admin','2026-01-21 05:06:28','Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,19,'N',NULL,'Y','N','Y',NULL,'SYS_SUBSYSTEM.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(297,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(298,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(299,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(300,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(301,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,19,'N',NULL,'N','N','N',NULL,'SYS_SUBSYSTEM.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(302,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1111111111',80,'NAME','varchar',255,NULL,19,'N',NULL,'N','N','Y',NULL,'SYS_SUBSYSTEM.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(303,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','序号','1111111111',90,'ORDERNO','int',NULL,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','序号',NULL,NULL,NULL,NULL),(304,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','网页链接','1111111111',100,'URL','varchar',255,NULL,19,'N',NULL,'Y','N','Y',NULL,'SYS_SUBSYSTEM.URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','网页链接',NULL,NULL,NULL,NULL),(305,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','icon','1111111111',110,'ICON','varchar',255,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.ICON','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','icon',NULL,NULL,NULL,NULL),(306,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',120,'DESCRIPTION','varchar',255,NULL,19,'N',NULL,'Y','N','N',NULL,'SYS_SUBSYSTEM.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(307,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,20,'N','Y','N','N','N',NULL,'SYS_TABLE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(308,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(309,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(310,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(311,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(312,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(313,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,20,'N',NULL,'N','N','N',NULL,'SYS_TABLE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(314,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1110111111',80,'NAME','varchar',255,NULL,20,'N',NULL,'N','Y','Y',NULL,'SYS_TABLE.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(315,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,20,'Y',NULL,'N','N','Y',NULL,'SYS_TABLE.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(316,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','实际数据库表','1111111111',100,'REAL_TABLE_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.REAL_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','实际数据库表',NULL,NULL,NULL,NULL),(317,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','数据过滤','1111111111',110,'FILTER','varchar',2000,NULL,20,'N',NULL,'Y','N','Y',NULL,'SYS_TABLE.FILTER','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',2,4,NULL,'Y','数据过滤',NULL,NULL,NULL,NULL),(318,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示主键(DK)','1111111111',120,'DK_COLUMN_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.DK_COLUMN_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','显示主键(DK)',NULL,NULL,NULL,NULL),(319,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','输入主键(AK)','1111111111',130,'AK_COLUMN_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.AK_COLUMN_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','输入主键(AK)',NULL,NULL,NULL,NULL),(320,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','表单规则(支持：A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废)','1111111111',140,'MASK','char',10,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.MASK','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','表单规则(支持：A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废)',NULL,NULL,NULL,NULL),(321,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','表类别','1111111111',150,'SYS_TABLECATEGORY_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.SYS_TABLECATEGORY_ID','Y','fk',21,343,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','表类别',NULL,NULL,NULL,NULL),(322,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','排序','1111111111',160,'ORDERNO','int',NULL,NULL,20,'N',NULL,'Y','N','Y',NULL,'SYS_TABLE.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y','排序',NULL,NULL,NULL,NULL),(323,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','网页连接','1111111111',170,'URL','varchar',255,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','网页连接',NULL,NULL,NULL,NULL),(324,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','rpc 方法','1111111111',180,'RPC_NAME','varchar',255,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.RPC_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','rpc 方法',NULL,NULL,NULL,NULL),(325,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否菜单(Y:是,N:否)','1111111111',190,'IS_MENU','char',1,NULL,20,'N',NULL,'Y','N','Y',NULL,'SYS_TABLE.IS_MENU','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否菜单(Y:是,N:否)',NULL,NULL,NULL,NULL),(326,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','表单ICO图片','1111111111',200,'ICO_IMG','varchar',255,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.ICO_IMG','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','表单ICO图片',NULL,NULL,NULL,NULL),(327,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否下拉框','1111111111',210,'IS_DROPDOWN','char',1,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.IS_DROPDOWN','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否下拉框',NULL,NULL,NULL,NULL),(328,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:24:06','Y','显示配置','1111111111',220,'SYS_OBJUICONF_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.SYS_OBJUICONF_ID','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示配置',NULL,NULL,NULL,NULL),(329,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','安全目录','1111111111',230,'SYS_DIRECTORY_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.SYS_DIRECTORY_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','安全目录',NULL,NULL,NULL,NULL),(330,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','父表','1111111111',240,'SYS_PARENT_TABLE_ID','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.SYS_PARENT_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','父表',NULL,NULL,NULL,NULL),(331,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','统计行数','1111111111',250,'ROWCNT','int',NULL,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.ROWCNT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','统计行数',NULL,NULL,NULL,NULL),(332,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否海量','1111111111',260,'IS_BIG','char',1,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.IS_BIG','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否海量',NULL,NULL,NULL,NULL),(333,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','扩展属性','1111111111',270,'PROPS','varchar',2000,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.PROPS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'json',2,4,NULL,'Y','扩展属性',NULL,NULL,NULL,NULL),(334,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',280,'DESCRIPTION','varchar',2000,NULL,20,'N',NULL,'Y','N','N',NULL,'SYS_TABLE.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',3,3,NULL,'Y','备注',NULL,NULL,NULL,NULL),(335,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,21,'N','N','N','N','N',NULL,'SYS_TABLE_CATEGORY.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(336,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(337,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(338,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(339,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(340,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(341,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,21,'N',NULL,'N','N','N',NULL,'SYS_TABLE_CATEGORY.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(342,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属子系统','1111111111',80,'SYS_SUBSYSTEM_ID','int',NULL,NULL,21,'N',NULL,'Y','N','Y',NULL,'SYS_TABLE_CATEGORY.SYS_SUBSYSTEM_ID','Y','fk',19,302,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属子系统',NULL,NULL,NULL,NULL),(343,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','名称','1111111111',90,'NAME','varchar',255,NULL,21,'Y','Y','N','N','Y',NULL,'SYS_TABLE_CATEGORY.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','名称',NULL,NULL,NULL,NULL),(344,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','序号','1111111111',100,'ORDERNO','int',NULL,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','序号',NULL,NULL,NULL,NULL),(345,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','icon图标','1111111111',110,'ICON','varchar',255,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.ICON','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','icon图标',NULL,NULL,NULL,NULL),(346,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','网页连接','1111111111',120,'URL','varchar',255,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.URL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','网页连接',NULL,NULL,NULL,NULL),(347,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',130,'DESCRIPTION','varchar',255,NULL,21,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CATEGORY.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(348,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,22,'Y','Y','N','N','N',NULL,'SYS_TABLE_CMD.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(349,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(350,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(351,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(352,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(353,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(354,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,22,'N',NULL,'N','N','N',NULL,'SYS_TABLE_CMD.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(355,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属表单','1111111111',80,'SYS_TABLE_ID','int',NULL,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属表单',NULL,NULL,NULL,NULL),(356,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','按钮类型(1:系统按钮)','1111111111',90,'ACTION_TYPE','char',1,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.ACTION_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','按钮类型(1:系统按钮)',NULL,NULL,NULL,NULL),(357,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','按钮(A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废,I:导入,E:导出)','1111111111',100,'ACTION','char',1,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.ACTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','按钮(A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废,I:导入,E:导出)',NULL,NULL,NULL,NULL),(358,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','按钮名称','1111111111',110,'ACTION_NAME','varchar',255,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.ACTION_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','按钮名称',NULL,NULL,NULL,NULL),(359,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','事件前后(begin:开始,end:结束)','1111111111',120,'EVENT','varchar',255,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.EVENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','事件前后(begin:开始,end:结束)',NULL,NULL,NULL,NULL),(360,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','执行操作(存储过程/action动作)','1111111111',130,'CONTENT','varchar',255,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.CONTENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','执行操作(存储过程/action动作)',NULL,NULL,NULL,NULL),(361,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','动作类型','1111111111',140,'CONTENT_TYPE','varchar',255,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.CONTENT_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','动作类型',NULL,NULL,NULL,NULL),(362,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','序号','1111111111',150,'ORDERNO','int',NULL,NULL,22,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_CMD.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','序号',NULL,NULL,NULL,NULL),(363,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,23,'Y','Y','N','N','N',NULL,'SYS_TABLE_REF.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(364,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(365,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(366,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(367,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(368,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(369,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,23,'N',NULL,'N','N','N',NULL,'SYS_TABLE_REF.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(370,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主表','1111111111',80,'SYS_TABLE_ID','int',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','主表',NULL,NULL,NULL,NULL),(371,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','序号','1111111111',90,'ORDERNO','int',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','序号',NULL,NULL,NULL,NULL),(372,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示描述','1111111111',100,'DISPLAY_NAME','varchar',255,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示描述',NULL,NULL,NULL,NULL),(373,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联表','1111111111',110,'REF_TABLE_ID','int',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.REF_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联表',NULL,NULL,NULL,NULL),(374,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联字段','1111111111',120,'REF_COLUMN_ID','int',NULL,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.REF_COLUMN_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联字段',NULL,NULL,NULL,NULL),(375,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','过滤条件','1111111111',130,'FILTER','varchar',255,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.FILTER','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','过滤条件',NULL,NULL,NULL,NULL),(376,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联方式(1 : 1对1, n: 1对n )','1111111111',140,'ASSOCTYPE','char',1,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.ASSOCTYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','关联方式(1 : 1对1, n: 1对n )',NULL,NULL,NULL,NULL),(377,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','编辑方式(Y:标准(新增和修改行时都可在内嵌窗口编辑),\r\nN:无(无内嵌编辑窗口),NP:非内嵌，允许弹出,NS:非内嵌，禁止弹出,A:仅显示新增字段，修改直接修改)','1111111111',150,'EDIT_TYPE','char',2,NULL,23,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_REF.EDIT_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','编辑方式(Y:标准(新增和修改行时都可在内嵌窗口编辑),\r\nN:无(无内嵌编辑窗口),NP:非内嵌，允许弹出,NS:非内嵌，禁止弹出,A:仅显示新增字段，修改直接修改)',NULL,NULL,NULL,NULL),(378,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,24,'Y','Y','N','N','N',NULL,'SYS_TABLE_SQL.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(379,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(380,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(381,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(382,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(383,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(384,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,24,'N',NULL,'N','N','N',NULL,'SYS_TABLE_SQL.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(385,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属表单','1111111111',80,'SYS_TABLE_ID','int',NULL,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属表单',NULL,NULL,NULL,NULL),(386,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','表单sql','1111111111',90,'SQL','varchar',5000,NULL,24,'N',NULL,'Y','N','N',NULL,'SYS_TABLE_SQL.SQL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','表单sql',NULL,NULL,NULL,NULL),(387,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,25,'Y','Y','N','N','N',NULL,'SYS_USER.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(388,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(389,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(390,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(391,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(392,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(393,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,25,'N',NULL,'N','N','N',NULL,'SYS_USER.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(394,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','真实名称','1111111111',80,'TRUE_NAME','varchar',255,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.TRUE_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','真实名称',NULL,NULL,NULL,NULL),(395,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','用户名称','1111111111',90,'USERNAME','varchar',255,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.USERNAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','用户名称',NULL,NULL,NULL,NULL),(396,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','密码','1111111111',100,'PASSWORD','varchar',255,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.PASSWORD','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','密码',NULL,NULL,NULL,NULL),(397,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','手机号','1111111111',110,'PHONE','varchar',20,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.PHONE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','手机号',NULL,NULL,NULL,NULL),(398,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','邮箱','1111111111',120,'EMAIL','varchar',255,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.EMAIL','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','邮箱',NULL,NULL,NULL,NULL),(399,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','语言','1111111111',130,'LANGUAGE','varchar',255,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.LANGUAGE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','语言',NULL,NULL,NULL,NULL),(400,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否管理员','1111111111',140,'IS_ADMIN','char',2,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.IS_ADMIN','Y','byPage',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否管理员',NULL,NULL,NULL,NULL),(401,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','字段访问级别','1111111111',150,'SGRADE','int',NULL,NULL,25,'N',NULL,'Y','N','N',NULL,'SYS_USER.SGRADE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','字段访问级别',NULL,NULL,NULL,NULL),(402,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,26,'Y','Y','N','N','N',NULL,'SYS_USER_ENV.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(403,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(404,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(405,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(406,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(407,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(408,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,26,'N',NULL,'N','N','N',NULL,'SYS_USER_ENV.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(409,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','变量名称','1111111111',80,'NAME','varchar',255,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','变量名称',NULL,NULL,NULL,NULL),(410,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','值来源','1111111111',90,'VALUE','varchar',255,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','值来源',NULL,NULL,NULL,NULL),(411,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','备注','1111111111',100,'DESCRIPTION','varchar',255,NULL,26,'N',NULL,'Y','N','N',NULL,'SYS_USER_ENV.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','备注',NULL,NULL,NULL,NULL),(412,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,27,'Y','Y','N','N','N',NULL,'SYS_USER_GROUPS.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(413,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(414,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(415,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(416,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(417,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(418,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,27,'N',NULL,'N','N','N',NULL,'SYS_USER_GROUPS.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(419,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','用户','1111111111',80,'SYS_USER_ID','int',NULL,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.SYS_USER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','用户',NULL,NULL,NULL,NULL),(420,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','权限组','1111111111',90,'SYS_DIRECTORY_ID','int',NULL,NULL,27,'N',NULL,'Y','N','N',NULL,'SYS_USER_GROUPS.SYS_DIRECTORY_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','权限组',NULL,NULL,NULL,NULL),(421,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','主键','0000000000',10,'ID','int',NULL,NULL,28,'Y','Y','N','N','N',NULL,'SYS_USER_MESSAGE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(422,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,28,'N',NULL,'Y','N','N',NULL,'SYS_USER_MESSAGE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(423,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,28,'N',NULL,'Y','N','N',NULL,'SYS_USER_MESSAGE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(424,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(425,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,28,'N',NULL,'Y','N','N',NULL,'SYS_USER_MESSAGE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(426,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,'CURRENT_TIMESTAMP',NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(427,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(428,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','消息ID','1111111111',80,'MESSAGE_ID','int',NULL,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.MESSAGE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','消息ID',NULL,NULL,NULL,NULL),(429,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','用户ID','1111111111',90,'USER_ID','int',NULL,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.USER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','用户ID',NULL,NULL,NULL,NULL),(430,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否已读 Y/N','1111111111',100,'IS_READ','char',1,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.IS_READ','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否已读 Y/N',NULL,NULL,NULL,NULL),(431,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','读取时间','1111111111',110,'READ_TIME','datetime',NULL,NULL,28,'N',NULL,'Y','N','N',NULL,'SYS_USER_MESSAGE.READ_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','读取时间',NULL,NULL,NULL,NULL),(432,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否星标 Y/N','1111111111',120,'IS_STARRED','char',1,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.IS_STARRED','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否星标 Y/N',NULL,NULL,NULL,NULL),(433,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否归档 Y/N','1111111111',130,'IS_ARCHIVED','char',1,NULL,28,'N',NULL,'N','N','N',NULL,'SYS_USER_MESSAGE.IS_ARCHIVED','Y','byPage',NULL,NULL,NULL,NULL,NULL,'N',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','是否归档 Y/N',NULL,NULL,NULL,NULL),(434,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','删除时间（软删除）','1111111111',140,'DELETED_AT','datetime',NULL,NULL,28,'N',NULL,'Y','N','N',NULL,'SYS_USER_MESSAGE.DELETED_AT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','删除时间（软删除）',NULL,NULL,NULL,NULL),(435,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,29,'Y','Y','N','N','N',NULL,'SYS_USER_SESSION.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(436,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','用户ID','1111111111',20,'USER_ID','int',NULL,NULL,29,'N',NULL,'N','N','N',NULL,'SYS_USER_SESSION.USER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','用户ID',NULL,NULL,NULL,NULL),(437,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','公司ID','1111111111',30,'COMPANY_ID','int',NULL,NULL,29,'N',NULL,'N','N','N',NULL,'SYS_USER_SESSION.COMPANY_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','公司ID',NULL,NULL,NULL,NULL),(438,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','Access Token','1111111111',40,'TOKEN','varchar',500,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.TOKEN','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','Access Token',NULL,NULL,NULL,NULL),(439,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','Refresh Token','1111111111',50,'REFRESH_TOKEN','varchar',500,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.REFRESH_TOKEN','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','Refresh Token',NULL,NULL,NULL,NULL),(440,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','客户端类型','1111111111',60,'CLIENT_TYPE','varchar',20,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.CLIENT_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','客户端类型',NULL,NULL,NULL,NULL),(441,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','设备ID','1111111111',70,'DEVICE_ID','varchar',255,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.DEVICE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','设备ID',NULL,NULL,NULL,NULL),(442,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','设备名称','1111111111',80,'DEVICE_NAME','varchar',255,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.DEVICE_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','设备名称',NULL,NULL,NULL,NULL),(443,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','IP地址','1111111111',90,'IP_ADDRESS','varchar',50,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.IP_ADDRESS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','IP地址',NULL,NULL,NULL,NULL),(444,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','User Agent','1111111111',100,'USER_AGENT','varchar',500,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.USER_AGENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','User Agent',NULL,NULL,NULL,NULL),(445,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','登录时间','1111111111',110,'LOGIN_TIME','datetime',NULL,NULL,29,'N',NULL,'N','N','N',NULL,'SYS_USER_SESSION.LOGIN_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','登录时间',NULL,NULL,NULL,NULL),(446,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','最后活跃时间','1111111111',120,'LAST_ACTIVE_TIME','datetime',NULL,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.LAST_ACTIVE_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','最后活跃时间',NULL,NULL,NULL,NULL),(447,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','过期时间','1111111111',130,'EXPIRE_TIME','datetime',NULL,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.EXPIRE_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','过期时间',NULL,NULL,NULL,NULL),(448,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','是否有效(Y/N)','0000000000',1050,'IS_ACTIVE','char',1,NULL,29,'N',NULL,'Y','N','N',NULL,'SYS_USER_SESSION.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y/N)',NULL,NULL,NULL,NULL),(449,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,30,'Y','Y','N','N','N',NULL,'WF_DEFINITION.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(450,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(451,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(452,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(453,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(454,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(455,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,30,'N',NULL,'N','N','N',NULL,'WF_DEFINITION.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(456,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流程名称','1111111111',80,'NAME','varchar',80,NULL,30,'N',NULL,'N','N','N',NULL,'WF_DEFINITION.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','流程名称',NULL,NULL,NULL,NULL),(457,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',90,'DISPLAY_NAME','varchar',255,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(458,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','版本号','1111111111',100,'VERSION','int',NULL,NULL,30,'N',NULL,'N','N','N',NULL,'WF_DEFINITION.VERSION','Y','byPage',NULL,NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','版本号',NULL,NULL,NULL,NULL),(459,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','状态(draft:草稿,published:已发布,archived:已归档)','1111111111',110,'STATUS','varchar',20,NULL,30,'N',NULL,'N','N','N',NULL,'WF_DEFINITION.STATUS','Y','byPage',NULL,NULL,NULL,NULL,NULL,'draft',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','状态(draft:草稿,published:已发布,archived:已归档)',NULL,NULL,NULL,NULL),(460,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联的业务表','1111111111',120,'SYS_TABLE_ID','int',NULL,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联的业务表',NULL,NULL,NULL,NULL),(461,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','描述','1111111111',130,'DESCRIPTION','varchar',2000,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.DESCRIPTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','描述',NULL,NULL,NULL,NULL),(462,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','JSON配置','1111111111',140,'CONFIG','text',65535,NULL,30,'N',NULL,'Y','N','N',NULL,'WF_DEFINITION.CONFIG','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','JSON配置',NULL,NULL,NULL,NULL),(463,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,31,'Y','Y','N','N','N',NULL,'WF_INSTANCE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(464,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(465,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(466,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(467,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(468,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(469,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,31,'N',NULL,'N','N','N',NULL,'WF_INSTANCE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(470,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流程定义ID','1111111111',80,'WF_DEFINITION_ID','int',NULL,NULL,31,'N',NULL,'N','N','N',NULL,'WF_INSTANCE.WF_DEFINITION_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','流程定义ID',NULL,NULL,NULL,NULL),(471,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','关联的业务表','1111111111',90,'SYS_TABLE_ID','int',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.SYS_TABLE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','关联的业务表',NULL,NULL,NULL,NULL),(472,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','业务数据ID','1111111111',100,'BUSINESS_ID','int',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.BUSINESS_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','业务数据ID',NULL,NULL,NULL,NULL),(473,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','状态(running:运行中,completed:已完成,terminated:已终止,suspended:已挂起)','1111111111',110,'STATUS','varchar',20,NULL,31,'N',NULL,'N','N','N',NULL,'WF_INSTANCE.STATUS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','状态(running:运行中,completed:已完成,terminated:已终止,suspended:已挂起)',NULL,NULL,NULL,NULL),(474,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','当前节点ID','1111111111',120,'CURRENT_NODE_ID','int',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.CURRENT_NODE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','当前节点ID',NULL,NULL,NULL,NULL),(475,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','发起人','1111111111',130,'START_USER_ID','int',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.START_USER_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','发起人',NULL,NULL,NULL,NULL),(476,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','开始时间','1111111111',140,'START_TIME','datetime',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.START_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','开始时间',NULL,NULL,NULL,NULL),(477,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','结束时间','1111111111',150,'END_TIME','datetime',NULL,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.END_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','结束时间',NULL,NULL,NULL,NULL),(478,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流程变量(JSON)','1111111111',160,'VARIABLES','text',65535,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.VARIABLES','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','流程变量(JSON)',NULL,NULL,NULL,NULL),(479,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流程标题','1111111111',170,'TITLE','varchar',255,NULL,31,'N',NULL,'Y','N','N',NULL,'WF_INSTANCE.TITLE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','流程标题',NULL,NULL,NULL,NULL),(480,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,32,'Y','Y','N','N','N',NULL,'WF_NODE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(481,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(482,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(483,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(484,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(485,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(486,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,32,'N',NULL,'N','N','N',NULL,'WF_NODE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(487,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属流程定义','1111111111',80,'WF_DEFINITION_ID','int',NULL,NULL,32,'N',NULL,'N','N','N',NULL,'WF_NODE.WF_DEFINITION_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属流程定义',NULL,NULL,NULL,NULL),(488,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','节点名称','1111111111',90,'NAME','varchar',80,NULL,32,'N',NULL,'N','N','N',NULL,'WF_NODE.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','节点名称',NULL,NULL,NULL,NULL),(489,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','显示名称','1111111111',100,'DISPLAY_NAME','varchar',255,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.DISPLAY_NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','显示名称',NULL,NULL,NULL,NULL),(490,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','节点类型(start:开始,end:结束,user:用户任务,auto:自动任务,gateway:网关)','1111111111',110,'NODE_TYPE','varchar',20,NULL,32,'N',NULL,'N','N','N',NULL,'WF_NODE.NODE_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','节点类型(start:开始,end:结束,user:用户任务,auto:自动任务,gateway:网关)',NULL,NULL,NULL,NULL),(491,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','分配类型(user:指定用户,starter:发起人,role:角色,expression:表达式)','1111111111',120,'ASSIGN_TYPE','varchar',20,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.ASSIGN_TYPE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','分配类型(user:指定用户,starter:发起人,role:角色,expression:表达式)',NULL,NULL,NULL,NULL),(492,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','分配值','1111111111',130,'ASSIGN_VALUE','varchar',500,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.ASSIGN_VALUE','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','分配值',NULL,NULL,NULL,NULL),(493,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','自动任务关联的动作ID','1111111111',140,'ACTION_ID','int',NULL,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.ACTION_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','自动任务关联的动作ID',NULL,NULL,NULL,NULL),(494,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','JSON配置','1111111111',150,'CONFIG','text',65535,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.CONFIG','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','JSON配置',NULL,NULL,NULL,NULL),(495,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','节点X坐标','1111111111',160,'POS_X','int',NULL,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.POS_X','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','节点X坐标',NULL,NULL,NULL,NULL),(496,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','节点Y坐标','1111111111',170,'POS_Y','int',NULL,NULL,32,'N',NULL,'Y','N','N',NULL,'WF_NODE.POS_Y','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','节点Y坐标',NULL,NULL,NULL,NULL),(497,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,33,'Y','Y','N','N','N',NULL,'WF_TASK.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(498,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(499,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(500,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(501,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(502,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(503,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,33,'N',NULL,'N','N','N',NULL,'WF_TASK.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(504,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流程实例ID','1111111111',80,'WF_INSTANCE_ID','int',NULL,NULL,33,'N',NULL,'N','N','N',NULL,'WF_TASK.WF_INSTANCE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','流程实例ID',NULL,NULL,NULL,NULL),(505,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流程节点ID','1111111111',90,'WF_NODE_ID','int',NULL,NULL,33,'N',NULL,'N','N','N',NULL,'WF_TASK.WF_NODE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','流程节点ID',NULL,NULL,NULL,NULL),(506,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','任务执行人','1111111111',100,'ASSIGNEE_ID','int',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.ASSIGNEE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','任务执行人',NULL,NULL,NULL,NULL),(507,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','状态(pending:待处理,completed:已完成,rejected:已拒绝,transferred:已转交)','1111111111',110,'STATUS','varchar',20,NULL,33,'N',NULL,'N','N','N',NULL,'WF_TASK.STATUS','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','状态(pending:待处理,completed:已完成,rejected:已拒绝,transferred:已转交)',NULL,NULL,NULL,NULL),(508,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','操作(approve:同意,reject:拒绝,transfer:转交)','1111111111',120,'ACTION','varchar',20,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.ACTION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','操作(approve:同意,reject:拒绝,transfer:转交)',NULL,NULL,NULL,NULL),(509,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','审批意见','1111111111',130,'COMMENT','varchar',2000,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.COMMENT','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','审批意见',NULL,NULL,NULL,NULL),(510,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','签收时间','1111111111',140,'CLAIM_TIME','datetime',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.CLAIM_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','签收时间',NULL,NULL,NULL,NULL),(511,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','完成时间','1111111111',150,'COMPLETE_TIME','datetime',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.COMPLETE_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','完成时间',NULL,NULL,NULL,NULL),(512,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','截止时间','1111111111',160,'DUE_TIME','datetime',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.DUE_TIME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','截止时间',NULL,NULL,NULL,NULL),(513,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','优先级','1111111111',170,'PRIORITY','int',NULL,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.PRIORITY','Y','byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','优先级',NULL,NULL,NULL,NULL),(514,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','任务变量(JSON)','1111111111',180,'VARIABLES','text',65535,NULL,33,'N',NULL,'Y','N','N',NULL,'WF_TASK.VARIABLES','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y','任务变量(JSON)',NULL,NULL,NULL,NULL),(515,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','ID','0000000000',10,'ID','int',NULL,NULL,34,'Y','Y','N','N','N',NULL,'WF_TRANSITION.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','',NULL,NULL,NULL,NULL),(516,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属公司','0000000000',20,'SYS_COMPANY_ID','int',NULL,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.SYS_COMPANY_ID','N','fk',4,91,'noAction',NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(517,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建人','0000000000',1010,'CREATE_BY','varchar',80,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(518,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','创建时间','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(519,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新人','0000000000',1030,'UPDATE_BY','varchar',80,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','更新人',NULL,NULL,NULL,NULL),(520,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','更新时间','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','更新时间',NULL,NULL,NULL,NULL),(521,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','可用','0000000000',1050,'IS_ACTIVE','char',1,NULL,34,'N',NULL,'N','N','N',NULL,'WF_TRANSITION.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,'1','Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','是否有效(Y:可用,N:不可用)',NULL,NULL,NULL,NULL),(522,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','所属流程定义','1111111111',80,'WF_DEFINITION_ID','int',NULL,NULL,34,'N',NULL,'N','N','N',NULL,'WF_TRANSITION.WF_DEFINITION_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属流程定义',NULL,NULL,NULL,NULL),(523,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','起始节点','1111111111',90,'FROM_NODE_ID','int',NULL,NULL,34,'N',NULL,'N','N','N',NULL,'WF_TRANSITION.FROM_NODE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','起始节点',NULL,NULL,NULL,NULL),(524,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','目标节点','1111111111',100,'TO_NODE_ID','int',NULL,NULL,34,'N',NULL,'N','N','N',NULL,'WF_TRANSITION.TO_NODE_ID','Y','fk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','目标节点',NULL,NULL,NULL,NULL),(525,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流转名称','1111111111',110,'NAME','varchar',80,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.NAME','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','流转名称',NULL,NULL,NULL,NULL),(526,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','流转条件表达式','1111111111',120,'CONDITION','varchar',500,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.CONDITION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','流转条件表达式',NULL,NULL,NULL,NULL),(527,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','优先级顺序','1111111111',130,'ORDERNO','int',NULL,NULL,34,'N',NULL,'Y','N','N',NULL,'WF_TRANSITION.ORDERNO','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','优先级顺序',NULL,NULL,NULL,NULL),(528,1,'admin','2026-01-21 03:34:42','admin','2026-01-21 11:54:01','Y','域名','1111111110',100,'DOMAIN','varchar',255,NULL,4,'N','N','Y','N','Y',NULL,'SYS_COMPANY.DOMAIN','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'N',NULL,NULL,NULL,NULL,NULL),(665,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.ID','0000000000',1,'ID','int',NULL,NULL,35,'Y','Y','N','N','N',NULL,'LIVE_DOMAIN.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(666,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,35,'N',NULL,'Y','N','N',NULL,'LIVE_DOMAIN.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(667,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.(LIVE_DOMAIN.ID+100)','0010011001',1000,'(LIVE_DOMAIN.ID+100)','varchar',30,NULL,35,'N',NULL,'Y','N','N',NULL,'LIVE_DOMAIN.(LIVE_DOMAIN.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(668,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.CREATE_BY','0000000000',1010,'CREATE_BY','varchar',80,NULL,35,'N',NULL,'Y','N','N',NULL,'LIVE_DOMAIN.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(669,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.UPDATE_BY','0000000000',1030,'UPDATE_BY','varchar',80,NULL,35,'N',NULL,'Y','N','N',NULL,'LIVE_DOMAIN.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(670,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.CREATE_TIME','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,35,'N',NULL,'Y','N','N',NULL,'LIVE_DOMAIN.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(671,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.UPDATE_TIME','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,35,'N',NULL,'Y','N','N',NULL,'LIVE_DOMAIN.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(672,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN.IS_ACTIVE','0000000000',1050,'IS_ACTIVE','char',1,NULL,35,'N',NULL,'N','N','N',NULL,'LIVE_DOMAIN.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(673,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.ID','0000000000',1,'ID','int',NULL,NULL,37,'Y','Y','N','N','N',NULL,'CLOUD_ITEM.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(674,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,37,'N',NULL,'Y','N','N',NULL,'CLOUD_ITEM.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(675,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.(CLOUD_ITEM.ID+100)','0010011001',1000,'(CLOUD_ITEM.ID+100)','varchar',30,NULL,37,'N',NULL,'Y','N','N',NULL,'CLOUD_ITEM.(CLOUD_ITEM.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(676,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.CREATE_BY','0000000000',1010,'CREATE_BY','varchar',80,NULL,37,'N',NULL,'Y','N','N',NULL,'CLOUD_ITEM.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(677,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.UPDATE_BY','0000000000',1030,'UPDATE_BY','varchar',80,NULL,37,'N',NULL,'Y','N','N',NULL,'CLOUD_ITEM.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(678,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.CREATE_TIME','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,37,'N',NULL,'Y','N','N',NULL,'CLOUD_ITEM.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(679,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.UPDATE_TIME','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,37,'N',NULL,'Y','N','N',NULL,'CLOUD_ITEM.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(680,1,'admin','2026-01-31 00:18:56',NULL,NULL,'Y','CLOUD_ITEM.IS_ACTIVE','0000000000',1050,'IS_ACTIVE','char',1,NULL,37,'N',NULL,'N','N','N',NULL,'CLOUD_ITEM.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(681,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.ID','0000000000',1,'ID','int',NULL,NULL,38,'Y','Y','N','N','N',NULL,'LIVE_STREAM.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(682,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,38,'N',NULL,'Y','N','N',NULL,'LIVE_STREAM.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(683,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.(LIVE_STREAM.ID+100)','0010011001',1000,'(LIVE_STREAM.ID+100)','varchar',30,NULL,38,'N',NULL,'Y','N','N',NULL,'LIVE_STREAM.(LIVE_STREAM.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(684,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.CREATE_BY','0000000000',1010,'CREATE_BY','varchar',80,NULL,38,'N',NULL,'Y','N','N',NULL,'LIVE_STREAM.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(685,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.UPDATE_BY','0000000000',1030,'UPDATE_BY','varchar',80,NULL,38,'N',NULL,'Y','N','N',NULL,'LIVE_STREAM.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(686,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.CREATE_TIME','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,38,'N',NULL,'Y','N','N',NULL,'LIVE_STREAM.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(687,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.UPDATE_TIME','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,38,'N',NULL,'Y','N','N',NULL,'LIVE_STREAM.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(688,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM.IS_ACTIVE','0000000000',1050,'IS_ACTIVE','char',1,NULL,38,'N',NULL,'N','N','N',NULL,'LIVE_STREAM.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(689,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.ID','0000000000',1,'ID','int',NULL,NULL,39,'Y','Y','N','N','N',NULL,'LIVE_PULL.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(690,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,39,'N',NULL,'Y','N','N',NULL,'LIVE_PULL.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(691,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.(LIVE_PULL.ID+100)','0010011001',1000,'(LIVE_PULL.ID+100)','varchar',30,NULL,39,'N',NULL,'Y','N','N',NULL,'LIVE_PULL.(LIVE_PULL.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(692,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.CREATE_BY','0000000000',1010,'CREATE_BY','varchar',80,NULL,39,'N',NULL,'Y','N','N',NULL,'LIVE_PULL.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(693,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.UPDATE_BY','0000000000',1030,'UPDATE_BY','varchar',80,NULL,39,'N',NULL,'Y','N','N',NULL,'LIVE_PULL.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(694,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.CREATE_TIME','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,39,'N',NULL,'Y','N','N',NULL,'LIVE_PULL.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(695,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.UPDATE_TIME','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,39,'N',NULL,'Y','N','N',NULL,'LIVE_PULL.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(696,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL.IS_ACTIVE','0000000000',1050,'IS_ACTIVE','char',1,NULL,39,'N',NULL,'N','N','N',NULL,'LIVE_PULL.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(697,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.ID','0000000000',1,'ID','int',NULL,NULL,40,'Y','Y','N','N','N',NULL,'LIVE_CUT.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(698,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,40,'N',NULL,'Y','N','N',NULL,'LIVE_CUT.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(699,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.(LIVE_CUT.ID+100)','0010011001',1000,'(LIVE_CUT.ID+100)','varchar',30,NULL,40,'N',NULL,'Y','N','N',NULL,'LIVE_CUT.(LIVE_CUT.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(700,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.CREATE_BY','0000000000',1010,'CREATE_BY','varchar',80,NULL,40,'N',NULL,'Y','N','N',NULL,'LIVE_CUT.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(701,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.UPDATE_BY','0000000000',1030,'UPDATE_BY','varchar',80,NULL,40,'N',NULL,'Y','N','N',NULL,'LIVE_CUT.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(702,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.CREATE_TIME','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,40,'N',NULL,'Y','N','N',NULL,'LIVE_CUT.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(703,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.UPDATE_TIME','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,40,'N',NULL,'Y','N','N',NULL,'LIVE_CUT.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(704,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT.IS_ACTIVE','0000000000',1050,'IS_ACTIVE','char',1,NULL,40,'N',NULL,'N','N','N',NULL,'LIVE_CUT.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(705,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.ID','0000000000',1,'ID','int',NULL,NULL,41,'Y','Y','N','N','N',NULL,'LIVE_RECODE.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(706,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,41,'N',NULL,'Y','N','N',NULL,'LIVE_RECODE.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(707,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.(LIVE_RECODE.ID+100)','0010011001',1000,'(LIVE_RECODE.ID+100)','varchar',30,NULL,41,'N',NULL,'Y','N','N',NULL,'LIVE_RECODE.(LIVE_RECODE.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(708,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.CREATE_BY','0000000000',1010,'CREATE_BY','varchar',80,NULL,41,'N',NULL,'Y','N','N',NULL,'LIVE_RECODE.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(709,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.UPDATE_BY','0000000000',1030,'UPDATE_BY','varchar',80,NULL,41,'N',NULL,'Y','N','N',NULL,'LIVE_RECODE.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(710,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.CREATE_TIME','0000000000',1020,'CREATE_TIME','datetime',NULL,NULL,41,'N',NULL,'Y','N','N',NULL,'LIVE_RECODE.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(711,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.UPDATE_TIME','0000000000',1040,'UPDATE_TIME','datetime',NULL,NULL,41,'N',NULL,'Y','N','N',NULL,'LIVE_RECODE.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(712,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE.IS_ACTIVE','0000000000',1050,'IS_ACTIVE','char',1,NULL,41,'N',NULL,'N','N','N',NULL,'LIVE_RECODE.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(713,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','ID',NULL,10,'ID','bigint',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(714,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','直播间名称',NULL,20,'ROOM_NAME','varchar',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(715,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','直播间类型',NULL,30,'ROOM_TYPE','varchar',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'6',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(716,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','播出形式',NULL,40,'BROADCAST_FORMAT','varchar',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'7',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(717,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','直播间阶段',NULL,50,'ROOM_STAGE','varchar',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'8',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(718,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','显示方式',NULL,60,'DISPLAY_MODE','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'9',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(719,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','开始时间',NULL,70,'START_TIME','datetime',NULL,NULL,54,'N',NULL,'Y','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(720,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','结束时间',NULL,80,'END_TIME','datetime',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(721,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','直播间封面',NULL,90,'COVER_IMAGE','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'image',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(722,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','观看方式',NULL,100,'VIEWING_METHOD','varchar',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'10',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(723,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','观看密码',NULL,110,'VIEWING_PASSWORD','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'password',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(724,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','观看价格',NULL,120,'VIEWING_PRICE','decimal',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(725,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','回放方式',NULL,130,'PLAYBACK_METHOD','varchar',NULL,NULL,54,'N',NULL,'N','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'11',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(726,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','回放有效期',NULL,140,'PLAYBACK_VALIDITY','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'12',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(727,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','回放开始时间',NULL,150,'PLAYBACK_START_TIME','time',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'time',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(728,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','回放结束时间',NULL,160,'PLAYBACK_END_TIME','time',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'time',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(729,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','流名称',NULL,170,'STREAM_NAME','varchar',NULL,NULL,54,'N',NULL,'Y','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(730,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','推流地址',NULL,180,'PUSH_URL','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(731,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','播放地址',NULL,190,'PLAY_URL','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(732,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','状态',NULL,200,'STATUS','varchar',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,'13',NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(733,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','观看人数',NULL,210,'VIEWER_COUNT','int',NULL,NULL,54,'N',NULL,'N','N','N',NULL,NULL,NULL,'ignore',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(734,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','峰值观看人数',NULL,220,'PEAK_VIEWER_COUNT','int',NULL,NULL,54,'N',NULL,'N','N','N',NULL,NULL,NULL,'ignore',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(735,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','直播时长(秒)',NULL,230,'DURATION','int',NULL,NULL,54,'N',NULL,'N','N','N',NULL,NULL,NULL,'ignore',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(736,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','直播间描述',NULL,240,'DESCRIPTION','text',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(737,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','扩展属性',NULL,250,'PROPS','text',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'json',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(738,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','公司ID',NULL,255,'SYS_COMPANY_ID','bigint',NULL,NULL,54,'N',NULL,'N','N','N',NULL,NULL,NULL,'ignore',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(739,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','创建人',NULL,260,'CREATE_BY','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(740,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','创建时间',NULL,270,'CREATE_TIME','datetime',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(741,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','更新人',NULL,280,'UPDATE_BY','varchar',NULL,NULL,54,'N',NULL,'Y','N','N',NULL,NULL,NULL,'operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(742,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','更新时间',NULL,290,'UPDATE_TIME','datetime',NULL,NULL,54,'N',NULL,'N','N','N',NULL,NULL,NULL,'sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(743,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','是否有效',NULL,300,'IS_ACTIVE','char',NULL,NULL,54,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(744,1,NULL,NULL,NULL,NULL,'Y','ID','1111111111',1,'ID','int',NULL,NULL,11,'N',NULL,'N','N','N',NULL,NULL,NULL,'pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(745,1,NULL,NULL,NULL,NULL,'Y','权限组名称','1111111111',10,'NAME','varchar',NULL,NULL,11,'N',NULL,'N','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,'^.{2,50}$','权限组名称长度必须在2-50个字符之间',NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(746,1,NULL,NULL,NULL,NULL,'Y','描述','1111111111',20,'DESCRIPTION','varchar',NULL,NULL,11,'N',NULL,'Y','N','N',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'textarea',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(747,1,NULL,NULL,NULL,NULL,'Y','安全等级','1111111111',30,'SGRADE','int',NULL,NULL,11,'N',NULL,'Y','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,'0',NULL,NULL,NULL,'number',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(748,1,NULL,NULL,NULL,NULL,'Y','创建人','1000000000',1001,'CREATE_BY','varchar',NULL,NULL,11,'N',NULL,'Y','N','N',NULL,NULL,NULL,'createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(749,1,NULL,NULL,NULL,NULL,'Y','创建时间','1000000000',1002,'CREATE_TIME','datetime',NULL,NULL,11,'N',NULL,'Y','N','N',NULL,NULL,NULL,'sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(750,1,NULL,NULL,NULL,NULL,'Y','更新人','1000000000',1003,'UPDATE_BY','varchar',NULL,NULL,11,'N',NULL,'Y','N','N',NULL,NULL,NULL,'operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(751,1,NULL,NULL,NULL,NULL,'Y','更新时间','1000000000',1004,'UPDATE_TIME','datetime',NULL,NULL,11,'N',NULL,'Y','N','N',NULL,NULL,NULL,'sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(752,1,NULL,NULL,NULL,NULL,'Y','状态','1111111111',1005,'IS_ACTIVE','varchar',NULL,NULL,11,'N',NULL,'Y','N','Y',NULL,NULL,NULL,'byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(753,1,NULL,NULL,NULL,NULL,'Y','ID','1111111111',1,'ID','int',NULL,NULL,10,'N',NULL,'N','N','N',NULL,NULL,NULL,'pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(754,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','ID','0000000000',1,'ID','int',NULL,NULL,55,'Y','Y','N','N','N',NULL,'SYS_COMPANY_CONF.ID','N','pk',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','主键',NULL,NULL,NULL,NULL),(755,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','SYS_COMPANY_ID','0000000000',2,'SYS_COMPANY_ID','int',NULL,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.SYS_COMPANY_ID','N','object',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'select',NULL,NULL,NULL,'Y','所属公司',NULL,NULL,NULL,NULL),(756,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','日志','0010011001',1000,'(SYS_COMPANY_CONF.ID+100)','varchar',30,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.(SYS_COMPANY_CONF.ID+100)','N','trigger',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'hr',NULL,NULL,NULL,'N','日志',NULL,NULL,NULL,NULL),(757,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','创建人','0010001001',1001,'CREATE_BY','varchar',80,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.CREATE_BY','N','createBy',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','创建人',NULL,NULL,NULL,NULL),(758,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','更新人','0010101101',1002,'UPDATE_BY','varchar',80,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.UPDATE_BY','N','operator',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y','修改人',NULL,NULL,NULL,NULL),(759,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','创建时间','0010001001',1003,'CREATE_TIME','datetime',NULL,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.CREATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','创建时间',NULL,NULL,NULL,NULL),(760,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','更新时间','0010101101',1004,'UPDATE_TIME','datetime',NULL,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.UPDATE_TIME','N','sysdate',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'datetime',NULL,NULL,NULL,'Y','修改时间',NULL,NULL,NULL,NULL),(761,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','状态','0011101000',10000,'IS_ACTIVE','char',1,NULL,55,'N',NULL,'N','N','N',NULL,'SYS_COMPANY_CONF.IS_ACTIVE','Y','select',NULL,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,'check',NULL,NULL,NULL,'Y','可用',NULL,NULL,NULL,NULL),(762,1,'admin','2026-03-04 02:24:17','admin','2026-03-13 03:54:50','Y','secretID','1111111111',10,'SECRET_ID','varchar',255,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.SECRET_ID','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(763,1,'admin','2026-03-13 03:54:29',NULL,NULL,'Y','secretKey','1111111111',20,'SECRET_KEY','varchar',255,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.SECRET_KEY','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL),(765,1,'admin','2026-03-13 03:57:23',NULL,NULL,'Y','区域','1111111111',30,'REGION','varchar',255,NULL,55,'N',NULL,'Y','N','N',NULL,'SYS_COMPANY_CONF.SREGION','Y','byPage',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'text',NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL);
/*!40000 ALTER TABLE `sys_column` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_company`
--

DROP TABLE IF EXISTS `sys_company`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_company` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '公司名称',
  `DOMAIN` varchar(255) DEFAULT NULL COMMENT '域名',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=37 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='模板表单';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_company`
--

LOCK TABLES `sys_company` WRITE;
/*!40000 ALTER TABLE `sys_company` DISABLE KEYS */;
INSERT INTO `sys_company` VALUES (1,1,'admin','2026-01-12 20:52:13','1','2026-03-13 02:47:31','Y','先桦科技','pan.xh-tec.cn');
/*!40000 ALTER TABLE `sys_company` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_company_conf`
--

DROP TABLE IF EXISTS `sys_company_conf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_company_conf` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SECRET_ID` varchar(255) DEFAULT NULL COMMENT 'secretID',
  `SECRET_KEY` varchar(255) DEFAULT NULL COMMENT 'secretKey',
  `REGION` varchar(255) DEFAULT NULL COMMENT 'region',
  `STORAGE_TYPE` varchar(20) DEFAULT 'local' COMMENT '存储类型: local, aliyunOSS, tencentCOS',
  `LOCAL_BASE_PATH` varchar(500) DEFAULT 'uploads' COMMENT '本地存储基础路径',
  `LOCAL_BASE_URL` varchar(500) DEFAULT '/files' COMMENT '本地存储基础URL',
  `ALIYUN_OSS_ENDPOINT` varchar(255) DEFAULT NULL COMMENT '阿里云OSS Endpoint',
  `ALIYUN_OSS_ACCESS_KEY_ID` varchar(255) DEFAULT NULL COMMENT '阿里云OSS AccessKeyID',
  `ALIYUN_OSS_ACCESS_KEY_SECRET` varchar(255) DEFAULT NULL COMMENT '阿里云OSS AccessKeySecret',
  `ALIYUN_OSS_BUCKET_NAME` varchar(255) DEFAULT NULL COMMENT '阿里云OSS Bucket名称',
  `ALIYUN_OSS_CDN_DOMAIN` varchar(500) DEFAULT NULL COMMENT '阿里云OSS CDN域名',
  `TENCENT_COS_BUCKET_URL` varchar(500) DEFAULT NULL COMMENT '腾讯云COS Bucket URL',
  `TENCENT_COS_SECRET_ID` varchar(255) DEFAULT NULL COMMENT '腾讯云COS SecretID',
  `TENCENT_COS_SECRET_KEY` varchar(255) DEFAULT NULL COMMENT '腾讯云COS SecretKey',
  `TENCENT_COS_BUCKET_NAME` varchar(255) DEFAULT NULL COMMENT '腾讯云COS Bucket名称',
  `TENCENT_COS_REGION` varchar(50) DEFAULT NULL COMMENT '腾讯云COS 区域',
  `TENCENT_COS_CDN_DOMAIN` varchar(500) DEFAULT NULL COMMENT '腾讯云COS CDN域名',
  `TENCENT_CLOUD_SECRET_ID` varchar(255) DEFAULT NULL COMMENT '腾讯云SecretID',
  `TENCENT_CLOUD_SECRET_KEY` varchar(255) DEFAULT NULL COMMENT '腾讯云SecretKey',
  `TENCENT_CLOUD_REGION` varchar(255) DEFAULT NULL COMMENT '腾讯云区域',
  `TENCENT_CLOUD_CALLBACK_KEY` varchar(255) DEFAULT NULL COMMENT '腾讯云回调密钥',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='公司配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_company_conf`
--

LOCK TABLES `sys_company_conf` WRITE;
/*!40000 ALTER TABLE `sys_company_conf` DISABLE KEYS */;
INSERT INTO `sys_company_conf` VALUES (1,1,'1',NULL,'1',NULL,'Y','YOUR_SECRET_ID','YOUR_SECRET_KEY',NULL,'tencentCOS','uploads','/files',NULL,'YOUR_ALIYUN_ACCESS_KEY_ID','YOUR_ALIYUN_ACCESS_KEY_SECRET',NULL,NULL,'https://zhibo-1301212747.cos.ap-nanjing.myqcloud.com','YOUR_TENCENT_SECRET_ID','YOUR_TENCENT_SECRET_KEY','zhibo-1301212747','ap-nanjin',NULL,'YOUR_TENCENT_SECRET_ID','YOUR_TENCENT_SECRET_KEY','ap-shanghai',NULL);
/*!40000 ALTER TABLE `sys_company_conf` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_dict`
--

DROP TABLE IF EXISTS `sys_dict`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dict` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字典名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `TYPE` int DEFAULT '0' COMMENT '字段类型(0: String, 1: int)',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  `DEFAULT_VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '默认值',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='数据字典';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_dict`
--

LOCK TABLES `sys_dict` WRITE;
/*!40000 ALTER TABLE `sys_dict` DISABLE KEYS */;
INSERT INTO `sys_dict` VALUES (1,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','YESNO','是否',0,'是/否选择','Y'),(4,1,'admin','2026-01-26 21:00:25',NULL,NULL,'Y','abc','123',0,NULL,NULL),(5,1,'admin','2026-01-26 21:02:49',NULL,NULL,'Y','asdf','asdf',0,NULL,'c'),(6,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','live_room_type','直播间类型',0,NULL,NULL),(7,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','broadcast_format','播出形式',0,NULL,NULL),(8,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','room_stage','直播间阶段',0,NULL,NULL),(9,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','display_mode','显示方式',0,NULL,NULL),(10,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','viewing_method','观看方式',0,NULL,NULL),(11,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','playback_method','回放方式',0,NULL,NULL),(12,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','playback_validity','回放有效期',0,NULL,NULL),(13,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y','live_room_status','直播间状态',0,NULL,'draft');
/*!40000 ALTER TABLE `sys_dict` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_dict_item`
--

DROP TABLE IF EXISTS `sys_dict_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dict_item` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_DICT_ID` int unsigned NOT NULL COMMENT '所属字典',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字典值',
  `ORDERNO` int DEFAULT NULL COMMENT '排序',
  `CSSCLASS` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'css',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  `IS_DEFAULT_VALUE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '是否默认值(Y:是,N:否)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='数据字典明细';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_dict_item`
--

LOCK TABLES `sys_dict_item` WRITE;
/*!40000 ALTER TABLE `sys_dict_item` DISABLE KEYS */;
INSERT INTO `sys_dict_item` VALUES (1,1,'system','2026-01-12 20:52:14','admin','2026-03-13 04:45:06','Y',1,'是','Y',10,NULL,NULL,'Y'),(2,1,'system','2026-01-12 20:52:14','admin','2026-01-26 03:30:18','Y',1,'否','N',20,'disabled',NULL,'N'),(3,1,'admin','2026-01-28 01:16:17','admin','2026-01-28 02:49:23','Y',5,'a','a',30,NULL,NULL,'N'),(4,1,'admin','2026-01-28 02:44:20','admin','2026-01-28 02:48:50','Y',5,'b','b',20,NULL,NULL,'N'),(6,1,'admin','2026-01-28 02:48:32','admin','2026-01-28 02:48:46','Y',5,'c','c',10,NULL,NULL,'Y'),(10,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',6,'视频直播','video',10,NULL,NULL,NULL),(11,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',6,'图片直播','image',20,NULL,NULL,NULL),(12,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',6,'VR直播','vr',30,NULL,NULL,NULL),(13,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',6,'语音直播','audio',40,NULL,NULL,NULL),(14,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',6,'图文直播','graphic',50,NULL,NULL,NULL),(15,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',7,'直播','live',10,NULL,NULL,NULL),(16,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',7,'点播/录播','vod',20,NULL,NULL,NULL),(17,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',7,'伪直播','pseudo',30,NULL,NULL,NULL),(18,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',8,'正式直播','formal',10,NULL,NULL,NULL),(19,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',8,'测试直播','test',20,NULL,NULL,NULL),(20,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',9,'横屏','landscape',10,NULL,NULL,NULL),(21,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',9,'竖屏','portrait',20,NULL,NULL,NULL),(22,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',9,'三分屏','three_screen',30,NULL,NULL,NULL),(23,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',10,'公开','public',10,NULL,NULL,NULL),(24,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',10,'加密','encrypted',20,NULL,NULL,NULL),(25,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',10,'付费','paid',30,NULL,NULL,NULL),(26,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',10,'购票进入','ticket',40,NULL,NULL,NULL),(27,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',10,'企业成员观看','enterprise',50,NULL,NULL,NULL),(28,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',10,'自建成员观看','custom',60,NULL,NULL,NULL),(29,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',11,'结束后回放','post_end',10,NULL,NULL,NULL),(30,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',11,'实时回放','real_time',20,NULL,NULL,NULL),(31,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',11,'结束后不回放','no_playback',30,NULL,NULL,NULL),(32,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',12,'无限制','unlimited',10,NULL,NULL,NULL),(33,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',12,'全天','all_day',20,NULL,NULL,NULL),(34,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',12,'部分时段','partial',30,NULL,NULL,NULL),(35,1,'system','2026-02-22 23:58:17','admin','2026-03-04 02:04:22','Y',13,'草稿','draft',10,NULL,NULL,'Y'),(36,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',13,'已排期','scheduled',20,NULL,NULL,'N'),(37,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',13,'直播中','live',30,NULL,NULL,'N'),(38,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',13,'已结束','ended',40,NULL,NULL,'N'),(39,1,'system','2026-02-22 23:58:17','system','2026-02-22 23:58:17','Y',13,'已归档','archived',50,NULL,NULL,'N');
/*!40000 ALTER TABLE `sys_dict_item` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_directory`
--

DROP TABLE IF EXISTS `sys_directory`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_directory` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示名称',
  `SYS_TABLE_CATEGORY_ID` int DEFAULT NULL COMMENT '所属表类别',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '位置',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '对应表',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=57 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='安全目录';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_directory`
--

LOCK TABLES `sys_directory` WRITE;
/*!40000 ALTER TABLE `sys_directory` DISABLE KEYS */;
INSERT INTO `sys_directory` VALUES (1,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','AUDIT_LOG','审计日志',NULL,'',1),(2,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_ACTION','动作定义',NULL,'',2),(3,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_COLUMN','系统表字段',NULL,'',3),(4,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_COMPANY','模板表单',NULL,'',4),(5,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_DICT','数据字典',NULL,'',5),(6,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_DICT_ITEM','数据字典明细',NULL,'',6),(7,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_DIRECTORY','安全目录',NULL,'',7),(8,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_EMAIL_CONFIG','邮件配置表',NULL,'',8),(9,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_FILE','系统文件表',NULL,'',9),(10,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_GROUP_PREM','权限组明细',NULL,'',10),(11,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_GROUPS','权限组',NULL,'',11),(12,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_MESSAGE','系统消息表',NULL,'',12),(13,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_MESSAGE_TEMPLATE','消息模板表',NULL,'',13),(14,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_MODEL','模板表单',NULL,'',14),(15,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_NOTIFICATION_LOG','通知日志表',NULL,'',15),(16,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_OBJUICONF','对象显示配置',NULL,'',16),(17,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_PARAM','模板表单',NULL,'',17),(18,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_SEQ','序号生成器',NULL,'',18),(19,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_SUBSYSTEM','子系统',NULL,'',19),(20,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE','系统表单',NULL,'',20),(21,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_CATEGORY','表类别',NULL,'',21),(22,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_CMD','表单功能扩展',NULL,'',22),(23,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_REF','关联表',NULL,'',23),(24,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_SQL','表单sql\r\n',NULL,'',24),(25,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER','系统用户',NULL,'',25),(26,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_ENV','用户环境变量',NULL,'',26),(27,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_GROUPS','用户权限组',NULL,'',27),(28,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_MESSAGE','用户消息关联表',NULL,'',28),(29,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_SESSION','用户会话表',NULL,'',29),(30,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_DEFINITION','工作流定义',NULL,'',30),(31,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_INSTANCE','工作流实例',NULL,'',31),(32,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_NODE','工作流节点',NULL,'',32),(33,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_TASK','工作流任务',NULL,'',33),(34,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_TRANSITION','工作流流转',NULL,'',34),(51,1,'admin','2026-01-30 23:12:44',NULL,NULL,'Y','LIVE_DOMAIN_LIST','直播域名',0,NULL,35),(52,1,'admin','2026-01-31 17:18:50',NULL,NULL,'Y','LIVE_STREAM_LIST','直播流管理',2,NULL,38),(53,1,'admin','2026-02-02 00:08:45',NULL,NULL,'Y','LIVE_PULL_LIST','拉流转推管理',2,NULL,39),(54,1,'admin','2026-02-02 16:27:53',NULL,NULL,'Y','LIVE_CUT_LIST','直播切片',2,NULL,40),(55,1,'admin','2026-02-02 16:28:52',NULL,NULL,'Y','LIVE_RECODE_LIST','直播录制',2,NULL,41),(56,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','SYS_COMPANY_CONF_LIST','公司其他配置',0,NULL,55);
/*!40000 ALTER TABLE `sys_directory` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_email_config`
--

DROP TABLE IF EXISTS `sys_email_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_email_config` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `SMTP_HOST` varchar(100) NOT NULL COMMENT 'SMTP服务器地址',
  `SMTP_PORT` int NOT NULL COMMENT 'SMTP端口',
  `SMTP_USER` varchar(100) NOT NULL COMMENT 'SMTP用户名',
  `SMTP_PASSWORD` varchar(255) NOT NULL COMMENT 'SMTP密码（加密存储）',
  `FROM_EMAIL` varchar(100) NOT NULL COMMENT '发件人邮箱',
  `FROM_NAME` varchar(100) DEFAULT NULL COMMENT '发件人名称',
  `USE_TLS` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否使用TLS Y/N',
  `IS_DEFAULT` char(1) NOT NULL DEFAULT 'N' COMMENT '是否默认配置 Y/N',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`ID`),
  KEY `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='邮件配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_email_config`
--

LOCK TABLES `sys_email_config` WRITE;
/*!40000 ALTER TABLE `sys_email_config` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_email_config` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_file`
--

DROP TABLE IF EXISTS `sys_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_file` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `FILE_NAME` varchar(255) NOT NULL COMMENT '原始文件名',
  `STORAGE_NAME` varchar(255) NOT NULL COMMENT '存储文件名（唯一）',
  `FILE_PATH` varchar(500) NOT NULL COMMENT '文件路径',
  `FILE_SIZE` bigint NOT NULL COMMENT '文件大小（字节）',
  `FILE_TYPE` varchar(100) DEFAULT NULL COMMENT '文件类型/MIME类型',
  `FILE_EXT` varchar(20) DEFAULT NULL COMMENT '文件扩展名',
  `STORAGE_TYPE` varchar(20) NOT NULL DEFAULT 'local' COMMENT '存储类型：local, oss, s3',
  `BUCKET_NAME` varchar(100) DEFAULT NULL COMMENT '存储桶名称（云存储）',
  `ACCESS_URL` varchar(500) DEFAULT NULL COMMENT '访问URL',
  `THUMBNAIL_URL` varchar(500) DEFAULT NULL COMMENT '缩略图URL',
  `MD5` varchar(32) DEFAULT NULL COMMENT '文件MD5值',
  `UPLOAD_IP` varchar(50) DEFAULT NULL COMMENT '上传IP',
  `DOWNLOAD_COUNT` int NOT NULL DEFAULT '0' COMMENT '下载次数',
  `CATEGORY` varchar(50) DEFAULT NULL COMMENT '文件分类',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '文件描述',
  `EXPIRE_TIME` datetime DEFAULT NULL COMMENT '过期时间',
  PRIMARY KEY (`ID`),
  KEY `idx_storage_name` (`STORAGE_NAME`),
  KEY `idx_md5` (`MD5`),
  KEY `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统文件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_file`
--

LOCK TABLES `sys_file` WRITE;
/*!40000 ALTER TABLE `sys_file` DISABLE KEYS */;
INSERT INTO `sys_file` VALUES (1,0,'admin','2026-01-28 18:42:16','admin','2026-01-28 18:42:16','Y','logo.png','dce39eec-3d96-4142-91a9-6704f916b7e1.png','uploads\\2026\\01\\28\\dce39eec-3d96-4142-91a9-6704f916b7e1.png',1623813,'image/png','.png','local','','/api/v1/files/download/dce39eec-3d96-4142-91a9-6704f916b7e1.png','','037359d2f1ccfb6c03c4a0c706e751f6','127.0.0.1',0,'default','',NULL),(2,0,'admin','2026-01-28 18:42:56','admin','2026-01-28 18:42:56','Y','logo.png','dce39eec-3d96-4142-91a9-6704f916b7e1.png','uploads\\2026\\01\\28\\dce39eec-3d96-4142-91a9-6704f916b7e1.png',1623813,'image/png','.png','local','','/api/v1/files/download/dce39eec-3d96-4142-91a9-6704f916b7e1.png','','037359d2f1ccfb6c03c4a0c706e751f6','127.0.0.1',0,'default','',NULL),(3,0,'admin','2026-01-28 18:49:26','admin','2026-01-28 18:49:26','Y','logo.png','dce39eec-3d96-4142-91a9-6704f916b7e1.png','uploads\\2026\\01\\28\\dce39eec-3d96-4142-91a9-6704f916b7e1.png',1623813,'image/png','.png','local','','/api/v1/files/access/dce39eec-3d96-4142-91a9-6704f916b7e1.png','','037359d2f1ccfb6c03c4a0c706e751f6','127.0.0.1',0,'default','',NULL),(4,0,'admin','2026-01-28 18:52:06','admin','2026-01-28 18:52:06','Y','logo.png','dce39eec-3d96-4142-91a9-6704f916b7e1.png','uploads\\2026\\01\\28\\dce39eec-3d96-4142-91a9-6704f916b7e1.png',1623813,'image/png','.png','local','','/api/v1/files/access/dce39eec-3d96-4142-91a9-6704f916b7e1.png','','037359d2f1ccfb6c03c4a0c706e751f6','127.0.0.1',0,'default','',NULL),(5,0,'admin','2026-01-28 18:52:24','admin','2026-01-28 18:52:24','Y','logo.png','dce39eec-3d96-4142-91a9-6704f916b7e1.png','uploads\\2026\\01\\28\\dce39eec-3d96-4142-91a9-6704f916b7e1.png',1623813,'image/png','.png','local','','/api/v1/files/access/dce39eec-3d96-4142-91a9-6704f916b7e1.png','','037359d2f1ccfb6c03c4a0c706e751f6','127.0.0.1',0,'default','',NULL),(6,0,'admin','2026-01-28 18:52:47','admin','2026-01-28 18:52:47','Y','logo.png','dce39eec-3d96-4142-91a9-6704f916b7e1.png','uploads\\2026\\01\\28\\dce39eec-3d96-4142-91a9-6704f916b7e1.png',1623813,'image/png','.png','local','','/api/v1/files/access/dce39eec-3d96-4142-91a9-6704f916b7e1.png','','037359d2f1ccfb6c03c4a0c706e751f6','127.0.0.1',0,'default','',NULL),(7,0,'admin','2026-01-28 18:52:54','admin','2026-01-28 18:52:54','Y','ChatGPT Image 2026年1月28日 04_47_12.png','a16dd4fc-4412-4367-a882-668f6583b04a.png','uploads\\2026\\01\\28\\a16dd4fc-4412-4367-a882-668f6583b04a.png',2025971,'image/png','.png','local','','/api/v1/files/access/a16dd4fc-4412-4367-a882-668f6583b04a.png','','859c6e876f68fdb254e645aa665f7116','127.0.0.1',0,'default','',NULL),(8,0,'admin','2026-02-24 01:15:27','admin','2026-02-24 01:15:27','Y','1.png','75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png','uploads\\2026\\02\\24\\75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png',60997,'image/png','.png','local','','/api/v1/files/access/75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png','','eb08780c3981499274d214f787323c59','127.0.0.1',0,'live_cover','',NULL),(9,0,'admin','2026-02-24 01:17:47','admin','2026-02-24 01:17:47','Y','1.png','75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png','uploads\\2026\\02\\24\\75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png',60997,'image/png','.png','local','','/api/v1/files/access/75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png','','eb08780c3981499274d214f787323c59','127.0.0.1',0,'live_cover','',NULL),(10,0,'admin','2026-02-24 01:18:46','admin','2026-02-24 01:18:46','Y','1.png','75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png','uploads\\2026\\02\\24\\75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png',60997,'image/png','.png','local','','/api/v1/files/access/75c2174a-eb91-41b0-8bbd-5dc073ce1f27.png','','eb08780c3981499274d214f787323c59','127.0.0.1',0,'live_cover','',NULL),(11,0,'admin','2026-02-24 01:20:06','admin','2026-02-24 01:20:06','Y','2.png','0f2e81e4-dba7-4468-941b-3741e3adcce5.png','uploads\\2026\\02\\24\\0f2e81e4-dba7-4468-941b-3741e3adcce5.png',22680,'image/png','.png','local','','/api/v1/files/access/0f2e81e4-dba7-4468-941b-3741e3adcce5.png','','62487fbcd838f9639151f88c4f22b2d9','127.0.0.1',0,'live_cover','',NULL),(12,0,'admin','2026-02-27 14:27:54','admin','2026-02-27 14:27:54','Y','微信图片_20260226161630_534_88.png','f38abfa3-5856-4fe3-ba21-901998d85b13.png','uploads\\2026\\02\\27\\f38abfa3-5856-4fe3-ba21-901998d85b13.png',132499,'image/png','.png','local','','/api/v1/files/access/f38abfa3-5856-4fe3-ba21-901998d85b13.png','','47cf64d9918f5025c2c4f984bbb7fa34','127.0.0.1',0,'live_cover','',NULL),(13,0,'admin','2026-02-27 15:38:35','admin','2026-02-27 15:38:35','Y','截屏2026-02-26 14.34.41.png','f0a0fc80-aa81-44e1-805d-afdda7e95501.png','uploads\\2026\\02\\27\\f0a0fc80-aa81-44e1-805d-afdda7e95501.png',3610611,'image/png','.png','local','','/api/v1/files/access/f0a0fc80-aa81-44e1-805d-afdda7e95501.png','','78515b543d5257783d2546706d0e42bf','127.0.0.1',0,'live_cover','',NULL);
/*!40000 ALTER TABLE `sys_file` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_group_prem`
--

DROP TABLE IF EXISTS `sys_group_prem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_group_prem` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_GROUPS_ID` int DEFAULT NULL COMMENT '权限组',
  `SYS_DIRECTORY_ID` int DEFAULT NULL COMMENT '目录\r\n',
  `PERMISSION` int DEFAULT NULL COMMENT '权限(1:读;3:读,写;5:读,提交;……)',
  `FILTER_OBJ` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '数据过滤({sql:"",display:"",other:""})',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='权限组明细';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_group_prem`
--

LOCK TABLES `sys_group_prem` WRITE;
/*!40000 ALTER TABLE `sys_group_prem` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_group_prem` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_groups`
--

DROP TABLE IF EXISTS `sys_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_groups` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '名称',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  `SGRADE` int DEFAULT NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='权限组';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_groups`
--

LOCK TABLES `sys_groups` WRITE;
/*!40000 ALTER TABLE `sys_groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_message`
--

DROP TABLE IF EXISTS `sys_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_message` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `TITLE` varchar(255) NOT NULL COMMENT '消息标题',
  `CONTENT` text NOT NULL COMMENT '消息内容',
  `MESSAGE_TYPE` varchar(50) NOT NULL COMMENT '消息类型: system, workflow, business, notice',
  `PRIORITY` int NOT NULL DEFAULT '0' COMMENT '优先级: 0=普通, 1=重要, 2=紧急',
  `CATEGORY` varchar(50) DEFAULT NULL COMMENT '消息分类',
  `SENDER_ID` int unsigned DEFAULT NULL COMMENT '发送者ID（系统消息为NULL）',
  `SENDER_NAME` varchar(100) DEFAULT NULL COMMENT '发送者姓名',
  `TARGET_TYPE` varchar(20) NOT NULL DEFAULT 'user' COMMENT '目标类型: user, role, group, all',
  `TARGET_IDS` varchar(1000) DEFAULT NULL COMMENT '目标ID列表（逗号分隔）',
  `LINK_URL` varchar(500) DEFAULT NULL COMMENT '关联URL',
  `LINK_TYPE` varchar(50) DEFAULT NULL COMMENT '链接类型: internal, external',
  `PARAMS` text COMMENT '消息参数（JSON）',
  `TEMPLATE_ID` int unsigned DEFAULT NULL COMMENT '消息模板ID',
  `READ_COUNT` int NOT NULL DEFAULT '0' COMMENT '已读人数',
  `TOTAL_COUNT` int NOT NULL DEFAULT '0' COMMENT '总接收人数',
  `EXPIRE_TIME` datetime DEFAULT NULL COMMENT '过期时间',
  `STATUS` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态: active, expired, deleted',
  PRIMARY KEY (`ID`),
  KEY `idx_message_type` (`MESSAGE_TYPE`),
  KEY `idx_priority` (`PRIORITY`),
  KEY `idx_category` (`CATEGORY`),
  KEY `idx_sender` (`SENDER_ID`),
  KEY `idx_template` (`TEMPLATE_ID`),
  KEY `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统消息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_message`
--

LOCK TABLES `sys_message` WRITE;
/*!40000 ALTER TABLE `sys_message` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_message` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_message_template`
--

DROP TABLE IF EXISTS `sys_message_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_message_template` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `CODE` varchar(50) NOT NULL COMMENT '模板代码',
  `NAME` varchar(100) NOT NULL COMMENT '模板名称',
  `MESSAGE_TYPE` varchar(50) NOT NULL COMMENT '消息类型',
  `TITLE` varchar(255) NOT NULL COMMENT '标题模板',
  `CONTENT` text NOT NULL COMMENT '内容模板',
  `VARIABLES` varchar(500) DEFAULT NULL COMMENT '变量列表（逗号分隔）',
  `DESCRIPTION` varchar(500) DEFAULT NULL COMMENT '描述',
  `IS_ENABLED` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用 Y/N',
  `CATEGORY` varchar(50) DEFAULT NULL COMMENT '分类',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idx_code` (`CODE`),
  KEY `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='消息模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_message_template`
--

LOCK TABLES `sys_message_template` WRITE;
/*!40000 ALTER TABLE `sys_message_template` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_message_template` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_model`
--

DROP TABLE IF EXISTS `sys_model`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_model` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='模板表单';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_model`
--

LOCK TABLES `sys_model` WRITE;
/*!40000 ALTER TABLE `sys_model` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_model` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_notification_log`
--

DROP TABLE IF EXISTS `sys_notification_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_notification_log` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `MESSAGE_ID` int unsigned DEFAULT NULL COMMENT '消息ID',
  `USER_ID` int unsigned DEFAULT NULL COMMENT '接收用户ID',
  `NOTIFY_TYPE` varchar(20) NOT NULL COMMENT '通知类型: websocket, email, sms',
  `STATUS` varchar(20) NOT NULL COMMENT '状态: pending, sent, failed, read',
  `SENT_TIME` datetime DEFAULT NULL COMMENT '发送时间',
  `READ_TIME` datetime DEFAULT NULL COMMENT '读取时间',
  `ERROR_MESSAGE` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `RETRY_COUNT` int NOT NULL DEFAULT '0' COMMENT '重试次数',
  PRIMARY KEY (`ID`),
  KEY `idx_message` (`MESSAGE_ID`),
  KEY `idx_user` (`USER_ID`),
  KEY `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='通知日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_notification_log`
--

LOCK TABLES `sys_notification_log` WRITE;
/*!40000 ALTER TABLE `sys_notification_log` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_notification_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_objuiconf`
--

DROP TABLE IF EXISTS `sys_objuiconf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_objuiconf` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示名称',
  `TABLE_PARAM_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'tableid参数名',
  `PK_PARAM_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'id参数名',
  `CSS_CLASS` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'CSS类',
  `COLS` int DEFAULT NULL COMMENT '每行字段个数',
  `DEFAULT_ACTION` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '缺省动作',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='对象显示配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_objuiconf`
--

LOCK TABLES `sys_objuiconf` WRITE;
/*!40000 ALTER TABLE `sys_objuiconf` DISABLE KEYS */;
INSERT INTO `sys_objuiconf` VALUES (1,1,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL,1,NULL),(2,1,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL,2,NULL),(3,1,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL,3,NULL),(4,1,NULL,NULL,NULL,NULL,'Y',NULL,NULL,NULL,NULL,NULL,4,NULL);
/*!40000 ALTER TABLE `sys_objuiconf` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_param`
--

DROP TABLE IF EXISTS `sys_param`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_param` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '名称',
  `DEFAULT_VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '默认值',
  `VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '当前值',
  `VALUE_TYPE` char(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '值类型',
  `VALUE_LIST` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '值列表',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='模板表单';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_param`
--

LOCK TABLES `sys_param` WRITE;
/*!40000 ALTER TABLE `sys_param` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_param` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_seq`
--

DROP TABLE IF EXISTS `sys_seq`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_seq` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示名称',
  `VFORMAT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '格式',
  `INCRE` int DEFAULT NULL COMMENT '递增',
  `CYCLETYPE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '循环方式',
  `PREFIX` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '前缀',
  `SUFFIX` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '后缀',
  `CUR_DATE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '当前周期值',
  `CUR_NUM` int DEFAULT NULL COMMENT '当前流水号',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='序号生成器';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_seq`
--

LOCK TABLES `sys_seq` WRITE;
/*!40000 ALTER TABLE `sys_seq` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_seq` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_subsystem`
--

DROP TABLE IF EXISTS `sys_subsystem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_subsystem` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '名称',
  `ORDERNO` int DEFAULT NULL COMMENT '序号',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '网页链接',
  `ICON` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'icon',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  `KEY` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='子系统';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_subsystem`
--

LOCK TABLES `sys_subsystem` WRITE;
/*!40000 ALTER TABLE `sys_subsystem` DISABLE KEYS */;
INSERT INTO `sys_subsystem` VALUES (1,1,NULL,NULL,'admin','2026-02-08 01:31:27','Y','系统管理',1000,'/system-management',NULL,NULL,NULL),(7,1,'admin','2026-01-22 03:19:35','admin','2026-02-13 23:41:05','Y','视频直播',10,'/LiveGuide',NULL,NULL,'live'),(10,1,'admin',NULL,NULL,NULL,'Y','云盘中心',20,'/cloud',NULL,NULL,'cloud');
/*!40000 ALTER TABLE `sys_subsystem` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_table`
--

DROP TABLE IF EXISTS `sys_table`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_table` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `REAL_TABLE_ID` int unsigned DEFAULT NULL COMMENT '实际数据库表',
  `FILTER` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '数据过滤SQL',
  `DK_COLUMN_ID` int unsigned DEFAULT NULL COMMENT '显示主键(DK)',
  `AK_COLUMN_ID` int DEFAULT NULL COMMENT '输入主键(AK)',
  `MASK` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '表单规则(支持：A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废)',
  `SYS_TABLECATEGORY_ID` int unsigned DEFAULT NULL COMMENT '表类别',
  `ORDERNO` int DEFAULT NULL COMMENT '排序',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '网页连接',
  `RPC_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'rpc 方法',
  `IS_MENU` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'N' COMMENT '是否菜单(Y:是,N:否)',
  `ICO_IMG` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '表单ICO图片',
  `IS_DROPDOWN` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '是否下拉框',
  `SYS_OBJUICONF_ID` int DEFAULT NULL COMMENT '显示配置',
  `SYS_DIRECTORY_ID` int DEFAULT NULL COMMENT '安全目录',
  `SYS_PARENT_TABLE_ID` int DEFAULT NULL COMMENT '父表',
  `ROWCNT` int DEFAULT NULL COMMENT '统计行数',
  `IS_BIG` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '是否海量',
  `PROPS` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '扩展属性',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE KEY `IDX_SYSTABLE_NAME` (`NAME`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=57 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='系统表单';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_table`
--

LOCK TABLES `sys_table` WRITE;
/*!40000 ALTER TABLE `sys_table` DISABLE KEYS */;
INSERT INTO `sys_table` VALUES (1,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','AUDIT_LOG','审计日志',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,'审计日志'),(2,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_ACTION','动作定义',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,2,NULL,NULL,NULL,NULL,'动作定义'),(3,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_COLUMN','系统表字段',NULL,NULL,NULL,NULL,'AMDQ',1,50,'/metadata/list/3',NULL,'Y',NULL,NULL,2,3,NULL,NULL,NULL,NULL,'系统表字段'),(4,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_COMPANY','公司',NULL,NULL,91,91,'AMDQ',1,10,'/metadata/list/4',NULL,'Y',NULL,'Y',2,4,NULL,NULL,NULL,NULL,'模板表单'),(5,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_DICT','数据字典',NULL,NULL,NULL,NULL,'AMDQ',1,60,'/metadata/list/5',NULL,'Y',NULL,NULL,NULL,5,NULL,NULL,NULL,NULL,'数据字典'),(6,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_DICT_ITEM','数据字典明细',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,'',NULL,'N',NULL,NULL,1,6,NULL,NULL,NULL,NULL,'数据字典明细'),(7,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_DIRECTORY','安全目录',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,7,NULL,NULL,NULL,NULL,'安全目录'),(8,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_EMAIL_CONFIG','邮件配置表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,8,NULL,NULL,NULL,NULL,'邮件配置表'),(9,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_FILE','系统文件表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,'',NULL,'N',NULL,NULL,NULL,9,NULL,NULL,NULL,NULL,'系统文件表'),(10,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_GROUP_PREM','权限明细',NULL,NULL,NULL,NULL,'AMDQ',3,NULL,NULL,NULL,'N',NULL,NULL,NULL,10,NULL,NULL,NULL,NULL,'权限组明细，定义权限组对各个目录的访问权限'),(11,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_GROUPS','权限组管理',NULL,NULL,NULL,NULL,'AMDQE',3,NULL,NULL,NULL,'Y',NULL,NULL,NULL,11,NULL,NULL,NULL,NULL,'权限组管理，用于定义不同的权限组并分配权限'),(12,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_MESSAGE','系统消息表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,12,NULL,NULL,NULL,NULL,'系统消息表'),(13,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_MESSAGE_TEMPLATE','消息模板表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,13,NULL,NULL,NULL,NULL,'消息模板表'),(14,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_MODEL','模板表单',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,14,NULL,NULL,NULL,NULL,'模板表单'),(15,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_NOTIFICATION_LOG','通知日志表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,15,NULL,NULL,NULL,NULL,'通知日志表'),(16,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_OBJUICONF','对象显示配置',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,16,NULL,NULL,NULL,NULL,'对象显示配置'),(17,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_PARAM','模板表单',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,17,NULL,NULL,NULL,NULL,'模板表单'),(18,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_SEQ','序号生成器',NULL,NULL,NULL,NULL,'AMDQ',1,70,'/metadata/list/18',NULL,'Y',NULL,NULL,NULL,18,NULL,NULL,NULL,NULL,'序号生成器'),(19,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_SUBSYSTEM','子系统',NULL,NULL,NULL,NULL,'AMDQ',1,20,'/metadata/list/19',NULL,'Y',NULL,'Y',3,19,NULL,NULL,NULL,NULL,'子系统'),(20,1,'system','2026-01-12 20:52:14','admin','2026-01-28 13:51:42','Y','SYS_TABLE','系统表单',NULL,NULL,NULL,NULL,'AMDQ',1,40,'/metadata/list/20',NULL,'Y',NULL,'',3,20,NULL,NULL,NULL,'{\"a\":\"b\",\"c\":\"d\"}',NULL),(21,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_CATEGORY','表类别',NULL,NULL,NULL,NULL,'AMDQ',1,30,'/metadata/list/21',NULL,'Y',NULL,'Y',4,21,NULL,NULL,NULL,NULL,'表类别'),(22,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_CMD','表单功能扩展',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,22,NULL,NULL,NULL,NULL,'表单功能扩展'),(23,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_REF','关联表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,23,NULL,NULL,NULL,NULL,'关联表'),(24,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_TABLE_SQL','表单sql\r\n',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,24,NULL,NULL,NULL,NULL,'表单sql\r\n'),(25,1,'system','2026-01-12 20:52:14','admin','2026-02-02 20:44:30','Y','SYS_USER','系统用户',NULL,NULL,NULL,NULL,'AMDQ',3,10,'/metadata/list/25',NULL,'Y',NULL,NULL,NULL,25,NULL,NULL,NULL,NULL,'系统用户'),(26,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_ENV','用户环境变量',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,26,NULL,NULL,NULL,NULL,'用户环境变量'),(27,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_GROUPS','用户权限组',NULL,NULL,NULL,NULL,'AMDQ',3,NULL,NULL,NULL,'N',NULL,NULL,NULL,27,NULL,NULL,NULL,NULL,'用户权限组关联，定义用户所属的权限组'),(28,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_MESSAGE','用户消息关联表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,28,NULL,NULL,NULL,NULL,'用户消息关联表'),(29,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','SYS_USER_SESSION','用户会话表',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,29,NULL,NULL,NULL,NULL,'用户会话表'),(30,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_DEFINITION','工作流定义',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,30,NULL,NULL,NULL,NULL,'工作流定义'),(31,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_INSTANCE','工作流实例',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,31,NULL,NULL,NULL,NULL,'工作流实例'),(32,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_NODE','工作流节点',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,32,NULL,NULL,NULL,NULL,'工作流节点'),(33,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_TASK','工作流任务',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,33,NULL,NULL,NULL,NULL,'工作流任务'),(34,1,'system','2026-01-12 20:52:14',NULL,NULL,'Y','WF_TRANSITION','工作流流转',NULL,NULL,NULL,NULL,'AMDQ',1,NULL,NULL,NULL,'N',NULL,NULL,NULL,34,NULL,NULL,NULL,NULL,'工作流流转'),(35,1,'admin','2026-01-30 23:12:44','admin','2026-01-30 23:26:31','Y','LIVE_DOMAIN','直播域名',NULL,NULL,NULL,NULL,'AMDQ',2,10,'/live/domains',NULL,'Y',NULL,NULL,NULL,51,NULL,NULL,NULL,NULL,NULL),(37,1,'admin','2026-01-31 00:18:56','admin','2026-01-31 17:39:25','Y','CLOUD_ITEM','云盘',NULL,NULL,NULL,NULL,'AMDQ',4,100,'/cloud',NULL,'Y',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),(38,1,'admin','2026-01-31 17:18:50','admin','2026-01-31 17:19:15','Y','LIVE_STREAM','直播流管理',NULL,NULL,NULL,NULL,'AMDQ',2,30,'/live/streams',NULL,'Y',NULL,NULL,NULL,52,NULL,NULL,NULL,NULL,NULL),(39,1,'admin','2026-02-02 00:08:45','admin','2026-03-01 23:03:07','Y','LIVE_PULL','社媒分发',NULL,NULL,NULL,NULL,'AMDQ',2,40,'/live/pull-stream',NULL,'Y',NULL,NULL,NULL,53,NULL,NULL,NULL,NULL,NULL),(40,1,'admin','2026-02-02 16:27:53','admin','2026-02-28 08:53:31','Y','LIVE_CUT','直播切片',NULL,NULL,NULL,NULL,'AMDQ',2,50,'/live/highlight-clips',NULL,'N',NULL,NULL,NULL,54,NULL,NULL,NULL,NULL,NULL),(41,1,'admin','2026-02-02 16:28:52','admin','2026-03-01 23:03:21','Y','LIVE_RECODE','直播录制',NULL,NULL,NULL,NULL,'AMDQ',2,60,'/live/recordings',NULL,'Y',NULL,NULL,NULL,55,NULL,NULL,NULL,NULL,NULL),(54,1,'system','2026-02-22 23:50:27','admin','2026-02-28 04:45:28','Y','LIVE_ROOM','直播间管理',NULL,NULL,NULL,NULL,'AMDQSUVE',2,35,'/live/rooms',NULL,'Y',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'直播间管理模块'),(55,1,'admin','2026-03-04 02:22:59',NULL,NULL,'Y','SYS_COMPANY_CONF','公司其他配置',NULL,NULL,NULL,NULL,NULL,NULL,10,NULL,NULL,'N',NULL,NULL,NULL,56,NULL,NULL,NULL,NULL,NULL);
/*!40000 ALTER TABLE `sys_table` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_table_category`
--

DROP TABLE IF EXISTS `sys_table_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_table_category` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_SUBSYSTEM_ID` int unsigned DEFAULT NULL COMMENT '所属子系统',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '名称',
  `ORDERNO` int DEFAULT NULL COMMENT '序号',
  `ICON` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'icon图标',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '网页连接',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_subsystem` (`SYS_SUBSYSTEM_ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='表类别';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_table_category`
--

LOCK TABLES `sys_table_category` WRITE;
/*!40000 ALTER TABLE `sys_table_category` DISABLE KEYS */;
INSERT INTO `sys_table_category` VALUES (1,1,NULL,NULL,'admin','2026-01-22 03:24:28','Y',1,'元数据',10,NULL,NULL,NULL),(2,1,'admin','2026-01-22 03:24:14','admin','2026-01-22 03:28:02','Y',7,'直播中心',10,NULL,NULL,NULL),(3,1,'admin','2026-02-02 20:41:51',NULL,NULL,'Y',1,'用户权限',10,NULL,NULL,NULL),(4,1,NULL,NULL,NULL,NULL,'Y',10,'云盘',10,NULL,NULL,NULL);
/*!40000 ALTER TABLE `sys_table_category` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_table_cmd`
--

DROP TABLE IF EXISTS `sys_table_cmd`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_table_cmd` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '所属表单',
  `ACTION_TYPE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '按钮类型(1:系统按钮)',
  `ACTION` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '按钮(A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废,I:导入,E:导出)',
  `ACTION_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '按钮名称',
  `EVENT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '事件前后(begin:开始,end:结束)',
  `CONTENT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '执行操作(存储过程/action动作)',
  `CONTENT_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '动作类型',
  `ORDERNO` int DEFAULT NULL COMMENT '序号',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='表单功能扩展';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_table_cmd`
--

LOCK TABLES `sys_table_cmd` WRITE;
/*!40000 ALTER TABLE `sys_table_cmd` DISABLE KEYS */;
INSERT INTO `sys_table_cmd` VALUES (1,1,NULL,NULL,NULL,NULL,'Y',20,'1','A','新增','end','SYS_TABLE_AFTER_CREATE','go',1),(2,1,NULL,NULL,NULL,NULL,'Y',20,'1','D','删除','begin','SYS_TABLE_BEFORE_DELETE','go',2),(3,1,NULL,NULL,NULL,NULL,'Y',6,'1','A','新增','end','SYS_DICT_ITEM_AFTER_EDIT','go',1),(4,1,NULL,NULL,NULL,NULL,'Y',6,'1','M','修改','end','SYS_DICT_ITEM_AFTER_EDIT','go',1),(5,1,NULL,NULL,NULL,NULL,'Y',54,'1','A','新增','end','LIVE_ROOM_AFTER_CREATE','go',1);
/*!40000 ALTER TABLE `sys_table_cmd` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_table_ref`
--

DROP TABLE IF EXISTS `sys_table_ref`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_table_ref` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '主表',
  `ORDERNO` int DEFAULT NULL COMMENT '序号',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示描述',
  `REF_TABLE_ID` int DEFAULT NULL COMMENT '关联表',
  `REF_COLUMN_ID` int DEFAULT NULL COMMENT '关联字段',
  `FILTER` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '过滤条件',
  `ASSOCTYPE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '关联方式(1 : 1对1, n: 1对n )',
  `EDIT_TYPE` char(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '编辑方式(Y:标准(新增和修改行时都可在内嵌窗口编辑),\r\nN:无(无内嵌编辑窗口),NP:非内嵌，允许弹出,NS:非内嵌，禁止弹出,A:仅显示新增字段，修改直接修改)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_table_ref`
--

LOCK TABLES `sys_table_ref` WRITE;
/*!40000 ALTER TABLE `sys_table_ref` DISABLE KEYS */;
INSERT INTO `sys_table_ref` VALUES (1,1,'admin',NULL,'admin',NULL,'Y',5,10,'明细',6,104,NULL,'n','A'),(2,1,'admin',NULL,'admin',NULL,'Y',4,10,'公司配置',55,755,NULL,'1','Y');
/*!40000 ALTER TABLE `sys_table_ref` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_table_sql`
--

DROP TABLE IF EXISTS `sys_table_sql`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_table_sql` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '所属表单',
  `SQL` varchar(5000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '表单sql',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='表单sql\r\n';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_table_sql`
--

LOCK TABLES `sys_table_sql` WRITE;
/*!40000 ALTER TABLE `sys_table_sql` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_table_sql` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_user`
--

DROP TABLE IF EXISTS `sys_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `TRUE_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '真实名称',
  `USERNAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '用户名称',
  `PASSWORD` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '密码',
  `PHONE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '手机号',
  `EMAIL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '邮箱',
  `LANGUAGE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '语言',
  `IS_ADMIN` char(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'Y' COMMENT '是否管理员',
  `SGRADE` int DEFAULT NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='系统用户';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_user`
--

LOCK TABLES `sys_user` WRITE;
/*!40000 ALTER TABLE `sys_user` DISABLE KEYS */;
INSERT INTO `sys_user` VALUES (1,1,'admin','2026-01-12 20:52:13','admin','2026-01-12 20:52:13','Y','系统管理员','admin','$2a$10$iztoR7MeHZKyoBNpJM4pjOZ729KAoy.5x5ayetl1Rnb3TBgVCO0jy','13800138000','admin@example.com','zh-cn','Y',99);
/*!40000 ALTER TABLE `sys_user` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_user_env`
--

DROP TABLE IF EXISTS `sys_user_env`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_env` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '变量名称',
  `VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '值来源',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='用户环境变量';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_user_env`
--

LOCK TABLES `sys_user_env` WRITE;
/*!40000 ALTER TABLE `sys_user_env` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_user_env` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_user_groups`
--

DROP TABLE IF EXISTS `sys_user_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_groups` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_USER_ID` int DEFAULT NULL COMMENT '用户',
  `SYS_DIRECTORY_ID` int DEFAULT NULL COMMENT '权限组',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='用户权限组';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_user_groups`
--

LOCK TABLES `sys_user_groups` WRITE;
/*!40000 ALTER TABLE `sys_user_groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_user_groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_user_message`
--

DROP TABLE IF EXISTS `sys_user_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_message` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '公司ID',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  `MESSAGE_ID` int unsigned NOT NULL COMMENT '消息ID',
  `USER_ID` int unsigned NOT NULL COMMENT '用户ID',
  `IS_READ` char(1) NOT NULL DEFAULT 'N' COMMENT '是否已读 Y/N',
  `READ_TIME` datetime DEFAULT NULL COMMENT '读取时间',
  `IS_STARRED` char(1) NOT NULL DEFAULT 'N' COMMENT '是否星标 Y/N',
  `IS_ARCHIVED` char(1) NOT NULL DEFAULT 'N' COMMENT '是否归档 Y/N',
  `DELETED_AT` datetime DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`ID`),
  KEY `idx_user_msg` (`USER_ID`,`MESSAGE_ID`),
  KEY `idx_is_read` (`IS_READ`),
  KEY `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户消息关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_user_message`
--

LOCK TABLES `sys_user_message` WRITE;
/*!40000 ALTER TABLE `sys_user_message` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_user_message` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_user_session`
--

DROP TABLE IF EXISTS `sys_user_session`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_session` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `USER_ID` int unsigned NOT NULL COMMENT '用户ID',
  `COMPANY_ID` int unsigned NOT NULL COMMENT '公司ID',
  `TOKEN` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Access Token',
  `REFRESH_TOKEN` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Refresh Token',
  `CLIENT_TYPE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '客户端类型',
  `DEVICE_ID` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '设备ID',
  `DEVICE_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '设备名称',
  `IP_ADDRESS` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'IP地址',
  `USER_AGENT` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'User Agent',
  `LOGIN_TIME` datetime NOT NULL COMMENT '登录时间',
  `LAST_ACTIVE_TIME` datetime DEFAULT NULL COMMENT '最后活跃时间',
  `EXPIRE_TIME` datetime DEFAULT NULL COMMENT '过期时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'Y' COMMENT '是否有效(Y/N)',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE KEY `idx_session_device` (`DEVICE_ID`) USING BTREE,
  KEY `idx_session_user` (`USER_ID`) USING BTREE,
  KEY `idx_session_token` (`TOKEN`(255)) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户会话表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_user_session`
--

LOCK TABLES `sys_user_session` WRITE;
/*!40000 ALTER TABLE `sys_user_session` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_user_session` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `wf_definition`
--

DROP TABLE IF EXISTS `wf_definition`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wf_definition` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '流程名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示名称',
  `VERSION` int NOT NULL DEFAULT '1' COMMENT '版本号',
  `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'draft' COMMENT '状态(draft:草稿,published:已发布,archived:已归档)',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '关联的业务表',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '描述',
  `CONFIG` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT 'JSON配置',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_wf_def_table` (`SYS_TABLE_ID`) USING BTREE,
  KEY `idx_wf_def_status` (`STATUS`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='工作流定义';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `wf_definition`
--

LOCK TABLES `wf_definition` WRITE;
/*!40000 ALTER TABLE `wf_definition` DISABLE KEYS */;
/*!40000 ALTER TABLE `wf_definition` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `wf_instance`
--

DROP TABLE IF EXISTS `wf_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wf_instance` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `WF_DEFINITION_ID` int unsigned NOT NULL COMMENT '流程定义ID',
  `SYS_TABLE_ID` int DEFAULT NULL COMMENT '关联的业务表',
  `BUSINESS_ID` int unsigned DEFAULT NULL COMMENT '业务数据ID',
  `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '状态(running:运行中,completed:已完成,terminated:已终止,suspended:已挂起)',
  `CURRENT_NODE_ID` int unsigned DEFAULT NULL COMMENT '当前节点ID',
  `START_USER_ID` int unsigned DEFAULT NULL COMMENT '发起人',
  `START_TIME` datetime DEFAULT NULL COMMENT '开始时间',
  `END_TIME` datetime DEFAULT NULL COMMENT '结束时间',
  `VARIABLES` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '流程变量(JSON)',
  `TITLE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '流程标题',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_wf_inst_def` (`WF_DEFINITION_ID`) USING BTREE,
  KEY `idx_wf_inst_biz` (`BUSINESS_ID`) USING BTREE,
  KEY `idx_wf_inst_status` (`STATUS`) USING BTREE,
  KEY `idx_wf_inst_node` (`CURRENT_NODE_ID`) USING BTREE,
  KEY `idx_wf_inst_user` (`START_USER_ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='工作流实例';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `wf_instance`
--

LOCK TABLES `wf_instance` WRITE;
/*!40000 ALTER TABLE `wf_instance` DISABLE KEYS */;
/*!40000 ALTER TABLE `wf_instance` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `wf_node`
--

DROP TABLE IF EXISTS `wf_node`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wf_node` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `WF_DEFINITION_ID` int unsigned NOT NULL COMMENT '所属流程定义',
  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '节点名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '显示名称',
  `NODE_TYPE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '节点类型(start:开始,end:结束,user:用户任务,auto:自动任务,gateway:网关)',
  `ASSIGN_TYPE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '分配类型(user:指定用户,starter:发起人,role:角色,expression:表达式)',
  `ASSIGN_VALUE` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '分配值',
  `ACTION_ID` int unsigned DEFAULT NULL COMMENT '自动任务关联的动作ID',
  `CONFIG` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT 'JSON配置',
  `POS_X` int DEFAULT NULL COMMENT '节点X坐标',
  `POS_Y` int DEFAULT NULL COMMENT '节点Y坐标',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_wf_node_def` (`WF_DEFINITION_ID`) USING BTREE,
  KEY `idx_wf_node_action` (`ACTION_ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='工作流节点';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `wf_node`
--

LOCK TABLES `wf_node` WRITE;
/*!40000 ALTER TABLE `wf_node` DISABLE KEYS */;
/*!40000 ALTER TABLE `wf_node` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `wf_task`
--

DROP TABLE IF EXISTS `wf_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wf_task` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `WF_INSTANCE_ID` int unsigned NOT NULL COMMENT '流程实例ID',
  `WF_NODE_ID` int unsigned NOT NULL COMMENT '流程节点ID',
  `ASSIGNEE_ID` int unsigned DEFAULT NULL COMMENT '任务执行人',
  `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '状态(pending:待处理,completed:已完成,rejected:已拒绝,transferred:已转交)',
  `ACTION` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '操作(approve:同意,reject:拒绝,transfer:转交)',
  `COMMENT` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '审批意见',
  `CLAIM_TIME` datetime DEFAULT NULL COMMENT '签收时间',
  `COMPLETE_TIME` datetime DEFAULT NULL COMMENT '完成时间',
  `DUE_TIME` datetime DEFAULT NULL COMMENT '截止时间',
  `PRIORITY` int DEFAULT '0' COMMENT '优先级',
  `VARIABLES` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '任务变量(JSON)',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_wf_task_inst` (`WF_INSTANCE_ID`) USING BTREE,
  KEY `idx_wf_task_node` (`WF_NODE_ID`) USING BTREE,
  KEY `idx_wf_task_assignee` (`ASSIGNEE_ID`) USING BTREE,
  KEY `idx_wf_task_status` (`STATUS`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='工作流任务';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `wf_task`
--

LOCK TABLES `wf_task` WRITE;
/*!40000 ALTER TABLE `wf_task` DISABLE KEYS */;
/*!40000 ALTER TABLE `wf_task` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `wf_transition`
--

DROP TABLE IF EXISTS `wf_transition`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wf_transition` (
  `ID` int unsigned NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int unsigned DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `WF_DEFINITION_ID` int unsigned NOT NULL COMMENT '所属流程定义',
  `FROM_NODE_ID` int unsigned NOT NULL COMMENT '起始节点',
  `TO_NODE_ID` int unsigned NOT NULL COMMENT '目标节点',
  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '流转名称',
  `CONDITION` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '流转条件表达式',
  `ORDERNO` int DEFAULT NULL COMMENT '优先级顺序',
  PRIMARY KEY (`ID`) USING BTREE,
  KEY `idx_wf_trans_def` (`WF_DEFINITION_ID`) USING BTREE,
  KEY `idx_wf_trans_from` (`FROM_NODE_ID`) USING BTREE,
  KEY `idx_wf_trans_to` (`TO_NODE_ID`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='工作流流转';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `wf_transition`
--

LOCK TABLES `wf_transition` WRITE;
/*!40000 ALTER TABLE `wf_transition` DISABLE KEYS */;
/*!40000 ALTER TABLE `wf_transition` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping routines for database 'skyserver'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-04-10  2:31:37
