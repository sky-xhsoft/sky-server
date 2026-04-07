package repository

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// CloudStorageConfigRepository 云存储配置仓储接口
type CloudStorageConfigRepository interface {
	// 获取公司的存储配置
	GetByCompanyID(ctx context.Context, sysCompanyID uint) (*entity.CloudStorageConfig, error)

	// 获取所有有效的存储配置
	GetAllActiveConfigs(ctx context.Context) ([]*entity.CloudStorageConfig, error)

	// 获取默认公司（ID=1）的存储配置
	GetDefaultConfig(ctx context.Context) (*entity.CloudStorageConfig, error)

	// 创建存储配置
	Create(ctx context.Context, config *entity.CloudStorageConfig) error

	// 更新存储配置
	Update(ctx context.Context, config *entity.CloudStorageConfig) error

	// 删除存储配置（软删除）
	Delete(ctx context.Context, id uint) error

	// 获取配置总数
	GetCount(ctx context.Context) (int64, error)

	// 事务操作
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}
