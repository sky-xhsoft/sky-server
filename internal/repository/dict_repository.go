package repository

import "github.com/sky-xhsoft/sky-server/internal/model/entity"

// DictRepository 数据字典仓储接口
type DictRepository interface {
	// 获取所有字典
	GetAllDicts() ([]*entity.SysDict, error)

	// 根据名称获取字典
	GetDictByName(name string) (*entity.SysDict, error)

	// 根据ID获取字典
	GetDictByID(id uint) (*entity.SysDict, error)

	// 获取字典的所有项
	GetDictItems(dictID uint) ([]*entity.SysDictItem, error)

	// 根据字典名称获取字典项
	GetDictItemsByName(dictName string) ([]*entity.SysDictItem, error)
}
