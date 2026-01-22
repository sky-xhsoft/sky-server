package crud

import (
	"encoding/json"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// applyDataFilter 应用数据过滤条件
func (s *service) applyDataFilter(query *gorm.DB, filterJSON string, columns []*entity.SysColumn) *gorm.DB {
	if filterJSON == "" {
		return query
	}

	var filter map[string]interface{}
	if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
		return query
	}

	return s.applyFilters(query, filter, columns)
}

// applyFilters 应用过滤条件
// 对于 text 类型字段使用 LIKE 模糊匹配，其他类型使用精确匹配
func (s *service) applyFilters(query *gorm.DB, filters map[string]interface{}, columns []*entity.SysColumn) *gorm.DB {
	// 构建字段类型映射，方便快速查找
	columnTypeMap := make(map[string]string)
	if columns != nil {
		for _, col := range columns {
			columnTypeMap[col.DbName] = col.DisplayType
		}
	}

	for field, value := range filters {
		// 跳过 nil 值
		if value == nil {
			continue
		}

		// 获取字段的显示类型
		displayType := columnTypeMap[field]

		// 对于 text 类型字段（text, textarea），使用 LIKE 模糊查询
		// 根据 DisplayType 字段定义：blank,button,hr,check,file,image,select,text,textarea,date,datetime,clob,xml,json
		if displayType == "text" || displayType == "textarea" || displayType == "clob" {
			// 转换为字符串并添加通配符
			if strValue, ok := value.(string); ok && strValue != "" {
				query = query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+strValue+"%")
			}
		} else {
			// 其他类型使用精确匹配（select, date, datetime, check 等）
			query = query.Where(fmt.Sprintf("%s = ?", field), value)
		}
	}

	return query
}
