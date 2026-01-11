# Phase 10 - 完善旧权限组体系方案

## 决策说明

**用户决定：使用旧的权限组体系（sys_groups系列）**

理由可能包括：
- 旧系统已经在生产环境使用
- 基于"安全目录"和"权限组"的模型更符合业务需求
- 位运算的权限值更高效
- 数据过滤条件(FILTER_OBJ)的设计更灵活
- 不想大规模重构现有代码

## 当前权限组体系架构

### 数据模型

```
sys_groups (权限组)
  ├── NAME: 权限组名称
  ├── DESCRIPTION: 描述
  └── SGRADE: 字段访问级别

sys_user_groups (用户权限组关联)
  ├── SYS_USER_ID: 用户ID
  └── SYS_DIRECTORY_ID: 目录ID

sys_directory (安全目录)
  ├── NAME: 目录名称
  ├── SYS_TABLE_ID: 关联表ID
  ├── PARENT_ID: 父目录ID
  ├── ORDERNO: 排序号
  └── DESCRIPTION: 描述

sys_group_prem (权限组明细)
  ├── SYS_GROUPS_ID: 权限组ID
  ├── SYS_DIRECTORY_ID: 目录ID
  ├── PERMISSION: 权限值(位运算)
  └── FILTER_OBJ: 数据过滤条件(JSON)
```

### 权限流程

```
用户
  ↓
sys_user_groups (用户关联权限组)
  ↓
sys_groups (权限组)
  ↓
sys_group_prem (权限明细)
  ↓
sys_directory (安全目录) + PERMISSION (位运算) + FILTER_OBJ (数据过滤)
  ↓
访问控制
```

### 权限位运算定义

```go
const (
    PermNone   = 0      // 无权限
    PermRead   = 1 << 0 // 0001 = 1   读取
    PermCreate = 1 << 1 // 0010 = 2   创建
    PermUpdate = 1 << 2 // 0100 = 4   更新
    PermDelete = 1 << 3 // 1000 = 8   删除
    PermExport = 1 << 4 // 10000 = 16 导出
    PermImport = 1 << 5 // 100000 = 32 导入
    PermAll    = 63     // 111111     所有权限
)

// 示例：用户有读取和更新权限
// PERMISSION = 1 | 4 = 5 (0101)
```

## Phase 10 实施方案

### 目标

1. ✅ 为权限组体系创建完整的API接口
2. ✅ 创建权限检查中间件
3. ✅ 完善权限组管理服务
4. ✅ 处理新旧系统的冲突
5. ✅ 统一权限检查逻辑

### 方案选择

#### 方案A: 完全废弃新RBAC，纯用旧系统 (推荐)

**优点:**
- ✅ 系统最简洁
- ✅ 无冲突
- ✅ 维护成本最低

**缺点:**
- ❌ 需要删除Phase 8的代码
- ❌ 菜单系统需要调整

**实施步骤:**
1. 为旧权限组体系补充API
2. 创建权限检查中间件
3. 删除新RBAC相关代码和表
4. 调整菜单系统使用旧权限体系

#### 方案B: 新RBAC专门用于菜单，旧系统用于数据权限

**优点:**
- ✅ 各司其职，职责清晰
- ✅ 不删除已实现的代码

**缺点:**
- ❌ 维护两套系统
- ❌ 概念混淆

**分工:**
- **新RBAC (sys_role)**: 仅用于菜单和按钮的显示控制
- **旧权限组 (sys_groups)**: 用于数据的增删改查权限控制

#### 方案C: 将新RBAC作为旧系统的上层封装

**优点:**
- ✅ 对外提供统一的RBAC接口
- ✅ 底层使用旧权限组

**缺点:**
- ❌ 架构复杂
- ❌ 性能损失

## 推荐实施：方案A - 完全使用旧权限组体系

### 第一步：补充权限组API接口

#### 1.1 权限组管理服务

```go
// internal/service/groups/groups_service.go

package groups

import (
    "context"
    "github.com/sky-xhsoft/sky-server/internal/model/entity"
    "gorm.io/gorm"
)

// Service 权限组服务接口
type Service interface {
    // 权限组管理
    CreateGroup(ctx context.Context, group *entity.SysGroups) error
    UpdateGroup(ctx context.Context, group *entity.SysGroups) error
    DeleteGroup(ctx context.Context, id uint) error
    GetGroup(ctx context.Context, id uint) (*entity.SysGroups, error)
    ListGroups(ctx context.Context, req *ListGroupsRequest) ([]*entity.SysGroups, int64, error)

    // 安全目录管理
    CreateDirectory(ctx context.Context, dir *entity.SysDirectory) error
    UpdateDirectory(ctx context.Context, dir *entity.SysDirectory) error
    DeleteDirectory(ctx context.Context, id uint) error
    GetDirectory(ctx context.Context, id uint) (*entity.SysDirectory, error)
    GetDirectoryTree(ctx context.Context, parentID uint) ([]*DirectoryNode, error)

    // 权限组明细管理
    AssignPermissions(ctx context.Context, groupID uint, permissions []*GroupPermission) error
    GetGroupPermissions(ctx context.Context, groupID uint) ([]*entity.SysGroupPrem, error)

    // 用户权限组管理
    AssignGroupToUser(ctx context.Context, userID uint, directoryIDs []uint) error
    GetUserGroups(ctx context.Context, userID uint) ([]*entity.SysGroups, error)

    // 权限检查
    CheckUserPermission(ctx context.Context, userID uint, directoryID uint, permission int) (bool, error)
    GetUserDirectoryPermission(ctx context.Context, userID uint, directoryID uint) (int, error)
    CheckUserTablePermission(ctx context.Context, userID uint, tableID uint, permission int) (bool, error)
}

// GroupPermission 权限组权限
type GroupPermission struct {
    DirectoryID uint
    Permission  int    // 位运算权限值
    FilterObj   string // JSON格式的过滤条件
}

// DirectoryNode 目录树节点
type DirectoryNode struct {
    *entity.SysDirectory
    Children []*DirectoryNode `json:"children"`
}

// ListGroupsRequest 查询权限组请求
type ListGroupsRequest struct {
    Name     string
    Page     int
    PageSize int
}
```

#### 1.2 权限检查中间件

```go
// internal/api/middleware/group_permission.go

package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/sky-xhsoft/sky-server/internal/service/groups"
    "github.com/sky-xhsoft/sky-server/pkg/errors"
)

// 权限位定义
const (
    PermRead   = 1 << 0 // 1 - 读取
    PermCreate = 1 << 1 // 2 - 创建
    PermUpdate = 1 << 2 // 4 - 更新
    PermDelete = 1 << 3 // 8 - 删除
    PermExport = 1 << 4 // 16 - 导出
    PermImport = 1 << 5 // 32 - 导入
)

// DirectoryPermissionRequired 安全目录权限检查中间件
func DirectoryPermissionRequired(groupService groups.Service, directoryID uint, requiredPerm int) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取用户ID
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "code":    errors.ErrUnauthorized,
                "message": "未登录",
            })
            c.Abort()
            return
        }

        // 检查权限
        hasPermission, err := groupService.CheckUserPermission(
            c.Request.Context(),
            userID.(uint),
            directoryID,
            requiredPerm,
        )

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "code":    errors.ErrInternal,
                "message": "权限检查失败",
            })
            c.Abort()
            return
        }

        if !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{
                "code":    errors.ErrForbidden,
                "message": "无权限访问",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// TablePermissionRequired 表权限检查中间件
func TablePermissionRequired(groupService groups.Service, tableID uint, requiredPerm int) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "code":    errors.ErrUnauthorized,
                "message": "未登录",
            })
            c.Abort()
            return
        }

        // 检查表权限
        hasPermission, err := groupService.CheckUserTablePermission(
            c.Request.Context(),
            userID.(uint),
            tableID,
            requiredPerm,
        )

        if err != nil || !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{
                "code":    errors.ErrForbidden,
                "message": "无权限访问该表",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// GetUserPermission 获取用户权限(不阻止请求)
func GetUserPermission(groupService groups.Service, directoryID uint) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.Set("userPermission", 0)
            c.Next()
            return
        }

        permission, err := groupService.GetUserDirectoryPermission(
            c.Request.Context(),
            userID.(uint),
            directoryID,
        )

        if err != nil {
            permission = 0
        }

        c.Set("userPermission", permission)
        c.Next()
    }
}
```

#### 1.3 API Handler

```go
// internal/api/handler/groups_handler.go

package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/sky-xhsoft/sky-server/internal/service/groups"
)

type GroupsHandler struct {
    groupService groups.Service
}

func NewGroupsHandler(groupService groups.Service) *GroupsHandler {
    return &GroupsHandler{
        groupService: groupService,
    }
}

// CreateGroup 创建权限组
func (h *GroupsHandler) CreateGroup(c *gin.Context) {
    // 实现创建权限组
}

// ListGroups 查询权限组列表
func (h *GroupsHandler) ListGroups(c *gin.Context) {
    // 实现查询权限组列表
}

// GetGroup 获取权限组详情
func (h *GroupsHandler) GetGroup(c *gin.Context) {
    // 实现获取权限组详情
}

// AssignPermissions 分配权限给权限组
func (h *GroupsHandler) AssignPermissions(c *gin.Context) {
    // 实现分配权限
}

// CreateDirectory 创建安全目录
func (h *GroupsHandler) CreateDirectory(c *gin.Context) {
    // 实现创建安全目录
}

// GetDirectoryTree 获取目录树
func (h *GroupsHandler) GetDirectoryTree(c *gin.Context) {
    // 实现获取目录树
}

// AssignGroupToUser 分配权限组给用户
func (h *GroupsHandler) AssignGroupToUser(c *gin.Context) {
    // 实现分配权限组给用户
}

// GetUserGroups 获取用户的权限组
func (h *GroupsHandler) GetUserGroups(c *gin.Context) {
    // 实现获取用户权限组
}

// CheckPermission 检查用户权限
func (h *GroupsHandler) CheckPermission(c *gin.Context) {
    // 实现权限检查
}
```

### 第二步：调整菜单系统

由于菜单系统(Phase 9)依赖权限系统，需要调整：

#### 选项1: 菜单直接关联安全目录

```go
// internal/model/entity/menu.go

type Menu struct {
    // ... 其他字段
    // 删除: PermCode string
    // 新增:
    DirectoryID *uint `gorm:"column:DIRECTORY_ID;index" json:"directoryId"` // 关联安全目录
}
```

#### 选项2: 菜单系统保持独立(推荐)

菜单只控制显示，不控制权限。真正的权限检查在API层通过权限组中间件实现。

```go
// 菜单表保持不变
// 去掉菜单的权限过滤，所有用户看到相同菜单
// 在API接口上使用DirectoryPermissionRequired中间件控制访问
```

### 第三步：删除新RBAC系统

#### 3.1 删除代码文件

```bash
# 删除实体
rm internal/model/entity/role.go
rm internal/model/entity/permission.go
rm internal/model/entity/user_role.go
rm internal/model/entity/role_permission.go

# 删除服务
rm -rf internal/service/role/
rm -rf internal/service/perm/

# 删除Handler
rm internal/api/handler/role_handler.go
rm internal/api/handler/permission_handler.go

# 删除中间件
# 编辑 internal/api/middleware/permission.go，删除RBAC相关中间件
```

#### 3.2 删除数据库表

```sql
-- sqls/cleanup_rbac.sql

-- 删除新RBAC表
DROP TABLE IF EXISTS sys_role_permission;
DROP TABLE IF EXISTS sys_user_role;
DROP TABLE IF EXISTS sys_permission;
DROP TABLE IF EXISTS sys_role;
```

#### 3.3 更新路由配置

```go
// internal/api/router/router.go

// 删除
// registerRoleRoutes(v1, jwtUtil, services.Role)
// registerPermissionRoutes(v1, jwtUtil, services.Permission)

// 新增
registerGroupsRoutes(v1, jwtUtil, services.Groups)
registerDirectoryRoutes(v1, jwtUtil, services.Groups)
```

#### 3.4 更新main.go

```go
// cmd/server/main.go

// 删除
// import "github.com/sky-xhsoft/sky-server/internal/service/role"
// import "github.com/sky-xhsoft/sky-server/internal/service/perm"
// roleService := role.NewService(db)
// permissionService := perm.NewService(db)

// 新增
import "github.com/sky-xhsoft/sky-server/internal/service/groups"
groupService := groups.NewService(db)
```

### 第四步：完善旧系统的SQL脚本

```sql
-- sqls/groups_enhanced.sql

-- 确保所有表结构完整

-- 1. 权限组表
CREATE TABLE IF NOT EXISTS `sys_groups` (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `UPDATE_BY` varchar(80) NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  `NAME` varchar(255) NOT NULL COMMENT '权限组名称',
  `DESCRIPTION` varchar(255) NULL COMMENT '描述',
  `SGRADE` int NULL COMMENT '字段访问级别',
  PRIMARY KEY (`ID`),
  INDEX `idx_name` (`NAME`),
  INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限组';

-- 2. 用户权限组关联表
CREATE TABLE IF NOT EXISTS `sys_user_groups` (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL,
  `CREATE_BY` varchar(80) NULL,
  `CREATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `UPDATE_BY` varchar(80) NULL,
  `UPDATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y',
  `SYS_USER_ID` int UNSIGNED NOT NULL COMMENT '用户ID',
  `SYS_DIRECTORY_ID` int UNSIGNED NOT NULL COMMENT '目录ID',
  PRIMARY KEY (`ID`),
  INDEX `idx_user_groups` (`SYS_USER_ID`, `SYS_DIRECTORY_ID`),
  INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户权限组关联';

-- 3. 安全目录表
CREATE TABLE IF NOT EXISTS `sys_directory` (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL,
  `CREATE_BY` varchar(80) NULL,
  `CREATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `UPDATE_BY` varchar(80) NULL,
  `UPDATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y',
  `NAME` varchar(255) NOT NULL COMMENT '目录名称',
  `SYS_TABLE_ID` int UNSIGNED NULL COMMENT '关联表ID',
  `PARENT_ID` int UNSIGNED NULL COMMENT '父目录ID',
  `ORDERNO` int NULL DEFAULT 0 COMMENT '排序号',
  `DESCRIPTION` varchar(255) NULL COMMENT '描述',
  PRIMARY KEY (`ID`),
  INDEX `idx_parent_id` (`PARENT_ID`),
  INDEX `idx_table_id` (`SYS_TABLE_ID`),
  INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='安全目录';

-- 4. 权限组明细表
CREATE TABLE IF NOT EXISTS `sys_group_prem` (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL,
  `CREATE_BY` varchar(80) NULL,
  `CREATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `UPDATE_BY` varchar(80) NULL,
  `UPDATE_TIME` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y',
  `SYS_GROUPS_ID` int UNSIGNED NOT NULL COMMENT '权限组ID',
  `SYS_DIRECTORY_ID` int UNSIGNED NOT NULL COMMENT '目录ID',
  `PERMISSION` int NOT NULL DEFAULT 0 COMMENT '权限值(位运算)',
  `FILTER_OBJ` varchar(255) NULL COMMENT '数据过滤条件(JSON)',
  PRIMARY KEY (`ID`),
  INDEX `idx_group_dir` (`SYS_GROUPS_ID`, `SYS_DIRECTORY_ID`),
  INDEX `idx_is_active` (`IS_ACTIVE`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限组明细';

-- 初始数据
INSERT INTO `sys_groups` (`NAME`, `DESCRIPTION`, `SGRADE`, `IS_ACTIVE`) VALUES
('管理员组', '系统管理员权限组', 1, 'Y'),
('普通用户组', '普通用户权限组', 5, 'Y');

-- 安全目录示例
INSERT INTO `sys_directory` (`NAME`, `PARENT_ID`, `ORDERNO`, `DESCRIPTION`, `IS_ACTIVE`) VALUES
('系统管理', NULL, 1, '系统管理根目录', 'Y'),
('业务管理', NULL, 2, '业务管理根目录', 'Y');
```

## 实施计划

### Week 1: 服务层开发
- [ ] 创建 groups_service.go (权限组服务)
- [ ] 实现所有Service接口方法
- [ ] 编写单元测试

### Week 2: API层开发
- [ ] 创建 groups_handler.go
- [ ] 实现所有API接口
- [ ] 创建权限检查中间件

### Week 3: 清理和集成
- [ ] 删除新RBAC代码
- [ ] 删除新RBAC数据库表
- [ ] 调整菜单系统
- [ ] 更新路由配置

### Week 4: 测试和文档
- [ ] 全面测试权限功能
- [ ] 编写API文档
- [ ] 更新系统文档

## API接口设计

### 权限组管理 (9个接口)

| 接口路径 | 方法 | 功能 |
|---------|------|------|
| `/api/v1/groups` | POST | 创建权限组 |
| `/api/v1/groups` | GET | 查询权限组列表 |
| `/api/v1/groups/:id` | GET | 获取权限组详情 |
| `/api/v1/groups/:id` | PUT | 更新权限组 |
| `/api/v1/groups/:id` | DELETE | 删除权限组 |
| `/api/v1/groups/:id/permissions` | POST | 分配权限 |
| `/api/v1/groups/:id/permissions` | GET | 获取权限组权限 |
| `/api/v1/groups/users/:userId` | POST | 分配权限组给用户 |
| `/api/v1/groups/users/:userId` | GET | 获取用户权限组 |

### 安全目录管理 (6个接口)

| 接口路径 | 方法 | 功能 |
|---------|------|------|
| `/api/v1/directories` | POST | 创建安全目录 |
| `/api/v1/directories` | GET | 查询目录列表 |
| `/api/v1/directories/tree` | GET | 获取目录树 |
| `/api/v1/directories/:id` | GET | 获取目录详情 |
| `/api/v1/directories/:id` | PUT | 更新目录 |
| `/api/v1/directories/:id` | DELETE | 删除目录 |

### 权限检查 (2个接口)

| 接口路径 | 方法 | 功能 |
|---------|------|------|
| `/api/v1/permissions/check` | POST | 检查权限 |
| `/api/v1/permissions/user` | GET | 获取用户权限 |

**总计: 17个接口**

## 使用示例

### 创建权限组并分配权限

```javascript
// 1. 创建权限组
POST /api/v1/groups
{
  "name": "销售部门",
  "description": "销售部门权限组",
  "sgrade": 3
}
// 返回 groupId = 10

// 2. 创建安全目录
POST /api/v1/directories
{
  "name": "客户管理",
  "sysTableId": 5,  // 关联客户表
  "parentId": null
}
// 返回 directoryId = 20

// 3. 分配权限给权限组
POST /api/v1/groups/10/permissions
{
  "permissions": [
    {
      "directoryId": 20,
      "permission": 7,  // 1+2+4 = 读取+创建+更新
      "filterObj": "{\"department\":\"sales\"}"  // 只能看销售部门的客户
    }
  ]
}

// 4. 分配权限组给用户
POST /api/v1/groups/users/100
{
  "directoryIds": [20]
}
```

### 使用中间件保护API

```go
// 保护CRUD接口
router.GET("/data/:tableName/:id",
    middleware.AuthRequired(jwtUtil),
    middleware.TablePermissionRequired(groupService, tableID, middleware.PermRead),
    handler.GetOne)

router.POST("/data/:tableName",
    middleware.AuthRequired(jwtUtil),
    middleware.TablePermissionRequired(groupService, tableID, middleware.PermCreate),
    handler.Create)

router.PUT("/data/:tableName/:id",
    middleware.AuthRequired(jwtUtil),
    middleware.TablePermissionRequired(groupService, tableID, middleware.PermUpdate),
    handler.Update)

router.DELETE("/data/:tableName/:id",
    middleware.AuthRequired(jwtUtil),
    middleware.TablePermissionRequired(groupService, tableID, middleware.PermDelete),
    handler.Delete)
```

## 优势总结

使用权限组体系的优势：

1. **高效的位运算**: 权限检查速度快
2. **灵活的数据过滤**: FILTER_OBJ支持复杂过滤条件
3. **目录树结构**: 与文件系统类似，易于理解
4. **字段级权限**: SGRADE支持字段访问控制
5. **表关联**: 直接关联数据表，权限控制更直接

## 下一步行动

请确认是否按此方案实施：

1. ✅ 完善旧权限组体系API
2. ✅ 删除新RBAC系统
3. ✅ 调整菜单系统
4. ✅ 统一使用权限组进行权限控制

如果确认，我可以立即开始实施Phase 10。
