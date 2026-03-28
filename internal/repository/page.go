package repository

import (
	"context"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// PageRepository 分页Repository接口
type PageRepository interface {
	GetPage(ctx context.Context, query *PageQuery, dest interface{}) (*PageResult, error)
	GetPageWithQuery(ctx context.Context, query *gorm.DB, page, pageSize int, dest interface{}) (*PageResult, error)
}

// PageRepositoryImpl 分页Repository实现
type PageRepositoryImpl struct {
	db *gorm.DB
}

// NewPageRepository 创建分页Repository
func NewPageRepository(db *gorm.DB) *PageRepositoryImpl {
	return &PageRepositoryImpl{
		db: db,
	}
}

// GetPage 获取分页列表
func (r *PageRepositoryImpl) GetPage(ctx context.Context, query *PageQuery, dest interface{}) (*PageResult, error) {
	if query == nil {
		query = &PageQuery{
			Page:     1,
			PageSize: 10,
		}
	}

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	if query.PageSize > 100 {
		query.PageSize = 100
	}

	db := r.db.WithContext(ctx).Model(dest)

	// 处理排序
	if query.OrderBy != "" {
		order := "asc"
		if query.Order != "" {
			order = strings.ToLower(query.Order)
		}
		db = db.Order(query.OrderBy + " " + order)
	} else {
		db = db.Order("CREATE_TIME DESC")
	}

	// 查询总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(dest).Error; err != nil {
		return nil, err
	}

	return &PageResult{
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
		Data:     dest,
	}, nil
}

// GetPageWithQuery 使用自定义查询获取分页列表
func (r *PageRepositoryImpl) GetPageWithQuery(ctx context.Context, query *gorm.DB, page, pageSize int, dest interface{}) (*PageResult, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	if pageSize > 100 {
		pageSize = 100
	}

	// 复制查询对象用于计数
	countQuery := query.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(dest).Error; err != nil {
		return nil, err
	}

	return &PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     dest,
	}, nil
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return totalPages
}

// CreatePageResult 创建分页结果
func CreatePageResult(data interface{}, page, pageSize int, total int64) *PageResult {
	return &PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}
}

// BuildOrderBy 构建排序条件
func BuildOrderBy(defaultOrderBy, orderBy, order string) string {
	if orderBy == "" {
		return defaultOrderBy
	}

	if order == "" {
		order = "asc"
	}

	order = strings.ToLower(order)
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	return orderBy + " " + order
}

// BuildOffset 计算偏移量
func BuildOffset(page, pageSize int) int {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return (page - 1) * pageSize
}

// IsValidPage 检查分页参数是否有效
func IsValidPage(page, pageSize int) bool {
	if page < 1 {
		return false
	}
	if pageSize < 1 || pageSize > 1000 {
		return false
	}
	return true
}

// ValidateAndFixPage 验证并修正分页参数
func ValidateAndFixPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	return page, pageSize
}

// GetSliceType 获取切片元素类型
func GetSliceType(slice interface{}) reflect.Type {
	sliceType := reflect.TypeOf(slice)
	if sliceType.Kind() == reflect.Ptr {
		sliceType = sliceType.Elem()
	}
	if sliceType.Kind() == reflect.Slice {
		return sliceType.Elem()
	}
	return sliceType
}
