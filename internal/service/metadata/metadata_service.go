package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// Service 元数据服务接口
type Service interface {
	// 获取表元数据
	GetTable(tableName string) (*entity.SysTable, error)

	// 获取表的所有字段
	GetColumns(tableID uint) ([]*entity.SysColumn, error)

	// 获取表的关联关系
	GetTableRefs(tableID uint) ([]*entity.SysTableRef, error)

	// 获取表的所有动作
	GetActions(tableID uint) ([]*entity.SysAction, error)

	// 刷新缓存
	RefreshCache() error

	// 获取元数据版本号
	GetMetadataVersion() string
}

// service 元数据服务实现
type service struct {
	repo         repository.MetadataRepository
	redisClient  *redis.Client
	cacheTTL     time.Duration
	metaVersion  string
	ctx          context.Context
}

// NewService 创建元数据服务
func NewService(repo repository.MetadataRepository, redisClient *redis.Client, cacheTTL int) Service {
	return &service{
		repo:        repo,
		redisClient: redisClient,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
		metaVersion: time.Now().Format("20060102150405"),
		ctx:         context.Background(),
	}
}

// GetTable 获取表元数据
func (s *service) GetTable(tableName string) (*entity.SysTable, error) {
	cacheKey := fmt.Sprintf("metadata:table:%s", tableName)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var table entity.SysTable
		if err := json.Unmarshal([]byte(cached), &table); err == nil {
			return &table, nil
		}
	}

	// 从数据库查询
	table, err := s.repo.GetTableByName(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 缓存结果
	data, _ := json.Marshal(table)
	s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)

	return table, nil
}

// GetColumns 获取表的所有字段
func (s *service) GetColumns(tableID uint) ([]*entity.SysColumn, error) {
	cacheKey := fmt.Sprintf("metadata:columns:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var columns []*entity.SysColumn
		if err := json.Unmarshal([]byte(cached), &columns); err == nil {
			return columns, nil
		}
	}

	// 从数据库查询
	columns, err := s.repo.GetColumnsByTableID(tableID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询字段失败", err)
	}

	// 缓存结果
	data, _ := json.Marshal(columns)
	s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)

	return columns, nil
}

// GetTableRefs 获取表的关联关系
func (s *service) GetTableRefs(tableID uint) ([]*entity.SysTableRef, error) {
	cacheKey := fmt.Sprintf("metadata:refs:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var refs []*entity.SysTableRef
		if err := json.Unmarshal([]byte(cached), &refs); err == nil {
			return refs, nil
		}
	}

	// 从数据库查询
	refs, err := s.repo.GetTableRefsByTableID(tableID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询关联关系失败", err)
	}

	// 缓存结果
	data, _ := json.Marshal(refs)
	s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)

	return refs, nil
}

// GetActions 获取表的所有动作
func (s *service) GetActions(tableID uint) ([]*entity.SysAction, error) {
	cacheKey := fmt.Sprintf("metadata:actions:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var actions []*entity.SysAction
		if err := json.Unmarshal([]byte(cached), &actions); err == nil {
			return actions, nil
		}
	}

	// 从数据库查询
	actions, err := s.repo.GetActionsByTableID(tableID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询动作失败", err)
	}

	// 缓存结果
	data, _ := json.Marshal(actions)
	s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)

	return actions, nil
}

// RefreshCache 刷新缓存
func (s *service) RefreshCache() error {
	// 删除所有元数据缓存
	pattern := "metadata:*"
	iter := s.redisClient.Scan(s.ctx, 0, pattern, 0).Iterator()

	keys := []string{}
	for iter.Next(s.ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return errors.Wrap(errors.ErrCache, "扫描缓存失败", err)
	}

	// 批量删除
	if len(keys) > 0 {
		if err := s.redisClient.Del(s.ctx, keys...).Err(); err != nil {
			return errors.Wrap(errors.ErrCache, "删除缓存失败", err)
		}
	}

	// 更新元数据版本号
	s.metaVersion = time.Now().Format("20060102150405")

	return nil
}

// GetMetadataVersion 获取元数据版本号
func (s *service) GetMetadataVersion() string {
	return s.metaVersion
}
