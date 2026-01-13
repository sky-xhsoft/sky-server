package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/router"
	_ "github.com/sky-xhsoft/sky-server/api/swagger" // Swagger docs
	"github.com/sky-xhsoft/sky-server/internal/config"
	jwtPkg "github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	ws "github.com/sky-xhsoft/sky-server/internal/pkg/websocket"
	"github.com/sky-xhsoft/sky-server/internal/repository/mysql"
	"github.com/sky-xhsoft/sky-server/internal/repository/redis"
	"github.com/sky-xhsoft/sky-server/internal/service/action"
	"github.com/sky-xhsoft/sky-server/internal/service/audit"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"github.com/sky-xhsoft/sky-server/internal/service/crud"
	"github.com/sky-xhsoft/sky-server/internal/service/dict"
	"github.com/sky-xhsoft/sky-server/internal/service/file"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/idgen"
	"github.com/sky-xhsoft/sky-server/internal/service/menu"
	"github.com/sky-xhsoft/sky-server/internal/service/message"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/internal/service/sequence"
	"github.com/sky-xhsoft/sky-server/internal/service/sso"
	"github.com/sky-xhsoft/sky-server/internal/service/workflow"
	"github.com/sky-xhsoft/sky-server/plugins"
	"go.uber.org/zap"
)

// @title Sky-Server API
// @version 1.0.0
// @description 元数据驱动的企业级应用框架
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9090
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log, err := logger.Init(&cfg.Log)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	logger.Info("Starting Sky-Server",
		zap.String("version", cfg.App.Version),
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.Port),
	)

	// 3. 初始化数据库连接
	db, err := mysql.Init(&cfg.Database.MySQL, log)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer mysql.Close()
	logger.Info("Database connected successfully")

	// 4. 初始化Redis连接
	redisClient, err := redis.Init(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer redis.Close()
	logger.Info("Redis connected successfully")

	// 5. 初始化仓储层
	userRepo := mysql.NewUserRepository(db)
	metadataRepo := mysql.NewMetadataRepository(db)
	dictRepo := mysql.NewDictRepository(db)
	seqRepo := mysql.NewSequenceRepository(db)
	logger.Info("Repositories initialized")

	// 6. 初始化JWT工具
	jwtUtil := jwtPkg.New(cfg.JWT.Secret)
	logger.Info("JWT utility initialized")

	// 7. 初始化服务层
	ssoService := sso.NewService(
		userRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpire,
		cfg.JWT.RefreshTokenExpire,
	)

	metadataService := metadata.NewService(
		metadataRepo,
		redisClient,
		cfg.Cache.MetadataTTL,
	)

	dictService := dict.NewService(
		dictRepo,
		redisClient,
		cfg.Cache.DictTTL,
	)

	seqService := sequence.NewService(
		seqRepo,
		redisClient,
	)

	// 初始化权限组服务（CRUD和Action服务依赖它）
	groupsService := groups.NewService(db)

	// 初始化ID生成服务（基于Redis缓存）
	idgenService := idgen.NewService(db, redisClient)

	// 初始化插件管理器并注册所有钩子函数
	_ = plugins.Setup(db)

	crudService := crud.NewService(
		db,
		metadataService,
		groupsService,
		metadataRepo,
		userRepo,
		idgenService,
	)

	actionService := action.NewService(
		db,
		metadataService,
		groupsService,
		cfg.Action.ScriptTimeout,
	)

	workflowService := workflow.NewService(
		db,
		actionService,
	)

	auditService := audit.NewService(db)

	// 初始化菜单服务
	menuService := menu.NewService(db)

	// 初始化文件服务
	logger.Info("Initializing file service",
		zap.String("uploadDir", cfg.File.UploadDir),
		zap.Int64("maxFileSize", cfg.File.MaxFileSize),
		zap.Float64("maxFileSizeGB", float64(cfg.File.MaxFileSize)/(1024*1024*1024)),
	)
	fileService := file.NewService(db, &file.Config{
		UploadDir:   cfg.File.UploadDir,
		MaxFileSize: cfg.File.MaxFileSize,
		AllowedExts: cfg.File.AllowedExts,
	})

	// 初始化WebSocket管理器
	wsManager := ws.NewManager(log)
	go wsManager.Run() // 在goroutine中运行WebSocket管理器
	logger.Info("WebSocket manager started")

	// 初始化消息服务
	messageService := message.NewService(db, wsManager)

	// 初始化云盘存储
	cloudStorage, err := storage.NewLocalStorage(&storage.LocalStorageConfig{
		BasePath: cfg.File.UploadDir + "/cloud", // 使用 uploads/cloud 作为云盘存储目录
		BaseURL:  fmt.Sprintf("http://localhost:%d/files/cloud", cfg.App.Port),
	})
	if err != nil {
		logger.Fatal("Failed to initialize cloud storage", zap.Error(err))
	}

	// 初始化云盘服务
	cloudService := cloud.NewService(db, cloudStorage)

	logger.Info("Services initialized")

	// 8. 初始化Gin引擎
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// 设置文件上传大小限制（32MB内存缓存，超过的部分会写入临时文件）
	// 这样可以支持大文件上传而不会占用过多内存
	engine.MaxMultipartMemory = 32 << 20 // 32 MB

	// 9. 注册路由
	services := &router.Services{
		SSO:       ssoService,
		Metadata:  metadataService,
		Dict:      dictService,
		Sequence:  seqService,
		CRUD:      crudService,
		Action:    actionService,
		Workflow:  workflowService,
		Audit:     auditService,
		Groups:    groupsService,
		Menu:      menuService,
		File:      fileService,
		Message:   messageService,
		Cloud:     cloudService,
		WSManager: wsManager,
	}
	router.Setup(engine, cfg, jwtUtil, services, log)
	logger.Info("Routes registered successfully")

	// 10. 启动HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: engine,
	}

	// 在goroutine中启动服务器
	go func() {
		logger.Info("Server started", zap.Int("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// 11. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
