package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// registerSpeechTaskRoutes 注册声影同传任务路由
func registerSpeechTaskRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, speechTaskService liveService.SpeechTaskService, groupService groups.Service) {
	speechTaskHandler := handler.NewSpeechTaskHandler(speechTaskService)

	// 声影同传任务路由组（需要认证 + 表权限, 表ID=39 同直播权限）
	group := r.Group("/live/speech-tasks")
	group.Use(middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, 39, groups.PermCreate))
	{
		group.GET("", speechTaskHandler.ListSpeechTasks)          // 查询声影同传任务列表
		group.POST("", speechTaskHandler.CreateSpeechTask)        // 创建声影同传任务
		group.POST("/delete", speechTaskHandler.DeleteSpeechTask) // 删除声影同传任务
	}
}