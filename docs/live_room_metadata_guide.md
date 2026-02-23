# 直播间管理功能说明

## 实现方式

本系统采用**元数据驱动架构**，直播间管理功能通过配置元数据自动生成，无需编写自定义页面代码。

## 配置步骤

### 1. 执行数据库迁移

```bash
# 创建直播间表
mysql -u root -p your_database < sqls/migrations/create_live_room.sql

# 配置元数据
mysql -u root -p your_database < sqls/migrations/configure_live_room_metadata.sql
```

### 2. 菜单配置

SQL 脚本已自动配置菜单，菜单项信息：
- **菜单名称**：直播间管理
- **父菜单**：视频直播（live_stream）
- **URL 格式**：`/metadata/list?tableId={live_room表ID}`
- **排序**：60

### 3. 前端访问

配置完成后，系统会自动：
1. 在"视频直播"菜单下显示"直播间管理"菜单项
2. 点击菜单后，自动加载 `MetadataListView` 组件
3. `MetadataListView` 会根据 `tableId` 参数加载 `live_room` 表的元数据配置
4. 自动生成列表页面，包括：
   - 搜索框（根据 `IS_QUERY='Y'` 的字段）
   - 数据表格（显示所有配置的字段）
   - 操作按钮（新增、编辑、删除等，根据 `MASK` 字段控制）
5. 点击"新增"或"编辑"按钮时，自动加载 `MetadataFormView` 组件
6. `MetadataFormView` 根据元数据配置自动生成表单，包括：
   - 所有字段的输入控件（根据 `DISPLAY_TYPE` 自动选择）
   - 下拉选项（根据 `SYS_DICT_ID` 关联的数据字典）
   - 字段验证（根据 `NULL_ABLE`、`REG_EXPRESSION` 等配置）

## API 端点

系统使用通用的数据 API，无需单独开发：

### 通用 CRUD API
- `POST /data/live_room/query` - 查询列表
- `GET /data/live_room/{id}` - 获取详情
- `POST /data/live_room` - 创建记录
- `PUT /data/live_room/{id}` - 更新记录
- `DELETE /data/live_room/{id}` - 删除记录

这些 API 由后端的通用数据处理器自动提供，根据元数据配置动态处理。

## 字段配置说明

### 显示类型（DISPLAY_TYPE）
- `text` - 文本输入框
- `textarea` - 多行文本框
- `number` - 数字输入框
- `select` - 下拉选择框（需配置 SYS_DICT_ID）
- `datetime` - 日期时间选择器
- `time` - 时间选择器
- `image` - 图片上传
- `password` - 密码输入框
- `json` - JSON 编辑器

### 赋值方式（SET_VALUE_TYPE）
- `pk` - 主键（自动生成）
- `byPage` - 界面输入
- `createBy` - 创建人（自动填充）
- `operator` - 操作人（自动填充）
- `sysdate` - 系统时间（自动填充）
- `ignore` - 忽略（不显示在表单中）

### 权限控制（MASK）
- `A` - 新增
- `M` - 修改
- `D` - 删除
- `Q` - 查询
- `S` - 提交
- `U` - 反提交
- `V` - 作废
- `E` - 导出

## 自定义页面（可选）

如果需要更复杂的 UI 或特殊交互，可以创建自定义页面：

### 已创建的自定义组件（供参考）
- `LiveRoomList.vue` - 自定义列表页面
- `LiveRoomForm.vue` - 自定义表单页面

这些组件展示了如何实现类似截图的复杂 UI，但在元数据驱动架构下，**推荐使用系统自带的 MetadataListView 和 MetadataFormView**，它们已经提供了完整的 CRUD 功能。

### 何时使用自定义页面
- 需要特殊的布局或交互
- 需要集成第三方组件
- 需要复杂的业务逻辑

### 如何使用自定义页面
1. 在 BasicLayout.vue 中注册组件
2. 在 loadComponent 函数中添加路由处理
3. 修改菜单配置，将 URL 改为自定义路由

## 扩展功能

### 添加自定义字段渲染器
如果需要特殊的字段显示方式，可以创建自定义字段渲染器：

```vue
<!-- src/modules/metadata/components/FieldRenderers/CustomField.vue -->
<template>
  <div>
    <!-- 自定义字段UI -->
  </div>
</template>

<script setup lang="ts">
// 自定义字段逻辑
</script>
```

然后在 `DynamicFormItem.vue` 中注册：

```typescript
case 'custom_field_type':
  return h(CustomField, {
    modelValue: props.modelValue,
    'onUpdate:modelValue': (value) => emit('update:modelValue', value)
  })
```

### 添加自定义操作按钮
可以通过 `sys_action` 表配置自定义操作按钮，支持：
- URL 跳转
- JavaScript 脚本
- 存储过程调用
- 后台任务

## 优势

使用元数据驱动架构的优势：
1. **零代码开发**：通过配置即可实现 CRUD 功能
2. **统一风格**：所有页面保持一致的 UI 风格
3. **易于维护**：修改字段、调整顺序只需修改数据库配置
4. **权限控制**：通过 MASK 字段统一控制操作权限
5. **快速迭代**：新增字段或修改配置无需重新编译发布

## 故障排查

### 菜单点击无反应
1. 检查 `sys_directory` 表中的 URL 格式是否正确
2. 检查 `SYS_TABLE_ID` 是否正确关联到 `live_room` 表
3. 检查 BasicLayout.vue 中的 `loadComponent` 函数是否处理了 `/metadata/list` 路由

### 404 错误
1. 确认后端通用数据 API 已启用
2. 检查表名是否正确（应为 `live_room`）
3. 检查数据库连接和权限

### 字段不显示
1. 检查 `sys_column` 表中的字段配置
2. 确认 `IS_ACTIVE='Y'`
3. 检查 `SET_VALUE_TYPE` 是否为 `ignore`

### 下拉选项为空
1. 检查 `SYS_DICT_ID` 是否正确关联到数据字典
2. 确认 `sys_dict_item` 表中有对应的选项数据
3. 检查选项的 `IS_ACTIVE='Y'`
