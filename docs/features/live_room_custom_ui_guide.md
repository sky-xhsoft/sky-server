# 直播间管理功能 - 定制界面 + 标准 API

## 实现方案

- **前端界面**：定制化设计，完全按照产品截图实现
- **后端 API**：使用系统标准的通用 CRUD 接口

## API 接口映射

### 标准通用接口

| 功能 | 方法 | 接口路径 | 说明 |
|------|------|---------|------|
| 查询列表 | POST | `/data/live_room/query` | 支持分页、搜索、过滤 |
| 获取详情 | GET | `/data/live_room/{id}` | 根据 ID 获取单条记录 |
| 创建记录 | POST | `/data/live_room` | 创建新的直播间 |
| 更新记录 | PUT | `/data/live_room/{id}` | 更新指定直播间 |
| 删除记录 | DELETE | `/data/live_room/{id}` | 删除指定直播间（软删除） |

### 查询接口请求格式

```json
POST /data/live_room/query
{
  "page": 1,
  "pageSize": 10,
  "keyword": "搜索关键词",
  "filters": {
    "roomType": "video",
    "status": "live"
  }
}
```

### 查询接口响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "roomName": "测试直播间",
      "roomType": "video",
      "broadcastFormat": "live",
      "status": "draft",
      ...
    }
  ],
  "total": 100
}
```

## 前端页面

### 1. LiveRoomList.vue - 直播间列表页面

**功能特性**：
- 顶部提示信息（可用分数、升级提醒）
- 操作按钮栏（创建、批量创建、全局设置、教程等）
- 搜索和筛选功能
- 数据表格展示
- 行操作（编辑、复制、删除、开始直播等）

**API 调用**：
```typescript
// 查询列表
fetch('/data/live_room/query', {
  method: 'POST',
  body: JSON.stringify({ page, pageSize, keyword })
})

// 删除
fetch(`/data/live_room/${id}`, { method: 'DELETE' })

// 复制（先获取再创建）
fetch(`/data/live_room/${id}`) // GET
fetch('/data/live_room', { method: 'POST', body: copyData })

// 开始直播（更新状态）
fetch(`/data/live_room/${id}`, {
  method: 'PUT',
  body: JSON.stringify({ status: 'live' })
})
```

### 2. LiveRoomForm.vue - 直播间创建/编辑表单

**功能特性**：
- 直播类型选择（视频、图片、VR、语音、图文）
- 直播形式选择（直播、点播/录播、伪直播）
- 直播阶段选择（正式、测试）
- 基础信息输入（名称、时间、封面）
- 显示模式选择（横屏、竖屏、三分屏）
- 观看方式配置（公开、加密、付费等）
- 回放设置（回放方式、有效期）
- 支持创建和编辑两种模式

**API 调用**：
```typescript
// 加载数据（编辑模式）
fetch(`/data/live_room/${roomId}`)

// 创建
fetch('/data/live_room', {
  method: 'POST',
  body: JSON.stringify(formData)
})

// 更新
fetch(`/data/live_room/${roomId}`, {
  method: 'PUT',
  body: JSON.stringify(formData)
})
```

## 路由配置

### BasicLayout.vue 中的路由处理

```typescript
case '/live/rooms':
  currentComponent.value = componentRegistry.LiveRoomList
  navigationStore.navigateTo('LiveRoomList', '直播间管理', {}, false)
  break

case '/live/rooms/create':
  currentComponent.value = componentRegistry.LiveRoomForm
  navigationStore.navigateTo('LiveRoomForm', '创建直播间', {}, false)
  break

case '/live/rooms/edit':
  currentComponent.value = componentRegistry.LiveRoomForm
  const roomId = new URLSearchParams(window.location.search).get('id')
  navigationStore.navigateTo('LiveRoomForm', '编辑直播间', { roomId }, false)
  break
```

## 菜单配置

### 数据库配置

```sql
-- 在 sys_directory 表中添加菜单项
INSERT INTO sys_directory (
  NAME, DISPLAY_NAME, PARENT_ID, ORDERNO, URL,
  SYS_COMPANY_ID, CREATE_BY, CREATE_TIME, IS_ACTIVE
) VALUES (
  'live_room_management',
  '直播间管理',
  (SELECT ID FROM sys_directory WHERE NAME = 'live_stream'),
  60,
  '/live/rooms',
  1,
  'system',
  NOW(),
  'Y'
);
```

## 数据字段映射

### 前端字段 → 数据库字段

| 前端字段 | 数据库字段 | 类型 | 说明 |
|---------|-----------|------|------|
| roomName | ROOM_NAME | varchar | 直播间名称 |
| roomType | ROOM_TYPE | varchar | 直播间类型 |
| broadcastFormat | BROADCAST_FORMAT | varchar | 播出形式 |
| roomStage | ROOM_STAGE | varchar | 直播间阶段 |
| displayMode | DISPLAY_MODE | varchar | 显示方式 |
| startTime | START_TIME | datetime | 开始时间 |
| endTime | END_TIME | datetime | 结束时间 |
| coverImage | COVER_IMAGE | varchar | 直播间封面 |
| viewingMethod | VIEWING_METHOD | varchar | 观看方式 |
| viewingPassword | VIEWING_PASSWORD | varchar | 观看密码 |
| viewingPrice | VIEWING_PRICE | decimal | 观看价格 |
| playbackMethod | PLAYBACK_METHOD | varchar | 回放方式 |
| playbackValidity | PLAYBACK_VALIDITY | varchar | 回放有效期 |
| status | STATUS | varchar | 状态 |

## 后端要求

### 通用数据处理器

后端需要提供通用的数据处理器，支持：

1. **动态表操作**：根据表名（live_room）自动处理 CRUD
2. **元数据驱动**：根据 sys_table、sys_column 配置自动验证和处理
3. **权限控制**：根据 MASK 字段控制操作权限
4. **数据过滤**：根据公司 ID 自动过滤数据
5. **审计字段**：自动填充 CREATE_BY、CREATE_TIME、UPDATE_BY、UPDATE_TIME

### 示例实现（Go）

```go
// 通用查询接口
func (h *DataHandler) Query(c *gin.Context) {
    tableName := c.Param("table")

    var req QueryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 根据表名和元数据配置查询
    data, total, err := h.service.Query(tableName, req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "data": data,
        "total": total
    })
}

// 通用创建接口
func (h *DataHandler) Create(c *gin.Context) {
    tableName := c.Param("table")

    var data map[string]interface{}
    if err := c.ShouldBindJSON(&data); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 自动填充审计字段
    data["CREATE_BY"] = c.GetString("userId")
    data["CREATE_TIME"] = time.Now()
    data["SYS_COMPANY_ID"] = c.GetUint("companyId")

    // 根据表名和元数据配置创建
    id, err := h.service.Create(tableName, data)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"id": id})
}
```

## 优势

1. **界面定制化**：完全按照产品需求设计，提供最佳用户体验
2. **后端标准化**：使用通用 API，减少重复代码，易于维护
3. **灵活扩展**：前端可以自由调整 UI，后端无需修改
4. **统一管理**：所有表的 CRUD 都使用相同的 API 模式
5. **权限统一**：通过元数据配置统一控制权限

## 测试验证

### 1. 列表页面测试
- 访问 `/live/rooms`
- 验证数据加载
- 测试搜索功能
- 测试分页功能
- 测试操作按钮

### 2. 创建功能测试
- 点击"创建直播"按钮
- 填写表单
- 提交创建
- 验证数据保存

### 3. 编辑功能测试
- 点击"编辑"按钮
- 修改数据
- 提交更新
- 验证数据更新

### 4. 删除功能测试
- 点击"删除"按钮
- 确认删除
- 验证数据删除（软删除）

## 注意事项

1. **API 路径**：确保后端已实现 `/data/{tableName}/*` 通用接口
2. **认证授权**：所有请求需要携带认证 token
3. **公司隔离**：后端需要根据当前用户的公司 ID 自动过滤数据
4. **错误处理**：前端需要处理各种错误情况并给出友好提示
5. **数据验证**：前端表单验证 + 后端数据验证双重保障
