package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// dictRepository 数据字典仓储MySQL实现
type dictRepository struct {
	db *gorm.DB
}

// NewDictRepository 创建数据字典仓储
func NewDictRepository(db *gorm.DB) repository.DictRepository {
	return &dictRepository{db: db}
}

func (r *dictRepository) GetAllDicts() ([]*entity.SysDict, error) {
	var dicts []*entity.SysDict
	err := r.db.Where("IS_ACTIVE = ?", "Y").Find(&dicts).Error
	return dicts, err
}

func (r *dictRepository) GetDictByName(name string) (*entity.SysDict, error) {
	var dict entity.SysDict
	err := r.db.Where("NAME = ? AND IS_ACTIVE = ?", name, "Y").First(&dict).Error
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

func (r *dictRepository) GetDictByID(id uint) (*entity.SysDict, error) {
	var dict entity.SysDict
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&dict).Error
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

func (r *dictRepository) GetDictItems(dictID uint) ([]*entity.SysDictItem, error) {
	var items []*entity.SysDictItem
	err := r.db.Where("SYS_DICT_ID = ? AND IS_ACTIVE = ?", dictID, "Y").
		Order("ORDERNO ASC").
		Find(&items).Error
	return items, err
}

func (r *dictRepository) GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error) {
	// 先查询字典
	dict, err := r.GetDictByName(dictName)
	if err != nil {
		return nil, err
	}

	// 再查询字典项
	return r.GetDictItems(dict.ID)
}
