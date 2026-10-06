package main

import (
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

func TestSecurityWarnings(t *testing.T) {
	weak := &config.Config{}
	weak.App.Mode = "release"
	weak.JWT.Secret = "dev_short"
	weak.Database.MySQL.Username = "root"
	// 回调密钥、短 JWT 密钥、开发默认值、root、无 Redis 密码、无字段加密密钥
	if got := len(securityWarnings(weak)); got != 6 {
		t.Errorf("弱配置应产生 6 条告警, got %d: %v", got, securityWarnings(weak))
	}

	strong := &config.Config{}
	strong.App.Mode = "release"
	strong.JWT.Secret = "0123456789abcdef0123456789abcdef"
	strong.Database.MySQL.Username = "sky_app"
	strong.Redis.Password = "redis-pass"
	strong.Security.FieldEncryptKey = "field-key"
	strong.TencentCloud.CallbackKey = "callback-key"
	if got := securityWarnings(strong); len(got) != 0 {
		t.Errorf("安全配置不应告警, got %v", got)
	}

	// debug 模式只提示回调密钥（debug 下不校验回调签名）
	debug := &config.Config{}
	debug.App.Mode = "debug"
	if got := securityWarnings(debug); len(got) != 1 {
		t.Errorf("debug 模式只应提示回调密钥, got %v", got)
	}
	debug.TencentCloud.CallbackKey = "callback-key"
	if got := securityWarnings(debug); len(got) != 0 {
		t.Errorf("debug 模式配置回调密钥后不应告警, got %v", got)
	}
}
