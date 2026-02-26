package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/sky-xhsoft/sky-server/internal/config"
)

// StorageManager 存储管理器
type StorageManager struct {
	storages       map[string]Storage // 存储后端映射
	defaultStorage string             // 默认存储类型
}

// NewStorageManager 创建存储管理器
func NewStorageManager(cfg *config.StorageConfig) (*StorageManager, error) {
	manager := &StorageManager{
		storages:       make(map[string]Storage),
		defaultStorage: cfg.Default,
	}

	// 初始化本地存储
	if err := manager.initLocalStorage(&cfg.Local); err != nil {
		return nil, fmt.Errorf("初始化本地存储失败: %w", err)
	}

	// 初始化阿里云OSS存储
	if err := manager.initAliyunOSS(&cfg.AliyunOSS); err != nil {
		return nil, fmt.Errorf("初始化阿里云OSS存储失败: %w", err)
	}

	// 初始化腾讯云COS存储
	if err := manager.initTencentCOS(&cfg.TencentCOS); err != nil {
		return nil, fmt.Errorf("初始化腾讯云COS存储失败: %w", err)
	}

	// 验证默认存储类型
	if _, ok := manager.storages[manager.defaultStorage]; !ok {
		// 如果默认存储类型未初始化，尝试使用已配置的存储类型
		if len(manager.storages) > 0 {
			// 优先使用本地存储
			if _, exists := manager.storages["local"]; exists {
				manager.defaultStorage = "local"
			} else {
				// 如果没有本地存储，使用第一个可用的存储类型
				for storageType := range manager.storages {
					manager.defaultStorage = storageType
					break
				}
			}
		} else {
			// 如果没有任何存储类型配置，默认初始化本地存储（使用临时目录）
			storage, err := NewLocalStorage(&LocalStorageConfig{
				BasePath: "./uploads", // 默认临时目录
				BaseURL:  "/uploads",
			})
			if err != nil {
				return nil, fmt.Errorf("无法初始化任何存储类型: %w", err)
			}
			manager.storages["local"] = storage
			manager.defaultStorage = "local"
		}
	}

	return manager, nil
}

// initLocalStorage 初始化本地存储
func (m *StorageManager) initLocalStorage(cfg *config.LocalStorageConfig) error {
	if cfg.BasePath == "" {
		return nil // 未配置，跳过初始化
	}

	storage, err := NewLocalStorage(&LocalStorageConfig{
		BasePath: cfg.BasePath,
		BaseURL:  cfg.BaseURL,
	})
	if err != nil {
		return err
	}

	m.storages["local"] = storage
	return nil
}

// initAliyunOSS 初始化阿里云OSS存储
func (m *StorageManager) initAliyunOSS(cfg *config.AliyunOSSConfig) error {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
		return nil // 未配置，跳过初始化
	}

	storage, err := NewAliyunOSS(&AliyunOSSConfig{
		Endpoint:        cfg.Endpoint,
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		BucketName:      cfg.BucketName,
		CDNDomain:       cfg.CDNDomain,
	})
	if err != nil {
		return err
	}

	m.storages["aliyunOSS"] = storage
	return nil
}

// initTencentCOS 初始化腾讯云COS存储
func (m *StorageManager) initTencentCOS(cfg *config.TencentCOSConfig) error {
	if cfg.BucketURL == "" || cfg.SecretID == "" || cfg.SecretKey == "" || cfg.BucketName == "" || cfg.Region == "" {
		return nil // 未配置，跳过初始化
	}

	storage, err := NewTencentCOS(&TencentCOSConfig{
		BucketURL:  cfg.BucketURL,
		SecretID:   cfg.SecretID,
		SecretKey:  cfg.SecretKey,
		BucketName: cfg.BucketName,
		Region:     cfg.Region,
		CDNDomain:  cfg.CDNDomain,
	})
	if err != nil {
		return err
	}

	m.storages["tencentCOS"] = storage
	return nil
}

// GetStorage 获取指定类型的存储后端
func (m *StorageManager) GetStorage(storageType string) (Storage, error) {
	if storageType == "" {
		storageType = m.defaultStorage
	}

	storage, ok := m.storages[storageType]
	if !ok {
		return nil, fmt.Errorf("存储类型 %s 不存在", storageType)
	}

	return storage, nil
}

// GetDefaultStorage 获取默认存储后端
func (m *StorageManager) GetDefaultStorage() (Storage, error) {
	return m.GetStorage(m.defaultStorage)
}

// Upload 上传文件（使用默认存储）
func (m *StorageManager) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return "", err
	}
	return storage.Upload(ctx, path, reader, contentType)
}

// Download 下载文件（使用默认存储）
func (m *StorageManager) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return nil, err
	}
	return storage.Download(ctx, path)
}

// Delete 删除文件（使用默认存储）
func (m *StorageManager) Delete(ctx context.Context, path string) error {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return err
	}
	return storage.Delete(ctx, path)
}

// Exists 检查文件是否存在（使用默认存储）
func (m *StorageManager) Exists(ctx context.Context, path string) (bool, error) {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return false, err
	}
	return storage.Exists(ctx, path)
}

// GetURL 获取文件访问URL（使用默认存储）
func (m *StorageManager) GetURL(ctx context.Context, path string, expireSeconds int) (string, error) {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return "", err
	}
	return storage.GetURL(ctx, path, expireSeconds)
}

// ListObjects 列出对象（使用默认存储）
func (m *StorageManager) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]Object, error) {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return nil, err
	}
	return storage.ListObjects(ctx, prefix, maxKeys)
}

// CopyObject 复制对象（使用默认存储）
func (m *StorageManager) CopyObject(ctx context.Context, srcPath, dstPath string) error {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return err
	}
	return storage.CopyObject(ctx, srcPath, dstPath)
}

// GetObjectInfo 获取对象信息（使用默认存储）
func (m *StorageManager) GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error) {
	storage, err := m.GetDefaultStorage()
	if err != nil {
		return nil, err
	}
	return storage.GetObjectInfo(ctx, path)
}
