package crud

import (
	"context"
	"fmt"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/domain/search"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetOne 查询单条记录
func (s *service) GetOne(ctx context.Context, tableName string, id uint, userID uint) (map[string]interface{}, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查读权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermRead)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无查询权限")
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 构建查询字段（根据MASK控制）
	selectFields, err := s.buildSelectFields(columns, userID, "edit")
	if err != nil {
		return nil, err
	}
	logger.Debug("GetOne query fields", zap.String("fields", selectFields))

	// 获取数据过滤条件
	dataFilter, err := s.groupsService.GetUserDataFilter(ctx, userID, table.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}

	// 构建查询
	query := s.db.Table(table.Name).Select(selectFields)

	// 应用表级数据过滤条件
	if table.Filter != "" {
		query = s.applyDataFilter(query, table.Filter, columns)
	}

	// 添加ID条件
	query = query.Where("ID = ?", id)
	// 公司隔离：与列表一致，只能读取本公司或公共数据
	query = s.applyCompanyScope(query, table.Name, columns, userID)
	logger.Debug("GetOne table and ID", zap.String("table", table.Name), zap.Uint("id", id))

	// 添加数据过滤条件
	if dataFilter != nil && len(dataFilter) > 0 {
		logger.Debug("GetOne dataFilter", zap.Any("filter", dataFilter))

		// 从数据过滤条件中移除 IS_ACTIVE（GetOne 不应该过滤 IS_ACTIVE）
		filteredDataFilter := make(map[string]interface{})
		for k, v := range dataFilter {
			if k != "IS_ACTIVE" {
				filteredDataFilter[k] = v
			}
		}

		if len(filteredDataFilter) > 0 {
			logger.Debug("GetOne filtered dataFilter", zap.Any("filter", filteredDataFilter))
			query = s.applyFilters(query, filteredDataFilter, columns)
		}
	}

	// 注意：GetOne 通过 ID 精确查询，不过滤 IS_ACTIVE
	// 用户可能需要查看或编辑已停用的记录
	// IS_ACTIVE 过滤主要用于列表查询（GetList）

	// 打印最终的 SQL（用于调试）
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Take(&map[string]interface{}{})
	})
	logger.Debug("GetOne SQL", zap.String("sql", sql))

	// 执行查询
	var result map[string]interface{}
	if err := query.Take(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "记录不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询失败", err)
	}

	// 处理外键字段，将ID转换为显示名称
	results := []map[string]interface{}{result}
	if err := s.processForeignKeys(ctx, columns, results, userID); err != nil {
		// FK 转换失败不影响主查询，只记录错误
		logger.Warn("processFK failed", zap.Error(err))
	}

	return result, nil
}

// applyCompanyScope 公司隔离：表有 SYS_COMPANY_ID 字段时只返回本公司与公共（0/-1）数据。
// 豁免：附属配置/子表（如 SYS_COMPANY_CONF 公司配置）不按当前用户公司隔离，
// 其可见性由所在公司的详情页上下文决定（外部已按外键过滤）
func (s *service) applyCompanyScope(query *gorm.DB, tableName string, columns []*entity.SysColumn, userID uint) *gorm.DB {
	for _, rt := range []string{"SYS_COMPANY_CONF"} {
		if strings.EqualFold(tableName, rt) {
			return query
		}
	}
	for _, col := range columns {
		if col.DbName == "SYS_COMPANY_ID" {
			user, userErr := s.userRepo.GetUserByID(userID)
			if userErr == nil && user != nil {
				query = query.Where("SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)", user.SysCompanyID)
			}
			break
		}
	}
	return query
}

// ensureInCompanyScope 写操作前确认目标记录都在当前用户的公司范围内（与查询一致），否则视为记录不存在。
// 写操作只按 ID 定位记录，缺少这一步会让有修改/删除权限的用户改动其他公司的数据。
func (s *service) ensureInCompanyScope(tableName string, tableID uint, userID uint, ids ...uint) error {
	uniq := make(map[uint]bool, len(ids))
	for _, id := range ids {
		uniq[id] = true
	}
	if len(uniq) == 0 {
		return nil
	}
	columns, err := s.metadataService.GetColumns(tableID)
	if err != nil {
		return err
	}
	list := make([]uint, 0, len(uniq))
	for id := range uniq {
		list = append(list, id)
	}
	var n int64
	q := s.applyCompanyScope(s.db.Table(tableName).Where("ID IN ?", list), tableName, columns, userID)
	if err := q.Count(&n).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
	}
	if n < int64(len(list)) {
		return errors.New(errors.ErrResourceNotFound, "记录不存在")
	}
	return nil
}

// GetList 查询列表
func (s *service) GetList(ctx context.Context, req *QueryRequest, userID uint) (*QueryResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 15
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取表元数据
	table, err := s.metadataService.GetTable(req.TableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查读权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermRead)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无查询权限")
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 构建允许的字段名集合
	allowedFields := make(map[string]bool)
	for _, col := range columns {
		allowedFields[strings.ToUpper(col.DbName)] = true
	}
	// 也允许 ID 字段（主键）
	allowedFields["ID"] = true

	// 构建查询字段（根据MASK控制）
	selectFields, err := s.buildSelectFields(columns, userID, "list")
	if err != nil {
		return nil, err
	}

	// 获取数据过滤条件
	dataFilter, err := s.groupsService.GetUserDataFilter(ctx, userID, table.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}

	// 构建查询
	query := s.db.Table(table.Name).Select(selectFields)

	// 应用表级数据过滤条件
	if table.Filter != "" {
		query = s.applyDataFilter(query, table.Filter, columns)
	}

	// 添加数据过滤条件
	if dataFilter != nil && len(dataFilter) > 0 {
		// 从数据过滤条件中移除 IS_ACTIVE（不应该默认过滤 IS_ACTIVE）
		filteredDataFilter := make(map[string]interface{})
		for k, v := range dataFilter {
			if k != "IS_ACTIVE" {
				filteredDataFilter[k] = v
			}
		}

		if len(filteredDataFilter) > 0 {
			query = s.applyFilters(query, filteredDataFilter, columns)
		}
	}

	// 注意：不再默认过滤 IS_ACTIVE = 'Y'
	// 用户可以通过查询条件主动过滤 IS_ACTIVE
	// 例如：filters: { IS_ACTIVE: 'Y' }

	// 添加公司过滤（如果表有 SYS_COMPANY_ID 字段）
	query = s.applyCompanyScope(query, table.Name, columns, userID)

	// 添加过滤条件
	if req.Filters != nil && len(req.Filters) > 0 {
		query = s.applyFilters(query, req.Filters, columns)
	}

	// 计算总数
	var total int64
	if table.IsBig == "Y" {
		total = -1
	} else {
		if err := query.Count(&total).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
		}
	}

	// 添加排序（支持多字段，逗号分隔）- 安全版本
	if req.OrderBy != "" {
		// 分割排序字段和方向
		orderByFields := strings.Split(req.OrderBy, ",")
		orderDirections := strings.Split(req.Order, ",")

		// 确保每个字段都有对应的排序方向，默认 ASC
		for i, field := range orderByFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			// 验证字段名是否在允许的列表中
			fieldUpper := strings.ToUpper(field)
			if !allowedFields[fieldUpper] {
				logger.Warn("Invalid order by field skipped", zap.String("field", field))
				continue
			}

			// 验证排序方向
			direction := "ASC"
			if i < len(orderDirections) {
				dir := strings.TrimSpace(strings.ToUpper(orderDirections[i]))
				if dir == "DESC" {
					direction = "DESC"
				}
			}

			// 字段已在白名单中校验；原先嵌套 gorm.Expr 生成的 SQL 不含 ORDER BY，排序不生效
			query = query.Order(fmt.Sprintf("`%s` %s", fieldUpper, direction))
		}
	} else {
		// 检查是否有ORDERNO字段，如果有则默认按ORDERNO升序排序
		hasOrderNo := false
		for _, col := range columns {
			if strings.EqualFold(col.DbName, "ORDERNO") {
				hasOrderNo = true
				break
			}
		}
		if hasOrderNo {
			query = query.Order("ORDERNO ASC")
		} else {
			query = query.Order("ID DESC")
		}
	}

	// 添加分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Limit(req.PageSize).Offset(offset)

	// 执行查询
	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询失败", err)
	}

	// 处理外键字段，将ID转换为显示名称
	if err := s.processForeignKeys(ctx, columns, results, userID); err != nil {
		// FK 转换失败不影响主查询，只记录错误
		logger.Warn("processFK failed", zap.Error(err))
	}

	return &QueryResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     results,
	}, nil
}

// Search 高级多条件搜索
func (s *service) Search(ctx context.Context, req *search.SearchRequest, tableName string, userID uint) (*QueryResponse, error) {
	req.Validate()

	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermRead)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无查询权限")
	}

	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 构建可查询字段白名单（只有 IS_QUERY='Y' 的字段才能被搜索）
	queryableFields := make(map[string]bool)
	for _, col := range columns {
		queryableFields[strings.ToUpper(col.DbName)] = true
	}
	queryableFields["ID"] = true

	// 校验搜索字段必须在白名单内
	if req.HasFilters() {
		for _, filter := range req.Filters {
			if !queryableFields[strings.ToUpper(filter.Field)] {
				return nil, errors.New(errors.ErrInvalidParam, "字段不允许搜索: "+filter.Field)
			}
		}
	}

	selectFields, err := s.buildSelectFields(columns, userID, "list")
	if err != nil {
		return nil, err
	}

	dataFilter, err := s.groupsService.GetUserDataFilter(ctx, userID, table.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}

	query := s.db.Table(table.Name).Select(selectFields)

	if table.Filter != "" {
		query = s.applyDataFilter(query, table.Filter, columns)
	}

	if dataFilter != nil && len(dataFilter) > 0 {
		filteredDataFilter := make(map[string]interface{})
		for k, v := range dataFilter {
			if k != "IS_ACTIVE" {
				filteredDataFilter[k] = v
			}
		}
		if len(filteredDataFilter) > 0 {
			query = s.applyFilters(query, filteredDataFilter, columns)
		}
	}

	query = s.applyCompanyScope(query, table.Name, columns, userID)

	// 应用高级搜索条件
	if req.HasFilters() {
		query, err = s.buildAdvancedSearch(query, req)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInvalidParam, "搜索条件构建失败", err)
		}
	}

	var total int64
	if table.IsBig == "Y" {
		total = -1
	} else {
		if err := query.Count(&total).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
		}
	}

	allowedFields := make(map[string]bool)
	for _, col := range columns {
		allowedFields[strings.ToUpper(col.DbName)] = true
	}
	allowedFields["ID"] = true

	if req.OrderBy != "" {
		orderByFields := strings.Split(req.OrderBy, ",")
		orderDirections := strings.Split(req.Order, ",")

		for i, field := range orderByFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			fieldUpper := strings.ToUpper(field)
			if !allowedFields[fieldUpper] {
				logger.Warn("Invalid order by field skipped", zap.String("field", field))
				continue
			}

			direction := "ASC"
			if i < len(orderDirections) {
				dir := strings.TrimSpace(strings.ToUpper(orderDirections[i]))
				if dir == "DESC" {
					direction = "DESC"
				}
			}

			query = query.Order(fmt.Sprintf("`%s` %s", fieldUpper, direction))
		}
	} else {
		hasOrderNo := false
		for _, col := range columns {
			if strings.EqualFold(col.DbName, "ORDERNO") {
				hasOrderNo = true
				break
			}
		}
		if hasOrderNo {
			query = query.Order("ORDERNO ASC")
		} else {
			query = query.Order("ID DESC")
		}
	}

	offset := (req.Page - 1) * req.PageSize
	query = query.Limit(req.PageSize).Offset(offset)

	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询失败", err)
	}

	if err := s.processForeignKeys(ctx, columns, results, userID); err != nil {
		logger.Warn("processFK failed", zap.Error(err))
	}

	return &QueryResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     results,
	}, nil
}
