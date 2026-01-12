package hotload

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"go.uber.org/zap"
)

// HotloadManager 热加载管理器
// 自动扫描 runtime 目录，编译、加载所有插件
type HotloadManager struct {
	pluginManager *core.Manager
	compiler      *Compiler
	loader        *Loader
	watcher       *Watcher

	runtimeDir string // 插件源码目录

	mu      sync.RWMutex
	started bool
}

// Config 热加载管理器配置
type Config struct {
	RuntimeDir   string        // 插件源码目录
	CompiledDir  string        // 编译输出目录
	ModulePath   string        // Go 模块路径
	DebounceTime time.Duration // 防抖动时间
	EnableWatch  bool          // 是否启用文件监听
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		RuntimeDir:   "plugins/runtime",
		CompiledDir:  "plugins/compiled",
		ModulePath:   "github.com/sky-xhsoft/sky-server",
		DebounceTime: 2 * time.Second,
		EnableWatch:  true,
	}
}

// NewHotloadManager 创建热加载管理器
func NewHotloadManager(pluginManager *core.Manager, config *Config) (*HotloadManager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建编译器
	compiler := NewCompiler(config.RuntimeDir, config.CompiledDir, config.ModulePath)

	// 创建加载器
	loader := NewLoader(pluginManager)

	manager := &HotloadManager{
		pluginManager: pluginManager,
		compiler:      compiler,
		loader:        loader,
		runtimeDir:    config.RuntimeDir,
		started:       false,
	}

	// 创建文件监听器（如果启用）
	if config.EnableWatch {
		watcher, err := NewWatcher(config.RuntimeDir, config.DebounceTime, manager.handleFileChange)
		if err != nil {
			return nil, fmt.Errorf("创建文件监听器失败: %w", err)
		}
		manager.watcher = watcher
	}

	return manager, nil
}

// Start 启动热加载管理器
// 1. 扫描 runtime 目录发现所有插件
// 2. 编译并加载所有插件
// 3. 启动文件监听器
func (m *HotloadManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("热加载管理器已启动")
	}

	// 检查平台支持
	if runtime.GOOS == "windows" {
		return fmt.Errorf("hot reload not supported on this platform")
	}

	logger.Info("启动热加载管理器...")

	// 1. 扫描并加载所有插件
	if err := m.scanAndLoadAll(); err != nil {
		logger.Error("扫描加载插件失败", zap.Error(err))
		// 不阻止管理器启动
	}

	// 2. 启动文件监听器（如果配置了）
	if m.watcher != nil {
		if err := m.watcher.Start(); err != nil {
			logger.Error("启动文件监听器失败", zap.Error(err))
			// 不阻止管理器启动
		}
	}

	m.started = true

	logger.Info("✨ 热加载管理器启动完成")

	return nil
}

// Stop 停止热加载管理器
func (m *HotloadManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	logger.Info("停止热加载管理器...")

	// 停止文件监听器
	if m.watcher != nil {
		if err := m.watcher.Stop(); err != nil {
			logger.Error("停止文件监听器失败", zap.Error(err))
		}
	}

	m.started = false

	logger.Info("热加载管理器已停止")

	return nil
}

// scanAndLoadAll 扫描并加载所有插件
func (m *HotloadManager) scanAndLoadAll() error {
	logger.Info("扫描插件目录...", zap.String("dir", m.runtimeDir))

	// 获取 runtime 目录的绝对路径
	absRuntimeDir, err := filepath.Abs(m.runtimeDir)
	if err != nil {
		return fmt.Errorf("获取 runtime 目录绝对路径失败: %w", err)
	}

	// 检查目录是否存在
	if _, err := os.Stat(absRuntimeDir); os.IsNotExist(err) {
		logger.Warn("插件目录不存在，跳过加载", zap.String("dir", absRuntimeDir))
		return nil
	}

	// 读取目录
	entries, err := os.ReadDir(absRuntimeDir)
	if err != nil {
		return fmt.Errorf("读取插件目录失败: %w", err)
	}

	pluginCount := 0
	for _, entry := range entries {
		// 只处理目录
		if !entry.IsDir() {
			continue
		}

		pluginName := entry.Name()

		// 跳过隐藏目录和特殊目录
		if strings.HasPrefix(pluginName, ".") || strings.HasPrefix(pluginName, "_") {
			logger.Debug("跳过特殊目录", zap.String("dir", pluginName))
			continue
		}

		// 加载插件
		logger.Info("发现插件", zap.String("plugin", pluginName))
		if err := m.loadPlugin(pluginName); err != nil {
			logger.Error("加载插件失败",
				zap.String("plugin", pluginName),
				zap.Error(err))
			// 继续加载其他插件
			continue
		}

		pluginCount++
	}

	logger.Info("插件扫描完成",
		zap.Int("total", pluginCount),
		zap.Int("loaded", pluginCount))

	return nil
}

// loadPlugin 加载单个插件
func (m *HotloadManager) loadPlugin(pluginPath string) error {
	logger.Info("加载插件",
		zap.String("plugin", pluginPath))

	// 1. 编译插件
	result := m.compiler.Compile(pluginPath)
	if !result.Success {
		return result.Error
	}

	logger.Info("插件编译成功",
		zap.String("plugin", pluginPath),
		zap.Duration("duration", result.CompileDuration))

	// 2. 加载插件
	// 注意：HookPoint 和 Priority 由插件在 Register() 中指定
	// 这里我们只需要加载，不需要知道钩子点
	loadResult := m.loader.Load(
		pluginPath,    // 插件名称
		result.OutputPath,
		"",  // hookPoint 由插件自己注册时指定
		0)   // priority 由插件自己注册时指定

	if !loadResult.Success {
		return loadResult.Error
	}

	logger.Info("插件加载成功", zap.String("plugin", pluginPath))

	return nil
}

// handleFileChange 处理文件变化事件
func (m *HotloadManager) handleFileChange(event FileChangeEvent) {
	logger.Info("检测到插件文件变化",
		zap.String("plugin", event.PluginPath),
		zap.String("event", event.EventType))

	// 重新加载插件
	if err := m.loadPlugin(event.PluginPath); err != nil {
		logger.Error("重新加载插件失败",
			zap.String("plugin", event.PluginPath),
			zap.Error(err))
	} else {
		logger.Info("插件热重载成功 ✨",
			zap.String("plugin", event.PluginPath))
	}
}

// GetLoadedPlugins 获取已加载的插件列表
func (m *HotloadManager) GetLoadedPlugins() []string {
	return m.loader.GetLoadedPlugins()
}

// GetStatus 获取热加载管理器状态
func (m *HotloadManager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := map[string]interface{}{
		"started":       m.started,
		"watcherActive": m.watcher != nil && m.watcher.IsRunning(),
		"loadedPlugins": m.loader.GetLoadedPlugins(),
	}

	return status
}
