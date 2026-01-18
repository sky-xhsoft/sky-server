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
// 支持多种类型转换：uint, int, int64, float64
func GetUintFromParams(params map[string]interface{}, key string) (uint, error) {
	value, exists := params[key]
	if !exists {
		return 0, fmt.Errorf("参数 %s 不存在", key)
	}

	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case int64:
		return uint(v), nil
	case float64:
		return uint(v), nil
	default:
		return 0, fmt.Errorf("参数 %s 类型不正确: %T", key, value)
	}
}

// GetUintOrZero 从 params 中获取 uint 类型的值，失败时返回 0
func GetUintOrZero(params map[string]interface{}, key string) uint {
	value, err := GetUintFromParams(params, key)
	if err != nil {
		return 0
	}
	return value
}

// GetStringFromParams 从 params 中获取 string 类型的值
func GetStringFromParams(params map[string]interface{}, key string) (string, error) {
	value, exists := params[key]
	if !exists {
		return "", fmt.Errorf("参数 %s 不存在", key)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("参数 %s 类型不正确: %T", key, value)
	}

	return str, nil
}

// GetStringOrEmpty 从 params 中获取 string 类型的值，失败时返回空字符串
func GetStringOrEmpty(params map[string]interface{}, key string) string {
	value, err := GetStringFromParams(params, key)
	if err != nil {
		return ""
	}
	return value
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
