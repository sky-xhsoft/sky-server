package repository

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// TranslationRepository 翻译配置仓储接口
type TranslationRepository interface {
	// 列出翻译
	ListTranslations(companyID uint, module string, page, pageSize int) ([]*entity.SysTranslation, int, error)
	// 根据ID获取翻译
	GetTranslationByID(id uint) (*entity.SysTranslation, error)
	// 根据键获取翻译
	GetTranslationByKey(companyID uint, key string, language string) (*entity.SysTranslation, error)
	// 根据模块获取翻译
	GetTranslationsByModule(companyID uint, module string, language string) ([]*entity.SysTranslation, error)
	// 添加翻译
	AddTranslation(translation *entity.SysTranslation) error
	// 更新翻译
	UpdateTranslation(translation *entity.SysTranslation) error
	// 删除翻译
	DeleteTranslation(id uint) error
	// 批量添加翻译
	BatchAddTranslations(translations []*entity.SysTranslation) error
	// 批量更新翻译
	BatchUpdateTranslations(translations []*entity.SysTranslation) error
}
