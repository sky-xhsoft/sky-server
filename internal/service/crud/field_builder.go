package crud

import (
	"fmt"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/mask"
)

// buildSelectFields 构建查询字段列表（根据MASK控制）
func (s *service) buildSelectFields(columns []*entity.SysColumn, userID uint, operation string) (string, error) {
	var fields []string

	// 定义标准的系统审计字段（无论MASK如何配置都应包含）
	standardFields := map[string]bool{
		"ID":             true,
		"SYS_COMPANY_ID": true,
		"CREATE_BY":      true,
		"CREATE_TIME":    true,
		"UPDATE_BY":      true,
		"UPDATE_TIME":    true,
		"IS_ACTIVE":      true,
	}

	// 跟踪哪些系统字段已经被添加
	addedStandardFields := make(map[string]bool)

	for _, col := range columns {
		// TODO: 检查字段权限（基于SGRADE）- 需要集成groups权限服务
		// 暂时允许所有字段访问

		// 主键字段和系统审计字段始终包含，不受MASK限制
		isPrimaryKey := col.IsAK == "Y" || col.SetValueType == "pk"
		isStandardField := standardFields[col.DbName]

		if isStandardField {
			addedStandardFields[col.DbName] = true
		}

		if !isPrimaryKey && !isStandardField {
			// 对于业务字段，检查MASK可见性
			if col.Mask != "" {
				fieldMask := mask.ParseMask(col.Mask)
				if !fieldMask.IsVisible(operation) {
					continue
				}
			}
		}

		fields = append(fields, col.DbName)
	}

	// 确保所有系统审计字段都被包含（即使它们不在 sys_column 中定义）
	for fieldName := range standardFields {
		if !addedStandardFields[fieldName] {
			fields = append(fields, fieldName)
			fmt.Printf("[DEBUG] 强制添加系统字段: %s\n", fieldName)
		}
	}

	if len(fields) == 0 {
		return "*", nil
	}

	return strings.Join(fields, ", "), nil
}

// processFieldsForCreate 处理创建时的字段
func (s *service) processFieldsForCreate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	for _, col := range columns {
		// TODO: 检查字段权限 - 需要集成groups权限服务

		// 检查MASK可编辑性
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("add") {
				continue
			}
		}

		// 获取字段值
		value, exists := data[col.DbName]
		if exists {
			// 验证FK字段值
			if col.SetValueType == "fk" && col.RefTableID != nil && value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			processedData[col.DbName] = value
		}
	}

	return processedData, nil
}

// processFieldsForUpdate 处理更新时的字段
func (s *service) processFieldsForUpdate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	for _, col := range columns {
		// TODO: 检查字段权限 - 需要集成groups权限服务

		// 检查MASK可编辑性
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("edit") {
				continue
			}
		}

		// 获取字段值
		value, exists := data[col.DbName]
		if exists {
			// 验证FK字段值
			if col.SetValueType == "fk" && col.RefTableID != nil && value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			processedData[col.DbName] = value
		}
	}

	return processedData, nil
}
