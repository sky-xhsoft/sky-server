# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Sky-Server 是一个用 Go 语言编写的元数据驱动应用框架。系统通过数据库元数据定义来动态生成表单、字段、权限和UI配置。这是一个早期阶段的项目，目前主要包含数据库表结构定义。

## 数据库架构

数据库架构文件位于 `sqls/create_skyserver.sql`，定义了一个元数据驱动的系统，包含以下核心概念：

### 核心元数据表

- **sys_table**: 定义系统中的逻辑表/表单。每个表包含显示名称、过滤规则、访问掩码（A:新增, M:修改, D:删除, Q:查询, S:提交, U:反提交, V:作废）以及UI配置引用
- **sys_column**: 定义表的字段/列。包括字段类型、校验规则、显示类型和赋值策略（pk, docno, createBy, byPage, select, fk, sysdate, operator, ignore）
- **sys_table_ref**: 定义表之间的关联关系（1:1 或 1:n 关联）

### 权限控制

- **sys_company**: 多租户公司隔离
- **sys_user**: 用户账户，包含 SGRADE 字段访问级别
- **sys_groups**: 权限组，包含 SGRADE 访问级别
- **sys_group_prem**: 将权限组映射到目录，定义权限（读=1, 读写=3, 读提交=5 ……）
  - **sys_directory**: 安全目录，映射到表

### 动态行为

- **sys_action**: 定义可附加到表的自定义动作（类型：url, sp(存储过程), job, js(JavaScript), bsh(Bash), py(Python)）
- **sys_table_cmd**: 表级命令钩子，在标准动作前后执行（A, M, D, Q, S, U, V, I, E）
- **sys_seq**: 单据编号生成器，支持循环类型和格式化

### UI配置

- **sys_objuiconf**: 对象显示配置（CSS类、列数、默认动作）
- **sys_dict** / **sys_dict_item**: 数据字典，用于下拉选项
- **sys_subsystem**: 子系统/模块定义，包含URL和图标

### 关键元数据模式

1. **字段赋值策略 (SET_VALUE_TYPE)**:
   - `pk`: 主键（自动生成）
   - `docno`: 单据编号（使用 sys_seq）
   - `createBy`: 创建人
   - `byPage`: 界面输入
   - `select`: 下拉选择（使用 sys_dict）
   - `fk`: 外键关联
   - `sysdate`: 操作时间
   - `operator`: 操作用户
   - `ignore`: 忽略

2. **标准字段**: 所有表都包含：
   - `ID`: 主键
   - `SYS_COMPANY_ID`: 公司隔离
   - `CREATE_BY`, `CREATE_TIME`: 创建跟踪
   - `UPDATE_BY`, `UPDATE_TIME`: 修改跟踪
   - `IS_ACTIVE`: 软删除标志（Y/N）

3. **动作显示类型**:
   - `list_button`: 列表栏按钮
   - `list_menu_item`: 列表栏菜单
   - `obj_button`: 单对象界面按钮
   - `obj_menu_item`: 单对象界面菜单
   - `tab_button`: 单对象标签页按钮

## 开发命令

由于这是一个早期阶段的项目，尚未包含 Go 代码，将使用标准 Go 命令：

```bash
# 初始化 Go 模块（创建项目时）
go mod init github.com/sky-xhsoft/sky-server

# 构建项目
go build -o sky-server ./cmd/server

# 运行测试
go test ./...

# 运行单个测试包
go test ./pkg/metadata

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...

# 数据库初始化
mysql -u root -p < sqls/create_skyserver.sql
```

## 架构说明

该系统采用**元数据驱动架构**：

1. 数据库表不在应用程序中硬编码
2. 表定义、列和关系存储在 `sys_table`、`sys_column` 和 `sys_table_ref` 中
3. UI表单和字段控件根据元数据动态生成
4. 权限和数据过滤通过元数据配置应用
5. 自定义动作和业务逻辑可以通过 `sys_action` 和 `sys_table_cmd` 注入

实现 Go 应用时的关键点：
- 创建元数据加载器从数据库读取表/列定义
- 构建基于元数据的通用 CRUD 引擎，可处理任何表
- 实现表达式求值器用于过滤器和校验规则
- 支持架构中定义的各种字段赋值策略和动作类型
