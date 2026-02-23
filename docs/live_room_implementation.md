# 直播间功能实现说明

## 概述

直播间功能已按照元数据驱动架构实现，支持视频直播、图片直播、VR直播、语音直播和图文直播等多种类型。

## 已创建的文件

### 1. 数据模型
- `internal/model/entity/live_room.go` - 直播间实体定义

### 2. 数据库迁移
- `sqls/migrations/create_live_room.sql` - 创建直播间表
- `sqls/migrations/configure_live_room_metadata.sql` - 配置元数据（sys_table、sys_column、sys_dict等）

### 3. 服务层
- `internal/service/live/live_room_service.go` - 直播间业务逻辑服务

### 4. API层
- `api/handler/live_room_handler.go` - 直播间HTTP处理器

## 功能特性

### 直播间类型
- video - 视频直播
- image - 图片直播
- vr - VR直播
- audio - 语音直播
- graphic - 图文直播

### 播出形式
- live - 直播
- vod - 点播/录播
- pseudo - 伪直播

### 直播间阶段
- formal - 正式直播
- test - 测试直播

### 显示方式
- landscape - 横屏
- portrait - 竖屏
- three_screen - 三分屏

### 观看方式
- public - 公开
- encrypted - 加密
- paid - 付费
- ticket - 购票进入
- enterprise - 企业成员观看
- custom - 自建成员观看

### 回放方式
- post_end - 结束后回放
- real_time - 实时回放
- no_playback - 结束后不回放

### 回放有效期
- unlimited - 无限制
- all_day - 全天
- partial - 部分时段

### 直播间状态
- draft - 草稿
- scheduled - 已排期
- live - 直播中
- ended - 已结束
- archived - 已归档

## API接口

### 创建直播间
```
POST /api/live/rooms
```

### 更新直播间
```
PUT /api/live/rooms
```

### 删除直播间
```
DELETE /api/live/rooms/{id}
```

### 获取直播间详情
```
GET /api/live/rooms/{id}
```

### 查询直播间列表
```
GET /api/live/rooms
```

### 开始直播
```
POST /api/live/rooms/{id}/start
```

### 结束直播
```
POST /api/live/rooms/{id}/end
```

### 更新观看人数
```
PUT /api/live/rooms/{id}/viewer-count
```

## 部署步骤

### 1. 执行数据库迁移

```bash
# 创建直播间表
mysql -u root -p < sqls/migrations/create_live_room.sql

# 配置元数据
mysql -u root -p < sqls/migrations/configure_live_room_metadata.sql
```

### 2. 注册服务和路由

在主程序中注册直播间服务和路由（需要在相应的初始化文件中添加）：

```go
// 创建服务实例
liveRoomService := live.NewLiveRoomService(db)

// 创建处理器
liveRoomHandler := handler.NewLiveRoomHandler(liveRoomService)

// 注册路由
liveGroup := router.Group("/api/live")
{
    liveGroup.POST("/rooms", liveRoomHandler.CreateRoom)
    liveGroup.PUT("/rooms", liveRoomHandler.UpdateRoom)
    liveGroup.DELETE("/rooms/:id", liveRoomHandler.DeleteRoom)
    liveGroup.GET("/rooms/:id", liveRoomHandler.GetRoom)
    liveGroup.GET("/rooms", liveRoomHandler.ListRooms)
    liveGroup.POST("/rooms/:id/start", liveRoomHandler.StartLive)
    liveGroup.POST("/rooms/:id/end", liveRoomHandler.EndLive)
    liveGroup.PUT("/rooms/:id/viewer-count", liveRoomHandler.UpdateViewerCount)
}
```

### 3. 前端集成

由于系统采用元数据驱动架构，前端可以直接使用现有的 `MetadataListView` 和 `MetadataFormView` 组件：

```
访问URL: /metadata/list?tableId=<live_room的表ID>
```

菜单已自动配置在"视频直播"模块下，名称为"直播间管理"。

## 业务逻辑说明

### 创建直播间
- 必填字段：直播间名称、直播间类型、播出形式、直播间阶段
- 默认状态为"草稿"
- 默认观看方式为"公开"
- 默认回放方式为"结束后回放"

### 更新直播间
- 直播中的直播间不允许修改类型、播出形式和流名称
- 其他字段可以随时修改

### 删除直播间
- 采用软删除（IS_ACTIVE = 'N'）
- 直播中的直播间不允许删除

### 开始直播
- 只有草稿或已排期状态的直播间可以开始直播
- 开始直播时自动记录开始时间
- 状态变更为"直播中"

### 结束直播
- 只有直播中的直播间可以结束
- 结束时自动计算直播时长
- 状态变更为"已结束"

### 观看人数统计
- 实时更新当前观看人数
- 自动记录峰值观看人数

## 扩展建议

### 1. 与腾讯云直播集成
可以在服务层添加方法，自动生成推流地址和播放地址：

```go
func (s *liveRoomService) GenerateStreamURLs(ctx context.Context, roomID uint) error {
    // 调用腾讯云API生成推流地址和播放地址
    // 更新直播间的 PUSH_URL 和 PLAY_URL 字段
}
```

### 2. 直播间成员管理
可以创建 `live_room_member` 表，管理直播间的观看成员：

```sql
CREATE TABLE live_room_member (
    ID bigint PRIMARY KEY,
    LIVE_ROOM_ID bigint NOT NULL,
    SYS_USER_ID bigint NOT NULL,
    JOIN_TIME datetime,
    LEAVE_TIME datetime,
    ...
);
```

### 3. 直播间消息/弹幕
可以创建 `live_room_message` 表，存储直播间的消息和弹幕：

```sql
CREATE TABLE live_room_message (
    ID bigint PRIMARY KEY,
    LIVE_ROOM_ID bigint NOT NULL,
    SYS_USER_ID bigint NOT NULL,
    MESSAGE_TYPE varchar(50), -- text, emoji, gift
    CONTENT text,
    SEND_TIME datetime,
    ...
);
```

### 4. 直播间统计
可以创建 `live_room_statistics` 表，记录每场直播的详细统计数据：

```sql
CREATE TABLE live_room_statistics (
    ID bigint PRIMARY KEY,
    LIVE_ROOM_ID bigint NOT NULL,
    DATE date,
    TOTAL_VIEWERS int,
    PEAK_VIEWERS int,
    AVG_WATCH_DURATION int,
    MESSAGE_COUNT int,
    ...
);
```

## 注意事项

1. 所有API接口都需要通过认证中间件，确保用户已登录
2. 公司隔离：每个公司只能看到自己的直播间
3. 权限控制：通过元数据配置的MASK字段控制增删改查权限
4. 数据验证：在服务层进行业务逻辑验证，确保数据完整性
5. 并发控制：直播状态变更时需要考虑并发场景，建议使用乐观锁或分布式锁
