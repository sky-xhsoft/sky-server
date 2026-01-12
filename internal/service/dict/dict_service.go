package dict

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// Service 数据字典服务接口
type Service interface {
	// 获取字典项列表
	GetDictItems(dictID uint) ([]*entity.SysDictItem, error)

	// 根据字典名称获取项
	GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error)

	// 获取默认值
	GetDefaultValue(dictName string) (string, error)

	// 刷新字典缓存
	RefreshDictCache() error
}

// service 数据字典服务实现
type service struct {
	repo        repository.DictRepository
	redisClient *redis.Client
	cacheTTL    time.Duration
	ctx         context.Context
}

// NewService 创建数据字典服务
func NewService(repo repository.DictRepository, redisClient *redis.Client, cacheTTL int) Service {
	return &service{
		repo:        repo,
		redisClient: redisClient,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
		ctx:         context.Background(),
	}
}

// GetDictItems 获取字典项列表
func (s *service) GetDictItems(dictID uint) ([]*entity.SysDictItem, error) {
	cacheKey := fmt.Sprintf("dict:items:%d", dictID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var items []*entity.SysDictItem
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	// 从数据库查询
	items, err := s.repo.GetDictItems(dictID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询字典项失败", err)
	}

	// 缓存结果
	data, _ := json.Marshal(items)
	s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)

	return items, nil
}

// GetDictItemsByName 根据字典名称获取项
func (s *service) GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error) {
	cacheKey := fmt.Sprintf("dict:items:name:%s", dictName)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var items []*entity.SysDictItem
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	// 从数据库查询
	items, err := s.repo.GetDictItemsByName(dictName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "字典不存在", err)
	}

	// 缓存结果
	data, _ := json.Marshal(items)
	s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)

	return items, nil
}

// GetDefaultValue 获取默认值
func (s *service) GetDefaultValue(dictName string) (string, error) {
	items, err := s.GetDictItemsByName(dictName)
	if err != nil {
		return "", err
	}

	// 查找默认值
	for _, item := range items {
		if item.IsDefaultValue == "Y" {
			return item.Value, nil
		}
	}

	// 如果没有默认值，返回第一个
	if len(items) > 0 {
		return items[0].Value, nil
	}

	return "", nil
}

// RefreshDictCache 刷新字典缓存
func (s *service) RefreshDictCache() error {
	// 删除所有字典缓存
	pattern := "dict:*"
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

	return nil
}
