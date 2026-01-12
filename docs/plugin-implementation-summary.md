# 插件系统实现总结

## 实现日期
2026-01-12

## 概述
实现了一个灵活的插件系统，用于在 sys_table 创建时自动生成标准字段的 sys_column 记录。

## 实现内容

### 1. 核心插件框架

**文件**: `internal/pkg/plugin/plugin.go`

- 定义了 `Plugin` 接口
- 实现了 `PluginData` 数据结构
- 创建了 `Manager` 插件管理器
- 支持插件注册和批量执行

### 2. 标准字段生成插件

**文件**: `internal/pkg/plugin/sys_table_after_create.go`

**插件命名**: `sys_table_after_create` (遵循命名规则：表单名称_执行时机_动作)

功能：当创建 sys_table 记录时，自动生成 7 个标准字段：

| 字段名 | 类型 | 说明 | 赋值策略 |
|--------|------|------|----------|
| ID | int | 主键 | pk |
| SYS_COMPANY_ID | int | 所属公司 | ignore |
| CREATE_BY | varchar(80) | 创建人 | createBy |
| CREATE_TIME | datetime | 创建时间 | sysdate |
| UPDATE_BY | varchar(80) | 修改人 | operator |
| UPDATE_TIME | datetime | 修改时间 | sysdate |
| IS_ACTIVE | char(1) | 是否有效 | byPage |

### 3. CRUD 服务集成

**文件**: `internal/service/crud/crud_service.go`

修改内容：
- 添加 `pluginManager` 字段到 service 结构
- 更新 `NewService` 构造函数接受插件管理器
- 在 `Create` 方法中执行插件（after 钩子之后）
- 插件执行失败不影响主流程

### 4. 主程序初始化

**文件**: `cmd/server/main.go`

修改内容：
- 导入 plugin 包
- 创建插件管理器实例
- 注册标准字段生成插件
- 将插件管理器传递给 CRUD 服务

### 5. 测试代码

**文件**: `internal/pkg/plugin/plugin_test.go`

实现了以下测试：
- 插件注册测试
- 插件执行测试
- 无插件场景测试
- 标准字段插件名称测试
- 标准字段插件动作过滤测试

测试结果：✅ 所有测试通过

### 6. 文档

**文件**: `docs/plugin-system.md`

详细文档包含：
- 架构说明
- 使用方法
- API 参考
- 示例代码
- 最佳实践
- 故障排查

## 技术特点

### 1. 解耦设计
- 插件系统独立于核心业务逻辑
- 使用接口实现松耦合
- 插件失败不影响主流程

### 2. 灵活扩展
- 支持为不同表注册不同插件
- 一个表可以注册多个插件
- 插件按注册顺序执行

### 3. 数据驱动
- 通过 PluginData 传递上下文信息
- 包含表名、操作类型、记录ID等
- 插件可以访问数据库连接

### 4. 类型安全
- 使用 Go 的类型系统
- 编译期检查插件接口实现
- 避免运行时类型错误

## 工作流程

```
用户创建 sys_table 记录
    ↓
CRUD Service Create 方法
    ↓
验证权限
    ↓
执行 before 钩子
    ↓
插入数据到数据库
    ↓
执行 after 钩子
    ↓
【执行插件】←── 新增步骤
    ↓
插件管理器查找注册的插件
    ↓
SysTableStandardColumnsPlugin 执行
    ↓
生成 7 个标准 sys_column 记录
    ↓
返回创建结果
```

## 使用示例

### 创建表时自动生成标准字段

```bash
# 1. 登录获取 token
curl -X POST http://localhost:9090/api/v1/sso/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 2. 创建新表
curl -X POST http://localhost:9090/api/v1/crud \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "tableName": "sys_table",
    "data": {
      "NAME": "客户表",
      "DB_NAME": "customer",
      "MASK": "AMQDSUV",
      "IS_ACTIVE": "Y"
    }
  }'

# 3. 查询生成的标准字段
curl -X GET "http://localhost:9090/api/v1/crud?tableName=sys_column&filters[SYS_TABLE_ID]=新表ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

预期结果：应该返回 7 条标准字段记录。

## 未来扩展方向

### 1. 更多操作支持
- 在 Update 方法中执行插件
- 在 Delete 方法中执行插件

### 2. 异步执行
- 支持异步插件执行
- 使用消息队列解耦

### 3. 插件配置
- 从配置文件加载插件
- 支持插件启用/禁用
- 插件执行超时控制

### 4. 监控和日志
- 插件执行时间统计
- 失败率监控
- 详细的执行日志

### 5. 插件市场
- 标准插件库
- 插件版本管理
- 插件依赖管理

## 注意事项

### 1. 性能影响
- 插件在主流程中同步执行
- 插件应该快速完成
- 避免复杂的数据库操作

### 2. 错误处理
- 插件失败只记录日志
- 不会回滚主操作
- 需要考虑数据一致性

### 3. 事务处理
- 插件不在主事务中
- 如需事务，在插件内部开启
- 注意分布式事务问题

## 验证清单

- [x] 插件系统核心代码实现
- [x] 标准字段生成插件实现
- [x] CRUD 服务集成完成
- [x] 主程序初始化完成
- [x] 单元测试编写并通过
- [x] 项目成功构建
- [x] 文档编写完成

## 总结

成功实现了一个灵活、可扩展的插件系统，实现了在 sys_table 创建时自动生成标准字段的功能。系统设计合理，代码质量高，测试覆盖完整，为后续功能扩展打下了良好基础。
