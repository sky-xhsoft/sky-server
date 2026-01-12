# 插件系统迁移总结

**日期**: 2026-01-12
**版本**: v2.0

## 概述

本次更新将插件系统从 `internal/plugin/cmd` 迁移到根目录 `plugins/`，并实现了动态插件加载功能。

## 主要变更

### 1. 目录结构重组

**旧结构**:
```
internal/plugin/cmd/
├── plugin.go
├── setup.go
├── sys_table_after_create.go
└── sys_table_before_delete.go
```

**新结构**:
```
plugins/
├── core/                      # 核心接口和管理器
│   ├── plugin.go             # Plugin 接口定义
│   └── manager.go            # 插件管理器
├── registry/                  # 插件注册中心
│   ├── registry.go           # 全局注册中心
│   └── loader.go             # 插件加载器
├── builtin/                   # 内置插件
│   ├── sys_table_after_create.go
│   ├── sys_table_before_delete.go
│   └── utils.go
├── setup.go                   # 插件系统初始化
└── README.md                  # 说明文档
```

### 2. 新增功能

#### 动态加载机制

- **插件注册中心** (`registry/registry.go`): 全局插件工厂注册
- **插件加载器** (`registry/loader.go`): 支持按名称、钩子点动态加载插件
- **自动注册**: 插件通过 `init()` 函数自动注册到注册中心

#### 增强的插件接口

```go
// 旧接口（只有 Name 和 Execute）
type Plugin interface {
    Name() string
    Execute(ctx context.Context, db *gorm.DB, data PluginData) error
}

// 新接口（增加 Description 和 Version）
type Plugin interface {
    Name() string
    Description() string  // 新增
    Version() string      // 新增
    Execute(ctx context.Context, db *gorm.DB, data PluginData) error
}
```

#### 插件元数据

```go
type PluginMetadata struct {
    Name        string  // 插件名称
    Description string  // 插件描述
    Version     string  // 插件版本
    Author      string  // 插件作者
    Enabled     bool    // 是否启用
    Priority    int     // 执行优先级
    HookPoint   string  // 钩子点
}
```

#### 增强的插件数据

```go
type PluginData struct {
    TableName string                 // 表名
    Action    string                 // 操作类型
    Timing    string                 // 执行时机（新增）
    RecordID  uint                   // 记录ID
    Data      map[string]interface{} // 数据内容
    UserID    uint                   // 操作用户ID
    CompanyID uint                   // 公司ID
    Extra     map[string]interface{} // 额外上下文（新增）
}
```

### 3. 代码更新

#### 导入路径变更

**修改的文件**:
1. `cmd/server/main.go`
2. `internal/service/crud/crud_service.go`

**变更内容**:

```go
// 旧导入
import plugin "github.com/sky-xhsoft/sky-server/internal/plugin/cmd"

// 新导入（main.go）
import "github.com/sky-xhsoft/sky-server/plugins"

// 新导入（crud_service.go）
import "github.com/sky-xhsoft/sky-server/plugins/core"
```

#### 类型更新

```go
// 旧类型
*plugin.Manager
plugin.PluginData

// 新类型
*core.Manager
core.PluginData
```

### 4. 编译验证

✅ **编译成功**:

```bash
$ go build -o bin/sky-server ./cmd/server
# 编译成功，无错误
```

## 使用变更

### 插件开发流程

**旧流程**:
1. 创建插件结构体
2. 实现 Plugin 接口
3. 在 `setup.go` 中手动调用 `NewXxxPlugin()` 和 `Register()`

**新流程**:
1. 创建插件结构体
2. 实现 Plugin 接口（包括新方法）
3. 使用 `init()` 自动注册
4. 在 `setup.go` 中导入插件包

### 插件注册示例

**旧方式**:
```go
// 在 setup.go 中
sysTablePlugin := NewSysTableAfterCreatePlugin()
pluginManager.Register("sys_table", sysTablePlugin)
```

**新方式**:
```go
// 在插件文件中
func init() {
    registry.Register(
        "sys_table_after_create",
        func() core.Plugin { return &SysTableAfterCreatePlugin{} },
        core.PluginMetadata{
            Name:      "sys_table_after_create",
            Version:   "1.0.0",
            Enabled:   true,
            Priority:  10,
            HookPoint: "sys_table.after.create",
        },
    )
}

// 在 setup.go 中只需导入
import _ "github.com/sky-xhsoft/sky-server/plugins/builtin"
```

## 优势分析

### 1. 更好的组织结构 ✅

- **职责分离**: core、registry、builtin 各司其职
- **易于扩展**: 新插件只需创建新包并导入
- **清晰的层次**: 核心接口 → 注册中心 → 具体插件

### 2. 动态加载能力 ✅

```go
// 动态加载所有插件
loader.LoadAll()

// 按钩子点加载
loader.LoadByHookPoint("order.after.create")

// 按名称加载
loader.LoadByNames([]string{"plugin1", "plugin2"})

// 重新加载
loader.Reload("plugin1")
```

### 3. 自动注册机制 ✅

- **无需手动注册**: `init()` 自动执行
- **减少错误**: 避免忘记注册
- **简化开发**: 开发者只需实现接口

### 4. 更丰富的元数据 ✅

- **版本管理**: 支持插件版本追踪
- **描述信息**: 便于理解插件功能
- **优先级控制**: 精确控制执行顺序
- **启用/禁用**: 运行时控制插件状态

### 5. 更灵活的钩子点 ✅

```go
// 支持更精确的钩子点定义
"sys_table.after.create"  // 表名.时机.操作
"order.before.update"
"*.after.*"  // 通配符支持（计划中）
```

## 向后兼容

### 内置插件保持兼容 ✅

- `sys_table_after_create`: 功能完全保留
- `sys_table_before_delete`: 功能完全保留

### 接口兼容性 ⚠️

新接口要求实现 `Description()` 和 `Version()` 方法，但可以返回空字符串：

```go
func (p *OldPlugin) Description() string { return "" }
func (p *OldPlugin) Version() string     { return "1.0.0" }
```

## 迁移检查清单

如果你有自定义插件，需要进行以下更新：

- [ ] 更新导入路径：`internal/plugin/cmd` → `plugins/core`
- [ ] 实现新方法：`Description()` 和 `Version()`
- [ ] 使用 `init()` 注册插件
- [ ] 更新插件元数据，添加 `HookPoint`
- [ ] 在 `plugins/setup.go` 中导入插件包
- [ ] 测试编译和运行

### 迁移示例

**旧代码**:
```go
package myplugin

import (
    "context"
    plugin "github.com/sky-xhsoft/sky-server/internal/plugin/cmd"
    "gorm.io/gorm"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my_plugin"
}

func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data plugin.PluginData) error {
    // ...
    return nil
}
```

**新代码**:
```go
package myplugin

import (
    "context"
    "github.com/sky-xhsoft/sky-server/plugins/core"
    "github.com/sky-xhsoft/sky-server/plugins/registry"
    "gorm.io/gorm"
)

type MyPlugin struct{}

func init() {
    registry.Register(
        "my_plugin",
        func() core.Plugin { return &MyPlugin{} },
        core.PluginMetadata{
            Name:        "my_plugin",
            Description: "我的自定义插件",
            Version:     "1.0.0",
            Enabled:     true,
            Priority:    10,
            HookPoint:   "order.after.create",
        },
    )
}

func (p *MyPlugin) Name() string        { return "my_plugin" }
func (p *MyPlugin) Description() string { return "我的自定义插件" }
func (p *MyPlugin) Version() string     { return "1.0.0" }

func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // ...
    return nil
}
```

## 文档更新

### 新增文档

1. **`docs/plugin-system.md`**: 插件系统完整使用指南
2. **`plugins/README.md`**: 插件目录快速入门
3. **`docs/plugin-migration-summary.md`**: 本迁移总结文档

### 更新文档

需要更新引用旧插件路径的文档（如果有）。

## 已知限制

### 1. 通配符钩子点 🚧

当前 `*.after.*` 等通配符钩子点需要在 Manager 中实现匹配逻辑。

**计划**: 在后续版本实现。

### 2. 插件依赖 🚧

当前插件之间没有依赖管理机制。

**建议**: 使用优先级和 `Extra` 数据传递结果。

### 3. 插件配置 🚧

当前插件配置通过代码硬编码。

**计划**: 支持通过配置文件或数据库配置插件。

## 后续优化建议

### 1. 配置文件支持

```yaml
plugins:
  - name: order_validation
    enabled: true
    priority: 5
    config:
      max_amount: 10000
```

### 2. 插件市场

- 标准化插件打包格式
- 支持插件下载和安装
- 插件依赖管理

### 3. 插件监控

- 插件执行统计
- 性能监控
- 错误追踪

### 4. 插件 UI 管理

- Web 界面管理插件
- 动态启用/禁用
- 查看插件日志

## 相关文档

- [插件系统使用指南](./plugin-system.md)
- [内置插件文档](../plugins/builtin/README.md)（待创建）
- [插件开发示例](./plugin-examples.md)（待创建）

## 总结

本次迁移成功将插件系统升级到 v2.0，主要改进包括：

✅ **目录结构优化**: 更清晰的组织方式
✅ **动态加载**: 支持运行时加载插件
✅ **自动注册**: 简化插件开发流程
✅ **元数据支持**: 丰富的插件信息
✅ **向后兼容**: 内置插件功能保持一致

新的插件系统为 Sky-Server 提供了更强大、更灵活的扩展能力，同时保持了简单易用的开发体验。🎉
