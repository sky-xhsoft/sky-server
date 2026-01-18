package hooks

import (
	"github.com/sky-xhsoft/sky-server/internal/pkg/executor"
	"github.com/sky-xhsoft/sky-server/plugins/core"
)

// HookRegistrar 定义 hook 注册器接口
// 所有 hook 必须实现此接口以支持自动注册
type HookRegistrar interface {
	// Name 返回 hook 函数的名称（用于 executor.RegisterGoFunc）
	Name() string

	// Register 执行注册逻辑，将 hook 注册到 executor
	Register(manager *core.Manager)
}

// hookRegistry 存储所有已注册的 hook
var hookRegistry []HookRegistrar

// Register 注册一个 hook（在各个 hook 文件的 init() 中调用）
func Register(hook HookRegistrar) {
	hookRegistry = append(hookRegistry, hook)
}

// RegisterAll 注册所有已收集的 hooks 到 executor
// 在 plugins.Setup() 中调用
func RegisterAll(manager *core.Manager) {
	for _, hook := range hookRegistry {
		hook.Register(manager)
	}
}

// GetRegisteredHooks 获取所有已注册的 hooks（用于调试）
func GetRegisteredHooks() []string {
	names := make([]string, 0, len(hookRegistry))
	for _, hook := range hookRegistry {
		names = append(names, hook.Name())
	}
	return names
}

// BaseHook 提供基础的 hook 实现
// 可以嵌入到具体的 hook 中复用代码
type BaseHook struct {
	hookName string
	handler  func(manager *core.Manager) func(map[string]interface{}) (interface{}, error)
}

// NewBaseHook 创建基础 hook
func NewBaseHook(name string, handler func(manager *core.Manager) func(map[string]interface{}) (interface{}, error)) *BaseHook {
	return &BaseHook{
		hookName: name,
		handler:  handler,
	}
}

// Name 返回 hook 名称
func (h *BaseHook) Name() string {
	return h.hookName
}

// Register 注册到 executor
func (h *BaseHook) Register(manager *core.Manager) {
	executor.RegisterGoFunc(h.hookName, h.handler(manager))
}
