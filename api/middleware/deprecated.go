package middleware

import (
	"github.com/gin-gonic/gin"
)

// DeprecatedAPI 返回一个废弃API警告中间件
// 该中间件会在响应头中添加废弃警告信息，但不影响功能
func DeprecatedAPI(newEndpoint string, deprecationDate string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加废弃警告头
		c.Header("X-API-Deprecated", "true")
		c.Header("X-API-Deprecated-Date", deprecationDate)
		c.Header("X-API-New-Endpoint", newEndpoint)
		c.Header("Warning", `299 - "This API endpoint is deprecated and will be removed in a future version. Please use `+newEndpoint+` instead."`)

		// 继续处理请求
		c.Next()
	}
}
