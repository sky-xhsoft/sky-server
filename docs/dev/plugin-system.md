# Sky-Server 插件系统使用指南

**实现日期**: 2026-01-12
**版本**: 2.0

## 概述

Sky-Server 插件系统提供了一个灵活、可扩展的插件架构，支持动态加载插件，在表单操作的不同阶段注入自定义逻辑。

### 核心特性

- ✅ **自动注册**: 使用 init() 函数自动注册插件
- ✅ **动态加载**: 支持运行时加载和卸载插件
- ✅ **优先级控制**: 控制插件执行顺序
- ✅ **钩子点系统**: 灵活的时机和操作组合
- ✅ **事务支持**: 插件执行在事务中，保证数据一致性
- ✅ **易于扩展**: 简单的接口，快速开发新插件

## 目录结构

```
plugins/
├── core/                      # 核心接口和管理器
│   ├── plugin.go             # 插件接口定义
│   └── manager.go            # 插件管理器
├── registry/                  # 插件注册中心
│   ├── registry.go           # 全局注册中心
│   └── loader.go             # 插件加载器
├── builtin/                   # 内置插件
│   ├── sys_table_after_create.go
│   ├── sys_table_before_delete.go
│   └── utils.go
└── setup.go                   # 插件系统初始化
```

## 核心概念

### 1. 插件接口

所有插件必须实现 `core.Plugin` 接口：

```go
type Plugin interface {
    Name() string                                          // 插件唯一标识
    Description() string                                   // 插件描述
    Version() string                                       // 版本号
    Execute(ctx context.Context, db *gorm.DB, data PluginData) error
}
```

### 2. 钩子点 (Hook Points)

钩子点定义插件的执行时机，格式：`{tableName}.{timing}.{action}`

**示例**:
- `sys_table.after.create` - sys_table 表创建后
- `sys_user.before.update` - sys_user 表更新前
- `order.after.delete` - order 表删除后

**时机**: `before` / `after`
**操作**: `create` / `update` / `delete` / `query` / `submit` / `unsubmit`

### 3. 插件数据

```go
type PluginData struct {
    TableName string                 // 表名
    Action    string                 // 操作类型
    Timing    string                 // 执行时机
    RecordID  uint                   // 记录ID
    Data      map[string]interface{} // 数据内容
    UserID    uint                   // 操作用户ID
    CompanyID uint                   // 公司ID
    Extra     map[string]interface{} // 额外上下文
}
```

## 快速开始

### 步骤 1: 创建插件

在 `plugins/` 目录下创建新的插件包：

```go
package myplugin

import (
    "context"
    "github.com/sky-xhsoft/sky-server/plugins/core"
    "github.com/sky-xhsoft/sky-server/plugins/registry"
    "gorm.io/gorm"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string        { return "my_plugin" }
func (p *MyPlugin) Description() string { return "我的自定义插件" }
func (p *MyPlugin) Version() string     { return "1.0.0" }

func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // 插件逻辑
    return nil
}
```

### 步骤 2: 注册插件

使用 `init()` 函数自动注册：

```go
func init() {
    registry.Register(
        "my_plugin",
        func() core.Plugin { return &MyPlugin{} },
        core.PluginMetadata{
            Name:        "my_plugin",
            Description: "我的自定义插件",
            Version:     "1.0.0",
            Author:      "Your Name",
            Enabled:     true,
            Priority:    10,
            HookPoint:   "sys_user.after.create",
        },
    )
}
```

### 步骤 3: 导入插件

在 `plugins/setup.go` 中添加导入：

```go
import (
    _ "github.com/sky-xhsoft/sky-server/plugins/builtin"
    _ "github.com/sky-xhsoft/sky-server/plugins/myplugin" // 导入你的插件
)
```

完成！插件会在系统启动时自动加载。

## 实用示例

### 示例 1: 数据验证插件

```go
package validation

import (
    "context"
    "fmt"
    "github.com/sky-xhsoft/sky-server/plugins/core"
    "github.com/sky-xhsoft/sky-server/plugins/registry"
    "gorm.io/gorm"
)

type OrderValidationPlugin struct{}

func init() {
    registry.Register("order_validation",
        func() core.Plugin { return &OrderValidationPlugin{} },
        core.PluginMetadata{
            HookPoint: "order.before.create",
            Priority:  5, // 高优先级
            Enabled:   true,
        })
}

func (p *OrderValidationPlugin) Name() string        { return "order_validation" }
func (p *OrderValidationPlugin) Description() string { return "订单验证" }
func (p *OrderValidationPlugin) Version() string     { return "1.0.0" }

func (p *OrderValidationPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // 验证订单金额
    amount, ok := data.Data["AMOUNT"].(float64)
    if !ok || amount <= 0 {
        return fmt.Errorf("订单金额必须大于0")
    }

    // 检查客户
    customerID := data.Data["CUSTOMER_ID"]
    var count int64
    db.Table("customer").Where("ID = ?", customerID).Count(&count)
    if count == 0 {
        return fmt.Errorf("客户不存在")
    }

    return nil
}
```

### 示例 2: 通知插件

```go
package notification

type NotificationPlugin struct{}

func init() {
    registry.Register("order_notification",
        func() core.Plugin { return &NotificationPlugin{} },
        core.PluginMetadata{
            HookPoint: "order.after.create",
            Priority:  20, // 低优先级
            Enabled:   true,
        })
}

func (p *NotificationPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // 发送通知
    orderNo := data.Data["ORDER_NO"]
    fmt.Printf("订单 %v 已创建\n", orderNo)

    // 调用消息服务（示例）
    // messageService.SendNotification(ctx, ...)

    return nil
}
```

### 示例 3: 审计日志插件

```go
package audit

type AuditLogPlugin struct{}

func init() {
    registry.Register("audit_log",
        func() core.Plugin { return &AuditLogPlugin{} },
        core.PluginMetadata{
            HookPoint: "*.after.*", // 所有表的所有after操作
            Priority:  100,
            Enabled:   true,
        })
}

func (p *AuditLogPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // 记录审计日志
    return db.Table("sys_audit_log").Create(map[string]interface{}{
        "TABLE_NAME": data.TableName,
        "ACTION":     data.Action,
        "RECORD_ID":  data.RecordID,
        "USER_ID":    data.UserID,
        "IS_ACTIVE":  "Y",
    }).Error
}
```

## 内置插件

### sys_table_after_create

**钩子点**: `sys_table.after.create`
**功能**: sys_table 表创建后自动生成标准字段

**执行内容**:
1. 验证 MASK 字段格式
2. 自动生成 orderno
3. 创建 directory（对于主表）
4. 创建标准字段：ID, SYS_COMPANY_ID, CREATE_BY, UPDATE_BY, CREATE_TIME, UPDATE_TIME, IS_ACTIVE
5. 设置表的 pk_column_id

### sys_table_before_delete

**钩子点**: `sys_table.before.delete`
**功能**: sys_table 删除前级联删除字段和目录

**执行内容**:
1. 删除 sys_column 中的所有字段配置
2. 删除 sys_directory 中的关联目录

## 最佳实践

### 1. 插件命名

格式：`{table}_{timing}_{action}`

```go
// 好的命名
"sys_table_after_create"
"order_before_update"
"user_after_delete"

// 不好的命名
"plugin1"
"my_plugin"
"test"
```

### 2. 优先级设置

```
1-10:   验证类插件（先验证）
10-50:  业务逻辑插件（核心功能）
50-100: 通知和审计插件（后处理）
```

### 3. 错误处理

始终返回明确的错误：

```go
func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    if err := validate(data); err != nil {
        return fmt.Errorf("验证失败: %w", err)
    }
    if err := process(db, data); err != nil {
        return fmt.Errorf("处理失败: %w", err)
    }
    return nil
}
```

### 4. 事务安全

插件的 `db` 参数可能是事务连接，返回错误会触发回滚：

```go
func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // db 在事务中
    if err := db.Table("related").Create(&record).Error; err != nil {
        return err // 触发事务回滚
    }
    return nil
}
```

### 5. 性能考虑

避免耗时操作，使用异步处理：

```go
func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // 快速执行
    quickOperation(data)

    // 耗时操作异步执行
    go func() {
        longRunningOperation(data)
    }()

    return nil
}
```

## 插件管理

### 查看已注册的插件

```go
import "github.com/sky-xhsoft/sky-server/plugins/registry"

// 列出所有插件
pluginNames := registry.ListPlugins()
for _, name := range pluginNames {
    fmt.Println(name)
}

// 获取插件详情
factories := registry.GetAllFactories()
for name, info := range factories {
    fmt.Printf("%s (v%s): %s\n",
        name, info.Metadata.Version, info.Metadata.Description)
}
```

### 动态启用/禁用

```go
// 禁用插件
manager.DisablePlugin("order.after.create", "notification_plugin")

// 启用插件
manager.EnablePlugin("order.after.create", "notification_plugin")
```

### 查看钩子点

```go
// 列出所有钩子点
hookPoints := manager.ListHookPoints()

// 查看钩子点的插件
plugins := manager.GetPlugins("order.after.create")
for _, info := range plugins {
    fmt.Printf("%s [优先级: %d, 启用: %v]\n",
        info.Plugin.Name(), info.Metadata.Priority, info.Metadata.Enabled)
}
```

## 常见问题

### Q: 同一钩子点可以注册多个插件吗？

A: 可以。系统会按优先级顺序执行。

### Q: 插件执行失败会怎样？

A: 整个操作会回滚，错误返回给调用方，后续插件不执行。

### Q: 插件可以修改数据吗？

A: 可以。在 `before` 钩子中修改 `data.Data` 会影响保存到数据库的数据。

### Q: 如何调试插件？

A: 使用 logger 输出日志：

```go
import "github.com/sky-xhsoft/sky-server/internal/pkg/logger"
import "go.uber.org/zap"

func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    logger.Info("执行插件",
        zap.String("plugin", p.Name()),
        zap.Any("data", data))
    // ...
    return nil
}
```

## 进阶功能

### 通配符钩子点

```go
core.PluginMetadata{
    HookPoint: "*.after.create",  // 所有表的 after.create
    // 或
    HookPoint: "order.after.*",   // order 表的所有 after 操作
}
```

### 插件间通信

通过 `PluginData.Extra` 传递数据：

```go
// 插件1（优先级低）
func (p *Plugin1) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    result := doSomething()
    if data.Extra == nil {
        data.Extra = make(map[string]interface{})
    }
    data.Extra["result"] = result
    return nil
}

// 插件2（优先级高）
func (p *Plugin2) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    if result, ok := data.Extra["result"]; ok {
        useResult(result)
    }
    return nil
}
```

## 迁移指南

### 从旧版本迁移

如果你的代码引用了 `internal/plugin/cmd`，需要更新导入：

**旧代码**:
```go
import plugin "github.com/sky-xhsoft/sky-server/internal/plugin/cmd"

pluginManager := plugin.Setup(db)
```

**新代码**:
```go
import "github.com/sky-xhsoft/sky-server/plugins"

pluginManager := plugins.Setup(db)
```

**类型更新**:
```go
// 旧
*plugin.Manager → *core.Manager
plugin.PluginData → core.PluginData

// 新
import "github.com/sky-xhsoft/sky-server/plugins/core"
*core.Manager
core.PluginData
```

## 相关文档

- [元数据初始化工具](./metadata-init-guide.md)
- [插件开发示例](./plugin-examples.md)
- [权限系统](./admin-permission-feature.md)

## 总结

Sky-Server 插件系统提供了强大而灵活的扩展机制，通过简单的接口和自动注册机制，可以快速开发和部署插件，在不修改核心代码的情况下扩展业务功能。

**关键优势**:
- ✅ 自动注册，简化开发
- ✅ 优先级控制，灵活编排
- ✅ 事务支持，保证一致性
- ✅ 动态管理，运行时控制
- ✅ 易于扩展，快速开发

通过合理使用插件系统，可以保持代码的可维护性和可扩展性！🎉
