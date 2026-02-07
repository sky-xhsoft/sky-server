package repository

import "github.com/sky-xhsoft/sky-server/internal/model/entity"

// LivePushDomainConfigRepository 直播推流域名配置仓储接口
type LivePushDomainConfigRepository interface {
	// Create 创建推流域名配置
	Create(config *entity.LivePushDomainConfig) error

	// Update 更新推流域名配置
	Update(config *entity.LivePushDomainConfig) error

	// Delete 删除推流域名配置
	Delete(id uint) error

	// GetByID 根据ID获取配置
	GetByID(id uint) (*entity.LivePushDomainConfig, error)

	// GetByCompanyID 获取公司的所有推流域名配置
	GetByCompanyID(companyID uint) ([]*entity.LivePushDomainConfig, error)

	// GetActiveByCompanyID 获取公司的所有启用的推流域名配置
	GetActiveByCompanyID(companyID uint) ([]*entity.LivePushDomainConfig, error)

	// GetDefaultByCompanyID 获取公司的默认推流域名配置
	GetDefaultByCompanyID(companyID uint) (*entity.LivePushDomainConfig, error)

	// GetByDomainName 根据域名获取配置
	GetByDomainName(companyID uint, domainName string) (*entity.LivePushDomainConfig, error)

	// SetDefault 设置默认域名（会取消其他默认）
	SetDefault(companyID uint, id uint) error
}
