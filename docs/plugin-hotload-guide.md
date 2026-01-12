# Go 插件热加载系统使用指南 🔥

**实现日期**: 2026-01-13
**版本**: v3.0 - 自动发现

## 🎯 概述

Sky-Server 现在支持 **JSP 风格的插件热加载**！

开发者只需：
1. 编写 `.go` 插件源码
2. 放到 `plugins/runtime/` 目录
3. **系统自动编译、加载、热重载** ✨

**无需数据库配置！无需重启服务器！修改插件代码后自动生效！**

## 🆚 对比：静态插件 vs 热加载插件

| 特性 | 静态插件 | 热加载插件 |
|------|---------|-----------|
| 编译时机 | 编译主程序时 | 运行时动态编译 |
| 加载方式 | 静态导入 | 动态加载 .so |
| 修改生效 | 需要重启服务器 | **自动热重载** ✨ |
| 适用场景 | 核心系统插件 | 业务定制插件 |
| 配置方式 | 代码注册 | **自动发现** ✨ |
| 平台支持 | 跨平台 | Linux/macOS |

## 🏗️ 架构说明

### 目录结构

```
plugins/
├── runtime/              # 插件源码目录（用户编写）
│   ├── example_hotload/  # 示例插件
│   │   └── plugin.go
│   ├── my_plugin/        # 你的插件
│   │   └── plugin.go
│   └── TEMPLATE.go       # 插件模板
│
├── compiled/             # 编译后的 .so 文件（自动生成）
│   ├── example_hotload.so
│   └── my_plugin.so
│
├── hotload/              # 热加载管理器（系统代码）
│   ├── manager.go        # 热加载管理器
│   ├── compiler.go       # 运行时编译器
│   ├── loader.go         # 动态加载器
│   └── watcher.go        # 文件监听器
│
└── ...
```

### 工作流程

```
1. 用户编写插件源码
   ├─ plugins/runtime/myplugin/plugin.go
   └─ 必须导出 Register() 函数

2. 系统启动时
   ├─ 自动扫描 plugins/runtime/ 目录
   ├─ 发现所有插件目录
   ├─ 自动编译 .go → .so
   ├─ 动态加载 .so 文件
   └─ 启动文件监听器

3. 运行时（JSP 风格热重载）
   ├─ 监听文件变化
   ├─ 检测到修改
   ├─ 自动重新编译
   ├─ 自动重新加载
   └─ 插件立即生效 ✨
```

## 📚 快速开始

### 步骤 1: 创建插件源码

```bash
# 创建插件目录
mkdir -p plugins/runtime/order_notify

# 复制模板
cp plugins/runtime/TEMPLATE.go plugins/runtime/order_notify/plugin.go

# 编辑插件代码
vim plugins/runtime/order_notify/plugin.go
```

示例插件代码：

```go
package main

import (
    "context"
    "fmt"
    "github.com/sky-xhsoft/sky-server/internal/pkg/logger"
    "github.com/sky-xhsoft/sky-server/plugins/core"
    "github.com/sky-xhsoft/sky-server/plugins/registry"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type OrderNotifyPlugin struct{}

// Register 必须导出（首字母大写）
func Register() {
    registry.Register(
        "order_notify",
        func() core.Plugin { return &OrderNotifyPlugin{} },
        core.PluginMetadata{
            Name:        "order_notify",
            Description: "订单通知插件",
            Version:     "1.0.0",
            Author:      "Your Name",
            Enabled:     true,
            Priority:    50,
            HookPoint:   "order.after.create",
        },
    )
}

func (p *OrderNotifyPlugin) Name() string { return "order_notify" }
func (p *OrderNotifyPlugin) Description() string { return "订单通知插件" }
func (p *OrderNotifyPlugin) Version() string { return "1.0.0" }

func (p *OrderNotifyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    logger.Info("📧 发送订单通知", zap.Uint("orderID", data.RecordID))

    orderNo := data.Data["ORDER_NO"]
    fmt.Printf("✉️ 新订单: %v\n", orderNo)

    // 这里添加你的业务逻辑
    // - 发送邮件
    // - 发送短信
    // - 调用外部API

    return nil
}
```

### 步骤 2: 启动服务器

```bash
# 启动服务器
go run cmd/server/main.go
```

系统会自动：
1. ✅ 扫描 `plugins/runtime/` 目录
2. ✅ 发现所有插件（包括 order_notify）
3. ✅ 编译插件：`plugins/runtime/order_notify/` → `plugins/compiled/order_notify.so`
4. ✅ 加载插件到运行时
5. ✅ 启动文件监听器

### 步骤 3: 测试热重载

```bash
# 1. 修改插件代码
vim plugins/runtime/order_notify/plugin.go

# 2. 修改 Execute 方法，例如改个日志输出
func (p *OrderNotifyPlugin) Execute(...) error {
    logger.Info("🚀 这是修改后的新日志！")  # 新增这行
    // ...
}

# 3. 保存文件
# 系统会自动：
#   - 检测到文件变化
#   - 重新编译
#   - 重新加载
#   - 新代码立即生效 ✨

# 4. 触发钩子验证
curl -X POST http://localhost:8080/api/crud/order -d '{...}'

# 5. 查看日志，应该看到新的日志输出
```

## 🔧 插件开发规范

### 必须遵守的规则

#### 1. Package 必须是 main

```go
package main  // ✅ 正确

package myplugin  // ❌ 错误，无法编译成插件
```

#### 2. 必须导出 Register 函数

```go
// ✅ 正确：首字母大写，导出符号
func Register() {
    registry.Register(...)
}

// ❌ 错误：首字母小写，无法被系统调用
func register() {
    registry.Register(...)
}
```

#### 3. 插件名称必须一致

```go
// 数据库配置
PLUGIN_NAME = 'order_notify'

// 代码中也必须一致
registry.Register(
    "order_notify",  // 必须与数据库一致
    ...
)
```

#### 4. 钩子点必须一致

```go
// 数据库配置
HOOK_POINT = 'order.after.create'

// 代码中也必须一致
core.PluginMetadata{
    HookPoint: "order.after.create",  // 必须与数据库一致
}
```

### 推荐最佳实践

#### 1. 使用模板创建插件

```bash
cp plugins/runtime/TEMPLATE.go plugins/runtime/myplugin/plugin.go
```

#### 2. 合理设置优先级

- 验证类插件: 1-10（先执行）
- 业务逻辑: 10-50
- 通知类插件: 50-100（后执行）

#### 3. 错误处理

```go
func (p *MyPlugin) Execute(...) error {
    // 如果返回错误，会中断后续插件执行
    if err := validateData(data.Data); err != nil {
        return fmt.Errorf("数据验证失败: %w", err)
    }

    // 通知类操作应该处理错误，不要影响主流程
    if err := sendNotification(); err != nil {
        logger.Error("发送通知失败", zap.Error(err))
        // 不返回错误，让流程继续
    }

    return nil
}
```

#### 4. 使用事务

```go
func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    // db 已经在事务中，可以直接使用
    var relatedData SomeModel
    if err := db.Where("...").First(&relatedData).Error; err != nil {
        return err
    }

    // 创建相关记录也会在同一事务中
    if err := db.Create(&NewRecord{...}).Error; err != nil {
        return err
    }

    return nil  // 成功时事务会提交，失败时会回滚
}
```

## 📂 目录管理

### 自动发现规则

插件通过目录结构自动发现，无需数据库配置：

1. **目录即插件**：
   - 每个子目录代表一个插件
   - 目录名即插件名
   - 示例：`plugins/runtime/order_notify/` → 插件名为 `order_notify`

2. **跳过规则**：
   - 以 `.` 开头的目录（隐藏目录）
   - 以 `_` 开头的目录（禁用标记）
   - 非目录文件会被忽略

3. **自动加载**：
   - 所有符合规则的插件都会被自动编译和加载
   - 不需要在数据库中配置
   - 不需要手动导入

### 插件禁用方法

如果需要禁用某个插件：

1. **方法 1：重命名目录**（推荐）
   ```bash
   mv plugins/runtime/order_notify plugins/runtime/_order_notify
   ```

2. **方法 2：删除目录**
   ```bash
   rm -rf plugins/runtime/order_notify
   ```

3. **方法 3：移出目录**
   ```bash
   mv plugins/runtime/order_notify /tmp/
   ```

## ⚙️ 配置选项

### 热加载管理器配置

在 `plugins/setup.go` 中可以自定义配置：

```go
config := &hotload.Config{
    RuntimeDir:   "plugins/runtime",    // 源码目录
    CompiledDir:  "plugins/compiled",   // 编译输出目录
    ModulePath:   "github.com/sky-xhsoft/sky-server",  // Go 模块路径
    DebounceTime: 2 * time.Second,      // 防抖动时间
    EnableWatch:  true,                 // 是否启用文件监听
}

hotloadMgr, err := hotload.NewHotloadManager(pluginManager, config)
```

## 🚨 常见问题

### Q1: 编译失败怎么办？

**A**: 查看服务器日志输出的错误信息：

```
[ERROR] 插件编译失败 plugin=order_notify error=...
```

常见错误：
- **找不到包**: 检查 import 路径是否正确
- **未定义的函数**: 确保导入了必要的包
- **package main**: 插件必须使用 `package main`
- **Register 未导出**: 确保函数名首字母大写

### Q2: 修改代码后没有自动重载？

**A**: 检查以下几点：
1. 查看服务器日志，确认文件监听器已启动
2. 确认文件变化被检测到（查看日志输出）
3. 检查防抖动时间（默认 2 秒），等待几秒后再触发钩子
4. 确保插件目录结构正确

### Q3: 插件加载成功，但没有执行？

**A**: 检查：
1. `HOOK_POINT` 在插件代码中是否正确设置
2. 触发的操作是否匹配钩子点
3. 查看日志确认插件是否真的被调用
4. 检查插件的 Execute 方法是否返回错误

### Q4: Windows 系统支持吗？

**A**: 当前版本仅支持 Linux/macOS，因为使用了 Go 的 plugin 包。
Windows 用户可以：
- 使用 WSL2
- 使用 Linux 虚拟机
- 等待未来的 Yaegi 解释器支持（跨平台）

### Q5: 如何禁用文件监听？

**A**: 修改 `plugins/setup.go` 中的配置：

```go
config := &hotload.Config{
    EnableWatch: false,  // 禁用文件监听
    // ...
}
```

### Q6: 如何临时禁用某个插件？

**A**: 重命名插件目录，添加下划线前缀：

```bash
mv plugins/runtime/order_notify plugins/runtime/_order_notify
```

需要时再改回来即可。

## 📝 日志说明

系统会输出详细的日志，帮助调试：

```
# 启动时
[INFO] 初始化插件系统...
[INFO] 启动热加载管理器...
[INFO] 扫描插件目录 dir=plugins/runtime
[INFO] 发现插件 plugin=order_notify
[INFO] 加载插件 plugin=order_notify
[INFO] 插件编译成功 plugin=order_notify duration=1.5s
[INFO] 插件加载成功 plugin=order_notify
[INFO] 插件扫描完成 total=3 loaded=3
[INFO] ✨ 热加载管理器已启动，支持 JSP 风格的插件动态编译和热重载

# 文件变化时
[INFO] 检测到插件文件变化 plugin=order_notify event=write
[INFO] 加载插件 plugin=order_notify
[INFO] 插件编译成功 plugin=order_notify duration=1.2s
[INFO] 插件加载成功 plugin=order_notify
[INFO] 插件热重载成功 ✨ plugin=order_notify
```

## 🎯 下一步

- [ ] 创建更多插件到 `plugins/runtime/` 目录
- [ ] 测试热重载功能
- [ ] 开发实际业务插件
- [ ] 监控插件执行性能

## 🔗 相关文档

- [快速入门指南](./plugin-hotload-quickstart.md)
- [插件模板](../plugins/runtime/TEMPLATE.go)
- [示例插件](../plugins/runtime/example_hotload/plugin.go)
- [组件文档](../plugins/hotload/README.md)

---

**享受 JSP 风格的热加载开发体验！无需数据库配置，自动发现加载！** 🔥✨
