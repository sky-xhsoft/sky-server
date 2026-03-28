package service

import (
	"sync"
)

// Container 依赖注入容器
type Container struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

var (
	instance *Container
	once     sync.Once
)

// GetContainer 获取容器单例
func GetContainer() *Container {
	once.Do(func() {
		instance = &Container{
			services: make(map[string]interface{}),
		}
	})
	return instance
}

// Register 注册服务
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// Get 获取服务
func (c *Container) Get(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, ok := c.services[name]
	return service, ok
}

// MustGet 获取服务，不存在则panic
func (c *Container) MustGet(name string) interface{} {
	service, ok := c.Get(name)
	if !ok {
		panic("service not registered: " + name)
	}
	return service
}

// Clear 清空容器
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[string]interface{})
}
