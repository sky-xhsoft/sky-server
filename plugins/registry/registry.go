package registry

import (
	"fmt"
	"sync"

	"github.com/sky-xhsoft/sky-server/plugins/core"
)

// PluginFactory 插件工厂函数类型
// 用于创建插件实例
type PluginFactory func() core.Plugin

// Registry 全局插件注册中心
// 插件通过 init() 函数自动注册到这里
type Registry struct {
	factories map[string]*FactoryInfo
	mu        sync.RWMutex
}

// FactoryInfo 工厂信息
type FactoryInfo struct {
	Factory  PluginFactory
	Metadata core.PluginMetadata
}

var (
	// globalRegistry 全局注册中心实例
	globalRegistry = &Registry{
		factories: make(map[string]*FactoryInfo),
	}
)

// Register 注册插件工厂到全局注册中心
// 插件应该在 init() 函数中调用此方法
func Register(name string, factory PluginFactory, metadata core.PluginMetadata) {
	globalRegistry.register(name, factory, metadata)
}

// register 内部注册方法
func (r *Registry) register(name string, factory PluginFactory, metadata core.PluginMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		panic(fmt.Sprintf("插件 %s 已注册", name))
	}

	// 填充元数据默认值
	if metadata.Name == "" {
		metadata.Name = name
	}
	if !metadata.Enabled {
		metadata.Enabled = true // 默认启用
	}

	r.factories[name] = &FactoryInfo{
		Factory:  factory,
		Metadata: metadata,
	}
}

// GetFactory 获取插件工厂
func (r *Registry) GetFactory(name string) (PluginFactory, core.PluginMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.factories[name]
	if !exists {
		return nil, core.PluginMetadata{}, fmt.Errorf("插件 %s 未注册", name)
	}

	return info.Factory, info.Metadata, nil
}

// ListPlugins 列出所有已注册的插件
func (r *Registry) ListPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// GetAllFactories 获取所有插件工厂
func (r *Registry) GetAllFactories() map[string]*FactoryInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*FactoryInfo, len(r.factories))
	for name, info := range r.factories {
		result[name] = info
	}
	return result
}

// GetFactory 获取全局插件工厂
func GetFactory(name string) (PluginFactory, core.PluginMetadata, error) {
	return globalRegistry.GetFactory(name)
}

// ListPlugins 列出全局已注册的插件
func ListPlugins() []string {
	return globalRegistry.ListPlugins()
}

// GetAllFactories 获取全局所有插件工厂
func GetAllFactories() map[string]*FactoryInfo {
	return globalRegistry.GetAllFactories()
}
