package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 为每个请求添加唯一请求ID，便于日志追踪和问题排查
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用客户端传入的 X-Request-ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}