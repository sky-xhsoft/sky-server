package cloud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StorageConfigService 存储配置管理服务接口
type StorageConfigService interface {
	// 创建存储配置
	CreateConfig(ctx context.Context, config *entity.CloudStorageConfig) error

	// 更新存储配置
	UpdateConfig(ctx context.Context, config *entity.CloudStorageConfig) error

	// 获取公司配置
	GetCompanyConfig(ctx context.Context, companyID uint) (*entity.CloudStorageConfig, error)

	// 获取所有配置
	GetAllConfigs(ctx context.Context) ([]*entity.CloudStorageConfig, error)

	// 删除配置
	DeleteConfig(ctx context.Context, id uint) error

	// 刷新缓存
	RefreshCache(ctx context.Context, companyID uint) error

	// 检查配置是否有效
	ValidateConfig(ctx context.Context, config *entity.CloudStorageConfig) error
}

// storageConfigService 存储配置管理服务实现
type storageConfigService struct {
	db             *gorm.DB
	configRepo     repository.CloudStorageConfigRepository
	storageManager *storage.CompanyStorageManager
}

// NewStorageConfigService 创建存储配置管理服务
func NewStorageConfigService(
	db *gorm.DB,
	configRepo repository.CloudStorageConfigRepository,
	storageManager *storage.CompanyStorageManager,
) StorageConfigService {
	return &storageConfigService{
		db:             db,
		configRepo:     configRepo,
		storageManager: storageManager,
	}
}

// CreateConfig 创建存储配置
func (s *storageConfigService) CreateConfig(ctx context.Context, config *entity.CloudStorageConfig) error {
	logger.Info("创建存储配置",
		zap.Uint("companyID", config.SysCompanyID),
		zap.String("storageType", config.StorageType))

	// 检查是否已存在配置
	existing, err := s.configRepo.GetByCompanyID(ctx, config.SysCompanyID)
	if err != nil {
		logger.Error("检查存储配置是否已存在失败", zap.Error(err))
		return errors.Wrap(errors.ErrDatabase, "检查存储配置是否已存在失败", err)
	}

	if existing != nil {
		logger.Warn("公司已存在存储配置", zap.Uint("companyID", config.SysCompanyID))
		return errors.Wrap(errors.ErrResourceExists, "公司已存在存储配置", nil)
	}

	// 验证配置
	if err := s.ValidateConfig(ctx, config); err != nil {
		logger.Error("存储配置验证失败", zap.Uint("companyID", config.SysCompanyID), zap.Error(err))
		return err
	}

	// 创建配置
	if err := s.configRepo.Create(ctx, config); err != nil {
		logger.Error("创建存储配置失败", zap.Uint("companyID", config.SysCompanyID), zap.Error(err))
		return errors.Wrap(errors.ErrDatabase, "创建存储配置失败", err)
	}

	// 刷新缓存
	if err := s.storageManager.RefreshCompanyConfig(ctx, config.SysCompanyID); err != nil {
		logger.Warn("刷新存储配置缓存失败", zap.Uint("companyID", config.SysCompanyID), zap.Error(err))
	}

	logger.Info("存储配置创建成功", zap.Uint("companyID", config.SysCompanyID))
	return nil
}

// UpdateConfig 更新存储配置
func (s *storageConfigService) UpdateConfig(ctx context.Context, config *entity.CloudStorageConfig) error {
	logger.Info("更新存储配置",
		zap.Uint("id", config.ID),
		zap.Uint("companyID", config.SysCompanyID),
		zap.String("storageType", config.StorageType))

	// 检查配置是否存在
	existing, err := s.getConfigByID(ctx, config.ID)
	if err != nil {
		return err
	}

	if existing == nil {
		logger.Warn("存储配置不存在", zap.Uint("id", config.ID))
		return errors.Wrap(errors.ErrResourceNotFound, "存储配置不存在", nil)
	}

	// 验证配置
	if err := s.ValidateConfig(ctx, config); err != nil {
		logger.Error("存储配置验证失败", zap.Uint("id", config.ID), zap.Error(err))
		return err
	}

	// 更新配置
	if err := s.configRepo.Update(ctx, config); err != nil {
		logger.Error("更新存储配置失败", zap.Uint("id", config.ID), zap.Error(err))
		return errors.Wrap(errors.ErrDatabase, "更新存储配置失败", err)
	}

	// 刷新缓存
	if err := s.storageManager.RefreshCompanyConfig(ctx, config.SysCompanyID); err != nil {
		logger.Warn("刷新存储配置缓存失败", zap.Uint("companyID", config.SysCompanyID), zap.Error(err))
	}

	logger.Info("存储配置更新成功", zap.Uint("id", config.ID))
	return nil
}

// GetCompanyConfig 获取公司配置
func (s *storageConfigService) GetCompanyConfig(ctx context.Context, companyID uint) (*entity.CloudStorageConfig, error) {
	logger.Debug("获取公司存储配置", zap.Uint("companyID", companyID))

	config, err := s.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		logger.Error("获取公司存储配置失败", zap.Uint("companyID", companyID), zap.Error(err))
		return nil, errors.Wrap(errors.ErrDatabase, "获取公司存储配置失败", err)
	}

	return config, nil
}

// GetAllConfigs 获取所有配置
func (s *storageConfigService) GetAllConfigs(ctx context.Context) ([]*entity.CloudStorageConfig, error) {
	logger.Debug("获取所有存储配置")

	configs, err := s.configRepo.GetAllActiveConfigs(ctx)
	if err != nil {
		logger.Error("获取所有存储配置失败", zap.Error(err))
		return nil, errors.Wrap(errors.ErrDatabase, "获取所有存储配置失败", err)
	}

	return configs, nil
}

// DeleteConfig 删除配置
func (s *storageConfigService) DeleteConfig(ctx context.Context, id uint) error {
	logger.Info("删除存储配置", zap.Uint("id", id))

	// 检查配置是否存在
	config, err := s.getConfigByID(ctx, id)
	if err != nil {
		return err
	}

	if config == nil {
		logger.Warn("存储配置不存在", zap.Uint("id", id))
		return errors.Wrap(errors.ErrResourceNotFound, "存储配置不存在", nil)
	}

	// 软删除配置
	if err := s.configRepo.Delete(ctx, id); err != nil {
		logger.Error("删除存储配置失败", zap.Uint("id", id), zap.Error(err))
		return errors.Wrap(errors.ErrDatabase, "删除存储配置失败", err)
	}

	// 刷新缓存
	if err := s.storageManager.RefreshCompanyConfig(ctx, config.SysCompanyID); err != nil {
		logger.Warn("刷新存储配置缓存失败", zap.Uint("companyID", config.SysCompanyID), zap.Error(err))
	}

	logger.Info("存储配置删除成功", zap.Uint("id", id))
	return nil
}

// RefreshCache 刷新缓存
func (s *storageConfigService) RefreshCache(ctx context.Context, companyID uint) error {
	logger.Info("刷新存储配置缓存", zap.Uint("companyID", companyID))

	if err := s.storageManager.RefreshCompanyConfig(ctx, companyID); err != nil {
		logger.Error("刷新存储配置缓存失败", zap.Uint("companyID", companyID), zap.Error(err))
		return errors.Wrap(errors.ErrInternal, "刷新存储配置缓存失败", err)
	}

	logger.Info("存储配置缓存刷新成功", zap.Uint("companyID", companyID))
	return nil
}

// ValidateConfig 验证配置
func (s *storageConfigService) ValidateConfig(ctx context.Context, config *entity.CloudStorageConfig) error {
	if config.SysCompanyID == 0 {
		return errors.Wrap(errors.ErrInvalidParam, "公司ID不能为空", nil)
	}

	if config.StorageType == "" {
		return errors.Wrap(errors.ErrInvalidParam, "存储类型不能为空", nil)
	}

	// 根据存储类型验证必填字段
	switch config.StorageType {
	case "local":
		if config.LocalBasePath == "" {
			return errors.Wrap(errors.ErrInvalidParam, "本地存储基础路径不能为空", nil)
		}
		if config.LocalBaseURL == "" {
			return errors.Wrap(errors.ErrInvalidParam, "本地存储基础URL不能为空", nil)
		}
	case "aliyunOSS":
		if config.AliyunOSSEndpoint == "" {
			return errors.Wrap(errors.ErrInvalidParam, "阿里云OSS端点不能为空", nil)
		}
		if config.AliyunOSSAccessKeyID == "" {
			return errors.Wrap(errors.ErrInvalidParam, "阿里云OSS AccessKeyID不能为空", nil)
		}
		if config.AliyunOSSAccessKeySecret == "" {
			return errors.Wrap(errors.ErrInvalidParam, "阿里云OSS AccessKeySecret不能为空", nil)
		}
		if config.AliyunOSSBucketName == "" {
			return errors.Wrap(errors.ErrInvalidParam, "阿里云OSS Bucket名称不能为空", nil)
		}
	case "tencentCOS":
		if config.TencentCOSBucketURL == "" {
			return errors.Wrap(errors.ErrInvalidParam, "腾讯云COS Bucket URL不能为空", nil)
		}
		if config.TencentCOSSecretID == "" {
			return errors.Wrap(errors.ErrInvalidParam, "腾讯云COS SecretID不能为空", nil)
		}
		if config.TencentCOSSecretKey == "" {
			return errors.Wrap(errors.ErrInvalidParam, "腾讯云COS SecretKey不能为空", nil)
		}
		if config.TencentCOSBucketName == "" {
			return errors.Wrap(errors.ErrInvalidParam, "腾讯云COS Bucket名称不能为空", nil)
		}
		if config.TencentCOSRegion == "" {
			return errors.Wrap(errors.ErrInvalidParam, "腾讯云COS区域不能为空", nil)
		}
	default:
		return errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("不支持的存储类型: %s", config.StorageType), nil)
	}

	// 验证配置的其他字段
	if config.IsActive != "Y" && config.IsActive != "N" {
		config.IsActive = "Y" // 默认激活
	}

	return nil
}

// getConfigByID 根据ID获取配置
func (s *storageConfigService) getConfigByID(ctx context.Context, id uint) (*entity.CloudStorageConfig, error) {
	var config entity.CloudStorageConfig
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("查询存储配置失败", zap.Uint("id", id), zap.Error(err))
		return nil, errors.Wrap(errors.ErrDatabase, "查询存储配置失败", err)
	}

	return &config, nil
}
