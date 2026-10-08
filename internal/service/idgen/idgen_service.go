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

// raiseCounterScript 计数器小于给定值时抬升到该值（并发安全），返回抬升后的值
var raiseCounterScript = redis.NewScript(`
	local key = KEYS[1]
	local target = tonumber(ARGV[1])
	local current = tonumber(redis.call('GET', key) or 0)
	if current < target then
		redis.call('SET', key, target)
		return target
	end
	return current
`)

// maxIDGenAttempts 取号与已有记录冲突时的最大重试次数
const maxIDGenAttempts = 5

// GetNextID 获取下一个ID
// 流程:
// 1. 从Redis计数器原子递增取号
// 2. 确认该ID在表中未被占用：同一张表也会被直接用数据库自增写入（如 /live/rooms、内部逻辑），
//    这些写入不推进计数器，取到的号可能已存在，此时把计数器抬升到表中最大ID后重新取号
func (s *service) GetNextID(ctx context.Context, tableName string) (uint, error) {
	cacheKey := fmt.Sprintf("table:maxid:%s", tableName)

	for attempt := 0; attempt < maxIDGenAttempts; attempt++ {
		newID, err := s.redisClient.Incr(ctx, cacheKey).Result()
		if err != nil {
			return 0, fmt.Errorf("Redis递增失败: %w", err)
		}

		taken, err := s.idExists(ctx, tableName, uint(newID))
		if err != nil {
			return 0, err
		}
		if !taken {
			// 设置过期时间（7天），过期后下次取号会按表中最大ID重新同步
			s.redisClient.Expire(ctx, cacheKey, 7*24*time.Hour)
			return uint(newID), nil
		}

		maxID, err := s.getMaxIDFromDB(ctx, tableName)
		if err != nil {
			return 0, err
		}
		if err := raiseCounterScript.Run(ctx, s.redisClient, []string{cacheKey}, maxID).Err(); err != nil {
			return 0, fmt.Errorf("同步Redis计数器失败: %w", err)
		}
	}
	return 0, fmt.Errorf("表 %s 取号连续冲突，请稍后重试", tableName)
}

// idExists 判断表中是否已存在该ID（含已停用的记录）
func (s *service) idExists(ctx context.Context, tableName string, id uint) (bool, error) {
	var n int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE ID = ?", tableName)
	if err := s.db.WithContext(ctx).Raw(query, id).Scan(&n).Error; err != nil {
		return false, fmt.Errorf("检查ID是否占用失败: %w", err)
	}
	return n > 0, nil
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
