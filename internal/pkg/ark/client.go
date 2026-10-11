// Package ark 火山方舟大模型（OpenAI 兼容的 chat/completions 接口）
package ark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL 火山方舟在线推理地址：固定使用北京区域（接入点须在北京区域创建）
const DefaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"

// Message 对话消息
type Message struct {
	Role    string `json:"role"` // system / user / assistant
	Content string `json:"content"`
}

// Client 火山方舟客户端
type Client struct {
	BaseURL    string
	APIKey     string
	Model      string // 推理接入点 ID（ep-…）或模型名
	HTTPClient *http.Client
}

// NewClient 创建客户端
func NewClient(apiKey, model string) *Client {
	return &Client{BaseURL: DefaultBaseURL, APIKey: apiKey, Model: model, HTTPClient: &http.Client{Timeout: 120 * time.Second}}
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Error 方舟接口错误
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("火山方舟返回错误（HTTP %d %s）：%s", e.Status, e.Code, e.Message)
}

// Chat 发送对话，返回模型回复内容
func (c *Client) Chat(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error) {
	body, _ := json.Marshal(chatRequest{Model: c.Model, Messages: messages, Temperature: temperature, MaxTokens: maxTokens})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用火山方舟失败：%w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	var out chatResponse
	_ = json.Unmarshal(raw, &out)
	if resp.StatusCode != http.StatusOK || out.Error != nil {
		e := &Error{Status: resp.StatusCode}
		if out.Error != nil {
			e.Code, e.Message = out.Error.Code, out.Error.Message
		} else {
			e.Message = strings.TrimSpace(string(raw))
			if len(e.Message) > 300 {
				e.Message = e.Message[:300]
			}
		}
		return "", e
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("火山方舟未返回结果")
	}
	return out.Choices[0].Message.Content, nil
}
