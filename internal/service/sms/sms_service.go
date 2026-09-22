package sms

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/config"
)

// 阿里云 dysmsapi 端点
const dysmsapiEndpoint = "https://dysmsapi.aliyuncs.com/"
const verifyCodeTTL = 5 * time.Minute // 验证码有效期

// Service 短信服务接口
type Service interface {
	// SendVerifyCode 发送验证码短信（存入Redis，5分钟有效）
	SendVerifyCode(ctx context.Context, phone string) error
	// VerifyCode 校验验证码（校验后立即作废）
	VerifyCode(ctx context.Context, phone, code string) (bool, error)
}

type service struct {
	cfg         config.SMSConfig
	redisClient *redis.Client
	client      *http.Client
}

// NewService 创建短信服务
func NewService(cfg config.SMSConfig, redisClient *redis.Client) Service {
	return &service{
		cfg:         cfg,
		redisClient: redisClient,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

// SendVerifyCode 生成6位验证码并发送
func (s *service) SendVerifyCode(ctx context.Context, phone string) error {
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	if !s.cfg.Enabled {
		return errors.New("短信服务未启用")
	}
	if s.cfg.SignName == "" || s.cfg.TemplateCode == "" {
		return errors.New("短信签名或模板未配置")
	}

	code := generateCode(6)

	paramJSON, _ := json.Marshal(map[string]string{"code": code})

	respBody, err := s.request(ctx, map[string]string{
		"Action":        "SendSms",
		"PhoneNumbers":  phone,
		"SignName":      s.cfg.SignName,
		"TemplateCode":  s.cfg.TemplateCode,
		"TemplateParam": string(paramJSON),
	})
	if err != nil {
		return err
	}

	// 解析响应
	var resp struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("解析短信响应失败: %w", err)
	}
	if resp.Code != "OK" {
		return fmt.Errorf("短信发送失败: [%s] %s", resp.Code, resp.Message)
	}

	// 存储验证码到Redis
	key := fmt.Sprintf("sms:verify:%s", phone)
	if err := s.redisClient.Set(ctx, key, code, verifyCodeTTL).Err(); err != nil {
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	return nil
}

// VerifyCode 校验验证码
func (s *service) VerifyCode(ctx context.Context, phone, code string) (bool, error) {
	// 通用验证码：配置了且匹配则直接放行（测试用，可绕过短信）
	if s.cfg.VerifyCode != "" && code == s.cfg.VerifyCode {
		return true, nil
	}
	key := fmt.Sprintf("sms:verify:%s", phone)
	storedCode, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // 验证码不存在或已过期
		}
		return false, fmt.Errorf("读取验证码失败: %w", err)
	}

	if storedCode != code {
		return false, nil
	}

	// 校验成功，作废验证码
	s.redisClient.Del(ctx, key)
	return true, nil
}

// request 调用阿里云 OpenAPI（V1 签名）
func (s *service) request(ctx context.Context, params map[string]string) ([]byte, error) {
	query := s.signParams(params)

	u := dysmsapiEndpoint + "?" + query
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用阿里云短信失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// signParams 生成带签名的Query参数
func (s *service) signParams(extra map[string]string) string {
	params := map[string]string{
		"AccessKeyId":      s.cfg.AccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"RegionId":         "cn-hangzhou",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d-%d", time.Now().UnixNano(), randInt()),
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}
	for k, v := range extra {
		params[k] = v
	}

	// 排序并拼接规范查询串（键值各自编码，= 和 & 为字面字符）
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canonical strings.Builder
	for i, k := range keys {
		if i > 0 {
			canonical.WriteString("&")
		}
		canonical.WriteString(percentEncode(k))
		canonical.WriteString("=")
		canonical.WriteString(percentEncode(params[k]))
	}

	// StringToSign = Method&%2F& + percentEncode(整个规范串)
	stringToSign := "GET&%2F&" + percentEncode(canonical.String())

	signature := s.hmacSHA1(stringToSign)
	params["Signature"] = signature

	// 重新编码全部参数（含 Signature）拼URL
	parts := make([]string, 0, len(params))
	finalKeys := make([]string, 0, len(params))
	for k := range params {
		finalKeys = append(finalKeys, k)
	}
	sort.Strings(finalKeys)
	for _, k := range finalKeys {
		parts = append(parts, percentEncode(k)+"="+percentEncode(params[k]))
	}
	return strings.Join(parts, "&")
}

func (s *service) hmacSHA1(value string) string {
	mac := hmac.New(sha1.New, []byte(s.cfg.AccessSecret+"&"))
	mac.Write([]byte(value))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// percentEncode 阿里云签名要求的特殊URL编码
func percentEncode(s string) string {
	// 先做标准URL编码（保留RFC3986规则外）
	escaped := url.QueryEscape(s)
	// 阿里云要求 + -> %20, * -> %2A, %7E -> ~
	escaped = strings.ReplaceAll(escaped, "+", "%20")
	escaped = strings.ReplaceAll(escaped, "*", "%2A")
	escaped = strings.ReplaceAll(escaped, "%7E", "~")
	return escaped
}

func randInt() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(9999999))
	if err != nil {
		return time.Now().UnixNano() % 9999999
	}
	return n.Int64()
}

// generateCode 生成随机数字验证码
func generateCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		b[i] = '0' + byte(n.Int64())
	}
	return string(b)
}
