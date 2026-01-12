// Example Hot-Reload Plugin
// 热加载示例插件 - 演示 JSP 风格的动态编译和热重载

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

// ExampleHotloadPlugin 热加载示例插件
type ExampleHotloadPlugin struct{}

// Register 注册插件（必须导出）
func Register() {
	registry.Register(
		"example_hotload",
		func() core.Plugin {
			return &ExampleHotloadPlugin{}
		},
		core.PluginMetadata{
			Name:        "example_hotload",
			Description: "热加载示例插件 - 演示动态编译和热重载",
			Version:     "1.0.0",
			Author:      "Sky-Server Team",
			Enabled:     true,
			Priority:    50,
			HookPoint:   "sys_user.after.create",
		},
	)

	logger.Info("📦 Example Hotload Plugin 已注册")
}

func (p *ExampleHotloadPlugin) Name() string {
	return "example_hotload"
}

func (p *ExampleHotloadPlugin) Description() string {
	return "热加载示例插件 - 演示动态编译和热重载"
}

func (p *ExampleHotloadPlugin) Version() string {
	return "1.0.0"
}

func (p *ExampleHotloadPlugin) Execute(ctx context.Context, db *gorm.DB, data core.PluginData) error {
	logger.Info("🎯 Example Hotload Plugin 执行",
		zap.String("table", data.TableName),
		zap.String("action", data.Action),
		zap.Uint("recordID", data.RecordID))

	// 示例：打印用户信息
	if username, ok := data.Data["USERNAME"]; ok {
		fmt.Printf("🚀 热加载插件检测到新用户: %v\n", username)
		fmt.Println("✨ 这是一个可以动态编译和热重载的插件！")
		fmt.Println("💡 提示：修改此文件后保存，系统会自动重新编译和加载")
	}

	return nil
}
