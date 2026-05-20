package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

// Service 邮件服务接口
type Service interface {
	SendVerificationCode(to, code string) error
	SendMail(to, subject, body string) error
}

type service struct {
	cfg *config.EmailConfig
}

// NewService 创建邮件服务
func NewService(cfg *config.EmailConfig) Service {
	return &service{cfg: cfg}
}

// SendMail 发送邮件
func (s *service) SendMail(to, subject, body string) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("邮件服务未启用")
	}

	msg := buildMessage(s.cfg.From, to, subject, body)
	addr := fmt.Sprintf("%s:%d", s.cfg.SmtpHost, s.cfg.SmtpPort)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.SmtpHost)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.cfg.SmtpHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("连接 SMTP 失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SmtpHost)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}
	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备邮件数据失败: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件数据失败: %w", err)
	}
	return w.Close()
}

// SendVerificationCode 发送验证码
func (s *service) SendVerificationCode(to, code string) error {
	subject := "验证码 - 企业注册"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:Arial,sans-serif;background:#f5f7fb;padding:40px;">
<div style="max-width:480px;margin:0 auto;background:#fff;border-radius:8px;padding:32px;box-shadow:0 4px 16px rgba(0,0,0,.08);">
  <h2 style="color:#1f2937;margin:0 0 8px;">企业注册验证码</h2>
  <p style="color:#6b7280;margin:0 0 24px;">您正在注册企业账户，请输入以下验证码：</p>
  <div style="background:#f0f5ff;border:1px solid #d6e4ff;border-radius:6px;padding:20px;text-align:center;margin-bottom:24px;">
    <span style="font-size:28px;font-weight:800;color:#165dff;letter-spacing:8px;">%s</span>
  </div>
  <p style="color:#9ca3af;font-size:13px;margin:0;">5 分钟内有效，请勿泄露。</p>
</div>
</body>
</html>`, code)
	return s.SendMail(to, subject, body)
}

func buildMessage(from, to, subject, body string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}
