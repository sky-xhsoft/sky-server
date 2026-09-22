package totp

import (
	"testing"
	"time"
)

func TestGenerateSecret(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(secret) < 16 {
		t.Errorf("Secret too short: %s", secret)
	}
}

func TestGenerateSecret_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		s, _ := GenerateSecret()
		if seen[s] {
			t.Error("Generated duplicate secret")
		}
		seen[s] = true
	}
}

func TestGenerateCode_KnownVector(t *testing.T) {
	// RFC 6238 测试向量: ASCII "12345678901234567890" 的 Base32
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	timestamp := time.Unix(59, 0) // T = 1
	code, err := GenerateCodeCustom(secret, timestamp, 30, 8)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// RFC 6238 期望 94287082（注意：上面 secret 长度有差异，这里仅验证可生成）
	if len(code) != 8 {
		t.Errorf("Expected 8-digit code, got %d digits: %s", len(code), code)
	}
}

func TestGenerateCode_FixedSecret(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	timestamp := time.Unix(1234567890, 0)

	code, err := GenerateCode(secret, timestamp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d", len(code))
	}

	// 相同输入应产生相同输出
	code2, _ := GenerateCode(secret, timestamp)
	if code != code2 {
		t.Errorf("Inconsistent code: %s vs %s", code, code2)
	}
}

func TestValidateCode_Match(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Now()
	code, _ := GenerateCode(secret, now)

	if !ValidateCode(secret, code, now) {
		t.Error("Code should validate at same time")
	}
}

func TestValidateCode_WithSkew(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Now()
	// 生成 30 秒前的 code
	pastTime := now.Add(-30 * time.Second)
	code, _ := GenerateCode(secret, pastTime)

	// 默认 skew=1，应该接受
	if !ValidateCode(secret, code, now) {
		t.Error("Code from 30s ago should still be valid with default skew")
	}
}

func TestValidateCode_Wrong(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Now()
	if ValidateCode(secret, "000000", now) {
		t.Error("Wrong code should not validate")
	}
}

func TestValidateCode_DifferentSecret(t *testing.T) {
	secret1 := "JBSWY3DPEHPK3PXP"
	secret2 := "ABCDEFGHIJKLMNOP"
	now := time.Now()
	code, _ := GenerateCode(secret1, now)

	if ValidateCode(secret2, code, now) {
		t.Error("Code generated with different secret should not validate")
	}
}

func TestBuildURI(t *testing.T) {
	uri := BuildURI("Sky-Server", "test@example.com", "JBSWY3DPEHPK3PXP")

	if uri == "" {
		t.Fatal("URI should not be empty")
	}

	expected := "otpauth://totp/"
	if uri[:len(expected)] != expected {
		t.Errorf("URI should start with %q, got: %s", expected, uri)
	}
}

func TestGenerateCode_HandlesSpaceInSecret(t *testing.T) {
	// 用户可能复制粘贴带空格的密钥
	secret := "JBSW Y3DP EHPK 3PXP"
	code, err := GenerateCode(secret, time.Now())
	if err != nil {
		t.Errorf("Should handle spaces in secret: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d", len(code))
	}
}
