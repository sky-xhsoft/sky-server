package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
)

// AuthRequired 认证中间件
func AuthRequired(jwtUtil *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      pkgerrors.ErrInvalidToken,
				"message":   "未提供认证令牌",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      pkgerrors.ErrInvalidToken,
				"message":   "认证令牌格式错误",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := jwtUtil.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      pkgerrors.ErrTokenExpired,
				"message":   "令牌无效或已过期",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 只接受访问令牌：刷新令牌（有效期更长）不能直接用来访问接口
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      pkgerrors.ErrInvalidToken,
				"message":   "认证令牌类型错误",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 账号已停用：已签发的令牌同样失效（状态缓存 30 秒）
		status, known := resolveUserStatus(claims.UserID)
		if known && !status.Active {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      pkgerrors.ErrInvalidToken,
				"message":   "账号已停用",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}
		c.Set("isAdmin", status.Admin)
		if known {
			c.Set("fieldGrade", status.FieldGrade)
		}

		// 将用户信息存入context
		c.Set("userID", claims.UserID)
		c.Set("companyID", claims.CompanyID)
		c.Set("username", claims.Username)
		c.Set("clientType", claims.ClientType)
		c.Set("deviceID", claims.DeviceID)
		c.Set("token", token)

		c.Next()
	}
}

// OptionalAuth 可选鉴权：未携带令牌按匿名放行；携带了令牌则必须有效（无效返回 401，前端可刷新令牌后重试）
func OptionalAuth(jwtUtil *jwt.JWT) gin.HandlerFunc {
	required := AuthRequired(jwtUtil)
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.Next()
			return
		}
		required(c)
	}
}
