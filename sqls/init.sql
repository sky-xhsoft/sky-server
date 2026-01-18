-- =================================================================
-- Sky-Server Complete Database Initialization Script
-- =================================================================
-- This script contains everything needed to set up Sky-Server:
--   1. Database creation
--   2. All table schemas
--   3. Initial test data (company + admin user)
--
-- Usage:
--   mysql -u root -p < sqls/init.sql
--
-- Default credentials after initialization:
--   Username: admin
--   Password: admin123
-- =================================================================

-- Create database
CREATE DATABASE IF NOT EXISTS skyserver CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE skyserver;

-- =================================================================
-- TABLE SCHEMAS
-- =================================================================


SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for sys_action
-- ----------------------------
DROP TABLE IF EXISTS `sys_action`;
CREATE TABLE `sys_action`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '动作名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示描述',
  `DISPLAY_TYPE` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示样式(list_button:列表栏按钮,list_menu_item:列表栏菜单,obj_button:单对象界面按钮,obj_menu_item:单对象界面菜单,tab_button:单对象标签页按钮)',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '排序',
  `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '所属表单',
  `FILTER` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示条件',
  `ACTION_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '动作类型(url:URL,sp:存储过程,job:任务程序,js:JavaScript,bsh: OS Shell,py:Python,)',
  `CONTENT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '动作内容',
  `SCRIPTS` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '脚本(javascript将直接部署到页面上)',
  `URLTARGET` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'URL目标页(_blank or div id 去哪里显示url内容)',
  `SAVE_OBJ` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '保存修改(针对ObjButton/ObjMenuItem/TabButton)',
  `COMMENTS` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '提醒 (如果有内容，针对Button和MenuItem, not ListXXX and TreeNode)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '动作定义' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_action
-- ----------------------------

-- ----------------------------
-- Table structure for sys_column
-- ----------------------------
DROP TABLE IF EXISTS `sys_column`;
CREATE TABLE `sys_column`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT 0 COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `MASK` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '字段读写规则',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '序号',
  `DB_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字段名称',
  `COL_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字段类型(varchar,datetime,int,decimal,float,char,datenumber,date)',
  `COL_LENGTH` int NULL DEFAULT NULL COMMENT '字段长度',
  `COL_PRECISION` int NULL DEFAULT NULL COMMENT '字段精度',
  `SYS_TABLE_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属表单',
  `IS_DK` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'N' COMMENT '显示键(DK)',
  `IS_AK` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '输入键(AK)',
  `NULL_ABLE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'Y' COMMENT '空值(Y: 是,N: 否)',
  `IS_UPPERCASE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'N' COMMENT '是否大写(Y:是,N:否)',
  `IS_QUERY` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'N' COMMENT '是否查询条件',
  `SUBMETHOD` varchar(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '统计方法(sum:求和)',
  `FULL_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '字段全名',
  `MODIFI_ABLE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '允许界面修改',
  `SET_VALUE_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '赋值方式(pk:pk,docno:单据编号,createBy:创建人,byPage:界面输入,select:下拉选项,fk:外键关联,sysdate:操作时间,operator:操作用户,ignore:忽略)',
  `REF_TABLE_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '关联表id',
  `REF_COLUMN_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '关联字段id',
  `REF_ON_DELETE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '外键删除动作(noAction:无动作)',
  `SEQ` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '单据编号生成器',
  `SYS_DICT_ID` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '数据字典',
  `DEFAULT_VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '默认值',
  `REG_EXPRESSION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '输入校验正则',
  `ERR_MSG` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '正则校验失败提醒',
  `FILTER` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '字段过滤器(sql)',
  `DISPLAY_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示控件(blank,button,hr,check,file,image,select,text,textarea,date,datetime)',
  `DISPLAY_COLS` int NULL DEFAULT NULL COMMENT '显示列数',
  `DISPLAY_ROWS` int NULL DEFAULT NULL COMMENT '显示行数',
  `PROPS` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '扩展属性',
  `IS_SHOW_TITLE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'Y' COMMENT '是否显示备注(Y:是,N:否)',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `SHOW_COLUMN_ID` int NULL DEFAULT NULL COMMENT '级联显示字段',
  `SHOW_COLUMN_VAL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '级联显示条件',
  `HR_COLUMN_ID` int NULL DEFAULT NULL COMMENT '关联HR折叠字段',
  `SGRADE` int NULL DEFAULT NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE INDEX `idx_cloumn_full_name`(`FULL_NAME` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '系统表字段' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_column
-- ----------------------------

-- ----------------------------
-- Table structure for sys_company
-- ----------------------------
DROP TABLE IF EXISTS `sys_company`;
CREATE TABLE `sys_company`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '公司名称',
  `DOMAIN` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '公司域名（用于多租户识别）',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE INDEX `idx_domain`(`DOMAIN` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '模板表单' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_company
-- ----------------------------

-- ----------------------------
-- Table structure for sys_dict
-- ----------------------------
DROP TABLE IF EXISTS `sys_dict`;
CREATE TABLE `sys_dict`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字典名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `TYPE` int NULL DEFAULT 0 COMMENT '字段类型(0: String, 1: int)',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `DEFAULT_VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '默认值',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '数据字典' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_dict
-- ----------------------------

-- ----------------------------
-- Table structure for sys_dict_item
-- ----------------------------
DROP TABLE IF EXISTS `sys_dict_item`;
CREATE TABLE `sys_dict_item`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_DICT_ID` int UNSIGNED NOT NULL COMMENT '所属字典',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '字典值',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '排序',
  `CSSCLASS` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'css',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `IS_DEFAULT_VALUE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '是否默认值(Y:是,N:否)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '数据字典明细' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_dict_item
-- ----------------------------

-- ----------------------------
-- Table structure for sys_directory
-- ----------------------------
DROP TABLE IF EXISTS `sys_directory`;
CREATE TABLE `sys_directory`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示名称',
  `SYS_TABLE_CATEGORY_ID` int NULL DEFAULT NULL COMMENT '所属表类别',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '位置',
  `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '对应表',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '安全目录' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_directory
-- ----------------------------

-- ----------------------------
-- Table structure for sys_group_prem
-- ----------------------------
DROP TABLE IF EXISTS `sys_group_prem`;
CREATE TABLE `sys_group_prem`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_GROUPS_ID` int NULL DEFAULT NULL COMMENT '权限组',
  `SYS_DIRECTORY_ID` int NULL DEFAULT NULL COMMENT '目录\r\n',
  `PERMISSION` int NULL DEFAULT NULL COMMENT '权限(1:读;3:读,写;5:读,提交;……)',
  `FILTER_OBJ` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '数据过滤({sql:\"\",display:\"\",other:\"\"})',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '权限组明细' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_group_prem
-- ----------------------------

-- ----------------------------
-- Table structure for sys_groups
-- ----------------------------
DROP TABLE IF EXISTS `sys_groups`;
CREATE TABLE `sys_groups`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '名称',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  `SGRADE` int NULL DEFAULT NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '权限组' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_groups
-- ----------------------------

-- ----------------------------
-- Table structure for sys_model
-- ----------------------------
DROP TABLE IF EXISTS `sys_model`;
CREATE TABLE `sys_model`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '模板表单' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_model
-- ----------------------------

-- ----------------------------
-- Table structure for sys_objuiconf
-- ----------------------------
DROP TABLE IF EXISTS `sys_objuiconf`;
CREATE TABLE `sys_objuiconf`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示名称',
  `TABLE_PARAM_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'tableid参数名',
  `PK_PARAM_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'id参数名',
  `CSS_CLASS` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'CSS类',
  `COLS` int NULL DEFAULT NULL COMMENT '每行字段个数',
  `DEFAULT_ACTION` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '缺省动作',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '对象显示配置' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_objuiconf
-- ----------------------------

-- ----------------------------
-- Table structure for sys_param
-- ----------------------------
DROP TABLE IF EXISTS `sys_param`;
CREATE TABLE `sys_param`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '名称',
  `DEFAULT_VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '默认值',
  `VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '当前值',
  `VALUE_TYPE` char(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '值类型',
  `VALUE_LIST` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '值列表',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '模板表单' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_param
-- ----------------------------

-- ----------------------------
-- Table structure for sys_seq
-- ----------------------------
DROP TABLE IF EXISTS `sys_seq`;
CREATE TABLE `sys_seq`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示名称',
  `VFORMAT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '格式',
  `INCRE` int NULL DEFAULT NULL COMMENT '递增',
  `CYCLETYPE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '循环方式',
  `PREFIX` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '前缀',
  `SUFFIX` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '后缀',
  `CUR_DATE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '当前周期值',
  `CUR_NUM` int NULL DEFAULT NULL COMMENT '当前流水号',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '序号生成器' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_seq
-- ----------------------------

-- ----------------------------
-- Table structure for sys_subsystem
-- ----------------------------
DROP TABLE IF EXISTS `sys_subsystem`;
CREATE TABLE `sys_subsystem`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '名称',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '序号',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '网页链接',
  `ICON` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'icon',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '子系统' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_subsystem
-- ----------------------------

-- ----------------------------
-- Table structure for sys_table
-- ----------------------------
DROP TABLE IF EXISTS `sys_table`;
CREATE TABLE `sys_table`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '名称',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '显示名称',
  `REAL_TABLE_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '实际数据库表',
  `FILTER` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '数据过滤SQL',
  `DK_COLUMN_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '显示主键(DK)',
  `AK_COLUMN_ID` int NULL DEFAULT NULL COMMENT '输入主键(AK)',
  `MASK` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '表单规则(支持：A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废)',
  `SYS_TABLECATEGORY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '表类别',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '排序',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '网页连接',
  `RPC_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'rpc 方法',
  `IS_MENU` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'N' COMMENT '是否菜单(Y:是,N:否)',
  `ICO_IMG` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '表单ICO图片',
  `IS_DROPDOWN` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '是否下拉框',
  `SYS_OBJUICONF_ID` int NULL DEFAULT NULL COMMENT '显示配置',
  `SYS_DIRECTORY_ID` int NULL DEFAULT NULL COMMENT '安全目录',
  `SYS_PARENT_TABLE_ID` int NULL DEFAULT NULL COMMENT '父表',
  `ROWCNT` int NULL DEFAULT NULL COMMENT '统计行数',
  `IS_BIG` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '是否海量',
  `PROPS` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '扩展属性',
  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE INDEX `IDX_SYSTABLE_NAME`(`NAME` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '系统表单' ROW_FORMAT = DYNAMIC;


-- ----------------------------
-- Table structure for sys_table_category
-- ----------------------------
DROP TABLE IF EXISTS `sys_table_category`;
CREATE TABLE `sys_table_category`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_SUBSYSTEM_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属子系统',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '名称',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '序号',
  `ICON` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'icon图标',
  `URL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '网页连接',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE,
  INDEX `idx_subsystem`(`SYS_SUBSYSTEM_ID` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '表类别' ROW_FORMAT = DYNAMIC;


-- ----------------------------
-- Table structure for sys_table_cmd
-- ----------------------------
CREATE TABLE `sys_table_cmd`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '所属表单',
  `ACTION_TYPE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '按钮类型(1:系统按钮)',
  `ACTION` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '按钮(A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废,I:导入,E:导出)',
  `ACTION_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '按钮名称',
  `EVENT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '事件前后(begin:开始,end:结束)',
  `CONTENT` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '执行操作(存储过程/action动作)',
  `CONTENT_TYPE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '动作类型',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '序号',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '表单功能扩展' ROW_FORMAT = DYNAMIC;


-- ----------------------------
-- Table structure for sys_table_ref
-- ----------------------------
DROP TABLE IF EXISTS `sys_table_ref`;
CREATE TABLE `sys_table_ref`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '主表',
  `ORDERNO` int NULL DEFAULT NULL COMMENT '序号',
  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示描述',
  `REF_TABLE_ID` int NULL DEFAULT NULL COMMENT '关联表',
  `REF_COLUMN_ID` int NULL DEFAULT NULL COMMENT '关联字段',
  `FILTER` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '过滤条件',
  `ASSOCTYPE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '关联方式(1 : 1对1, n: 1对n )',
  `EDIT_TYPE` char(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '编辑方式(Y:标准(新增和修改行时都可在内嵌窗口编辑),\r\nN:无(无内嵌编辑窗口),NP:非内嵌，允许弹出,NS:非内嵌，禁止弹出,A:仅显示新增字段，修改直接修改)',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '关联表' ROW_FORMAT = DYNAMIC;


-- ----------------------------
-- Table structure for sys_table_sql
-- ----------------------------
DROP TABLE IF EXISTS `sys_table_sql`;
CREATE TABLE `sys_table_sql`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '所属表单',
  `SQL` varchar(5000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '表单sql',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '表单sql\r\n' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_table_sql
-- ----------------------------

-- ----------------------------
-- Table structure for sys_user
-- ----------------------------
DROP TABLE IF EXISTS `sys_user`;
CREATE TABLE `sys_user`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `TRUE_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '真实名称',
  `USERNAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户名称',
  `PASSWORD` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '密码',
  `PHONE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '手机号',
  `EMAIL` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '邮箱',
  `LANGUAGE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '语言',
  `IS_ADMIN` char(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'Y' COMMENT '是否管理员',
  `SGRADE` int NULL DEFAULT NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '系统用户' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_user
-- ----------------------------

-- ----------------------------
-- Table structure for sys_user_env
-- ----------------------------
DROP TABLE IF EXISTS `sys_user_env`;
CREATE TABLE `sys_user_env`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '变量名称',
  `VALUE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '值来源',
  `DESCRIPTION` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户环境变量' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_user_env
-- ----------------------------

-- ----------------------------
-- Table structure for sys_user_groups
-- ----------------------------
DROP TABLE IF EXISTS `sys_user_groups`;
CREATE TABLE `sys_user_groups`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
  `SYS_USER_ID` int NULL DEFAULT NULL COMMENT '用户',
  `SYS_DIRECTORY_ID` int NULL DEFAULT NULL COMMENT '权限组',
  PRIMARY KEY (`ID`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户权限组' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of sys_user_groups
-- ----------------------------



-- 审计日志表

-- ----------------------------
-- Table structure for audit_log
-- ----------------------------
DROP TABLE IF EXISTS `audit_log`;
CREATE TABLE `audit_log`  (
                              `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                              `USER_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '操作用户ID',
                              `USERNAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '操作用户名',
                              `ACTION` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '操作类型(login,logout,create,update,delete等)',
                              `RESOURCE` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '资源类型(user,table,action,workflow等)',
                              `RESOURCE_ID` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '资源ID',
                              `RESOURCE_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '资源名称',
                              `METHOD` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'HTTP方法',
                              `PATH` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '请求路径',
                              `IP` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '客户端IP',
                              `USER_AGENT` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '用户代理',
                              `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '操作状态(success,failure)',
                              `ERROR_MESSAGE` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '错误信息',
                              `REQUEST_BODY` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '请求体',
                              `RESPONSE_BODY` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '响应体',
                              `OLD_VALUE` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '修改前的值(JSON)',
                              `NEW_VALUE` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '修改后的值(JSON)',
                              `DURATION` bigint NULL DEFAULT NULL COMMENT '执行时长(毫秒)',
                              `TAGS` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '标签(用于分类和搜索)',
                              `CREATED_AT` datetime NULL DEFAULT NULL COMMENT '创建时间',
                              `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
                              PRIMARY KEY (`ID`) USING BTREE,
                              INDEX `idx_audit_user`(`USER_ID` ASC) USING BTREE,
                              INDEX `idx_audit_action`(`ACTION` ASC) USING BTREE,
                              INDEX `idx_audit_resource`(`RESOURCE` ASC) USING BTREE,
                              INDEX `idx_audit_resource_id`(`RESOURCE_ID` ASC) USING BTREE,
                              INDEX `idx_audit_status`(`STATUS` ASC) USING BTREE,
                              INDEX `idx_audit_created`(`CREATED_AT` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '审计日志' ROW_FORMAT = DYNAMIC;



CREATE TABLE IF NOT EXISTS `sys_user_session` (
                                                  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                                                  `USER_ID` int UNSIGNED NOT NULL COMMENT '用户ID',
                                                  `COMPANY_ID` int UNSIGNED NOT NULL COMMENT '公司ID',
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
    INDEX `idx_session_user` (`USER_ID`) USING BTREE,
    INDEX `idx_session_token` (`TOKEN`(255)) USING BTREE,
    UNIQUE INDEX `idx_session_device` (`DEVICE_ID`) USING BTREE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户会话表';


-- =============================================
-- 菜单管理表
-- =============================================

-- =============================================
-- 消息和文件管理表（Phase 14 & 15）
-- =============================================

-- =============================================
-- 文件管理表
-- =============================================

-- sys_file 系统文件表
DROP TABLE IF EXISTS `sys_file`;
CREATE TABLE `sys_file` (
                            `ID` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                            `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '公司ID',
                            `CREATE_BY` varchar(80) NULL COMMENT '创建人',
                            `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                            `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
                            `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                            `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',

                            `FILE_NAME` varchar(255) NOT NULL COMMENT '原始文件名',
                            `STORAGE_NAME` varchar(255) NOT NULL COMMENT '存储文件名（唯一）',
                            `FILE_PATH` varchar(500) NOT NULL COMMENT '文件路径',
                            `FILE_SIZE` bigint NOT NULL COMMENT '文件大小（字节）',
                            `FILE_TYPE` varchar(100) NULL COMMENT '文件类型/MIME类型',
                            `FILE_EXT` varchar(20) NULL COMMENT '文件扩展名',
                            `STORAGE_TYPE` varchar(20) NOT NULL DEFAULT 'local' COMMENT '存储类型：local, oss, s3',
                            `BUCKET_NAME` varchar(100) NULL COMMENT '存储桶名称（云存储）',
                            `ACCESS_URL` varchar(500) NULL COMMENT '访问URL',
                            `THUMBNAIL_URL` varchar(500) NULL COMMENT '缩略图URL',
                            `MD5` varchar(32) NULL COMMENT '文件MD5值',
                            `UPLOAD_IP` varchar(50) NULL COMMENT '上传IP',
                            `DOWNLOAD_COUNT` int NOT NULL DEFAULT 0 COMMENT '下载次数',
                            `CATEGORY` varchar(50) NULL COMMENT '文件分类',
                            `DESCRIPTION` varchar(500) NULL COMMENT '文件描述',
                            `EXPIRE_TIME` datetime NULL COMMENT '过期时间',

                            PRIMARY KEY (`ID`),
                            INDEX `idx_storage_name` (`STORAGE_NAME`),
                            INDEX `idx_md5` (`MD5`),
                            INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统文件表';

-- =============================================
-- 消息管理表
-- =============================================

-- sys_message 系统消息表
DROP TABLE IF EXISTS `sys_message`;
CREATE TABLE `sys_message` (
                               `ID` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '公司ID',
                               `CREATE_BY` varchar(80) NULL COMMENT '创建人',
                               `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                               `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
                               `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                               `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',

                               `TITLE` varchar(255) NOT NULL COMMENT '消息标题',
                               `CONTENT` text NOT NULL COMMENT '消息内容',
                               `MESSAGE_TYPE` varchar(50) NOT NULL COMMENT '消息类型: system, workflow, business, notice',
                               `PRIORITY` int NOT NULL DEFAULT 0 COMMENT '优先级: 0=普通, 1=重要, 2=紧急',
                               `CATEGORY` varchar(50) NULL COMMENT '消息分类',
                               `SENDER_ID` int UNSIGNED NULL COMMENT '发送者ID（系统消息为NULL）',
                               `SENDER_NAME` varchar(100) NULL COMMENT '发送者姓名',
                               `TARGET_TYPE` varchar(20) NOT NULL DEFAULT 'user' COMMENT '目标类型: user, role, group, all',
                               `TARGET_IDS` varchar(1000) NULL COMMENT '目标ID列表（逗号分隔）',
                               `LINK_URL` varchar(500) NULL COMMENT '关联URL',
                               `LINK_TYPE` varchar(50) NULL COMMENT '链接类型: internal, external',
                               `PARAMS` text NULL COMMENT '消息参数（JSON）',
                               `TEMPLATE_ID` int UNSIGNED NULL COMMENT '消息模板ID',
                               `READ_COUNT` int NOT NULL DEFAULT 0 COMMENT '已读人数',
                               `TOTAL_COUNT` int NOT NULL DEFAULT 0 COMMENT '总接收人数',
                               `EXPIRE_TIME` datetime NULL COMMENT '过期时间',
                               `STATUS` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态: active, expired, deleted',

                               PRIMARY KEY (`ID`),
                               INDEX `idx_message_type` (`MESSAGE_TYPE`),
                               INDEX `idx_priority` (`PRIORITY`),
                               INDEX `idx_category` (`CATEGORY`),
                               INDEX `idx_sender` (`SENDER_ID`),
                               INDEX `idx_template` (`TEMPLATE_ID`),
                               INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统消息表';

-- sys_user_message 用户消息关联表
DROP TABLE IF EXISTS `sys_user_message`;
CREATE TABLE `sys_user_message` (
                                    `ID` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                                    `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '公司ID',
                                    `CREATE_BY` varchar(80) NULL COMMENT '创建人',
                                    `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                    `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
                                    `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                    `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',

                                    `MESSAGE_ID` int UNSIGNED NOT NULL COMMENT '消息ID',
                                    `USER_ID` int UNSIGNED NOT NULL COMMENT '用户ID',
                                    `IS_READ` char(1) NOT NULL DEFAULT 'N' COMMENT '是否已读 Y/N',
                                    `READ_TIME` datetime NULL COMMENT '读取时间',
                                    `IS_STARRED` char(1) NOT NULL DEFAULT 'N' COMMENT '是否星标 Y/N',
                                    `IS_ARCHIVED` char(1) NOT NULL DEFAULT 'N' COMMENT '是否归档 Y/N',
                                    `DELETED_AT` datetime NULL COMMENT '删除时间（软删除）',

                                    PRIMARY KEY (`ID`),
                                    INDEX `idx_user_msg` (`USER_ID`, `MESSAGE_ID`),
                                    INDEX `idx_is_read` (`IS_READ`),
                                    INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户消息关联表';

-- sys_message_template 消息模板表
DROP TABLE IF EXISTS `sys_message_template`;
CREATE TABLE `sys_message_template` (
                                        `ID` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                                        `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '公司ID',
                                        `CREATE_BY` varchar(80) NULL COMMENT '创建人',
                                        `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                        `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
                                        `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                        `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',

                                        `CODE` varchar(50) NOT NULL COMMENT '模板代码',
                                        `NAME` varchar(100) NOT NULL COMMENT '模板名称',
                                        `MESSAGE_TYPE` varchar(50) NOT NULL COMMENT '消息类型',
                                        `TITLE` varchar(255) NOT NULL COMMENT '标题模板',
                                        `CONTENT` text NOT NULL COMMENT '内容模板',
                                        `VARIABLES` varchar(500) NULL COMMENT '变量列表（逗号分隔）',
                                        `DESCRIPTION` varchar(500) NULL COMMENT '描述',
                                        `IS_ENABLED` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用 Y/N',
                                        `CATEGORY` varchar(50) NULL COMMENT '分类',

                                        PRIMARY KEY (`ID`),
                                        UNIQUE INDEX `idx_code` (`CODE`),
                                        INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息模板表';

-- sys_email_config 邮件配置表
DROP TABLE IF EXISTS `sys_email_config`;
CREATE TABLE `sys_email_config` (
                                    `ID` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                                    `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '公司ID',
                                    `CREATE_BY` varchar(80) NULL COMMENT '创建人',
                                    `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                    `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
                                    `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                    `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',

                                    `SMTP_HOST` varchar(100) NOT NULL COMMENT 'SMTP服务器地址',
                                    `SMTP_PORT` int NOT NULL COMMENT 'SMTP端口',
                                    `SMTP_USER` varchar(100) NOT NULL COMMENT 'SMTP用户名',
                                    `SMTP_PASSWORD` varchar(255) NOT NULL COMMENT 'SMTP密码（加密存储）',
                                    `FROM_EMAIL` varchar(100) NOT NULL COMMENT '发件人邮箱',
                                    `FROM_NAME` varchar(100) NULL COMMENT '发件人名称',
                                    `USE_TLS` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否使用TLS Y/N',
                                    `IS_DEFAULT` char(1) NOT NULL DEFAULT 'N' COMMENT '是否默认配置 Y/N',
                                    `DESCRIPTION` varchar(500) NULL COMMENT '描述',

                                    PRIMARY KEY (`ID`),
                                    INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件配置表';

-- sys_notification_log 通知日志表
DROP TABLE IF EXISTS `sys_notification_log`;
CREATE TABLE `sys_notification_log` (
                                        `ID` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                                        `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '公司ID',
                                        `CREATE_BY` varchar(80) NULL COMMENT '创建人',
                                        `CREATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                        `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
                                        `UPDATE_TIME` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                        `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y/N)',

                                        `MESSAGE_ID` int UNSIGNED NULL COMMENT '消息ID',
                                        `USER_ID` int UNSIGNED NULL COMMENT '接收用户ID',
                                        `NOTIFY_TYPE` varchar(20) NOT NULL COMMENT '通知类型: websocket, email, sms',
                                        `STATUS` varchar(20) NOT NULL COMMENT '状态: pending, sent, failed, read',
                                        `SENT_TIME` datetime NULL COMMENT '发送时间',
                                        `READ_TIME` datetime NULL COMMENT '读取时间',
                                        `ERROR_MESSAGE` varchar(500) NULL COMMENT '错误信息',
                                        `RETRY_COUNT` int NOT NULL DEFAULT 0 COMMENT '重试次数',

                                        PRIMARY KEY (`ID`),
                                        INDEX `idx_message` (`MESSAGE_ID`),
                                        INDEX `idx_user` (`USER_ID`),
                                        INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知日志表';


-- 工作流相关表

-- ----------------------------
-- Table structure for wf_definition
-- ----------------------------
DROP TABLE IF EXISTS `wf_definition`;
CREATE TABLE `wf_definition`  (
                                  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                                  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
                                  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
                                  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
                                  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
                                  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
                                  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
                                  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '流程名称',
                                  `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示名称',
                                  `VERSION` int NOT NULL DEFAULT 1 COMMENT '版本号',
                                  `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'draft' COMMENT '状态(draft:草稿,published:已发布,archived:已归档)',
                                  `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '关联的业务表',
                                  `DESCRIPTION` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '描述',
                                  `CONFIG` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT 'JSON配置',
                                  PRIMARY KEY (`ID`) USING BTREE,
                                  INDEX `idx_wf_def_table`(`SYS_TABLE_ID` ASC) USING BTREE,
                                  INDEX `idx_wf_def_status`(`STATUS` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '工作流定义' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for wf_node
-- ----------------------------
DROP TABLE IF EXISTS `wf_node`;
CREATE TABLE `wf_node`  (
                            `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                            `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
                            `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
                            `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
                            `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
                            `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
                            `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
                            `WF_DEFINITION_ID` int UNSIGNED NOT NULL COMMENT '所属流程定义',
                            `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '节点名称',
                            `DISPLAY_NAME` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '显示名称',
                            `NODE_TYPE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '节点类型(start:开始,end:结束,user:用户任务,auto:自动任务,gateway:网关)',
                            `ASSIGN_TYPE` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '分配类型(user:指定用户,starter:发起人,role:角色,expression:表达式)',
                            `ASSIGN_VALUE` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '分配值',
                            `ACTION_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '自动任务关联的动作ID',
                            `CONFIG` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT 'JSON配置',
                            `POS_X` int NULL DEFAULT NULL COMMENT '节点X坐标',
                            `POS_Y` int NULL DEFAULT NULL COMMENT '节点Y坐标',
                            PRIMARY KEY (`ID`) USING BTREE,
                            INDEX `idx_wf_node_def`(`WF_DEFINITION_ID` ASC) USING BTREE,
                            INDEX `idx_wf_node_action`(`ACTION_ID` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '工作流节点' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for wf_transition
-- ----------------------------
DROP TABLE IF EXISTS `wf_transition`;
CREATE TABLE `wf_transition`  (
                                  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                                  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
                                  `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
                                  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
                                  `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
                                  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
                                  `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
                                  `WF_DEFINITION_ID` int UNSIGNED NOT NULL COMMENT '所属流程定义',
                                  `FROM_NODE_ID` int UNSIGNED NOT NULL COMMENT '起始节点',
                                  `TO_NODE_ID` int UNSIGNED NOT NULL COMMENT '目标节点',
                                  `NAME` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '流转名称',
                                  `CONDITION` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '流转条件表达式',
                                  `ORDERNO` int NULL DEFAULT NULL COMMENT '优先级顺序',
                                  PRIMARY KEY (`ID`) USING BTREE,
                                  INDEX `idx_wf_trans_def`(`WF_DEFINITION_ID` ASC) USING BTREE,
                                  INDEX `idx_wf_trans_from`(`FROM_NODE_ID` ASC) USING BTREE,
                                  INDEX `idx_wf_trans_to`(`TO_NODE_ID` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '工作流流转' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for wf_instance
-- ----------------------------
DROP TABLE IF EXISTS `wf_instance`;
CREATE TABLE `wf_instance`  (
                                `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                                `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
                                `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
                                `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
                                `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
                                `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
                                `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
                                `WF_DEFINITION_ID` int UNSIGNED NOT NULL COMMENT '流程定义ID',
                                `SYS_TABLE_ID` int NULL DEFAULT NULL COMMENT '关联的业务表',
                                `BUSINESS_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '业务数据ID',
                                `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '状态(running:运行中,completed:已完成,terminated:已终止,suspended:已挂起)',
                                `CURRENT_NODE_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '当前节点ID',
                                `START_USER_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '发起人',
                                `START_TIME` datetime NULL DEFAULT NULL COMMENT '开始时间',
                                `END_TIME` datetime NULL DEFAULT NULL COMMENT '结束时间',
                                `VARIABLES` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '流程变量(JSON)',
                                `TITLE` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '流程标题',
                                PRIMARY KEY (`ID`) USING BTREE,
                                INDEX `idx_wf_inst_def`(`WF_DEFINITION_ID` ASC) USING BTREE,
                                INDEX `idx_wf_inst_biz`(`BUSINESS_ID` ASC) USING BTREE,
                                INDEX `idx_wf_inst_status`(`STATUS` ASC) USING BTREE,
                                INDEX `idx_wf_inst_node`(`CURRENT_NODE_ID` ASC) USING BTREE,
                                INDEX `idx_wf_inst_user`(`START_USER_ID` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '工作流实例' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for wf_task
-- ----------------------------
DROP TABLE IF EXISTS `wf_task`;
CREATE TABLE `wf_task`  (
                            `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
                            `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
                            `CREATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '创建人',
                            `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
                            `UPDATE_BY` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '更新人',
                            `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
                            `IS_ACTIVE` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Y' COMMENT '是否有效(Y:可用,N:不可用)',
                            `WF_INSTANCE_ID` int UNSIGNED NOT NULL COMMENT '流程实例ID',
                            `WF_NODE_ID` int UNSIGNED NOT NULL COMMENT '流程节点ID',
                            `ASSIGNEE_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '任务执行人',
                            `STATUS` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '状态(pending:待处理,completed:已完成,rejected:已拒绝,transferred:已转交)',
                            `ACTION` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '操作(approve:同意,reject:拒绝,transfer:转交)',
                            `COMMENT` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '审批意见',
                            `CLAIM_TIME` datetime NULL DEFAULT NULL COMMENT '签收时间',
                            `COMPLETE_TIME` datetime NULL DEFAULT NULL COMMENT '完成时间',
                            `DUE_TIME` datetime NULL DEFAULT NULL COMMENT '截止时间',
                            `PRIORITY` int NULL DEFAULT 0 COMMENT '优先级',
                            `VARIABLES` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '任务变量(JSON)',
                            PRIMARY KEY (`ID`) USING BTREE,
                            INDEX `idx_wf_task_inst`(`WF_INSTANCE_ID` ASC) USING BTREE,
                            INDEX `idx_wf_task_node`(`WF_NODE_ID` ASC) USING BTREE,
                            INDEX `idx_wf_task_assignee`(`ASSIGNEE_ID` ASC) USING BTREE,
                            INDEX `idx_wf_task_status`(`STATUS` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '工作流任务' ROW_FORMAT = DYNAMIC;


SET FOREIGN_KEY_CHECKS = 1;

-- =================================================================
-- INITIAL TEST DATA
-- =================================================================


    -- =====================================================
-- 云盘分片上传表结构
-- 功能：支持大文件分片上传和断点续传
-- 创建时间：2026-01-15
-- =====================================================
DROP TABLE IF EXISTS `cloud_upload_session`;
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

DROP TABLE IF EXISTS `cloud_chunk_record`;
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


-- Create test company (ID=1)
INSERT INTO sys_company (ID, SYS_COMPANY_ID, NAME, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, '测试公司', 'Y', 1, NOW(), 1, NOW())
ON DUPLICATE KEY UPDATE NAME = '测试公司';

-- Create admin user (ID=1)
-- Password: admin123 (bcrypt hash)
INSERT INTO sys_user (ID, SYS_COMPANY_ID, USERNAME, PASSWORD, TRUE_NAME, EMAIL, PHONE,
                      IS_ACTIVE, SGRADE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, 'admin',
        '$2a$10$iztoR7MeHZKyoBNpJM4pjOZ729KAoy.5x5ayetl1Rnb3TBgVCO0jy',
        '系统管理员', 'admin@example.com', '13800138000',
        'Y', 99, 1, NOW(), 1, NOW())
ON DUPLICATE KEY UPDATE PASSWORD = '$2a$10$iztoR7MeHZKyoBNpJM4pjOZ729KAoy.5x5ayetl1Rnb3TBgVCO0jy';

-- =================================================================
-- VERIFICATION
-- =================================================================

SELECT 'Database initialization completed!' AS Status;
SELECT CONCAT('Tables created: ', COUNT(*)) AS Result
FROM information_schema.tables
WHERE table_schema = 'skyserver';

SELECT 'Default admin user created' AS Info,
       'Username: admin' AS Username,
       'Password: admin123' AS Password;




-- ==========================================
-- 云盘系统表结构
-- ==========================================

-- 1. 云盘文件夹表
DROP TABLE IF EXISTS `cloud_folder`;
CREATE TABLE `cloud_folder` (
                                `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                `NAME` VARCHAR(255) NOT NULL COMMENT '文件夹名称',
                                `PARENT_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '父文件夹ID',
                                `PATH` VARCHAR(1000) NOT NULL COMMENT '完整路径',
                                `OWNER_ID` BIGINT UNSIGNED NOT NULL COMMENT '所有者ID',
                                `IS_PUBLIC` CHAR(1) DEFAULT 'N' COMMENT '是否公开 Y/N',
                                `SHARE_CODE` VARCHAR(50) DEFAULT NULL COMMENT '分享码',
                                `SHARE_EXPIRE` DATETIME DEFAULT NULL COMMENT '分享过期时间',
                                `DESCRIPTION` VARCHAR(500) DEFAULT NULL COMMENT '描述',
                                `FILE_COUNT` INT DEFAULT 0 COMMENT '文件数量',
                                `TOTAL_SIZE` BIGINT DEFAULT 0 COMMENT '总大小（字节）',
                                `SYS_COMPANY_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '公司ID',
                                `CREATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '创建人',
                                `CREATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                `UPDATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '更新人',
                                `UPDATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                `IS_ACTIVE` CHAR(1) DEFAULT 'Y' COMMENT '是否有效 Y/N',
                                PRIMARY KEY (`ID`),
                                UNIQUE KEY `uk_share_code` (`SHARE_CODE`),
                                KEY `idx_parent_id` (`PARENT_ID`),
                                KEY `idx_path` (`PATH`(255)),
                                KEY `idx_owner_id` (`OWNER_ID`),
                                KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘文件夹表';

-- 2. 云盘文件表
DROP TABLE IF EXISTS `cloud_file`;
CREATE TABLE `cloud_file` (
                              `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                              `FILE_NAME` VARCHAR(255) NOT NULL COMMENT '文件名',
                              `FOLDER_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '文件夹ID',
                              `PATH` VARCHAR(1000) NOT NULL COMMENT '完整路径',
                              `STORAGE_TYPE` VARCHAR(20) NOT NULL DEFAULT 'local' COMMENT '存储类型: local, oss',
                              `STORAGE_PATH` VARCHAR(500) NOT NULL COMMENT '存储路径',
                              `FILE_SIZE` BIGINT NOT NULL COMMENT '文件大小（字节）',
                              `FILE_TYPE` VARCHAR(100) DEFAULT NULL COMMENT '文件MIME类型',
                              `FILE_EXT` VARCHAR(20) DEFAULT NULL COMMENT '文件扩展名',
                              `MD5` VARCHAR(32) DEFAULT NULL COMMENT 'MD5值',
                              `OWNER_ID` BIGINT UNSIGNED NOT NULL COMMENT '所有者ID',
                              `IS_PUBLIC` CHAR(1) DEFAULT 'N' COMMENT '是否公开 Y/N',
                              `SHARE_CODE` VARCHAR(50) DEFAULT NULL COMMENT '分享码',
                              `SHARE_EXPIRE` DATETIME DEFAULT NULL COMMENT '分享过期时间',
                              `ACCESS_URL` VARCHAR(500) DEFAULT NULL COMMENT '访问URL',
                              `THUMBNAIL` VARCHAR(500) DEFAULT NULL COMMENT '缩略图URL',
                              `DOWNLOAD_COUNT` INT DEFAULT 0 COMMENT '下载次数',
                              `TAGS` VARCHAR(500) DEFAULT NULL COMMENT '标签（逗号分隔）',
                              `DESCRIPTION` VARCHAR(500) DEFAULT NULL COMMENT '描述',
                              `SYS_COMPANY_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '公司ID',
                              `CREATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '创建人',
                              `CREATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                              `UPDATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '更新人',
                              `UPDATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                              `IS_ACTIVE` CHAR(1) DEFAULT 'Y' COMMENT '是否有效 Y/N',
                              PRIMARY KEY (`ID`),
                              UNIQUE KEY `uk_share_code` (`SHARE_CODE`),
                              KEY `idx_folder_id` (`FOLDER_ID`),
                              KEY `idx_md5` (`MD5`),
                              KEY `idx_owner_id` (`OWNER_ID`),
                              KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘文件表';

-- 3. 云盘分享记录表
DROP TABLE IF EXISTS `cloud_share`;
CREATE TABLE `cloud_share` (
                               `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                               `SHARE_CODE` VARCHAR(50) NOT NULL COMMENT '分享码',
                               `RESOURCE_TYPE` VARCHAR(20) NOT NULL COMMENT '资源类型: file, folder',
                               `RESOURCE_ID` BIGINT UNSIGNED NOT NULL COMMENT '资源ID',
                               `SHARER_ID` BIGINT UNSIGNED NOT NULL COMMENT '分享者ID',
                               `SHARE_TYPE` VARCHAR(20) NOT NULL COMMENT '分享类型: public, password, private',
                               `PASSWORD` VARCHAR(50) DEFAULT NULL COMMENT '访问密码',
                               `EXPIRE_TIME` DATETIME DEFAULT NULL COMMENT '过期时间',
                               `MAX_DOWNLOADS` INT DEFAULT 0 COMMENT '最大下载次数（0=无限制）',
                               `DOWNLOAD_COUNT` INT DEFAULT 0 COMMENT '已下载次数',
                               `VIEW_COUNT` INT DEFAULT 0 COMMENT '查看次数',
                               `STATUS` VARCHAR(20) DEFAULT 'active' COMMENT '状态: active, expired, disabled',
                               `SYS_COMPANY_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '公司ID',
                               `CREATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '创建人',
                               `CREATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                               `UPDATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '更新人',
                               `UPDATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                               `IS_ACTIVE` CHAR(1) DEFAULT 'Y' COMMENT '是否有效 Y/N',
                               PRIMARY KEY (`ID`),
                               UNIQUE KEY `uk_share_code` (`SHARE_CODE`),
                               KEY `idx_resource_id` (`RESOURCE_ID`),
                               KEY `idx_sharer_id` (`SHARER_ID`),
                               KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘分享记录表';

-- 4. 云盘配额表
DROP TABLE IF EXISTS `cloud_quota`;
CREATE TABLE `cloud_quota` (
                               `ID` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                               `USER_ID` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                               `TOTAL_QUOTA` BIGINT NOT NULL COMMENT '总配额（字节）',
                               `USED_SPACE` BIGINT DEFAULT 0 COMMENT '已用空间（字节）',
                               `FILE_COUNT` INT DEFAULT 0 COMMENT '文件数量',
                               `FOLDER_COUNT` INT DEFAULT 0 COMMENT '文件夹数量',
                               `MAX_FILE_SIZE` BIGINT DEFAULT 0 COMMENT '单文件最大大小（字节）',
                               `QUOTA_TYPE` VARCHAR(20) DEFAULT 'standard' COMMENT '配额类型: standard, premium',
                               `SYS_COMPANY_ID` BIGINT UNSIGNED DEFAULT NULL COMMENT '公司ID',
                               `CREATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '创建人',
                               `CREATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                               `UPDATE_BY` VARCHAR(80) DEFAULT NULL COMMENT '更新人',
                               `UPDATE_TIME` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                               `IS_ACTIVE` CHAR(1) DEFAULT 'Y' COMMENT '是否有效 Y/N',
                               PRIMARY KEY (`ID`),
                               UNIQUE KEY `uk_user_id` (`USER_ID`),
                               KEY `idx_sys_company_id` (`SYS_COMPANY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云盘配额表';

-- ==========================================
-- 初始化默认配额
-- ==========================================

-- 为现有用户创建默认配额（10GB存储空间，20GB单文件限制）
INSERT INTO `cloud_quota` (`USER_ID`, `TOTAL_QUOTA`, `USED_SPACE`, `FILE_COUNT`, `FOLDER_COUNT`, `MAX_FILE_SIZE`, `QUOTA_TYPE`, `CREATE_BY`)
SELECT
    u.ID,
    107374182400,   -- 10GB = 10 * 1024 * 1024 * 1024
    0,
    0,
    0,
    214748364800,   -- 20GB = 20 * 1024 * 1024 * 1024
    'standard',
    'system'
FROM sys_user u
WHERE NOT EXISTS (
    SELECT 1 FROM cloud_quota q WHERE q.USER_ID = u.ID
);

SET FOREIGN_KEY_CHECKS = 1;
