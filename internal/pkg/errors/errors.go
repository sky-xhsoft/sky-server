package errors

import "fmt"

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Is 检查错误是否匹配
func Is(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}

	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}

	targetErr, ok := target.(*AppError)
	if !ok {
		return false
	}

	return appErr.Code == targetErr.Code
}

// New 创建新的应用错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// GetCode 获取错误码
func GetCode(err error) int {
	if err == nil {
		return 0
	}
	appErr, ok := err.(*AppError)
	if !ok {
		return ErrInternal
	}
	return appErr.Code
}

// 错误码定义
const (
	// 参数错误 10xxx
	ErrInvalidParam     = 10001
	ErrMissingParam     = 10002
	ErrInvalidParamType = 10003
	ErrValidation       = 10004

	// 认证错误 20xxx
	ErrInvalidCredentials = 20001
	ErrInvalidToken       = 20002
	ErrTokenExpired       = 20003
	ErrPermissionDenied   = 20004
	ErrUnauthorized       = 20005
	ErrForbidden          = 20006

	// 资源错误 30xxx
	ErrResourceNotFound = 30001
	ErrResourceExists   = 30002
	ErrResourceConflict = 30003

	// 数据库错误 40xxx
	ErrDatabase = 40001
	ErrCache    = 40002

	// 服务器错误 50xxx
	ErrInternal      = 50001
	ErrExternalAPI   = 50002
	ErrActionExecute = 50003
)

// 预定义错误

var (
	// 参数错误
	InvalidParam     = New(ErrInvalidParam, "参数错误")
	MissingParam     = New(ErrMissingParam, "参数缺失")
	InvalidParamType = New(ErrInvalidParamType, "参数类型错误")
	ValidationError  = New(ErrValidation, "验证失败")

	// 认证错误
	InvalidCredentials = New(ErrInvalidCredentials, "用户名或密码错误")
	InvalidToken       = New(ErrInvalidToken, "Token无效")
	TokenExpired       = New(ErrTokenExpired, "Token过期")
	PermissionDenied   = New(ErrPermissionDenied, "无权限")
	Unauthorized       = New(ErrUnauthorized, "未认证")
	Forbidden          = New(ErrForbidden, "禁止访问")

	// 资源错误
	ResourceNotFound = New(ErrResourceNotFound, "资源不存在")
	ResourceExists   = New(ErrResourceExists, "资源已存在")
	ResourceConflict = New(ErrResourceConflict, "资源冲突")

	// 数据库错误
	DatabaseError = New(ErrDatabase, "数据库错误")
	CacheError    = New(ErrCache, "缓存错误")

	// 服务器错误
	InternalError      = New(ErrInternal, "服务器内部错误")
	ExternalAPIError   = New(ErrExternalAPI, "第三方服务错误")
	ActionExecuteError = New(ErrActionExecute, "动作执行失败")
)
