# 插件热加载系统 🔥

**JSP 风格的 Go 插件热加载**

## 概述

此包实现了类似 JSP 的插件热加载机制：
- 运行时自动编译 `.go` 源码为 `.so` 插件
- 动态加载到运行中的程序
- 文件变化时自动重新编译和加载
- 无需重启服务器

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                      HotloadManager                     │
│  负责：自动扫描、编译、加载、监听、热重载                 │
│  自动发现 plugins/runtime/ 目录中的所有插件              │
└──────────────┬──────────────┬───────────────┬──────────┘
               │              │               │
       ┌───────▼──────┐ ┌────▼─────┐   ┌────▼────────┐
       │   Compiler   │ │  Loader  │   │   Watcher   │
       │  运行时编译   │ │ 动态加载  │   │  文件监听   │
       └──────────────┘ └──────────┘   └─────────────┘
```

## 核心组件

### 1. HotloadManager

**职责**：统一管理编译、加载、热重载流程

**关键方法**：
- `scanAndLoadAll()` - 扫描并加载所有插件
- `loadPlugin(pluginPath)` - 加载单个插件
- `Start()` - 启动自动扫描和文件监听
- `Stop()` - 停止文件监听

**设计原则**：
- ✅ 自动发现，扫描 plugins/runtime/ 目录
- ✅ 无需数据库配置
- ✅ 简单易用

### 2. Compiler

**职责**：将 Go 源码编译成 `.so` 插件

**关键方法**：
- `Compile(pluginPath)` - 编译插件
- `NeedsRecompile(pluginPath, lastHash)` - 检查是否需要重编译
- `Clean(pluginName)` - 清理编译产物

**技术实现**：
- 使用 `go build -buildmode=plugin` 编译
- SHA256 哈希检测文件变化
- 支持增量编译

### 3. Loader

**职责**：动态加载编译后的 `.so` 文件

**关键方法**：
- `Load(pluginName, soPath, hookPoint, priority)` - 加载插件
- `Unload(pluginName, hookPoint)` - 卸载插件
- `Reload(pluginName, ...)` - 重新加载

**技术实现**：
- 使用 Go 的 `plugin` 包
- 查找并调用 `Register()` 符号
- 支持插件覆盖（重新加载）

### 4. Watcher

**职责**：监听源码文件变化

**关键方法**：
- `Start()` - 启动监听
- `Stop()` - 停止监听

**技术实现**：
- 使用 `fsnotify` 包
- 防抖动处理（默认 2 秒）
- 仅监听 `.go` 文件

## 目录结构要求

### 插件目录规范

```
plugins/runtime/
├── my_plugin/          ✅ 正常插件（会被加载）
│   └── plugin.go
├── _disabled_plugin/   ⏭️ 下划线开头（跳过）
│   └── plugin.go
├── .hidden/            ⏭️ 点号开头（跳过）
│   └── plugin.go
└── TEMPLATE.go         ⏭️ 非目录文件（跳过）
```

### 自动发现规则

1. **扫描目录**：系统启动时自动扫描 `plugins/runtime/`
2. **识别插件**：每个子目录视为一个插件
3. **跳过规则**：
   - 以 `.` 开头的目录（隐藏）
   - 以 `_` 开头的目录（禁用标记）
   - 非目录项
4. **自动编译**：所有符合规则的插件都会被编译
5. **自动加载**：编译成功后自动加载到运行时

## 使用示例

### 基础用法

```go
// 1. 创建热加载管理器（使用默认配置）
mgr, err := hotload.NewHotloadManager(pluginManager, hotload.DefaultConfig())
if err != nil {
    log.Fatal(err)
}

// 2. 启动管理器（自动扫描、编译、加载所有插件）
if err := mgr.Start(); err != nil {
    log.Fatal(err)
}

// 就这么简单！系统会自动：
// - 扫描 plugins/runtime/ 目录
// - 编译所有找到的插件
// - 加载到运行时
// - 启动文件监听器
```

### 自定义配置

```go
config := &hotload.Config{
    RuntimeDir:   "plugins/runtime",        // 插件源码目录
    CompiledDir:  "plugins/compiled",       // 编译输出目录
    ModulePath:   "github.com/your/project", // Go 模块路径
    DebounceTime: 3 * time.Second,          // 文件变化防抖时间
    EnableWatch:  true,                     // 启用文件监听
}

mgr, _ := hotload.NewHotloadManager(pluginManager, config)
mgr.Start()
```

## 工作流程

### 启动时

```
1. HotloadManager.Start()
   └─ scanAndLoadAll()

2. scanAndLoadAll()
   ├─ 读取 plugins/runtime/ 目录
   ├─ 遍历所有子目录
   ├─ 跳过 . 和 _ 开头的目录
   └─ 对每个插件调用 loadPlugin()

3. loadPlugin(pluginPath)
   ├─ Compiler.Compile(path) → .so
   └─ Loader.Load(name, soPath)

4. 启动文件监听器
   └─ Watcher.Start()
```

### 运行时（热重载）

```
1. Watcher 检测到文件变化
   ├─ 防抖动处理（2秒）
   └─ 触发 handleFileChange

2. handleFileChange(pluginPath)
   └─ 调用 loadPlugin() 重新编译和加载

3. loadPlugin(pluginPath)
   ├─ Compiler.Compile(path) → 新 .so
   ├─ Loader.Load(name, newSoPath)
   └─ 覆盖旧版本（自动替换）

4. 新版本立即生效 ✨
```

## 限制与注意事项

### 平台限制

- ✅ **Linux/macOS**: 完全支持（Go plugin 包）
- ❌ **Windows**: 不支持

### 技术限制

1. **Go 版本一致性**：插件和主程序必须用相同 Go 版本编译
2. **依赖版本一致性**：依赖包版本必须一致
3. **内存无法释放**：Go plugin 包限制，卸载不释放内存
4. **CGO 要求**：必须启用 CGO (`CGO_ENABLED=1`)

### 插件要求

1. **Package main**: 插件必须是 `package main`
2. **导出 Register**: 必须导出 `func Register()`
3. **路径一致**：源码目录名必须与配置中的 CONTENT 一致

## 文件结构

```
plugins/
├── hotload/
│   ├── manager.go          # 热加载管理器
│   ├── compiler.go         # 运行时编译器
│   ├── loader.go           # 动态加载器
│   ├── watcher.go          # 文件监听器
│   ├── config_loader.go    # 配置加载器
│   └── README.md           # 本文件
│
├── runtime/                # 插件源码目录
│   ├── TEMPLATE.go         # 插件模板
│   └── my_plugin/          # 用户插件
│       └── plugin.go
│
└── compiled/               # 编译产物目录
    └── my_plugin.so
```

## 相关文档

- [快速入门](../../docs/plugin-hotload-quickstart.md)
- [完整指南](../../docs/plugin-hotload-guide.md)
- [SQL 示例](../../docs/plugin-hotload-examples.sql)
- [插件模板](../runtime/TEMPLATE.go)

## 性能考虑

- **编译时间**：首次编译约 1-3 秒
- **加载时间**：加载 .so 文件约 10-100ms
- **内存占用**：每个插件约 1-10MB
- **文件监听**：防抖动 2 秒，避免频繁重编译

## 故障排除

### 编译失败

- 检查 Go 版本一致性
- 检查依赖包是否安装
- 查看编译错误日志

### 加载失败

- 确认 .so 文件存在
- 检查 Register 函数是否导出
- 确认 package 是否为 main

### 热重载不生效

- 查看文件监听器是否启动（检查日志）
- 确认文件变化被检测到（查看日志）
- 检查防抖动时间（默认 2 秒），等待几秒后再测试
- 确认插件目录结构正确
