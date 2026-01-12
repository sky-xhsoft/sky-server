package builtin

import "fmt"

// 辅助函数：从 map 中获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// 辅助函数：从 map 中获取 int 值
func getIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case uint:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

// 辅助函数：从 map 中获取 uint 值
func getUintValue(m map[string]interface{}, key string) uint {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case uint:
			return val
		case int:
			if val >= 0 {
				return uint(val)
			}
		case int64:
			if val >= 0 {
				return uint(val)
			}
		case float64:
			if val >= 0 {
				return uint(val)
			}
		}
	}
	return 0
}

// 辅助函数：从 params 中获取 uint 类型的 ID（处理多种类型转换）
func extractUintID(params map[string]interface{}, key string) (uint, error) {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case uint:
			return val, nil
		case int:
			if val >= 0 {
				return uint(val), nil
			}
		case int64:
			if val >= 0 {
				return uint(val), nil
			}
		case float64:
			if val >= 0 {
				return uint(val), nil
			}
		}
	}
	return 0, fmt.Errorf("无法从参数中获取 %s", key)
}
