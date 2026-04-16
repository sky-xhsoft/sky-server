package plugins

import (
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"github.com/sky-xhsoft/sky-server/plugins/hotload"
	"github.com/sky-xhsoft/sky-server/plugins/registry"

	// 导入内置插件包以触发 init() 注册
	_ "github.com/sky-xhsoft/sky-server/plugins/builtin"

	// 导入 hooks 包以触发 init() 自动注册所有 hooks
	"github.com/sky-xhsoft/sky-server/plugins/hooks"

	// 注意：自定义插件会通过 plugins_gen.go 自动导入
	// 如需添加新插件，请运行: make plugin-scan
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup 初始化插件系统
// 1. 创建插件管理器
// 2. 从注册中心加载所有已注册的插件（静态编译插件）
// 3. 启动热加载管理器（动态加载插件）
// 4. 注册 Go 钩子函数到执行器
func Setup(db *gorm.DB, cfg *config.Config) *core.Manager {
	logger.Info("初始化插件系统...")

	// 1. 创建插件管理器
	pluginManager := core.NewManager(db)

	// 2. 创建插件加载器（用于静态编译的插件）
	loader := registry.NewLoader(pluginManager)

	// 3. 加载所有已注册的插件（静态编译插件）
	if err := loader.LoadAll(); err != nil {
		logger.Error("加载静态插件失败", zap.Error(err))
	}

	// 4. 启动热加载管理器（动态加载插件）
	setupHotloadManager(db, pluginManager)

	// 5. 注册 Go 钩子函数
	registerGoHooks(pluginManager, cfg)

	// 6. 输出已加载的插件信息
	logLoadedPlugins(pluginManager)

	logger.Info("插件系统初始化完成")

	return pluginManager
}

// registerGoHooks 注册 Go 钩子函数到执行器
// 将插件系统与 GoFuncRegistry 集成
// 所有 hooks 通过 init() 自动注册，这里只需调用 RegisterAll
func registerGoHooks(manager *core.Manager, cfg *config.Config) {
	// 自动注册所有 hooks
	hooks.RegisterAll(manager, cfg)

	// 输出已注册的 hooks
	registeredHooks := hooks.GetRegisteredHooks()
	logger.Info("Go 钩子函数已自动注册到执行器",
		zap.Int("count", len(registeredHooks)),
		zap.Strings("hooks", registeredHooks))
}

// setupHotloadManager 设置热加载管理器
func setupHotloadManager(db *gorm.DB, pluginManager *core.Manager) {
	logger.Info("设置热加载管理器...")

	// 创建热加载管理器
	hotloadMgr, err := hotload.NewHotloadManager(
		pluginManager,
		hotload.DefaultConfig())

	if err != nil {
		logger.Error("创建热加载管理器失败", zap.Error(err))
		return
	}

	// 启动管理器（自动扫描 runtime 目录并加载所有插件）
	if err := hotloadMgr.Start(); err != nil {
		// 如果是平台不支持错误，给出友好提示
		if err.Error() == "hot reload not supported on this platform" {
			logger.Warn("⚠️  热加载功能不支持当前平台（仅支持 Linux/macOS）")
			logger.Warn("💡 建议使用 WSL2 或 Linux 虚拟机来使用热加载功能")
			logger.Info("✅ 静态插件功能仍然可用")
			return
		}
		logger.Error("启动热加载管理器失败", zap.Error(err))
		return
	}

	logger.Info("✨ 热加载管理器已启动，支持 JSP 风格的插件动态编译和热重载")
	logger.Info("💡 只需将插件放入 plugins/runtime/ 目录即可自动加载")
}

// logLoadedPlugins 输出已加载的插件信息
func logLoadedPlugins(manager *core.Manager) {
	allPlugins := manager.GetAllPlugins()
	totalCount := 0

	for hookPoint, plugins := range allPlugins {
		totalCount += len(plugins)
		logger.Info("已加载插件",
			zap.String("hookPoint", hookPoint),
			zap.Int("count", len(plugins)))

		for _, info := range plugins {
			logger.Debug("插件详情",
				zap.String("name", info.Plugin.Name()),
				zap.String("version", info.Plugin.Version()),
				zap.String("description", info.Plugin.Description()),
				zap.Bool("enabled", info.Metadata.Enabled),
				zap.Int("priority", info.Metadata.Priority))
		}
	}

	logger.Info("插件加载完成",
		zap.Int("totalPlugins", totalCount),
		zap.Int("hookPoints", len(allPlugins)))
}
