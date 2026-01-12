# 菜单接口文档

## 概述

菜单系统采用三级目录结构：
1. **一级菜单（子系统）**: sys_subsystem
2. **二级菜单（表类别）**: sys_table_category
3. **三级菜单（表单）**: sys_table (IS_MENU='Y')

## API 接口

### 1. 获取完整菜单树

**接口地址**: `GET /api/v1/menus/tree`

**请求头**:
```
Authorization: Bearer {access_token}
```

**响应示例**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "系统管理",
      "icon": "system",
      "url": "",
      "orderno": 10,
      "type": "subsystem",
      "children": [
        {
          "id": 1,
          "name": "基础数据",
          "icon": "database",
          "url": "",
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
            },
            {
              "id": 2,
              "name": "SYS_COMPANY",
              "displayName": "公司管理",
              "icon": "building",
              "url": "/company/list",
              "orderno": 20,
              "type": "table"
            }
          ]
        },
        {
          "id": 2,
          "name": "权限管理",
          "icon": "lock",
          "url": "",
          "orderno": 20,
          "type": "category",
          "children": [
            {
              "id": 10,
              "name": "SYS_GROUPS",
              "displayName": "权限组",
              "icon": "team",
              "url": "/groups/list",
              "orderno": 10,
              "type": "table"
            }
          ]
        }
      ]
    },
    {
      "id": 2,
      "name": "业务管理",
      "icon": "appstore",
      "url": "",
      "orderno": 20,
      "type": "subsystem",
      "children": []
    }
  ],
  "message": "success"
}
```

**说明**:
- 返回系统中所有有效的菜单项（IS_ACTIVE='Y'）
- 只包含标记为菜单的表（IS_MENU='Y'）
- 自动过滤空分支（没有子菜单的分类和子系统不会显示）
- 按 ORDERNO 字段排序

**字段说明**:
| 字段 | 类型 | 说明 |
|-----|------|------|
| id | uint | 菜单项ID |
| name | string | 名称（表单使用数据库名称） |
| displayName | string | 显示名称（仅表单有） |
| icon | string | 图标 |
| url | string | 访问URL |
| orderno | int | 排序号 |
| type | string | 类型：subsystem/category/table |
| children | array | 子菜单列表 |

### 2. 获取用户菜单树

**接口地址**: `GET /api/v1/menus/user/tree`

**请求头**:
```
Authorization: Bearer {access_token}
```

**响应示例**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "系统管理",
      "icon": "system",
      "url": "",
      "orderno": 10,
      "type": "subsystem",
      "children": [
        {
          "id": 1,
          "name": "基础数据",
          "icon": "database",
          "url": "",
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

**说明**:
- 根据用户的权限组过滤菜单
- 只返回用户有权限访问的表单
- 权限判断流程：
  1. 查询用户所属的权限组（sys_user_groups）
  2. 查询权限组对应的目录（sys_group_prem）
  3. 查询目录对应的表（sys_directory）
  4. 只返回用户有权限的表单菜单
- 自动过滤空分支

## 数据准备

### 1. 创建子系统（一级菜单）

```sql
INSERT INTO sys_subsystem (NAME, ORDERNO, ICON, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES ('系统管理', 10, 'system', 'Y', NOW(), 1);
```

### 2. 创建表类别（二级菜单）

```sql
INSERT INTO sys_table_category (SYS_SUBSYSTEM_ID, NAME, ORDERNO, ICON, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES (1, '基础数据', 10, 'database', 'Y', NOW(), 1);
```

### 3. 标记表单为菜单（三级菜单）

```sql
UPDATE sys_table
SET IS_MENU = 'Y',
    SYS_TABLECATEGORY_ID = 1,  -- 关联到表类别
    ORDERNO = 10,
    ICO_IMG = 'user',
    URL = '/user/list'
WHERE NAME = 'SYS_USER';
```

## 权限配置

### 1. 创建目录

```sql
INSERT INTO sys_directory (NAME, DISPLAY_NAME, SYS_TABLE_ID, SYS_TABLE_CATEGORY_ID, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES ('SYS_USER_LIST', '用户管理', 1, 1, 'Y', NOW(), 1);
```

### 2. 分配权限给权限组

```sql
INSERT INTO sys_group_prem (SYS_GROUPS_ID, SYS_DIRECTORY_ID, PERMISSION, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES (1, 1, 7, 'Y', NOW(), 1);  -- 权限值：1=读, 3=读写, 5=读提交, 7=读写提交
```

### 3. 分配权限组给用户

```sql
INSERT INTO sys_user_groups (SYS_USER_ID, SYS_GROUPS_ID, IS_ACTIVE, CREATE_TIME, SYS_COMPANY_ID)
VALUES (1, 1, 'Y', NOW(), 1);
```

## 前端集成示例

### Vue 3 + Ant Design Vue

```vue
<template>
  <a-menu mode="inline" :items="menuItems" />
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getMenuTree, getUserMenuTree } from '@/api/menu'

const menuItems = ref([])

onMounted(async () => {
  try {
    // 获取用户菜单
    const res = await getUserMenuTree()
    if (res.code === 200) {
      menuItems.value = transformMenu(res.data)
    }
  } catch (error) {
    console.error('获取菜单失败:', error)
  }
})

// 转换菜单数据为 Ant Design 格式
function transformMenu(data) {
  return data.map(item => ({
    key: `${item.type}-${item.id}`,
    label: item.displayName || item.name,
    icon: item.icon,
    path: item.url,
    children: item.children ? transformMenu(item.children) : undefined
  }))
}
</script>
```

### React + Ant Design

```jsx
import { Menu } from 'antd'
import { useEffect, useState } from 'react'
import { getUserMenuTree } from './api/menu'

function MenuComponent() {
  const [menuItems, setMenuItems] = useState([])

  useEffect(() => {
    loadMenu()
  }, [])

  const loadMenu = async () => {
    try {
      const res = await getUserMenuTree()
      if (res.code === 200) {
        setMenuItems(transformMenu(res.data))
      }
    } catch (error) {
      console.error('获取菜单失败:', error)
    }
  }

  const transformMenu = (data) => {
    return data.map(item => ({
      key: `${item.type}-${item.id}`,
      label: item.displayName || item.name,
      icon: item.icon,
      path: item.url,
      children: item.children ? transformMenu(item.children) : undefined
    }))
  }

  return <Menu mode="inline" items={menuItems} />
}

export default MenuComponent
```

## 注意事项

1. **排序规则**:
   - 所有菜单项按 ORDERNO 升序排列
   - ORDERNO 相同时按 ID 升序排列

2. **空分支过滤**:
   - 没有子菜单的分类不会显示
   - 没有分类的子系统不会显示

3. **权限控制**:
   - `/menus/tree` 返回完整菜单（适合管理员）
   - `/menus/user/tree` 返回权限过滤后的菜单（适合普通用户）

4. **字段命名**:
   - sys_table 表中字段名为 `SYS_TABLECATEGORY_ID`（无下划线）
   - sys_table_category 表中字段名为 `SYS_SUBSYSTEM_ID`（有下划线）

5. **图标字段**:
   - sys_subsystem 使用 `ICON`
   - sys_table_category 使用 `ICON`
   - sys_table 使用 `ICO_IMG`
