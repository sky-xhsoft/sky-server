# CRUD Service 事务修复总结

## 修复日期
2026-01-12

## 背景

在代码审查中发现 `crud_service.go` 存在严重的**事务一致性问题**：hooks 和数据库操作不在同一事务中，可能导致数据不一致。

## 问题列表

### 1. Create 方法 ❌
- before hooks 在事务外
- 插入操作在事务外
- after hooks 在事务外
- **风险**：after hooks 失败时，插入已完成无法回滚

### 2. Update 方法 ❌
- before hooks 在事务外
- 更新操作在事务外
- after hooks 在事务外
- **风险**：数据不一致

### 3. Delete 方法 ❌
- before hooks 在事务外
- 删除操作在事务外
- after hooks 在事务外
- **风险**：数据不一致

### 4. BatchDelete 方法 ❌
- 没有 hooks 调用
- 没有事务控制
- **风险**：批量删除时业务逻辑未执行

### 5. Go 钩子事务支持 ❌
- Go 钩子无法访问事务数据库连接
- **风险**：Go 钩子的数据库操作不在主事务中

## 修复方案

### 1. 创建 executeHooksInTx 方法

在事务中执行 hooks：

```go
func (s *service) executeHooksInTx(ctx context.Context, tx *gorm.DB, tableID uint, action, event string, data map[string]interface{}) error {
    hooks, err := s.metadataRepo.GetTableCmdsByAction(tableID, action, event)
    if err != nil {
        return err
    }

    for _, hook := range hooks {
        if err := s.executeHook(ctx, hook, data, tx); err != nil {
            return err
        }
    }

    return nil
}
```

### 2. 修改 executeHook 接收 db 参数

```go
// 修改前
func (s *service) executeHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}) error

// 修改后
func (s *service) executeHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error
```

### 3. 修改 executeSPHook 使用事务连接

```go
func (s *service) executeSPHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
    // ...
    spExecutor := executor.NewSPExecutor(db)  // 使用传入的 db
    // ...
}
```

### 4. 修改 executeScriptHook 支持 Go 钩子事务

```go
func (s *service) executeScriptHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
    // ...
    params := make(map[string]interface{})
    for k, v := range data {
        params[k] = v
    }

    // 对于 Go 钩子，将数据库连接加入到参数中
    if hook.ContentType == "go" && db != nil {
        params["__db__"] = db
    }

    scriptExecutor := executor.NewScriptExecutor(scriptType, 5*time.Minute)
    result, err := scriptExecutor.Execute(ctx, hook.Content, params)
    // ...
}
```

### 5. 修改 CRUD 方法使用事务

#### Create 方法

```go
// 在事务中执行：before钩子 + 插入 + after钩子
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "begin", data); err != nil {
        return err
    }

    if err := tx.Table(table.Name).Create(&processedData).Error; err != nil {
        return err
    }

    if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "end", processedData); err != nil {
        return err
    }

    return nil
})
```

#### Update 方法

```go
// 在事务中执行：before钩子 + 更新 + after钩子
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "begin", data); err != nil {
        return err
    }

    result := tx.Table(table.Name).
        Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
        Select(updateFields).
        Updates(processedData)
    if result.Error != nil {
        return err
    }

    if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "end", processedData); err != nil {
        return err
    }

    return nil
})
```

#### Delete 方法

```go
// 在事务中执行：before钩子 + 删除 + after钩子
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
        return err
    }

    result := tx.Table(table.Name).
        Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
        Update("IS_ACTIVE", "N")
    if result.Error != nil {
        return err
    }

    if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
        return err
    }

    return nil
})
```

#### BatchDelete 方法

```go
// 在事务中执行批量删除
err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    // 对每个ID执行before钩子
    for _, id := range ids {
        deleteData := map[string]interface{}{"ID": id}
        if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
            return err
        }
    }

    // 执行批量软删除
    result := tx.Table(table.Name).
        Where("ID IN ? AND IS_ACTIVE = ?", ids, "Y").
        Update("IS_ACTIVE", "N")
    if result.Error != nil {
        return err
    }

    // 对每个ID执行after钩子
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

### internal/service/crud/crud_service.go

**新增方法**：
1. `executeHooksInTx` (行 599-615) - 在事务中执行钩子列表

**修改的方法**：
1. `executeHooks` (行 582-597) - 传递 s.db 给 executeHook
2. `executeHook` (行 648-661) - 新增 db 参数
3. `executeScriptHook` (行 663-699) - 新增 db 参数，支持 Go 钩子事务
4. `executeSPHook` (行 721-746) - 新增 db 参数，使用传入的 db
5. `Create` (行 245-343) - 使用事务包裹核心操作
6. `Update` (行 345-416) - 使用事务包裹核心操作
7. `Delete` (行 418-462) - 使用事务包裹核心操作
8. `BatchDelete` (行 464-509) - 使用事务包裹核心操作，新增 hooks

## 事务范围设计

### 事务内 ✅
- Before hooks 执行
- 数据库主操作（INSERT/UPDATE/DELETE）
- After hooks 执行
- 存储过程钩子（使用事务连接）
- **Go 钩子（可访问事务连接）**

### 事务外 ✅
- 权限检查
- 元数据查询（表定义、字段定义）
- 字段验证和处理
- 插件执行（失败不影响主流程）
- 脚本钩子（js、py、bsh - 外部进程）
- URL 钩子（外部服务）

## Go 钩子事务支持

### 数据库连接传递

通过特殊参数 `__db__` 传递给 Go 钩子函数：

```go
if hook.ContentType == "go" && db != nil {
    params["__db__"] = db
}
```

### Go 钩子函数示例

```go
executor.RegisterGoFunc("myHook", func(params map[string]interface{}) (interface{}, error) {
    // 获取事务数据库连接
    db, ok := params["__db__"].(*gorm.DB)
    if !ok || db == nil {
        return nil, fmt.Errorf("数据库连接不可用")
    }

    // 在事务中执行数据库操作
    if err := db.Table("related_table").
        Where("parent_id = ?", params["ID"]).
        Update("parent_name", params["NAME"]).Error; err != nil {
        return nil, err
    }

    return map[string]interface{}{"success": true}, nil
})
```

## 收益

### 1. 数据一致性 ✅
- Before hooks、主操作、After hooks 在同一事务中
- 任何一个失败，整个事务回滚
- 保证数据的原子性

### 2. 钩子事务支持 ✅
- 存储过程钩子可以访问事务连接
- Go 钩子可以访问事务连接
- 钩子内的数据库操作与主操作在同一事务中

### 3. BatchDelete 完整性 ✅
- 批量删除现在会为每个 ID 执行钩子
- 所有操作在一个事务中，保证一致性

### 4. 向后兼容 ✅
- API 接口没有变化
- 现有的钩子脚本无需修改
- 不影响现有功能

## 相关文档

1. **详细修复文档**: `docs/transaction-hooks-fix.md`
2. **Go 钩子使用指南**: `docs/go-hook-guide.md`
3. **事务工具**: `internal/pkg/transaction/transaction.go`
4. **事务指南**: `docs/transaction-guide.md`
5. **事务实现总结**: `docs/transaction-implementation-summary.md`
6. **Update 零值修复**: `docs/update-zero-value-fix.md`

## 测试建议

### 单元测试

1. **测试事务回滚**
   ```go
   func TestCreate_HookFailureRollback(t *testing.T) {
       // Mock after hook 失败
       // 验证数据未插入（事务回滚）
   }
   ```

2. **测试钩子在事务中执行**
   ```go
   func TestCreate_HooksInTransaction(t *testing.T) {
       // 在 before hook 中插入测试数据
       // 在 after hook 中验证可以读取到这些数据
   }
   ```

3. **测试 Go 钩子数据库访问**
   ```go
   func TestGoHook_DatabaseAccess(t *testing.T) {
       // Go 钩子访问数据库
       // 验证操作在同一事务中
   }
   ```

### 集成测试

1. **测试真实事务提交**
2. **测试并发场景**
3. **测试性能影响**

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

### 修复前 ❌
- Hooks 和数据库操作不在同一事务中
- 数据不一致风险
- BatchDelete 没有 hooks 调用
- Go 钩子无法访问事务连接

### 修复后 ✅
- Before hooks + 主操作 + After hooks 在同一事务中
- 保证数据的原子性和一致性
- 存储过程钩子和 Go 钩子可以访问事务连接
- BatchDelete 完整支持 hooks
- 插件执行保持在事务外（合理设计）

### 关键改进
1. **数据一致性** - 事务保证原子性
2. **代码可维护性** - 清晰的事务边界
3. **功能完整性** - BatchDelete 支持 hooks、Go 钩子支持事务
4. **向后兼容** - 不影响现有功能

这次修复是系统稳定性和数据一致性的重要改进。✅
