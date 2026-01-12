package builtin

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"github.com/sky-xhsoft/sky-server/plugins/registry"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SysTableBeforeDeletePlugin sys_table 删除前插件
// 删除表时级联删除相关的字段配置和目录
type SysTableBeforeDeletePlugin struct{}

// 通过 init 函数自动注册插件
func init() {
	registry.Register(
		"sys_table_before_delete",
		func() core.Plugin {
			return &SysTableBeforeDeletePlugin{}
		},
		core.PluginMetadata{
			Name:        "sys_table_before_delete",
			Description: "sys_table 删除前级联删除字段和目录",
			Version:     "1.0.0",
			Author:      "Sky-Server",
			Enabled:     true,
			Priority:    10,
			HookPoint:   "sys_table.before.delete",
		},
	)
}

// Name 返回插件名称
func (p *SysTableBeforeDeletePlugin) Name() string {
	return "sys_table_before_delete"
}

// Description 返回插件描述
func (p *SysTableBeforeDeletePlugin) Description() string {
	return "sys_table 删除前级联删除字段和目录"
}

// Version 返回插件版本
func (p *SysTableBeforeDeletePlugin) Version() string {
	return "1.0.0"
}

// Execute 执行插件逻辑
// 删除表时级联删除相关的字段配置和目录
func (p *SysTableBeforeDeletePlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
	tableID := data.RecordID
	if tableID == 0 {
		return fmt.Errorf("table ID is required")
	}

	// 删除该表的所有字段配置（物理删除）
	result := db.WithContext(ctx).Table("sys_column").Where("SYS_TABLE_ID = ?", tableID).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除表字段配置失败: %v", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info("删除表字段配置",
			zap.Uint("tableID", tableID),
			zap.Int64("count", result.RowsAffected))
	}

	// 删除该表关联的目录（物理删除）
	result = db.WithContext(ctx).Table("sys_directory").Where("SYS_TABLE_ID = ?", tableID).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除表目录配置失败: %v", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info("删除表目录配置",
			zap.Uint("tableID", tableID),
			zap.Int64("count", result.RowsAffected))
	}

	return nil
}
