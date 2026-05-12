package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

type languageRepository struct {
	db *gorm.DB
}

func NewLanguageRepository(db *gorm.DB) repository.LanguageRepository {
	return &languageRepository{db: db}
}

func (r *languageRepository) ListLanguages(companyID uint) ([]*entity.SysLanguage, error) {
	var languages []*entity.SysLanguage
	err := r.db.
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID = ?) AND IS_ACTIVE = ?", companyID, 0, "Y").
		Order("CASE WHEN IS_DEFAULT = 'Y' THEN 0 ELSE 1 END, CODE").
		Find(&languages).Error
	return languages, err
}

func (r *languageRepository) GetLanguageByID(id uint) (*entity.SysLanguage, error) {
	var language entity.SysLanguage
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&language).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) GetLanguageByCode(companyID uint, code string) (*entity.SysLanguage, error) {
	var language entity.SysLanguage
	err := r.db.
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID = ?) AND CODE = ? AND IS_ACTIVE = ?", companyID, 0, code, "Y").
		First(&language).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) AddLanguage(language *entity.SysLanguage) error {
	return r.db.Create(language).Error
}

func (r *languageRepository) UpdateLanguage(language *entity.SysLanguage) error {
	return r.db.Save(language).Error
}

func (r *languageRepository) DeleteLanguage(id uint) error {
	return r.db.Delete(&entity.SysLanguage{}, id).Error
}

func (r *languageRepository) GetDefaultLanguage(companyID uint) (*entity.SysLanguage, error) {
	var language entity.SysLanguage
	err := r.db.
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID = ?) AND IS_DEFAULT = ? AND IS_ACTIVE = ?", companyID, 0, "Y", "Y").
		First(&language).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) SetDefaultLanguage(id, companyID uint) error {
	// 首先将所有语言的默认值设置为N
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.
		Model(&entity.SysLanguage{}).
		Where("SYS_COMPANY_ID = ? OR SYS_COMPANY_ID = ?", companyID, 0).
		Update("IS_DEFAULT", "N").Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.
		Model(&entity.SysLanguage{}).
		Where("ID = ?", id).
		Update("IS_DEFAULT", "Y").Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
