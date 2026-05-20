package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
)

// registerCompanyRoutes 注册公司管理路由
func registerCompanyRoutes(rg *gin.RouterGroup, companyService company.Service) {
	companyHandler := handler.NewCompanyHandler(companyService)

	companyRg := rg.Group("/company")
	{
		companyRg.POST("/register", companyHandler.Register)
	}
}
