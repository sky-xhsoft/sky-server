// 插件模板
// 复制此文件到你的插件目录并修改
// 例如：cp TEMPLATE.go myplugin/plugin.go

package main

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"github.com/sky-xhsoft/sky-server/plugins/registry"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// =====================================================
// 1. 定义插件结构体
// =====================================================

// MyPlugin 你的插件名称
type MyPlugin struct {
	// 可以添加插件需要的字段
}

// =====================================================
// 2. Register 函数（必须导出）
//    这是插件的入口函数，系统会调用它来注册插件
// =====================================================

// Register 注册插件到系统
// 必须导出（首字母大写）
func Register() {
	registry.Register(
		"my_plugin", // 插件名称（必须与数据库配置一致）
		func() core.Plugin {
			return &MyPlugin{} // 返回插件实例
		},
		core.PluginMetadata{
			Name:        "my_plugin",
			Description: "我的插件描述",
			Version:     "1.0.0",
			Author:      "Your Name",
			Enabled:     true,
			Priority:    50,                      // 优先级（数字越小越先执行）
			HookPoint:   "sys_user.after.create", // 钩子点（必须与数据库配置一致）
		},
	)

	logger.Info("📦 MyPlugin 已注册")
}

// =====================================================
// 3. 实现 Plugin 接口的方法
// =====================================================

// Name 返回插件名称
func (p *MyPlugin) Name() string {
	return "my_plugin"
}

// Description 返回插件描述
func (p *MyPlugin) Description() string {
	return "我的插件描述"
}

// Version 返回插件版本
func (p *MyPlugin) Version() string {
	return "1.0.0"
}

// Execute 执行插件逻辑
// 这是插件的核心业务逻辑
func (p *MyPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
	logger.Info("🎯 MyPlugin 执行",
		zap.String("table", data.TableName),
		zap.String("action", data.Action),
		zap.Uint("recordID", data.RecordID))

	// =====================================================
	// 在这里编写你的插件逻辑
	// =====================================================

	// 示例 1: 访问记录数据
	if username, ok := data.Data["USERNAME"]; ok {
		fmt.Printf("✨ 新用户注册: %v\n", username)
	}

	// 示例 2: 使用数据库连接（在同一事务中）
	// var user model.SysUser
	// if err := db.First(&user, data.RecordID).Error; err != nil {
	//     return err
	// }

	// 示例 3: 调用外部服务
	// sendWelcomeEmail(username)

	// 示例 4: 记录日志
	logger.Info("插件执行完成",
		zap.String("plugin", p.Name()),
		zap.Uint("recordID", data.RecordID))

	return nil
}

// =====================================================
// 可选：添加辅助方法
// =====================================================

// 你可以添加更多辅助方法来组织代码
// func (p *MyPlugin) helperMethod() {
//     // ...
// }
