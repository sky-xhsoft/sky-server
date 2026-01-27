package hooks

import (
	"fmt"

	"gorm.io/gorm"
)

// GetDBFromParams 从 params 中获取数据库连接
func GetDBFromParams(params map[string]interface{}) (*gorm.DB, error) {
	txDB, ok := params["__db__"].(*gorm.DB)
	if !ok || txDB == nil {
		return nil, fmt.Errorf("无法获取数据库连接")
	}
	return txDB, nil
}

// GetUintFromParams 从 params 中获取 uint 类型的值
// 使用 ConvertToUint 进行类型转换，支持所有数字类型
func GetUintFromParams(params map[string]interface{}, key string) (uint, error) {
	value, exists := params[key]
	if !exists {
		return 0, fmt.Errorf("参数 %s 不存在", key)
	}

	result, err := ConvertToUint(value)
	if err != nil {
		return 0, fmt.Errorf("参数 %s 转换失败: %w (实际类型: %T, 实际值: %v)", key, err, value, value)
	}

	return result, nil
}

// GetUintOrZero 从 params 中获取 uint 类型的值，失败时返回 0
func GetUintOrZero(params map[string]interface{}, key string) uint {
	value, err := GetUintFromParams(params, key)
	if err != nil {
		return 0
	}
	return value
}

// ConvertToUint 将任意类型的值转换为 uint
// 支持所有数字类型：int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
func ConvertToUint(value interface{}) (uint, error) {
	if value == nil {
		return 0, fmt.Errorf("值为 nil")
	}

	switch v := value.(type) {
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %d", v)
		}
		return uint(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %d", v)
		}
		return uint(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %d", v)
		}
		return uint(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %d", v)
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %d", v)
		}
		return uint(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %f", v)
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为 uint: %f", v)
		}
		return uint(v), nil
	default:
		return 0, fmt.Errorf("不支持的类型: %T", value)
	}
}

// GetUintFromMap 从 map 中获取 uint 类型的值
// 使用 ConvertToUint 进行类型转换，支持所有数字类型
func GetUintFromMap(data map[string]interface{}, key string) (uint, error) {
	value, exists := data[key]
	if !exists {
		return 0, fmt.Errorf("字段 %s 不存在", key)
	}

	result, err := ConvertToUint(value)
	if err != nil {
		return 0, fmt.Errorf("字段 %s 转换失败: %w (实际类型: %T, 实际值: %v)", key, err, value, value)
	}

	return result, nil
}

// GetStringFromParams 从 params 中获取 string 类型的值
func GetStringFromParams(params map[string]interface{}, key string) (string, error) {
	value, exists := params[key]
	if !exists {
		return "", fmt.Errorf("参数 %s 不存在", key)
	}

	result, err := ConvertToString(value)
	if err != nil {
		return "", fmt.Errorf("参数 %s 转换失败: %w (实际类型: %T, 实际值: %v)", key, err, value, value)
	}

	return result, nil
}

// GetStringOrEmpty 从 params 中获取 string 类型的值，失败时返回空字符串
func GetStringOrEmpty(params map[string]interface{}, key string) string {
	value, err := GetStringFromParams(params, key)
	if err != nil {
		return ""
	}
	return value
}

// ConvertToString 将任意类型的值转换为 string
func ConvertToString(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("不支持的类型: %T", value)
	}
}

// GetStringFromMap 从 map 中获取 string 类型的值
func GetStringFromMap(data map[string]interface{}, key string) (string, error) {
	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("字段 %s 不存在", key)
	}

	result, err := ConvertToString(value)
	if err != nil {
		return "", fmt.Errorf("字段 %s 转换失败: %w (实际类型: %T, 实际值: %v)", key, err, value, value)
	}

	return result, nil
}

// SuccessResult 创建成功的返回结果
func SuccessResult(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"message": message,
	}
}

// ErrorResult 创建错误的返回结果
func ErrorResult(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": message,
	}
}
