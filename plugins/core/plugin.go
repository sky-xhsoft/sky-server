package core

import (
	"context"

	"gorm.io/gorm"
)

// Plugin 插件接口
// 所有插件必须实现此接口
type Plugin interface {
	// Name 返回插件唯一标识名称
	Name() string

	// Description 返回插件描述信息
	Description() string

	// Version 返回插件版本
	Version() string

	// Execute 执行插件逻辑
	Execute(ctx context.Context, db *gorm.DB, data PluginData) error
}

// PluginData 插件执行时的数据上下文
type PluginData struct {
	// TableName 表名
	TableName string `json:"tableName"`

	// Action 操作类型: create, update, delete, query, submit, unsubmit
	Action string `json:"action"`

	// Timing 执行时机: before, after
	Timing string `json:"timing"`

	// RecordID 记录ID
	RecordID uint `json:"recordId"`

	// Data 数据内容（来自 HTTP 请求或数据库记录）
	Data map[string]interface{} `json:"data"`

	// UserID 操作用户ID
	UserID uint `json:"userId"`

	// CompanyID 公司ID
	CompanyID uint `json:"companyId"`

	// Extra 额外的上下文数据
	Extra map[string]interface{} `json:"extra"`
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	// Name 插件名称
	Name string

	// Description 插件描述
	Description string

	// Version 插件版本
	Version string

	// Author 插件作者
	Author string

	// Enabled 是否启用
	Enabled bool

	// Priority 执行优先级（数字越小优先级越高）
	Priority int

	// HookPoint 钩子点（如 "sys_table.after.create"）
	HookPoint string
}

// PluginInfo 插件信息（包含插件实例和元数据）
type PluginInfo struct {
	Plugin   Plugin
	Metadata PluginMetadata
}
