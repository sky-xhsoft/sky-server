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
func registerCompanyRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, companyService company.Service, smsService sms.Service, storageConfig cloud.StorageConfigService) {
	companyHandler := handler.NewCompanyHandler(companyService, smsService)
	if storageConfig != nil {
		companyHandler.WithStorageRefresher(storageConfig)
	}

	companyRg := rg.Group("/company")
	{
		// 公开接口
		companyRg.POST("/register", companyHandler.Register)

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
