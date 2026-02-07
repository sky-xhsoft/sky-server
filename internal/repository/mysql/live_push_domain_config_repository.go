package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// livePushDomainConfigRepository 直播推流域名配置仓储MySQL实现
type livePushDomainConfigRepository struct {
	db *gorm.DB
}

// NewLivePushDomainConfigRepository 创建直播推流域名配置仓储
func NewLivePushDomainConfigRepository(db *gorm.DB) repository.LivePushDomainConfigRepository {
	return &livePushDomainConfigRepository{db: db}
}

func (r *livePushDomainConfigRepository) Create(config *entity.LivePushDomainConfig) error {
	return r.db.Create(config).Error
}

func (r *livePushDomainConfigRepository) Update(config *entity.LivePushDomainConfig) error {
	return r.db.Save(config).Error
}

func (r *livePushDomainConfigRepository) Delete(id uint) error {
	return r.db.Delete(&entity.LivePushDomainConfig{}, id).Error
}

func (r *livePushDomainConfigRepository) GetByID(id uint) (*entity.LivePushDomainConfig, error) {
	var config entity.LivePushDomainConfig
	err := r.db.Where("ID = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *livePushDomainConfigRepository) GetByCompanyID(companyID uint) ([]*entity.LivePushDomainConfig, error) {
	var configs []*entity.LivePushDomainConfig
	err := r.db.Where("SYS_COMPANY_ID = ?", companyID).
		Order("IS_DEFAULT DESC, CREATED_AT DESC").
		Find(&configs).Error
	return configs, err
}

func (r *livePushDomainConfigRepository) GetActiveByCompanyID(companyID uint) ([]*entity.LivePushDomainConfig, error) {
	var configs []*entity.LivePushDomainConfig
	err := r.db.Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", companyID, true).
		Order("IS_DEFAULT DESC, CREATED_AT DESC").
		Find(&configs).Error
	return configs, err
}

func (r *livePushDomainConfigRepository) GetDefaultByCompanyID(companyID uint) (*entity.LivePushDomainConfig, error) {
	var config entity.LivePushDomainConfig
	err := r.db.Where("SYS_COMPANY_ID = ? AND IS_DEFAULT = ? AND IS_ACTIVE = ?", companyID, true, true).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *livePushDomainConfigRepository) GetByDomainName(companyID uint, domainName string) (*entity.LivePushDomainConfig, error) {
	var config entity.LivePushDomainConfig
	err := r.db.Where("SYS_COMPANY_ID = ? AND DOMAIN_NAME = ?", companyID, domainName).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *livePushDomainConfigRepository) SetDefault(companyID uint, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先取消该公司所有域名的默认状态
		if err := tx.Model(&entity.LivePushDomainConfig{}).
			Where("SYS_COMPANY_ID = ?", companyID).
			Update("IS_DEFAULT", false).Error; err != nil {
			return err
		}

		// 设置指定域名为默认
		if err := tx.Model(&entity.LivePushDomainConfig{}).
			Where("ID = ? AND SYS_COMPANY_ID = ?", id, companyID).
			Update("IS_DEFAULT", true).Error; err != nil {
			return err
		}

		return nil
	})
}
