package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
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
		// 公司明确选择了云存储：报错提示补全配置，不能悄悄写到服务器本地
		if normalizeConfig(config).StorageType != StorageTypeLocal {
			return nil, fmt.Errorf("公司云存储配置不完整：%w", err)
		}
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

// 公司存储类型
const (
	StorageTypeLocal      = "local"
	StorageTypeAliyunOSS  = "aliyunOSS"
	StorageTypeTencentCOS = "tencentCOS"
)

// cosBucketURLPattern 腾讯云 COS Bucket 地址：https://{bucket}.cos.{region}.myqcloud.com
var cosBucketURLPattern = regexp.MustCompile(`^(?:https?://)?([^./]+)\.cos\.([a-z0-9-]+)\.myqcloud\.com`)

// ParseCOSBucketURL 从 COS Bucket 地址解析 Bucket 名称与区域；无法识别时返回空串
func ParseCOSBucketURL(bucketURL string) (bucket, region string) {
	m := cosBucketURLPattern.FindStringSubmatch(strings.TrimSpace(bucketURL))
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}

// normalizeConfig 规范化公司存储配置（返回副本，不修改原对象）：
//   - 存储类型别名：cos → tencentCOS，oss → aliyunOSS
//   - 类型为空但腾讯云 COS 的 Bucket 地址已填写时按 COS 处理
//   - COS 的 Bucket 名称、区域为空时从 Bucket 地址解析
//   - COS 密钥为空时沿用腾讯云直播的密钥（同一腾讯云账号无需重复填写）
//   - 其余为空时按本地存储
func normalizeConfig(config *entity.SysCompanyConf) *entity.SysCompanyConf {
	c := *config
	switch strings.ToLower(strings.TrimSpace(c.StorageType)) {
	case "local":
		c.StorageType = StorageTypeLocal
	case "cos", "tencentcos":
		c.StorageType = StorageTypeTencentCOS
	case "oss", "aliyunoss":
		c.StorageType = StorageTypeAliyunOSS
	case "":
		if strings.TrimSpace(c.TencentCOSBucketURL) != "" {
			c.StorageType = StorageTypeTencentCOS
		} else {
			c.StorageType = StorageTypeLocal
		}
	}
	if c.StorageType == StorageTypeTencentCOS {
		bucket, region := ParseCOSBucketURL(c.TencentCOSBucketURL)
		if c.TencentCOSBucketName == "" {
			c.TencentCOSBucketName = bucket
		}
		if c.TencentCOSRegion == "" {
			c.TencentCOSRegion = region
		}
		if c.TencentCOSSecretID == "" && c.TencentCOSSecretKey == "" {
			c.TencentCOSSecretID = c.TencentCloudSecretID
			c.TencentCOSSecretKey = c.TencentCloudSecretKey
		}
	}
	return &c
}

// createStorageFromConfig 根据数据库配置创建存储实例
func (m *CompanyStorageManager) createStorageFromConfig(config *entity.SysCompanyConf) (Storage, error) {
	c := normalizeConfig(config)
	switch c.StorageType {
	case StorageTypeLocal:
		return m.createLocalStorage(c)
	case StorageTypeAliyunOSS:
		return m.createAliyunOSSStorage(c)
	case StorageTypeTencentCOS:
		return m.createTencentCOSStorage(c)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", config.StorageType)
	}
}

// GetStorageByType 读写已有文件时按文件记录的存储类型取存储：
// 记录为本地存储、而公司已改用云存储时返回本地存储（使用公司的本地存储配置），保证历史文件仍可访问；
// 其余情况返回公司当前存储
func (m *CompanyStorageManager) GetStorageByType(ctx context.Context, companyID uint, storageType string) (Storage, error) {
	if storageType != StorageTypeLocal {
		return m.GetStorage(ctx, companyID)
	}
	config, err := m.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil || config == nil {
		return m.globalManager.GetStorage(StorageTypeLocal)
	}
	if normalizeConfig(config).StorageType == StorageTypeLocal {
		return m.GetStorage(ctx, companyID)
	}
	return m.createLocalStorage(config)
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
		return nil, fmt.Errorf("云存储 BucketURL 未配置")
	}
	if config.TencentCOSSecretID == "" {
		return nil, fmt.Errorf("云存储 SecretID 未配置")
	}
	if config.TencentCOSSecretKey == "" {
		return nil, fmt.Errorf("云存储 SecretKey 未配置")
	}
	if config.TencentCOSBucketName == "" {
		return nil, fmt.Errorf("云存储 BucketName 未配置")
	}
	if config.TencentCOSRegion == "" {
		return nil, fmt.Errorf("云存储 Region 未配置")
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
