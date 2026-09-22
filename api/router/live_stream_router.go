package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// SetupLiveStreamRoutes 设置直播流路由
func SetupLiveStreamRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, service liveService.Service, groupService groups.Service) {
	streamHandler := handler.NewLiveStreamHandler(service)

	streams := r.Group("/streams")
	streams.Use(middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, 38, groups.PermRead))
	{
		streams.GET("/online", streamHandler.GetOnlineStreams)              // 查询在线流列表
		streams.GET("/history", streamHandler.GetHistoryStreams)            // 查询历史流列表
		streams.GET("/events", streamHandler.GetStreamEvents)               // 查询推断流事件
		streams.GET("/:streamName/metrics", streamHandler.GetStreamMetrics) // 获取直播流实时指标
		streams.POST("/drop", streamHandler.DropStream)                     // 断开直播推流
		streams.POST("/resume", streamHandler.ResumeStream)                 // 恢复直播推流
	}
}

// registerLiveStreamRoutes 注册直播流路由
func registerLiveStreamRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, service liveService.Service, groupService groups.Service) {
	SetupLiveStreamRoutes(rg.Group("/live"), jwtUtil, service, groupService)
}
