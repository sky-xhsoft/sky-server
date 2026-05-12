package repository

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// LanguageRepository 语言配置仓储接口
type LanguageRepository interface {
	// 列出所有语言
	ListLanguages(companyID uint) ([]*entity.SysLanguage, error)
	// 根据ID获取语言
	GetLanguageByID(id uint) (*entity.SysLanguage, error)
	// 根据代码获取语言
	GetLanguageByCode(companyID uint, code string) (*entity.SysLanguage, error)
	// 添加语言
	AddLanguage(language *entity.SysLanguage) error
	// 更新语言
	UpdateLanguage(language *entity.SysLanguage) error
	// 删除语言
	DeleteLanguage(id uint) error
	// 获取默认语言
	GetDefaultLanguage(companyID uint) (*entity.SysLanguage, error)
	// 设置默认语言
	SetDefaultLanguage(id, companyID uint) error
}
