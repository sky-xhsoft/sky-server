//go:build wireinject
// +build wireinject

package wire

import (
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/dbmonitor"
	jwtPkg "github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/queue"
	ws "github.com/sky-xhsoft/sky-server/internal/pkg/websocket"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	mysqlRepo "github.com/sky-xhsoft/sky-server/internal/repository/mysql"
	redisRepo "github.com/sky-xhsoft/sky-server/internal/repository/redis"
	"github.com/sky-xhsoft/sky-server/internal/service/action"
	"github.com/sky-xhsoft/sky-server/internal/service/audit"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
	"github.com/sky-xhsoft/sky-server/internal/service/crud"
	"github.com/sky-xhsoft/sky-server/internal/service/dict"
	emailSvc "github.com/sky-xhsoft/sky-server/internal/service/email"
	"github.com/sky-xhsoft/sky-server/internal/service/file"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/idgen"
	"github.com/sky-xhsoft/sky-server/internal/service/language"
	"github.com/sky-xhsoft/sky-server/internal/service/menu"
	"github.com/sky-xhsoft/sky-server/internal/service/message"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/internal/service/sequence"
	"github.com/sky-xhsoft/sky-server/internal/service/sso"
	"github.com/sky-xhsoft/sky-server/internal/service/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用核心依赖容器
type App struct {
	Logger      *zap.Logger
	DB          *gorm.DB
	RedisClient *redis.Client
	JWTUtil     *jwtPkg.JWT
	DBMonitor   *dbmonitor.Monitor
	TaskQueue   queue.Queue
	WSManager   *ws.Manager

	// 仓储
	UserRepo    repository.UserRepository
	MetadataRepo repository.MetadataRepository
	DictRepo    repository.DictRepository
	SeqRepo     repository.SequenceRepository

	// 核心服务
	SSOService      sso.Service
	MetadataService metadata.Service
	DictService     dict.Service
	SequenceService sequence.Service
	GroupsService   groups.Service
	IDGenService    idgen.Service
	CRUDService     crud.Service
	ActionService   action.Service
	WorkflowService workflow.Service
	AuditService    audit.Service
	MenuService     menu.Service
	FileService     file.Service
	MessageService  message.Service
	EmailService    emailSvc.Service
	CompanyService  company.Service
	LanguageService language.Service
}

// --- 基础设施 ---

func provideConfig() (*config.Config, error)      { return config.Load() }
func provideLogger(cfg *config.Config) (*zap.Logger, error) { return logger.Init(&cfg.Log) }
func provideDB(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	return mysqlRepo.Init(&cfg.Database.MySQL, log)
}
func provideRedis(cfg *config.Config) (*redis.Client, error) { return redisRepo.Init(&cfg.Redis) }
func provideJWT(cfg *config.Config, redisClient *redis.Client) *jwtPkg.JWT {
	return jwtPkg.New(cfg.JWT.Secret, jwtPkg.WithRedis(redisClient, "jwt_blacklist:"))
}
func provideDBMonitor(db *gorm.DB) *dbmonitor.Monitor {
	return dbmonitor.NewMonitor(db, 30*time.Second)
}
func provideWSManager(log *zap.Logger) *ws.Manager { return ws.NewManager(log) }
func provideTaskQueue(redisClient *redis.Client, log *zap.Logger) queue.Queue {
	return queue.NewRedisQueue(redisClient, queue.DefaultConfig(), log)
}

// --- 仓储 ---

func provideUserRepo(db *gorm.DB) repository.UserRepository               { return mysqlRepo.NewUserRepository(db) }
func provideMetadataRepo(db *gorm.DB) repository.MetadataRepository        { return mysqlRepo.NewMetadataRepository(db) }
func provideDictRepo(db *gorm.DB) repository.DictRepository                { return mysqlRepo.NewDictRepository(db) }
func provideSeqRepo(db *gorm.DB) repository.SequenceRepository             { return mysqlRepo.NewSequenceRepository(db) }
func provideCloudStorageConfigRepo(db *gorm.DB) repository.CloudStorageConfigRepository { return mysqlRepo.NewCloudStorageConfigRepository(db) }
func provideSysCompanyConfRepo(db *gorm.DB) repository.SysCompanyConfRepository           { return mysqlRepo.NewSysCompanyConfRepository(db) }
func provideLangRepo(db *gorm.DB) repository.LanguageRepository              { return mysqlRepo.NewLanguageRepository(db) }
func provideTransRepo(db *gorm.DB) repository.TranslationRepository          { return mysqlRepo.NewTranslationRepository(db) }

// --- 服务 ---

func provideSSOService(userRepo repository.UserRepository, jwtUtil *jwtPkg.JWT, cfg *config.Config) sso.Service {
	return sso.NewService(userRepo, jwtUtil, cfg.JWT.AccessTokenExpire, cfg.JWT.RefreshTokenExpire)
}
func provideMetadataService(metadataRepo repository.MetadataRepository, userRepo repository.UserRepository, redisClient *redis.Client, cfg *config.Config) metadata.Service {
	return metadata.NewService(metadataRepo, userRepo, redisClient, cfg.Cache.MetadataTTL)
}
func provideDictService(dictRepo repository.DictRepository, redisClient *redis.Client, cfg *config.Config) dict.Service {
	return dict.NewService(dictRepo, redisClient, cfg.Cache.DictTTL)
}
func provideSequenceService(seqRepo repository.SequenceRepository, redisClient *redis.Client) sequence.Service {
	return sequence.NewService(seqRepo, redisClient)
}
func provideGroupsService(db *gorm.DB) groups.Service { return groups.NewService(db) }
func provideIDGenService(db *gorm.DB, redisClient *redis.Client) idgen.Service {
	return idgen.NewService(db, redisClient)
}
func provideCRUDService(db *gorm.DB, metadataService metadata.Service, groupsService groups.Service, dictService dict.Service,
	seqService sequence.Service, metadataRepo repository.MetadataRepository, userRepo repository.UserRepository, idgenService idgen.Service) crud.Service {
	return crud.NewService(db, metadataService, groupsService, dictService, seqService, metadataRepo, userRepo, idgenService)
}
func provideActionService(db *gorm.DB, metadataService metadata.Service, groupsService groups.Service, cfg *config.Config) action.Service {
	return action.NewService(db, metadataService, groupsService, cfg.Action.ScriptTimeout)
}
func provideWorkflowService(db *gorm.DB, actionService action.Service) workflow.Service {
	return workflow.NewService(db, actionService)
}
func provideAuditService(db *gorm.DB) audit.Service                 { return audit.NewService(db) }
func provideMenuService(db *gorm.DB) menu.Service                   { return menu.NewService(db) }
func provideFileService(db *gorm.DB, cfg *config.Config) file.Service {
	return file.NewService(db, &file.Config{UploadDir: cfg.File.UploadDir, MaxFileSize: cfg.File.MaxFileSize, AllowedExts: cfg.File.AllowedExts})
}
func provideMessageService(db *gorm.DB, wsManager *ws.Manager) message.Service {
	return message.NewService(db, wsManager)
}
func provideEmailService(cfg *config.Config, taskQueue queue.Queue) emailSvc.Service {
	return emailSvc.NewServiceWithQueue(&cfg.Email, taskQueue)
}
func provideCompanyService(db *gorm.DB) company.Service { return company.NewService(db) }
func provideLanguageService(langRepo repository.LanguageRepository, transRepo repository.TranslationRepository,
	redisClient *redis.Client, log *zap.Logger, cfg *config.Config) language.Service {
	return language.NewService(langRepo, transRepo, redisClient, log, cfg.Cache.LanguageTTL)
}

// InitializeApp 通过 Wire 编译时依赖注入初始化应用核心组件
func InitializeApp() (*App, error) {
	wire.Build(
		provideConfig, provideLogger, provideDB, provideRedis, provideJWT, provideDBMonitor,
		provideWSManager, provideTaskQueue,
		provideUserRepo, provideMetadataRepo, provideDictRepo, provideSeqRepo,
		provideLangRepo, provideTransRepo,
		provideSSOService, provideMetadataService, provideDictService, provideSequenceService,
		provideGroupsService, provideIDGenService, provideCRUDService, provideActionService,
		provideWorkflowService, provideAuditService, provideMenuService, provideFileService,
		provideMessageService, provideEmailService, provideCompanyService, provideLanguageService,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
