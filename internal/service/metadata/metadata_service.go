package metadata

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

// Service 元数据服务接口
type Service interface {
	// 获取表元数据（通过表名）
	GetTable(tableName string) (*entity.SysTable, error)

	// 获取表元数据（通过ID）
	GetTableByID(tableID uint) (*entity.SysTable, error)

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

	// 获取外键选项列表
	GetForeignKeyOptions(tableID uint, columnID *uint, search string, page, pageSize int) ([]map[string]interface{}, int64, error)

	// 获取外键显示值
	GetForeignKeyDisplayValue(tableID uint, value string, columnID *uint) (string, error)
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

// GetTable 获取表元数据（通过表名）
func (s *service) GetTable(tableName string) (*entity.SysTable, error) {
	cacheKey := fmt.Sprintf("metadata:table:name:%s", tableName)

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

// GetTableByID 获取表元数据（通过ID）
func (s *service) GetTableByID(tableID uint) (*entity.SysTable, error) {
	cacheKey := fmt.Sprintf("metadata:table:id:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var table entity.SysTable
		if err := json.Unmarshal([]byte(cached), &table); err == nil {
			return &table, nil
		}
	}

	// 从数据库查询
	table, err := s.repo.GetTableByID(tableID)
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

// GetForeignKeyOptions 获取外键选项列表
func (s *service) GetForeignKeyOptions(tableID uint, columnID *uint, search string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 1. 获取目标表的元数据
	table, err := s.GetTableByID(tableID)
	if err != nil {
		return nil, 0, errors.Wrap(errors.ErrNotFound, "表不存在", err)
	}

	// 2. 确定显示字段
	var displayColumn *entity.SysColumn
	if columnID != nil {
		// 使用指定的字段
		columns, err := s.GetColumns(tableID)
		if err != nil {
			return nil, 0, err
		}
		for _, col := range columns {
			if col.ID == *columnID {
				displayColumn = col
				break
			}
		}
	} else {
		// 使用表的 DK（显示键）字段
		columns, err := s.GetColumns(tableID)
		if err != nil {
			return nil, 0, err
		}
		for _, col := range columns {
			if col.IsDK == "Y" {
				displayColumn = col
				break
			}
		}
		// 如果没有 DK 字段，使用第一个 VARCHAR 类型的字段
		if displayColumn == nil {
			for _, col := range columns {
				if col.ColType == "varchar" || col.ColType == "char" {
					displayColumn = col
					break
				}
			}
		}
	}

	if displayColumn == nil {
		return nil, 0, errors.New(errors.ErrNotFound, "未找到显示字段")
	}

	// 3. 构建查询（使用原生 SQL 查询）
	offset := (page - 1) * pageSize
	options := []map[string]interface{}{}
	var total int64

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE IS_ACTIVE = 'Y'", table.Name)
	if search != "" {
		countSQL += fmt.Sprintf(" AND %s LIKE ?", displayColumn.DBName)
		if err := s.repo.CountBySql(countSQL, "%"+search+"%").Scan(&total).Error; err != nil {
			return nil, 0, errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
		}
	} else {
		if err := s.repo.CountBySql(countSQL).Scan(&total).Error; err != nil {
			return nil, 0, errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
		}
	}

	// 查询数据
	dataSQL := fmt.Sprintf("SELECT ID as value, %s as label FROM %s WHERE IS_ACTIVE = 'Y'", displayColumn.DBName, table.Name)
	if search != "" {
		dataSQL += fmt.Sprintf(" AND %s LIKE ?", displayColumn.DBName)
		dataSQL += fmt.Sprintf(" ORDER BY %s LIMIT ? OFFSET ?", displayColumn.DBName)
		if err := s.repo.RawQuery(dataSQL, "%"+search+"%", pageSize, offset).Scan(&options).Error; err != nil {
			return nil, 0, errors.Wrap(errors.ErrDatabase, "查询数据失败", err)
		}
	} else {
		dataSQL += fmt.Sprintf(" ORDER BY %s LIMIT ? OFFSET ?", displayColumn.DBName)
		if err := s.repo.RawQuery(dataSQL, pageSize, offset).Scan(&options).Error; err != nil {
			return nil, 0, errors.Wrap(errors.ErrDatabase, "查询数据失败", err)
		}
	}

	return options, total, nil
}

// GetForeignKeyDisplayValue 获取外键显示值
func (s *service) GetForeignKeyDisplayValue(tableID uint, value string, columnID *uint) (string, error) {
	// 1. 获取目标表的元数据
	table, err := s.GetTableByID(tableID)
	if err != nil {
		return "", errors.Wrap(errors.ErrNotFound, "表不存在", err)
	}

	// 2. 确定显示字段（逻辑同上）
	var displayColumn *entity.SysColumn
	if columnID != nil {
		columns, err := s.GetColumns(tableID)
		if err != nil {
			return "", err
		}
		for _, col := range columns {
			if col.ID == *columnID {
				displayColumn = col
				break
			}
		}
	} else {
		columns, err := s.GetColumns(tableID)
		if err != nil {
			return "", err
		}
		for _, col := range columns {
			if col.IsDK == "Y" {
				displayColumn = col
				break
			}
		}
		if displayColumn == nil {
			for _, col := range columns {
				if col.ColType == "varchar" || col.ColType == "char" {
					displayColumn = col
					break
				}
			}
		}
	}

	if displayColumn == nil {
		return "", errors.New(errors.ErrNotFound, "未找到显示字段")
	}

	// 3. 查询显示值
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE ID = ? AND IS_ACTIVE = 'Y'", displayColumn.DBName, table.Name)
	var displayValue string
	if err := s.repo.RawQuery(querySQL, value).Scan(&displayValue).Error; err != nil {
		return "", errors.Wrap(errors.ErrDatabase, "查询显示值失败", err)
	}

	return displayValue, nil
}
