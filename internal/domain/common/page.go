package common

// PageRequest 通用分页请求
type PageRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// PageResponse 通用分页响应
type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// Validate 校验并修正分页参数
func (p *PageRequest) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算偏移量
func (p *PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// NewPageResponse 创建分页响应
func NewPageResponse[T any](list []T, total int64, page, pageSize int) *PageResponse[T] {
	return &PageResponse[T]{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
