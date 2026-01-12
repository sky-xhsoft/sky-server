# SysTableAfterCreatePlugin 升级说明

## 升级日期
2026-01-12

## 背景

参考 Oracle 存储过程 `AD_TABLE_AC` 的实现，对 `SysTableAfterCreatePlugin` 进行了全面升级，使其功能更加完善和强大。

## 主要改进

### 1. MASK 字段验证 ✅

**功能**：验证表的 MASK 字段格式是否正确

**规则**：
- MASK 必须是 8 位字符
- 每位字符必须是 `AMDSQPGU` 中的一个

**MASK 字符含义**：
- `A` - Add (新增)
- `M` - Modify (修改)
- `D` - Delete (删除)
- `S` - Submit (提交)
- `Q` - Query (查询)
- `P` - Print (打印)
- `G` - Grant (授权)
- `U` - Unsubmit (反提交)

**示例**：
```go
// 有效的 MASK
"AMDSQPGU" ✅
"AMDQPSGU" ✅
"AAAA0000" ❌ (包含非法字符 '0')
"AMDQPS"   ❌ (长度不是8位)
```

### 2. 自动生成 ORDERNO ✅

**功能**：如果表的 ORDERNO 未设置，自动生成一个有序的编号

**算法**：
```go
// 获取同类别表的最大 ORDERNO
maxOrderno := SELECT MAX(ORDERNO) FROM sys_table
             WHERE SYS_TABLE_CATEGORY_ID = categoryID

// 按 10 递增
newOrderno = ((maxOrderno / 10) + 1) * 10
```

**示例**：
```
已有表的 ORDERNO: 10, 20, 30
新表的 ORDERNO: 40

已有表的 ORDERNO: 15, 23, 37
新表的 ORDERNO: 40
```

### 3. 自动创建 Directory ✅

**功能**：为符合条件的表自动创建对应的安全目录

**创建条件**：
- 表名不以 `ITEM` 结尾
- 表名不以 `LINE` 结尾
- 表的 `PARENT_TABLE_ID` 为空（不是子表）
- 表的 `SYS_DIRECTORY_ID` 为空（未手动指定目录）

**Directory 命名**：`<TABLE_NAME>_LIST`

**示例**：
```go
// 会创建 directory
表名: USERS         → USERS_LIST 目录
表名: PRODUCTS      → PRODUCTS_LIST 目录

// 不会创建 directory
表名: ORDER_ITEM    (以 ITEM 结尾)
表名: INVOICE_LINE  (以 LINE 结尾)
表名: ORDER_DETAIL  (PARENT_TABLE_ID 不为空)
```

**验证规则**：
- 如果表既没有 directory 也没有 parent_table，且需要创建 directory，则报错
- 确保每个主表都有对应的安全目录

### 4. 完善的标准字段 ✅

参考 Oracle 存储过程，创建以下标准字段：

#### 字段列表

| 序号 | 字段名 | 说明 | ORDERNO | MASK | SET_VALUE_TYPE |
|------|--------|------|---------|------|----------------|
| 1 | ID | 主键 | 1 | 0000000000 | pk |
| 2 | SYS_COMPANY_ID | 所属公司 | 2 | 0000000000 | object |
| 3 | (表名.ID+100) | 日志字段 | 1000 | 0010011001 | trigger |
| 4 | CREATE_BY | 创建人 | 1001 | 0010001001 | createBy |
| 5 | UPDATE_BY | 修改人 | 1002 | 0010101101 | operator |
| 6 | CREATE_TIME | 创建时间 | 1003 | 0010001001 | sysdate |
| 7 | UPDATE_TIME | 修改时间 | 1004 | 0010101101 | sysdate |
| 8 | IS_ACTIVE | 是否有效 | 10000 | 0011101000 | select |

#### 字段属性详解

**1. ID 字段**
```go
{
    "NAME":           "TABLE_NAME.ID",
    "DB_NAME":        "ID",
    "DESCRIPTION":    "主键",
    "COL_TYPE":       "int",
    "ORDERNO":        1,
    "NULL_ABLE":      "N",         // 不允许为空
    "MASK":           "0000000000", // 10位全0：不可见不可编辑
    "SET_VALUE_TYPE": "pk",         // 主键自动生成
    "MODIFI_ABLE":    "N",          // 不可修改
    "IS_SYSTEM":      "Y",          // 系统字段
    "IS_AK":          "Y",          // 是主键
    "IS_DK":          "Y",          // 是显示键
}
```

**2. SYS_COMPANY_ID 字段**
```go
{
    "NAME":           "TABLE_NAME.SYS_COMPANY_ID",
    "DB_NAME":        "SYS_COMPANY_ID",
    "DESCRIPTION":    "所属公司",
    "COL_TYPE":       "int",
    "ORDERNO":        2,
    "NULL_ABLE":      "Y",
    "MASK":           "0000000000",
    "SET_VALUE_TYPE": "object",     // 从上下文对象获取
    "MODIFI_ABLE":    "N",
    "IS_SYSTEM":      "Y",
}
```

**3. 日志字段 (虚拟字段)**
```go
{
    "NAME":           "TABLE_NAME.(TABLE_NAME.ID+100)",
    "DB_NAME":        "(TABLE_NAME.ID+100)",
    "DESCRIPTION":    "日志",
    "COL_TYPE":       "varchar",
    "COL_LENGTH":     30,
    "ORDERNO":        1000,
    "MASK":           "0010011001",
    "SET_VALUE_TYPE": "trigger",    // 触发器赋值
    "DISPLAY_TYPE":   "hr",         // 水平线分隔符
    "IS_SHOW_TITLE":  "N",          // 不显示标题
}
```

**4. CREATE_BY 创建人**
```go
{
    "NAME":           "TABLE_NAME.CREATE_BY",
    "DB_NAME":        "CREATE_BY",
    "DESCRIPTION":    "创建人",
    "COL_TYPE":       "varchar",
    "COL_LENGTH":     80,
    "ORDERNO":        1001,
    "MASK":           "0010001001",
    "SET_VALUE_TYPE": "createBy",   // 创建人赋值
    "MODIFI_ABLE":    "N",
}
```

**5. UPDATE_BY 修改人**
```go
{
    "NAME":           "TABLE_NAME.UPDATE_BY",
    "DB_NAME":        "UPDATE_BY",
    "DESCRIPTION":    "修改人",
    "COL_TYPE":       "varchar",
    "COL_LENGTH":     80,
    "ORDERNO":        1002,
    "MASK":           "0010101101",
    "SET_VALUE_TYPE": "operator",   // 操作人赋值
    "MODIFI_ABLE":    "N",
}
```

**6. CREATE_TIME 创建时间**
```go
{
    "NAME":           "TABLE_NAME.CREATE_TIME",
    "DB_NAME":        "CREATE_TIME",
    "DESCRIPTION":    "创建时间",
    "COL_TYPE":       "datetime",
    "ORDERNO":        1003,
    "MASK":           "0010001001",
    "SET_VALUE_TYPE": "sysdate",    // 系统时间
    "DISPLAY_TYPE":   "datetime",
}
```

**7. UPDATE_TIME 修改时间**
```go
{
    "NAME":           "TABLE_NAME.UPDATE_TIME",
    "DB_NAME":        "UPDATE_TIME",
    "DESCRIPTION":    "修改时间",
    "COL_TYPE":       "datetime",
    "ORDERNO":        1004,
    "MASK":           "0010101101",
    "SET_VALUE_TYPE": "sysdate",
    "DISPLAY_TYPE":   "datetime",
}
```

**8. IS_ACTIVE 是否有效**
```go
{
    "NAME":           "TABLE_NAME.IS_ACTIVE",
    "DB_NAME":        "IS_ACTIVE",
    "DESCRIPTION":    "可用",
    "COL_TYPE":       "char",
    "COL_LENGTH":     1,
    "ORDERNO":        10000,
    "NULL_ABLE":      "N",
    "MASK":           "0011101000",
    "SET_VALUE_TYPE": "select",     // 下拉选择
    "DEFAULT_VALUE":  "Y",
    "MODIFI_ABLE":    "Y",          // 可以修改
    "DISPLAY_TYPE":   "check",      // 复选框
}
```

### 5. 设置表的 PK_COLUMN_ID ✅

**功能**：创建完标准字段后，自动设置表的主键字段引用

**更新字段**：
```go
sys_table.PK_COLUMN_ID = ID字段的column_id
sys_table.AK_COLUMN_ID = ID字段的column_id  // 备用键
sys_table.DK_COLUMN_ID = ID字段的column_id  // 显示键
```

## MASK 字段详解

### MASK 格式

MASK 是一个 10 位字符串，每一位控制字段在特定操作下的可见性和可编辑性。

### 位置含义

| 位置 | 操作 | 说明 |
|------|------|------|
| 1 | Add-V | 新增时是否可见 |
| 2 | Add-E | 新增时是否可编辑 |
| 3 | Modify-V | 修改时是否可见 |
| 4 | Modify-E | 修改时是否可编辑 |
| 5 | Delete-V | 删除时是否可见 |
| 6 | Delete-E | 删除时是否可编辑 |
| 7 | Query-V | 查询时是否可见 |
| 8 | Query-E | 查询时是否可编辑 |
| 9-10 | 保留 | 未来扩展 |

### 值含义

- `0` - 不可见或不可编辑
- `1` - 可见或可编辑

### MASK 示例

```
"0000000000" - 所有操作都不可见不可编辑（系统字段）
"0010001001" - 新增不可见，修改可见不可编辑，查询可见不可编辑
"0010101101" - 新增不可见，修改可见可编辑，查询可见不可编辑
"0011101000" - 新增不可见，修改和删除可见可编辑，查询可见不可编辑
"1111111111" - 所有操作都可见可编辑
```

## 使用示例

### 创建普通表

```bash
curl -X POST http://localhost:9090/api/v1/data/sys_table \
  -H "Content-Type: application/json" \
  -d '{
    "DB_NAME": "USERS",
    "NAME": "用户表",
    "DESCRIPTION": "系统用户信息",
    "MASK": "AMDSQPGU",
    "SYS_TABLE_CATEGORY_ID": 1
  }'
```

**结果**：
- ✅ 自动验证 MASK 格式
- ✅ 自动生成 ORDERNO
- ✅ 自动创建 USERS_LIST 目录
- ✅ 自动创建 9 个标准字段
- ✅ 自动设置 PK_COLUMN_ID

### 创建子表

```bash
curl -X POST http://localhost:9090/api/v1/data/sys_table \
  -H "Content-Type: application/json" \
  -d '{
    "DB_NAME": "ORDER_ITEM",
    "NAME": "订单明细",
    "DESCRIPTION": "订单明细表",
    "MASK": "AMDSQPGU",
    "PARENT_TABLE_ID": 123,
    "SYS_TABLE_CATEGORY_ID": 1
  }'
```

**结果**：
- ✅ 不会创建 directory（因为表名以 ITEM 结尾）
- ✅ 使用父表的 directory
- ✅ 创建标准字段

## 辅助函数

插件新增了3个辅助函数：

### getStringValue
```go
func getStringValue(m map[string]interface{}, key string) string
```
安全地从 map 中获取字符串值

### getIntValue
```go
func getIntValue(m map[string]interface{}, key string) int
```
安全地从 map 中获取 int 值，支持多种数值类型转换

### getUintValue
```go
func getUintValue(m map[string]interface{}, key string) uint
```
安全地从 map 中获取 uint 值，支持多种数值类型转换

## 与 Oracle 存储过程的对应关系

| Oracle 存储过程功能 | Go 插件实现 | 状态 |
|---------------------|-------------|------|
| 验证 MASK 格式 | ✅ 已实现 | 完成 |
| 自动生成 orderno | ✅ 已实现 | 完成 |
| 创建 directory | ✅ 已实现 | 完成 |
| 创建 ad_tablesql | ❌ 未实现 | 架构不同 |
| 执行 DDL 创建表 | ❌ 未实现 | 需单独工具 |
| 创建标准字段 | ✅ 已实现 | 完成 |
| 设置 PK_COLUMN_ID | ✅ 已实现 | 完成 |
| 设置外键引用 | ⏳ 部分实现 | 待完善 |

## 未来改进

### 1. 外键引用
为 SYS_COMPANY_ID 等字段设置 REF_TABLE_ID 和 REF_COLUMN_ID：
```go
"REF_TABLE_ID":   <sys_company表的ID>,
"REF_COLUMN_ID":  <sys_company.ID字段的ID>,
```

### 2. 字典组引用
为 IS_ACTIVE 字段设置字典组：
```go
"SYS_DICT_ID": <YESNO字典组的ID>,
```

### 3. DDL 执行
实现自动创建物理表的功能

### 4. SYS_ORG_ID 字段
根据需求决定是否添加组织字段

## 测试建议

### 单元测试

```go
func TestSysTableAfterCreate_MaskValidation(t *testing.T) {
    // 测试 MASK 验证
    // 有效 MASK: "AMDSQPGU" ✅
    // 无效 MASK: "INVALID0" ❌
}

func TestSysTableAfterCreate_OrdernoGeneration(t *testing.T) {
    // 测试 ORDERNO 自动生成
}

func TestSysTableAfterCreate_DirectoryCreation(t *testing.T) {
    // 测试 directory 自动创建
}

func TestSysTableAfterCreate_StandardColumns(t *testing.T) {
    // 测试标准字段创建
    // 验证字段数量、属性、MASK等
}
```

## 总结

### 升级前 ❌
- 只创建基本的7个标准字段
- 没有 MASK 验证
- 没有 ORDERNO 自动生成
- 没有 directory 自动创建
- 字段属性不完整

### 升级后 ✅
- 创建完整的9个标准字段（包括日志字段）
- ✅ MASK 字段验证
- ✅ ORDERNO 自动生成
- ✅ Directory 自动创建
- ✅ 完整的字段属性（MASK, SET_VALUE_TYPE, IS_SYSTEM等）
- ✅ 设置表的 PK_COLUMN_ID
- ✅ 辅助函数支持

### 关键改进
1. **功能完整性** - 参考成熟的 Oracle 存储过程实现
2. **自动化程度** - 减少手动配置工作
3. **数据一致性** - 自动验证和设置
4. **可维护性** - 清晰的代码结构和注释

这次升级使插件更加强大和完善，为元数据驱动系统提供了坚实的基础。✅
