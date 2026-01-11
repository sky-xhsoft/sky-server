package utils

import (
	"strings"
)

// ToUpper 转换为大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 检查字符串是否非空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// DefaultString 返回默认值（如果为空）
func DefaultString(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// Contains 检查是否包含子字符串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ContainsAny 检查是否包含任一子字符串
func ContainsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if Contains(s, substr) {
			return true
		}
	}
	return false
}
