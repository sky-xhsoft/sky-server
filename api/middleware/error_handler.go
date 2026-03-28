package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		// 根据错误类型返回不同的响应
		var appErr *errors.AppError
		if errors.As(err, &appErr) {
			// 业务错误
			logger.Warn("Business error",
				zap.Int("code", appErr.Code),
				zap.String("message", appErr.Message),
				zap.Error(appErr.Err))

			c.JSON(getHTTPStatus(appErr.Code), gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"error":   err.Error(),
			})
			return
		}

		// 未知错误
		logger.Error("Unexpected error",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "服务器内部错误",
			"error":   err.Error(),
		})
	}
}

// getHTTPStatus 根据错误码获取HTTP状态码
func getHTTPStatus(errCode int) int {
	switch errCode {
	case errors.ErrValidation.Code:
		return http.StatusBadRequest
	case errors.ErrInvalidParameter.Code:
		return http.StatusBadRequest
	case errors.ErrResourceNotFound.Code:
		return http.StatusNotFound
	case errors.ErrUnauthorized.Code:
		return http.StatusUnauthorized
	case errors.ErrForbidden.Code:
		return http.StatusForbidden
	case errors.ErrConflict.Code:
		return http.StatusConflict
	case errors.ErrTooManyRequests.Code:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// PanicRecovery panic恢复中间件
func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic recovered",
					zap.String("path", c.Request.URL.Path),
					zap.Any("panic", r),
					zap.Stack("stack"))

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "服务器内部错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
