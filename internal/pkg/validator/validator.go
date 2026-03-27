package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// 使用字段名作为标签名称
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidationError 验证错误
type ValidationError struct {
	Errors map[string]string `json:"errors"`
}

// NewValidationError 创建验证错误
func NewValidationError(errs map[string]string) *ValidationError {
	return &ValidationError{
		Errors: errs,
	}
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation error"
	}

	var parts []string
	for field, msg := range e.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}

	return strings.Join(parts, "; ")
}

// ValidateStruct 验证结构体
func ValidateStruct(data interface{}) error {
	if data == nil {
		return fmt.Errorf("validate: data can not be nil")
	}

	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	// 处理验证错误
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return fmt.Errorf("validate: invalid validation error")
	}

	validationErrors := err.(validator.ValidationErrors)
	errs := make(map[string]string)
	for _, ve := range validationErrors {
		field := ve.Field()
		errs[field] = getValidationMessage(ve.Tag(), ve.Param())
	}

	return NewValidationError(errs)
}

// getValidationMessage 获取验证失败消息
func getValidationMessage(tag, param string) string {
	switch tag {
	case "required":
		return "此字段必填"
	case "email":
		return "请输入有效的邮箱地址"
	case "min":
		return fmt.Sprintf("最小值为 %s", param)
	case "max":
		return fmt.Sprintf("最大值为 %s", param)
	case "minLength":
		return fmt.Sprintf("最小长度为 %s", param)
	case "maxLength":
		return fmt.Sprintf("最大长度为 %s", param)
	case "len":
		return fmt.Sprintf("长度必须为 %s", param)
	case "url":
		return "请输入有效的URL"
	case "mobile":
		return "请输入有效的手机号码"
	default:
		return fmt.Sprintf("验证失败: %s", tag)
	}
}

// RegisterValidator 注册自定义验证器
func RegisterValidator(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}

// RegisterStructLevelValidator 注册结构体级别验证器
func RegisterStructLevelValidator(fn validator.StructLevelFunc, types ...interface{}) {
	validate.RegisterStructValidation(fn, types...)
}

// GetValidator 获取验证器实例
func GetValidator() *validator.Validate {
	return validate
}
