# 元数据初始化工具 - sys_directory 自动初始化功能

**实现日期**: 2026-01-12

## 功能概述

为 `metadata-init` 工具添加 **sys_directory 自动初始化**功能，从 sys_table 读取所有表记录，自动为每个表创建对应的安全目录，并建立双向关联关系。

## 业务背景

### 权限系统架构

Sky-Server 的权限系统基于目录-表的映射关系：

```
用户 → 权限组 → 安全目录 (sys_directory) → 表单 (sys_table)
```

**核心关联**:
- `sys_directory.SYS_TABLE_ID` → `sys_table.ID` （目录指向表）
- `sys_table.SYS_DIRECTORY_ID` → `sys_directory.ID` （表指向目录）

### 问题

在旧版本中：
1. ❌ 只初始化 sys_table 和 sys_column
2. ❌ 不会自动创建 sys_directory
3. ❌ 需要手动配置表和目录的关联
4. ❌ 导致权限系统无法使用（无法通过目录授权）

### 解决方案

新版本自动：
1. ✅ 读取 sys_table 中的所有表
2. ✅ 为每个表创建对应的 sys_directory
3. ✅ 建立双向关联关系
4. ✅ 确保权限系统开箱即用

## 技术实现

### 1. 执行时机

metadata-init 工具在两个时机执行 directory 初始化：

```go
// 执行流程
1. 加载配置
2. 连接数据库
3. 执行 init.sql（可选）
4. 初始化基础数据字典
5. ⭐ 第一次：为已存在的 sys_table 创建目录
6. 初始化表元数据（可能新增 sys_table 记录）
7. ⭐ 第二次：为新增的 sys_table 创建目录
8. 完成
```

### 2. 核心函数 - initDirectoriesFromTables

**功能**: 为 sys_table 中的每个表创建对应的 sys_directory 并建立关联

**实现位置**: `cmd/metadata-init/main.go`

**完整实现**:

```go
func initDirectoriesFromTables(ctx context.Context, db *gorm.DB) error {
    // 1. 查询所有 sys_table 记录
    var tables []struct {
        ID          uint
        Name        string
        DisplayName string
        URL         string
    }

    err := db.WithContext(ctx).
        Table("sys_table").
        Select("ID, NAME, DISPLAY_NAME, URL").
        Where("IS_ACTIVE = ?", "Y").
        Find(&tables).Error

    if err != nil {
        return fmt.Errorf("query sys_table failed: %w", err)
    }

    if len(tables) == 0 {
        logger.Info("No tables found in sys_table, skipping directory initialization")
        return nil
    }

    // 2. 为每个表创建或更新 sys_directory
    createdCount := 0
    updatedCount := 0
    skippedCount := 0

    for _, table := range tables {
        // 检查是否已存在对应的 sys_directory
        var existingDirID uint
        err := db.WithContext(ctx).
            Table("sys_directory").
            Select("ID").
            Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", table.ID, "Y").
            Scan(&existingDirID).Error

        if err != nil && err != gorm.ErrRecordNotFound {
            logger.Error("Failed to check existing directory",
                zap.String("table", table.Name),
                zap.Error(err))
            continue
        }

        if existingDirID > 0 {
            // 目录已存在，检查 sys_table 的 SYS_DIRECTORY_ID 是否已设置
            var tableDirectoryID *uint
            db.WithContext(ctx).
                Table("sys_table").
                Select("SYS_DIRECTORY_ID").
                Where("ID = ?", table.ID).
                Scan(&tableDirectoryID)

            if tableDirectoryID == nil || *tableDirectoryID == 0 {
                // 更新 sys_table 的 SYS_DIRECTORY_ID
                if err := db.WithContext(ctx).
                    Table("sys_table").
                    Where("ID = ?", table.ID).
                    Update("SYS_DIRECTORY_ID", existingDirID).Error; err != nil {
                    logger.Error("Failed to update table's directory ID",
                        zap.String("table", table.Name),
                        zap.Uint("dirID", existingDirID),
                        zap.Error(err))
                    continue
                }
                updatedCount++
                logger.Info("Updated table's directory link",
                    zap.String("table", table.Name),
                    zap.Uint("dirID", existingDirID))
            } else {
                skippedCount++
            }
            continue
        }

        // 3. 在事务中创建 sys_directory 并更新 sys_table
        err = db.Transaction(func(tx *gorm.DB) error {
            // 创建 sys_directory 记录
            directoryData := map[string]interface{}{
                "NAME":           table.Name,
                "DISPLAY_NAME":   table.DisplayName,
                "URL":            table.URL,
                "SYS_TABLE_ID":   table.ID,
                "IS_ACTIVE":      "Y",
                "CREATE_BY":      "system",
                "CREATE_TIME":    time.Now(),
                "SYS_COMPANY_ID": 1,
            }

            if err := tx.Table("sys_directory").Create(&directoryData).Error; err != nil {
                return fmt.Errorf("create sys_directory failed: %w", err)
            }

            // 获取新创建的目录ID
            var dirID uint
            if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&dirID).Error; err != nil {
                return fmt.Errorf("failed to get directory ID: %w", err)
            }

            if dirID == 0 {
                return fmt.Errorf("directory ID is 0")
            }

            // 更新 sys_table 的 SYS_DIRECTORY_ID
            if err := tx.Table("sys_table").
                Where("ID = ?", table.ID).
                Update("SYS_DIRECTORY_ID", dirID).Error; err != nil {
                return fmt.Errorf("update sys_table.SYS_DIRECTORY_ID failed: %w", err)
            }

            logger.Info("Created directory and linked to table",
                zap.String("table", table.Name),
                zap.Uint("tableID", table.ID),
                zap.Uint("dirID", dirID))

            return nil
        })

        if err != nil {
            logger.Error("Failed to create directory for table",
                zap.String("table", table.Name),
                zap.Error(err))
            continue
        }

        createdCount++
    }

    // 4. 输出统计结果
    logger.Info("Directory initialization completed",
        zap.Int("created", createdCount),
        zap.Int("updated", updatedCount),
        zap.Int("skipped", skippedCount),
        zap.Int("total", len(tables)))

    return nil
}
```

### 3. 处理逻辑

#### 3.1 新建目录场景

**条件**: sys_directory 中不存在该表对应的记录

**操作**:
1. 创建 sys_directory 记录
2. 设置 `SYS_TABLE_ID` = 表ID
3. 复制表的 `NAME`, `DISPLAY_NAME`, `URL`
4. 获取新创建的目录ID
5. 更新 sys_table 的 `SYS_DIRECTORY_ID` = 目录ID

**事务保证**: 在单个事务中完成，确保原子性

#### 3.2 更新关联场景

**条件**: 目录已存在，但 sys_table.SYS_DIRECTORY_ID 未设置

**操作**:
1. 查询已存在的目录ID
2. 更新 sys_table 的 `SYS_DIRECTORY_ID`

**用途**: 修复历史数据的关联关系

#### 3.3 跳过场景

**条件**: 目录已存在，且 sys_table.SYS_DIRECTORY_ID 已正确设置

**操作**: 跳过处理，记录到统计信息

**优势**: 幂等性，可以重复执行

### 4. 数据映射

从 sys_table 到 sys_directory 的字段映射：

| sys_table 字段 | sys_directory 字段 | 说明 |
|---------------|-------------------|------|
| ID | SYS_TABLE_ID | 目录指向表 |
| NAME | NAME | 目录名称（使用表名） |
| DISPLAY_NAME | DISPLAY_NAME | 显示名称 |
| URL | URL | URL路径 |
| - | IS_ACTIVE | 固定为 'Y' |
| - | CREATE_BY | 固定为 'system' |
| - | CREATE_TIME | 当前时间 |
| - | SYS_COMPANY_ID | 固定为 1 |

反向关联：

| sys_directory 字段 | sys_table 字段 | 说明 |
|-------------------|---------------|------|
| ID | SYS_DIRECTORY_ID | 表指向目录 |

## 使用示例

### 示例 1: 首次初始化

```bash
# 初始化所有表和目录
metadata-init

# 输出：
# 2026-01-12T10:00:00.000+0800    INFO    Base dictionaries initialized
# 2026-01-12T10:00:00.100+0800    INFO    Found tables in sys_table {"count": 35}
# 2026-01-12T10:00:00.150+0800    INFO    Created directory and linked to table {"table": "SYS_USER", "tableID": 1, "dirID": 1}
# 2026-01-12T10:00:00.200+0800    INFO    Created directory and linked to table {"table": "SYS_COMPANY", "tableID": 2, "dirID": 2}
# ...
# 2026-01-12T10:00:01.000+0800    INFO    Directory initialization completed {"created": 35, "updated": 0, "skipped": 0, "total": 35}
# 2026-01-12T10:00:01.050+0800    INFO    Directories initialized from sys_table
```

### 示例 2: 增量初始化

```bash
# 只初始化新增的业务表
metadata-init --exclude-sys

# 输出：
# 2026-01-12T10:00:00.000+0800    INFO    Base dictionaries initialized
# 2026-01-12T10:00:00.100+0800    INFO    Found tables in sys_table {"count": 35}
# 2026-01-12T10:00:00.150+0800    INFO    Updated table's directory link {"table": "SYS_USER", "dirID": 1}  ← 更新已存在的关联
# 2026-01-12T10:00:00.200+0800    INFO    Directory initialization completed {"created": 0, "updated": 1, "skipped": 34, "total": 35}
# ...
# 2026-01-12T10:00:02.000+0800    INFO    Found tables {"count": 5, "filter": "business tables (excluding sys_*)"}
# 2026-01-12T10:00:03.000+0800    INFO    Metadata initialization completed {"success": 5, "skipped": 0, "failed": 0, "total": 5}
# 2026-01-12T10:00:03.100+0800    INFO    Created directory and linked to table {"table": "ORDER", "tableID": 36, "dirID": 36}  ← 为新表创建目录
# 2026-01-12T10:00:03.150+0800    INFO    Directory initialization completed {"created": 5, "updated": 0, "skipped": 35, "total": 40}
# 2026-01-12T10:00:03.200+0800    INFO    Directories synchronized after metadata creation
```

### 示例 3: 修复历史数据

**场景**: 已有 sys_table 和 sys_directory，但关联关系缺失

```bash
# 运行工具修复关联
metadata-init

# 输出：
# 2026-01-12T10:00:00.100+0800    INFO    Found tables in sys_table {"count": 35}
# 2026-01-12T10:00:00.150+0800    INFO    Updated table's directory link {"table": "SYS_USER", "dirID": 1}
# 2026-01-12T10:00:00.200+0800    INFO    Updated table's directory link {"table": "SYS_COMPANY", "dirID": 2}
# ...
# 2026-01-12T10:00:01.000+0800    INFO    Directory initialization completed {"created": 0, "updated": 35, "skipped": 0, "total": 35}
```

## 数据库验证

### 1. 检查目录是否创建

```sql
-- 查看 sys_directory 记录
SELECT
    d.ID AS dir_id,
    d.NAME AS dir_name,
    d.SYS_TABLE_ID AS table_id,
    t.NAME AS table_name,
    t.SYS_DIRECTORY_ID AS table_dir_id
FROM sys_directory d
LEFT JOIN sys_table t ON d.SYS_TABLE_ID = t.ID
WHERE d.IS_ACTIVE = 'Y'
ORDER BY d.ID;
```

**预期结果**: 每个 sys_table 都有对应的 sys_directory

### 2. 检查双向关联

```sql
-- 检查关联完整性
SELECT
    t.ID AS table_id,
    t.NAME AS table_name,
    t.SYS_DIRECTORY_ID AS table_dir_id,
    d.ID AS dir_id,
    d.NAME AS dir_name,
    CASE
        WHEN t.SYS_DIRECTORY_ID = d.ID AND d.SYS_TABLE_ID = t.ID THEN '✓ 关联正确'
        WHEN t.SYS_DIRECTORY_ID IS NULL THEN '✗ 表未关联目录'
        WHEN d.ID IS NULL THEN '✗ 目录不存在'
        ELSE '✗ 关联不一致'
    END AS status
FROM sys_table t
LEFT JOIN sys_directory d ON t.SYS_DIRECTORY_ID = d.ID
WHERE t.IS_ACTIVE = 'Y'
ORDER BY t.ID;
```

**预期结果**: 所有记录的 status 都是 "✓ 关联正确"

### 3. 检查孤立记录

```sql
-- 查找没有目录的表
SELECT ID, NAME, DISPLAY_NAME
FROM sys_table
WHERE IS_ACTIVE = 'Y'
  AND (SYS_DIRECTORY_ID IS NULL OR SYS_DIRECTORY_ID = 0);

-- 查找没有表的目录
SELECT ID, NAME, DISPLAY_NAME
FROM sys_directory
WHERE IS_ACTIVE = 'Y'
  AND (SYS_TABLE_ID IS NULL OR SYS_TABLE_ID = 0);
```

**预期结果**: 都应该返回空结果

## 权限系统集成

### 1. 授权流程

创建目录后，可以为权限组分配目录权限：

```sql
-- 为权限组分配目录权限
INSERT INTO sys_group_prem (
    SYS_GROUPS_ID,      -- 权限组ID
    SYS_DIRECTORY_ID,   -- 目录ID（自动创建的）
    PERMISSION,         -- 权限值（位运算）
    IS_ACTIVE,
    CREATE_BY,
    CREATE_TIME,
    SYS_COMPANY_ID
) VALUES (
    1,                  -- 管理员组
    1,                  -- SYS_USER 表的目录
    63,                 -- 全部权限（读、创建、更新、删除、导出、导入）
    'Y',
    'admin',
    NOW(),
    1
);
```

### 2. 权限检查

权限检查通过目录进行：

```go
// CheckUserTablePermission 会：
// 1. 通过 sys_table.SYS_DIRECTORY_ID 找到目录
// 2. 检查用户在该目录的权限
// 3. 返回是否有权限

hasPermission, err := groupsService.CheckUserTablePermission(
    ctx,
    userID,
    tableID,
    groups.PermRead,  // 检查读权限
)
```

### 3. 菜单显示

菜单系统也依赖目录权限：

```go
// GetUserMenuTree 会：
// 1. 查询用户的权限组
// 2. 获取权限组的目录列表
// 3. 通过 sys_directory.SYS_TABLE_ID 找到表
// 4. 只显示有权限的菜单
```

## 优势分析

### 1. 自动化程度高 ✅

**对比**:
- **旧版本**: 手动创建目录 → 手动关联表 → 配置权限
- **新版本**: 自动创建目录 → 自动关联 → 只需配置权限

**节省时间**: 原来 100 张表需要手动配置 100 次，现在 0 次

### 2. 数据一致性 ✅

**保证**:
- 事务保证创建和关联的原子性
- 双向关联自动建立，不会遗漏
- 自动修复不一致的关联关系

### 3. 幂等性 ✅

**特性**:
- 多次运行不会重复创建
- 已存在的记录会被跳过
- 关联缺失会自动修复

**用途**: 可以安全地重复执行，用于修复数据

### 4. 向后兼容 ✅

**兼容性**:
- 不影响已存在的目录
- 不破坏现有的权限配置
- 只补充缺失的记录

### 5. 开箱即用 ✅

**体验**:
- 初始化后权限系统立即可用
- 无需额外配置即可授权
- 降低系统部署难度

## 注意事项

### 1. 目录命名规范 ⚠️

目录名称直接使用表名（大写），建议：

```sql
-- 表结构设计时就规划好名称
CREATE TABLE `sys_user` COMMENT '系统用户';

-- metadata-init 会创建：
-- sys_directory.NAME = 'SYS_USER'
-- sys_directory.DISPLAY_NAME = '系统用户'
```

### 2. URL字段处理 ⚠️

如果 sys_table.URL 为空：
- sys_directory.URL 也会是空
- 不影响权限功能
- 但菜单跳转可能需要手动配置

### 3. 父子目录 ⚠️

当前实现：
- 所有目录都是平级的（ParentID = NULL）
- 不会自动创建目录层级结构
- 如需层级，需手动调整 sys_directory.PARENT_ID

### 4. 表类别关联 ⚠️

当前实现：
- 不设置 sys_directory.SYS_TABLE_CATEGORY_ID
- 如需按类别组织目录，需手动配置

## 常见问题

### Q1: 如果表已有手动创建的目录怎么办？

**A**: 工具会检测已存在的目录，不会重复创建，只会补充缺失的 sys_table.SYS_DIRECTORY_ID 关联。

### Q2: 如何为新表创建目录？

**A**: 运行 metadata-init 即可，工具会自动为所有缺少目录的表创建目录。

```bash
metadata-init
```

### Q3: 目录名称可以修改吗？

**A**: 可以，但建议不修改 NAME（系统内部使用），只修改 DISPLAY_NAME（用户显示）。

```sql
UPDATE sys_directory
SET DISPLAY_NAME = '用户管理'
WHERE NAME = 'SYS_USER';
```

### Q4: 如何删除目录？

**A**: 软删除，设置 IS_ACTIVE='N'

```sql
-- 软删除目录
UPDATE sys_directory
SET IS_ACTIVE = 'N',
    UPDATE_BY = 'admin',
    UPDATE_TIME = NOW()
WHERE ID = 1;

-- 同时清除 sys_table 的关联
UPDATE sys_table
SET SYS_DIRECTORY_ID = NULL
WHERE SYS_DIRECTORY_ID = 1;
```

### Q5: 为什么要执行两次 initDirectoriesFromTables？

**A**:
1. **第一次**: 在初始化表元数据之前，为已存在的 sys_table 创建目录
2. **第二次**: 在初始化表元数据之后，为新创建的 sys_table 创建目录

这样确保：
- 历史数据补充目录
- 新增数据也有目录
- 一次运行完成所有工作

### Q6: 如果只想为某些表创建目录怎么办？

**A**: 工具会为所有 IS_ACTIVE='Y' 的 sys_table 创建目录。如果只想为部分表创建，可以：

```sql
-- 临时禁用不需要的表
UPDATE sys_table
SET IS_ACTIVE = 'N'
WHERE NAME IN ('TEMP_TABLE1', 'TEMP_TABLE2');

-- 运行工具
-- metadata-init

-- 恢复表状态
UPDATE sys_table
SET IS_ACTIVE = 'Y'
WHERE NAME IN ('TEMP_TABLE1', 'TEMP_TABLE2');
```

## 后续优化建议

### 1. 支持目录层级 ⭐

```go
// 根据 sys_table_category 自动创建目录层级
// - 子系统作为一级目录
// - 表类别作为二级目录
// - 表作为三级目录
```

### 2. 目录模板 ⭐

```go
// 支持目录创建模板
type DirectoryTemplate struct {
    DefaultPermission int    // 默认权限
    RequireAuth      bool   // 是否需要认证
    CustomFields     map[string]string
}
```

### 3. 批量操作优化 ⭐

```go
// 使用批量插入提升性能
// 当前：逐条创建
// 优化：批量创建
tx.CreateInBatches(directories, 100)
```

### 4. 配置化 ⭐

```yaml
# metadata-init.yaml
directory:
  auto_create: true              # 是否自动创建
  name_template: "${TABLE_NAME}" # 命名模板
  inherit_category: true         # 继承表类别
  default_permission: 1          # 默认权限（只读）
```

## 版本历史

| 版本 | 日期 | 变更说明 |
|-----|------|---------|
| v1.0.0 | 2026-01-12 | 初始实现，自动为 sys_table 创建 sys_directory |

## 相关文档

- [元数据初始化工具使用指南](./metadata-init-guide.md)
- [权限系统文档](./admin-permission-feature.md)
- [菜单系统文档](./menu-system.md)

## 总结

sys_directory 自动初始化功能带来的价值：

✅ **100% 自动化**: 无需手动创建目录和关联
✅ **数据一致性**: 事务保证原子性，自动修复不一致
✅ **幂等性**: 可安全重复执行，用于数据修复
✅ **权限系统可用**: 初始化后即可配置权限授权
✅ **降低门槛**: 简化系统部署和维护流程
✅ **向后兼容**: 不影响已有数据和配置

这个功能是权限系统的基础设施，确保元数据驱动架构的完整性！🎉
