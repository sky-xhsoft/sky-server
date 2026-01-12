package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
)

// AuthRequired 认证中间件
func AuthRequired(jwtUtil *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      20002,
				"message":   "未提供认证令牌",
				"timestamp": "2026-01-11T00:00:00Z",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      20002,
				"message":   "认证令牌格式错误",
				"timestamp": "2026-01-11T00:00:00Z",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证JWT token
		claims, err := jwtUtil.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      20003,
				"message":   "令牌无效或已过期",
				"timestamp": "2026-01-11T00:00:00Z",
			})
			c.Abort()
			return
		}

		// 将用户信息存入context
		c.Set("userID", claims.UserID)
		c.Set("companyID", claims.CompanyID)
		c.Set("username", claims.Username)
		c.Set("clientType", claims.ClientType)
		c.Set("deviceID", claims.DeviceID)

		c.Next()
	}
}
