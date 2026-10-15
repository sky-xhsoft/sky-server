package middleware

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/config"
)

// CORS 跨域中间件；subdomainSuffix 来自 app.subdomainBase（如 ".zhibo.skyzhou.cn"），为空时不动态放行子域
func CORS(cfg config.CORSConfig, subdomainSuffix string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowOriginFunc: func(origin string) bool {
			// 动态放行公司子域（*<subdomainSuffix>）的任何子域、任意端口（多公司自动适配，免改配置）
			// origin 形如 http://1.zhibo.skyzhou.cn:5173，需剥离协议与端口后取 host
			if subdomainSuffix == "" {
				return false
			}
			u, err := url.Parse(origin)
			if err == nil && u.Hostname() != "" && strings.HasSuffix(u.Hostname(), subdomainSuffix) {
				return true
			}
			return false
		},
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	})
}
