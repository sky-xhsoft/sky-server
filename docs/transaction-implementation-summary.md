# Service 层事务控制实现总结

## 实现日期
2026-01-12

## 背景

在代码审查中发现，Service 层的很多方法涉及多个数据库操作，但没有使用事务控制，可能导致数据不一致问题。为解决这个问题，我们实现了统一的事务管理工具和使用规范。

## 已完成的工作

### 1. 问题分析 ✅

**文档**: `docs/transaction-analysis.md`

分析了需要事务控制的场景：
- ✅ CRUD Service: Create/Update/Delete 方法
- ✅ SSO Service: Login 方法
- ✅ Groups Service: AssignPermissions 方法
- ✅ Workflow Service: 工作流操作
- ❌ Audit Service: 不需要（独立记录）
- ❌ Metadata/Dict/Sequence: 不需要（只读或单一操作）

### 2. 事务工具包 ✅

**文件**: `internal/pkg/transaction/transaction.go`

实现了以下工具：

#### RunInTransaction（推荐）
```go
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    // 在事务中执行操作
    return nil // 自动提交
    // return err // 自动回滚
})
```

#### Executor（面向对象）
```go
executor := transaction.NewExecutor(db)
err := executor.Execute(ctx, func(tx *gorm.DB) error {
    return nil
})
```

#### Context 事务传递
```go
err := transaction.ExecuteInTransaction(ctx, db, func(tx *gorm.DB) error {
    ctx := transaction.WithTx(ctx, tx)
    // 传递 context 到其他函数
    return nil
})
```

**特点**：
- 自动提交/回滚
- panic 安全
- 支持嵌套调用
- Context 传递

### 3. 单元测试 ✅

**文件**: `internal/pkg/transaction/transaction_test.go`

测试覆盖：
- ✅ Context 存取测试
- ✅ 函数类型测试
- ✅ 基础逻辑验证

注：完整的数据库集成测试需要在实际环境中运行。

### 4. 使用规范文档 ✅

**文件**: `docs/transaction-guide.md`

包含：
- 何时使用事务
- 如何使用事务工具
- 最佳实践
- 性能考虑
- 常见问题
- 迁移指南

### 5. 代码示例 ✅

#### CRUD Service 示例
**文件**: `examples/transaction_crud_example.go`

示例方法：
- `CreateWithTransaction` - 事务中创建记录
- `UpdateWithTransaction` - 事务中更新记录
- `DeleteWithTransaction` - 事务中删除记录
- `BatchDeleteWithTransaction` - 事务中批量删除

**要点**：
- before钩子、插入、after钩子在同一事务中
- 插件执行在事务外（失败不影响主流程）
- 权限检查、元数据查询可在事务外

#### SSO Service 示例
**文件**: `examples/transaction_sso_example.go`

示例方法：
- `LoginWithTransaction` - 事务中登录
- `LogoutWithTransaction` - 事务中登出
- `LogoutAllWithTransaction` - 事务中全部登出
- `RefreshTokenWithTransaction` - 事务中刷新Token
- `CleanExpiredSessionsWithTransaction` - 事务中清理过期会话

**要点**：
- 查询用户和创建/更新会话在同一事务中
- 防止并发登录导致的会话重复
- Token生成和会话保存原子性

#### Groups Service 示例
**文件**: `examples/transaction_groups_example.go`

示例方法：
- `AssignPermissionsWithTransaction` - 事务中分配权限
- `RemovePermissionsWithTransaction` - 事务中移除权限
- `AssignGroupsToUserWithTransaction` - 事务中分配用户权限组
- `DeleteGroupWithTransaction` - 事务中删除权限组
- `DeleteDirectoryWithTransaction` - 事务中删除目录
- `CopyGroupPermissionsWithTransaction` - 事务中复制权限

**要点**：
- 删除旧权限和创建新权限在同一事务中
- 删除前检查关联
- 批量操作保证原子性

## 架构设计

### 事务层次

```
+------------------+
|   Handler层      | HTTP请求处理
+------------------+
         ↓
+------------------+
|   Service层      | 业务逻辑 ← 在这里添加事务
+------------------+
         ↓
+------------------+
|  Repository层    | 数据访问
+------------------+
         ↓
+------------------+
|   Database       | MySQL
+------------------+
```

### 事务范围设计

```go
// ✅ 好的设计：事务包含必须的操作
RunInTransaction(db, func(tx *gorm.DB) error {
    // 1. before钩子（必须在事务中）
    executeHooks(tx, "before", data)

    // 2. 主操作（必须在事务中）
    tx.Create(data)

    // 3. after钩子（必须在事务中）
    executeHooks(tx, "after", data)

    return nil
})

// 插件在事务外执行（失败不影响主流程）
pluginManager.ExecutePlugins(ctx, data)
```

## 使用指南

### 快速开始

1. **引入包**
```go
import "github.com/sky-xhsoft/sky-server/internal/pkg/transaction"
```

2. **使用事务**
```go
err := transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
    // 多个操作
    if err := tx.Create(&obj1).Error; err != nil {
        return err
    }

    if err := tx.Update(&obj2).Error; err != nil {
        return err
    }

    return nil
})
```

3. **处理错误**
```go
if err != nil {
    log.Error("Transaction failed", zap.Error(err))
    return err
}
```

### 判断是否需要事务

使用以下决策树：

```
是否涉及多个数据库写操作？
├─ 是 → 必须使用事务 ✅
└─ 否 →
    ├─ 是否需要原子性？
    │   ├─ 是 → 使用事务 ✅
    │   └─ 否 → 不需要事务 ❌
    └─ 只读操作？ → 不需要事务 ❌
```

## 迁移计划

### 优先级

| 优先级 | Service | 方法 | 影响 | 状态 |
|--------|---------|------|------|------|
| 🔴 高 | CRUD | Create/Update/Delete | 数据一致性核心 | 📝 示例已完成 |
| 🔴 高 | SSO | Login | 会话状态不一致 | 📝 示例已完成 |
| 🟡 中 | Groups | AssignPermissions | 权限可能不完整 | 📝 示例已完成 |
| 🟡 中 | Workflow | 所有写操作 | 工作流状态混乱 | ⏳ 待实现 |
| 🟢 低 | Action | 取决于实现 | 取决于动作类型 | ⏳ 待分析 |

### 迁移步骤

1. **准备阶段** ✅
   - [x] 创建事务工具包
   - [x] 编写使用文档
   - [x] 创建代码示例

2. **实施阶段** ⏳
   - [ ] 修改 CRUD Service
   - [ ] 修改 SSO Service
   - [ ] 修改 Groups Service
   - [ ] 修改 Workflow Service

3. **测试阶段** ⏳
   - [ ] 单元测试
   - [ ] 集成测试
   - [ ] 压力测试

4. **上线阶段** ⏳
   - [ ] 灰度发布
   - [ ] 监控观察
   - [ ] 性能优化

## 技术要点

### 1. GORM 事务特性

```go
// GORM 自动处理
db.Transaction(func(tx *gorm.DB) error {
    // 返回 error 自动回滚
    // 返回 nil 自动提交
    // panic 自动回滚
    return nil
})
```

### 2. 事务隔离级别

默认使用 MySQL 的 `REPEATABLE READ`：
- 防止脏读
- 防止不可重复读
- 允许幻读（影响较小）

如需修改：
```go
db.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED")
```

### 3. 死锁避免

**原则**：
1. 按固定顺序访问表
2. 缩短事务时间
3. 合理使用索引
4. 避免长事务

**示例**：
```go
// ✅ 好：按固定顺序
tx.Table("orders").Update(...)      // 先 orders
tx.Table("order_items").Update(...) // 后 order_items

// ❌ 差：不同函数可能不同顺序，导致死锁
```

### 4. 性能优化

**原则**：
1. 事务范围最小化
2. 避免在事务中执行慢操作
3. 使用批量操作
4. 合理使用索引

**示例**：
```go
// ✅ 好：快速事务
RunInTransaction(db, func(tx *gorm.DB) error {
    return tx.Updates(data).Error
})

// ❌ 差：慢操作在事务中
RunInTransaction(db, func(tx *gorm.DB) error {
    time.Sleep(10 * time.Second) // 持有锁太久
    return tx.Updates(data).Error
})
```

## 监控和日志

### 日志记录

建议在事务中记录关键信息：

```go
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    log.Info("Transaction started",
        zap.String("operation", "create_order"),
        zap.Uint("userId", userID))

    // 操作...

    log.Info("Transaction completed successfully")
    return nil
})

if err != nil {
    log.Error("Transaction failed",
        zap.Error(err),
        zap.String("operation", "create_order"))
}
```

### 性能监控

可以添加事务耗时监控：

```go
start := time.Now()
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    // 操作...
    return nil
})
duration := time.Since(start)

if duration > time.Second {
    log.Warn("Slow transaction",
        zap.Duration("duration", duration))
}
```

## 常见问题

### Q1: 什么时候必须使用事务？

**A**: 当一个操作涉及多个数据库写操作，且它们必须全部成功或全部失败时。

示例：
- 转账（扣款+加款）
- 创建订单+订单项
- 删除用户+会话+权限

### Q2: 查询需要事务吗？

**A**: 通常不需要，除非：
- 需要 SELECT FOR UPDATE 加锁
- 需要保证多次查询的一致性视图

### Q3: 事务会影响性能吗？

**A**: 是的，但影响有限。优化方法：
- 缩短事务时间
- 只在必要时使用事务
- 避免在事务中执行慢操作

### Q4: 插件应该在事务内还是外？

**A**: 取决于插件的作用：
- **事务内**：插件失败应该回滚主操作
- **事务外**：插件是辅助功能（如发送通知）

### Q5: 如何测试事务回滚？

**A**:
```go
func TestTransactionRollback(t *testing.T) {
    // 模拟错误
    err := service.Create(data)
    assert.Error(t, err)

    // 验证数据未写入
    var count int64
    db.Model(&Model{}).Count(&count)
    assert.Equal(t, 0, count)
}
```

## 下一步工作

### 短期（1-2周）

1. ✅ 完成示例代码
2. ⏳ 修改实际的 Service 代码
3. ⏳ 添加集成测试
4. ⏳ Code Review

### 中期（1个月）

1. ⏳ 性能测试和优化
2. ⏳ 添加监控指标
3. ⏳ 编写迁移指南
4. ⏳ 团队培训

### 长期（持续）

1. ⏳ 收集反馈和改进
2. ⏳ 扩展到其他 Service
3. ⏳ 性能持续优化
4. ⏳ 最佳实践总结

## 参考资料

### 内部文档

- `docs/transaction-analysis.md` - 问题分析
- `docs/transaction-guide.md` - 使用指南
- `examples/transaction_*_example.go` - 代码示例

### 外部资料

- [GORM 事务文档](https://gorm.io/docs/transactions.html)
- [MySQL 事务隔离级别](https://dev.mysql.com/doc/refman/8.0/en/innodb-transaction-isolation-levels.html)
- [分布式事务最佳实践](https://microservices.io/patterns/data/saga.html)

## 总结

通过实现统一的事务管理工具和规范，我们解决了 Service 层缺乏事务控制的问题：

### 成果

1. **工具包** - 提供了简单易用的事务工具
2. **文档** - 详细的使用指南和最佳实践
3. **示例** - 三个核心 Service 的完整示例
4. **测试** - 基础的单元测试框架

### 优势

1. **数据一致性** - 保证多操作的原子性
2. **易于使用** - 简单的 API，自动提交/回滚
3. **安全可靠** - panic 安全，错误处理完善
4. **性能优化** - 合理的事务范围控制

### 下一步

现在工具和规范已经就绪，可以开始：
1. 在实际 Service 代码中应用
2. 编写完整的集成测试
3. 性能测试和优化
4. 团队培训和推广

**记住**：事务是确保数据一致性的关键，但也要注意性能影响。在使用时要权衡数据一致性需求和性能要求。
