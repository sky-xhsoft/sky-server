package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	j := New("test-secret-key")
	expire := time.Hour

	token, err := j.GenerateToken(1, 1, "admin", "web", "device-1", expire)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := j.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("UserID = %d, want 1", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("Username = %s, want admin", claims.Username)
	}
	if claims.ClientType != "web" {
		t.Errorf("ClientType = %s, want web", claims.ClientType)
	}
	if claims.TokenType != "access" {
		t.Errorf("TokenType = %s, want access", claims.TokenType)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	j := New("test-secret-key")
	expire := 24 * time.Hour

	token, err := j.GenerateRefreshToken(1, 1, "admin", "web", "device-1", expire)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := j.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.TokenType != "refresh" {
		t.Errorf("TokenType = %s, want refresh", claims.TokenType)
	}
}

func TestTokenExpired(t *testing.T) {
	j := New("test-secret-key")
	// 已过期的token
	expire := -1 * time.Hour

	token, err := j.GenerateToken(1, 1, "admin", "web", "device-1", expire)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = j.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected token expired error, got nil")
	}
}

func TestInvalidToken(t *testing.T) {
	j := New("test-secret-key")

	_, err := j.ParseToken("invalid-token-string")
	if err == nil {
		t.Fatal("Expected parse error for invalid token, got nil")
	}
}

func TestParseWithWrongSecret(t *testing.T) {
	j1 := New("secret-1")
	j2 := New("secret-2")
	expire := time.Hour

	token, err := j1.GenerateToken(1, 1, "admin", "web", "device-1", expire)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = j2.ParseToken(token)
	if err == nil {
		t.Fatal("Expected signature validation error, got nil")
	}
}

func TestRefreshToken(t *testing.T) {
	j := New("test-secret-key")
	refreshExpire := 24 * time.Hour
	accessExpire := time.Hour

	refreshToken, err := j.GenerateRefreshToken(1, 1, "admin", "web", "device-1", refreshExpire)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	newAccessToken, err := j.RefreshToken(refreshToken, accessExpire)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if newAccessToken == "" {
		t.Fatal("new access token is empty")
	}

	claims, err := j.ParseToken(newAccessToken)
	if err != nil {
		t.Fatalf("ParseToken for new access token failed: %v", err)
	}
	if claims.TokenType != "access" {
		t.Errorf("New token TokenType = %s, want access", claims.TokenType)
	}
}
