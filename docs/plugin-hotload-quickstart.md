# Go 插件热加载快速入门 🔥

**实现日期**: 2026-01-13
**版本**: v3.0 - 自动发现，无需数据库配置

## 🎯 什么是插件热加载？

类似 JSP 的热加载机制：
- 📝 编写 `.go` 源码到 `plugins/runtime/`
- 🔨 系统**自动编译**成 `.so` 插件
- ⚡ **自动加载**到运行时
- 🔄 修改代码后**自动热重载**，无需重启！
- ✨ **无需数据库配置**，自动发现所有插件！

## 🚀 2 步创建热加载插件

### 步骤 1: 创建插件源码

```bash
# 创建插件目录
mkdir -p plugins/runtime/order_notify

# 复制模板
cp plugins/runtime/TEMPLATE.go plugins/runtime/order_notify/plugin.go

# 编辑插件
vim plugins/runtime/order_notify/plugin.go
```

插件代码示例：

```go
package main

import (
    "context"
    "fmt"
    "github.com/sky-xhsoft/sky-server/plugins/core"
    "github.com/sky-xhsoft/sky-server/plugins/registry"
    "gorm.io/gorm"
)

type OrderNotifyPlugin struct{}

// Register 必须导出
func Register() {
    registry.Register("order_notify",
        func() core.Plugin { return &OrderNotifyPlugin{} },
        core.PluginMetadata{
            Name:        "order_notify",
            HookPoint:   "order.after.create",
            Priority:    50,
            Enabled:     true,
        })
}

func (p *OrderNotifyPlugin) Name() string { return "order_notify" }
func (p *OrderNotifyPlugin) Description() string { return "订单通知" }
func (p *OrderNotifyPlugin) Version() string { return "1.0.0" }

func (p *OrderNotifyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
    fmt.Printf("📧 新订单: %v\n", data.Data["ORDER_NO"])
    return nil
}
```

### 步骤 2: 启动服务器

```bash
go run cmd/server/main.go
```

系统会自动：
1. ✅ 扫描 `plugins/runtime/` 目录
2. ✅ 发现所有插件（包括 order_notify）
3. ✅ 编译：`plugins/runtime/order_notify/` → `plugins/compiled/order_notify.so`
4. ✅ 加载插件到运行时
5. ✅ 启动文件监听器

### 🎉 完成！测试热重载

```bash
# 修改插件代码
vim plugins/runtime/order_notify/plugin.go

# 保存后，系统自动：
# ✅ 检测文件变化
# ✅ 重新编译
# ✅ 重新加载
# ✅ 新代码立即生效！
```

## 📋 配置说明

### 自动发现规则

系统会自动扫描 `plugins/runtime/` 目录并加载所有插件：

- **目录名即插件名**：`plugins/runtime/order_notify/` → 插件名为 `order_notify`
- **跳过隐藏目录**：以 `.` 或 `_` 开头的目录会被跳过
- **自动编译加载**：满足规则的所有插件都会被编译和加载
- **无需数据库配置**：不需要在任何表中配置

## 🔧 常用操作

### 禁用插件

如果不想加载某个插件，可以：

1. **重命名目录**（添加下划线前缀）：
   ```bash
   mv plugins/runtime/order_notify plugins/runtime/_order_notify
   ```

2. **删除目录**：
   ```bash
   rm -rf plugins/runtime/order_notify
   ```

### 启用插件

恢复目录名称或重新创建即可：
```bash
mv plugins/runtime/_order_notify plugins/runtime/order_notify
```

### 查看已加载的插件

查看服务器启动日志：
```
[INFO] 插件扫描完成 total=3 loaded=3
[INFO] 插件加载完成 totalPlugins=3 hookPoints=5
```

## ⚠️ 注意事项

1. **目录结构要求**：
   - 插件必须放在独立目录中：`plugins/runtime/插件名/plugin.go`
   - 一个目录代表一个插件

2. **Package 必须是 main**：
   ```go
   package main  // ✅ 正确
   ```

3. **必须导出 Register 函数**：
   ```go
   func Register() {  // ✅ 首字母大写
       // ...
   }
   ```

4. **平台限制**：
   - ✅ Linux/macOS：完全支持
   - ❌ Windows：不支持（可用 WSL2）

5. **自动加载所有插件**：
   - 系统会加载 `plugins/runtime/` 下所有符合规则的插件
   - 如需禁用，请重命名或删除目录

## 📚 更多文档

- **完整文档**：`docs/plugin-hotload-guide.md`
- **插件模板**：`plugins/runtime/TEMPLATE.go`
- **示例插件**：`plugins/runtime/example_hotload/`

---

**享受 JSP 风格的热加载！只需放入目录，自动编译加载！** 🔥✨
