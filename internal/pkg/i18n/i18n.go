package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Language 定义支持的语言类型
type Language string

const (
	Chinese Language = "zh-CN"
	English Language = "en-US"
)

// DefaultLanguage 默认语言
const DefaultLanguage = Chinese

// Translations 翻译字典
var Translations = map[Language]map[string]string{
	Chinese: {
		"success":                 "success",
		"created":                 "创建成功",
		"deleted":                 "删除成功",
		"invalid_credentials":     "用户名或密码错误",
		"old_password_incorrect":  "旧密码不正确",
		"user_not_found":          "用户不存在",
		"change_password_success": "密码修改成功，请重新登录",
		"invalid_params":          "参数验证失败",
		"unauthorized":            "未授权",
		"forbidden":               "无权限",
		"not_found":               "资源不存在",
		"internal_error":          "服务器内部错误",
	},
	English: {
		"success":                 "success",
		"created":                 "Created successfully",
		"deleted":                 "Deleted successfully",
		"invalid_credentials":     "Invalid username or password",
		"old_password_incorrect":  "Old password is incorrect",
		"user_not_found":          "User not found",
		"change_password_success": "Password changed successfully, please login again",
		"invalid_params":          "Parameter validation failed",
		"unauthorized":            "Unauthorized",
		"forbidden":               "Forbidden",
		"not_found":               "Resource not found",
		"internal_error":          "Internal server error",
	},
}

// GetLanguageFromRequest 从请求中获取语言
func GetLanguageFromRequest(c *gin.Context) Language {
	// 从请求头中获取 Accept-Language
	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang == "" {
		return DefaultLanguage
	}

	// 简单的语言检测
	if strings.HasPrefix(strings.ToLower(acceptLang), "en") {
		return English
	}

	return DefaultLanguage
}

// Translate 翻译函数
func Translate(lang Language, key string) string {
	if langMap, ok := Translations[lang]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}

	// 回退到默认语言
	if langMap, ok := Translations[DefaultLanguage]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}

	// 如果都没有找到，返回原始 key
	return key
}

// ContextKey 用于在 gin.Context 中存储语言的键
type ContextKey string

const LanguageKey ContextKey = "language"

// Middleware 语言中间件，用于从请求中获取语言并存储到 gin.Context 中
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLanguageFromRequest(c)
		c.Set(string(LanguageKey), lang)
		c.Next()
	}
}

// GetLanguageFromContext 从 gin.Context 中获取语言
func GetLanguageFromContext(c *gin.Context) Language {
	if lang, exists := c.Get(string(LanguageKey)); exists {
		if l, ok := lang.(Language); ok {
			return l
		}
	}
	return DefaultLanguage
}

// TranslateContext 从 gin.Context 中获取语言并翻译
func TranslateContext(c *gin.Context, key string) string {
	lang := GetLanguageFromContext(c)
	return Translate(lang, key)
}
