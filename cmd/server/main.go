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
	_ "github.com/sky-xhsoft/sky-server/api/swagger"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
	"github.com/sky-xhsoft/sky-server/internal/pkg/volcengine"
	"github.com/sky-xhsoft/sky-server/internal/repository/mysql"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
	"github.com/sky-xhsoft/sky-server/internal/service/dashboard"
	"github.com/sky-xhsoft/sky-server/internal/service/housekeeping"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
	"github.com/sky-xhsoft/sky-server/internal/service/pref"
	"github.com/sky-xhsoft/sky-server/internal/service/schedule"
	"github.com/sky-xhsoft/sky-server/internal/service/sms"
	"github.com/sky-xhsoft/sky-server/plugins"
	"github.com/sky-xhsoft/sky-server/wire"
	"go.uber.org/zap"
)

// @title Sky-Server API
// @version 1.0.2
// @description 元数据驱动的企业级应用框架 - 完整 REST API 文档
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9090
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 通过 Wire 初始化所有核心依赖
	app, err := wire.InitializeApp()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	defer app.Logger.Sync()
	defer mysql.CloseDB(app.DB)
	defer app.DBMonitor.Stop()

	logger.Info("Starting Sky-Server",
		zap.String("version", "1.0.2"),
		zap.Int("port", 9090),
	)

	// 2. 加载配置（用于可选组件）
	cfg, _ := config.Load()

	// 3. 启动数据库监控和消息队列
	app.DBMonitor.Start()
	logger.Info("DB monitor started")

	ctx, cancelQueue := context.WithCancel(context.Background())
	defer cancelQueue()
	if err := app.TaskQueue.Start(ctx); err != nil {
		logger.Warn("Failed to start task queue", zap.Error(err))
	} else {
		defer app.TaskQueue.Stop()
	}

	// 4. 启动WebSocket管理器
	go app.WSManager.Run()
	logger.Info("WebSocket manager started")

	// 5. 初始化插件管理器
	_ = plugins.Setup(app.DB, cfg)
	logger.Info("Plugins initialized")

	// 6. 初始化云盘和直播服务（条件依赖）
	cloudStorageConfigRepo := mysql.NewCloudStorageConfigRepository(app.DB)
	sysCompanyConfRepo := mysql.NewSysCompanyConfRepository(app.DB)

	companyStorageManager, err := storage.NewCompanyStorageManager(&cfg.Storage, sysCompanyConfRepo, 0)
	if err != nil {
		logger.Fatal("Failed to initialize company storage manager", zap.Error(err))
	}

	storageConfigService := cloud.NewStorageConfigService(app.DB, cloudStorageConfigRepo, companyStorageManager)
	cloudService := cloud.NewService(app.DB, companyStorageManager, app.UserRepo)

	multipartService := cloud.NewMultipartUploadService(
		app.DB, companyStorageManager, app.UserRepo, cloudService,
		cfg.MultipartUpload.ChunkSize, cfg.MultipartUpload.SessionExpireHours,
	)
	logger.Info("Multipart upload service initialized")

	companyTencentManager, err := tencent.NewCompanyTencentManager(&cfg.TencentCloud, sysCompanyConfRepo, 0)
	if err != nil {
		logger.Warn("Failed to initialize company tencent manager", zap.Error(err))
	}

	companyVolcengineManager, err := volcengine.NewCompanyVolcengineManager(&cfg.Volcengine, sysCompanyConfRepo, 0)
	if err != nil {
		logger.Warn("Failed to initialize company volcengine manager", zap.Error(err))
	}

	// 公司配置更新后，同步刷新腾讯云/火山引擎客户端缓存，使新配置的密钥即时生效
	if setter, ok := app.CompanyService.(company.ConfigChangedSetter); ok {
		setter.SetConfigChangedCallback(func(companyID uint) {
			if companyTencentManager != nil {
				if cerr := companyTencentManager.RefreshCompanyConfig(context.Background(), companyID); cerr != nil {
					logger.Warn("Refresh tencent config cache failed", zap.Uint("companyID", companyID), zap.Error(cerr))
				}
			}
			if companyVolcengineManager != nil {
				if cerr := companyVolcengineManager.RefreshCompanyConfig(context.Background(), companyID); cerr != nil {
					logger.Warn("Refresh volcengine config cache failed", zap.Uint("companyID", companyID), zap.Error(cerr))
				}
			}
		})
	}

	var liveService live.Service
	if companyTencentManager != nil {
		liveService, err = live.NewServiceWithCompanyManager(companyTencentManager, app.RedisClient)
	} else {
		liveService, err = live.NewService(&cfg.TencentCloud, app.RedisClient)
	}
	if err != nil {
		logger.Warn("Failed to initialize live service", zap.Error(err))
	} else {
		logger.Info("Live service initialized successfully")
	}

	liveRoomService := live.NewLiveRoomService(app.DB, cloudService)
	dashboardService := dashboard.NewService(app.DB)
	pullStreamTaskService := live.NewPullStreamTaskService(app.DB)
	var liveTemplateService live.TemplateService
	if liveService != nil {
		liveTemplateService = live.NewTemplateService(app.DB, liveService, cfg.TencentCloud.Live.CallbackURL)
	}
	clipService := live.NewClipService(app.DB)
	var speechTaskService live.SpeechTaskService
	if companyVolcengineManager != nil {
		speechTaskService = live.NewSpeechTaskService(companyVolcengineManager, app.DB)
	}

	logger.Info("Services initialized")

	// 用户环境变量服务 + 自动任务调度器
	prefService := pref.NewService(app.DB, app.UserRepo)
	taskRunner := schedule.NewTaskRunner()
	schedSvc := schedule.NewScheduler(app.DB, taskRunner)

	// 注入动作变量替换器(动作内容/参数支持 $变量$)
	app.ActionService.SetVarReplacer(func(ctx context.Context, str string, userID uint) (string, error) {
		return prefService.ReplaceVariables(ctx, str, userID)
	})
	// 注入自动任务执行器(供 list_button 动作「执行全部任务」调用)
	app.ActionService.SetScheduleRunner(func() (string, error) {
		return schedule.RunAllEnabled(context.Background(), app.DB, taskRunner)
	})

	// 启动自动任务调度器(后台)
	schedSvc.Start()
	logger.Info("Auto-task scheduler started")

	// 7. 初始化Gin引擎
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.MaxMultipartMemory = 32 << 20

	// 8. 注册路由
	services := &router.Services{
		SSO:             app.SSOService,
		Metadata:        app.MetadataService,
		Dict:            app.DictService,
		Sequence:        app.SequenceService,
		CRUD:            app.CRUDService,
		Dashboard:       dashboardService,
		Action:          app.ActionService,
		Workflow:        app.WorkflowService,
		Audit:           app.AuditService,
		Groups:          app.GroupsService,
		Menu:            app.MenuService,
		File:            app.FileService,
		Message:         app.MessageService,
		Email:           app.EmailService,
		SMS:             sms.NewService(cfg.SMS, app.RedisClient),
		Company:         app.CompanyService,
		Cloud:           cloudService,
		MultipartUpload: multipartService,
		StorageConfig:   storageConfigService,
		Live:            liveService,
		LiveRoom:        liveRoomService,
		PullStreamTask:  pullStreamTaskService,
		LiveTemplate:    liveTemplateService,
		SpeechTask:      speechTaskService,
		ClipService:     clipService,
		WSManager:       app.WSManager,
		Language:        app.LanguageService,
		Pref:            prefService,
		Schedule:        schedSvc,
	}

	repos := &router.Repositories{
		DictRepo: app.DictRepo,
	}

	router.Setup(engine, cfg, app.JWTUtil, services, repos, app.Logger, app.DB)
	logger.Info("Routes registered successfully")

	// 9. 启动定时清理任务
	go func() {
		interval := cfg.MultipartUpload.CleanupInterval
		if interval <= 0 {
			interval = 3600
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		logger.Info("Multipart cleanup task started", zap.Int("intervalSeconds", interval))
		for range ticker.C {
			cleanCtx := context.Background()
			if err := multipartService.CleanupExpiredSessions(cleanCtx); err != nil {
				logger.Error("Cleanup expired sessions failed", zap.Error(err))
			} else {
				logger.Info("Expired sessions cleaned")
			}
		}
	}()

	// 启动自检：弱配置告警、元数据与表结构一致性（后台执行，不阻塞启动）
	go runStartupChecks(cfg, app.DB)

	// 过期会话、超过保留期的审计日志、已过期分享状态的定时清理
	housekeeping.Start(ctx, app.DB, housekeeping.Config{
		Interval:           time.Duration(cfg.Housekeeping.IntervalSeconds) * time.Second,
		AuditRetentionDays: cfg.Housekeeping.AuditRetentionDays,
	})

	// 10. 启动HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: engine,
	}

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
