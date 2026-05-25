package email

import (
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

func TestNewService(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  false,
		From:     "noreply@example.com",
		SmtpHost: "smtp.example.com",
		SmtpPort: 465,
	}

	svc := NewService(cfg)
	if svc == nil {
		t.Fatal("Service should not be nil")
	}
}

func TestNewServiceWithQueue_NilQueue(t *testing.T) {
	cfg := &config.EmailConfig{Enabled: true}

	// 队列为 nil 时应正常工作
	svc := NewServiceWithQueue(cfg, nil)
	if svc == nil {
		t.Fatal("Service should not be nil with nil queue")
	}
}

func TestSendMail_DisabledService(t *testing.T) {
	cfg := &config.EmailConfig{Enabled: false}
	svc := NewService(cfg)

	err := svc.SendMail("test@example.com", "Subject", "Body")
	if err == nil {
		t.Error("Expected error when email service disabled")
	}
}

func TestSendVerificationCode_DisabledService(t *testing.T) {
	cfg := &config.EmailConfig{Enabled: false}
	svc := NewService(cfg)

	err := svc.SendVerificationCode("test@example.com", "123456")
	if err == nil {
		t.Error("Expected error when email service disabled")
	}
}

func TestBuildMessage(t *testing.T) {
	msg := buildMessage("from@example.com", "to@example.com", "Test Subject", "<p>Test</p>")

	if msg == "" {
		t.Error("Message should not be empty")
	}

	// 验证关键字段
	expected := []string{
		"From: from@example.com",
		"To: to@example.com",
		"Subject: Test Subject",
		"MIME-Version: 1.0",
		"<p>Test</p>",
	}
	for _, s := range expected {
		if !contains(msg, s) {
			t.Errorf("Message should contain %q", s)
		}
	}
}

func TestBuildVerificationCodeHTML(t *testing.T) {
	html := buildVerificationCodeHTML("987654")
	if !contains(html, "987654") {
		t.Error("HTML should contain the verification code")
	}
	if !contains(html, "企业注册验证码") {
		t.Error("HTML should contain title")
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