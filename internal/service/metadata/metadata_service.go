package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	GetForeignKeyOptions(tableID uint, columnID *uint, search string, page, pageSize int, userID uint) ([]map[string]interface{}, int64, string, error)

	// 获取外键显示值
	GetForeignKeyDisplayValue(tableID uint, value string, columnID *uint, userID uint) (string, error)
}

// service 元数据服务实现
type service struct {
	repo         repository.MetadataRepository
	userRepo     repository.UserRepository
	redisClient  *redis.Client
	cacheTTL     time.Duration
	metaVersion  string
	ctx          context.Context
}

// NewService 创建元数据服务
func NewService(repo repository.MetadataRepository, userRepo repository.UserRepository, redisClient *redis.Client, cacheTTL int) Service {
	return &service{
		repo:        repo,
		userRepo:    userRepo,
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
	var columns []*entity.SysColumn
	cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &columns); err == nil {
			// 从缓存获取成功，但仍需填充 RefTableIsDropdown（因为缓存中可能没有）
			for _, col := range columns {
				if col.SetValueType == "fk" && col.RefTableID != nil && col.RefTableIsDropdown == "" {
					refTable, err := s.GetTableByID(*col.RefTableID)
					if err == nil && refTable != nil {
						col.RefTableIsDropdown = refTable.IsDropdown
					}
				}
			}
			return columns, nil
		}
	}

	// 从数据库查询
	columns, err = s.repo.GetColumnsByTableID(tableID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询字段失败", err)
	}

	// 为 FK 字段填充引用表的 IS_DROPDOWN 信息
	for _, col := range columns {
		if col.SetValueType == "fk" && col.RefTableID != nil {
			refTable, err := s.GetTableByID(*col.RefTableID)
			if err == nil && refTable != nil {
				// 将引用表的 IS_DROPDOWN 存储到列的扩展字段
				col.RefTableIsDropdown = refTable.IsDropdown
			}
		}
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
func (s *service) GetForeignKeyOptions(tableID uint, columnID *uint, search string, page, pageSize int, userID uint) ([]map[string]interface{}, int64, string, error) {
	// 1. 获取目标表的元数据
	table, err := s.GetTableByID(tableID)
	if err != nil {
		return nil, 0, "", errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 2. 确定显示字段
	var displayColumn *entity.SysColumn
	columns, err := s.GetColumns(tableID)
	if err != nil {
		return nil, 0, "", err
	}

	if columnID != nil {
		// 使用指定的字段
		for _, col := range columns {
			if col.ID == *columnID {
				displayColumn = col
				break
			}
		}
	}

	// 如果没有找到指定字段，使用 DK 字段
	if displayColumn == nil {
		for _, col := range columns {
			if col.IsDK == "Y" {
				displayColumn = col
				break
			}
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

	if displayColumn == nil {
		return nil, 0, "", errors.New(errors.ErrResourceNotFound, "未找到显示字段")
	}

	// 3. 获取用户信息以获取公司ID
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, 0, "", errors.Wrap(errors.ErrInternal, "获取用户信息失败", err)
	}

	// 4. 构建查询（使用原生 SQL 查询）
	offset := (page - 1) * pageSize
	options := []map[string]interface{}{}
	var total int64

	// 构建基础 WHERE 条件
	baseWhere := "IS_ACTIVE = 'Y'"
	whereParams := []interface{}{}

	fmt.Printf("[DEBUG] GetForeignKeyOptions: table.Name=%s, userID=%d, user.SysCompanyID=%d\n", table.Name, userID, user.SysCompanyID)

	// 特殊处理：sys_company 表本身没有 SYS_COMPANY_ID 字段
	// 用户只能看到自己所属的公司
	if strings.EqualFold(table.Name, "sys_company") {
		baseWhere += " AND ID = ?"
		whereParams = append(whereParams, user.SysCompanyID)
		fmt.Printf("[DEBUG] sys_company 特殊处理：添加 ID 过滤\n")
	} else {
		// 检查目标表是否有 SYS_COMPANY_ID 字段
		hasCompanyColumn := false
		for _, col := range columns {
			if col.DbName == "SYS_COMPANY_ID" {
				hasCompanyColumn = true
				break
			}
		}
		// 如果目标表有 SYS_COMPANY_ID 字段，添加公司过滤
		if hasCompanyColumn {
			baseWhere += " AND SYS_COMPANY_ID = ?"
			whereParams = append(whereParams, user.SysCompanyID)
			fmt.Printf("[DEBUG] 添加 SYS_COMPANY_ID 过滤\n")
		}
	}

	fmt.Printf("[DEBUG] baseWhere=%s, whereParams=%v\n", baseWhere, whereParams)

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table.Name, baseWhere)
	countParams := make([]interface{}, len(whereParams))
	copy(countParams, whereParams)

	if search != "" {
		countSQL += fmt.Sprintf(" AND %s LIKE ?", displayColumn.DbName)
		countParams = append(countParams, "%"+search+"%")
	}

	if err := s.repo.CountBySql(countSQL, countParams...).Scan(&total).Error; err != nil {
		return nil, 0, "", errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
	}

	// 查询数据 - 返回所有字段
	dataSQL := fmt.Sprintf("SELECT * FROM %s WHERE %s", table.Name, baseWhere)
	dataParams := make([]interface{}, len(whereParams))
	copy(dataParams, whereParams)

	if search != "" {
		dataSQL += fmt.Sprintf(" AND %s LIKE ?", displayColumn.DbName)
		dataParams = append(dataParams, "%"+search+"%")
	}

	dataSQL += fmt.Sprintf(" ORDER BY %s LIMIT ? OFFSET ?", displayColumn.DbName)
	dataParams = append(dataParams, pageSize, offset)

	if err := s.repo.RawQuery(dataSQL, dataParams...).Scan(&options).Error; err != nil {
		return nil, 0, "", errors.Wrap(errors.ErrDatabase, "查询数据失败", err)
	}

	// 为每条记录添加 value 和 label 字段
	for i := range options {
		if id, ok := options[i]["ID"]; ok {
			options[i]["value"] = id
		}
		if labelVal, ok := options[i][displayColumn.DbName]; ok {
			options[i]["label"] = labelVal
		}
	}

	// 获取表的 IS_DROPDOWN 配置，默认为 "Y"
	isDropdown := "Y"
	if table.IsDropdown != "" {
		isDropdown = table.IsDropdown
	}

	return options, total, isDropdown, nil
}

// GetForeignKeyDisplayValue 获取外键显示值
func (s *service) GetForeignKeyDisplayValue(tableID uint, value string, columnID *uint, userID uint) (string, error) {
	// 1. 获取目标表的元数据
	table, err := s.GetTableByID(tableID)
	if err != nil {
		return "", errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 2. 确定显示字段（逻辑同上）
	var displayColumn *entity.SysColumn
	columns, err := s.GetColumns(tableID)
	if err != nil {
		return "", err
	}

	if columnID != nil {
		// 使用指定的字段
		for _, col := range columns {
			if col.ID == *columnID {
				displayColumn = col
				break
			}
		}
	}

	// 如果没有找到指定字段，使用 DK 字段
	if displayColumn == nil {
		for _, col := range columns {
			if col.IsDK == "Y" {
				displayColumn = col
				break
			}
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

	if displayColumn == nil {
		return "", errors.New(errors.ErrResourceNotFound, "未找到显示字段")
	}

	// 3. 获取用户信息以获取公司ID
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return "", errors.Wrap(errors.ErrInternal, "获取用户信息失败", err)
	}

	// 4. 构建查询条件
	baseWhere := "ID = ? AND IS_ACTIVE = 'Y'"
	queryParams := []interface{}{value}

	// 特殊处理：sys_company 表本身没有 SYS_COMPANY_ID 字段
	// 但用户只能访问自己所属的公司，检查 value 是否等于用户的公司ID
	if strings.EqualFold(table.Name, "sys_company") {
		if value != fmt.Sprintf("%d", user.SysCompanyID) {
			return "", errors.New(errors.ErrPermissionDenied, "无权访问其他公司信息")
		}
	} else {
		// 检查目标表是否有 SYS_COMPANY_ID 字段
		hasCompanyColumn := false
		for _, col := range columns {
			if col.DbName == "SYS_COMPANY_ID" {
				hasCompanyColumn = true
				break
			}
		}
		// 如果目标表有 SYS_COMPANY_ID 字段，添加公司过滤
		if hasCompanyColumn {
			baseWhere += " AND SYS_COMPANY_ID = ?"
			queryParams = append(queryParams, user.SysCompanyID)
		}
	}

	// 6. 查询显示值
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", displayColumn.DbName, table.Name, baseWhere)
	var displayValue string
	if err := s.repo.RawQuery(querySQL, queryParams...).Scan(&displayValue).Error; err != nil {
		return "", errors.Wrap(errors.ErrDatabase, "查询显示值失败", err)
	}

	return displayValue, nil
}
