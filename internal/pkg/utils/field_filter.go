package utils

import (
	"reflect"
	"strings"
)

// FieldFilter 提供 API 响应字段裁剪功能
// 通过 fields 参数动态控制返回的字段，减少响应体积
// 使用示例: GET /api/v1/data/users?fields=id,name,email

// FilterFields 根据指定的字段列表过滤结构体字段
// obj 可以是 map[string]interface{} 或 struct
// fields 是要保留的字段名（逗号分隔），为空则返回全部字段
func FilterFields(obj interface{}, fields string) interface{} {
	if fields == "" {
		return obj
	}

	fieldSet := parseFields(fields)
	if len(fieldSet) == 0 {
		return obj
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return filterMapFields(v, fieldSet)
	case []map[string]interface{}:
		result := make([]map[string]interface{}, len(v))
		for i, item := range v {
			result[i] = filterMapFields(item, fieldSet)
		}
		return result
	default:
		return filterStructFields(obj, fieldSet)
	}
}

// parseFields 解析逗号分隔的字段列表
func parseFields(fields string) map[string]bool {
	result := make(map[string]bool)
	for _, f := range strings.Split(fields, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			result[f] = true
		}
	}
	return result
}

// filterMapFields 过滤 map 中的字段，只保留指定字段
func filterMapFields(m map[string]interface{}, fields map[string]bool) map[string]interface{} {
	if len(fields) == 0 {
		return m
	}

	result := make(map[string]interface{})
	for k, v := range m {
		if fields[k] {
			result[k] = v
		}
	}
	return result
}

// filterStructFields 使用反射过滤结构体的指定字段
func filterStructFields(obj interface{}, fields map[string]bool) interface{} {
	if len(fields) == 0 {
		return obj
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return obj
	}

	result := make(map[string]interface{})
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 获取 JSON 字段名
		jsonTag := getJSONFieldName(field)
		if jsonTag == "-" {
			continue
		}
		if jsonTag == "" {
			jsonTag = field.Name
		}

		if fields[jsonTag] || fields[field.Name] {
			result[jsonTag] = v.Field(i).Interface()
		}
	}

	return result
}

// getJSONFieldName 从 json 标签中提取字段名
func getJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return ""
	}

	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return field.Name
	}

	return parts[0]
}

// SelectFields 从 []string 转换为 map[string]bool
func SelectFields(fields []string) map[string]bool {
	result := make(map[string]bool, len(fields))
	for _, f := range fields {
		result[strings.TrimSpace(f)] = true
	}
	return result
}