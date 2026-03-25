# CRUD Service 事务和钩子一致性修复

## 实施日期
2026-01-12

## 问题描述

在代码审查中发现，`crud_service.go` 中的 Create、Update、Delete 和 BatchDelete 方法存在严重的**事务一致性问题**：

### 问题 1：钩子和数据库操作不在同一事务中

**Create 方法（修复前）**：
```go
// before hooks - 在事务外执行 ❌
s.executeHooks(ctx, table.ID, "A", "begin", data)

// 插入操作 - 在事务外执行 ❌
s.db.Table(table.Name).Create(&processedData)

// after hooks - 在事务外执行 ❌
s.executeHooks(ctx, table.ID, "A", "end", processedData)
```

**风险**：
- 如果 after hooks 失败，插入操作已完成，无法回滚
- 如果插入操作失败，before hooks 已经执行，可能造成副作用
- 多个操作之间没有原子性保证

**Update 和 Delete 方法存在同样的问题**。

### 问题 2：BatchDelete 没有钩子调用

**BatchDelete 方法（修复前）**：
```go
// 直接执行批量删除，没有任何 hooks ❌
s.db.Table(table.Name).
    Where("ID IN ? AND IS_ACTIVE = ?", ids, "Y").
    Update("IS_ACTIVE", "N")
```

**风险**：
- 批量删除时，相关的业务逻辑（钩子）没有执行
- 可能导致数据不一致

## 解决方案

### 核心原则

✅ **before hooks + 主操作 + after hooks** 必须在同一个事务中执行

✅ **插件执行保持在事务外**（插件失败不应影响主流程）

### 实现方法

#### 1. 创建 executeHooksInTx 方法

在 `crud_service.go` 中添加了新方法，支持在事务中执行钩子：

```go
// executeHooksInTx 在事务中执行钩子
func (s *service) executeHooksInTx(ctx context.Context, tx *gorm.DB, tableID uint, action, event string, data map[string]interface{}) error {
    // 获取钩子列表
    hooks, err := s.metadataRepo.GetTableCmdsByAction(tableID, action, event)
    if err != nil {
        return err
    }

    // 按顺序执行钩子（在事务中）
    for _, hook := range hooks {
        if err := s.executeHook(ctx, hook, data, tx); err != nil {
            return err
        }
    }

    return nil
}
```

#### 2. 修改 executeHook 接收 db 参数

```go
// 修改前
func (s *service) executeHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}) error

// 修改后
func (s *service) executeHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error
```

这样存储过程钩子可以使用传入的 db（可能是事务连接）：

```go
func (s *service) executeSPHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
    // ...
    spExecutor := executor.NewSPExecutor(db)  // 使用传入的 db
    // ...
}
```

#### 3. 修改 Create 方法

**修复后**：
```go
// 获取字段定义（在事务外，避免长时间持有锁）
columns, err := s.metadataService.GetColumns(table.ID)
// 验证和处理字段（在事务外）
processedData, err := s.processFieldsForCreate(columns, data, userID)

// 在事务中执行：before钩子 + 插入 + after钩子
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    // 执行before钩子（在事务中）✅
    if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "begin", data); err != nil {
        return err
    }

    // 执行插入（在事务中）✅
    if err := tx.Table(table.Name).Create(&processedData).Error; err != nil {
        return err
    }

    // 执行after钩子（在事务中）✅
    if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "end", processedData); err != nil {
        return err
    }

    return nil
})

// 插件在事务外执行 ✅
if s.pluginManager != nil && recordID > 0 {
    s.pluginManager.ExecutePlugins(ctx, pluginData)
}
```

#### 4. 修改 Update 方法

**修复后**：
```go
// 获取字段定义（在事务外）
columns, err := s.metadataService.GetColumns(table.ID)
// 验证和处理字段（在事务外）
processedData, err := s.processFieldsForUpdate(columns, data, userID)

// 在事务中执行：before钩子 + 更新 + after钩子
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    // 执行before钩子（在事务中）✅
    if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "begin", data); err != nil {
        return err
    }

    // 执行更新（在事务中，支持零值更新）✅
    result := tx.Table(table.Name).
        Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
        Select(updateFields).
        Updates(processedData)
    if result.Error != nil {
        return err
    }

    // 执行after钩子（在事务中）✅
    if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "end", processedData); err != nil {
        return err
    }

    return nil
})
```

#### 5. 修改 Delete 方法

**修复后**：
```go
// 在事务中执行：before钩子 + 删除 + after钩子
deleteData := map[string]interface{}{"ID": id}
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    // 执行before钩子（在事务中）✅
    if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
        return err
    }

    // 执行软删除（在事务中）✅
    result := tx.Table(table.Name).
        Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
        Update("IS_ACTIVE", "N")
    if result.Error != nil {
        return err
    }

    // 执行after钩子（在事务中）✅
    if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
        return err
    }

    return nil
})
```

#### 6. 修改 BatchDelete 方法

**修复后（新增钩子调用）**：
```go
// 在事务中执行批量删除
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    // 对每个ID执行before钩子（在事务中）✅
    for _, id := range ids {
        deleteData := map[string]interface{}{"ID": id}
        if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
            return err
        }
    }

    // 执行批量软删除（在事务中）✅
    result := tx.Table(table.Name).
        Where("ID IN ? AND IS_ACTIVE = ?", ids, "Y").
        Update("IS_ACTIVE", "N")
    if result.Error != nil {
        return err
    }

    // 对每个ID执行after钩子（在事务中）✅
    for _, id := range ids {
        deleteData := map[string]interface{}{"ID": id}
        if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
            return err
        }
    }

    return nil
})
```

## 修改文件

### 主文件：`internal/service/crud/crud_service.go`

#### 新增方法：
1. **executeHooksInTx** (行 599-615)
   - 在事务中执行钩子列表

#### 修改的方法：
1. **executeHooks** (行 582-597)
   - 传递 s.db 给 executeHook

2. **executeHook** (行 617-630)
   - 新增 db 参数
   - 将 db 传递给 executeSPHook

3. **executeSPHook** (行 688-713)
   - 新增 db 参数
   - 使用传入的 db 创建 SPExecutor

4. **Create** (行 245-343)
   - 使用 transaction.RunInTransaction 包裹核心操作
   - 在事务中调用 executeHooksInTx

5. **Update** (行 345-416)
   - 使用 transaction.RunInTransaction 包裹核心操作
   - 在事务中调用 executeHooksInTx
   - 保持零值更新支持

6. **Delete** (行 418-462)
   - 使用 transaction.RunInTransaction 包裹核心操作
   - 在事务中调用 executeHooksInTx

7. **BatchDelete** (行 464-509)
   - 使用 transaction.RunInTransaction 包裹核心操作
   - 新增：为每个 ID 调用 before/after 钩子

## 事务范围设计

### 事务内的操作
✅ **必须在事务中**：
- Before hooks 执行
- 数据库主操作（INSERT/UPDATE/DELETE）
- After hooks 执行
- 存储过程钩子（使用事务连接）
- **Go 钩子（可访问事务连接）**

### 事务外的操作
✅ **应该在事务外**：
- 权限检查
- 元数据查询（表定义、字段定义）
- 字段验证和处理
- 插件执行（失败不影响主流程）
- 脚本钩子（js、py、bsh - 外部进程）
- URL 钩子（外部服务）

## Go 钩子的事务支持

### 数据库连接传递

对于 Go 类型的钩子，数据库连接通过特殊参数 `__db__` 传递给注册的 Go 函数：

**代码实现**：
```go
// executeScriptHook 执行脚本钩子
func (s *service) executeScriptHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
    // ...

    // 对于 Go 钩子，将数据库连接加入到参数中
    if hook.ContentType == "go" && db != nil {
        params["__db__"] = db
    }

    scriptExecutor := executor.NewScriptExecutor(scriptType, 5*time.Minute)
    result, err := scriptExecutor.Execute(ctx, hook.Content, params)
    // ...
}
```

### Go 钩子函数示例

注册的 Go 钩子函数可以从 params 中获取数据库连接：

```go
import (
    "github.com/sky-xhsoft/sky-server/internal/pkg/executor"
    "gorm.io/gorm"
)

// 注册 Go 钩子函数
func init() {
    executor.RegisterGoFunc("validateAndUpdateRelated", func(params map[string]interface{}) (interface{}, error) {
        // 从 params 获取数据库连接（事务连接）
        db, ok := params["__db__"].(*gorm.DB)
        if !ok || db == nil {
            return nil, fmt.Errorf("数据库连接不可用")
        }

        // 获取业务数据
        recordID := params["ID"].(uint)
        name := params["NAME"].(string)

        // 在事务中执行数据库操作
        // 这些操作与主操作在同一事务中，保证一致性
        if err := db.Table("related_table").
            Where("parent_id = ?", recordID).
            Update("parent_name", name).Error; err != nil {
            return nil, err
        }

        // 验证逻辑
        var count int64
        if err := db.Table("related_table").
            Where("parent_id = ?", recordID).
            Count(&count).Error; err != nil {
            return nil, err
        }

        return map[string]interface{}{
            "validated": true,
            "related_count": count,
        }, nil
    })
}
```

### 使用场景

Go 钩子适合以下场景：

1. **复杂的业务逻辑验证**
   - 需要查询多个表
   - 复杂的数据验证规则
   - 与主操作在同一事务中

2. **关联数据更新**
   - 级联更新相关记录
   - 更新缓存表
   - 维护数据一致性

3. **业务规则执行**
   - 库存检查和扣减
   - 账户余额验证和更新
   - 状态机转换

### 优势

✅ **性能**：Go 函数在进程内执行，无需启动外部进程
✅ **类型安全**：可以使用 Go 的类型系统
✅ **事务一致性**：数据库操作与主操作在同一事务中
✅ **调试方便**：可以使用 Go 的调试工具

## 收益

### 1. 数据一致性 ✅
- Before hooks、主操作、After hooks 在同一事务中
- 任何一个失败，整个事务回滚
- 保证数据的原子性

### 2. 钩子事务支持 ✅
- 存储过程钩子现在可以访问事务连接
- 钩子内的数据库操作与主操作在同一事务中

### 3. BatchDelete 完整性 ✅
- 批量删除现在会为每个 ID 执行钩子
- 所有操作在一个事务中，保证一致性

### 4. 向后兼容 ✅
- API 接口没有变化
- 现有的钩子脚本无需修改
- 不影响现有功能

## 性能考虑

### 事务范围优化
- ✅ 元数据查询在事务外（避免持有锁）
- ✅ 字段验证在事务外（减少事务时间）
- ✅ 权限检查在事务外（快速失败）

### BatchDelete 性能
- ⚠️ 为每个 ID 调用 hooks 可能较慢
- ✅ 但保证了数据一致性
- 💡 如果性能是问题，可以考虑批量 hooks API

## 测试建议

### 单元测试

#### 测试事务回滚
```go
func TestCreate_HookFailureRollback(t *testing.T) {
    // Mock after hook 失败
    mockHook := createFailingHook()

    // 尝试创建
    _, err := service.Create(ctx, "users", data, userID)
    assert.Error(t, err)

    // 验证数据未插入（事务回滚）
    var count int64
    db.Table("users").Count(&count)
    assert.Equal(t, 0, count)
}
```

#### 测试钩子在事务中执行
```go
func TestCreate_HooksInTransaction(t *testing.T) {
    // 创建一个在 before hook 中插入测试数据的钩子
    // 在 after hook 中验证可以读取到这些数据（说明在同一事务中）
}
```

#### 测试 BatchDelete 钩子调用
```go
func TestBatchDelete_HooksExecuted(t *testing.T) {
    // 创建多条记录
    ids := []uint{1, 2, 3}

    // 批量删除
    err := service.BatchDelete(ctx, "users", ids, userID)
    assert.NoError(t, err)

    // 验证每个 ID 的 hooks 都被调用
    assert.Equal(t, 3, beforeHookCallCount)
    assert.Equal(t, 3, afterHookCallCount)
}
```

### 集成测试

#### 测试真实事务行为
```go
func TestCreate_TransactionCommit(t *testing.T) {
    // 使用真实数据库
    db := setupRealDB(t)
    service := NewService(db, ...)

    // 创建记录
    result, err := service.Create(ctx, "users", data, userID)
    assert.NoError(t, err)

    // 在新连接中验证数据已提交
    var user User
    newDB := connectToDB()
    err = newDB.Where("id = ?", result["ID"]).First(&user).Error
    assert.NoError(t, err)
}
```

## 相关文档

- **事务工具**: `internal/pkg/transaction/transaction.go`
- **事务指南**: `docs/transaction-guide.md`
- **事务实现总结**: `docs/transaction-implementation-summary.md`
- **事务示例**: `examples/transaction_crud_example.go`
- **Update 零值修复**: `docs/update-zero-value-fix.md`

## 下一步

### 短期
1. ✅ 修复代码（已完成）
2. ⏳ 编写单元测试
3. ⏳ 编写集成测试
4. ⏳ Code Review

### 中期
1. ⏳ 性能测试（特别是 BatchDelete）
2. ⏳ 监控事务时间
3. ⏳ 优化慢事务

### 长期
1. ⏳ 考虑批量 hooks API（性能优化）
2. ⏳ 添加事务重试机制（死锁处理）
3. ⏳ 事务隔离级别优化

## 总结

通过本次修复，我们解决了 CRUD Service 中严重的事务一致性问题：

### 修复前 ❌
- Hooks 和数据库操作不在同一事务中
- 数据不一致风险
- BatchDelete 没有 hooks 调用

### 修复后 ✅
- Before hooks + 主操作 + After hooks 在同一事务中
- 保证数据的原子性和一致性
- 存储过程钩子可以访问事务连接
- BatchDelete 完整支持 hooks
- 插件执行保持在事务外（合理设计）

### 关键改进
1. **数据一致性** - 事务保证原子性
2. **代码可维护性** - 清晰的事务边界
3. **功能完整性** - BatchDelete 支持 hooks
4. **向后兼容** - 不影响现有功能

这次修复是系统稳定性和数据一致性的重要改进。
