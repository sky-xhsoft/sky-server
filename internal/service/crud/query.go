package crud

import (
	"context"
	"fmt"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
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
	fmt.Printf("[DEBUG] GetOne - 查询字段: %s\n", selectFields)

	// 获取数据过滤条件
	dataFilter, err := s.groupsService.GetUserDataFilter(ctx, userID, table.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}

	// 构建查询
	query := s.db.Table(table.Name).Select(selectFields)

	// 添加ID条件
	query = query.Where("ID = ?", id)
	fmt.Printf("[DEBUG] GetOne - 查询表: %s, ID: %d\n", table.Name, id)

	// 添加数据过滤条件
	if dataFilter != nil && len(dataFilter) > 0 {
		fmt.Printf("[DEBUG] GetOne - 原始数据过滤条件: %+v\n", dataFilter)

		// 从数据过滤条件中移除 IS_ACTIVE（GetOne 不应该过滤 IS_ACTIVE）
		filteredDataFilter := make(map[string]interface{})
		for k, v := range dataFilter {
			if k != "IS_ACTIVE" {
				filteredDataFilter[k] = v
			}
		}

		if len(filteredDataFilter) > 0 {
			fmt.Printf("[DEBUG] GetOne - 过滤后的数据过滤条件: %+v\n", filteredDataFilter)
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
	fmt.Printf("[DEBUG] GetOne - 执行 SQL: %s\n", sql)

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
		fmt.Printf("[WARN] 处理外键字段失败: %v\n", err)
	}

	return result, nil
}

// GetList 查询列表
func (s *service) GetList(ctx context.Context, req *QueryRequest, userID uint) (*QueryResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
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

	// 添加过滤条件
	if req.Filters != nil && len(req.Filters) > 0 {
		query = s.applyFilters(query, req.Filters, columns)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
	}

	// 添加排序（支持多字段，逗号分隔）
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

			direction := "ASC"
			if i < len(orderDirections) {
				dir := strings.TrimSpace(strings.ToUpper(orderDirections[i]))
				if dir == "DESC" {
					direction = "DESC"
				}
			}

			query = query.Order(fmt.Sprintf("%s %s", field, direction))
		}
	} else {
		query = query.Order("ID DESC")
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
		fmt.Printf("[WARN] 处理外键字段失败: %v\n", err)
	}

	return &QueryResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     results,
	}, nil
}
