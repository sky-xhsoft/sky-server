# 元数据初始化工具升级总结

**升级日期**: 2026-01-12

## 升级概述

将 `metadata-init` 工具从只支持业务表初始化，升级为支持**全数据库所有表单的默认初始化**，并添加灵活的命令行参数控制。

## 主要改动

### 1. 新增命令行参数支持 ⭐

| 参数 | 功能 | 示例 |
|------|------|------|
| `--exclude-sys` | 排除 sys_ 开头的系统表 | `metadata-init --exclude-sys` |
| `--only-sys` | 只初始化 sys_ 开头的系统表 | `metadata-init --only-sys` |
| `--tables <names>` | 指定要初始化的表名（逗号分隔） | `metadata-init --tables user,order` |
| `--force` | 强制重新初始化已存在的表 | `metadata-init --force` |
| `--help` | 显示帮助信息 | `metadata-init --help` |

### 2. 默认行为变更 ⭐ 重要

**升级前**:
- 只初始化非 `sys_` 开头的业务表
- 代码和注释不一致（代码是 `LIKE 'sys_%'`，注释说排除）

**升级后**:
- **默认初始化所有表**（包括 sys_ 开头的系统表）
- 可通过 `--exclude-sys` 恢复旧行为

### 3. 新增功能

#### 3.1 强制重新初始化 (`--force`)

```bash
# 重新初始化所有表
metadata-init --force

# 重新初始化指定表
metadata-init --tables sys_user,sys_company --force
```

**功能说明**:
- 检测到已存在的表时，先删除旧元数据（sys_table 和 sys_column）
- 然后重新创建元数据
- 适用于表结构变更后需要更新元数据的场景

#### 3.2 指定表初始化 (`--tables`)

```bash
# 只初始化指定的几张表
metadata-init --tables user,order,product

# 只初始化系统表中的部分表
metadata-init --tables sys_user,sys_company
```

**功能说明**:
- 可以精确控制要初始化的表
- 使用逗号分隔多个表名
- 会自动忽略 `--exclude-sys` 和 `--only-sys` 参数

#### 3.3 过滤模式

```bash
# 只初始化业务表（旧行为）
metadata-init --exclude-sys

# 只初始化系统表
metadata-init --only-sys

# 初始化所有表（新默认行为）
metadata-init
```

### 4. 代码改进

#### 4.1 函数签名更新

**升级前**:
```go
func getTables(ctx context.Context, db *gorm.DB, dbName string) ([]TableInfo, error)
```

**升级后**:
```go
func getTables(ctx context.Context, db *gorm.DB, dbName string, excludeSys, onlySys bool) ([]TableInfo, error)
```

#### 4.2 新增函数

```go
// 显示帮助信息
func printHelp()

// 获取指定的表
func getSpecificTables(ctx context.Context, db *gorm.DB, dbName string, tableNames []string) ([]TableInfo, error)
```

#### 4.3 删除未使用的函数

```go
// 已删除
func getUintValue(m map[string]interface{}, key string) uint
```

#### 4.4 更新 initTableMetadata 函数

**新增 force 参数**:
```go
func initTableMetadata(ctx context.Context, db *gorm.DB, dbName string, table TableInfo, force bool) error
```

**功能增强**:
- 支持强制删除已存在的元数据
- 返回特殊错误 "table already exists" 用于跳过已存在的表
- 在事务中删除旧数据，确保数据一致性

### 5. 输出改进

**升级前**:
```
Found tables (count: 5)
Metadata initialization completed (success: 5, failed: 0, total: 5)
```

**升级后**:
```
Found tables (count: 35, filter: all tables)
Metadata initialization completed (success: 30, skipped: 3, failed: 2, total: 35)
```

**新增信息**:
- `filter`: 显示当前使用的过滤模式
- `skipped`: 显示跳过的表数量（已存在且未使用 --force）

## 使用示例

### 场景 1: 首次初始化整个数据库

```bash
# 初始化所有表（包括系统表）
metadata-init
```

**输出**:
```
Found tables (count: 35, filter: all tables)
Processing table: sys_user
Created table metadata (table: sys_user, table_id: 1, columns: 15)
...
Metadata initialization completed (success: 35, skipped: 0, failed: 0, total: 35)
```

### 场景 2: 只初始化新增的业务表

```bash
# 排除系统表，只看业务表
metadata-init --exclude-sys
```

**输出**:
```
Found tables (count: 8, filter: business tables (excluding sys_*))
Processing table: user
Table already exists in sys_table, skipping (table: user)
Processing table: order
Created table metadata (table: order, table_id: 36, columns: 10)
...
Metadata initialization completed (success: 1, skipped: 7, failed: 0, total: 8)
```

### 场景 3: 表结构变更后重新初始化

```bash
# 强制重新初始化指定的表
metadata-init --tables sys_user,user --force
```

**输出**:
```
Found specified tables (count: 2)
Processing table: sys_user
Force mode enabled, deleting existing metadata (table: sys_user, table_id: 1)
Deleted existing metadata (table: sys_user, table_id: 1)
Created table metadata (table: sys_user, table_id: 37, columns: 15)
...
Metadata initialization completed (success: 2, skipped: 0, failed: 0, total: 2)
```

### 场景 4: 查看帮助信息

```bash
metadata-init --help
```

**输出**:
```
元数据初始化工具 - Sky-Server Metadata Initializer

用法:
  metadata-init [参数]

参数:
  --exclude-sys      排除 sys_ 开头的系统表（只初始化业务表）
  --only-sys         只初始化 sys_ 开头的系统表
  --tables <names>   指定要初始化的表名（逗号分隔），如：user,order,product
  --force            强制重新初始化已存在的表（会删除原有元数据）
  --help             显示此帮助信息

示例:
  # 初始化所有表（默认）
  metadata-init

  # 只初始化业务表
  metadata-init --exclude-sys

  # 只初始化系统表
  metadata-init --only-sys

  # 初始化指定的表
  metadata-init --tables user,order,product

  # 强制重新初始化所有表
  metadata-init --force

注意:
  - 已存在的表默认会跳过，使用 --force 参数可强制重新初始化
  - --exclude-sys 和 --only-sys 不能同时使用
  - 指定 --tables 时会忽略 --exclude-sys 和 --only-sys
```

## 兼容性说明

### 向后兼容

✅ **完全兼容**: 原有的使用方式仍然有效
```bash
# 原有用法（不加参数）
make metadata-init
go run cmd/metadata-init/main.go
```

### 行为变更

⚠️ **注意**: 默认行为已变更

| 升级前 | 升级后 |
|-------|-------|
| 只初始化业务表 | 初始化所有表 |

**如需恢复旧行为**:
```bash
metadata-init --exclude-sys
```

## 文件变更

### 修改文件
- ✅ `cmd/metadata-init/main.go` - 添加命令行参数支持
- ✅ `docs/metadata-init-guide.md` - 更新使用文档

### 新增文件
- ✅ `docs/metadata-init-upgrade-summary.md` - 本升级总结文档

## 常见问题

### Q1: 为什么默认行为变更为初始化所有表？

**A**: 为了支持元数据驱动架构的完整性。系统表（sys_*）本身也需要元数据定义才能通过统一的 CRUD 接口进行管理。

### Q2: 如何只初始化新增的表？

**A**: 工具会自动跳过已存在的表，直接运行即可：
```bash
metadata-init
```
或者指定具体的表：
```bash
metadata-init --tables new_table1,new_table2
```

### Q3: --force 参数会影响什么？

**A**: `--force` 会删除并重新创建元数据，包括：
- sys_table 中的表记录
- sys_column 中的所有字段记录

**不会影响**:
- 实际数据库表结构
- 业务数据

### Q4: 参数冲突如何处理？

**A**: 工具会自动检测并提示：
```bash
# 错误用法
metadata-init --exclude-sys --only-sys

# 输出
错误: --exclude-sys 和 --only-sys 不能同时使用
```

### Q5: 如何验证初始化结果？

**A**: 查询 sys_table 和 sys_column：
```sql
-- 查看已初始化的表
SELECT ID, NAME, DISPLAY_NAME, CREATE_TIME
FROM sys_table
ORDER BY CREATE_TIME DESC;

-- 查看某表的字段
SELECT ID, DB_NAME, DISPLAY_NAME, COL_TYPE, SET_VALUE_TYPE
FROM sys_column
WHERE SYS_TABLE_ID = 1
ORDER BY ORDERNO;
```

## 后续建议

1. **首次使用**: 建议使用默认模式初始化所有表
   ```bash
   metadata-init
   ```

2. **日常维护**: 新增表后直接运行，会自动跳过已存在的表
   ```bash
   metadata-init
   ```

3. **表结构变更**: 使用 --force 重新初始化受影响的表
   ```bash
   metadata-init --tables affected_table1,affected_table2 --force
   ```

4. **测试环境**: 可以频繁使用 --force 重置元数据
   ```bash
   metadata-init --force
   ```

5. **生产环境**: 谨慎使用 --force，建议先备份元数据
   ```sql
   -- 备份元数据
   CREATE TABLE sys_table_backup AS SELECT * FROM sys_table;
   CREATE TABLE sys_column_backup AS SELECT * FROM sys_column;
   ```

## 总结

此次升级使 `metadata-init` 工具更加灵活和强大：

✅ **默认支持全库初始化** - 更符合元数据驱动架构的理念
✅ **灵活的过滤选项** - 可精确控制要初始化的表范围
✅ **强制重新初始化** - 支持表结构变更后的元数据更新
✅ **友好的帮助信息** - 内置完整的使用说明
✅ **向后兼容** - 原有使用方式仍然有效
✅ **更详细的输出** - 显示跳过、成功、失败的详细统计
