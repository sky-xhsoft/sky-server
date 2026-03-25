# Go 钩子使用指南

## 概述

Go 钩子（hook）是系统中执行效率最高的钩子类型，它允许在 CRUD 操作的 before/after 阶段执行自定义的 Go 函数。与其他类型的钩子（js、py、bsh）不同，Go 钩子在进程内执行，可以直接访问事务数据库连接，确保与主操作的数据一致性。

## 特性

### ✅ 优势

1. **高性能**：在进程内执行，无需启动外部进程
2. **类型安全**：使用 Go 的类型系统，编译时检查
3. **事务一致性**：可访问事务数据库连接，与主操作在同一事务中
4. **调试方便**：可使用 Go 的标准调试工具
5. **直接访问内部 API**：可以调用任何 Go 包和服务

### ⚠️ 注意事项

1. **需要编译**：Go 钩子需要编译到主程序中，不能动态加载
2. **需要重启**：修改 Go 钩子需要重新编译和重启服务
3. **错误处理**：Go 钩子中的错误会导致整个事务回滚

## 注册 Go 钩子

### 基本用法

在应用启动时注册 Go 钩子函数：

```go
package hooks

import (
    "fmt"
    "github.com/sky-xhsoft/sky-server/internal/pkg/executor"
    "gorm.io/gorm"
)

func init() {
    // 注册钩子函数
    executor.RegisterGoFunc("myHookFunction", myHookFunction)
}

// 钩子函数签名
func myHookFunction(params map[string]interface{}) (interface{}, error) {
    // 实现逻辑
    return map[string]interface{}{
        "success": true,
    }, nil
}
```

### 函数签名

所有 Go 钩子函数必须遵循以下签名：

```go
func(params map[string]interface{}) (interface{}, error)
```

**参数**：
- `params`: 包含业务数据和数据库连接的 map
  - 业务数据：从请求中传入的字段（ID, NAME, 等）
  - `__db__`: *gorm.DB 类型的数据库连接（在事务中时为事务连接）

**返回值**：
- `interface{}`: 执行结果数据（可选）
- `error`: 错误信息，如果返回非 nil，事务将回滚

## 访问数据库连接

### 获取事务连接

```go
func myHookWithDB(params map[string]interface{}) (interface{}, error) {
    // 从 params 获取数据库连接
    db, ok := params["__db__"].(*gorm.DB)
    if !ok || db == nil {
        return nil, fmt.Errorf("数据库连接不可用")
    }

    // 现在可以使用 db 执行数据库操作
    // 这些操作与主操作在同一事务中
    return nil, nil
}
```

### 在事务中查询数据

```go
func validateUser(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)
    userID := params["USER_ID"].(uint)

    // 查询用户
    var user User
    if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
        return nil, fmt.Errorf("用户不存在: %w", err)
    }

    // 验证用户状态
    if user.Status != "active" {
        return nil, fmt.Errorf("用户未激活")
    }

    return map[string]interface{}{
        "user": user,
        "validated": true,
    }, nil
}
```

### 在事务中更新数据

```go
func updateRelatedRecords(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)
    parentID := params["ID"].(uint)
    name := params["NAME"].(string)

    // 更新相关记录
    result := db.Table("child_table").
        Where("parent_id = ?", parentID).
        Update("parent_name", name)

    if result.Error != nil {
        return nil, result.Error
    }

    return map[string]interface{}{
        "updated_count": result.RowsAffected,
    }, nil
}
```

## 实战示例

### 示例 1：创建订单时扣减库存

```go
func init() {
    executor.RegisterGoFunc("deductInventory", deductInventory)
}

func deductInventory(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)

    // 获取订单项数据
    productID := params["PRODUCT_ID"].(uint)
    quantity := int(params["QUANTITY"].(float64))

    // 查询当前库存（使用 FOR UPDATE 加锁）
    var inventory struct {
        ID       uint
        Quantity int
    }
    if err := db.Table("inventory").
        Where("product_id = ?", productID).
        First(&inventory).Error; err != nil {
        return nil, fmt.Errorf("产品不存在: %w", err)
    }

    // 检查库存是否充足
    if inventory.Quantity < quantity {
        return nil, fmt.Errorf("库存不足: 需要 %d, 可用 %d", quantity, inventory.Quantity)
    }

    // 扣减库存
    if err := db.Table("inventory").
        Where("id = ?", inventory.ID).
        Update("quantity", gorm.Expr("quantity - ?", quantity)).Error; err != nil {
        return nil, fmt.Errorf("扣减库存失败: %w", err)
    }

    return map[string]interface{}{
        "deducted": quantity,
        "remaining": inventory.Quantity - quantity,
    }, nil
}
```

**sys_table_cmd 配置**：
```sql
INSERT INTO sys_table_cmd (SYS_TABLE_ID, CONTENT_TYPE, ACTION, EVENT, CONTENT, ORDER_NUM, IS_ACTIVE)
VALUES (
    (SELECT ID FROM sys_table WHERE NAME = 'order_items'),
    'go',           -- Go 类型钩子
    'A',            -- 创建操作
    'begin',        -- before 钩子（在插入前执行）
    'deductInventory',  -- 函数名
    1,
    'Y'
);
```

### 示例 2：更新用户时同步缓存表

```go
func init() {
    executor.RegisterGoFunc("syncUserCache", syncUserCache)
}

func syncUserCache(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)
    userID := params["ID"].(uint)

    // 查询用户完整信息
    var user struct {
        ID       uint
        Username string
        TrueName string
        Email    string
    }
    if err := db.Table("sys_user").
        Where("id = ?", userID).
        First(&user).Error; err != nil {
        return nil, err
    }

    // 更新缓存表
    cacheData := map[string]interface{}{
        "user_id":   user.ID,
        "username":  user.Username,
        "true_name": user.TrueName,
        "email":     user.Email,
        "updated_at": time.Now(),
    }

    // 使用 UPSERT 逻辑
    if err := db.Table("user_cache").
        Where("user_id = ?", userID).
        Updates(cacheData).Error; err != nil {
        // 如果不存在，则插入
        if err := db.Table("user_cache").Create(cacheData).Error; err != nil {
            return nil, fmt.Errorf("同步缓存失败: %w", err)
        }
    }

    return map[string]interface{}{
        "synced": true,
    }, nil
}
```

**sys_table_cmd 配置**：
```sql
INSERT INTO sys_table_cmd (SYS_TABLE_ID, CONTENT_TYPE, ACTION, EVENT, CONTENT, ORDER_NUM, IS_ACTIVE)
VALUES (
    (SELECT ID FROM sys_table WHERE NAME = 'sys_user'),
    'go',
    'M',            -- 更新操作
    'end',          -- after 钩子（更新后执行）
    'syncUserCache',
    1,
    'Y'
);
```

### 示例 3：删除前检查依赖

```go
func init() {
    executor.RegisterGoFunc("checkDependencies", checkDependencies)
}

func checkDependencies(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)
    categoryID := params["ID"].(uint)

    // 检查是否有产品依赖此分类
    var productCount int64
    if err := db.Table("products").
        Where("category_id = ?", categoryID).
        Count(&productCount).Error; err != nil {
        return nil, err
    }

    if productCount > 0 {
        return nil, fmt.Errorf("该分类下有 %d 个产品，无法删除", productCount)
    }

    // 检查是否有子分类
    var childCount int64
    if err := db.Table("categories").
        Where("parent_id = ?", categoryID).
        Count(&childCount).Error; err != nil {
        return nil, err
    }

    if childCount > 0 {
        return nil, fmt.Errorf("该分类下有 %d 个子分类，无法删除", childCount)
    }

    return map[string]interface{}{
        "can_delete": true,
    }, nil
}
```

**sys_table_cmd 配置**：
```sql
INSERT INTO sys_table_cmd (SYS_TABLE_ID, CONTENT_TYPE, ACTION, EVENT, CONTENT, ORDER_NUM, IS_ACTIVE)
VALUES (
    (SELECT ID FROM sys_table WHERE NAME = 'categories'),
    'go',
    'D',            -- 删除操作
    'begin',        -- before 钩子（删除前检查）
    'checkDependencies',
    1,
    'Y'
);
```

## 最佳实践

### 1. 错误处理

```go
func myHook(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)

    // ✅ 好：清晰的错误信息
    if err := db.Create(&record).Error; err != nil {
        return nil, fmt.Errorf("创建记录失败: %w", err)
    }

    // ❌ 差：模糊的错误信息
    if err := db.Create(&record).Error; err != nil {
        return nil, err
    }

    return nil, nil
}
```

### 2. 类型断言安全

```go
func myHook(params map[string]interface{}) (interface{}, error) {
    // ✅ 好：安全的类型断言
    db, ok := params["__db__"].(*gorm.DB)
    if !ok || db == nil {
        return nil, fmt.Errorf("数据库连接不可用")
    }

    quantity, ok := params["QUANTITY"].(float64)
    if !ok {
        return nil, fmt.Errorf("QUANTITY 字段类型错误")
    }

    // ❌ 差：不安全的类型断言（可能 panic）
    db := params["__db__"].(*gorm.DB)
    quantity := params["QUANTITY"].(float64)

    return nil, nil
}
```

### 3. 事务操作

```go
func myHook(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)

    // ✅ 好：使用传入的 db（已经在事务中）
    if err := db.Create(&record).Error; err != nil {
        return nil, err
    }

    // ❌ 差：不要创建新的事务（已经在事务中了）
    // db.Transaction(func(tx *gorm.DB) error {
    //     return tx.Create(&record).Error
    // })

    return nil, nil
}
```

### 4. 获取参数

```go
func myHook(params map[string]interface{}) (interface{}, error) {
    // ✅ 好：使用辅助函数获取参数
    userID, err := getUintParam(params, "USER_ID")
    if err != nil {
        return nil, err
    }

    name, err := getStringParam(params, "NAME")
    if err != nil {
        return nil, err
    }

    return nil, nil
}

// 辅助函数
func getUintParam(params map[string]interface{}, key string) (uint, error) {
    value, exists := params[key]
    if !exists {
        return 0, fmt.Errorf("参数 %s 不存在", key)
    }

    switch v := value.(type) {
    case uint:
        return v, nil
    case float64:
        return uint(v), nil
    case int:
        return uint(v), nil
    default:
        return 0, fmt.Errorf("参数 %s 类型错误", key)
    }
}

func getStringParam(params map[string]interface{}, key string) (string, error) {
    value, exists := params[key]
    if !exists {
        return "", fmt.Errorf("参数 %s 不存在", key)
    }

    str, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("参数 %s 类型错误", key)
    }

    return str, nil
}
```

## 调试

### 1. 添加日志

```go
import "go.uber.org/zap"

func myHook(params map[string]interface{}) (interface{}, error) {
    logger := zap.L()
    logger.Info("执行钩子", zap.Any("params", params))

    // 业务逻辑
    result, err := doSomething(params)

    if err != nil {
        logger.Error("钩子执行失败", zap.Error(err))
        return nil, err
    }

    logger.Info("钩子执行成功", zap.Any("result", result))
    return result, nil
}
```

### 2. 单元测试

```go
func TestMyHook(t *testing.T) {
    // 准备测试数据库
    db := setupTestDB(t)

    // 准备参数
    params := map[string]interface{}{
        "__db__": db,
        "ID":     uint(1),
        "NAME":   "测试",
    }

    // 执行钩子
    result, err := myHook(params)

    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, result)

    // 验证数据库状态
    var record Record
    err = db.First(&record, 1).Error
    assert.NoError(t, err)
}
```

## 性能优化

### 1. 减少数据库查询

```go
// ✅ 好：一次查询获取所有需要的数据
func myHook(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)

    var results []struct {
        ProductID uint
        Name      string
        Quantity  int
    }

    err := db.Table("products").
        Select("id as product_id, name, quantity").
        Where("category_id = ?", categoryID).
        Find(&results).Error

    // 处理数据...
    return nil, nil
}

// ❌ 差：N+1 查询问题
func myHook(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)

    var productIDs []uint
    db.Table("products").Where("category_id = ?", categoryID).Pluck("id", &productIDs)

    for _, id := range productIDs {
        var product Product
        db.First(&product, id)  // 每个 ID 一次查询
        // 处理...
    }
    return nil, nil
}
```

### 2. 使用批量操作

```go
// ✅ 好：批量插入
func myHook(params map[string]interface{}) (interface{}, error) {
    db := params["__db__"].(*gorm.DB)

    records := []Record{
        {Name: "A"},
        {Name: "B"},
        {Name: "C"},
    }

    // 批量插入
    if err := db.CreateInBatches(records, 100).Error; err != nil {
        return nil, err
    }

    return nil, nil
}
```

## 总结

Go 钩子是实现复杂业务逻辑的最佳选择：

### 何时使用 Go 钩子

✅ **适合**：
- 需要访问数据库的业务逻辑
- 需要与主操作在同一事务中
- 性能要求高的场景
- 复杂的数据验证和处理
- 需要调用内部服务和 API

❌ **不适合**：
- 需要动态修改逻辑（使用 js/py 钩子）
- 调用外部 HTTP 服务（使用 url 钩子）
- 执行系统命令（使用 bsh 钩子）

### 关键要点

1. **注册函数**：使用 `executor.RegisterGoFunc()` 注册钩子函数
2. **访问数据库**：通过 `params["__db__"]` 获取事务连接
3. **错误处理**：返回错误会导致事务回滚
4. **类型安全**：使用类型断言时要检查是否成功
5. **性能优化**：减少数据库查询，使用批量操作
