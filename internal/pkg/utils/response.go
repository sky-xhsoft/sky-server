package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
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
		Message:   "创建成功",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// NoContent 无内容响应（删除成功）
func NoContent(c *gin.Context) {
	c.JSON(204, Response{
		Code:      204,
		Message:   "删除成功",
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
		resp.Message = appErr.Message
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
