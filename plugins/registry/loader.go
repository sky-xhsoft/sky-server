package registry

import (
	"fmt"

	"github.com/sky-xhsoft/sky-server/plugins/core"
)

// Loader 插件加载器
// 负责从注册中心加载插件并注册到管理器
type Loader struct {
	manager *core.Manager
}

// NewLoader 创建插件加载器
func NewLoader(manager *core.Manager) *Loader {
	return &Loader{
		manager: manager,
	}
}

// LoadAll 加载所有已注册的插件
// 根据插件的 HookPoint 元数据自动注册到对应的钩子点
func (l *Loader) LoadAll() error {
	factories := GetAllFactories()

	for name, info := range factories {
		if !info.Metadata.Enabled {
			continue // 跳过未启用的插件
		}

		if err := l.Load(name); err != nil {
			return fmt.Errorf("加载插件 %s 失败: %w", name, err)
		}
	}

	return nil
}

// Load 加载指定的插件
func (l *Loader) Load(name string) error {
	factory, metadata, err := GetFactory(name)
	if err != nil {
		return err
	}

	if !metadata.Enabled {
		return fmt.Errorf("插件 %s 未启用", name)
	}

	// 创建插件实例
	plugin := factory()

	// 注册到管理器
	if metadata.HookPoint == "" {
		return fmt.Errorf("插件 %s 未指定钩子点", name)
	}

	return l.manager.Register(metadata.HookPoint, plugin, metadata)
}

// LoadByHookPoint 加载指定钩子点的所有插件
func (l *Loader) LoadByHookPoint(hookPoint string) error {
	factories := GetAllFactories()
	loaded := 0

	for name, info := range factories {
		if !info.Metadata.Enabled {
			continue
		}

		if info.Metadata.HookPoint == hookPoint {
			if err := l.Load(name); err != nil {
				return fmt.Errorf("加载插件 %s 失败: %w", name, err)
			}
			loaded++
		}
	}

	if loaded == 0 {
		return fmt.Errorf("钩子点 %s 没有可用的插件", hookPoint)
	}

	return nil
}

// LoadByNames 加载指定名称的插件列表
func (l *Loader) LoadByNames(names []string) error {
	for _, name := range names {
		if err := l.Load(name); err != nil {
			return err
		}
	}
	return nil
}

// Reload 重新加载插件
func (l *Loader) Reload(name string) error {
	factory, metadata, err := GetFactory(name)
	if err != nil {
		return err
	}

	// 先取消注册
	if metadata.HookPoint != "" {
		_ = l.manager.Unregister(metadata.HookPoint, name)
	}

	// 如果插件未启用，不重新注册
	if !metadata.Enabled {
		return nil
	}

	// 重新注册
	plugin := factory()
	return l.manager.Register(metadata.HookPoint, plugin, metadata)
}
