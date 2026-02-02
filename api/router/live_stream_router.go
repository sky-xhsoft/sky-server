package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// SetupLiveStreamRoutes 设置直播流路由
func SetupLiveStreamRoutes(r *gin.RouterGroup, service liveService.Service) {
	streamHandler := handler.NewLiveStreamHandler(service)

	streams := r.Group("/streams")
	{
		streams.GET("/online", streamHandler.GetOnlineStreams)   // 查询在线流列表
		streams.GET("/history", streamHandler.GetHistoryStreams) // 查询历史流列表
		streams.GET("/events", streamHandler.GetStreamEvents)    // 查询推断流事件
		streams.POST("/drop", streamHandler.DropStream)          // 断开直播推流
		streams.POST("/resume", streamHandler.ResumeStream)      // 恢复直播推流
	}
}
