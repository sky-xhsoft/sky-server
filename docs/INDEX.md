# Sky-Server 文档索引

> **版本**: v2.0
> **最后更新**: 2026-01-28
> **维护者**: Sky Team

---

## 📚 文档导航

### 🚀 快速开始

- [系统架构设计](./01-系统架构设计.md) - 整体架构、技术栈、模块划分
- [数据库设计](./02-数据库设计.md) - 数据库表结构、关系设计
- [API 设计](./03-API设计.md) - RESTful API 接口规范
- [开发计划](./04-开发计划.md) - 项目开发路线图

---

## 📦 核心模块

### ☁️ 云盘模块 (`modules/cloud/`)

**功能概述**: 完整的云盘文件管理系统，支持文件上传下载、分享、配额管理

- [上传功能指南](./modules/cloud/上传功能指南.md) - 基础上传和大文件上传（20GB）
- [分片上传](./modules/cloud/分片上传.md) - 分片上传和断点续传
- [断点续传分析](./modules/cloud/断点续传分析.md) - 断点续传技术分析
- [API 参考](./modules/cloud/API参考.md) - 完整的云盘 API 文档

**相关文档**:
- [云盘服务实现](./cloud-service-implementation.md)
- [文件上传故障排查](./file-upload-troubleshooting.md)
- [云盘日期时间修复](./cloud-datetime-fix.md)

### 🔌 插件系统 (`modules/plugin/`)

**功能概述**: 灵活的插件架构，支持动态加载、热重载和生命周期管理

- [插件系统概述](./modules/plugin/插件系统概述.md) - 架构设计、核心概念、开发指南
- [热加载指南](./modules/plugin/热加载指南.md) - 热重载功能详细说明
- [热加载快速开始](./modules/plugin/热加载快速开始.md) - 快速接入指南
- [测试指南](./modules/plugin/测试指南.md) - 插件测试编写指南

**相关文档**:
- [插件命名规范](./plugin-naming-convention.md)
- [插件实现总结](./plugin-implementation-summary.md)
- [插件重构总结](./plugin-refactoring-summary.md)
- [Hooks 自动注册](./hooks-auto-registration.md)

### 📊 元数据初始化 (`modules/metadata-init/`)

**功能概述**: 元数据初始化工具，用于数据库初始化和目录管理

- [使用指南](./modules/metadata-init/使用指南.md) - 工具使用说明
- [工具总结](./metadata-init-tool-summary.md) - 实现总结
- [数据库初始化功能](./metadata-init-init-db-feature.md)
- [目录管理功能](./metadata-init-directory-feature.md)

**相关文档**:
- [元数据初始化 README](./METADATA-INIT-README.md)
- [升级总结](./metadata-init-upgrade-summary.md)

### 🔄 事务管理 (`modules/transaction/`)

**功能概述**: 数据库事务处理和管理

- [使用指南](./modules/transaction/使用指南.md) - 事务使用说明
- [事务分析](./transaction-analysis.md) - 事务处理分析
- [实现总结](./transaction-implementation-summary.md)
- [Hooks 修复](./transaction-hooks-fix.md)

---

## 📖 功能文档

### 核心功能

- [文件上传 API](./file-upload-api.md) - 完整的文件上传功能文档
- [菜单 API](./menu-api.md) - 菜单管理接口文档
- [权限系统](./admin-permission-feature.md) - 管理员权限功能说明
- [多租户系统](./domain-based-multi-tenancy.md) - 基于域名的多租户实现

### 数据处理

- [更新行为说明](./update-behavior.md) - 数据更新行为说明
- [更新零值修复](./update-zero-value-fix.md) - 零值更新问题修复

---

## 📝 开发总结

### 阶段总结

- [Phase 1-9 完成总结](./Phase1-完成总结.md) - 各阶段开发总结
- [Phase 10 完善权限组体系方案](./Phase10-完善权限组体系方案.md)
- [Phase 11 权限集成完成总结](./Phase11-权限集成完成总结.md)
- [Phase 12 完成总结](./Phase12-完成总结.md)
- [Phase 13 云盘功能设计总结](./Phase13-云盘功能设计总结.md)
- [Phase 14 消息通知系统实现总结](./Phase14-消息通知系统实现总结.md)
- [Phase 15 WebSocket 实时推送实现总结](./Phase15-WebSocket实时推送实现总结.md)

### 功能实现

- [权限系统重复分析](./权限系统重复分析.md)
- [菜单实现总结](./menu-implementation-summary.md)

---

## 🔧 运维指南

- [脚本使用指南](./scripts/scripts-guide.md) - 运维脚本使用说明
- [SQL 指南](./guides/sql-guide.md) - SQL 脚本使用说明
- [测试结果](./guides/test-results.md) - 测试结果报告

---

## 📋 API 文档

- [API 废弃指南](./API_DEPRECATION_GUIDE.md) - API 废弃和迁移指南
- [最终验证报告](./FINAL_VERIFICATION_REPORT.md) - 系统验证报告
- [迁移验证](./MIGRATION_VERIFICATION.md) - 迁移验证报告

---

## 🗂️ 文档组织结构

```
docs/
├── INDEX.md                    # 本文档
├── modules/                    # 模块文档
│   ├── cloud/                  # 云盘模块
│   │   ├── 上传功能指南.md
│   │   ├── 分片上传.md
│   │   ├── 断点续传分析.md
│   │   └── API参考.md
│   ├── plugin/                 # 插件系统
│   │   ├── 插件系统概述.md
│   │   ├── 热加载指南.md
│   │   ├── 热加载快速开始.md
│   │   └── 测试指南.md
│   ├── metadata-init/          # 元数据初始化
│   │   └── 使用指南.md
│   └── transaction/            # 事务管理
│       └── 使用指南.md
├── guides/                     # 通用指南
├── scripts/                    # 脚本文档
└── plugins/                    # 插件文档

```

---

## 🔍 快速查找

### 我想了解...

- **如何开始开发？** → [系统架构设计](./01-系统架构设计.md) + [开发计划](./04-开发计划.md)
- **如何使用文件上传？** → [上传功能指南](./modules/cloud/上传功能指南.md)
- **如何实现分片上传？** → [分片上传](./modules/cloud/分片上传.md)
- **如何开发插件？** → [插件系统概述](./modules/plugin/插件系统概述.md)
- **如何使用元数据工具？** → [元数据初始化指南](./modules/metadata-init/使用指南.md)
- **如何实现云盘功能？** → [云盘 API 参考](./modules/cloud/API参考.md)
- **如何处理事务？** → [事务管理指南](./modules/transaction/使用指南.md)
- **如何配置权限？** → [权限系统](./admin-permission-feature.md)
- **如何实现多租户？** → [多租户系统](./domain-based-multi-tenancy.md)

---

## 📌 重要提示

1. **文档版本**: 所有文档都标注了版本号和更新日期，请注意查看
2. **模块化组织**: 文档已按模块重新组织，便于查找和维护
3. **代码示例**: 文档中的代码示例均经过测试，可直接使用
4. **问题反馈**: 如发现文档问题，请及时反馈给维护团队
5. **持续更新**: 文档会随着项目发展持续更新

---

## 📞 联系方式

- **项目地址**: https://github.com/sky-xhsoft/sky-server
- **维护团队**: Sky Team
- **更新日期**: 2026-01-28

---

**文档索引版本**: v2.0 (模块化重组)
**最后更新**: 2026-01-28
**维护者**: Sky Team
