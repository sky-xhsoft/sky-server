package crud

import (
	"encoding/json"
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

		// 主键字段、系统审计字段和外键字段始终包含，不受MASK限制
		isPrimaryKey := col.IsAK == "Y" || col.SetValueType == "pk"
		isStandardField := standardFields[col.DbName]
		isForeignKey := col.SetValueType == "fk"

		if isStandardField {
			addedStandardFields[col.DbName] = true
		}

		if !isPrimaryKey && !isStandardField && !isForeignKey {
			// 对于业务字段（非FK），检查MASK可见性
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

		// 获取字段值
		value, exists := data[col.DbName]

		// 特殊处理：外键字段（作为子表的父表关联字段）
		// 如果请求数据中包含外键字段值，即使 MASK 不允许编辑，也应该接受该值
		// 这是为了支持主从表（父子表）场景，前端会自动设置外键字段的值
		if exists && col.SetValueType == "fk" && col.RefTableID != nil {
			// 验证FK字段值
			if value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			// 外键字段允许传入，不受 MASK 限制
			processedData[col.DbName] = value
			continue
		}

		// 检查MASK可编辑性（非外键字段需要遵守 MASK）
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("add") {
				continue
			}
		}

		// 普通字段处理
		if exists {
			// 特殊处理：JSON 类型字段
			// 如果 DisplayType 是 json 且值是 map 或 slice，需要序列化为 JSON 字符串
			if col.DisplayType == "json" && value != nil {
				switch v := value.(type) {
				case map[string]interface{}, []interface{}:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				case string:
					// 如果已经是字符串，直接使用
					processedData[col.DbName] = v
				default:
					// 其他类型尝试序列化
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				}
			} else {
				// 特殊处理：自动转大写
				// 如果字段配置了 IS_UPPERCASE = 'Y'，且值是字符串类型，则转换为大写
				if col.IsUppercase == "Y" && value != nil {
					if strValue, ok := value.(string); ok {
						processedData[col.DbName] = strings.ToUpper(strValue)
					} else {
						processedData[col.DbName] = value
					}
				} else {
					processedData[col.DbName] = value
				}
			}
		}
	}

	return processedData, nil
}

// processFieldsForUpdate 处理更新时的字段
func (s *service) processFieldsForUpdate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	for _, col := range columns {
		// TODO: 检查字段权限 - 需要集成groups权限服务

		// 获取字段值
		value, exists := data[col.DbName]

		// 特殊处理：外键字段（作为子表的父表关联字段）
		// 如果请求数据中包含外键字段值，即使 MASK 不允许编辑，也应该接受该值
		// 这是为了支持主从表（父子表）场景的灵活性
		if exists && col.SetValueType == "fk" && col.RefTableID != nil {
			// 验证FK字段值
			if value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			// 外键字段允许传入，不受 MASK 限制
			processedData[col.DbName] = value
			continue
		}

		// 检查MASK可编辑性（非外键字段需要遵守 MASK）
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("edit") {
				continue
			}
		}

		// 普通字段处理
		if exists {
			// 特殊处理：JSON 类型字段
			// 如果 DisplayType 是 json 且值是 map 或 slice，需要序列化为 JSON 字符串
			if col.DisplayType == "json" && value != nil {
				switch v := value.(type) {
				case map[string]interface{}, []interface{}:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				case string:
					// 如果已经是字符串，直接使用
					processedData[col.DbName] = v
				default:
					// 其他类型尝试序列化
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				}
			} else {
				// 特殊处理：自动转大写
				// 如果字段配置了 IS_UPPERCASE = 'Y'，且值是字符串类型，则转换为大写
				if col.IsUppercase == "Y" && value != nil {
					if strValue, ok := value.(string); ok {
						processedData[col.DbName] = strings.ToUpper(strValue)
					} else {
						processedData[col.DbName] = value
					}
				} else {
					processedData[col.DbName] = value
				}
			}
		}
	}

	return processedData, nil
}
