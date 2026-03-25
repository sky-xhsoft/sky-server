# 元数据初始化工具 - 快速参考

## 一分钟快速开始

### 运行工具

```bash
make metadata-init
```

或

```bash
go run cmd/metadata-init/main.go
```

或

```bash
./bin/metadata-init.exe  # Windows
./bin/metadata-init      # Linux/Mac
```

## 功能说明

### 自动生成的内容

✅ **sys_table** - 表定义
- 从 `information_schema.TABLES` 读取所有业务表（排除 sys_* 表）
- 自动生成表名、描述、权限掩码

✅ **sys_column** - 字段定义
- 从 `information_schema.COLUMNS` 读取所有字段
- 自动映射数据类型（MySQL → 系统类型）
- 智能推断字段属性（赋值类型、显示类型、可修改性）

✅ **sys_dict & sys_dict_item** - 数据字典
- 自动创建 YESNO 字典（是/否）

### 智能识别

| 字段特征 | 识别结果 |
|---------|---------|
| PRIMARY KEY | SET_VALUE_TYPE: pk, MODIFI_ABLE: N |
| CREATE_BY | SET_VALUE_TYPE: createBy, MODIFI_ABLE: N |
| UPDATE_BY | SET_VALUE_TYPE: operator, MODIFI_ABLE: N |
| CREATE_TIME, UPDATE_TIME | SET_VALUE_TYPE: sysdate, MODIFI_ABLE: N |
| IS_ACTIVE | SET_VALUE_TYPE: select, DISPLAY_TYPE: check |
| *_ID (非ID) | SET_VALUE_TYPE: fk, DISPLAY_TYPE: select |
| 其他字段 | SET_VALUE_TYPE: byPage, DISPLAY_TYPE: text |

## 使用示例

### 示例 1: 从零开始

```bash
# 1. 创建数据库表
mysql -u root -p skyserver < business_tables.sql

# 2. 自动生成元数据
make metadata-init

# 3. 启动服务器
make run

# 4. 测试 API
curl http://localhost:9090/api/v1/data/users
```

### 示例 2: 已有数据库

```bash
# 已有数据库，想要迁移到元数据驱动架构

# 1. 确保元数据表已创建
mysql -u root -p skyserver < sqls/create_skyserver.sql

# 2. 运行元数据初始化
make metadata-init

# 3. 所有业务表立即可用！
```

## 输出示例

```
INFO    Starting metadata initialization
INFO    Database connected successfully
INFO    Database name   {"database": "skyserver"}
INFO    Base dictionaries initialized
INFO    Found tables    {"count": 5}

INFO    Processing table        {"table": "users"}
INFO    Created table metadata  {"table": "users", "table_id": 100, "columns": 12}
INFO    Table metadata initialized      {"table": "users"}

...

INFO    Metadata initialization completed       {"success": 5, "failed": 0, "total": 5}
```

## 重要特性

### 🔄 幂等性

可以安全地多次运行：
- 已存在的表会自动跳过
- 不会重复创建
- 不会覆盖已有配置

### 💾 事务保证

每个表在一个事务中创建：
- sys_table + sys_column 在同一事务
- 任一失败则全部回滚
- 保证数据一致性

### 🎯 智能推断

自动识别：
- 主键字段 (pk)
- 审计字段 (createBy, operator, sysdate)
- 外键字段 (fk)
- 特殊字段 (IS_ACTIVE → select + check)

## 常见问题

### Q: 如何重新初始化某个表？

```sql
-- 1. 删除已有元数据
DELETE FROM sys_column WHERE SYS_TABLE_ID IN (
  SELECT ID FROM sys_table WHERE DB_NAME = 'USERS'
);
DELETE FROM sys_table WHERE DB_NAME = 'USERS';

-- 2. 重新运行工具
make metadata-init
```

### Q: 哪些表会被导入？

- ✅ 所有业务表
- ❌ sys_* 开头的元数据表（自动排除）
- ❌ 已存在于 sys_table 的表（自动跳过）

### Q: 生成的元数据需要手动调整吗？

建议手动调整：
- 表的 SYS_TABLE_CATEGORY_ID（分类）
- 外键的 REF_TABLE_ID 和 REF_COLUMN_ID
- 字段的 ORDERNO（显示顺序）
- 枚举字段的 SYS_DICT_ID

## 下一步

### 初始化后的工作

1. **调整表属性**
   ```sql
   -- 设置表分类
   UPDATE sys_table SET SYS_TABLE_CATEGORY_ID = 2 WHERE DB_NAME = 'ORDERS';

   -- 设置主从关系
   UPDATE sys_table SET PARENT_TABLE_ID = 100 WHERE DB_NAME = 'ORDER_ITEMS';
   ```

2. **配置外键引用**
   ```sql
   -- 为 CUSTOMER_ID 字段设置引用
   UPDATE sys_column
   SET REF_TABLE_ID = 101, REF_COLUMN_ID = 1001
   WHERE SYS_TABLE_ID = 100 AND DB_NAME = 'CUSTOMER_ID';
   ```

3. **创建业务字典**
   ```sql
   -- 创建订单状态字典
   INSERT INTO sys_dict (CODE, NAME, IS_ACTIVE, CREATE_BY, SYS_COMPANY_ID)
   VALUES ('ORDER_STATUS', '订单状态', 'Y', 'admin', 1);

   -- 添加字典项
   INSERT INTO sys_dict_item (SYS_DICT_ID, CODE, NAME, ORDERNO, IS_ACTIVE)
   VALUES
     (10, 'PENDING', '待处理', 10, 'Y'),
     (10, 'PROCESSING', '处理中', 20, 'Y'),
     (10, 'COMPLETED', '已完成', 30, 'Y');
   ```

## 文档链接

- 📖 [详细使用指南](./metadata-init-guide.md)
- 📊 [实现总结](./metadata-init-tool-summary.md)
- 🏗️ [系统架构](./01-系统架构设计.md)

## 技术支持

- GitHub Issues: [报告问题](https://github.com/sky-xhsoft/sky-server/issues)
- 文档: [docs 目录](../docs/)

---

**提示**: 这个工具可以节省 90% 的元数据配置时间，让你专注于业务逻辑！✅
