package search

// Operator 搜索运算符
type Operator string

const (
	OpEq         Operator = "eq"
	OpNe         Operator = "ne"
	OpGt         Operator = "gt"
	OpGte        Operator = "gte"
	OpLt         Operator = "lt"
	OpLte        Operator = "lte"
	OpContains   Operator = "contains"
	OpStartsWith Operator = "startsWith"
	OpEndsWith   Operator = "endsWith"
	OpIn         Operator = "in"
	OpBetween    Operator = "between"
	OpIsNull     Operator = "isNull"
	OpIsNotNull  Operator = "isNotNull"
)

// SearchFilter 单个搜索条件
type SearchFilter struct {
	Field     string        `json:"field"`
	Operator  Operator      `json:"operator"`
	Value     interface{}   `json:"value,omitempty"`
	ValueFrom interface{}   `json:"valueFrom,omitempty"`
	ValueTo   interface{}   `json:"valueTo,omitempty"`
	Values    []interface{} `json:"values,omitempty"`
}

// SearchRequest 多条件搜索请求
type SearchRequest struct {
	Filters  []*SearchFilter `json:"filters"`
	OrderBy  string          `json:"orderBy"`
	Order    string          `json:"order"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// Validate 校验搜索请求
func (r *SearchRequest) Validate() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 10
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	if r.Order == "" {
		r.Order = "desc"
	}
}

// HasFilters 是否有搜索条件
func (r *SearchRequest) HasFilters() bool {
	return len(r.Filters) > 0
}
