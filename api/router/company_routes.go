package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
	"github.com/sky-xhsoft/sky-server/internal/service/sms"
)

// registerCompanyRoutes 注册公司管理路由
// registerCheckLimiter 注册查重：每个 IP 每秒 2 次，突发 10 次
var registerCheckLimiter = middleware.NewIPRateLimiter(2, 10)

func registerCompanyRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, companyService company.Service, smsService sms.Service, storageConfig cloud.StorageConfigService) {
	companyHandler := handler.NewCompanyHandler(companyService, smsService)
	if storageConfig != nil {
		companyHandler.WithStorageRefresher(storageConfig)
	}

	companyRg := rg.Group("/company")
	{
		// 公开接口
		companyRg.POST("/register", companyHandler.Register)
		// 注册前查重（公开，按 IP 限流，防止批量探测账号）
		companyRg.GET("/register/check", registerCheckLimiter.Limit(), companyHandler.CheckRegister)

		// 需要认证的公司配置接口
		config := companyRg.Group("/config")
		config.Use(middleware.AuthRequired(jwtUtil))
		{
			config.GET("", companyHandler.GetCompanyConfig)
			config.PUT("", middleware.AdminRequired(), companyHandler.UpdateCompanyConfig)
			// 云盘存储：列出、新建存储桶（用云盘存储密钥，不修改配置）
			config.GET("/cos-buckets", middleware.AdminRequired(), companyHandler.ListStorageBuckets)
			config.POST("/cos-buckets", middleware.AdminRequired(), companyHandler.CreateStorageBucket)
		}
	}
}
