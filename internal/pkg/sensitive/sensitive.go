// Package sensitive 遮盖请求/响应内容中的密码、密钥、令牌等敏感字段，用于审计日志等落库场景
package sensitive

import (
	"encoding/json"
	"regexp"
	"strings"
)

// sensitiveKeyParts 字段名（转小写并去掉 _ 和 -）包含任一片段即视为敏感
var sensitiveKeyParts = []string{"password", "passwd", "pwd", "secret", "token", "accesskey", "apikey", "privatekey", "credential", "callbackkey"}

const maskedValue = "***"

// formSensitivePattern 非 JSON 请求体（如表单）中 key=value 形式的敏感字段
var formSensitivePattern = regexp.MustCompile(`(?i)([a-z0-9_\-]*(password|passwd|pwd|secret|token|access_?key|api_?key|private_?key|credential|callback_?key)[a-z0-9_\-]*=)[^&\s]*`)

// isSensitiveKey 判断字段名是否敏感（不区分大小写，忽略 _ 和 -）
func isSensitiveKey(key string) bool {
	k := strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(key))
	for _, part := range sensitiveKeyParts {
		if strings.Contains(k, part) {
			return true
		}
	}
	return false
}

// maskSensitive 递归遮盖 JSON 结构中敏感字段的值
func maskSensitive(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if isSensitiveKey(k) && val != nil {
				t[k] = maskedValue
			} else {
				t[k] = maskSensitive(val)
			}
		}
	case []interface{}:
		for i, val := range t {
			t[i] = maskSensitive(val)
		}
	}
	return v
}

// Filter 遮盖请求/响应体中的密码、密钥、令牌等敏感字段的值
func Filter(data string) string {
	trimmed := strings.TrimSpace(data)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		dec := json.NewDecoder(strings.NewReader(trimmed))
		dec.UseNumber()
		var v interface{}
		if err := dec.Decode(&v); err == nil {
			if out, err := json.Marshal(maskSensitive(v)); err == nil {
				return string(out)
			}
		}
	}
	return formSensitivePattern.ReplaceAllString(data, "${1}"+maskedValue)
}
