# 云盘字段类型修复

## 问题描述

在使用云盘功能时遇到两个 MySQL 错误：

### 错误 1: 日期时间字段错误
```
Error 1292 (22007): Incorrect datetime value: '' for column 'SHARE_EXPIRE'
```

### 错误 2: 唯一索引冲突
```
Error 1062 (23000): Duplicate entry '' for key 'cloud_folder.uk_share_code'
```

## 问题原因

### 原因 1: 日期时间字段使用 string 类型

在 Go 结构体中，`ShareExpire` 和 `ExpireTime` 字段定义为 `string` 类型：

```go
// 错误的定义
ShareExpire string `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`
```

当这些字段不设置值时，GORM 会使用字符串的零值（空字符串 `""`），而 MySQL 的 `datetime` 字段不接受空字符串，只接受 `NULL` 或有效的日期时间格式。

### 原因 2: 分享码字段使用 string 类型

`ShareCode` 字段也定义为 `string` 类型，且有唯一索引：

```go
// 错误的定义
ShareCode string `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"`
```

问题：
- 字符串零值是空字符串 `""`
- 数据库有唯一索引 `uk_share_code`
- 创建多个未分享的文件夹时，都会插入空字符串 `""`
- 违反唯一约束，导致 `Duplicate entry ''` 错误

## 解决方案

### 方案 1: 日期时间字段改为指针类型

将日期时间字段改为指针类型 `*time.Time`：

```go
// 正确的定义
ShareExpire *time.Time `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`
```

使用指针类型的好处：
- 零值是 `nil`（而不是空字符串）
- GORM 会将 `nil` 正确映射为 SQL 的 `NULL`
- 可以清晰区分"未设置"和"设置为某个时间"

### 方案 2: 分享码字段改为指针类型

将分享码字段改为指针类型 `*string`：

```go
// 正确的定义
ShareCode *string `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"`
```

使用指针类型的好处：
- 零值是 `nil`（而不是空字符串）
- GORM 会将 `nil` 正确映射为 SQL 的 `NULL`
- 多个 `NULL` 值不违反唯一索引约束
- 可以清晰区分"未分享"（`nil`）和"已分享"（非 `nil`）

## 修改的文件

### 1. `internal/model/entity/cloud_disk.go`

修改了三个结构体的字段类型：

**CloudFolder**
```go
type CloudFolder struct {
    BaseModel
    Name        string     `gorm:"column:NAME;size:255;not null" json:"name"`
    ParentID    *uint      `gorm:"column:PARENT_ID;index" json:"parentId"`
    Path        string     `gorm:"column:PATH;size:1000;not null;index" json:"path"`
    OwnerID     uint       `gorm:"column:OWNER_ID;index;not null" json:"ownerId"`
    IsPublic    string     `gorm:"column:IS_PUBLIC;size:1;default:N" json:"isPublic"`
    ShareCode   *string    `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"`        // 改为 *string
    ShareExpire *time.Time `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`          // 改为 *time.Time
    // ...
}
```

**CloudFile**
```go
type CloudFile struct {
    BaseModel
    FileName      string     `gorm:"column:FILE_NAME;size:255;not null" json:"fileName"`
    // ...
    ShareCode     *string    `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"`     // 改为 *string
    ShareExpire   *time.Time `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`       // 改为 *time.Time
    // ...
}
```

**CloudShare**
```go
type CloudShare struct {
    BaseModel
    ShareCode     string     `gorm:"column:SHARE_CODE;size:50;uniqueIndex;not null" json:"shareCode"` // 必填，保持 string
    // ...
    ExpireTime    *time.Time `gorm:"column:EXPIRE_TIME;type:datetime" json:"expireTime"`               // 改为 *time.Time
    // ...
}
```

### 2. `internal/service/cloud/cloud_service.go`

修改 `CreateShare` 方法中过期时间的处理：

```go
// 修改前
var expireTime string
if req.ExpireDays > 0 {
    expireTime = time.Now().AddDate(0, 0, req.ExpireDays).Format("2006-01-02 15:04:05")
}

// 修改后
var expireTime *time.Time
if req.ExpireDays > 0 {
    t := time.Now().AddDate(0, 0, req.ExpireDays)
    expireTime = &t
}
```

## 验证

### 1. 重新编译

```bash
cd F:\work\golang\src\github.com\sky-xhsoft\sky-server
go build ./cmd/server
```

### 2. 重启服务器

```bash
./sky-server.exe
```

### 3. 测试创建文件夹

```bash
curl -X POST http://localhost:9090/api/v1/cloud/folders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "folderName": "测试文件夹",
    "parentId": null,
    "description": "测试描述"
  }'
```

应该返回成功，不再出现 datetime 错误。

### 4. 测试创建分享

```bash
curl -X POST http://localhost:9090/api/v1/cloud/shares \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resourceType": "file",
    "resourceId": 1,
    "shareType": "public",
    "expireDays": 7
  }'
```

分享记录应该正确创建，ExpireTime 字段会被设置为 7 天后的日期。

## 最佳实践

在 Go + GORM + MySQL 中处理可选字段时：

1. **可选日期时间字段使用 `*time.Time`** 而不是 `string`
   - ✅ 类型安全
   - ✅ 自动处理 NULL
   - ✅ 避免空字符串错误

2. **有唯一索引的可选字符串字段使用 `*string`** 而不是 `string`
   - ✅ 多个 NULL 值不违反唯一约束
   - ✅ 明确区分"未设置"和"空字符串"
   - ✅ 避免 Duplicate entry 错误

3. **设置过期时间示例**
   ```go
   // 设置 7 天后过期
   t := time.Now().AddDate(0, 0, 7)
   entity.ExpireTime = &t

   // 永不过期（NULL）
   entity.ExpireTime = nil
   ```

4. **设置分享码示例**
   ```go
   // 设置分享码
   shareCode := "abc123xyz"
   entity.ShareCode = &shareCode

   // 未分享（NULL）
   entity.ShareCode = nil
   ```

5. **读取可选字段示例**
   ```go
   // 读取过期时间
   if entity.ExpireTime != nil {
       fmt.Println("过期时间:", entity.ExpireTime.Format("2006-01-02 15:04:05"))
   } else {
       fmt.Println("永不过期")
   }

   // 读取分享码
   if entity.ShareCode != nil {
       fmt.Println("分享码:", *entity.ShareCode)
   } else {
       fmt.Println("未分享")
   }
   ```

6. **JSON 序列化**
   - 指针类型会自动正确序列化
   - `nil` 会被序列化为 `null`
   - 非 nil 值会被序列化为对应的 JSON 类型

## 相关问题

### 遇到 datetime 错误时
如果遇到 `Incorrect datetime value: ''` 错误：
1. 检查字段类型是否为 `*time.Time`
2. 确保 SQL 表允许 NULL（或设置了默认值）
3. 避免使用 `string` 类型存储日期时间

### 遇到唯一索引冲突时
如果遇到 `Duplicate entry '' for key` 错误：
1. 检查有唯一索引的字段是否使用了指针类型
2. 对于可选字符串字段使用 `*string` 而不是 `string`
3. 确保数据库表允许 NULL（不要设置 `NOT NULL`）

## 数据库表定义

确保数据库表的可选字段允许 NULL：

```sql
CREATE TABLE cloud_folder (
  -- ...
  SHARE_CODE VARCHAR(50) DEFAULT NULL COMMENT '分享码',
  SHARE_EXPIRE DATETIME DEFAULT NULL COMMENT '分享过期时间',
  -- ...
  UNIQUE KEY `uk_share_code` (`SHARE_CODE`)
);
```

关键点：
- `DEFAULT NULL` 允许字段为空
- 唯一索引允许多个 NULL 值（SQL 标准）
- 不要设置 `NOT NULL` 约束在可选字段上
