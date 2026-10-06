package main

import (
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

func TestSecurityWarnings(t *testing.T) {
	weak := &config.Config{}
	weak.App.Mode = "release"
	weak.JWT.Secret = "short"
	weak.Database.MySQL.Username = "root"
	if got := len(securityWarnings(weak)); got != 4 {
		t.Errorf("弱配置应产生 4 条告警, got %d", got)
	}

	strong := &config.Config{}
	strong.App.Mode = "release"
	strong.JWT.Secret = "0123456789abcdef0123456789abcdef"
	strong.Database.MySQL.Username = "sky_app"
	strong.Redis.Password = "redis-pass"
	strong.Security.FieldEncryptKey = "field-key"
	if got := securityWarnings(strong); len(got) != 0 {
		t.Errorf("安全配置不应告警, got %v", got)
	}

	debug := &config.Config{}
	debug.App.Mode = "debug"
	if got := securityWarnings(debug); len(got) != 0 {
		t.Errorf("debug 模式不检查, got %v", got)
	}
}
