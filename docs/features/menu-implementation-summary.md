# 菜单系统实现总结

## 概述

本次实现将原有的 sys_menu 单独菜单表改造为基于三级目录的元数据驱动菜单系统：

**sys_subsystem** (一级) → **sys_table_category** (二级) → **sys_table (IS_MENU='Y')** (三级)

## 数据库变更

### 1. init.sql 更新

#### sys_table_category 表修改
- ✅ 添加 `SYS_SUBSYSTEM_ID` 字段，关联到 sys_subsystem
- ✅ 添加 `ICON` 字段，存储分类图标
- ✅ 添加索引 `idx_subsystem` 在 SYS_SUBSYSTEM_ID

#### sys_table 表修改
- ✅ 添加 `ORDERNO` 字段（**本次新增**），用于菜单排序
- ✅ 已有 `ICO_IMG` 字段，存储表单图标
- ✅ 已有 `SYS_DIRECTORY_ID` 字段，关联安全目录
- ✅ 已有 `SYS_PARENT_TABLE_ID` 字段，关联父表
- ✅ 已有 `ROWCNT` 字段，统计行数
- ✅ 已有 `DESCRIPTION` 字段，备注说明

#### 删除 sys_menu 表
- ✅ 从 init.sql 中完全移除 sys_menu 表定义（原第 646-813 行）

### 2. 迁移脚本

为现有数据库提供了两个迁移脚本：

#### migration_name_to_display_name.sql
```bash
mysql -u root -p skyserver < sqls/migration_name_to_display_name.sql
```
- 将 sys_column 表的 NAME 字段重命名为 DISPLAY_NAME

#### migration_add_orderno_to_sys_table.sql（**本次新增**）
```bash
mysql -u root -p skyserver < sqls/migration_add_orderno_to_sys_table.sql
```
- 为 sys_table 表添加 ORDERNO 字段

## 代码实现

### 1. 实体定义

**文件**: `internal/model/entity/menu.go`

```go
// MenuNode - 通用菜单节点结构
type MenuNode struct {
    ID          uint
    Name        string
    DisplayName string      // 仅 sys_table 使用
    Icon        string
    URL         string
    OrderNo     int
    Type        string      // subsystem, category, table
    Children    []*MenuNode
}

// SysSubsystem - 子系统（一级菜单）
type SysSubsystem struct {
    BaseModel
    Name        string
    OrderNo     int
    URL         string
    Icon        string
    Description string
}

// SysTableCategory - 表类别（二级菜单）
type SysTableCategory struct {
    BaseModel
    SysSubsystemID uint
    Name           string
    OrderNo        int
    Icon           string
    URL            string
    Description    string
}
```

**文件**: `internal/model/entity/sys_table.go`

更新 SysTable 实体，添加缺失字段：
- ✅ OrderNo (int) - 排序
- ✅ IcoImg (string) - 图标
- ✅ SysDirectoryID (*uint) - 安全目录
- ✅ SysParentTableID (*uint) - 父表
- ✅ RowCnt (*int) - 统计行数
- ✅ Description (string) - 备注

### 2. 服务层

**文件**: `internal/service/menu/menu_service.go`

实现两个核心接口：

```go
type Service interface {
    // 获取完整菜单树（管理员用）
    GetMenuTree(ctx context.Context, companyID uint) ([]*entity.MenuNode, error)

    // 获取用户权限过滤后的菜单树（普通用户用）
    GetUserMenuTree(ctx context.Context, userID, companyID uint) ([]*entity.MenuNode, error)
}
```

**权限过滤逻辑**：
1. 查询用户所属权限组（sys_user_groups）
2. 查询权限组对应目录（sys_group_prem）
3. 查询目录对应的表（sys_directory）
4. 只返回用户有权限的表单菜单

**自动过滤空分支**：
- 没有子菜单的分类不显示
- 没有分类的子系统不显示

**排序规则**：
- 按 ORDERNO 升序排列
- ORDERNO 相同时按 ID 升序排列

### 3. 处理器层

**文件**: `api/handler/menu_handler.go`

提供两个 HTTP 接口：

```go
// GET /api/v1/menus/tree - 获取完整菜单树
func (h *MenuHandler) GetMenuTree(c *gin.Context)

// GET /api/v1/menus/user/tree - 获取用户菜单树
func (h *MenuHandler) GetUserMenuTree(c *gin.Context)
```

### 4. 路由注册

**文件**: `api/router/router.go`

```go
// 添加 Menu 服务到 Services 结构
type Services struct {
    // ...
    Menu      menu.Service
    // ...
}

// 注册菜单路由
func registerMenuRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, menuService menu.Service) {
    menuHandler := handler.NewMenuHandler(menuService)
    menus := rg.Group("/menus")
    menus.Use(middleware.AuthRequired(jwtUtil))
    {
        menus.GET("/tree", menuHandler.GetMenuTree)
        menus.GET("/user/tree", menuHandler.GetUserMenuTree)
    }
}
```

**文件**: `cmd/server/main.go`

```go
// 初始化菜单服务
menuService := menu.NewService(db)

// 添加到服务集合
services := &router.Services{
    // ...
    Menu: menuService,
    // ...
}
```

### 5. 插件修复

**文件**: `internal/plugin/cmd/sys_table_after_create.go`

修复字段名称不一致问题：
```go
// 修改前：
categoryID := getUintValue(tableInfo, "SYS_TABLE_CATEGORY_ID")

// 修改后：
categoryID := getUintValue(tableInfo, "SYS_TABLECATEGORY_ID")
```

## API 文档

详细的 API 使用文档请参考：**docs/menu-api.md**

包含内容：
- 接口地址和参数说明
- 请求响应示例
- 数据准备 SQL 示例
- 权限配置说明
- 前端集成示例（Vue 3 和 React）

## 测试数据准备

### 1. 创建子系统
```sql
INSERT INTO sys_subsystem (NAME, ORDERNO, ICON, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES ('系统管理', 10, 'system', 'Y', NOW(), 1);
```

### 2. 创建表类别
```sql
INSERT INTO sys_table_category (SYS_SUBSYSTEM_ID, NAME, ORDERNO, ICON, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES (1, '基础数据', 10, 'database', 'Y', NOW(), 1);
```

### 3. 标记表单为菜单
```sql
UPDATE sys_table
SET IS_MENU = 'Y',
    SYS_TABLECATEGORY_ID = 1,
    ORDERNO = 10,
    ICO_IMG = 'user',
    URL = '/user/list'
WHERE NAME = 'SYS_USER';
```

### 4. 配置权限
```sql
-- 创建目录
INSERT INTO sys_directory (NAME, DISPLAY_NAME, SYS_TABLE_ID, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES ('SYS_USER_LIST', '用户管理', 1, 'Y', NOW(), 1);

-- 分配权限给权限组
INSERT INTO sys_group_prem (SYS_GROUPS_ID, SYS_DIRECTORY_ID, PERMISSION, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES (1, 1, 7, 'Y', NOW(), 1);

-- 分配权限组给用户
INSERT INTO sys_user_groups (SYS_USER_ID, SYS_GROUPS_ID, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES (1, 1, 'Y', NOW(), 1);
```

## 接口测试

### 获取完整菜单树
```bash
curl -H "Authorization: Bearer {token}" \
  http://localhost:9090/api/v1/menus/tree
```

### 获取用户菜单树
```bash
curl -H "Authorization: Bearer {token}" \
  http://localhost:9090/api/v1/menus/user/tree
```

## 响应结构示例

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "系统管理",
      "icon": "system",
      "orderno": 10,
      "type": "subsystem",
      "children": [
        {
          "id": 1,
          "name": "基础数据",
          "icon": "database",
          "orderno": 10,
          "type": "category",
          "children": [
            {
              "id": 1,
              "name": "SYS_USER",
              "displayName": "用户管理",
              "icon": "user",
              "url": "/user/list",
              "orderno": 10,
              "type": "table"
            }
          ]
        }
      ]
    }
  ],
  "message": "success"
}
```

## 文件清单

### 新增文件
1. ✅ `internal/model/entity/menu.go` - 菜单实体定义
2. ✅ `internal/service/menu/menu_service.go` - 菜单服务实现
3. ✅ `api/handler/menu_handler.go` - 菜单处理器
4. ✅ `sqls/migration_add_orderno_to_sys_table.sql` - ORDERNO 字段迁移脚本
5. ✅ `docs/menu-api.md` - 菜单 API 文档
6. ✅ `docs/menu-implementation-summary.md` - 本实现总结（当前文件）

### 修改文件
1. ✅ `sqls/init.sql` - 添加 ORDERNO 字段到 sys_table
2. ✅ `sqls/README.md` - 添加迁移脚本说明
3. ✅ `internal/model/entity/sys_table.go` - 添加缺失字段
4. ✅ `api/router/router.go` - 注册菜单服务和路由
5. ✅ `cmd/server/main.go` - 初始化菜单服务
6. ✅ `internal/plugin/cmd/sys_table_after_create.go` - 修复字段名称

### 删除内容
1. ✅ `sqls/init.sql` - 删除 sys_menu 表定义（第 646-813 行）

## 注意事项

1. **字段命名不一致**：
   - sys_table 表使用 `SYS_TABLECATEGORY_ID`（无下划线）
   - sys_table_category 表使用 `SYS_SUBSYSTEM_ID`（有下划线）

2. **图标字段不同**：
   - sys_subsystem 和 sys_table_category 使用 `ICON`
   - sys_table 使用 `ICO_IMG`

3. **排序字段**：
   - 本次为 sys_table 新增了 `ORDERNO` 字段
   - 现有数据库需要运行迁移脚本

4. **权限控制**：
   - `/menus/tree` 返回完整菜单（适合管理员）
   - `/menus/user/tree` 根据权限过滤（适合普通用户）

5. **空分支过滤**：
   - 自动过滤没有子菜单的分支
   - 确保前端展示的是有效的菜单结构

## 后续工作建议

1. **性能优化**：
   - 考虑添加菜单缓存（Redis）
   - 优化数据库查询（减少 JOIN）

2. **功能增强**：
   - 添加菜单国际化支持
   - 支持动态菜单配置（不依赖代码部署）
   - 添加菜单访问统计

3. **测试覆盖**：
   - 单元测试（service 层）
   - 集成测试（API 层）
   - 权限过滤测试

4. **文档完善**：
   - 添加 Swagger 注释
   - 补充业务流程图
   - 添加常见问题 FAQ
