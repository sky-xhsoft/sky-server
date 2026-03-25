# 基于域名的多租户实现方案

## 📋 概述

本系统支持通过域名自动识别不同公司（多租户），无需在每个请求中手动传递公司 ID。系统会根据请求的 `Host` 头自动匹配对应的公司，并在整个请求生命周期中使用该公司上下文。

## 🏗️ 实现架构

### 1. 数据库层面

#### sys_company 表结构
```sql
CREATE TABLE `sys_company`  (
  `ID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `SYS_COMPANY_ID` int UNSIGNED NULL DEFAULT NULL COMMENT '所属公司',
  `CREATE_BY` varchar(80) NULL DEFAULT NULL COMMENT '创建人',
  `CREATE_TIME` datetime NULL DEFAULT NULL COMMENT '创建时间',
  `UPDATE_BY` varchar(80) NULL DEFAULT NULL COMMENT '更新人',
  `UPDATE_TIME` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `IS_ACTIVE` char(1) NOT NULL DEFAULT 'Y' COMMENT '是否有效',
  `NAME` varchar(255) NULL DEFAULT NULL COMMENT '公司名称',
  `DOMAIN` varchar(255) NULL DEFAULT NULL COMMENT '公司域名（用于多租户识别）',
  PRIMARY KEY (`ID`) USING BTREE,
  UNIQUE INDEX `idx_domain`(`DOMAIN` ASC) USING BTREE
) ENGINE = InnoDB;
```

**关键字段**：
- `DOMAIN`: 公司绑定的域名，必须唯一
- `idx_domain`: 唯一索引，确保每个域名只能绑定一个公司

#### 数据示例
```sql
INSERT INTO sys_company (NAME, DOMAIN, IS_ACTIVE) VALUES
('公司A', 'companya.example.com', 'Y'),
('公司B', 'companyb.example.com', 'Y'),
('公司C', 'app.companyc.com', 'Y'),
('本地开发', NULL, 'Y'); -- 不使用域名识别
```

### 2. 实体模型层面

#### SysCompany 实体
```go
type SysCompany struct {
    BaseModel
    Name        string  `gorm:"column:NAME;size:255;not null" json:"name"`
    Code        string  `gorm:"column:CODE;size:50;uniqueIndex" json:"code"`
    Domain      *string `gorm:"column:DOMAIN;size:255;uniqueIndex" json:"domain"`
    Description string  `gorm:"column:DESCRIPTION;size:500" json:"description"`
    Status      string  `gorm:"column:STATUS;size:1;default:Y" json:"status"`
}
```

**注意**：`Domain` 使用指针类型 `*string`，允许 NULL 值（不使用域名识别的公司）。

### 3. 中间件层面

#### DomainTenant 中间件

文件位置：`api/middleware/domain_tenant.go`

**核心功能**：
```go
func DomainTenant(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取请求 Host
        host := c.Request.Host

        // 2. 移除端口号
        if idx := strings.Index(host, ":"); idx != -1 {
            host = host[:idx]
        }

        // 3. 跳过本地开发环境
        if host == "localhost" || strings.HasPrefix(host, "127.") {
            c.Next()
            return
        }

        // 4. 查询数据库匹配公司
        var company entity.SysCompany
        err := db.Where("DOMAIN = ? AND IS_ACTIVE = ?", host, "Y").
            First(&company).Error

        // 5. 设置到上下文
        if err == nil {
            c.Set("companyID", company.ID)
            c.Set("companyName", company.Name)
            c.Set("companyDomain", host)
        }

        c.Next()
    }
}
```

**辅助函数**：
```go
// 获取公司 ID
func GetCompanyID(c *gin.Context) *uint {
    if companyID, exists := c.Get("companyID"); exists {
        if id, ok := companyID.(uint); ok {
            return &id
        }
    }
    return nil
}

// 获取公司名称
func GetCompanyName(c *gin.Context) string {
    if companyName, exists := c.Get("companyName"); exists {
        if name, ok := companyName.(string); ok {
            return name
        }
    }
    return ""
}

// 要求必须识别公司（可选使用）
func RequireCompany() gin.HandlerFunc {
    return func(c *gin.Context) {
        if _, exists := c.Get("companyID"); !exists {
            c.JSON(403, gin.H{
                "code": 403,
                "message": "无法识别公司域名，请使用正确的域名访问",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 4. 路由配置

#### router.Setup 函数

文件位置：`api/router/router.go`

```go
func Setup(engine *gin.Engine, cfg *config.Config, jwtUtil *jwt.JWT,
           services *Services, logger *zap.Logger, db *gorm.DB) {
    // 全局中间件
    engine.Use(middleware.Logger())
    engine.Use(middleware.Recovery())
    engine.Use(middleware.CORS(cfg.CORS))

    // 域名多租户识别中间件（自动）
    engine.Use(middleware.DomainTenant(db))

    // ... 其他路由配置
}
```

## 🚀 使用方法

### 1. 数据库配置

#### 新数据库初始化
直接运行 `sqls/init.sql`，已包含 DOMAIN 字段定义。

#### 现有数据库迁移
运行迁移脚本：
```bash
mysql -u root -p your_database < sqls/migrations/add_company_domain.sql
```

#### 配置公司域名
```sql
-- 为现有公司配置域名
UPDATE sys_company SET DOMAIN = 'company1.example.com' WHERE ID = 1;
UPDATE sys_company SET DOMAIN = 'company2.example.com' WHERE ID = 2;

-- 或插入新公司
INSERT INTO sys_company (NAME, DOMAIN, IS_ACTIVE)
VALUES ('新公司', 'newcompany.example.com', 'Y');
```

### 2. DNS 配置

#### 生产环境
在 DNS 服务商处添加 A 记录或 CNAME 记录：
```
company1.example.com  →  服务器IP
company2.example.com  →  服务器IP
*.example.com         →  服务器IP (泛域名)
```

#### 开发环境
修改本地 hosts 文件：

**Windows**: `C:\Windows\System32\drivers\etc\hosts`
**Linux/Mac**: `/etc/hosts`

```
127.0.0.1  company1.local
127.0.0.1  company2.local
```

### 3. 前端配置

#### 方式一：不同域名部署
```javascript
// 公司A前端部署在 http://company1.example.com
const API_BASE = window.location.origin + '/api/v1';
// 请求会自动带上 Host: company1.example.com

// 公司B前端部署在 http://company2.example.com
const API_BASE = window.location.origin + '/api/v1';
// 请求会自动带上 Host: company2.example.com
```

#### 方式二：单页应用动态切换
```javascript
// 根据子域名确定 API 域名
const subdomain = window.location.hostname.split('.')[0];
const API_BASE = `https://${subdomain}.example.com/api/v1`;

axios.get(`${API_BASE}/data/orders`);
// 请求头会自动包含对应的 Host
```

### 4. 在业务代码中使用

#### Handler 中获取公司信息
```go
func MyHandler(c *gin.Context) {
    // 获取当前请求的公司 ID
    companyID := middleware.GetCompanyID(c)
    if companyID != nil {
        // 已识别到公司
        logger.Info("Processing request for company",
            zap.Uint("companyID", *companyID))
    } else {
        // 未识别到公司（可能是本地开发或未配置域名）
        logger.Warn("No company identified for request")
    }

    // 获取公司名称
    companyName := middleware.GetCompanyName(c)

    // 使用公司 ID 进行数据过滤
    var orders []Order
    query := db.Where("IS_ACTIVE = ?", "Y")
    if companyID != nil {
        query = query.Where("SYS_COMPANY_ID = ?", *companyID)
    }
    query.Find(&orders)
}
```

#### Service 层自动过滤（推荐）
```go
type OrderService struct {
    db *gorm.DB
}

func (s *OrderService) ListOrders(ctx context.Context, c *gin.Context) ([]Order, error) {
    var orders []Order

    // 自动添加公司过滤
    query := s.db.WithContext(ctx).Where("IS_ACTIVE = ?", "Y")

    if companyID := middleware.GetCompanyID(c); companyID != nil {
        query = query.Where("SYS_COMPANY_ID = ?", *companyID)
    }

    err := query.Find(&orders).Error
    return orders, err
}
```

#### 使用 RequireCompany 强制要求识别公司
```go
// 某些路由必须通过域名访问
v1.GET("/company-specific",
    middleware.RequireCompany(),  // 未识别公司时返回 403
    handler.MyHandler)
```

## 📝 测试示例

### 1. 测试域名识别

```bash
# 使用不同域名访问
curl -H "Host: company1.example.com" http://localhost:9090/api/v1/data/orders
curl -H "Host: company2.example.com" http://localhost:9090/api/v1/data/orders

# 使用 localhost（不会识别公司）
curl http://localhost:9090/api/v1/data/orders
```

### 2. 验证公司上下文

创建测试端点：
```go
engine.GET("/debug/company", func(c *gin.Context) {
    companyID := middleware.GetCompanyID(c)
    companyName := middleware.GetCompanyName(c)

    c.JSON(200, gin.H{
        "companyID":   companyID,
        "companyName": companyName,
        "host":        c.Request.Host,
    })
})
```

测试：
```bash
curl -H "Host: company1.example.com" http://localhost:9090/debug/company
# 响应：
# {
#   "companyID": 1,
#   "companyName": "公司A",
#   "host": "company1.example.com"
# }
```

## ⚙️ 高级配置

### 1. 泛域名支持

如果需要支持 `*.example.com` 的任意子域名：

```sql
-- 方式1：为每个子域名单独配置
INSERT INTO sys_company (NAME, DOMAIN) VALUES
('客户1', 'customer1.example.com'),
('客户2', 'customer2.example.com');

-- 方式2：使用模糊匹配（需要修改中间件）
-- 在中间件中使用 LIKE 查询
db.Where("DOMAIN LIKE ? AND IS_ACTIVE = ?", "%"+subdomain+".example.com", "Y")
```

### 2. 多域名绑定同一公司

如果一个公司有多个域名，建议使用关联表：

```sql
CREATE TABLE sys_company_domains (
  ID BIGINT PRIMARY KEY AUTO_INCREMENT,
  SYS_COMPANY_ID INT UNSIGNED NOT NULL,
  DOMAIN VARCHAR(255) NOT NULL,
  IS_PRIMARY CHAR(1) DEFAULT 'N',
  UNIQUE KEY idx_domain (DOMAIN),
  FOREIGN KEY (SYS_COMPANY_ID) REFERENCES sys_company(ID)
);
```

### 3. 域名白名单验证

在中间件中添加额外验证：
```go
// 只允许特定后缀的域名
allowedSuffixes := [".example.com", ".myapp.com"]
allowed := false
for _, suffix := range allowedSuffixes {
    if strings.HasSuffix(host, suffix) {
        allowed = true
        break
    }
}
if !allowed {
    c.AbortWithStatusJSON(403, gin.H{"error": "域名不在白名单内"})
    return
}
```

## 🔒 安全注意事项

1. **HTTPS 强制**：生产环境必须使用 HTTPS，防止域名欺骗
2. **Host 头验证**：防止 Host 头注入攻击
3. **CORS 配置**：正确配置 CORS 允许的域名
4. **域名所有权验证**：确保只有经过验证的域名才能绑定公司

## 🐛 故障排查

### 问题 1：域名识别失败

**症状**：请求未识别到公司

**检查步骤**：
1. 确认 DNS 配置正确
2. 检查数据库中 DOMAIN 字段是否正确配置
3. 验证请求的 Host 头：`curl -v -H "Host: xxx.com" http://...`
4. 检查域名是否包含端口号（中间件会自动移除）
5. 确认公司状态为 `IS_ACTIVE = 'Y'`

### 问题 2：本地开发环境识别失败

**解决方案**：
- 使用 `localhost` 或 `127.0.0.1`（不会尝试识别）
- 或在 hosts 文件中配置测试域名
- 或直接使用 `curl -H "Host: test.local"` 模拟

### 问题 3：跨域问题

**原因**：不同域名被视为不同源

**解决**：
```go
// 在 CORS 中间件配置允许的域名
engine.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "https://company1.example.com",
        "https://company2.example.com",
        "https://*.example.com", // 泛域名
    },
    AllowCredentials: true,
}))
```

## 📊 性能优化

### 1. 域名查询缓存

```go
var domainCache = cache.New(5*time.Minute, 10*time.Minute)

func DomainTenant(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        host := extractHost(c.Request.Host)

        // 先查缓存
        if cached, found := domainCache.Get(host); found {
            if company, ok := cached.(*entity.SysCompany); ok {
                c.Set("companyID", company.ID)
                c.Set("companyName", company.Name)
                c.Next()
                return
            }
        }

        // 缓存未命中，查询数据库
        var company entity.SysCompany
        if err := db.Where("DOMAIN = ?", host).First(&company).Error; err == nil {
            // 放入缓存
            domainCache.Set(host, &company, cache.DefaultExpiration)
            c.Set("companyID", company.ID)
            c.Set("companyName", company.Name)
        }

        c.Next()
    }
}
```

### 2. 数据库索引

确保 DOMAIN 字段有唯一索引（已在表结构中定义）：
```sql
CREATE UNIQUE INDEX idx_domain ON sys_company(DOMAIN);
```

## 📚 相关文档

- [多租户架构设计](./multi-tenancy-architecture.md)
- [数据隔离策略](./data-isolation.md)
- [API 接口文档](./api-documentation.md)

## ✅ 实现清单

- ✅ 数据库表添加 DOMAIN 字段
- ✅ 创建唯一索引 idx_domain
- ✅ 更新 SysCompany 实体模型
- ✅ 实现 DomainTenant 中间件
- ✅ 集成到路由配置
- ✅ 提供辅助函数获取公司信息
- ✅ 创建数据库迁移脚本
- ✅ 编写完整文档
- ⏳ 添加单元测试（待实现）
- ⏳ 添加域名管理 UI（待实现）

## 🔄 后续优化建议

1. **域名管理界面**：提供 UI 界面管理公司域名绑定
2. **域名验证**：在绑定前验证域名所有权（如 DNS TXT 记录）
3. **域名历史记录**：记录域名变更历史
4. **性能监控**：监控域名识别的性能和成功率
5. **多级域名支持**：支持 `app.sub.company.com` 等多级域名
6. **自定义路由规则**：允许不同公司使用不同的路由前缀
