package crud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/sqlutils"
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
		// 跳过 nil 值和空字符串
		if value == nil {
			continue
		}
		if strValue, ok := value.(string); ok && strValue == "" {
			continue
		}

		// 校验字段名为合法 SQL 标识符
		safeField, err := sqlutils.SanitizeIdentifier(field)
		if err != nil {
			continue
		}

		// 获取字段的显示类型
		displayType := columnTypeMap[safeField]

		// 对于 text 类型字段（text, textarea），使用 LIKE 模糊查询
		if displayType == "text" || displayType == "textarea" || displayType == "clob" {
			// 转换为字符串并添加通配符
			if strValue, ok := value.(string); ok && strValue != "" {
				query = query.Where(fmt.Sprintf("`%s` LIKE ?", safeField), "%"+strValue+"%")
			}
		} else {
			// 其他类型使用精确匹配
			// 如果值是逗号分隔的字符串，使用 IN 查询
			if strValue, ok := value.(string); ok && strings.Contains(strValue, ",") {
				parts := strings.Split(strValue, ",")
				var values []interface{}
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						values = append(values, p)
					}
				}
				if len(values) > 0 {
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1]
					query = query.Where(fmt.Sprintf("`%s` IN (%s)", safeField, placeholders), values...)
				}
			} else {
				query = query.Where(fmt.Sprintf("`%s` = ?", safeField), value)
			}
		}
	}

	return query
}
