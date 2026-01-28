# Sky-Server 测试目录

> **版本**: v1.0
> **最后更新**: 2026-01-28
> **维护者**: Sky Team

---

## 📁 目录结构

```
test/
├── README.md             # 本文档
├── TEST_RESULTS.md       # 测试结果文档
└── unit/                 # 单元测试
    ├── mask/             # MASK 权限系统测试
    │   └── mask_test.go
    ├── permission/       # 权限系统测试
    │   └── permission_test.go
    └── transaction/      # 事务管理测试
        └── transaction_test.go
```

---

## 🧪 测试类型

### 1. 单元测试 (Unit Tests)

**目录**: `test/unit/`

**框架**: Go 标准测试库 (`testing`)

**运行命令**:

```bash
# 运行所有单元测试
go test ./test/unit/...

# 运行特定模块的测试
go test ./test/unit/mask
go test ./test/unit/permission
go test ./test/unit/transaction

# 运行测试并显示详细输出
go test -v ./test/unit/...

# 运行测试并生成覆盖率报告
go test -cover ./test/unit/...

# 生成覆盖率 HTML 报告
go test -coverprofile=coverage.out ./test/unit/...
go tool cover -html=coverage.out -o coverage.html
```

**测试模块说明**:

#### MASK 权限系统测试 (`test/unit/mask/`)

测试 MASK 字段权限系统的功能：
- MASK 值解析和验证
- 权限位检查（A, M, D, Q, S, U, V）
- 权限组合和继承

#### 权限系统测试 (`test/unit/permission/`)

测试权限管理系统的功能：
- 用户权限验证
- 角色权限管理
- 权限组功能
- 数据访问控制

#### 事务管理测试 (`test/unit/transaction/`)

测试事务管理功能：
- 事务开启和提交
- 事务回滚
- 嵌套事务处理
- 事务钩子函数

---

## 📝 编写测试

### 测试文件命名规范

- 测试文件命名: `*_test.go`
- 测试函数命名: `Test<FunctionName>`
- 基准测试命名: `Benchmark<FunctionName>`
- 示例函数命名: `Example<FunctionName>`

### 单元测试示例

```go
package mask

import (
    "testing"
)

func TestMaskParse(t *testing.T) {
    tests := []struct {
        name     string
        mask     string
        expected int
        wantErr  bool
    }{
        {
            name:     "valid mask",
            mask:     "AMDQ",
            expected: 15,
            wantErr:  false,
        },
        {
            name:     "invalid mask",
            mask:     "XYZ",
            expected: 0,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParseMask(tt.mask)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseMask() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("ParseMask() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### 表格驱动测试

Go 推荐使用表格驱动测试（Table-Driven Tests）:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // 测试用例
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

---

## 🔧 测试工具

### 常用测试命令

```bash
# 运行测试
go test ./test/unit/...

# 详细输出
go test -v ./test/unit/...

# 覆盖率
go test -cover ./test/unit/...

# 覆盖率详情
go test -coverprofile=coverage.out ./test/unit/...
go tool cover -func=coverage.out

# HTML 覆盖率报告
go tool cover -html=coverage.out

# 运行基准测试
go test -bench=. ./test/unit/...

# 运行特定测试
go test -run TestMaskParse ./test/unit/mask

# 并行测试
go test -parallel 4 ./test/unit/...

# 超时设置
go test -timeout 30s ./test/unit/...
```

### Mock 和 Stub

对于需要 mock 的测试，推荐使用以下工具：

- [gomock](https://github.com/golang/mock) - 官方 mock 框架
- [testify/mock](https://github.com/stretchr/testify) - 流行的测试工具包
- [sqlmock](https://github.com/DATA-DOG/go-sqlmock) - 数据库 mock

---

## 📊 测试覆盖率

### 生成覆盖率报告

```bash
# 生成覆盖率文件
go test -coverprofile=coverage.out ./test/unit/...

# 查看覆盖率统计
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

### 覆盖率目标

- **核心模块**: 目标覆盖率 ≥ 80%
- **工具函数**: 目标覆盖率 ≥ 90%
- **业务逻辑**: 目标覆盖率 ≥ 70%

---

## 🎯 测试最佳实践

### 1. 测试独立性

每个测试应该独立运行，不依赖其他测试的执行顺序或结果。

```go
func TestExample(t *testing.T) {
    // 每个测试都应该有自己的 setup
    setup := createTestSetup()
    defer cleanup(setup)

    // 测试逻辑
}
```

### 2. 使用子测试

使用 `t.Run()` 创建子测试，提高测试的组织性：

```go
func TestUser(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // 测试创建用户
    })

    t.Run("Update", func(t *testing.T) {
        // 测试更新用户
    })

    t.Run("Delete", func(t *testing.T) {
        // 测试删除用户
    })
}
```

### 3. 测试边界条件

不仅测试正常情况，也要测试边界和异常情况：

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    int
        want    int
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"zero divisor", 10, 0, 0, true},  // 边界条件
        {"negative", -10, 2, -5, false},
    }
    // ...
}
```

### 4. 清晰的错误消息

使用清晰的错误消息，方便定位问题：

```go
if got != want {
    t.Errorf("Function(%v) = %v, want %v", input, got, want)
}
```

### 5. 使用 Helper 函数

对于重复的测试逻辑，提取为 helper 函数：

```go
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()  // 标记为 helper 函数
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

---

## 🔍 持续集成

### GitHub Actions 配置示例

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Run tests
        run: go test -v -cover ./test/unit/...
      - name: Generate coverage
        run: |
          go test -coverprofile=coverage.out ./test/unit/...
          go tool cover -html=coverage.out -o coverage.html
      - name: Upload coverage
        uses: actions/upload-artifact@v2
        with:
          name: coverage
          path: coverage.html
```

---

## 📚 相关文档

- [Go Testing 官方文档](https://golang.org/pkg/testing/)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)
- [测试结果文档](./TEST_RESULTS.md)
- [项目文档索引](../docs/INDEX.md)

---

## 📞 联系方式

- **项目地址**: https://github.com/sky-xhsoft/sky-server
- **维护团队**: Sky Team
- **更新日期**: 2026-01-28

---

**文档版本**: v1.0
**最后更新**: 2026-01-28
**维护者**: Sky Team
