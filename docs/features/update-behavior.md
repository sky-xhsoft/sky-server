# Update 操作行为说明

## 当前行为

### 核心原则
**只更新传入的字段，忽略未传入的字段** ✅

### 实现机制

#### 1. Handler 层
```go
// 接收客户端传入的数据
var data map[string]interface{}
c.ShouldBindJSON(&data)

// 传递给 Service
service.Update(ctx, tableName, id, data, userID)
```

#### 2. Service 层处理流程

```go
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
    // 1. 获取字段定义
    columns := s.metadataService.GetColumns(table.ID)

    // 2. 处理字段（只保留传入的字段）
    processedData := s.processFieldsForUpdate(columns, data, userID)

    // 3. 执行更新（只更新 processedData 中的字段）
    s.db.Table(table.Name).Where("ID = ?", id).Updates(processedData)
}
```

#### 3. 字段处理逻辑

```go
func (s *service) processFieldsForUpdate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
    processedData := make(map[string]interface{})

    for _, col := range columns {
        // 检查 MASK 权限
        if col.Mask != "" {
            fieldMask := mask.ParseMask(col.Mask)
            if !fieldMask.IsEditable("edit") {
                continue // 不可编辑的字段跳过
            }
        }

        // 只处理传入的字段
        value, exists := data[col.DbName]
        if exists {
            processedData[col.DbName] = value
        }
    }

    return processedData, nil
}
```

### 示例

#### 场景1：部分更新

**数据库当前记录**：
```json
{
  "ID": 1,
  "NAME": "张三",
  "AGE": 30,
  "EMAIL": "zhangsan@example.com",
  "PHONE": "13800138000"
}
```

**客户端请求**：
```json
PUT /api/v1/data/users/1
{
  "NAME": "李四"
}
```

**结果**：
```json
{
  "ID": 1,
  "NAME": "李四",          ← 已更新
  "AGE": 30,              ← 未变
  "EMAIL": "zhangsan@example.com",  ← 未变
  "PHONE": "13800138000"  ← 未变
}
```

✅ **正确**：只更新了传入的 NAME 字段

#### 场景2：多字段更新

**客户端请求**：
```json
PUT /api/v1/data/users/1
{
  "NAME": "王五",
  "EMAIL": "wangwu@example.com"
}
```

**结果**：
```json
{
  "ID": 1,
  "NAME": "王五",         ← 已更新
  "AGE": 30,             ← 未变
  "EMAIL": "wangwu@example.com",  ← 已更新
  "PHONE": "13800138000" ← 未变
}
```

✅ **正确**：只更新了传入的 NAME 和 EMAIL 字段

## 零值问题

### 当前限制

GORM 的 `Updates` 方法会**忽略零值**：

| Go类型 | 零值 | 行为 |
|--------|------|------|
| int, uint | 0 | ❌ 被忽略 |
| string | "" | ❌ 被忽略 |
| bool | false | ❌ 被忽略 |
| float | 0.0 | ❌ 被忽略 |
| pointer | nil | ❌ 被忽略 |

### 问题示例

**场景：清空邮箱**

**客户端请求**：
```json
PUT /api/v1/data/users/1
{
  "EMAIL": ""
}
```

**预期结果**：EMAIL 字段被清空

**实际结果**：EMAIL 字段未变化 ❌

**原因**：GORM 忽略了空字符串

## 解决方案

### 方案1：使用 Select（推荐）

修改 Update 方法，明确指定要更新的字段：

```go
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
    // ... 前面的逻辑不变 ...

    // 获取要更新的字段列表
    updateFields := make([]string, 0, len(processedData))
    for field := range processedData {
        updateFields = append(updateFields, field)
    }

    // 使用 Select 明确指定要更新的字段（包括零值）
    result := s.db.Table(table.Name).
        Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
        Select(updateFields).  // ← 关键：明确指定字段
        Updates(processedData)

    // ... 后续逻辑不变 ...
}
```

**优点**：
- ✅ 支持零值更新
- ✅ 只更新传入的字段
- ✅ 代码改动小

**缺点**：
- ⚠️ 需要 GORM v1.20+

### 方案2：使用 Update 单字段

对每个字段单独更新：

```go
for field, value := range processedData {
    if err := tx.Table(table.Name).
        Where("ID = ?", id).
        Update(field, value).Error; err != nil {
        return err
    }
}
```

**优点**：
- ✅ 支持零值更新
- ✅ 兼容所有 GORM 版本

**缺点**：
- ❌ 性能较差（多次SQL）
- ❌ 不在同一个原子操作中

### 方案3：区分零值和未传入

使用指针或特殊标记：

```go
type UpdateRequest struct {
    Name  *string `json:"name"`   // nil = 未传入，"" = 清空
    Age   *int    `json:"age"`    // nil = 未传入，0 = 设为0
    Email *string `json:"email"`
}
```

**优点**：
- ✅ 明确区分"未传入"和"零值"
- ✅ 类型安全

**缺点**：
- ❌ 需要为每个表定义结构体
- ❌ 失去了通用性

## 已实现的解决方案 ✅

### Update 方法（已修改）

```go
// Update 更新记录
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
    // 获取表元数据
    table, err := s.metadataService.GetTable(tableName)
    if err != nil {
        return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
    }

    // 检查更新权限
    hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermUpdate)
    if err != nil {
        return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
    }
    if !hasPermission {
        return errors.New(errors.ErrPermissionDenied, "无修改权限")
    }

    // 添加ID到数据中供钩子使用
    data["ID"] = id

    // 执行before钩子
    if err := s.executeHooks(ctx, table.ID, "M", "begin", data); err != nil {
        return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
    }

    // 获取字段定义
    columns, err := s.metadataService.GetColumns(table.ID)
    if err != nil {
        return err
    }

    // 验证和处理字段（根据MASK和权限）
    processedData, err := s.processFieldsForUpdate(columns, data, userID)
    if err != nil {
        return err
    }

    // 如果没有要更新的字段，直接返回
    if len(processedData) == 0 {
        return errors.New(errors.ErrValidation, "没有可更新的字段")
    }

    // 添加审计字段
    // TODO: 从context获取用户名
    processedData["UPDATE_TIME"] = time.Now()
    // processedData["UPDATE_BY"] = username

    // 获取要更新的字段列表
    updateFields := make([]string, 0, len(processedData))
    for field := range processedData {
        updateFields = append(updateFields, field)
    }

    // 执行更新（使用 Select 支持零值更新）
    result := s.db.Table(table.Name).
        Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
        Select(updateFields).  // ← 关键改动
        Updates(processedData)

    if result.Error != nil {
        return errors.Wrap(errors.ErrDatabase, "更新失败", result.Error)
    }

    if result.RowsAffected == 0 {
        return errors.New(errors.ErrResourceNotFound, "记录不存在")
    }

    // 执行after钩子
    processedData["ID"] = id
    if err := s.executeHooks(ctx, table.ID, "M", "end", processedData); err != nil {
        return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
    }

    return nil
}
```

### 关键改动

```go
// 改动前
result := s.db.Table(table.Name).
    Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
    Updates(processedData)

// 改动后
updateFields := make([]string, 0, len(processedData))
for field := range processedData {
    updateFields = append(updateFields, field)
}

result := s.db.Table(table.Name).
    Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
    Select(updateFields).  // ← 明确指定要更新的字段
    Updates(processedData)
```

## 测试用例

### 测试1：部分字段更新
```go
func TestUpdate_PartialFields(t *testing.T) {
    // 创建记录
    record := map[string]interface{}{
        "NAME": "张三",
        "AGE": 30,
        "EMAIL": "zhangsan@example.com",
    }
    service.Create(ctx, "users", record, userID)

    // 只更新 NAME
    updateData := map[string]interface{}{
        "NAME": "李四",
    }
    err := service.Update(ctx, "users", 1, updateData, userID)
    assert.NoError(t, err)

    // 验证：NAME已更新，其他字段未变
    result, _ := service.GetOne(ctx, "users", 1, userID)
    assert.Equal(t, "李四", result["NAME"])
    assert.Equal(t, 30, result["AGE"])
    assert.Equal(t, "zhangsan@example.com", result["EMAIL"])
}
```

### 测试2：零值更新
```go
func TestUpdate_ZeroValue(t *testing.T) {
    // 创建记录
    record := map[string]interface{}{
        "NAME": "张三",
        "AGE": 30,
        "EMAIL": "zhangsan@example.com",
    }
    service.Create(ctx, "users", record, userID)

    // 更新 AGE 为 0
    updateData := map[string]interface{}{
        "AGE": 0,
    }
    err := service.Update(ctx, "users", 1, updateData, userID)
    assert.NoError(t, err)

    // 验证：AGE 已更新为 0
    result, _ := service.GetOne(ctx, "users", 1, userID)
    assert.Equal(t, 0, result["AGE"])
}
```

### 测试3：清空字符串
```go
func TestUpdate_EmptyString(t *testing.T) {
    // 创建记录
    record := map[string]interface{}{
        "NAME": "张三",
        "EMAIL": "zhangsan@example.com",
    }
    service.Create(ctx, "users", record, userID)

    // 清空 EMAIL
    updateData := map[string]interface{}{
        "EMAIL": "",
    }
    err := service.Update(ctx, "users", 1, updateData, userID)
    assert.NoError(t, err)

    // 验证：EMAIL 已清空
    result, _ := service.GetOne(ctx, "users", 1, userID)
    assert.Equal(t, "", result["EMAIL"])
}
```

## 注意事项

### 1. 系统字段保护

某些字段不应该被客户端直接更新：

```go
// 在 processFieldsForUpdate 中过滤系统字段
protectedFields := []string{"ID", "CREATE_BY", "CREATE_TIME", "IS_ACTIVE"}

for _, col := range columns {
    // 跳过受保护的字段
    if contains(protectedFields, col.DbName) {
        continue
    }

    // ... 其他逻辑
}
```

### 2. MASK 权限控制

某些字段可能根据 MASK 不可编辑：

```go
if col.Mask != "" {
    fieldMask := mask.ParseMask(col.Mask)
    if !fieldMask.IsEditable("edit") {
        continue // 不可编辑的字段忽略
    }
}
```

### 3. 审计字段自动填充

UPDATE_TIME 和 UPDATE_BY 应该自动填充：

```go
processedData["UPDATE_TIME"] = time.Now()
processedData["UPDATE_BY"] = getCurrentUsername(ctx)
```

## API 使用示例

### 更新单个字段
```bash
curl -X PUT http://localhost:9090/api/v1/data/users/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "NAME": "新名字"
  }'
```

### 更新多个字段
```bash
curl -X PUT http://localhost:9090/api/v1/data/users/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "NAME": "新名字",
    "EMAIL": "new@example.com",
    "PHONE": "13900139000"
  }'
```

### 清空字段（零值）
```bash
curl -X PUT http://localhost:9090/api/v1/data/users/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "EMAIL": "",
    "AGE": 0
  }'
```

## 总结

### 实现状态

✅ **已经实现**："只更新传入的字段，忽略未传入的字段"

✅ **已经修复**：零值更新问题（2026-01-12 实现）

### 实现方案

使用 `Select` 方法明确指定要更新的字段：

```go
// 获取要更新的字段列表（支持零值更新）
updateFields := make([]string, 0, len(processedData))
for field := range processedData {
    updateFields = append(updateFields, field)
}

// 执行更新（使用 Select 明确指定要更新的字段，包括零值）
result := s.db.Table(table.Name).
    Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
    Select(updateFields).
    Updates(processedData)
```

### 收益

- ✅ 支持零值更新
- ✅ 保持"只更新传入字段"的语义
- ✅ 代码改动最小
- ✅ 性能优化（单个SQL语句）

### 修改文件

- `internal/service/crud/crud_service.go` (行 375-385)
