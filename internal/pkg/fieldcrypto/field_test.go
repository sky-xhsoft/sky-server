package fieldcrypto

import (
	"strings"
	"testing"
)

func TestEncryptField_RoundTrip(t *testing.T) {
	Setup("test-field-key")
	plaintext := "tencent-secret-key-1234567890abcd"

	encrypted, err := EncryptField(plaintext)
	if err != nil {
		t.Fatalf("EncryptField failed: %v", err)
	}
	if !strings.HasPrefix(encrypted, EncPrefix) {
		t.Errorf("Expected prefix %q, got %q", EncPrefix, encrypted)
	}

	decrypted, err := DecryptField(encrypted)
	if err != nil {
		t.Fatalf("DecryptField failed: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptField_EmptyAndAlreadyEncrypted(t *testing.T) {
	Setup("test-field-key")

	if v, _ := EncryptField(""); v != "" {
		t.Errorf("Empty value should stay empty, got %q", v)
	}

	once, _ := EncryptField("secret")
	twice, err := EncryptField(once)
	if err != nil {
		t.Fatalf("EncryptField failed: %v", err)
	}
	if twice != once {
		t.Error("Already encrypted value should not be encrypted again")
	}
}

func TestDecryptField_LegacyPlaintext(t *testing.T) {
	Setup("test-field-key")
	legacy := "plain-legacy-secret"

	v, err := DecryptField(legacy)
	if err != nil {
		t.Fatalf("DecryptField failed: %v", err)
	}
	if v != legacy {
		t.Errorf("Legacy plaintext should be returned as-is, got %q", v)
	}
}

func TestDecryptField_WrongKey(t *testing.T) {
	Setup("key-a")
	encrypted, _ := EncryptField("secret")

	Setup("key-b")
	if _, err := DecryptField(encrypted); err == nil {
		t.Error("Expected error when decrypting with a different key")
	}
}
