# 插件命名规范

## 命名规则

插件遵循统一的命名规则：**表单名称_执行时机_动作**

### 格式说明

```
<table_name>_<timing>_<action>
```

- **表单名称 (table_name)**: 插件作用的目标表，使用小写和下划线分隔
- **执行时机 (timing)**: before（前） 或 after（后）
- **动作 (action)**: create、update、delete 等操作

### 示例

| 插件名称 | 说明 |
|---------|------|
| `sys_table_after_create` | sys_table表创建后执行 |
| `sys_table_before_delete` | sys_table删除前执行 |
| `sys_user_after_update` | sys_user更新后执行 |
| `sys_column_before_create` | sys_column创建前执行 |

## 已实现的插件

### sys_table_after_create

**文件**: `internal/pkg/plugin/sys_table_after_create.go`

**功能**: 在 sys_table 创建后，自动生成标准字段到 sys_column 表

**命名解析**:
- 表单名称：`sys_table`
- 执行时机：`after`（后）
- 动作：`create`

**完整名称**: `sys_table_after_create`

**生成的标准字段**：
1. ID - 主键
2. SYS_COMPANY_ID - 公司隔离
3. CREATE_BY - 创建人
4. CREATE_TIME - 创建时间
5. UPDATE_BY - 修改人
6. UPDATE_TIME - 修改时间
7. IS_ACTIVE - 软删除标志

## 代码结构

### 文件命名

文件名应与插件名称一致：

```
internal/pkg/plugin/{table_name}_{timing}_{action}.go
```

**示例**:
- `sys_table_after_create.go`
- `sys_table_before_delete.go`

### 类型命名

结构体使用 PascalCase：

```go
// 将插件名转换为 PascalCase
// sys_table_after_create -> SysTableAfterCreatePlugin

type SysTableAfterCreatePlugin struct{}
```

### 方法命名

```go
// 构造函数
func NewSysTableAfterCreatePlugin() *SysTableAfterCreatePlugin

// Name() 方法返回插件的唯一标识（使用下划线格式）
func (p *SysTableAfterCreatePlugin) Name() string {
    return "sys_table_after_create"
}
```

## 命名规范详解

### 1. 表单名称 (Table Name)

使用实际的数据库表名，保持小写和下划线格式：
- `sys_table`
- `sys_column`
- `sys_user`
- `customer`
- `product_order`

### 2. 执行时机 (Timing)

| 时机 | 说明 | 使用场景 |
|------|------|---------|
| `before` | 操作前执行 | 数据验证、权限检查、数据预处理 |
| `after` | 操作后执行 | 日志记录、关联数据创建、通知发送 |

### 3. 动作 (Action)

| 动作 | 说明 | 对应的 CRUD 操作 |
|------|------|-----------------|
| `create` | 创建操作 | INSERT |
| `update` | 更新操作 | UPDATE |
| `delete` | 删除操作 | DELETE/软删除 |
| `query` | 查询操作 | SELECT (可选) |

## 实际应用示例

### 示例 1: 创建后发送通知

```go
// 文件: sys_user_after_create.go
type SysUserAfterCreatePlugin struct{}

func (p *SysUserAfterCreatePlugin) Name() string {
    return "sys_user_after_create"
}

func (p *SysUserAfterCreatePlugin) Execute(ctx context.Context, db *gorm.DB, data PluginData) error {
    if data.Action != "create" {
        return nil
    }

    // 发送欢迎邮件
    email := data.Data["EMAIL"]
    // sendWelcomeEmail(email)

    return nil
}
```

### 示例 2: 删除前检查关联

```go
// 文件: sys_table_before_delete.go
type SysTableBeforeDeletePlugin struct{}

func (p *SysTableBeforeDeletePlugin) Name() string {
    return "sys_table_before_delete"
}

func (p *SysTableBeforeDeletePlugin) Execute(ctx context.Context, db *gorm.DB, data PluginData) error {
    if data.Action != "delete" {
        return nil
    }

    // 检查是否有关联的字段
    var count int64
    db.Table("sys_column").Where("SYS_TABLE_ID = ?", data.RecordID).Count(&count)

    if count > 0 {
        return fmt.Errorf("无法删除表，存在 %d 个关联字段", count)
    }

    return nil
}
```

### 示例 3: 更新后记录历史

```go
// 文件: product_after_update.go
type ProductAfterUpdatePlugin struct{}

func (p *ProductAfterUpdatePlugin) Name() string {
    return "product_after_update"
}

func (p *ProductAfterUpdatePlugin) Execute(ctx context.Context, db *gorm.DB, data PluginData) error {
    if data.Action != "update" {
        return nil
    }

    // 记录价格变更历史
    history := map[string]interface{}{
        "PRODUCT_ID":   data.RecordID,
        "OLD_PRICE":    data.Data["OLD_PRICE"],
        "NEW_PRICE":    data.Data["PRICE"],
        "CHANGE_TIME":  time.Now(),
        "CHANGE_BY":    data.UserID,
    }

    return db.Table("product_price_history").Create(&history).Error
}
```

## 注册插件

在 `cmd/server/main.go` 中注册插件：

```go
pluginManager := plugin.NewManager(db)

// 按照命名规则注册
pluginManager.Register("sys_table", plugin.NewSysTableAfterCreatePlugin())
pluginManager.Register("sys_user", plugin.NewSysUserAfterCreatePlugin())
pluginManager.Register("product", plugin.NewProductAfterUpdatePlugin())
```

## 最佳实践

### 1. 单一职责

每个插件只做一件事：

✅ **好的做法**：
- `sys_table_after_create` - 只生成标准字段
- `sys_user_after_create` - 只发送欢迎邮件

❌ **不好的做法**：
- `sys_table_after_create` - 生成字段 + 发送通知 + 创建目录

### 2. 清晰命名

命名应该清楚表达插件的作用：

✅ **好的命名**：
- `sys_table_after_create` - 表明在sys_table创建后执行
- `order_before_delete` - 表明在order删除前执行

❌ **不好的命名**：
- `table_plugin` - 不清楚什么时机
- `create_columns` - 不清楚针对哪个表

### 3. 注释完整

```go
// SysTableAfterCreatePlugin sys_table表创建后自动生成标准字段的插件
// 插件命名规则：表单名称_执行时机_动作
type SysTableAfterCreatePlugin struct{}
```

### 4. 操作验证

始终验证操作类型：

```go
func (p *Plugin) Execute(ctx context.Context, db *gorm.DB, data PluginData) error {
    // 确保只在目标操作时执行
    if data.Action != "create" {
        return nil
    }

    // 执行逻辑
    return nil
}
```

## 总结

遵循 **表单名称_执行时机_动作** 的命名规范，可以让插件系统：

1. **易于理解** - 从名称就知道插件的作用
2. **易于管理** - 统一的命名规则便于维护
3. **易于扩展** - 清晰的规则便于添加新插件
4. **避免冲突** - 规范的命名减少重复
