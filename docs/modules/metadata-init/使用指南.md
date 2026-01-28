# 元数据初始化工具使用指南

## 概述

`metadata-init` 是一个命令行工具，用于从现有数据库表结构自动生成元数据定义，将物理表映射到系统的元数据驱动架构中。

## 功能

### 1. 自动读取数据库结构 ✅

从 MySQL 数据库的 `information_schema` 中读取：
- 所有业务表（排除 `sys_` 开头的元数据表）
- 每个表的所有字段定义
- 表和字段的注释信息

### 2. 生成元数据记录 ✅

自动创建以下元数据：

**sys_table 记录**:
- `DB_NAME`: 表名（大写）
- `NAME`: 显示名称（优先使用表注释）
- `DESCRIPTION`: 表描述
- `MASK`: 默认权限 `AMDSQPGU`（增删改查提交打印授权反提交）
- `SYS_TABLE_CATEGORY_ID`: 默认类别 1
- 其他系统字段

**sys_column 记录**:
- 所有字段的完整定义
- 自动映射数据类型（MySQL → 系统类型）
- 自动推断字段属性：
  - 赋值类型 (SET_VALUE_TYPE)
  - 是否可修改 (MODIFI_ABLE)
  - 显示类型 (DISPLAY_TYPE)
  - 是否可空 (NULL_ABLE)

**sys_dict 和 sys_dict_item 记录**:
- 创建基础数据字典 `YESNO`
  - Y: 是
  - N: 否

**sys_directory 记录** ⭐ 新增:
- 为每个 sys_table 自动创建对应的安全目录
- 建立双向关联关系：
  - sys_directory.SYS_TABLE_ID → sys_table.ID
  - sys_table.SYS_DIRECTORY_ID → sys_directory.ID
- 用于权限系统的目录授权

### 3. 智能字段识别 ✅

自动识别特殊字段：

**主键字段 (PK)**:
- SET_VALUE_TYPE: `pk`
- MODIFI_ABLE: `N`
- IS_AK: `Y`
- IS_DK: `Y`

**标准审计字段**:
- `CREATE_BY` → createBy
- `UPDATE_BY` → operator
- `CREATE_TIME` / `UPDATE_TIME` → sysdate
- `SYS_COMPANY_ID` / `SYS_ORG_ID` → object

**外键字段** (以 `_ID` 结尾):
- SET_VALUE_TYPE: `fk`
- DISPLAY_TYPE: `select`

**是否字段**:
- `IS_ACTIVE` → SET_VALUE_TYPE: `select`, DISPLAY_TYPE: `check`

**普通字段**:
- SET_VALUE_TYPE: `byPage` (界面输入)
- MODIFI_ABLE: `Y`

## 使用方法

### 前提条件

1. 数据库已创建并导入表结构
2. 元数据表已创建 (sys_table, sys_column, sys_dict, sys_dict_item)
3. 配置文件 `configs/config.yaml` 已正确配置数据库连接

### 命令行参数 ⭐ 新增

```bash
metadata-init [参数]
```

**可用参数**:

| 参数 | 说明 |
|------|------|
| `--exclude-sys` | 排除 sys_ 开头的系统表（只初始化业务表） |
| `--only-sys` | 只初始化 sys_ 开头的系统表 |
| `--tables <names>` | 指定要初始化的表名（逗号分隔） |
| `--force` | 强制重新初始化已存在的表（会删除原有元数据） |
| `--init-db` | 在元数据初始化前先执行 sqls/init.sql 初始化数据库 |
| `--help` | 显示帮助信息 |

### 运行示例

**1. 初始化所有表（默认）** ⭐ 推荐
```bash
make metadata-init
# 或
go run cmd/metadata-init/main.go
```

**2. 只初始化业务表**
```bash
go run cmd/metadata-init/main.go --exclude-sys
```

**3. 只初始化系统表**
```bash
go run cmd/metadata-init/main.go --only-sys
```

**4. 初始化指定的表**
```bash
go run cmd/metadata-init/main.go --tables user,order,product
```

**5. 强制重新初始化所有表**
```bash
go run cmd/metadata-init/main.go --force
```

**6. 强制重新初始化指定表**
```bash
go run cmd/metadata-init/main.go --tables sys_user,sys_company --force
```

**7. 先执行 init.sql 初始化数据库，再初始化元数据** ⭐ 新增
```bash
go run cmd/metadata-init/main.go --init-db
```

**8. 组合使用参数**
```bash
# 执行 init.sql 后强制重新初始化所有表
go run cmd/metadata-init/main.go --init-db --force

# 执行 init.sql 后只初始化系统表
go run cmd/metadata-init/main.go --init-db --only-sys
```

**9. 编译后运行**
```bash
go build -o bin/metadata-init cmd/metadata-init/main.go
./bin/metadata-init --help
```

### 运行流程

```
1. 加载配置文件
   ↓
2. 连接数据库
   ↓
3. 执行 init.sql（如果指定 --init-db）⭐ 新增
   ↓
4. 初始化基础数据字典 (YESNO)
   ↓
5. 初始化 sys_directory（为已存在的 sys_table 创建安全目录）⭐ 新增
   ↓
6. 读取所有业务表
   ↓
7. 对每个表：
   a. 检查是否已存在于 sys_table
   b. 如已存在，跳过（--force 可强制重新初始化）
   c. 读取表的所有字段
   d. 在事务中创建 sys_table 和 sys_column 记录
   ↓
8. 再次初始化 sys_directory（为新增的 sys_table 创建安全目录）⭐ 新增
   ↓
9. 输出统计结果
```

## 输出示例

```
2026-01-12T10:00:00.000+0800    INFO    Starting metadata initialization
2026-01-12T10:00:00.100+0800    INFO    Database connected successfully
2026-01-12T10:00:00.150+0800    INFO    Database name   {"database": "skyserver"}
2026-01-12T10:00:00.200+0800    INFO    Base dictionaries initialized
2026-01-12T10:00:00.250+0800    INFO    Found tables in sys_table    {"count": 0}  ⭐ 新增
2026-01-12T10:00:00.300+0800    INFO    Directories initialized from sys_table  ⭐ 新增
2026-01-12T10:00:00.350+0800    INFO    Found tables    {"count": 5}

2026-01-12T10:00:00.300+0800    INFO    Processing table        {"table": "users"}
2026-01-12T10:00:00.350+0800    INFO    Created table metadata  {"table": "users", "table_id": 100, "columns": 12}
2026-01-12T10:00:00.400+0800    INFO    Table metadata initialized      {"table": "users"}

2026-01-12T10:00:00.450+0800    INFO    Processing table        {"table": "orders"}
2026-01-12T10:00:00.500+0800    INFO    Created table metadata  {"table": "orders", "table_id": 101, "columns": 15}
2026-01-12T10:00:00.550+0800    INFO    Table metadata initialized      {"table": "orders"}

2026-01-12T10:00:01.000+0800    INFO    Metadata initialization completed       {"success": 5, "failed": 0, "total": 5}
2026-01-12T10:00:01.050+0800    INFO    Created directory and linked to table   {"table": "USERS", "tableID": 100, "dirID": 1}  ⭐ 新增
2026-01-12T10:00:01.100+0800    INFO    Created directory and linked to table   {"table": "ORDERS", "tableID": 101, "dirID": 2}  ⭐ 新增
2026-01-12T10:00:01.500+0800    INFO    Directory initialization completed      {"created": 5, "updated": 0, "skipped": 0, "total": 5}  ⭐ 新增
2026-01-12T10:00:01.550+0800    INFO    Directories synchronized after metadata creation  ⭐ 新增
```

## 数据类型映射

### MySQL → 系统类型

| MySQL 类型 | 系统类型 | 说明 |
|-----------|---------|------|
| int, bigint, smallint, tinyint, mediumint | int | 整数 |
| decimal, numeric, float, double | decimal | 小数 |
| date | date | 日期 |
| datetime, timestamp | datetime | 日期时间 |
| time | time | 时间 |
| char | char | 定长字符 |
| text, mediumtext, longtext | text | 长文本 |
| varchar, 其他 | varchar | 变长字符 |

## SET_VALUE_TYPE 映射规则

| 条件 | SET_VALUE_TYPE | 说明 |
|------|---------------|------|
| 主键字段 (PRI) | pk | 主键自动生成 |
| CREATE_BY | createBy | 创建人 |
| UPDATE_BY | operator | 操作人 |
| CREATE_TIME, UPDATE_TIME | sysdate | 系统时间 |
| SYS_COMPANY_ID, SYS_ORG_ID | object | 上下文对象 |
| IS_ACTIVE | select | 下拉选择 |
| 以 _ID 结尾（非ID） | fk | 外键 |
| 其他 | byPage | 界面输入 |

## DISPLAY_TYPE 映射规则

| 条件 | DISPLAY_TYPE | 说明 |
|------|-------------|------|
| 主键字段 | text | 文本框 |
| IS_ACTIVE | check | 复选框 |
| date | date | 日期选择器 |
| datetime, timestamp | datetime | 日期时间选择器 |
| time | time | 时间选择器 |
| text, mediumtext, longtext | textarea | 文本域 |
| 以 _ID 结尾（非ID） | select | 下拉选择框 |
| 其他 | text | 文本框 |

## 示例

### 数据库表定义

```sql
CREATE TABLE `users` (
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `USERNAME` varchar(50) NOT NULL COMMENT '用户名',
  `EMAIL` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `PHONE` varchar(20) DEFAULT NULL COMMENT '手机号',
  `STATUS` char(1) DEFAULT 'Y' COMMENT '状态',
  `SYS_COMPANY_ID` bigint DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) DEFAULT NULL COMMENT '修改人',
  `UPDATE_TIME` datetime DEFAULT NULL COMMENT '修改时间',
  `IS_ACTIVE` char(1) DEFAULT 'Y' COMMENT '是否有效',
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB COMMENT='用户表';
```

### 生成的元数据

**sys_table 记录**:
```
ID: 100
DB_NAME: USERS
NAME: 用户表
DESCRIPTION: 用户表
MASK: AMDSQPGU
SYS_TABLE_CATEGORY_ID: 1
IS_ACTIVE: Y
CREATE_BY: system
```

**sys_column 记录** (部分):

```
1. ID 字段
   SYS_TABLE_ID: 100
   NAME: 主键
   DB_NAME: ID
   FULL_NAME: USERS.ID
   COL_TYPE: int
   ORDERNO: 10
   NULL_ABLE: N
   SET_VALUE_TYPE: pk
   MODIFI_ABLE: N
   DISPLAY_TYPE: text
   IS_AK: Y
   IS_DK: Y

2. USERNAME 字段
   SYS_TABLE_ID: 100
   NAME: 用户名
   DB_NAME: USERNAME
   FULL_NAME: USERS.USERNAME
   COL_TYPE: varchar
   COL_LENGTH: 50
   ORDERNO: 20
   NULL_ABLE: N
   SET_VALUE_TYPE: byPage
   MODIFI_ABLE: Y
   DISPLAY_TYPE: text

3. CREATE_BY 字段
   SYS_TABLE_ID: 100
   NAME: 创建人
   DB_NAME: CREATE_BY
   FULL_NAME: USERS.CREATE_BY
   COL_TYPE: varchar
   COL_LENGTH: 80
   ORDERNO: 70
   NULL_ABLE: Y
   SET_VALUE_TYPE: createBy
   MODIFI_ABLE: N
   DISPLAY_TYPE: text

4. IS_ACTIVE 字段
   SYS_TABLE_ID: 100
   NAME: 是否有效
   DB_NAME: IS_ACTIVE
   FULL_NAME: USERS.IS_ACTIVE
   COL_TYPE: char
   COL_LENGTH: 1
   ORDERNO: 110
   NULL_ABLE: Y
   SET_VALUE_TYPE: select
   MODIFI_ABLE: Y
   DISPLAY_TYPE: check
```

## 重要说明

### 1. 幂等性 ✅

工具会检查表是否已存在于 sys_table：
- 如果已存在，跳过该表
- 如果不存在，创建元数据
- 可以安全地多次运行

### 2. 事务保证 ✅

每个表的元数据创建在一个事务中：
- sys_table 记录创建成功
- 所有 sys_column 记录创建成功
- 任一失败则全部回滚

### 3. 表过滤 ✅

自动排除元数据表：
- 只处理业务表（非 `sys_` 开头）
- 避免循环引用
- 保护元数据表结构

### 4. 字段顺序 ✅

字段按 ORDERNO 排序：
- 第1个字段: ORDERNO = 10
- 第2个字段: ORDERNO = 20
- 第3个字段: ORDERNO = 30
- ...

## 后续优化

### 手动调整建议

初始化后，建议手动调整以下内容：

1. **表属性**:
   - 调整 MASK 权限
   - 设置正确的 SYS_TABLE_CATEGORY_ID
   - 设置 PARENT_TABLE_ID（主从表关系）
   - 创建或关联 SYS_DIRECTORY_ID

2. **字段属性**:
   - 调整 ORDERNO（字段显示顺序）
   - 设置外键的 REF_TABLE_ID 和 REF_COLUMN_ID
   - 为下拉字段设置 SYS_DICT_ID
   - 调整 MASK（字段级权限）
   - 设置验证规则 (VALIDATE_RULE)

3. **数据字典**:
   - 创建业务相关的字典
   - 为枚举字段创建字典项

### 配合插件使用

初始化后，如果使用 `SysTableAfterCreatePlugin`：
- 不建议重新创建已有表
- 新建表会自动生成标准字段
- 可能需要调整字段顺序避免冲突

## 故障排除

### 问题 1: 连接数据库失败

**错误**:
```
Failed to connect to database: dial tcp xxx: connect: connection refused
```

**解决方案**:
1. 检查 `configs/config.yaml` 中的数据库配置
2. 确保数据库服务已启动
3. 检查网络连接和防火墙

### 问题 2: 表已存在，但想重新初始化

**解决方案**:
```sql
-- 删除已有元数据（谨慎操作！）
DELETE FROM sys_column WHERE SYS_TABLE_ID IN (
  SELECT ID FROM sys_table WHERE DB_NAME = 'USERS'
);
DELETE FROM sys_table WHERE DB_NAME = 'USERS';
```

然后重新运行初始化工具。

### 问题 3: 某些表初始化失败

**日志示例**:
```
Failed to initialize table metadata     {"table": "orders", "error": "..."}
```

**解决方案**:
1. 查看详细错误日志
2. 检查表结构是否符合要求
3. 手动处理失败的表

## 最佳实践

### 1. 数据库设计建议

为了最大化自动化效果：
- 为表和字段添加 COMMENT
- 使用标准审计字段（CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME, IS_ACTIVE）
- 外键字段统一以 `_ID` 结尾
- 主键字段命名为 `ID`

### 2. 初始化流程建议

```bash
# 1. 创建数据库表结构
mysql -u root -p < sqls/create_business_tables.sql

# 2. 运行元数据初始化
make metadata-init

# 3. 手动调整元数据
# - 设置正确的 SYS_TABLE_CATEGORY_ID
# - 配置外键引用
# - 创建业务字典

# 4. 验证元数据
# - 通过 API 测试 CRUD 操作
# - 检查字段显示和权限
```

### 3. 团队协作建议

- 将表结构 SQL 文件纳入版本控制
- 记录手动调整的元数据
- 为特殊配置编写迁移脚本

## 总结

`metadata-init` 工具提供了快速从现有数据库生成元数据的能力：

✅ **优点**:
- 自动化程度高
- 智能字段识别
- 事务保证数据一致性
- 幂等性，可重复运行
- 节省大量手工配置时间

⚠️ **注意**:
- 生成的元数据是基础配置
- 仍需手动调整业务特性
- 不会覆盖已存在的表

🎯 **适用场景**:
- 新项目快速启动
- 现有系统迁移到元数据驱动架构
- 批量导入业务表
- 开发测试环境快速搭建

这个工具大大简化了元数据系统的初始化工作，让您可以快速开始业务开发！✅
