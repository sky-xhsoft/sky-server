# Sky-Server

元数据驱动的企业级应用框架

## 项目简介

Sky-Server 是一个基于元数据驱动的企业级应用框架，通过元数据配置快速构建企业管理系统，减少重复的CRUD代码开发工作。

### 核心特性

- **元数据驱动**: 表结构、字段、UI、权限等均通过元数据定义
- **动态表单生成**: 基于元数据自动生成数据录入表单和列表视图
- **灵活的权限控制**: 支持多租户、角色组、字段级权限控制
- **可扩展的动作系统**: 支持URL、存储过程、JavaScript、Go方法、Bash等多种动作类型
- **单据编号管理**: 内置单据编号生成器，支持多种循环规则
- **数据字典**: 统一管理下拉选项等枚举数据
- **审核流程**: 支持单据提交、审核、作废等完整生命周期管理
- **审计日志**: 记录所有关键操作，支持数据追溯
- **单点登录**: 支持多客户端（Web、移动端、桌面端）统一认证

## 技术栈

- **开发语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0+
- **缓存**: Redis
- **配置管理**: Viper
- **日志**: Zap
- **API文档**: Swagger

## 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+
- Redis

### 安装

1. 克隆项目

```bash
git clone https://github.com/sky-xhsoft/sky-server.git
cd sky-server
```

2. 安装依赖

```bash
make init
```

3. 配置数据库

```bash
# 复制配置文件
cp configs/config.example.yaml configs/config.yaml

# 编辑配置文件，修改数据库连接信息
vi configs/config.yaml
```

4. 初始化数据库

```bash
# 创建元数据表
mysql -u root -p < sqls/create_skyserver.sql

# 如果有业务表，可以从数据库自动生成元数据
make metadata-init
```

5. 运行项目

```bash
make run
```

6. 访问API文档

打开浏览器访问: http://localhost:9090/swagger/index.html

## 项目结构

```
sky-server/
├── cmd/
│   ├── server/              # 主程序入口
│   ├── hash/                # 密码哈希工具
│   └── metadata-init/       # 元数据初始化工具
├── internal/                # 内部包
│   ├── api/                 # API层
│   ├── service/             # 业务逻辑层
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   └── pkg/                 # 内部工具包
├── configs/                 # 配置文件
├── sqls/                    # SQL文件
└── docs/                    # 文档
```

## 开发命令

```bash
make help          # 显示帮助信息
make init          # 初始化项目（安装依赖）
make build         # 编译项目
make run           # 运行项目
make test          # 运行测试
make clean         # 清理编译产物
make tidy          # 整理依赖
make swagger       # 生成Swagger文档
make metadata-init # 从数据库初始化元数据
```

## 文档

详细文档请查看 [docs](./docs) 目录：

**设计文档**:
- [系统架构设计](./docs/01-系统架构设计.md)
- [数据库设计](./docs/02-数据库设计.md)
- [API设计](./docs/03-API设计.md)
- [开发计划](./docs/04-开发计划.md)

**使用指南**:
- [元数据初始化工具使用指南](./docs/metadata-init-guide.md)
- [插件测试指南](./docs/plugin-tests-guide.md)

**技术文档**:
- [事务修复总结](./docs/transaction-fix-summary.md)
- [Update 零值修复](./docs/update-zero-value-fix.md)
- [插件升级说明](./docs/plugin-sys-table-after-create-upgrade.md)

## 许可证

Apache 2.0
