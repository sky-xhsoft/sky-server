package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"go.uber.org/zap"
)

// CompanyStorageManager 公司级存储管理器
type CompanyStorageManager struct {
	globalManager    *StorageManager                     // 全局存储管理器（作为fallback）
	configRepo       repository.SysCompanyConfRepository // 存储配置仓储
	companyStorages  map[uint]*companyStorageEntry       // 公司ID -> 存储实例缓存
	mu               sync.RWMutex                        // 读写锁
	defaultCompanyID uint                                // 默认公司ID
}

// companyStorageEntry 公司存储缓存条目
type companyStorageEntry struct {
	storage  Storage                // 存储实例
	config   *entity.SysCompanyConf // 配置信息
	cachedAt time.Time              // 缓存时间
}

const (
	// ConfigCacheTTL 配置缓存TTL（秒）
	ConfigCacheTTL = 300 // 5分钟缓存
)

// NewCompanyStorageManager 创建公司级存储管理器
func NewCompanyStorageManager(
	globalConfig *config.StorageConfig,
	configRepo repository.SysCompanyConfRepository,
	defaultCompanyID uint,
) (*CompanyStorageManager, error) {

	// 创建全局存储管理器作为fallback
	globalManager, err := NewStorageManager(globalConfig)
	if err != nil {
		logger.Error("创建全局存储管理器失败", zap.Error(err))
		return nil, fmt.Errorf("创建全局存储管理器失败: %w", err)
	}

	manager := &CompanyStorageManager{
		globalManager:    globalManager,
		configRepo:       configRepo,
		companyStorages:  make(map[uint]*companyStorageEntry),
		defaultCompanyID: defaultCompanyID,
	}

	return manager, nil
}

// GetStorage 获取指定公司的存储实例
// 优先级：公司配置 > 配置文件 > 默认值
func (m *CompanyStorageManager) GetStorage(ctx context.Context, companyID uint) (Storage, error) {
	// 首先尝试从缓存获取
	m.mu.RLock()
	if entry, exists := m.companyStorages[companyID]; exists {
		// 检查缓存是否过期
		if time.Since(entry.cachedAt) < ConfigCacheTTL*time.Second {
			m.mu.RUnlock()
			return entry.storage, nil
		}
		// 缓存过期，需要重新加载
	}
	m.mu.RUnlock()

	// 从数据库读取配置
	config, err := m.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		logger.Error("读取公司存储配置失败",
			zap.Uint("companyID", companyID),
			zap.Error(err))
		// 数据库读取失败，返回全局存储
		return m.globalManager.GetDefaultStorage()
	}

	// 如果数据库中没有配置，返回全局存储
	if config == nil {
		logger.Info("公司未配置存储，使用全局存储",
			zap.Uint("companyID", companyID))
		return m.globalManager.GetDefaultStorage()
	}

	// 根据配置创建存储实例
	storage, err := m.createStorageFromConfig(config)
	if err != nil {
		logger.Error("根据公司配置创建存储实例失败",
			zap.Uint("companyID", companyID),
			zap.Error(err))
		// 创建失败，返回全局存储
		return m.globalManager.GetDefaultStorage()
	}

	// 更新缓存
	m.mu.Lock()
	m.companyStorages[companyID] = &companyStorageEntry{
		storage:  storage,
		config:   config,
		cachedAt: time.Now(),
	}
	m.mu.Unlock()

	logger.Info("使用公司配置的存储",
		zap.Uint("companyID", companyID),
		zap.String("storageType", config.StorageType))

	return storage, nil
}

// GetDefaultStorage 获取默认存储（全局存储）
func (m *CompanyStorageManager) GetDefaultStorage() (Storage, error) {
	return m.globalManager.GetDefaultStorage()
}

// GetGlobalManager 获取全局存储管理器
func (m *CompanyStorageManager) GetGlobalManager() *StorageManager {
	return m.globalManager
}

// RefreshCompanyConfig 刷新公司存储配置缓存
func (m *CompanyStorageManager) RefreshCompanyConfig(ctx context.Context, companyID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 删除缓存
	delete(m.companyStorages, companyID)

	// 如果有上下文，尝试重新加载
	if ctx != nil {
		config, err := m.configRepo.GetByCompanyID(ctx, companyID)
		if err != nil {
			logger.Error("刷新公司存储配置失败",
				zap.Uint("companyID", companyID),
				zap.Error(err))
			return err
		}

		if config != nil {
			storage, err := m.createStorageFromConfig(config)
			if err != nil {
				logger.Error("刷新后创建存储实例失败",
					zap.Uint("companyID", companyID),
					zap.Error(err))
				return err
			}

			m.companyStorages[companyID] = &companyStorageEntry{
				storage:  storage,
				config:   config,
				cachedAt: time.Now(),
			}

			logger.Info("刷新公司存储配置成功",
				zap.Uint("companyID", companyID),
				zap.String("storageType", config.StorageType))
		}
	}

	return nil
}

// ClearAllCache 清除所有缓存
func (m *CompanyStorageManager) ClearAllCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.companyStorages = make(map[uint]*companyStorageEntry)
	logger.Info("清除所有公司存储配置缓存")
}

// createStorageFromConfig 根据数据库配置创建存储实例
func (m *CompanyStorageManager) createStorageFromConfig(config *entity.SysCompanyConf) (Storage, error) {
	switch config.StorageType {
	case "local":
		return m.createLocalStorage(config)
	case "aliyunOSS":
		return m.createAliyunOSSStorage(config)
	case "tencentCOS":
		return m.createTencentCOSStorage(config)
	default:
		// 未知类型，使用默认的本地存储
		logger.Warn("未知的存储类型，使用本地存储",
			zap.String("storageType", config.StorageType))
		return m.createLocalStorage(config)
	}
}

// createLocalStorage 创建本地存储
func (m *CompanyStorageManager) createLocalStorage(config *entity.SysCompanyConf) (Storage, error) {
	basePath := config.LocalBasePath
	if basePath == "" {
		basePath = "uploads"
	}

	baseURL := config.LocalBaseURL
	if baseURL == "" {
		baseURL = "/files"
	}

	return NewLocalStorage(&LocalStorageConfig{
		BasePath: basePath,
		BaseURL:  baseURL,
	})
}

// createAliyunOSSStorage 创建阿里云OSS存储
func (m *CompanyStorageManager) createAliyunOSSStorage(config *entity.SysCompanyConf) (Storage, error) {
	// 验证必要的配置
	if config.AliyunOSSEndpoint == "" {
		return nil, fmt.Errorf("阿里云OSS Endpoint未配置")
	}
	if config.AliyunOSSAccessKeyID == "" {
		return nil, fmt.Errorf("阿里云OSS AccessKeyID未配置")
	}
	if config.AliyunOSSAccessKeySecret == "" {
		return nil, fmt.Errorf("阿里云OSS AccessKeySecret未配置")
	}
	if config.AliyunOSSBucketName == "" {
		return nil, fmt.Errorf("阿里云OSS BucketName未配置")
	}

	return NewAliyunOSS(&AliyunOSSConfig{
		Endpoint:        config.AliyunOSSEndpoint,
		AccessKeyID:     config.AliyunOSSAccessKeyID,
		AccessKeySecret: config.AliyunOSSAccessKeySecret,
		BucketName:      config.AliyunOSSBucketName,
		CDNDomain:       config.AliyunOSSCdnDomain,
	})
}

// createTencentCOSStorage 创建腾讯云COS存储
func (m *CompanyStorageManager) createTencentCOSStorage(config *entity.SysCompanyConf) (Storage, error) {
	// 验证必要的配置
	if config.TencentCOSBucketURL == "" {
		return nil, fmt.Errorf("腾讯云COS BucketURL未配置")
	}
	if config.TencentCOSSecretID == "" {
		return nil, fmt.Errorf("腾讯云COS SecretID未配置")
	}
	if config.TencentCOSSecretKey == "" {
		return nil, fmt.Errorf("腾讯云COS SecretKey未配置")
	}
	if config.TencentCOSBucketName == "" {
		return nil, fmt.Errorf("腾讯云COS BucketName未配置")
	}
	if config.TencentCOSRegion == "" {
		return nil, fmt.Errorf("腾讯云COS Region未配置")
	}

	return NewTencentCOS(&TencentCOSConfig{
		BucketURL:  config.TencentCOSBucketURL,
		SecretID:   config.TencentCOSSecretID,
		SecretKey:  config.TencentCOSSecretKey,
		BucketName: config.TencentCOSBucketName,
		Region:     config.TencentCOSRegion,
		CDNDomain:  config.TencentCOSCdnDomain,
	})
}

// Upload 使用公司配置上传文件
func (m *CompanyStorageManager) Upload(ctx context.Context, companyID uint, path string, reader io.Reader, contentType string) (string, error) {
	storage, err := m.GetStorage(ctx, companyID)
	if err != nil {
		return "", err
	}

	return storage.Upload(ctx, path, reader, contentType)
}

// GetConfig 获取公司的存储配置（带缓存）
func (m *CompanyStorageManager) GetConfig(ctx context.Context, companyID uint) (*entity.SysCompanyConf, error) {
	// 先检查缓存
	m.mu.RLock()
	if entry, exists := m.companyStorages[companyID]; exists {
		if entry.config != nil && time.Since(entry.cachedAt) < ConfigCacheTTL*time.Second {
			m.mu.RUnlock()
			return entry.config, nil
		}
	}
	m.mu.RUnlock()

	// 从数据库读取
	config, err := m.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// HasCompanyConfig 检查公司是否有配置
func (m *CompanyStorageManager) HasCompanyConfig(ctx context.Context, companyID uint) bool {
	config, err := m.GetConfig(ctx, companyID)
	if err != nil {
		return false
	}
	return config != nil
}
