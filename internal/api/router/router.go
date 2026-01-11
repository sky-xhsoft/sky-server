package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/api/handler"
	"github.com/sky-xhsoft/sky-server/internal/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/action"
	"github.com/sky-xhsoft/sky-server/internal/service/audit"
	"github.com/sky-xhsoft/sky-server/internal/service/crud"
	"github.com/sky-xhsoft/sky-server/internal/service/dict"
	"github.com/sky-xhsoft/sky-server/internal/service/file"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/menu"
	"github.com/sky-xhsoft/sky-server/internal/service/message"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/internal/service/sequence"
	"github.com/sky-xhsoft/sky-server/internal/service/sso"
	"github.com/sky-xhsoft/sky-server/internal/service/workflow"
	ws "github.com/sky-xhsoft/sky-server/internal/pkg/websocket"
	"go.uber.org/zap"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Services 服务集合
type Services struct {
	SSO       sso.Service
	Metadata  metadata.Service
	Dict      dict.Service
	Sequence  sequence.Service
	CRUD      crud.Service
	Action    action.Service
	Workflow  workflow.Service
	Audit     audit.Service
	Groups    groups.Service
	Menu      menu.Service
	File      file.Service
	Message   message.Service
	WSManager *ws.Manager
}

// Setup 设置路由
func Setup(engine *gin.Engine, cfg *config.Config, jwtUtil *jwt.JWT, services *Services, logger *zap.Logger) {
	// 全局中间件
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS(cfg.CORS))

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "UP",
			"timestamp": "2026-01-11T00:00:00Z",
		})
	})

	// Swagger文档
	if cfg.Swagger.Enabled {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API路由组
	v1 := engine.Group("/api/v1")
	v1.Use(middleware.AuditLogger(services.Audit)) // 审计日志中间件
	{
		// 注册认证路由
		registerAuthRoutes(v1, jwtUtil, services.SSO)

		// 注册元数据路由
		registerMetadataRoutes(v1, jwtUtil, services.Metadata)

		// 注册字典路由
		registerDictRoutes(v1, jwtUtil, services.Dict)

		// 注册序号路由
		registerSequenceRoutes(v1, jwtUtil, services.Sequence)

		// 注册通用CRUD路由
		registerCRUDRoutes(v1, jwtUtil, services.CRUD)

		// 注册动作路由
		registerActionRoutes(v1, jwtUtil, services.Action)

		// 注册工作流路由
		registerWorkflowRoutes(v1, jwtUtil, services.Workflow)

		// 注册审计日志路由
		registerAuditRoutes(v1, jwtUtil, services.Audit)

		// 注册权限组管理路由
		registerGroupsRoutes(v1, jwtUtil, services.Groups)

		// 注册安全目录管理路由
		registerDirectoryRoutes(v1, jwtUtil, services.Groups)

		// 注册菜单管理路由
		registerMenuRoutes(v1, jwtUtil, services.Menu)

		// 注册文件管理路由
		registerFileRoutes(v1, jwtUtil, services.File)

		// 注册消息通知路由
		registerMessageRoutes(v1, jwtUtil, services.Message)

		// 注册WebSocket路由
		registerWebSocketRoutes(v1, jwtUtil, services.WSManager, logger)

		// 临时测试路由
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}
}

// registerAuthRoutes 注册认证路由
func registerAuthRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, ssoService sso.Service) {
	authHandler := handler.NewAuthHandler(ssoService)

	auth := rg.Group("/auth")
	{
		// 公开路由（不需要认证）
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		// 需要认证的路由
		authenticated := auth.Group("")
		authenticated.Use(middleware.AuthRequired(jwtUtil))
		{
			authenticated.POST("/logout", authHandler.Logout)
			authenticated.POST("/logout-all", authHandler.LogoutAll)
			authenticated.GET("/sessions", authHandler.GetActiveSessions)
			authenticated.POST("/kick-device", authHandler.KickDevice)
		}
	}
}

// registerMetadataRoutes 注册元数据路由
func registerMetadataRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, metadataService metadata.Service) {
	metadataHandler := handler.NewMetadataHandler(metadataService)

	metadata := rg.Group("/metadata")
	metadata.Use(middleware.AuthRequired(jwtUtil))
	{
		metadata.GET("/tables/:tableName", metadataHandler.GetTable)
		metadata.GET("/tables/:tableName/columns", metadataHandler.GetColumns)
		metadata.GET("/tables/:tableName/refs", metadataHandler.GetTableRefs)
		metadata.GET("/tables/:tableName/actions", metadataHandler.GetActions)
		metadata.POST("/refresh", metadataHandler.RefreshCache)
		metadata.GET("/version", metadataHandler.GetMetadataVersion)
	}
}

// registerDictRoutes 注册字典路由
func registerDictRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, dictService dict.Service) {
	dictHandler := handler.NewDictHandler(dictService)

	dicts := rg.Group("/dicts")
	dicts.Use(middleware.AuthRequired(jwtUtil))
	{
		dicts.GET("/:dictId/items", dictHandler.GetDictItems)
		dicts.GET("/name/:dictName/items", dictHandler.GetDictItemsByName)
		dicts.GET("/:dictId/default", dictHandler.GetDefaultValue)
		dicts.POST("/refresh", dictHandler.RefreshCache)
	}
}

// registerSequenceRoutes 注册序号路由
func registerSequenceRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, sequenceService sequence.Service) {
	sequenceHandler := handler.NewSequenceHandler(sequenceService)

	sequences := rg.Group("/sequences")
	sequences.Use(middleware.AuthRequired(jwtUtil))
	{
		sequences.POST("/:seqName/next", sequenceHandler.NextValue)
		sequences.POST("/batch", sequenceHandler.BatchNextValue)
		sequences.GET("/:seqName/current", sequenceHandler.GetCurrentValue)
		sequences.POST("/:seqName/reset", sequenceHandler.ResetSequence)
	}
}

// registerCRUDRoutes 注册通用CRUD路由
func registerCRUDRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, crudService crud.Service) {
	crudHandler := handler.NewCrudHandler(crudService)

	data := rg.Group("/data")
	data.Use(middleware.AuthRequired(jwtUtil))
	{
		data.GET("/:tableName/:id", crudHandler.GetOne)
		data.POST("/:tableName/query", crudHandler.GetList)
		data.POST("/:tableName", crudHandler.Create)
		data.PUT("/:tableName/:id", crudHandler.Update)
		data.DELETE("/:tableName/:id", crudHandler.Delete)
		data.POST("/:tableName/batch-delete", crudHandler.BatchDelete)
	}
}

// registerActionRoutes 注册动作路由
func registerActionRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, actionService action.Service) {
	actionHandler := handler.NewActionHandler(actionService)

	actions := rg.Group("/actions")
	actions.Use(middleware.AuthRequired(jwtUtil))
	{
		actions.GET("/:actionId", actionHandler.GetAction)
		actions.POST("/:actionId/execute", actionHandler.ExecuteAction)
		actions.POST("/:actionId/batch-execute", actionHandler.BatchExecuteAction)
		actions.POST("/by-name/:tableName/:actionName/execute", actionHandler.ExecuteActionByName)
	}
}

// registerWorkflowRoutes 注册工作流路由
func registerWorkflowRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, workflowService workflow.Service) {
	workflowHandler := handler.NewWorkflowHandler(workflowService)

	workflow := rg.Group("/workflow")
	workflow.Use(middleware.AuthRequired(jwtUtil))
	{
		// 流程定义管理
		definitions := workflow.Group("/definitions")
		{
			definitions.POST("", workflowHandler.CreateDefinition)
			definitions.GET("", workflowHandler.ListDefinitions)
			definitions.GET("/:id", workflowHandler.GetDefinition)
			definitions.PUT("/:id", workflowHandler.UpdateDefinition)
			definitions.POST("/:id/publish", workflowHandler.PublishDefinition)
		}

		// 流程节点管理
		nodes := workflow.Group("/nodes")
		{
			nodes.POST("", workflowHandler.CreateNode)
			nodes.GET("", workflowHandler.GetNodes)
			nodes.PUT("/:id", workflowHandler.UpdateNode)
			nodes.DELETE("/:id", workflowHandler.DeleteNode)
		}

		// 流程流转管理
		transitions := workflow.Group("/transitions")
		{
			transitions.POST("", workflowHandler.CreateTransition)
			transitions.GET("", workflowHandler.GetTransitions)
			transitions.DELETE("/:id", workflowHandler.DeleteTransition)
		}

		// 流程实例管理
		instances := workflow.Group("/instances")
		{
			instances.POST("/start", workflowHandler.StartProcess)
			instances.GET("", workflowHandler.ListInstances)
			instances.GET("/:id", workflowHandler.GetInstance)
			instances.POST("/:id/terminate", workflowHandler.TerminateInstance)
		}

		// 任务管理
		tasks := workflow.Group("/tasks")
		{
			tasks.GET("/my", workflowHandler.ListMyTasks)
			tasks.GET("/:id", workflowHandler.GetTask)
			tasks.POST("/complete", workflowHandler.CompleteTask)
			tasks.POST("/:id/claim", workflowHandler.ClaimTask)
			tasks.POST("/:id/transfer", workflowHandler.TransferTask)
		}
	}
}

// registerAuditRoutes 注册审计日志路由
func registerAuditRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, auditService audit.Service) {
	auditHandler := handler.NewAuditHandler(auditService)

	audit := rg.Group("/audit")
	audit.Use(middleware.AuthRequired(jwtUtil))
	{
		audit.GET("/logs", auditHandler.QueryLogs)
		audit.GET("/logs/:id", auditHandler.GetLog)
		audit.GET("/users/:userId/logs", auditHandler.GetUserLogs)
		audit.GET("/resources/:resource/:resourceId/logs", auditHandler.GetResourceLogs)
		audit.GET("/statistics", auditHandler.GetStatistics)
		audit.POST("/clean", auditHandler.CleanExpiredLogs)
	}
}

// registerGroupsRoutes 注册权限组管理路由
func registerGroupsRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, groupService groups.Service) {
	groupHandler := handler.NewGroupsHandler(groupService)

	groupsRg := rg.Group("/groups")
	groupsRg.Use(middleware.AuthRequired(jwtUtil))
	{
		groupsRg.POST("", groupHandler.CreateGroup)
		groupsRg.GET("", groupHandler.ListGroups)
		groupsRg.GET("/:id", groupHandler.GetGroup)
		groupsRg.PUT("/:id", groupHandler.UpdateGroup)
		groupsRg.DELETE("/:id", groupHandler.DeleteGroup)
		groupsRg.POST("/:id/permissions", groupHandler.AssignPermissions)
		groupsRg.GET("/:id/permissions", groupHandler.GetGroupPermissions)
		groupsRg.POST("/users/:userId", groupHandler.AssignGroupsToUser)
		groupsRg.GET("/users/:userId", groupHandler.GetUserGroups)
	}

	// 权限检查接口
	perms := rg.Group("/permissions")
	perms.Use(middleware.AuthRequired(jwtUtil))
	{
		perms.POST("/check", groupHandler.CheckPermission)
		perms.GET("/user", groupHandler.GetUserPermission)
	}
}

// registerDirectoryRoutes 注册安全目录管理路由
func registerDirectoryRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, groupService groups.Service) {
	dirHandler := handler.NewDirectoryHandler(groupService)

	dirs := rg.Group("/directories")
	dirs.Use(middleware.AuthRequired(jwtUtil))
	{
		dirs.POST("", dirHandler.CreateDirectory)
		dirs.GET("", dirHandler.ListDirectories)
		dirs.GET("/tree", dirHandler.GetDirectoryTree)
		dirs.GET("/:id", dirHandler.GetDirectory)
		dirs.PUT("/:id", dirHandler.UpdateDirectory)
		dirs.DELETE("/:id", dirHandler.DeleteDirectory)
	}
}

// registerMenuRoutes 注册菜单管理路由
func registerMenuRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, menuService menu.Service) {
	menuHandler := handler.NewMenuHandler(menuService)

	menus := rg.Group("/menus")
	menus.Use(middleware.AuthRequired(jwtUtil))
	{
		menus.POST("", menuHandler.CreateMenu)
		menus.GET("", menuHandler.ListMenus)
		menus.GET("/tree", menuHandler.GetMenuTree)
		menus.GET("/user/tree", menuHandler.GetUserMenuTree)
		menus.GET("/user/routers", menuHandler.GetUserRouters)
		menus.GET("/:id", menuHandler.GetMenu)
		menus.PUT("/:id", menuHandler.UpdateMenu)
		menus.DELETE("/:id", menuHandler.DeleteMenu)
	}
}

// registerFileRoutes 注册文件管理路由
func registerFileRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, fileService file.Service) {
	fileHandler := handler.NewFileHandler(fileService)

	files := rg.Group("/files")
	files.Use(middleware.AuthRequired(jwtUtil))
	{
		// 上传接口
		files.POST("/upload", fileHandler.UploadFile)
		files.POST("/upload/multiple", fileHandler.UploadMultipleFiles)

		// 下载和预览接口
		files.GET("/download/:id", fileHandler.DownloadFile)
		files.GET("/preview/:id", fileHandler.PreviewFile)

		// 文件管理接口
		files.GET("/:id", fileHandler.GetFile)
		files.POST("/list", fileHandler.ListFiles)
		files.DELETE("/:id", fileHandler.DeleteFile)

		// 直接访问（通过存储名称）
		files.GET("/access/:storageName", fileHandler.GetFileByPath)
	}
}

// registerMessageRoutes 注册消息通知路由
func registerMessageRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, messageService message.Service) {
	messageHandler := handler.NewMessageHandler(messageService)

	messages := rg.Group("/messages")
	messages.Use(middleware.AuthRequired(jwtUtil))
	{
		// 消息发送
		messages.POST("/send", messageHandler.SendMessage)
		messages.POST("/send/template", messageHandler.SendTemplateMessage)
		messages.POST("/send/batch", messageHandler.SendBatchMessage)
		messages.POST("/send/all", messageHandler.SendToAll)

		// 消息查询
		messages.GET("/:id", messageHandler.GetMessage)
		messages.POST("/list", messageHandler.ListUserMessages)
		messages.GET("/unread/count", messageHandler.GetUnreadCount)
		messages.GET("/unread/list", messageHandler.GetUnreadMessages)

		// 消息操作
		messages.POST("/:id/read", messageHandler.MarkAsRead)
		messages.POST("/read-all", messageHandler.MarkAllAsRead)
		messages.POST("/:id/star", messageHandler.StarMessage)
		messages.POST("/:id/archive", messageHandler.ArchiveMessage)
		messages.DELETE("/:id", messageHandler.DeleteMessage)
	}
}

// registerWebSocketRoutes 注册WebSocket路由
func registerWebSocketRoutes(rg *gin.RouterGroup, jwtUtil *jwt.JWT, wsManager *ws.Manager, logger *zap.Logger) {
	wsHandler := handler.NewWebSocketHandler(wsManager, logger)

	wsGroup := rg.Group("/ws")
	{
		// WebSocket连接（需要认证）
		wsGroup.GET("/messages", middleware.AuthRequired(jwtUtil), wsHandler.HandleConnection)

		// WebSocket管理接口（需要认证）
		wsGroup.GET("/online/users", middleware.AuthRequired(jwtUtil), wsHandler.GetOnlineUsers)
		wsGroup.GET("/online/check", middleware.AuthRequired(jwtUtil), wsHandler.CheckUserOnline)
		wsGroup.POST("/broadcast", middleware.AuthRequired(jwtUtil), wsHandler.BroadcastMessage)
	}
}
