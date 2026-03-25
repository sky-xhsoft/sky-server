# Service 层事务控制分析

## 当前问题

Service 层中很多涉及多个数据库操作的方法没有使用事务控制，可能导致数据不一致问题。

## 需要事务控制的场景

### 1. SSO Service

#### Login 方法
**当前代码**:
```go
func (s *service) Login(req *LoginRequest) (*LoginResponse, error) {
    // 1. 查询用户
    user, err := s.userRepo.GetUserByUsername(req.Username)

    // 2. 验证密码、生成Token...

    // 3. 检查现有会话
    existingSession, err := s.userRepo.GetSessionByDeviceID(user.ID, deviceID)
    if err == nil {
        // 更新会话
        s.userRepo.UpdateSession(session)
    } else {
        // 创建会话
        s.userRepo.CreateSession(session)
    }
}
```

**问题**:
- 查询、更新/创建会话操作不在同一事务中
- 如果创建/更新会话失败，Token已生成并返回，导致状态不一致

**应该使用事务**: ✅

### 2. CRUD Service

#### Create 方法
**当前代码**:
```go
func (s *service) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) {
    // 1. 执行before钩子
    s.executeHooks(ctx, table.ID, "A", "begin", data)

    // 2. 插入数据
    s.db.Table(table.Name).Create(&processedData)

    // 3. 执行after钩子
    s.executeHooks(ctx, table.ID, "A", "end", processedData)

    // 4. 执行插件
    s.pluginManager.ExecutePlugins(ctx, pluginData)
}
```

**问题**:
- before钩子、插入、after钩子不在同一事务中
- 如果after钩子执行失败，数据已插入
- 插件执行失败不会回滚主操作（这个可能是期望行为）

**应该使用事务**: ✅ (钩子+插入，插件可独立)

#### Update 方法
**当前代码**:
```go
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) {
    // 1. 执行before钩子
    s.executeHooks(ctx, table.ID, "M", "begin", data)

    // 2. 更新数据
    s.db.Table(table.Name).Where("ID = ?", id).Updates(processedData)

    // 3. 执行after钩子
    s.executeHooks(ctx, table.ID, "M", "end", processedData)
}
```

**问题**: 同Create方法

**应该使用事务**: ✅

#### Delete 方法
**问题**: 同Create方法

**应该使用事务**: ✅

### 3. Groups Service

#### AssignPermissions 方法
**可能的实现**:
```go
func (s *service) AssignPermissions(ctx context.Context, groupID uint, permissions []*GroupPermission) error {
    // 1. 删除现有权限
    db.Where("group_id = ?", groupID).Delete(&SysGroupPrem{})

    // 2. 插入新权限（多条）
    for _, perm := range permissions {
        db.Create(perm)
    }
}
```

**问题**:
- 删除和批量插入不在同一事务
- 如果部分插入失败，已删除的权限无法恢复

**应该使用事务**: ✅

### 4. Workflow Service

工作流相关操作通常涉及多个表：
- wf_instance (实例)
- wf_task (任务)
- wf_node (节点状态)

**应该使用事务**: ✅

### 5. Action Service

执行动作时可能调用存储过程或执行多个SQL操作。

**应该使用事务**: 取决于具体实现

### 6. Audit Service

审计日志记录通常应该独立于主操作。

**应该使用事务**: ❌ (应该独立记录)

## 不需要事务的场景

### 1. 单一查询操作
- GetOne, GetList
- 只读操作，不需要事务

### 2. Metadata Service
- 元数据查询（通常有缓存）
- 不需要事务

### 3. Dict Service
- 字典查询
- 不需要事务

### 4. Sequence Service
- 序列号生成（使用Redis）
- 不需要事务

## GORM 事务使用

### 方式1: 手动事务
```go
func (s *service) SomeMethod() error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if err := tx.Create(&obj1).Error; err != nil {
        tx.Rollback()
        return err
    }

    if err := tx.Update(&obj2).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

### 方式2: 自动事务（推荐）
```go
func (s *service) SomeMethod() error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&obj1).Error; err != nil {
            return err // 自动回滚
        }

        if err := tx.Update(&obj2).Error; err != nil {
            return err // 自动回滚
        }

        return nil // 自动提交
    })
}
```

## 改进建议

### 1. 创建事务辅助工具

创建一个统一的事务管理工具，简化事务使用。

### 2. 修改 Service 接口

某些 Service 需要接受 `*gorm.DB` 参数，以支持在已有事务中执行。

### 3. 编写事务使用规范

明确哪些操作需要事务，如何正确使用事务。

### 4. 添加事务测试

测试事务回滚是否正常工作。

## 优先级

| Service | 方法 | 优先级 | 影响 |
|---------|------|--------|------|
| CRUD | Create/Update/Delete | 🔴 高 | 数据一致性核心 |
| SSO | Login | 🔴 高 | 会话状态不一致 |
| Groups | AssignPermissions | 🟡 中 | 权限可能不完整 |
| Workflow | 所有写操作 | 🟡 中 | 工作流状态混乱 |
| Action | 取决于实现 | 🟡 中 | 取决于动作类型 |

## 下一步

1. 创建事务辅助工具
2. 修复 CRUD Service
3. 修复 SSO Service
4. 修复 Groups Service
5. 编写测试验证
6. 编写使用规范文档
