# 事务使用指南

## 概述

为确保数据一致性，Service 层的多数据库操作必须使用事务控制。本文档说明何时使用事务以及如何正确使用。

## 事务工具包

位置：`internal/pkg/transaction`

### 主要方法

#### 1. RunInTransaction (推荐)

最简单的事务执行方式：

```go
import "github.com/sky-xhsoft/sky-server/internal/pkg/transaction"

err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&record1).Error; err != nil {
        return err // 自动回滚
    }

    if err := tx.Update(&record2).Error; err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})
```

#### 2. Executor (面向对象)

```go
executor := transaction.NewExecutor(db)

// 执行事务
err := executor.Execute(ctx, func(tx *gorm.DB) error {
    // 操作...
    return nil
})

// 带返回值的事务
result, err := executor.ExecuteWithResult(ctx, func(tx *gorm.DB) (interface{}, error) {
    model := &Model{}
    if err := tx.Create(model).Error; err != nil {
        return nil, err
    }
    return model, nil
})
```

#### 3. 使用 Context 传递事务

适用于需要在多个函数间共享事务：

```go
err := transaction.ExecuteInTransaction(ctx, db, func(tx *gorm.DB) error {
    // 将事务保存到context
    ctx := transaction.WithTx(ctx, tx)

    // 调用其他函数，传递context
    if err := someFunction(ctx, data); err != nil {
        return err
    }

    return nil
})

// 在被调用的函数中
func someFunction(ctx context.Context, data interface{}) error {
    // 从context获取事务
    tx := transaction.GetTxFromContext(ctx)
    if tx == nil {
        return errors.New("no transaction in context")
    }

    return tx.Create(data).Error
}
```

## 何时使用事务

### ✅ 需要事务的场景

#### 1. 多个写操作

```go
// ✅ 正确：使用事务
func (s *service) TransferMoney(from, to uint, amount float64) error {
    return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // 扣款
        if err := tx.Model(&Account{}).Where("id = ?", from).
            Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }

        // 加款
        if err := tx.Model(&Account{}).Where("id = ?", to).
            Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }

        return nil
    })
}
```

#### 2. 创建+关联数据

```go
// ✅ 正确：使用事务
func (s *service) CreateOrder(order *Order, items []*OrderItem) error {
    return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // 创建订单
        if err := tx.Create(order).Error; err != nil {
            return err
        }

        // 创建订单项
        for _, item := range items {
            item.OrderID = order.ID
            if err := tx.Create(item).Error; err != nil {
                return err
            }
        }

        return nil
    })
}
```

#### 3. 删除+清理关联

```go
// ✅ 正确：使用事务
func (s *service) DeleteUser(userID uint) error {
    return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // 删除用户会话
        if err := tx.Where("user_id = ?", userID).Delete(&UserSession{}).Error; err != nil {
            return err
        }

        // 删除用户权限
        if err := tx.Where("user_id = ?", userID).Delete(&UserPermission{}).Error; err != nil {
            return err
        }

        // 删除用户
        if err := tx.Delete(&User{}, userID).Error; err != nil {
            return err
        }

        return nil
    })
}
```

#### 4. 钩子+主操作

```go
// ✅ 正确：使用事务
func (s *service) CreateWithHooks(data map[string]interface{}) error {
    return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // before钩子
        if err := s.executeHooks(tx, "before", data); err != nil {
            return err
        }

        // 主操作
        if err := tx.Create(data).Error; err != nil {
            return err
        }

        // after钩子
        if err := s.executeHooks(tx, "after", data); err != nil {
            return err
        }

        return nil
    })
}
```

### ❌ 不需要事务的场景

#### 1. 单一查询

```go
// ❌ 不需要：单一查询
func (s *service) GetUser(id uint) (*User, error) {
    var user User
    if err := s.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

#### 2. 单一写操作

```go
// ❌ 不需要：单个更新
func (s *service) UpdateUserName(id uint, name string) error {
    return s.db.Model(&User{}).Where("id = ?", id).
        Update("name", name).Error
}
```

#### 3. 批量查询

```go
// ❌ 不需要：只读操作
func (s *service) ListUsers(page, pageSize int) ([]*User, error) {
    var users []*User
    offset := (page - 1) * pageSize
    if err := s.db.Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
        return nil, err
    }
    return users, nil
}
```

## 实际应用示例

### CRUD Service

#### 修改前（无事务）

```go
func (s *service) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) error {
    // before钩子
    s.executeHooks(ctx, table.ID, "A", "begin", data)

    // 插入数据
    s.db.Table(table.Name).Create(&processedData)

    // after钩子
    s.executeHooks(ctx, table.ID, "A", "end", processedData)

    // 执行插件
    s.pluginManager.ExecutePlugins(ctx, pluginData)

    return nil
}
```

**问题**：如果 after钩子失败，数据已插入；如果插件失败，数据和钩子已执行。

#### 修改后（使用事务）

```go
import "github.com/sky-xhsoft/sky-server/internal/pkg/transaction"

func (s *service) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) error {
    var result map[string]interface{}

    // 在事务中执行主操作
    err := transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // before钩子
        if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "begin", data); err != nil {
            return err
        }

        // 插入数据
        if err := tx.Table(table.Name).Create(&processedData).Error; err != nil {
            return err
        }

        // after钩子
        if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "end", processedData); err != nil {
            return err
        }

        result = processedData
        return nil
    })

    if err != nil {
        return err
    }

    // 插件在事务外执行（失败不影响主流程）
    if err := s.pluginManager.ExecutePlugins(ctx, pluginData); err != nil {
        // 只记录日志
        log.Error("Plugin execution failed", zap.Error(err))
    }

    return nil
}
```

### SSO Service

#### 修改前（无事务）

```go
func (s *service) Login(req *LoginRequest) (*LoginResponse, error) {
    // 查询用户
    user, _ := s.userRepo.GetUserByUsername(req.Username)

    // 生成Token...

    // 创建或更新会话
    existingSession, err := s.userRepo.GetSessionByDeviceID(user.ID, deviceID)
    if err == nil {
        s.userRepo.UpdateSession(session)
    } else {
        s.userRepo.CreateSession(session)
    }

    return &LoginResponse{...}, nil
}
```

**问题**：查询和创建/更新会话不在同一事务，并发时可能出现问题。

#### 修改后（使用事务）

```go
func (s *service) Login(req *LoginRequest) (*LoginResponse, error) {
    var response *LoginResponse

    err := transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // 查询用户（在事务中加锁）
        var user User
        if err := tx.Where("username = ?", req.Username).First(&user).Error; err != nil {
            return errors.InvalidCredentials
        }

        // 验证密码...

        // 生成Token...

        // 创建或更新会话（在事务中）
        var session UserSession
        err := tx.Where("user_id = ? AND device_id = ?", user.ID, deviceID).First(&session).Error

        if err == gorm.ErrRecordNotFound {
            // 创建新会话
            session = UserSession{...}
            if err := tx.Create(&session).Error; err != nil {
                return err
            }
        } else if err == nil {
            // 更新现有会话
            session.Token = token
            session.LastActiveTime = time.Now()
            if err := tx.Save(&session).Error; err != nil {
                return err
            }
        } else {
            return err
        }

        response = &LoginResponse{...}
        return nil
    })

    return response, err
}
```

## 最佳实践

### 1. 事务范围最小化

```go
// ✅ 好：事务只包含必要操作
func (s *service) Process() error {
    // 准备数据（事务外）
    data := prepareData()

    // 事务操作
    return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        return tx.Create(data).Error
    })
}

// ❌ 差：事务包含不必要的操作
func (s *service) Process() error {
    return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
        // 复杂计算（应该在事务外）
        data := complexCalculation()

        return tx.Create(data).Error
    })
}
```

### 2. 明确错误处理

```go
// ✅ 好：明确的错误处理
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    if err := tx.Create(&obj).Error; err != nil {
        return fmt.Errorf("创建失败: %w", err)
    }
    return nil
})

if err != nil {
    log.Error("Transaction failed", zap.Error(err))
    return err
}
```

### 3. 避免嵌套事务

```go
// ❌ 差：嵌套事务
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    // 外层事务

    return transaction.RunInTransaction(tx, func(tx2 *gorm.DB) error {
        // 内层事务 - 可能导致死锁
        return nil
    })
})

// ✅ 好：使用同一事务
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    // 所有操作使用同一个 tx
    tx.Create(&obj1)
    tx.Create(&obj2)
    return nil
})
```

### 4. 长事务处理

```go
// ❌ 差：事务中包含耗时操作
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    tx.Create(&obj)

    // 调用外部API（可能很慢）
    callExternalAPI()

    tx.Update(&obj)
    return nil
})

// ✅ 好：拆分事务
// 第一个事务
err1 := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    return tx.Create(&obj).Error
})

// 外部调用（事务外）
callExternalAPI()

// 第二个事务
err2 := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    return tx.Update(&obj).Error
})
```

## 性能考虑

### 1. 事务锁

事务会持有数据库锁，应尽快完成：

```go
// ✅ 快速事务
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    return tx.Where("id = ?", id).Updates(updates).Error
})

// ❌ 慢事务
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    time.Sleep(10 * time.Second) // 持有锁太久
    return tx.Updates(updates).Error
})
```

### 2. 死锁避免

始终以相同顺序访问表：

```go
// ✅ 好：按固定顺序
err := transaction.RunInTransaction(db, func(tx *gorm.DB) error {
    tx.Table("orders").Where(...).Update(...)      // 先orders
    tx.Table("order_items").Where(...).Update(...) // 后order_items
    return nil
})

// ❌ 可能死锁：不同顺序
// 事务A: orders -> order_items
// 事务B: order_items -> orders
```

## 测试

### 单元测试

```go
func TestService_CreateWithTransaction(t *testing.T) {
    // 使用真实数据库或testcontainers

    service := NewService(db)

    err := service.Create(...)
    if err != nil {
        t.Errorf("Create failed: %v", err)
    }

    // 验证数据
    var count int64
    db.Model(&Model{}).Count(&count)
    if count != 1 {
        t.Errorf("Expected 1 record")
    }
}
```

### 集成测试

```go
func TestTransactionRollback(t *testing.T) {
    // 模拟失败场景
    service := NewService(db)

    // 注入失败
    mockRepo.EXPECT().Something().Return(errors.New("simulated error"))

    err := service.Process()
    if err == nil {
        t.Error("Expected error")
    }

    // 验证回滚
    var count int64
    db.Model(&Model{}).Count(&count)
    if count != 0 {
        t.Errorf("Expected rollback, got %d records", count)
    }
}
```

## 常见问题

### Q: 什么时候用事务？

A: 当一个操作涉及多个数据库写操作，且它们必须全部成功或全部失败时。

### Q: 查询需要事务吗？

A: 通常不需要。除非你需要确保读取的一致性（如 SELECT FOR UPDATE）。

### Q: 插件应该在事务内还是外？

A: 取决于插件的作用。如果插件失败应该回滚主操作，放在事务内；如果插件是辅助功能（如发送通知），放在事务外。

### Q: 如何处理部分失败？

A: 不推荐部分提交。要么全部成功，要么全部回滚。如果真需要，拆分成多个独立事务。

## 迁移指南

### 步骤

1. 识别需要事务的方法
2. 引入 transaction 包
3. 使用 RunInTransaction 包装操作
4. 测试事务提交和回滚
5. 更新文档

### 优先级

1. 🔴 **高优先级**：CRUD操作、SSO登录
2. 🟡 **中优先级**：权限分配、工作流
3. 🟢 **低优先级**：单一操作、只读操作

## 总结

事务是确保数据一致性的关键机制。正确使用事务可以：

1. **保证数据一致性** - 多个操作要么全部成功，要么全部失败
2. **避免脏数据** - 防止部分操作失败导致数据不一致
3. **简化错误处理** - 自动回滚，无需手动清理

记住：**不是所有操作都需要事务，但涉及多个写操作的场景必须使用事务**。
