package core

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"gorm.io/gorm"
)

// Manager 插件管理器
// 负责插件的注册、管理和执行
type Manager struct {
	db      *gorm.DB
	plugins map[string][]*PluginInfo // hookPoint -> []*PluginInfo
	mu      sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:      db,
		plugins: make(map[string][]*PluginInfo),
	}
}

// Register 注册插件到指定钩子点
// hookPoint 格式：tableName.timing.action
// 例如：sys_table.after.create, sys_user.before.update
func (m *Manager) Register(hookPoint string, plugin Plugin, metadata PluginMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查插件是否已注册
	for _, info := range m.plugins[hookPoint] {
		if info.Plugin.Name() == plugin.Name() {
			return fmt.Errorf("插件 %s 已在钩子点 %s 注册", plugin.Name(), hookPoint)
		}
	}

	// 填充元数据的默认值
	if metadata.Name == "" {
		metadata.Name = plugin.Name()
	}
	if metadata.Description == "" {
		metadata.Description = plugin.Description()
	}
	if metadata.Version == "" {
		metadata.Version = plugin.Version()
	}
	metadata.HookPoint = hookPoint

	// 如果未指定优先级，使用当前插件数量作为优先级
	if metadata.Priority == 0 {
		metadata.Priority = len(m.plugins[hookPoint]) + 1
	}

	// 添加插件
	info := &PluginInfo{
		Plugin:   plugin,
		Metadata: metadata,
	}

	m.plugins[hookPoint] = append(m.plugins[hookPoint], info)

	// 按优先级排序（数字越小优先级越高）
	sort.Slice(m.plugins[hookPoint], func(i, j int) bool {
		return m.plugins[hookPoint][i].Metadata.Priority < m.plugins[hookPoint][j].Metadata.Priority
	})

	return nil
}

// Unregister 取消注册插件
func (m *Manager) Unregister(hookPoint string, pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, exists := m.plugins[hookPoint]
	if !exists {
		return fmt.Errorf("钩子点 %s 不存在", hookPoint)
	}

	for i, info := range plugins {
		if info.Plugin.Name() == pluginName {
			// 从切片中删除
			m.plugins[hookPoint] = append(plugins[:i], plugins[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("插件 %s 未在钩子点 %s 注册", pluginName, hookPoint)
}

// Execute 执行指定钩子点的所有插件（使用默认数据库连接）
func (m *Manager) Execute(ctx context.Context, data PluginData) error {
	return m.ExecuteWithDB(ctx, m.db, data)
}

// ExecuteWithDB 使用指定的数据库连接执行插件（支持事务）
func (m *Manager) ExecuteWithDB(ctx context.Context, db *gorm.DB, data PluginData) error {
	// 构建钩子点名称
	hookPoint := fmt.Sprintf("%s.%s.%s", data.TableName, data.Timing, data.Action)

	m.mu.RLock()
	plugins := m.plugins[hookPoint]
	m.mu.RUnlock()

	if len(plugins) == 0 {
		return nil // 没有插件，直接返回
	}

	// 按优先级执行所有启用的插件
	for _, info := range plugins {
		if !info.Metadata.Enabled {
			continue // 跳过未启用的插件
		}

		if err := info.Plugin.Execute(ctx, db, data); err != nil {
			return fmt.Errorf("插件 %s 执行失败: %w", info.Plugin.Name(), err)
		}
	}

	return nil
}

// ExecutePlugins 执行指定表的插件（兼容旧接口）
func (m *Manager) ExecutePlugins(ctx context.Context, data PluginData) error {
	return m.Execute(ctx, data)
}

// GetPlugins 获取指定钩子点的所有插件
func (m *Manager) GetPlugins(hookPoint string) []*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := m.plugins[hookPoint]
	result := make([]*PluginInfo, len(plugins))
	copy(result, plugins)
	return result
}

// GetAllPlugins 获取所有已注册的插件
func (m *Manager) GetAllPlugins() map[string][]*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]*PluginInfo)
	for hookPoint, plugins := range m.plugins {
		pluginsCopy := make([]*PluginInfo, len(plugins))
		copy(pluginsCopy, plugins)
		result[hookPoint] = pluginsCopy
	}
	return result
}

// EnablePlugin 启用插件
func (m *Manager) EnablePlugin(hookPoint string, pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, exists := m.plugins[hookPoint]
	if !exists {
		return fmt.Errorf("钩子点 %s 不存在", hookPoint)
	}

	for _, info := range plugins {
		if info.Plugin.Name() == pluginName {
			info.Metadata.Enabled = true
			return nil
		}
	}

	return fmt.Errorf("插件 %s 未在钩子点 %s 注册", pluginName, hookPoint)
}

// DisablePlugin 禁用插件
func (m *Manager) DisablePlugin(hookPoint string, pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, exists := m.plugins[hookPoint]
	if !exists {
		return fmt.Errorf("钩子点 %s 不存在", hookPoint)
	}

	for _, info := range plugins {
		if info.Plugin.Name() == pluginName {
			info.Metadata.Enabled = false
			return nil
		}
	}

	return fmt.Errorf("插件 %s 未在钩子点 %s 注册", pluginName, hookPoint)
}

// ListHookPoints 列出所有钩子点
func (m *Manager) ListHookPoints() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hookPoints := make([]string, 0, len(m.plugins))
	for hookPoint := range m.plugins {
		hookPoints = append(hookPoints, hookPoint)
	}
	sort.Strings(hookPoints)
	return hookPoints
}
