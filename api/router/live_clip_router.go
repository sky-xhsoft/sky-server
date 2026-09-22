package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// registerLiveClipRoutes 注册直播切片/录制路由
func registerLiveClipRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, clipService liveService.ClipService, groupService groups.Service) {
	clipHandler := handler.NewLiveClipHandler(clipService)

	// 直播切片路由（需要认证）
	clips := r.Group("/live/clips")
	clips.Use(middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, 40, groups.PermRead))
	{
		clips.GET("", clipHandler.ListClips) // 查询切片列表
	}

	// 直播录制路由（需要认证）
	recordings := r.Group("/live/recordings")
	recordings.Use(middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, 41, groups.PermRead))
	{
		recordings.GET("", clipHandler.ListRecordings) // 查询录制列表
	}
}
