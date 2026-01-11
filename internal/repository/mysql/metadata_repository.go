package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// metadataRepository 元数据仓储MySQL实现
type metadataRepository struct {
	db *gorm.DB
}

// NewMetadataRepository 创建元数据仓储
func NewMetadataRepository(db *gorm.DB) repository.MetadataRepository {
	return &metadataRepository{db: db}
}

// ========== 表定义 ==========

func (r *metadataRepository) GetAllTables() ([]*entity.SysTable, error) {
	var tables []*entity.SysTable
	err := r.db.Where("IS_ACTIVE = ?", "Y").Find(&tables).Error
	return tables, err
}

func (r *metadataRepository) GetTableByName(name string) (*entity.SysTable, error) {
	var table entity.SysTable
	err := r.db.Where("NAME = ? AND IS_ACTIVE = ?", name, "Y").First(&table).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *metadataRepository) GetTableByID(id uint) (*entity.SysTable, error) {
	var table entity.SysTable
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&table).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

// ========== 字段定义 ==========

func (r *metadataRepository) GetColumnsByTableID(tableID uint) ([]*entity.SysColumn, error) {
	var columns []*entity.SysColumn
	err := r.db.Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", tableID, "Y").
		Order("ORDERNO ASC").
		Find(&columns).Error
	return columns, err
}

func (r *metadataRepository) GetColumnByFullName(fullName string) (*entity.SysColumn, error) {
	var column entity.SysColumn
	err := r.db.Where("FULL_NAME = ? AND IS_ACTIVE = ?", fullName, "Y").First(&column).Error
	if err != nil {
		return nil, err
	}
	return &column, nil
}

func (r *metadataRepository) GetColumnByID(id uint) (*entity.SysColumn, error) {
	var column entity.SysColumn
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&column).Error
	if err != nil {
		return nil, err
	}
	return &column, nil
}

// ========== 表关联关系 ==========

func (r *metadataRepository) GetTableRefsByTableID(tableID uint) ([]*entity.SysTableRef, error) {
	var refs []*entity.SysTableRef
	err := r.db.Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", tableID, "Y").
		Order("ORDERNO ASC").
		Find(&refs).Error
	return refs, err
}

// ========== 动作定义 ==========

func (r *metadataRepository) GetActionsByTableID(tableID uint) ([]*entity.SysAction, error) {
	var actions []*entity.SysAction
	err := r.db.Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", tableID, "Y").
		Order("ORDERNO ASC").
		Find(&actions).Error
	return actions, err
}

func (r *metadataRepository) GetActionByID(id uint) (*entity.SysAction, error) {
	var action entity.SysAction
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&action).Error
	if err != nil {
		return nil, err
	}
	return &action, nil
}

// ========== 表命令钩子 ==========

func (r *metadataRepository) GetTableCmdsByTableID(tableID uint) ([]*entity.SysTableCmd, error) {
	var cmds []*entity.SysTableCmd
	err := r.db.Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", tableID, "Y").
		Order("ORDERNO ASC").
		Find(&cmds).Error
	return cmds, err
}

func (r *metadataRepository) GetTableCmdsByAction(tableID uint, action, event string) ([]*entity.SysTableCmd, error) {
	var cmds []*entity.SysTableCmd
	err := r.db.Where("SYS_TABLE_ID = ? AND ACTION = ? AND EVENT = ? AND IS_ACTIVE = ?", tableID, action, event, "Y").
		Order("ORDERNO ASC").
		Find(&cmds).Error
	return cmds, err
}
