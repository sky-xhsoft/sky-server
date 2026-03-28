package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/cache"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CacheableRepository 可缓存的Repository接口
type CacheableRepository interface {
	GetByID(ctx context.Context, entity interface{}, id uint) error
	GetAll(ctx context.Context, entity interface{}) error
	GetOne(ctx context.Context, entity interface{}, opts ...interface{}) error
}

// CacheRepository 缓存装饰器
type CacheRepository struct {
	repository CacheableRepository
	cache      cache.Cache
	ttl        time.Duration
}

// NewCacheRepository 创建缓存装饰器
func NewCacheRepository(
	repository CacheableRepository,
	cache cache.Cache,
	ttl time.Duration,
) *CacheRepository {
	return &CacheRepository{
		repository: repository,
		cache:      cache,
		ttl:        ttl,
	}
}

// GetByID 获取单条记录（带缓存）
func (c *CacheRepository) GetByID(ctx context.Context, entity interface{}, id uint) error {
	entityType := reflect.TypeOf(entity)
	for entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	cacheKey := fmt.Sprintf("cache:%s:%d", entityType.Name(), id)

	// 尝试从缓存中获取
	var cacheData []byte
	if err := c.cache.Get(ctx, cacheKey, &cacheData); err == nil {
		if err := json.Unmarshal(cacheData, entity); err == nil {
			return nil
		}
		logger.Warn("Failed to unmarshal cache data", zap.String("key", cacheKey), zap.Error(err))
	}

	// 从数据库获取
	if err := c.repository.GetByID(ctx, entity, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			// 避免缓存穿透，设置空值
			c.cache.Set(ctx, cacheKey, "null", time.Minute*5)
		}
		return err
	}

	// 缓存到Redis
	if data, err := json.Marshal(entity); err == nil {
		_ = c.cache.Set(ctx, cacheKey, data, c.ttl)
	}

	return nil
}

// GetAll 获取所有记录（带缓存）
func (c *CacheRepository) GetAll(ctx context.Context, entity interface{}) error {
	entityType := reflect.TypeOf(entity)
	for entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Slice {
		entityType = entityType.Elem()
	}
	cacheKey := fmt.Sprintf("cache:%s:all", entityType.Name())

	// 尝试从缓存中获取
	var cacheData []byte
	if err := c.cache.Get(ctx, cacheKey, &cacheData); err == nil {
		if err := json.Unmarshal(cacheData, entity); err == nil {
			return nil
		}
		logger.Warn("Failed to unmarshal cache data", zap.String("key", cacheKey), zap.Error(err))
	}

	// 从数据库获取
	if err := c.repository.GetAll(ctx, entity); err != nil {
		return err
	}

	// 缓存到Redis
	if data, err := json.Marshal(entity); err == nil {
		_ = c.cache.Set(ctx, cacheKey, data, c.ttl)
	}

	return nil
}

// GetOne 获取单条记录（带缓存）
func (c *CacheRepository) GetOne(ctx context.Context, entity interface{}, opts ...interface{}) error {
	entityType := reflect.TypeOf(entity)
	for entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// 构建查询条件缓存键
	conds := make([]string, 0)
	for _, opt := range opts {
		if condMap, ok := opt.(map[string]interface{}); ok {
			for k, v := range condMap {
				conds = append(conds, fmt.Sprintf("%s=%v", k, v))
			}
		}
	}
	cacheKey := fmt.Sprintf("cache:%s:one:%s", entityType.Name(), strings.Join(conds, "-"))

	// 尝试从缓存中获取
	var cacheData []byte
	if err := c.cache.Get(ctx, cacheKey, &cacheData); err == nil {
		if err := json.Unmarshal(cacheData, entity); err == nil {
			return nil
		}
		logger.Warn("Failed to unmarshal cache data", zap.String("key", cacheKey), zap.Error(err))
	}

	// 从数据库获取
	if err := c.repository.GetOne(ctx, entity, opts...); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.cache.Set(ctx, cacheKey, "null", time.Minute*5)
		}
		return err
	}

	// 缓存到Redis
	if data, err := json.Marshal(entity); err == nil {
		_ = c.cache.Set(ctx, cacheKey, data, c.ttl)
	}

	return nil
}

// InvalidateCache 使缓存失效
func (c *CacheRepository) InvalidateCache(ctx context.Context, entityType string, id ...uint) {
	var keys []string

	if entityType == "" {
		logger.Warn("Entity type is required for cache invalidation")
		return
	}

	if len(id) > 0 {
		for _, i := range id {
			keys = append(keys, fmt.Sprintf("cache:%s:%d", entityType, i))
		}
	} else {
		keys = append(keys, fmt.Sprintf("cache:%s:all", entityType))
		keys = append(keys, fmt.Sprintf("cache:%s:one:*", entityType))
		keys = append(keys, fmt.Sprintf("cache:%s:page:*", entityType))
	}

	for _, key := range keys {
		if err := c.cache.Delete(ctx, key); err != nil {
			logger.Warn("Failed to delete cache", zap.String("key", key), zap.Error(err))
		}
	}
}

// CacheKeyGenerator 缓存键生成器
type CacheKeyGenerator struct{}

func NewCacheKeyGenerator() *CacheKeyGenerator {
	return &CacheKeyGenerator{}
}

func (g *CacheKeyGenerator) GenerateKey(entityType string, id uint) string {
	return fmt.Sprintf("cache:%s:%d", entityType, id)
}

func (g *CacheKeyGenerator) GenerateListKey(entityType string, params ...interface{}) string {
	parts := []string{entityType, "list"}
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("%v", p))
	}
	return fmt.Sprintf("cache:%s", strings.Join(parts, ":"))
}

func (g *CacheKeyGenerator) GeneratePageKey(entityType string, page, pageSize int, params ...interface{}) string {
	parts := []string{entityType, "page", fmt.Sprintf("page:%d", page), fmt.Sprintf("size:%d", pageSize)}
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("%v", p))
	}
	return fmt.Sprintf("cache:%s", strings.Join(parts, ":"))
}

func (g *CacheKeyGenerator) GenerateCustomKey(entityType string, method string, params ...interface{}) string {
	parts := []string{entityType, method}
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("%v", p))
	}
	return fmt.Sprintf("cache:%s", strings.Join(parts, ":"))
}

// CacheManager 缓存管理器
type CacheManager struct {
	caches map[string]cache.Cache
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]cache.Cache),
	}
}

func (cm *CacheManager) Register(name string, c cache.Cache) {
	cm.caches[name] = c
}

func (cm *CacheManager) Get(name string) (cache.Cache, bool) {
	c, ok := cm.caches[name]
	return c, ok
}

func (cm *CacheManager) GetDefault() (cache.Cache, bool) {
	c, ok := cm.caches["default"]
	return c, ok
}
