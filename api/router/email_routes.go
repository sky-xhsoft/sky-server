package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/internal/service/email"
)

// registerEmailRoutes 注册邮件路由
func registerEmailRoutes(rg *gin.RouterGroup, emailService email.Service) {
	emailHandler := handler.NewEmailHandler(emailService)
	emailRg := rg.Group("/email")
	{
		emailRg.POST("/verify-code", emailHandler.SendVerificationCode)
	}
}
