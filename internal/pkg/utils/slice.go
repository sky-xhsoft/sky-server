package utils

// ContainsInt 检查int切片是否包含指定元素
func ContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsString 检查string切片是否包含指定元素
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicateInt 去除int切片中的重复元素
func RemoveDuplicateInt(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// RemoveDuplicateString 去除string切片中的重复元素
func RemoveDuplicateString(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
