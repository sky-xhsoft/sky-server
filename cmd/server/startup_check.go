package main

import (
	"context"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// securityWarnings release 模式下的弱配置检查
func securityWarnings(cfg *config.Config) []string {
	var warnings []string
	// 回调密钥为空时 debug 模式会接受未签名的直播回调，任何模式都提示
	if cfg.TencentCloud.CallbackKey == "" {
		warnings = append(warnings, "tencentCloud.callbackKey 未配置：release 模式将拒绝全部直播回调，debug 模式不校验回调签名")
	}
	if cfg.App.Mode != "release" {
		return warnings
	}
	if len(cfg.JWT.Secret) < 32 {
		warnings = append(warnings, "jwt.secret 少于 32 位，容易被暴力破解")
	}
	if lower := strings.ToLower(cfg.JWT.Secret); strings.HasPrefix(lower, "dev") || strings.Contains(lower, "change") || strings.Contains(lower, "secret") {
		warnings = append(warnings, "jwt.secret 看起来是开发/示例默认值，请更换为随机密钥")
	}
	if cfg.Database.MySQL.Username == "root" {
		warnings = append(warnings, "数据库使用 root 账号，建议改用仅有本库权限的专用账号")
	}
	if cfg.Redis.Password == "" {
		warnings = append(warnings, "Redis 未设置密码")
	}
	if cfg.Security.FieldEncryptKey == "" {
		warnings = append(warnings, "security.fieldEncryptKey 未配置，敏感字段加密回退使用 jwt.secret")
	}
	return warnings
}

// runStartupChecks 启动自检：只记录告警，不阻止启动
func runStartupChecks(cfg *config.Config, db *gorm.DB) {
	for _, w := range securityWarnings(cfg) {
		logger.Warn("Startup check: insecure config", zap.String("detail", w))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	issues, err := metadata.CheckSchema(ctx, db)
	if err != nil {
		logger.Warn("Startup check: schema check failed", zap.Error(err))
		return
	}
	for _, issue := range issues {
		logger.Warn("Startup check: metadata does not match database", zap.String("detail", issue))
	}
	if len(issues) == 0 {
		logger.Info("Startup check: metadata matches database schema")
	}
}
