# 插件命名规范重构总结

## 重构日期
2026-01-12

## 重构目标

将插件命名规范统一为：**表单名称_执行时机_动作**

## 修改内容

### 1. 文件重命名

**变更**:
```
internal/pkg/plugin/sys_table_standard_columns.go
→
internal/pkg/plugin/sys_table_after_create.go
```

### 2. 代码重构

#### 类型名称

**变更前**:
```go
type SysTableStandardColumnsPlugin struct{}
func NewSysTableStandardColumnsPlugin() *SysTableStandardColumnsPlugin
```

**变更后**:
```go
type SysTableAfterCreatePlugin struct{}
func NewSysTableAfterCreatePlugin() *SysTableAfterCreatePlugin
```

#### 插件名称

**变更前**:
```go
func (p *SysTableStandardColumnsPlugin) Name() string {
    return "SysTableStandardColumnsPlugin"
}
```

**变更后**:
```go
func (p *SysTableAfterCreatePlugin) Name() string {
    return "sys_table_after_create"
}
```

### 3. 主程序修改

**文件**: `cmd/server/main.go`

**变更前**:
```go
pluginManager.Register("sys_table", plugin.NewSysTableStandardColumnsPlugin())
```

**变更后**:
```go
pluginManager.Register("sys_table", plugin.NewSysTableAfterCreatePlugin())
```

### 4. 测试代码修改

**文件**: `internal/pkg/plugin/plugin_test.go`

**变更**:
- `TestSysTableStandardColumnsPlugin_Name` → `TestSysTableAfterCreatePlugin_Name`
- `TestSysTableStandardColumnsPlugin_OnlyExecuteOnCreate` → `TestSysTableAfterCreatePlugin_OnlyExecuteOnCreate`
- 测试中的期望值从 `"SysTableStandardColumnsPlugin"` 改为 `"sys_table_after_create"`

### 5. 文档更新

#### 新增文档

- **plugin-naming-convention.md** - 插件命名规范完整指南

#### 更新文档

- **plugin-system.md** - 更新插件命名说明
- **plugin-implementation-summary.md** - 更新插件文件名和命名
- **plugin-quick-start.md** - 更新示例代码

## 命名规范

### 格式

```
{表单名称}_{执行时机}_{动作}
```

### 组成部分

1. **表单名称**: 触发插件的表名（如 `sys_table`）
2. **执行时机**:
   - `before` - 在操作前执行
   - `after` - 在操作后执行
3. **动作**:
   - `create` - 创建操作
   - `update` - 更新操作
   - `delete` - 删除操作

### 示例

| 插件名称 | 说明 |
|---------|------|
| `sys_table_after_create` | sys_table创建后执行 |
| `sys_table_before_delete` | sys_table删除前执行 |
| `sys_user_after_update` | sys_user更新后执行 |
| `customer_before_create` | customer创建前执行 |

## 代码结构映射

### 文件命名

```
internal/pkg/plugin/{table_name}_{timing}_{action}.go
```

### 类型命名 (PascalCase)

```go
type {TableName}{Timing}{Action}Plugin struct{}
```

示例：
- `sys_table_after_create` → `SysTableAfterCreatePlugin`
- `sys_user_before_delete` → `SysUserBeforeDeletePlugin`

### 插件标识 (snake_case)

```go
func (p *SysTableAfterCreatePlugin) Name() string {
    return "sys_table_after_create"
}
```

## 验证结果

### 测试通过 ✅

```bash
$ go test ./internal/pkg/plugin -v

=== RUN   TestManager_Register
--- PASS: TestManager_Register (0.00s)
=== RUN   TestManager_ExecutePlugins
--- PASS: TestManager_ExecutePlugins (0.00s)
=== RUN   TestManager_ExecutePlugins_NoPlugins
--- PASS: TestManager_ExecutePlugins_NoPlugins (0.00s)
=== RUN   TestSysTableAfterCreatePlugin_Name
--- PASS: TestSysTableAfterCreatePlugin_Name (0.00s)
=== RUN   TestSysTableAfterCreatePlugin_OnlyExecuteOnCreate
--- PASS: TestSysTableAfterCreatePlugin_OnlyExecuteOnCreate (0.00s)
PASS
ok  	github.com/sky-xhsoft/sky-server/internal/pkg/plugin	0.009s
```

### 构建成功 ✅

```bash
$ go build -o bin/sky-server.exe ./cmd/server
# 构建成功，无错误
```

## 优势

### 1. 语义清晰

从名称就能清楚知道：
- 哪个表触发
- 什么时机执行
- 什么操作

示例：`sys_table_after_create` 一眼就知道是"sys_table创建后"

### 2. 易于管理

统一的命名规则：
- 便于查找相关插件
- 便于按表或操作分组
- 便于代码审查

### 3. 可扩展性强

清晰的规则便于添加新插件：
- 开发者知道如何命名
- 避免命名冲突
- 保持代码一致性

### 4. 自文档化

命名本身就是文档：
```go
// 无需过多注释，名称已说明一切
plugin.NewSysTableAfterCreatePlugin()
plugin.NewSysUserBeforeDeletePlugin()
plugin.NewOrderAfterUpdatePlugin()
```

## 向后兼容性

本次重构是重命名操作，不影响：
- ✅ 插件系统架构
- ✅ 插件接口定义
- ✅ 插件执行逻辑
- ✅ 数据库结构

只影响：
- 代码中的类型名称
- 插件的字符串标识
- 文档和注释

## 未来规划

基于新的命名规范，可以轻松添加：

### 1. Before 插件

```go
// sys_table_before_delete.go
type SysTableBeforeDeletePlugin struct{}

func (p *SysTableBeforeDeletePlugin) Name() string {
    return "sys_table_before_delete"
}
```

### 2. Update 插件

```go
// product_after_update.go
type ProductAfterUpdatePlugin struct{}

func (p *ProductAfterUpdatePlugin) Name() string {
    return "product_after_update"
}
```

### 3. 业务表插件

```go
// order_after_create.go
type OrderAfterCreatePlugin struct{}

func (p *OrderAfterCreatePlugin) Name() string {
    return "order_after_create"
}
```

## 最佳实践

### 命名时应考虑

1. **表名**: 使用实际的数据库表名
2. **时机**: before 或 after，考虑插件的执行时机
3. **动作**: create、update、delete，要明确
4. **功能**: 插件名应反映其主要功能

### 示例

✅ **好的命名**:
```
sys_table_after_create     - 创建sys_table后生成标准字段
sys_user_after_create      - 创建用户后发送欢迎邮件
order_before_delete        - 删除订单前检查权限
product_after_update       - 更新产品后记录历史
```

❌ **不好的命名**:
```
table_plugin              - 不明确哪个表
create_standard_fields    - 不知道针对哪个表
user_plugin               - 不知道什么时机和动作
```

## 总结

通过这次重构，插件系统的命名更加规范、清晰和易于维护。新的命名规则 **表单名称_执行时机_动作** 使得：

1. **代码更易读** - 一目了然
2. **管理更方便** - 统一规范
3. **扩展更简单** - 清晰模板
4. **维护更轻松** - 自文档化

所有测试通过，构建成功，可以投入使用。
