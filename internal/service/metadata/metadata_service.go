package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/sqlutils"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"go.uber.org/zap"
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
	GetForeignKeyOptions(tableID uint, columnID *uint, search string, filters map[string]string, page, pageSize int, userID uint) ([]map[string]interface{}, int64, string, error)

	// 获取外键显示值
	GetForeignKeyDisplayValue(tableID uint, value string, columnID *uint, userID uint) (string, error)

	// ResolveDisplayColumn 解析外键显示字段（指定列→DK→ID）
	ResolveDisplayColumn(tableID uint, columnID *uint) (*entity.SysColumn, error)

	// ResolveInputColumn 解析外键输入字段（AK→DK→指定列→ID）
	ResolveInputColumn(tableID uint, columnID *uint) (*entity.SysColumn, error)
}

// service 元数据服务实现
type service struct {
	repo        repository.MetadataRepository
	userRepo    repository.UserRepository
	redisClient *redis.Client
	cacheTTL    time.Duration
	metaVersion string
}

// NewService 创建元数据服务
func NewService(repo repository.MetadataRepository, userRepo repository.UserRepository, redisClient *redis.Client, cacheTTL int) Service {
	return &service{
		repo:        repo,
		userRepo:    userRepo,
		redisClient: redisClient,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
		metaVersion: time.Now().Format("20060102150405"),
	}
}

// GetTable 获取表元数据（通过表名）
func (s *service) GetTable(tableName string) (*entity.SysTable, error) {
	cacheKey := fmt.Sprintf("metadata:table:name:%s", tableName)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
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
	s.redisClient.Set(context.Background(), cacheKey, data, s.cacheTTL)

	return table, nil
}

// GetTableByID 获取表元数据（通过ID）
func (s *service) GetTableByID(tableID uint) (*entity.SysTable, error) {
	cacheKey := fmt.Sprintf("metadata:table:id:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
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
	s.redisClient.Set(context.Background(), cacheKey, data, s.cacheTTL)

	return table, nil
}

// GetColumns 获取表的所有字段
func (s *service) GetColumns(tableID uint) ([]*entity.SysColumn, error) {
	cacheKey := fmt.Sprintf("metadata:columns:%d", tableID)

	// 尝试从缓存获取
	var columns []*entity.SysColumn
	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
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
	s.redisClient.Set(context.Background(), cacheKey, data, s.cacheTTL)

	return columns, nil
}

// GetTableRefs 获取表的关联关系
func (s *service) GetTableRefs(tableID uint) ([]*entity.SysTableRef, error) {
	cacheKey := fmt.Sprintf("metadata:refs:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
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
	s.redisClient.Set(context.Background(), cacheKey, data, s.cacheTTL)

	return refs, nil
}

// GetActions 获取表的所有动作
func (s *service) GetActions(tableID uint) ([]*entity.SysAction, error) {
	cacheKey := fmt.Sprintf("metadata:actions:%d", tableID)

	// 尝试从缓存获取
	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
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
	s.redisClient.Set(context.Background(), cacheKey, data, s.cacheTTL)

	return actions, nil
}

// RefreshCache 刷新缓存
func (s *service) RefreshCache() error {
	// 删除所有元数据缓存
	pattern := "metadata:*"
	iter := s.redisClient.Scan(context.Background(), 0, pattern, 0).Iterator()

	keys := []string{}
	for iter.Next(context.Background()) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return errors.Wrap(errors.ErrCache, "扫描缓存失败", err)
	}

	// 批量删除
	if len(keys) > 0 {
		if err := s.redisClient.Del(context.Background(), keys...).Err(); err != nil {
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

// syncTableDkAk 懒回填系统表单的 DK_COLUMN_ID / AK_COLUMN_ID
// dkID>0 时更新显示键；akID>0 时更新输入键。
func (s *service) syncTableDkAk(tableID uint, dkID, akID uint) {
	var sets []string
	var args []interface{}
	if dkID > 0 {
		sets = append(sets, "DK_COLUMN_ID = ?")
		args = append(args, dkID)
	}
	if akID > 0 {
		sets = append(sets, "AK_COLUMN_ID = ?")
		args = append(args, akID)
	}
	if len(sets) == 0 {
		return
	}
	args = append(args, tableID)
	sql := "UPDATE sys_table SET " + strings.Join(sets, ", ") + " WHERE ID = ?"
	if err := s.repo.Exec(sql, args...); err != nil {
		// 静默失败：懒回填是优化，失败不影响主流程
		_ = err
	}
	// 使缓存失效，避免读到旧表配置
	s.redisClient.Del(context.Background(), fmt.Sprintf("metadata:table:id:%d", tableID))
}

// ResolveDisplayColumn 解析外键显示字段（判定：DK列ID→指定列→DK→ID）
// 优先读取系统表单缓存好的 DK_COLUMN_ID（O(1)），未设置再遍历回退。
func (s *service) ResolveDisplayColumn(tableID uint, columnID *uint) (*entity.SysColumn, error) {
	columns, err := s.GetColumns(tableID)
	if err != nil {
		return nil, err
	}

	isPkLike := func(col *entity.SysColumn) bool {
		return col.SetValueType == "pk" || strings.EqualFold(col.DbName, "ID") || strings.EqualFold(col.DbName, "PK")
	}

	lookup := func(id uint) (*entity.SysColumn, bool) {
		for _, col := range columns {
			if col.ID == id {
				return col, true
			}
		}
		return nil, false
	}

	// 0. 优先使用系统表单缓存的 DK_COLUMN_ID（查询更快，不必逐列判断 IS_DK）
	if t, err := s.GetTableByID(tableID); err == nil && t != nil && t.DkColumnID != nil && uint64(*t.DkColumnID) > 0 {
		if col, ok := lookup(uint(*t.DkColumnID)); ok && !isPkLike(col) {
			return col, nil
		}
	}

	// 1. 其次优先使用指定的显示列（排除主键/ID 列，避免 FK 指向 ID 时把数字当名称显示）
	if columnID != nil {
		for _, col := range columns {
			if col.ID == *columnID {
				if !isPkLike(col) {
					return col, nil
				}
				break
			}
		}
	}

	// 2. 再次使用 DK 字段（显示键，排除主键/ID 列——DK 常常标在 ID 列上）
	for _, col := range columns {
		if col.IsDK == "Y" && !isPkLike(col) {
			// 懒回填：系统表单 DK_COLUMN_ID 为空时，自动写入（字段页设过 IS_DK 无需再手改系统表单）
			if t, err := s.GetTableByID(tableID); err == nil && t != nil && t.DkColumnID == nil {
				s.syncTableDkAk(tableID, col.ID, 0)
			}
			return col, nil
		}
	}

	// 3. 最后默认显示 ID 字段
	for _, col := range columns {
		if strings.EqualFold(col.DbName, "ID") {
			return col, nil
		}
	}

	return nil, errors.New(errors.ErrResourceNotFound, "未找到显示字段")
}

// ResolveInputColumn 解析外键输入字段（判定：AK列ID→AK→DK→指定列→ID）
// 优先读取系统表单缓存的 AK_COLUMN_ID（查询更快），未设置再遍历回退。
func (s *service) ResolveInputColumn(tableID uint, columnID *uint) (*entity.SysColumn, error) {
	columns, err := s.GetColumns(tableID)
	if err != nil {
		return nil, err
	}

	isPkLike := func(col *entity.SysColumn) bool {
		return col.SetValueType == "pk" || strings.EqualFold(col.DbName, "ID") || strings.EqualFold(col.DbName, "PK")
	}

	lookup := func(id uint) (*entity.SysColumn, bool) {
		for _, col := range columns {
			if col.ID == id {
				return col, true
			}
		}
		return nil, false
	}

	// 0. 优先使用系统表单缓存的 AK_COLUMN_ID（查询更快）
	if t, err := s.GetTableByID(tableID); err == nil && t != nil && t.AkColumnID != nil && *t.AkColumnID > 0 {
		if col, ok := lookup(uint(*t.AkColumnID)); ok && !isPkLike(col) {
			return col, nil
		}
	}

	// 1. 优先使用 AK 字段（输入键）
	for _, col := range columns {
		if col.IsAK == "Y" && !isPkLike(col) {
			// 懒回填：系统表单 AK_COLUMN_ID 为空时自动写入
			if t, err := s.GetTableByID(tableID); err == nil && t != nil && t.AkColumnID == nil {
				s.syncTableDkAk(tableID, 0, col.ID)
			}
			return col, nil
		}
	}

	// 2. 其次 DK 字段
	for _, col := range columns {
		if col.IsDK == "Y" && !isPkLike(col) {
			return col, nil
		}
	}

	// 3. 指定列（非主键/ID）
	if columnID != nil {
		for _, col := range columns {
			if col.ID == *columnID && !isPkLike(col) {
				return col, nil
			}
		}
	}

	// 4. 默认 ID 字段
	for _, col := range columns {
		if strings.EqualFold(col.DbName, "ID") {
			return col, nil
		}
	}

	return nil, errors.New(errors.ErrResourceNotFound, "未找到输入字段")
}

// GetForeignKeyOptions 获取外键选项列表
func (s *service) GetForeignKeyOptions(tableID uint, columnID *uint, search string, filters map[string]string, page, pageSize int, userID uint) ([]map[string]interface{}, int64, string, error) {
	// 1. 获取目标表的元数据
	table, err := s.GetTableByID(tableID)
	if err != nil {
		return nil, 0, "", errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 2. 确定输入字段（统一判定：AK→DK→指定列→ID）——下拉选项用目标表单 AK 字段值，交由用户输入校验
	columns, err := s.GetColumns(tableID)
	if err != nil {
		return nil, 0, "", err
	}
	displayColumn, err := s.ResolveInputColumn(tableID, columnID)
	if err != nil {
		return nil, 0, "", err
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

	logger.Debug("GetForeignKeyOptions", zap.String("tableName", table.Name), zap.Uint("userID", userID), zap.Uint("companyID", user.SysCompanyID))

	// 特殊处理：sys_company 表本身没有 SYS_COMPANY_ID 字段
	// 用户只能看到自己所属的公司
	if strings.EqualFold(table.Name, "sys_company") {
		baseWhere += " AND ID = ?"
		whereParams = append(whereParams, user.SysCompanyID)
		logger.Debug("sys_company special handling: added ID filter")
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
			baseWhere += " AND (SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1))"
			whereParams = append(whereParams, user.SysCompanyID)
			logger.Debug("Added SYS_COMPANY_ID filter with global fallback")
		}
	}

	logger.Debug("GetForeignKeyOptions SQL", zap.String("baseWhere", baseWhere), zap.Any("whereParams", whereParams))

	// 多条件过滤：查询弹窗的用户输入，按目标表真实列名等值匹配（白名单防注入）
	columnNames := make(map[string]bool)
	for _, col := range columns {
		columnNames[strings.ToUpper(col.DbName)] = true
	}
	for key, val := range filters {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		upperKey := strings.ToUpper(strings.TrimSpace(key))
		if !columnNames[upperKey] {
			continue // 非真实列名，忽略
		}
		filterCol, _ := sqlutils.SanitizeIdentifier(upperKey)
		if filterCol == "" {
			filterCol = upperKey
		}
		// 查询条件用模糊匹配（输入部分关键字可命中，如 姓"张"匹配"张伟"）
		baseWhere += " AND " + filterCol + " LIKE ?"
		whereParams = append(whereParams, "%"+val+"%")
	}

	safeCol, _ := sqlutils.SanitizeIdentifier(displayColumn.DbName)
	if safeCol == "" {
		safeCol = displayColumn.DbName
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table.Name, baseWhere)
	countParams := make([]interface{}, len(whereParams))
	copy(countParams, whereParams)

	if search != "" {
		countSQL += fmt.Sprintf(" AND %s LIKE ?", safeCol)
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
		dataSQL += fmt.Sprintf(" AND %s LIKE ?", safeCol)
		dataParams = append(dataParams, "%"+search+"%")
	}

	dataSQL += fmt.Sprintf(" ORDER BY %s LIMIT ? OFFSET ?", safeCol)
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

	// 2. 确定显示字段（统一判定）
	columns, err := s.GetColumns(tableID)
	if err != nil {
		return "", err
	}
	displayColumn, err := s.ResolveDisplayColumn(tableID, columnID)
	if err != nil {
		return "", err
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
		// 如果目标表有 SYS_COMPANY_ID 字段，添加公司过滤(全局记录 company_id=0 或 -1，或当前用户公司)
		if hasCompanyColumn {
			baseWhere += " AND (SYS_COMPANY_ID = 0 OR SYS_COMPANY_ID = -1 OR SYS_COMPANY_ID = ?)"
			queryParams = append(queryParams, user.SysCompanyID)
		}
	}

	// 6. 查询显示值
	safeCol, _ := sqlutils.SanitizeIdentifier(displayColumn.DbName)
	if safeCol == "" {
		return "", errors.New(errors.ErrInvalidParam, "无效的列名")
	}
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", safeCol, table.Name, baseWhere)
	var displayValue string
	if err := s.repo.RawQuery(querySQL, queryParams...).Scan(&displayValue).Error; err != nil {
		return "", errors.Wrap(errors.ErrDatabase, "查询显示值失败", err)
	}

	return displayValue, nil
}
