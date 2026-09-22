package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/queue"
	"go.uber.org/zap"
)

// Service 邮件服务接口
type Service interface {
	// SendVerificationCode 发送验证码邮件（同步）
	SendVerificationCode(to, code string) error
	// SendMail 同步发送邮件
	SendMail(to, subject, body string) error
	// SendMailAsync 异步发送邮件（推入消息队列）
	SendMailAsync(ctx context.Context, to, subject, body string) error
}

// TaskType 邮件任务类型常量
const TaskTypeSendEmail = "email.send"

type service struct {
	cfg   *config.EmailConfig
	queue queue.Queue // 可选：用于异步发送
}

// NewService 创建邮件服务（同步模式）
func NewService(cfg *config.EmailConfig) Service {
	return &service{cfg: cfg}
}

// NewServiceWithQueue 创建支持异步发送的邮件服务
// 同时会注册一个 email.send 任务消费者
func NewServiceWithQueue(cfg *config.EmailConfig, q queue.Queue) Service {
	s := &service{cfg: cfg, queue: q}

	// 注册邮件发送任务处理器
	if q != nil {
		q.Subscribe(TaskTypeSendEmail, func(ctx context.Context, task *queue.Task) error {
			to, _ := task.Payload["to"].(string)
			subject, _ := task.Payload["subject"].(string)
			body, _ := task.Payload["body"].(string)

			if to == "" || subject == "" {
				return fmt.Errorf("invalid email task: missing to or subject")
			}

			if err := s.SendMail(to, subject, body); err != nil {
				logger.Error("Async email send failed",
					zap.String("to", to),
					zap.String("subject", subject),
					zap.Error(err))
				return err
			}

			logger.Info("Async email sent",
				zap.String("to", to),
				zap.String("subject", subject))
			return nil
		})
	}

	return s
}

// SendMail 同步发送邮件
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

// SendMailAsync 异步发送邮件
// 如果队列不可用，会降级为同步发送
func (s *service) SendMailAsync(ctx context.Context, to, subject, body string) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("邮件服务未启用")
	}

	// 队列不可用时降级为同步
	if s.queue == nil {
		go func() {
			if err := s.SendMail(to, subject, body); err != nil {
				logger.Error("Fallback sync email failed", zap.Error(err))
			}
		}()
		return nil
	}

	return s.queue.Publish(ctx, TaskTypeSendEmail, map[string]interface{}{
		"to":      to,
		"subject": subject,
		"body":    body,
	})
}

// SendVerificationCode 发送验证码
func (s *service) SendVerificationCode(to, code string) error {
	subject := "验证码 - 企业注册"
	body := buildVerificationCodeHTML(code)
	return s.SendMail(to, subject, body)
}

// buildVerificationCodeHTML 构建验证码邮件 HTML 模板
func buildVerificationCodeHTML(code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
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
