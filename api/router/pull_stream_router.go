package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// registerPullStreamRoutes 注册拉流任务路由
func registerPullStreamRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, service liveService.Service) {
	pullStreamHandler := handler.NewPullStreamHandler(service)

	// 拉流任务路由组（需要认证）
	pullStream := r.Group("/live/pull-stream")
	pullStream.Use(middleware.AuthRequired(jwtUtil))
	{
		tasks := pullStream.Group("/tasks")
		{
			tasks.POST("", pullStreamHandler.CreatePullStreamTask)              // 创建拉流任务
			tasks.GET("", pullStreamHandler.GetPullStreamTasks)                 // 查询拉流任务列表
			tasks.PUT("/:id", pullStreamHandler.UpdatePullStreamTask)           // 更新拉流任务
			tasks.DELETE("/:id", pullStreamHandler.DeletePullStreamTask)        // 删除拉流任务
			tasks.GET("/:id/status", pullStreamHandler.GetPullStreamTaskStatus) // 查询拉流任务状态
			tasks.POST("/:id/restart", pullStreamHandler.RestartPullStreamTask) // 重启拉流任务
		}
	}
}
