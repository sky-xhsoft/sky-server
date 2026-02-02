package tencent

import (
	"fmt"
	"sync"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
)

// Client 腾讯云客户端
type Client struct {
	config      *config.TencentCloudConfig
	credential  *common.Credential
	liveClient  *live.Client
	liveClients map[string]*live.Client // 多地域客户端缓存
	mu          sync.RWMutex
}

// NewClient 创建腾讯云客户端
func NewClient(cfg *config.TencentCloudConfig) (*Client, error) {
	credential := common.NewCredential(
		cfg.SecretID,
		cfg.SecretKey,
	)

	// 创建默认地域的直播客户端
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "live.tencentcloudapi.com"

	liveClient, err := live.NewClient(credential, cfg.Region, cpf)
	if err != nil {
		return nil, err
	}

	return &Client{
		config:      cfg,
		credential:  credential,
		liveClient:  liveClient,
		liveClients: make(map[string]*live.Client),
	}, nil
}

// GetLiveClient 获取直播客户端（默认地域）
func (c *Client) GetLiveClient() *live.Client {
	return c.liveClient
}

// GetLiveClientForRegion 获取指定地域的直播客户端
func (c *Client) GetLiveClientForRegion(region string) (*live.Client, error) {
	// 如果是默认地域，直接返回默认客户端
	if region == "" || region == c.config.Region {
		return c.liveClient, nil
	}

	// 检查缓存
	c.mu.RLock()
	client, exists := c.liveClients[region]
	c.mu.RUnlock()

	if exists {
		return client, nil
	}

	// 创建新的地域客户端
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if client, exists := c.liveClients[region]; exists {
		return client, nil
	}

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "live.tencentcloudapi.com"

	newClient, err := live.NewClient(c.credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for region %s: %w", region, err)
	}

	c.liveClients[region] = newClient
	return newClient, nil
}

// GetConfig 获取配置
func (c *Client) GetConfig() *config.TencentCloudConfig {
	return c.config
}
