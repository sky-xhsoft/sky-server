package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

type translationRepository struct {
	db *gorm.DB
}

func NewTranslationRepository(db *gorm.DB) repository.TranslationRepository {
	return &translationRepository{db: db}
}

func (r *translationRepository) ListTranslations(companyID uint, module string, page, pageSize int) ([]*entity.SysTranslation, int, error) {
	var translations []*entity.SysTranslation
	var total int64

	query := r.db.
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y")

	if module != "" {
		query = query.Where("MODULE = ?", module)
	}

	// 计算总数
	if err := query.Model(&entity.SysTranslation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		err := query.
			Order("MODULE, KEY, LANGUAGE").
			Offset(offset).
			Limit(pageSize).
			Find(&translations).Error
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := query.
			Order("MODULE, KEY, LANGUAGE").
			Find(&translations).Error
		if err != nil {
			return nil, 0, err
		}
	}

	return translations, int(total), nil
}

func (r *translationRepository) GetTranslationByID(id uint) (*entity.SysTranslation, error) {
	var translation entity.SysTranslation
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&translation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &translation, nil
}

func (r *translationRepository) GetTranslationByKey(companyID uint, key string, language string) (*entity.SysTranslation, error) {
	var translation entity.SysTranslation
	err := r.db.
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND KEY = ? AND LANGUAGE = ? AND IS_ACTIVE = ?", companyID, key, language, "Y").
		First(&translation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &translation, nil
}

func (r *translationRepository) GetTranslationsByModule(companyID uint, module string, language string) ([]*entity.SysTranslation, error) {
	var translations []*entity.SysTranslation
	query := r.db.
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y")

	if module != "" {
		query = query.Where("MODULE = ?", module)
	}

	if language != "" {
		query = query.Where("LANGUAGE = ?", language)
	}

	err := query.
		Order("KEY").
		Find(&translations).Error
	return translations, err
}

func (r *translationRepository) AddTranslation(translation *entity.SysTranslation) error {
	return r.db.Create(translation).Error
}

func (r *translationRepository) UpdateTranslation(translation *entity.SysTranslation) error {
	return r.db.Save(translation).Error
}

func (r *translationRepository) DeleteTranslation(id uint) error {
	return r.db.Delete(&entity.SysTranslation{}, id).Error
}

func (r *translationRepository) BatchAddTranslations(translations []*entity.SysTranslation) error {
	if len(translations) == 0 {
		return nil
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, translation := range translations {
		if err := tx.Create(translation).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *translationRepository) BatchUpdateTranslations(translations []*entity.SysTranslation) error {
	if len(translations) == 0 {
		return nil
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, translation := range translations {
		if err := tx.Save(translation).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
