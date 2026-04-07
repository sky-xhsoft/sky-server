package tencent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"go.uber.org/zap"
)

// CompanyTencentManager 公司级腾讯云客户端管理器
type CompanyTencentManager struct {
	globalClient     *Client                             // 全局客户端（作为fallback）
	configRepo       repository.SysCompanyConfRepository // 公司配置仓储
	companyClients   map[uint]*companyTencentEntry       // 公司ID -> 腾讯云客户端缓存
	mu               sync.RWMutex                        // 读写锁
	defaultCompanyID uint                                // 默认公司ID
}

// companyTencentEntry 公司腾讯云客户端缓存条目
type companyTencentEntry struct {
	client   *Client                // 腾讯云客户端
	config   *entity.SysCompanyConf // 配置信息
	cachedAt time.Time              // 缓存时间
}

const (
	// ConfigCacheTTL 配置缓存TTL（秒）
	TencentConfigCacheTTL = 300 // 5分钟缓存
)

// NewCompanyTencentManager 创建公司级腾讯云客户端管理器
func NewCompanyTencentManager(
	globalConfig *config.TencentCloudConfig,
	configRepo repository.SysCompanyConfRepository,
	defaultCompanyID uint,
) (*CompanyTencentManager, error) {

	// 创建全局客户端作为fallback
	globalClient, err := NewClient(globalConfig)
	if err != nil {
		logger.Error("创建全局腾讯云客户端失败", zap.Error(err))
		return nil, fmt.Errorf("创建全局腾讯云客户端失败: %w", err)
	}

	manager := &CompanyTencentManager{
		globalClient:     globalClient,
		configRepo:       configRepo,
		companyClients:   make(map[uint]*companyTencentEntry),
		defaultCompanyID: defaultCompanyID,
	}

	return manager, nil
}

// GetClient 获取指定公司的腾讯云客户端
// 优先级：公司配置 > 配置文件 > 默认值
func (m *CompanyTencentManager) GetClient(ctx context.Context, companyID uint) (*Client, error) {
	// 首先尝试从缓存获取
	m.mu.RLock()
	if entry, exists := m.companyClients[companyID]; exists {
		// 检查缓存是否过期
		if time.Since(entry.cachedAt) < TencentConfigCacheTTL*time.Second {
			m.mu.RUnlock()
			return entry.client, nil
		}
		// 缓存过期，需要重新加载
	}
	m.mu.RUnlock()

	// 从数据库读取配置
	config, err := m.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		logger.Error("读取公司腾讯云配置失败",
			zap.Uint("companyID", companyID),
			zap.Error(err))
		// 数据库读取失败，返回全局客户端
		return m.globalClient, nil
	}

	// 如果数据库中没有配置，返回全局客户端
	if config == nil {
		logger.Info("公司未配置腾讯云，使用全局配置",
			zap.Uint("companyID", companyID))
		return m.globalClient, nil
	}

	// 如果公司配置中的腾讯云参数不完整，返回全局客户端
	if config.TencentCloudSecretID == "" || config.TencentCloudSecretKey == "" {
		logger.Info("公司腾讯云配置不完整，使用全局配置",
			zap.Uint("companyID", companyID))
		return m.globalClient, nil
	}

	// 根据配置创建腾讯云客户端
	client, err := m.createClientFromConfig(config)
	if err != nil {
		logger.Error("根据公司配置创建腾讯云客户端失败",
			zap.Uint("companyID", companyID),
			zap.Error(err))
		// 创建失败，返回全局客户端
		return m.globalClient, nil
	}

	// 更新缓存
	m.mu.Lock()
	m.companyClients[companyID] = &companyTencentEntry{
		client:   client,
		config:   config,
		cachedAt: time.Now(),
	}
	m.mu.Unlock()

	logger.Info("使用公司配置的腾讯云",
		zap.Uint("companyID", companyID))

	return client, nil
}

// GetDefaultClient 获取默认客户端（全局客户端）
func (m *CompanyTencentManager) GetDefaultClient() *Client {
	return m.globalClient
}

// GetGlobalClient 获取全局客户端
func (m *CompanyTencentManager) GetGlobalClient() *Client {
	return m.globalClient
}

// RefreshCompanyConfig 刷新公司腾讯云配置缓存
func (m *CompanyTencentManager) RefreshCompanyConfig(ctx context.Context, companyID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 删除缓存
	delete(m.companyClients, companyID)

	// 如果有上下文，尝试重新加载
	if ctx != nil {
		config, err := m.configRepo.GetByCompanyID(ctx, companyID)
		if err != nil {
			logger.Error("刷新公司腾讯云配置失败",
				zap.Uint("companyID", companyID),
				zap.Error(err))
			return err
		}

		if config != nil && config.TencentCloudSecretID != "" && config.TencentCloudSecretKey != "" {
			client, err := m.createClientFromConfig(config)
			if err != nil {
				logger.Error("刷新后创建腾讯云客户端失败",
					zap.Uint("companyID", companyID),
					zap.Error(err))
				return err
			}

			m.companyClients[companyID] = &companyTencentEntry{
				client:   client,
				config:   config,
				cachedAt: time.Now(),
			}

			logger.Info("刷新公司腾讯云配置成功",
				zap.Uint("companyID", companyID))
		}
	}

	return nil
}

// ClearAllCache 清除所有缓存
func (m *CompanyTencentManager) ClearAllCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.companyClients = make(map[uint]*companyTencentEntry)
	logger.Info("清除所有公司腾讯云配置缓存")
}

// createClientFromConfig 根据数据库配置创建腾讯云客户端
func (m *CompanyTencentManager) createClientFromConfig(sysConf *entity.SysCompanyConf) (*Client, error) {
	// 使用公司配置中的值，如果没有则使用全局配置的默认值
	globalCfg := m.globalClient.GetConfig()

	region := sysConf.TencentCloudRegion
	if region == "" {
		region = globalCfg.Region
	}

	tencentCfg := &config.TencentCloudConfig{
		SecretID:    sysConf.TencentCloudSecretID,
		SecretKey:   sysConf.TencentCloudSecretKey,
		Region:      region,
		CallbackKey: sysConf.TencentCloudCallbackKey,
		Live:        globalCfg.Live, // 使用全局的直播配置
	}

	return NewClient(tencentCfg)
}

// HasCompanyConfig 检查公司是否有配置
func (m *CompanyTencentManager) HasCompanyConfig(ctx context.Context, companyID uint) bool {
	config, err := m.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return false
	}
	if config == nil {
		return false
	}
	return config.TencentCloudSecretID != "" && config.TencentCloudSecretKey != ""
}
