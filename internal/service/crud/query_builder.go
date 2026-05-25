package crud

import (
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/domain/search"
	"github.com/sky-xhsoft/sky-server/internal/pkg/sqlutils"
	"gorm.io/gorm"
)

// buildAdvancedSearch 根据 SearchRequest 构建高级搜索条件
func (s *service) buildAdvancedSearch(query *gorm.DB, sr *search.SearchRequest) (*gorm.DB, error) {
	if sr == nil || !sr.HasFilters() {
		return query, nil
	}

	for _, filter := range sr.Filters {
		safeField, err := sqlutils.SanitizeIdentifier(filter.Field)
		if err != nil {
			return query, fmt.Errorf("invalid field name: %s", filter.Field)
		}

		var condition string
		var args []interface{}

		switch filter.Operator {
		case search.OpEq:
			condition = fmt.Sprintf("`%s` = ?", safeField)
			args = []interface{}{filter.Value}
		case search.OpNe:
			condition = fmt.Sprintf("`%s` != ?", safeField)
			args = []interface{}{filter.Value}
		case search.OpGt:
			condition = fmt.Sprintf("`%s` > ?", safeField)
			args = []interface{}{filter.Value}
		case search.OpGte:
			condition = fmt.Sprintf("`%s` >= ?", safeField)
			args = []interface{}{filter.Value}
		case search.OpLt:
			condition = fmt.Sprintf("`%s` < ?", safeField)
			args = []interface{}{filter.Value}
		case search.OpLte:
			condition = fmt.Sprintf("`%s` <= ?", safeField)
			args = []interface{}{filter.Value}
		case search.OpContains:
			condition = fmt.Sprintf("`%s` LIKE ?", safeField)
			args = []interface{}{"%" + toString(filter.Value) + "%"}
		case search.OpStartsWith:
			condition = fmt.Sprintf("`%s` LIKE ?", safeField)
			args = []interface{}{toString(filter.Value) + "%"}
		case search.OpEndsWith:
			condition = fmt.Sprintf("`%s` LIKE ?", safeField)
			args = []interface{}{"%" + toString(filter.Value)}
		case search.OpIn:
			if len(filter.Values) == 0 {
				continue
			}
			placeholders := ""
			for i := range filter.Values {
				if i > 0 {
					placeholders += ","
				}
				placeholders += "?"
			}
			condition = fmt.Sprintf("`%s` IN (%s)", safeField, placeholders)
			args = filter.Values
		case search.OpBetween:
			condition = fmt.Sprintf("`%s` BETWEEN ? AND ?", safeField)
			args = []interface{}{filter.ValueFrom, filter.ValueTo}
		case search.OpIsNull:
			condition = fmt.Sprintf("`%s` IS NULL", safeField)
		case search.OpIsNotNull:
			condition = fmt.Sprintf("`%s` IS NOT NULL", safeField)
		default:
			return query, fmt.Errorf("unsupported operator: %s", filter.Operator)
		}

		if condition != "" {
			query = query.Where(condition, args...)
		}
	}

	return query, nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
