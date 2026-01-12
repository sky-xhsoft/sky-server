package hotload

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"go.uber.org/zap"
)

// Loader 插件加载器
// 负责动态加载编译后的 .so 插件文件
type Loader struct {
	manager       *core.Manager        // 插件管理器
	loadedPlugins map[string]*plugin.Plugin // 已加载的插件
	mu            sync.RWMutex         // 互斥锁
}

// LoadResult 加载结果
type LoadResult struct {
	Success bool   // 是否成功
	Error   error  // 错误信息
}

// NewLoader 创建加载器
func NewLoader(manager *core.Manager) *Loader {
	return &Loader{
		manager:       manager,
		loadedPlugins: make(map[string]*plugin.Plugin),
	}
}

// Load 加载插件
// pluginName: 插件名称
// soPath: 编译后的 .so 文件路径
// hookPoint: 钩子点
// priority: 优先级
func (l *Loader) Load(pluginName string, soPath string, hookPoint string, priority int) *LoadResult {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := &LoadResult{}

	logger.Info("开始加载插件",
		zap.String("plugin", pluginName),
		zap.String("soPath", soPath),
		zap.String("hookPoint", hookPoint))

	// 1. 检查是否已加载
	if _, exists := l.loadedPlugins[pluginName]; exists {
		logger.Warn("插件已加载，将先卸载旧版本",
			zap.String("plugin", pluginName))
		// Note: Go plugin 无法真正卸载，只能覆盖注册
		delete(l.loadedPlugins, pluginName)
	}

	// 2. 打开插件文件
	p, err := plugin.Open(soPath)
	if err != nil {
		result.Error = fmt.Errorf("打开插件文件失败: %w", err)
		logger.Error("打开插件文件失败",
			zap.String("plugin", pluginName),
			zap.String("soPath", soPath),
			zap.Error(err))
		return result
	}

	// 3. 查找 Register 符号
	symbol, err := p.Lookup("Register")
	if err != nil {
		result.Error = fmt.Errorf("查找 Register 函数失败: %w (插件必须导出 Register 函数)", err)
		logger.Error("查找 Register 函数失败",
			zap.String("plugin", pluginName),
			zap.Error(err))
		return result
	}

	// 4. 类型断言为注册函数
	registerFunc, ok := symbol.(func())
	if !ok {
		result.Error = fmt.Errorf("Register 符号类型错误，期望 func()，实际 %T", symbol)
		logger.Error("Register 符号类型错误",
			zap.String("plugin", pluginName))
		return result
	}

	// 5. 调用注册函数
	// 注册函数应该调用 registry.Register() 来注册插件
	logger.Info("调用插件注册函数",
		zap.String("plugin", pluginName))

	registerFunc()

	// 6. 保存加载的插件引用
	l.loadedPlugins[pluginName] = p

	result.Success = true

	logger.Info("插件加载成功",
		zap.String("plugin", pluginName),
		zap.String("hookPoint", hookPoint))

	return result
}

// Unload 卸载插件
// 注意：Go plugin 包不支持真正的卸载，只能从管理器中移除
func (l *Loader) Unload(pluginName string, hookPoint string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	logger.Info("卸载插件",
		zap.String("plugin", pluginName),
		zap.String("hookPoint", hookPoint))

	// 从已加载列表中移除
	delete(l.loadedPlugins, pluginName)

	// 从插件管理器中移除
	// 注意：这只是从注册表中移除，不会释放内存
	if err := l.manager.DisablePlugin(hookPoint, pluginName); err != nil {
		logger.Error("从管理器中禁用插件失败",
			zap.String("plugin", pluginName),
			zap.Error(err))
		return err
	}

	logger.Info("插件已卸载",
		zap.String("plugin", pluginName))

	return nil
}

// Reload 重新加载插件
func (l *Loader) Reload(pluginName string, soPath string, hookPoint string, priority int) *LoadResult {
	logger.Info("重新加载插件",
		zap.String("plugin", pluginName),
		zap.String("hookPoint", hookPoint))

	// 1. 卸载旧版本
	if err := l.Unload(pluginName, hookPoint); err != nil {
		logger.Warn("卸载旧插件失败，将继续加载新版本",
			zap.String("plugin", pluginName),
			zap.Error(err))
	}

	// 2. 加载新版本
	return l.Load(pluginName, soPath, hookPoint, priority)
}

// IsLoaded 检查插件是否已加载
func (l *Loader) IsLoaded(pluginName string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	_, exists := l.loadedPlugins[pluginName]
	return exists
}

// GetLoadedPlugins 获取所有已加载的插件名称
func (l *Loader) GetLoadedPlugins() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	plugins := make([]string, 0, len(l.loadedPlugins))
	for name := range l.loadedPlugins {
		plugins = append(plugins, name)
	}
	return plugins
}
