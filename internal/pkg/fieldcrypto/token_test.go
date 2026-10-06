package fieldcrypto

import (
	"strings"
	"testing"
)

func TestSealOpenToken_RoundTrip(t *testing.T) {
	Setup("test-token-key")
	payload := "v1|25|1700000000|abcd"

	token, err := SealToken(payload)
	if err != nil {
		t.Fatalf("SealToken failed: %v", err)
	}
	if strings.ContainsAny(token, "+/=") {
		t.Errorf("Token should be URL safe, got %q", token)
	}

	got, err := OpenToken(token)
	if err != nil {
		t.Fatalf("OpenToken failed: %v", err)
	}
	if got != payload {
		t.Errorf("Expected %q, got %q", payload, got)
	}
}

func TestOpenToken_Tampered(t *testing.T) {
	Setup("test-token-key")
	token, _ := SealToken("v1|25|1700000000|abcd")

	// 修改中间字符（末位字符的低位在无填充 base64 中可能被忽略）
	tampered := []byte(token)
	mid := len(tampered) / 2
	if tampered[mid] == 'A' {
		tampered[mid] = 'B'
	} else {
		tampered[mid] = 'A'
	}
	if _, err := OpenToken(string(tampered)); err == nil {
		t.Error("Expected error for tampered token")
	}
}

func TestOpenToken_Invalid(t *testing.T) {
	Setup("test-token-key")
	for _, token := range []string{"", "not-base64!!", "YWJj"} {
		if _, err := OpenToken(token); err == nil {
			t.Errorf("Expected error for token %q", token)
		}
	}
}

func TestOpenToken_WrongKey(t *testing.T) {
	Setup("key-a")
	token, _ := SealToken("payload")

	Setup("key-b")
	if _, err := OpenToken(token); err == nil {
		t.Error("Expected error when opening with a different key")
	}
}
