package config

import (
	"os"
	"testing"
)

func TestExpandEnvStringWithDefault(t *testing.T) {
	t.Setenv("S0_TEST_VALUE", "from-env")

	if got := expandEnvString("${S0_TEST_VALUE:-fallback}"); got != "from-env" {
		t.Fatalf("expandEnvString() = %q, want %q", got, "from-env")
	}

	if got := expandEnvString("${S0_TEST_MISSING:-fallback}"); got != "fallback" {
		t.Fatalf("expandEnvString() = %q, want %q", got, "fallback")
	}
}

func TestExpandEnvDefaults(t *testing.T) {
	t.Setenv("S0_JWT_SECRET", "safe-secret")
	os.Unsetenv("S0_DB_HOST")

	cfg := &Config{
		JWT: JWTConfig{Secret: "${S0_JWT_SECRET}"},
		Database: DatabaseConfig{MySQL: MySQLConfig{
			Host:     "${S0_DB_HOST:-localhost}",
			Password: "${S0_DB_PASSWORD}",
		}},
	}

	expandEnvDefaults(cfg)

	if cfg.JWT.Secret != "safe-secret" {
		t.Fatalf("JWT.Secret = %q, want safe-secret", cfg.JWT.Secret)
	}
	if cfg.Database.MySQL.Host != "localhost" {
		t.Fatalf("Database.MySQL.Host = %q, want localhost", cfg.Database.MySQL.Host)
	}
	if cfg.Database.MySQL.Password != "" {
		t.Fatalf("Database.MySQL.Password = %q, want empty", cfg.Database.MySQL.Password)
	}
}
