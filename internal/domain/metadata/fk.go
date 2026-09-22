package metadata

// ForeignKeyOption 外键选项
type ForeignKeyOption struct {
	ID    uint        `json:"id"`
	Label string      `json:"label"`
	Data  interface{} `json:"data,omitempty"`
}

// ForeignKeyQuery 外键查询参数
type ForeignKeyQuery struct {
	TableID  uint                   `json:"tableId"`
	ColumnID *uint                  `json:"columnId,omitempty"`
	Search   string                 `json:"search,omitempty"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

// ForeignKeyResult 外键查询结果
type ForeignKeyResult struct {
	Options []*ForeignKeyOption `json:"options"`
	Total   int64               `json:"total"`
}

// ForeignKeyDisplayQuery 外键显示值查询参数
type ForeignKeyDisplayQuery struct {
	TableID  uint   `json:"tableId"`
	Value    string `json:"value"`
	ColumnID *uint  `json:"columnId,omitempty"`
}
