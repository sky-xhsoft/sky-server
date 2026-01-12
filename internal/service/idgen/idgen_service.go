package idgen

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Service ID生成服务接口
type Service interface {
	// GetNextID 获取下一个ID
	GetNextID(ctx context.Context, tableName string) (uint, error)

	// ResetCache 重置某个表的ID缓存（用于数据导入等场景）
	ResetCache(ctx context.Context, tableName string) error
}

// service ID生成服务实现
type service struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewService 创建ID生成服务
func NewService(db *gorm.DB, redisClient *redis.Client) Service {
	return &service{
		db:          db,
		redisClient: redisClient,
	}
}

// GetNextID 获取下一个ID
// 流程:
// 1. 尝试从Redis缓存中原子递增获取ID
// 2. 如果Redis中不存在该表的键，从数据库查询最大ID并初始化缓存
// 3. 返回递增后的ID
func (s *service) GetNextID(ctx context.Context, tableName string) (uint, error) {
	cacheKey := fmt.Sprintf("table:maxid:%s", tableName)

	// 1. 尝试从Redis原子递增获取ID
	newID, err := s.redisClient.Incr(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("Redis递增失败: %w", err)
	}

	// 2. 如果是第一次访问（newID == 1），需要从数据库初始化
	if newID == 1 {
		// 从数据库查询当前最大ID
		maxID, err := s.getMaxIDFromDB(ctx, tableName)
		if err != nil {
			return 0, err
		}

		// 如果数据库中有数据，重置Redis缓存为最大ID+1
		if maxID > 0 {
			// 使用SET命令重置缓存（只在不存在或小于maxID+1时设置）
			// 为了保证并发安全，使用Lua脚本原子操作
			script := redis.NewScript(`
				local key = KEYS[1]
				local maxID = tonumber(ARGV[1])
				local current = tonumber(redis.call('GET', key) or 0)
				if current < maxID then
					redis.call('SET', key, maxID)
					return maxID
				end
				return current
			`)

			result, err := script.Run(ctx, s.redisClient, []string{cacheKey}, maxID+1).Int64()
			if err != nil {
				return 0, fmt.Errorf("初始化Redis缓存失败: %w", err)
			}

			// 再次递增获取新ID
			newID, err = s.redisClient.Incr(ctx, cacheKey).Result()
			if err != nil {
				return 0, fmt.Errorf("Redis递增失败: %w", err)
			}

			// 设置过期时间（7天，防止缓存无限增长）
			s.redisClient.Expire(ctx, cacheKey, 7*24*time.Hour)

			return uint(result), nil
		}

		// 如果数据库为空，newID=1就是正确的
		// 设置过期时间
		s.redisClient.Expire(ctx, cacheKey, 7*24*time.Hour)
	}

	return uint(newID), nil
}

// getMaxIDFromDB 从数据库查询表的最大ID
func (s *service) getMaxIDFromDB(ctx context.Context, tableName string) (uint, error) {
	var maxID uint

	// 查询最大ID
	query := fmt.Sprintf("SELECT IFNULL(MAX(ID), 0) as max_id FROM %s", tableName)
	err := s.db.WithContext(ctx).Raw(query).Scan(&maxID).Error
	if err != nil {
		return 0, fmt.Errorf("查询最大ID失败: %w", err)
	}

	return maxID, nil
}

// ResetCache 重置某个表的ID缓存（用于数据导入等场景）
func (s *service) ResetCache(ctx context.Context, tableName string) error {
	cacheKey := fmt.Sprintf("table:maxid:%s", tableName)

	// 从数据库查询最大ID
	maxID, err := s.getMaxIDFromDB(ctx, tableName)
	if err != nil {
		return err
	}

	// 重置Redis缓存
	err = s.redisClient.Set(ctx, cacheKey, maxID, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("重置Redis缓存失败: %w", err)
	}

	return nil
}
