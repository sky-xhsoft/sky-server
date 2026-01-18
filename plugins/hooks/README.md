# Hooks 自动注册机制

## 概述

本包提供了一个自动注册机制，用于管理 Go 钩子函数。所有 hooks 通过 `init()` 函数自动注册，无需在 `setup.go` 中手动调用注册函数。

## 核心概念

### 1. HookRegistrar 接口

所有 hook 必须实现此接口：

```go
type HookRegistrar interface {
    Name() string                          // 返回 hook 名称
    Register(manager *core.Manager)        // 执行注册逻辑
}
```

### 2. 自动注册流程

1. 每个 hook 文件在 `init()` 中调用 `hooks.Register(hook)` 注册自己
2. `plugins.Setup()` 调用 `hooks.RegisterAll(manager)` 统一注册所有 hooks
3. 所有 hooks 自动注册到 `executor.GoFuncRegistry`

## 如何添加新的 Hook

### 方法一：使用 BaseHook（推荐）

这是最简单的方式，适合大多数场景：

```go
package hooks

import (
    "context"
    "github.com/sky-xhsoft/sky-server/internal/pkg/logger"
    "github.com/sky-xhsoft/sky-server/plugins/core"
    "go.uber.org/zap"
)

// SysUserAfterCreateHook 示例：用户创建后钩子
type SysUserAfterCreateHook struct {
    *BaseHook
}

// 在 init() 中自动注册
func init() {
    hook := &SysUserAfterCreateHook{
        BaseHook: NewBaseHook("SYS_USER_AFTER_CREATE", sysUserAfterCreateHandler),
    }
    Register(hook)
}

// 处理函数
func sysUserAfterCreateHandler(manager *core.Manager) func(map[string]interface{}) (interface{}, error) {
    return func(params map[string]interface{}) (interface{}, error) {
        logger.Info("执行 SYS_USER_AFTER_CREATE 钩子", zap.Any("params", params))

        // 获取数据库连接
        txDB, err := GetDBFromParams(params)
        if err != nil {
            return nil, err
        }

        // 获取必要的参数
        recordID, err := GetUintFromParams(params, "ID")
        if err != nil {
            return nil, err
        }

        companyID := GetUintOrZero(params, "SYS_COMPANY_ID")
        username := GetStringOrEmpty(params, "USERNAME")

        // 构造插件数据
        pluginData := core.PluginData{
            TableName: "sys_user",
            Action:    "create",
            Timing:    "after",
            RecordID:  recordID,
            CompanyID: companyID,
            Data:      params,
        }

        // 执行插件
        ctx := context.Background()
        if err := manager.ExecuteWithDB(ctx, txDB, pluginData); err != nil {
            logger.Error("执行插件失败", zap.Error(err))
            return nil, err
        }

        // 这里可以添加自定义逻辑
        logger.Info("用户创建成功",
            zap.Uint("recordID", recordID),
            zap.String("username", username))

        return SuccessResult("sys_user 创建后钩子执行成功"), nil
    }
}
```

### 方法二：自定义实现

如果需要更复杂的逻辑，可以自己实现 `HookRegistrar` 接口：

```go
package hooks

import (
    "github.com/sky-xhsoft/sky-server/internal/pkg/executor"
    "github.com/sky-xhsoft/sky-server/plugins/core"
)

// CustomHook 自定义 hook 示例
type CustomHook struct {
    name    string
    manager *core.Manager
}

func init() {
    Register(&CustomHook{name: "CUSTOM_HOOK"})
}

func (h *CustomHook) Name() string {
    return h.name
}

func (h *CustomHook) Register(manager *core.Manager) {
    h.manager = manager

    executor.RegisterGoFunc(h.name, func(params map[string]interface{}) (interface{}, error) {
        // 自定义处理逻辑
        return map[string]interface{}{
            "success": true,
            "message": "自定义钩子执行成功",
        }, nil
    })
}
```

## 工具函数

### 参数提取

- `GetDBFromParams(params)` - 获取数据库连接
- `GetUintFromParams(params, key)` - 获取 uint 参数（失败抛错）
- `GetUintOrZero(params, key)` - 获取 uint 参数（失败返回 0）
- `GetStringFromParams(params, key)` - 获取 string 参数（失败抛错）
- `GetStringOrEmpty(params, key)` - 获取 string 参数（失败返回空字符串）

### 返回结果

- `SuccessResult(message)` - 创建成功的返回结果
- `ErrorResult(message)` - 创建错误的返回结果

## 命名约定

### Hook 名称格式

```
<TABLE_NAME>_<TIMING>_<ACTION>
```

示例：
- `SYS_TABLE_AFTER_CREATE`
- `SYS_TABLE_BEFORE_DELETE`
- `SYS_USER_AFTER_UPDATE`
- `SYS_COMPANY_BEFORE_CREATE`

### 文件命名

```
<table_name>_<timing>_<action>.go
```

示例：
- `sys_table_after_create.go`
- `sys_table_before_delete.go`
- `sys_user_after_update.go`

### 类型命名

```go
type <TableName><Timing><Action>Hook struct {
    *BaseHook
}
```

示例：
- `SysTableAfterCreateHook`
- `SysUserBeforeDeleteHook`

## 调试

### 查看已注册的 Hooks

在应用启动时，会在日志中看到：

```
INFO    Go 钩子函数已自动注册到执行器    {"count": 2, "hooks": ["SYS_TABLE_AFTER_CREATE", "SYS_TABLE_BEFORE_DELETE"]}
```

### 获取已注册的 Hooks 列表

```go
registeredHooks := hooks.GetRegisteredHooks()
fmt.Println(registeredHooks)
// 输出: ["SYS_TABLE_AFTER_CREATE", "SYS_TABLE_BEFORE_DELETE"]
```

## 优势

### 与旧方式对比

**旧方式（手动注册）：**
```go
// 在 setup.go 中
func registerGoHooks(manager *core.Manager) {
    registerSysTableAfterCreateHook(manager)
    registerSysTableBeforeDeleteHook(manager)
    registerSysUserAfterCreateHook(manager)  // 每次添加都要改这里
    // ... 更多 hook 注册
}
```

**新方式（自动注册）：**
```go
// 在 setup.go 中
func registerGoHooks(manager *core.Manager) {
    hooks.RegisterAll(manager)  // 一行搞定，新增 hook 无需修改
}

// 在新的 hook 文件中
func init() {
    hook := &MyNewHook{...}
    Register(hook)  // 自动注册
}
```

### 优点

1. ✅ **零配置添加**：新增 hook 只需创建文件，无需修改 `setup.go`
2. ✅ **统一管理**：所有 hooks 在一个包中管理
3. ✅ **代码复用**：通过 `BaseHook` 和工具函数减少重复代码
4. ✅ **类型安全**：编译时检查，避免运行时错误
5. ✅ **易于测试**：每个 hook 独立文件，便于单元测试
6. ✅ **自动发现**：通过 `GetRegisteredHooks()` 可以查看所有已注册的 hooks

## 示例项目结构

```
plugins/
├── hooks/
│   ├── README.md                          # 本文档
│   ├── registry.go                        # 注册机制核心
│   ├── utils.go                           # 工具函数
│   ├── sys_table_after_create.go         # 表创建后 hook
│   ├── sys_table_before_delete.go        # 表删除前 hook
│   ├── sys_user_after_create.go          # 用户创建后 hook (示例)
│   └── sys_company_before_update.go      # 公司更新前 hook (示例)
└── setup.go                               # 只需要 hooks.RegisterAll(manager)
```

## 注意事项

1. **包导入顺序**：确保在 `setup.go` 中导入 hooks 包
   ```go
   import (
       // 导入 hooks 包以触发 init() 自动注册所有 hooks
       "github.com/sky-xhsoft/sky-server/plugins/hooks"
   )
   ```

2. **init() 函数**：每个 hook 文件必须有 `init()` 函数来注册自己

3. **Manager 依赖**：所有 hooks 都依赖 `*core.Manager`，在 handler 闭包中使用

4. **事务安全**：从 `params["__db__"]` 获取的是事务连接，确保在事务中执行

## 迁移指南

如果你有旧的手动注册的 hooks，按以下步骤迁移：

1. 在 `plugins/hooks/` 目录创建新文件
2. 将 handler 逻辑复制到新文件的 handler 函数中
3. 添加 `init()` 函数注册 hook
4. 从 `setup.go` 删除旧的 `registerXxxHook` 函数
5. 验证 hooks 正常工作

完成！🎉
