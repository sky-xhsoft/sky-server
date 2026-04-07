package repository

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// SysCompanyConfRepository 公司配置仓储接口
type SysCompanyConfRepository interface {
	// 获取公司配置
	GetByCompanyID(ctx context.Context, sysCompanyID uint) (*entity.SysCompanyConf, error)

	// 获取默认公司（ID=1）的配置
	GetDefaultConfig(ctx context.Context) (*entity.SysCompanyConf, error)

	// 创建公司配置
	Create(ctx context.Context, config *entity.SysCompanyConf) error

	// 更新公司配置
	Update(ctx context.Context, config *entity.SysCompanyConf) error

	// 删除公司配置（软删除）
	Delete(ctx context.Context, id uint) error

	// 事务操作
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}
