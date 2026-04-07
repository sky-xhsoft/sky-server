package mysql

import (
	"context"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// sysCompanyConfRepository 公司配置仓储MySQL实现
type sysCompanyConfRepository struct {
	db *gorm.DB
}

// NewSysCompanyConfRepository 创建公司配置仓储
func NewSysCompanyConfRepository(db *gorm.DB) repository.SysCompanyConfRepository {
	return &sysCompanyConfRepository{db: db}
}

// GetByCompanyID 获取公司配置
func (r *sysCompanyConfRepository) GetByCompanyID(ctx context.Context, sysCompanyID uint) (*entity.SysCompanyConf, error) {
	// 创建一个具有超时的上下文，防止数据库查询无限期挂起
	// 同时避免直接使用可能已被取消的HTTP请求上下文
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config entity.SysCompanyConf
	err := r.db.WithContext(dbCtx).
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", sysCompanyID, "Y").
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		// 忽略上下文取消错误，返回nil而不是错误
		if err == context.Canceled || err == context.DeadlineExceeded {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetDefaultConfig 获取默认公司（ID=1）的配置
func (r *sysCompanyConfRepository) GetDefaultConfig(ctx context.Context) (*entity.SysCompanyConf, error) {
	// 创建一个具有超时的上下文，防止数据库查询无限期挂起
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config entity.SysCompanyConf
	err := r.db.WithContext(dbCtx).
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", 1, "Y").
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		// 忽略上下文取消错误，返回nil而不是错误
		if err == context.Canceled || err == context.DeadlineExceeded {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Create 创建公司配置
func (r *sysCompanyConfRepository) Create(ctx context.Context, config *entity.SysCompanyConf) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新公司配置
func (r *sysCompanyConfRepository) Update(ctx context.Context, config *entity.SysCompanyConf) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除公司配置（软删除）
func (r *sysCompanyConfRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.SysCompanyConf{}).
		Where("ID = ?", id).
		Update("IS_ACTIVE", "N").Error
}

// Transaction 事务操作
func (r *sysCompanyConfRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
