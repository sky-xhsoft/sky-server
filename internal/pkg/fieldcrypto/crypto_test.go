package fieldcrypto

import (
	"strings"
	"testing"
)

func TestNewEncryptor_InvalidKeyLength(t *testing.T) {
	_, err := NewEncryptor([]byte("too-short"))
	if err == nil {
		t.Error("Expected error for invalid key length")
	}
}

func TestNewEncryptor_Success(t *testing.T) {
	key := make([]byte, 32)
	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if enc == nil {
		t.Fatal("Encryptor should not be nil")
	}
}

func TestNewEncryptorFromString(t *testing.T) {
	enc, err := NewEncryptorFromString("my-secret-key")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if enc == nil {
		t.Fatal("Encryptor should not be nil")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	enc, _ := NewEncryptorFromString("test-encryption-key-12345")
	plaintext := "sensitive data: SSN 123-45-6789"

	encrypted, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted == plaintext {
		t.Error("Encrypted should differ from plaintext")
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncrypt_EmptyString(t *testing.T) {
	enc, _ := NewEncryptorFromString("key")
	result, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Empty input should produce empty output, got %q", result)
	}
}

func TestEncrypt_DifferentNoncePerCall(t *testing.T) {
	enc, _ := NewEncryptorFromString("test-key-12345")
	plaintext := "same input"

	enc1, _ := enc.Encrypt(plaintext)
	enc2, _ := enc.Encrypt(plaintext)

	if enc1 == enc2 {
		t.Error("Same plaintext should produce different ciphertext (random nonce)")
	}

	// 都能正确解密
	dec1, _ := enc.Decrypt(enc1)
	dec2, _ := enc.Decrypt(enc2)
	if dec1 != plaintext || dec2 != plaintext {
		t.Error("Both ciphertexts should decrypt to same plaintext")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	enc, _ := NewEncryptorFromString("key")
	_, err := enc.Decrypt("!!!not-base64!!!")
	if err == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	enc, _ := NewEncryptorFromString("key")
	_, err := enc.Decrypt("aGk=") // valid base64 "hi" but too short
	if err == nil {
		t.Error("Expected error for too short ciphertext")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	enc, _ := NewEncryptorFromString("test-key")
	encrypted, _ := enc.Encrypt("original message")

	// 篡改密文（替换最后一个字符）
	tampered := encrypted[:len(encrypted)-1] + "X"
	if tampered == encrypted {
		tampered = encrypted[:len(encrypted)-2] + "XX"
	}

	_, err := enc.Decrypt(tampered)
	if err == nil {
		t.Error("Tampered ciphertext should fail to decrypt (GCM authentication)")
	}
}

func TestEncryptDecrypt_Bytes(t *testing.T) {
	enc, _ := NewEncryptorFromString("byte-test-key")
	data := []byte{0x00, 0x01, 0x02, 0xff, 0xfe}

	encrypted, err := enc.EncryptBytes(data)
	if err != nil {
		t.Fatalf("EncryptBytes failed: %v", err)
	}

	decrypted, err := enc.DecryptBytes(encrypted)
	if err != nil {
		t.Fatalf("DecryptBytes failed: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Error("Byte round-trip mismatch")
	}
}

func TestEncrypt_LongData(t *testing.T) {
	enc, _ := NewEncryptorFromString("long-data-test")
	plaintext := strings.Repeat("Long sensitive data block. ", 1000)

	encrypted, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt long data failed: %v", err)
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt long data failed: %v", err)
	}

	if decrypted != plaintext {
		t.Error("Long data round-trip mismatch")
	}
}