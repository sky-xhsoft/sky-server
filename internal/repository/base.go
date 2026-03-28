package repository

import (
	"context"
	"reflect"
	"time"

	"gorm.io/gorm"
)

// QueryOptions 查询选项
type QueryOptions struct {
	Fields     []string      // 查询字段列表
	OrderBy    string        // 排序字段
	Order      string        // 排序方向: "asc" or "desc"
	Preloads   []string      // 预加载关系
	WithCount  bool          // 是否返回总数 (用于分页)
	Distinct   bool          // 是否去重
}

// PageQuery 分页查询参数
type PageQuery struct {
	Page      int         `json:"page" binding:"required,min=1"`           // 页码
	PageSize  int         `json:"pageSize" binding:"required,min=1,max=100"` // 每页大小
	Keyword   string      `json:"keyword"`                                // 搜索关键字
	OrderBy   string      `json:"orderBy"`                                // 排序字段
	Order     string      `json:"order"`                                  // 排序方向: "asc" or "desc"
	Filters   interface{} `json:"filters"`                                // 筛选条件
}

// PageResult 统一分页查询结果
type PageResult struct {
	Total    int64       `json:"total"`                                    // 总数量
	Page     int         `json:"page"`                                     // 当前页码
	PageSize int         `json:"pageSize"`                                 // 每页大小
	Data     interface{} `json:"data"`                                     // 数据列表
}

// BaseRepository 基础Repository接口
type BaseRepository interface {
	GetByID(ctx context.Context, entity interface{}, id uint) error
	GetOne(ctx context.Context, entity interface{}, opts ...interface{}) error
	GetList(ctx context.Context, entity interface{}, opts ...interface{}) error
	GetPage(ctx context.Context, query *PageQuery, entity interface{}) (*PageResult, error)
	Create(ctx context.Context, entity interface{}) error
	Update(ctx context.Context, entity interface{}, updates interface{}) error
	Delete(ctx context.Context, entity interface{}, id uint) error
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// Repository 仓库接口扩展
type Repository interface {
	GetDB() *gorm.DB
	GetContext() context.Context
	GetTable(entity interface{}) string
}

// BaseRepositoryImpl 基础Repository实现
type BaseRepositoryImpl struct {
	db     *gorm.DB
	ctx    context.Context
	logger Logger
	cache  Cache
	ttl    time.Duration
}

// Logger 仓库日志接口
type Logger interface {
	Debug(ctx context.Context, format string, args ...interface{})
	Info(ctx context.Context, format string, args ...interface{})
	Warn(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

// Cache 仓库缓存接口
type Cache interface {
	Get(key string, obj interface{}) error
	Set(key string, obj interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
}

// NewBaseRepository 创建基础Repository实例
func NewBaseRepository(db *gorm.DB, logger Logger, cache Cache, ttl time.Duration) *BaseRepositoryImpl {
	if logger == nil {
		logger = &NoopLogger{}
	}
	if cache == nil {
		cache = &NoopCache{}
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &BaseRepositoryImpl{
		db:     db,
		ctx:    context.Background(),
		logger: logger,
		cache:  cache,
		ttl:    ttl,
	}
}

// NoopLogger 空日志实现
type NoopLogger struct{}

func (l *NoopLogger) Debug(ctx context.Context, format string, args ...interface{}) {}
func (l *NoopLogger) Info(ctx context.Context, format string, args ...interface{})  {}
func (l *NoopLogger) Warn(ctx context.Context, format string, args ...interface{})  {}
func (l *NoopLogger) Error(ctx context.Context, format string, args ...interface{}) {}

// NoopCache 空缓存实现
type NoopCache struct{}

func (c *NoopCache) Get(key string, obj interface{}) error          { return gorm.ErrRecordNotFound }
func (c *NoopCache) Set(key string, obj interface{}, ttl time.Duration) error { return nil }
func (c *NoopCache) Delete(key string) error                        { return nil }
func (c *NoopCache) Exists(key string) (bool, error)                 { return false, nil }

// GetDB 获取数据库实例
func (r *BaseRepositoryImpl) GetDB() *gorm.DB {
	return r.db
}

// GetContext 获取上下文
func (r *BaseRepositoryImpl) GetContext() context.Context {
	return r.ctx
}

// GetTable 获取表名
func (r *BaseRepositoryImpl) GetTable(entity interface{}) string {
	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}

	typeInfo := val.Type()
	if tableName, ok := typeInfo.FieldByName("TableName"); ok && tableName.IsExported() {
		return tableName.Tag.Get("gorm")
	}

	if method, exists := typeInfo.MethodByName("TableName"); exists {
		return method.Type.Out(0).Name()
	}

	return typeInfo.Name()
}

// Transaction 在事务中执行操作
func (r *BaseRepositoryImpl) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// GetByID 根据ID获取实体
func (r *BaseRepositoryImpl) GetByID(ctx context.Context, entity interface{}, id uint) error {
	return r.db.WithContext(ctx).First(entity, id).Error
}

// GetOne 获取单个实体
func (r *BaseRepositoryImpl) GetOne(ctx context.Context, entity interface{}, opts ...interface{}) error {
	db := r.db.WithContext(ctx)
	for _, opt := range opts {
		if cond, ok := opt.(map[string]interface{}); ok {
			db = db.Where(cond)
		}
	}
	return db.First(entity).Error
}

// GetList 获取实体列表
func (r *BaseRepositoryImpl) GetList(ctx context.Context, entity interface{}, opts ...interface{}) error {
	db := r.db.WithContext(ctx)
	for _, opt := range opts {
		if cond, ok := opt.(map[string]interface{}); ok {
			db = db.Where(cond)
		}
	}
	return db.Find(entity).Error
}

// GetPage 分页查询（使用 PageQuery）
func (r *BaseRepositoryImpl) GetPage(ctx context.Context, query *PageQuery, entity interface{}) (*PageResult, error) {
	pageRepo := NewPageRepository(r.db)
	return pageRepo.GetPage(ctx, query, entity)
}

// Create 创建实体
func (r *BaseRepositoryImpl) Create(ctx context.Context, entity interface{}) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update 更新实体
func (r *BaseRepositoryImpl) Update(ctx context.Context, entity interface{}, updates interface{}) error {
	return r.db.WithContext(ctx).Model(entity).Updates(updates).Error
}

// Delete 删除实体
func (r *BaseRepositoryImpl) Delete(ctx context.Context, entity interface{}, id uint) error {
	return r.db.WithContext(ctx).Delete(entity, id).Error
}
