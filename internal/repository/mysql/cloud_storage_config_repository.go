package mysql

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// cloudStorageConfigRepository 云存储配置仓储MySQL实现
type cloudStorageConfigRepository struct {
	db *gorm.DB
}

// NewCloudStorageConfigRepository 创建云存储配置仓储
func NewCloudStorageConfigRepository(db *gorm.DB) repository.CloudStorageConfigRepository {
	return &cloudStorageConfigRepository{db: db}
}

// GetByCompanyID 获取公司的存储配置
func (r *cloudStorageConfigRepository) GetByCompanyID(ctx context.Context, sysCompanyID uint) (*entity.CloudStorageConfig, error) {
	var config entity.CloudStorageConfig
	err := r.db.WithContext(ctx).
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", sysCompanyID, "Y").
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetAllActiveConfigs 获取所有有效的存储配置
func (r *cloudStorageConfigRepository) GetAllActiveConfigs(ctx context.Context) ([]*entity.CloudStorageConfig, error) {
	var configs []*entity.CloudStorageConfig
	err := r.db.WithContext(ctx).
		Where("IS_ACTIVE = ?", "Y").
		Order("SYS_COMPANY_ID ASC").
		Find(&configs).Error

	return configs, err
}

// GetDefaultConfig 获取默认公司（ID=1）的存储配置
func (r *cloudStorageConfigRepository) GetDefaultConfig(ctx context.Context) (*entity.CloudStorageConfig, error) {
	var config entity.CloudStorageConfig
	err := r.db.WithContext(ctx).
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", 1, "Y").
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Create 创建存储配置
func (r *cloudStorageConfigRepository) Create(ctx context.Context, config *entity.CloudStorageConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新存储配置
func (r *cloudStorageConfigRepository) Update(ctx context.Context, config *entity.CloudStorageConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除存储配置（软删除）
func (r *cloudStorageConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.CloudStorageConfig{}).
		Where("ID = ?", id).
		Update("IS_ACTIVE", "N").Error
}

// GetCount 获取配置总数
func (r *cloudStorageConfigRepository) GetCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.CloudStorageConfig{}).
		Where("IS_ACTIVE = ?", "Y").
		Count(&count).Error

	return count, err
}

// Transaction 事务操作
func (r *cloudStorageConfigRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
