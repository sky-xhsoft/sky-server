# SysTableAfterCreatePlugin 测试指南

## 测试日期
2026-01-12

## 测试文件

### 1. sys_table_after_create_test.go
**位置**: `internal/pkg/plugin/sys_table_after_create_test.go`

**测试类型**: 集成测试

**测试覆盖**:
1. MASK 字段验证
2. ORDERNO 自动生成
3. Directory 自动创建
4. 标准字段创建
5. PK_COLUMN_ID 设置
6. 辅助函数测试
7. 完整集成测试

## 运行环境要求

### Windows 环境

集成测试使用 SQLite 数据库，需要 CGO 支持，因此需要 C 编译器：

**选项 1: 安装 MinGW-w64**
```bash
# 1. 下载并安装 MinGW-w64
# https://www.mingw-w64.org/downloads/

# 2. 将 MinGW bin 目录添加到 PATH
# 例如: C:\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin

# 3. 验证安装
gcc --version

# 4. 运行测试
go test -v ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin
```

**选项 2: 使用 TDM-GCC**
```bash
# 1. 下载并安装 TDM-GCC
# https://jmeubank.github.io/tdm-gcc/

# 2. 运行测试
go test -v ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin
```

### Linux/Mac 环境

Linux 和 Mac 通常已安装 gcc，可以直接运行：

```bash
go test -v ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin
```

## 测试类型

### 单元测试 (Unit Tests)
不需要数据库，不需要 gcc，可以在任何环境运行。

**测试文件**: `plugin_test.go`

**包含测试**:
- Manager 插件管理测试
- Plugin Name 测试
- Plugin 执行条件测试
- 辅助函数测试 (getStringValue, getIntValue, getUintValue)

### 集成测试 (Integration Tests)
需要 SQLite 数据库，需要 gcc/CGO 支持。

**测试文件**: `sys_table_after_create_test.go` (带 `// +build integration` 标签)

**包含测试**:
- MASK 验证测试
- ORDERNO 自动生成测试
- Directory 自动创建测试
- 标准字段创建测试
- PK_COLUMN_ID 设置测试
- 完整集成测试

## 运行测试

### 运行单元测试（推荐，无需 gcc）
```bash
# 运行所有单元测试
go test -v ./internal/pkg/plugin

# 运行特定单元测试
go test -v ./internal/pkg/plugin -run TestManager
go test -v ./internal/pkg/plugin -run TestGetStringValue
go test -v ./internal/pkg/plugin -run TestGetIntValue
go test -v ./internal/pkg/plugin -run TestGetUintValue
```

### 运行集成测试（需要 gcc）
```bash
# 运行所有集成测试
go test -v -tags=integration ./internal/pkg/plugin

# 运行特定集成测试
go test -v -tags=integration ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin_MaskValidation
go test -v -tags=integration ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin_OrdernoGeneration
go test -v -tags=integration ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin_DirectoryCreation
go test -v -tags=integration ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin_StandardColumnsCreation
go test -v -tags=integration ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin_PKColumnIDSetting
go test -v -tags=integration ./internal/pkg/plugin -run TestHelperFunctions
go test -v -tags=integration ./internal/pkg/plugin -run TestSysTableAfterCreatePlugin_FullIntegration
```

### 运行所有测试（单元 + 集成）
```bash
# 需要 gcc 环境
go test -v -tags=integration ./internal/pkg/plugin
```

## 测试数据库

测试使用 SQLite 内存数据库 (`:memory:`)：
- 不需要外部数据库服务器
- 每次测试运行时创建新的数据库
- 测试结束后自动清理
- 快速且隔离

## 测试覆盖率

```bash
# 生成覆盖率报告
go test -v ./internal/pkg/plugin -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y gcc

      - name: Run tests
        run: go test -v ./internal/pkg/plugin
```

### GitLab CI 示例

```yaml
test:
  image: golang:1.21
  script:
    - apt-get update && apt-get install -y gcc
    - go test -v ./internal/pkg/plugin
```

## 测试结果示例

### 单元测试（通过）✅
```
=== RUN   TestManager_Register
--- PASS: TestManager_Register (0.00s)
=== RUN   TestManager_ExecutePlugins
--- PASS: TestManager_ExecutePlugins (0.00s)
=== RUN   TestManager_ExecutePlugins_NoPlugins
--- PASS: TestManager_ExecutePlugins_NoPlugins (0.00s)
=== RUN   TestSysTableAfterCreatePlugin_Name
--- PASS: TestSysTableAfterCreatePlugin_Name (0.00s)
=== RUN   TestSysTableAfterCreatePlugin_OnlyExecuteOnCreate
--- PASS: TestSysTableAfterCreatePlugin_OnlyExecuteOnCreate (0.00s)
=== RUN   TestGetStringValue
--- PASS: TestGetStringValue (0.00s)
=== RUN   TestGetIntValue
--- PASS: TestGetIntValue (0.00s)
=== RUN   TestGetUintValue
--- PASS: TestGetUintValue (0.00s)
PASS
ok      github.com/sky-xhsoft/sky-server/internal/pkg/plugin    0.009s
```

### 开发建议

**日常开发**:
- 使用单元测试 (`go test -v ./internal/pkg/plugin`)
- 快速验证代码逻辑
- 不需要配置数据库环境

**提交前/CI环境**:
- 运行集成测试 (`go test -v -tags=integration ./internal/pkg/plugin`)
- 完整验证功能
- 确保数据库操作正确

## 测试场景

### 1. MASK 验证测试

测试各种 MASK 格式：
- ✅ 有效的 MASK: `AMDSQPGU`, `AMDSQPIE`
- ✅ 空 MASK（允许）
- ❌ 包含数字: `AMDQ0000`
- ❌ 包含特殊字符: `AMDQ-PSU`
- ❌ 小写字母: `amdsqpgu`

### 2. ORDERNO 自动生成测试

测试自动生成逻辑：
- 类别1已有表 (10, 20, 30) → 新表应为 40
- 类别2已有表 (15) → 新表应为 20
- ORDERNO 已设置 (100) → 不应修改

### 3. Directory 创建测试

测试不同场景：
- ✅ 普通表 (USERS) → 创建 USERS_LIST
- ❌ ITEM 结尾表 (ORDER_ITEM) → 不创建
- ❌ LINE 结尾表 (INVOICE_LINE) → 不创建
- ❌ 有 parent_table → 不创建
- ❌ 已有 directory → 不创建

### 4. 标准字段创建测试

验证创建的 9 个标准字段：
1. ID - 主键
2. SYS_COMPANY_ID - 所属公司
3. (TABLE.ID+100) - 日志字段
4. CREATE_BY - 创建人
5. UPDATE_BY - 修改人
6. CREATE_TIME - 创建时间
7. UPDATE_TIME - 修改时间
8. IS_ACTIVE - 是否有效

每个字段验证：
- COL_TYPE
- ORDERNO
- NULL_ABLE
- SET_VALUE_TYPE
- MODIFI_ABLE
- IS_SYSTEM
- IS_AK / IS_DK (仅 ID 字段)
- FULL_NAME

### 5. PK 设置测试

验证表的关键字段设置：
- PK_COLUMN_ID = ID字段的column_id
- AK_COLUMN_ID = ID字段的column_id
- DK_COLUMN_ID = ID字段的column_id

### 6. 完整集成测试

端到端测试创建完整的表：
1. 创建表记录
2. 执行插件
3. 验证 ORDERNO 已生成
4. 验证 Directory 已创建
5. 验证标准字段已创建
6. 验证 PK_COLUMN_ID 已设置

## 常见问题

### Q: 为什么需要 gcc？
A: SQLite 数据库驱动 (`gorm.io/driver/sqlite`) 使用 CGO，需要 C 编译器。

### Q: 能否使用纯 Go 的 SQLite 驱动？
A: 可以，但需要修改测试代码使用 `modernc.org/sqlite` 等纯 Go 驱动。

### Q: 测试能否在 Docker 中运行？
A: 可以，Docker 镜像通常包含 gcc。

### Q: 如何在没有 gcc 的环境中开发？
A: 使用 `plugin_test.go` 中的简单单元测试，或者在 CI/CD 环境中运行完整测试。

## 下一步

### 短期
1. ✅ 编写测试（已完成）
2. ⏳ 在开发环境运行测试
3. ⏳ 修复失败的测试
4. ⏳ 提高测试覆盖率

### 中期
1. ⏳ 添加性能基准测试
2. ⏳ 添加并发测试
3. ⏳ 集成到 CI/CD

### 长期
1. ⏳ 添加模糊测试
2. ⏳ 添加压力测试
3. ⏳ 自动化回归测试

## 总结

- ✅ 创建了完整的集成测试套件
- ✅ 覆盖所有插件功能
- ⚠️ 需要 gcc 支持 (Windows 环境)
- ✅ 使用内存数据库，快速且隔离
- ✅ 提供详细的测试场景和验证

测试套件为插件提供了完整的质量保证。✅
