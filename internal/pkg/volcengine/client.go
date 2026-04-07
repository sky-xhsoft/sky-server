package volcengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

// Client 火山引擎API客户端
type Client struct {
	config *config.VolcengineConfig
	client *http.Client
	signer *Signer
}

// NewClient 创建火山引擎API客户端
func NewClient(cfg *config.VolcengineConfig) (*Client, error) {
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建签名器
	signer := NewSigner(
		cfg.AccessKeyId,
		cfg.AccessKeySecret,
		cfg.Region,
		cfg.Service,
	)

	return &Client{
		config: cfg,
		client: httpClient,
		signer: signer,
	}, nil
}

// DoRequest 发送请求并返回响应
func (c *Client) DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	// 获取请求体
	var body []byte
	if req.Body != nil {
		var err error
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	// 签名请求
	if err := c.signer.SignRequest(req, body); err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DoAPIRequest 发送API请求并解析响应
func (c *Client) DoAPIRequest(ctx context.Context, method, url string, requestBody interface{}, response interface{}) error {
	// 创建请求体
	var body []byte
	if requestBody != nil {
		var err error
		body, err = json.Marshal(requestBody)
		if err != nil {
			return err
		}
	}

	// 创建HTTP请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// 设置上下文
	req = req.WithContext(ctx)

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.DoRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	if response != nil {
		if err := json.Unmarshal(respBody, response); err != nil {
			return err
		}
	}

	return nil
}

// GetRequest 创建GET请求
func (c *Client) GetRequest(ctx context.Context, url string, response interface{}) error {
	return c.DoAPIRequest(ctx, http.MethodGet, url, nil, response)
}

// PostRequest 创建POST请求
func (c *Client) PostRequest(ctx context.Context, url string, requestBody interface{}, response interface{}) error {
	return c.DoAPIRequest(ctx, http.MethodPost, url, requestBody, response)
}

// PutRequest 创建PUT请求
func (c *Client) PutRequest(ctx context.Context, url string, requestBody interface{}, response interface{}) error {
	return c.DoAPIRequest(ctx, http.MethodPut, url, requestBody, response)
}

// DeleteRequest 创建DELETE请求
func (c *Client) DeleteRequest(ctx context.Context, url string, response interface{}) error {
	return c.DoAPIRequest(ctx, http.MethodDelete, url, nil, response)
}

// GetConfig 获取配置
func (c *Client) GetConfig() *config.VolcengineConfig {
	return c.config
}
