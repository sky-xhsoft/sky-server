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
func registerSpeechTaskRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, speechTaskService liveService.SpeechTaskService, glossary liveService.GlossaryExtractor, groupService groups.Service) {
	speechTaskHandler := handler.NewSpeechTaskHandler(speechTaskService)

	// 声影同传任务路由组（需要认证 + 表权限, 表ID=39 同直播权限）
	group := r.Group("/live/speech-tasks")
	group.Use(middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, 39, groups.PermCreate))
	{
		group.GET("", speechTaskHandler.ListSpeechTasks)                // 查询声影同传任务列表
		group.POST("", speechTaskHandler.CreateSpeechTask)              // 创建声影同传任务
		group.POST("/delete", speechTaskHandler.DeleteSpeechTask)       // 彻底删除（火山连同历史）
		group.POST("/stop", speechTaskHandler.StopSpeechTask)           // 停止（保留历史，可复制新建）
		group.POST("/restart", speechTaskHandler.RestartSpeechTask)     // 重启运行中 / 启动中的任务
		group.GET("/:taskId/rule", speechTaskHandler.GetSpeechTaskRule) // 创建参数（复制新建）
		if glossary != nil {
			group.POST("/glossary/extract", handler.NewSpeechGlossaryHandler(glossary).Extract) // 整篇原文译文智能提取词库
		}
	}
}
