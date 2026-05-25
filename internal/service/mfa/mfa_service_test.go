package mfa

import (
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/totp"
)

func TestInMemoryStore_GetSetDelete(t *testing.T) {
	store := NewInMemoryStore()

	// 初始为空
	_, err := store.Get(1)
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound, got %v", err)
	}

	// 设置后可获取
	if err := store.Set(1, "secret-abc"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := store.Get(1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != "secret-abc" {
		t.Errorf("Expected 'secret-abc', got %s", got)
	}

	// 删除后再次为空
	store.Delete(1)
	_, err = store.Get(1)
	if err != ErrSecretNotFound {
		t.Error("Expected ErrSecretNotFound after delete")
	}
}

func TestMFAService_EnrollUser(t *testing.T) {
	svc := NewService(NewInMemoryStore(), "Test-App")

	secret, uri, err := svc.EnrollUser(1, "alice@example.com")
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}
	if secret == "" {
		t.Error("Secret should not be empty")
	}
	if uri == "" {
		t.Error("URI should not be empty")
	}
	// URI 应包含 issuer 和账户
	expectedPrefix := "otpauth://totp/"
	if uri[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("URI should start with %s, got %s", expectedPrefix, uri)
	}
}

func TestMFAService_VerifyAndEnable(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store, "Test-App")

	// 步骤 1: Enroll
	secret, _, err := svc.EnrollUser(1, "alice")
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// 步骤 2: 生成正确的 TOTP 码并验证
	code, _ := totp.GenerateCode(secret, time.Now())
	if err := svc.VerifyCode(1, code); err != nil {
		t.Fatalf("VerifyCode should succeed: %v", err)
	}

	// 步骤 3: 验证密钥已持久化
	storedSecret, _ := store.Get(1)
	if storedSecret != secret {
		t.Error("Secret should be persisted after successful verification")
	}

	// 步骤 4: 后续验证应该用 store 中的密钥
	code2, _ := totp.GenerateCode(secret, time.Now())
	if err := svc.VerifyCode(1, code2); err != nil {
		t.Errorf("Subsequent verify should succeed: %v", err)
	}
}

func TestMFAService_VerifyCode_Invalid(t *testing.T) {
	svc := NewService(NewInMemoryStore(), "App")
	svc.EnrollUser(1, "alice")

	err := svc.VerifyCode(1, "000000")
	if err != ErrInvalidCode {
		t.Errorf("Expected ErrInvalidCode, got %v", err)
	}
}

func TestMFAService_VerifyCode_UserNotEnrolled(t *testing.T) {
	svc := NewService(NewInMemoryStore(), "App")

	err := svc.VerifyCode(999, "123456")
	if err == nil {
		t.Error("Expected error for un-enrolled user")
	}
}

func TestMFAService_IsEnabled(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store, "App")

	if svc.IsEnabled(1) {
		t.Error("Should not be enabled initially")
	}

	store.Set(1, "secret")
	if !svc.IsEnabled(1) {
		t.Error("Should be enabled after Set")
	}
}

func TestMFAService_DisableMFA(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store, "App")

	// 启用
	secret, _, _ := svc.EnrollUser(1, "alice")
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyCode(1, code)

	if !svc.IsEnabled(1) {
		t.Fatal("Should be enabled before disable")
	}

	// 禁用
	if err := svc.DisableMFA(1); err != nil {
		t.Errorf("Disable failed: %v", err)
	}

	if svc.IsEnabled(1) {
		t.Error("Should not be enabled after disable")
	}
}

func TestMFAService_DefaultIssuer(t *testing.T) {
	svc := NewService(NewInMemoryStore(), "")
	_, uri, _ := svc.EnrollUser(1, "alice")

	// 默认 issuer 为 "Sky-Server"
	if !contains(uri, "Sky-Server") {
		t.Errorf("URI should contain default issuer 'Sky-Server', got: %s", uri)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
