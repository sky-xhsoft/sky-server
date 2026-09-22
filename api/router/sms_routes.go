package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/internal/service/sms"
)

// registerSmsRoutes 注册短信路由
func registerSmsRoutes(rg *gin.RouterGroup, smsService sms.Service) {
	smsHandler := handler.NewSmsHandler(smsService)
	smsRg := rg.Group("/sms")
	{
		smsRg.POST("/send-code", smsHandler.SendVerifyCode)
	}
}