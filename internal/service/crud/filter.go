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
// 使用智能转换:字符串字段默认 LIKE、支持 =精确/*通配;NUMBER 支持 =/>/</→/区间;DATE 支持日期范围(仿 Bos toSQLClause)
func (s *service) applyFilters(query *gorm.DB, filters map[string]interface{}, columns []*entity.SysColumn) *gorm.DB {
	// 构建字段→(显示类型,DB列类型) 映射
	colInfo := make(map[string]*entity.SysColumn)
	if columns != nil {
		for _, col := range columns {
			colInfo[col.DbName] = col
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

		col := colInfo[safeField]
		// 非字符串类型(如 FK 下拉选值、bool)仍走精确匹配
		if _, isStr := value.(string); !isStr || col == nil {
			query = query.Where(fmt.Sprintf("`%s` = ?", safeField), value)
			continue
		}

		// 字符串 → 推断 db 类型并智能转换
		dbType := inferDBType(col)
		sqlClause, args := toSQLClause(safeField, value.(string), dbType)
		query = query.Where(sqlClause, args...)
	}

	return query
}

// inferDBType 由列定义推断 MySQL 语义类型(string/number/date/text)
func inferDBType(col *entity.SysColumn) string {
	if col == nil {
		return "string"
	}
	ct := strings.ToUpper(col.ColType)
	ct = strings.TrimSpace(strings.SplitN(ct, "(", 2)[0])
	switch {
	case strings.Contains(ct, "INT"), strings.Contains(ct, "DECIMAL"), strings.Contains(ct, "NUMERIC"),
		strings.Contains(ct, "FLOAT"), strings.Contains(ct, "DOUBLE"), strings.Contains(ct, "BIGINT"):
		return "number"
	case strings.Contains(ct, "DATE"), strings.Contains(ct, "TIME"), strings.Contains(ct, "YEAR"):
		return "date"
	case strings.Contains(ct, "TEXT"), strings.Contains(ct, "BLOB"), strings.Contains(ct, "CLOB"):
		return "text"
	default:
		return "string"
	}
}
