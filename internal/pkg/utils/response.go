package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/i18n"
	"github.com/sky-xhsoft/sky-server/internal/pkg/validator"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// PageResult 分页结果
type PageResult struct {
	List       interface{} `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Response{
		Code:      201,
		Message:   i18n.TranslateContext(c, "created"),
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// NoContent 无内容响应（删除成功）
func NoContent(c *gin.Context) {
	c.JSON(204, Response{
		Code:      204,
		Message:   i18n.TranslateContext(c, "deleted"),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, err error) {
	resp := Response{
		Code:      httpStatus,
		Message:   err.Error(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 如果是AppError，使用其错误码
	if appErr, ok := err.(*errors.AppError); ok {
		resp.Code = appErr.Code
		// 尝试根据错误码找到对应的翻译
		var key string
		switch appErr.Code {
		case errors.ErrInvalidCredentials:
			key = "invalid_credentials"
		case errors.ErrInvalidParam, errors.ErrValidation:
			key = "invalid_params"
		case errors.ErrUnauthorized:
			key = "unauthorized"
		case errors.ErrForbidden:
			key = "forbidden"
		case errors.ErrInternal:
			key = "internal_error"
		default:
			// 使用原始消息
			key = ""
		}
		if key != "" {
			resp.Message = i18n.TranslateContext(c, key)
		} else {
			resp.Message = appErr.Message
		}
	}

	// 如果是验证错误，格式化错误响应
	if valErr, ok := err.(*validator.ValidationError); ok {
		resp.Code = 400
		resp.Message = i18n.TranslateContext(c, "invalid_params")
		resp.Error = valErr.Errors
		httpStatus = 400
	}

	c.JSON(httpStatus, resp)
}

// BadRequest 参数错误响应
func BadRequest(c *gin.Context, message string) {
	c.JSON(400, Response{
		Code:      400,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Unauthorized 未认证响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, Response{
		Code:      401,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Forbidden 无权限响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(403, Response{
		Code:      403,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	c.JSON(404, Response{
		Code:      404,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// InternalError 服务器内部错误响应
func InternalError(c *gin.Context, message string) {
	c.JSON(500, Response{
		Code:      500,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// PageResponse 分页响应
func PageResponse(c *gin.Context, list interface{}, page, pageSize, total int) {
	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data: PageResult{
			List: list,
			Pagination: Pagination{
				Page:       page,
				PageSize:   pageSize,
				Total:      total,
				TotalPages: totalPages,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// HTTPStatusFor 按业务错误码映射 HTTP 状态码；非 AppError 视为 500
// 登录失败映射为 400 而非 401：前端遇到 401 会按登录过期处理并跳转登录页
func HTTPStatusFor(err error) int {
	var appErr *errors.AppError
	if !errors.As(err, &appErr) {
		return 500
	}
	switch code := appErr.Code; {
	case code >= 10000 && code < 20000, code == errors.ErrInvalidCredentials:
		return 400
	case code == errors.ErrInvalidToken, code == errors.ErrTokenExpired:
		return 401
	case code == errors.ErrPermissionDenied, code == errors.ErrForbidden, code == errors.ErrUnauthorized:
		return 403
	case code == errors.ErrResourceNotFound:
		return 404
	case code == errors.ErrResourceExists, code == errors.ErrResourceConflict:
		return 409
	}
	return 500
}

// FailFromError 按错误码返回对应的 HTTP 状态：4xx 只返回业务提示，5xx 附带前缀与原始错误便于排查
func FailFromError(c *gin.Context, prefix string, err error) {
	status := HTTPStatusFor(err)
	message := prefix + err.Error()
	var appErr *errors.AppError
	if status < 500 && errors.As(err, &appErr) {
		message = appErr.Message
	}
	c.JSON(status, Response{
		Code:      status,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
