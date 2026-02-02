package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// registerLiveDomainRoutes 注册直播域名路由
func registerLiveDomainRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, service liveService.Service) {
	domainHandler := handler.NewLiveDomainHandler(service)

	// 直播域名路由组（需要认证）
	domains := r.Group("/live/domains")
	domains.Use(middleware.AuthRequired(jwtUtil))
	{
		domains.POST("", domainHandler.AddDomain)                       // 添加域名
		domains.GET("", domainHandler.ListDomains)                      // 查询域名列表
		domains.GET("/verify", domainHandler.VerifyDomainOwner)         // 验证域名归属
		domains.GET("/:domainName", domainHandler.GetDomain)            // 查询域名信息
		domains.DELETE("/:domainName", domainHandler.DeleteDomain)      // 删除域名
		domains.POST("/:domainName/enable", domainHandler.EnableDomain) // 启用域名
		domains.POST("/:domainName/forbid", domainHandler.ForbidDomain) // 禁用域名
	}
}
