package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"gorm.io/gorm"
)

// SetupLiveCallbackRoutes 设置直播回调路由
func SetupLiveCallbackRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, db *gorm.DB, callbackKey string) {
	callbackHandler := handler.NewLiveCallbackHandler(db, callbackKey)

	callback := rg.Group("/callback")
	{
		// 公开回调接口（不需要认证，由腾讯云调用）
		// 基础事件
		callback.POST("/push", callbackHandler.HandlePushStream)
		callback.POST("/disconnect", callbackHandler.HandleDisconnectStream)

		// 录制相关
		callback.POST("/recording-file", callbackHandler.HandleRecordingFile)
		callback.POST("/recording-status", callbackHandler.HandleRecordingStatus)
		callback.POST("/record-exception", callbackHandler.HandleRecordException)

		// 截图和审核
		callback.POST("/screenshot", callbackHandler.HandleScreenshot)
		callback.POST("/video-audit", callbackHandler.HandleVideoAudit)
		callback.POST("/audio-audit", callbackHandler.HandleAudioAudit)

		// 质检和评测
		callback.POST("/quality-inspection", callbackHandler.HandleQualityInspection)
		callback.POST("/quality-threshold", callbackHandler.HandleQualityThreshold)
		callback.POST("/quality-average", callbackHandler.HandleQualityAverage)

		// AI功能
		callback.POST("/smart-erase", callbackHandler.HandleSmartErase)
		callback.POST("/subtitle", callbackHandler.HandleSubtitle)
		callback.POST("/summary", callbackHandler.HandleSummary)
		callback.POST("/highlight", callbackHandler.HandleHighlight)

		// 异常和监控
		callback.POST("/push-exception", callbackHandler.HandlePushException)
		callback.POST("/pull-stream", callbackHandler.HandlePullStream)
		callback.POST("/monitor", callbackHandler.HandleMonitor)

		// 需要认证的查询接口
		authenticated := callback.Group("")
		authenticated.Use(middleware.AuthRequired(jwtUtil))
		{
			authenticated.GET("/events", callbackHandler.QueryCallbackEvents)
			authenticated.DELETE("/events/:id", callbackHandler.DeleteCallbackEvent)
			authenticated.DELETE("/events/batch", callbackHandler.BatchDeleteCallbackEvents)
		}
	}
}
