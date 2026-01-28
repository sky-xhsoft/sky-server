# Sky-Server 文档索引

> **版本**: v1.0
> **最后更新**: 2026-01-28
> **维护者**: Sky Team

---

## 📚 文档导航

### 🏗️ 系统架构与设计

- [系统架构设计](./01-系统架构设计.md) - 整体架构、技术栈、模块划分
- [数据库设计](./02-数据库设计.md) - 数据库表结构、关系设计
- [API 设计](./03-API设计.md) - RESTful API 接口规范
- [开发计划](./04-开发计划.md) - 项目开发路线图

### 📖 开发指南

#### 核心功能

- [文件上传 API](./file-upload-api.md) - 完整的文件上传功能文档
- [菜单 API](./menu-api.md) - 菜单管理接口文档
- [权限系统](./admin-permission-feature.md) - 管理员权限功能说明
- [多租户系统](./domain-based-multi-tenancy.md) - 基于域名的多租户实现

#### 元数据系统

- [元数据初始化工具指南](./metadata-init-guide.md) - 元数据初始化工具使用说明
- [元数据初始化工具 README](./METADATA-INIT-README.md) - 工具概述和快速开始
- [元数据初始化工具总结](./metadata-init-tool-summary.md) - 工具实现总结
- [元数据初始化升级总结](./metadata-init-upgrade-summary.md) - 工具升级记录
- [元数据初始化数据库功能](./metadata-init-init-db-feature.md) - 数据库初始化功能
- [元数据初始化目录功能](./metadata-init-directory-feature.md) - 目录管理功能

#### 云盘功能

- [云盘 API](./cloud-api.md) - 云盘接口文档
- [云盘 API 快速开始](./cloud-api-quickstart.md) - 快速接入指南
- [云盘服务实现](./cloud-service-implementation.md) - 服务层实现细节
- [云盘服务迁移总结](./cloud_service_migration_summary.md) - 服务迁移记录
- [云盘日期时间修复](./cloud-datetime-fix.md) - 日期时间字段修复
- [文件夹 ID 零值修复](./folder-id-zero-fix.md) - 文件夹 ID 问题修复
- [大文件上传](./large-file-upload.md) - 大文件上传实现
- [文件上传故障排查](./file-upload-troubleshooting.md) - 上传问题排查指南
- [云盘断点续传分析](./cloud-resumable-upload-analysis.md) - 断点续传技术分析
- [分片上传快速开始](./multipart-upload-quickstart.md) - 分片上传快速指南
- [分片上传实现总结](./multipart-upload-implementation-summary.md) - 分片上传实现记录

#### 插件系统

- [插件系统概述](./plugins/plugins-overview.md) - 插件系统架构和使用
- [插件系统文档](./plugin-system.md) - 插件系统详细文档
- [插件命名规范](./plugin-naming-convention.md) - 插件命名约定
- [插件实现总结](./plugin-implementation-summary.md) - 插件系统实现记录
- [插件重构总结](./plugin-refactoring-summary.md) - 插件重构记录
- [插件迁移总结](./plugin-migration-summary.md) - 插件迁移记录
- [插件测试指南](./plugin-tests-guide.md) - 插件测试编写指南
- [插件测试总结](./plugin-tests-summary.md) - 插件测试实现记录
- [插件热加载快速开始](./plugin-hotload-quickstart.md) - 热加载快速指南
- [插件热加载指南](./plugin-hotload-guide.md) - 热加载详细文档
- [插件热加载文档](./plugins/hotload.md) - 热加载插件说明
- [Hooks 插件文档](./plugins/hooks.md) - Hooks 插件说明
- [Hooks 自动注册](./hooks-auto-registration.md) - Hooks 自动注册机制
- [Go Hook 指南](./go-hook-guide.md) - Go 语言 Hook 编写指南
- [sys_table_after_create 插件升级](./plugin-sys-table-after-create-upgrade.md) - 插件升级记录

#### 事务管理

- [事务分析](./transaction-analysis.md) - 事务处理分析
- [事务指南](./transaction-guide.md) - 事务使用指南
- [事务实现总结](./transaction-implementation-summary.md) - 事务实现记录
- [事务 Hooks 修复](./transaction-hooks-fix.md) - 事务 Hooks 问题修复
- [事务修复总结](./transaction-fix-summary.md) - 事务修复记录

#### 数据更新

- [更新行为说明](./update-behavior.md) - 数据更新行为说明
- [更新零值修复](./update-zero-value-fix.md) - 零值更新问题修复

### 📝 开发总结

#### 阶段总结

- [Phase 1 完成总结](./Phase1-完成总结.md)
- [Phase 2 完成总结](./Phase2-完成总结.md)
- [Phase 3 完成总结](./Phase3-完成总结.md)
- [Phase 4 完成总结](./Phase4-完成总结.md)
- [Phase 5 完成总结](./Phase5-完成总结.md)
- [Phase 6 完成总结](./Phase6-完成总结.md)
- [Phase 7 完成总结](./Phase7-完成总结.md)
- [Phase 8 完成总结](./Phase8-完成总结.md)
- [Phase 9 完成总结](./Phase9-完成总结.md)
- [Phase 10 完善权限组体系方案](./Phase10-完善权限组体系方案.md)
- [Phase 10 完成总结](./Phase10-完成总结.md)
- [Phase 11 权限集成完成总结](./Phase11-权限集成完成总结.md)
- [Phase 12 完成总结](./Phase12-完成总结.md)
- [Phase 13 云盘功能设计总结](./Phase13-云盘功能设计总结.md)
- [Phase 14 消息通知系统实现总结](./Phase14-消息通知系统实现总结.md)
- [Phase 15 WebSocket 实时推送实现总结](./Phase15-WebSocket实时推送实现总结.md)

#### 功能实现

- [权限系统重复分析](./权限系统重复分析.md)
- [菜单实现总结](./menu-implementation-summary.md)

### 🔧 运维指南

- [脚本使用指南](./scripts/scripts-guide.md) - 运维脚本使用说明
- [SQL 指南](./guides/sql-guide.md) - SQL 脚本使用说明
- [测试结果](./guides/test-results.md) - 测试结果报告

### 📋 API 文档

- [API 废弃指南](./API_DEPRECATION_GUIDE.md) - API 废弃和迁移指南
- [最终验证报告](./FINAL_VERIFICATION_REPORT.md) - 系统验证报告
- [迁移验证](./MIGRATION_VERIFICATION.md) - 迁移验证报告

---

## 🗂️ 文档分类

### 按功能分类

#### 核心系统
- 系统架构、数据库设计、API 设计

#### 文件管理
- 文件上传、云盘功能、大文件上传、分片上传

#### 权限管理
- 权限系统、多租户、菜单管理

#### 元数据系统
- 元数据初始化工具、元数据管理

#### 插件系统
- 插件开发、热加载、Hooks

#### 数据处理
- 事务管理、数据更新

### 按文档类型分类

#### 设计文档
- 系统架构设计、数据库设计、API 设计

#### 开发指南
- 各功能模块的开发指南和使用说明

#### 实现总结
- 各阶段的开发总结和实现记录

#### 故障排查
- 问题修复记录和故障排查指南

---

## 🔍 快速查找

### 我想了解...

- **如何开始开发？** → [系统架构设计](./01-系统架构设计.md) + [开发计划](./04-开发计划.md)
- **如何使用文件上传？** → [文件上传 API](./file-upload-api.md)
- **如何开发插件？** → [插件系统概述](./plugins/plugins-overview.md) + [插件系统文档](./plugin-system.md)
- **如何使用元数据工具？** → [元数据初始化工具指南](./metadata-init-guide.md)
- **如何实现云盘功能？** → [云盘 API 快速开始](./cloud-api-quickstart.md)
- **如何处理事务？** → [事务指南](./transaction-guide.md)
- **如何配置权限？** → [权限系统](./admin-permission-feature.md)
- **如何实现多租户？** → [多租户系统](./domain-based-multi-tenancy.md)

---

## 📌 重要提示

1. **文档版本**: 所有文档都标注了版本号和更新日期，请注意查看
2. **代码示例**: 文档中的代码示例均经过测试，可直接使用
3. **问题反馈**: 如发现文档问题，请及时反馈给维护团队
4. **持续更新**: 文档会随着项目发展持续更新

---

## 📞 联系方式

- **项目地址**: https://github.com/sky-xhsoft/sky-server
- **维护团队**: Sky Team
- **更新日期**: 2026-01-28

---

**文档索引版本**: v1.0
**最后更新**: 2026-01-28
**维护者**: Sky Team
