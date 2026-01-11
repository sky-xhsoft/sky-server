package repository

import "github.com/sky-xhsoft/sky-server/internal/model/entity"

// MetadataRepository 元数据仓储接口
type MetadataRepository interface {
	// ========== 表定义 ==========
	// 获取所有表
	GetAllTables() ([]*entity.SysTable, error)

	// 根据表名获取表
	GetTableByName(name string) (*entity.SysTable, error)

	// 根据ID获取表
	GetTableByID(id uint) (*entity.SysTable, error)

	// ========== 字段定义 ==========
	// 获取表的所有字段
	GetColumnsByTableID(tableID uint) ([]*entity.SysColumn, error)

	// 根据字段全名获取字段
	GetColumnByFullName(fullName string) (*entity.SysColumn, error)

	// 根据ID获取字段
	GetColumnByID(id uint) (*entity.SysColumn, error)

	// ========== 表关联关系 ==========
	// 获取表的所有关联关系
	GetTableRefsByTableID(tableID uint) ([]*entity.SysTableRef, error)

	// ========== 动作定义 ==========
	// 获取表的所有动作
	GetActionsByTableID(tableID uint) ([]*entity.SysAction, error)

	// 根据ID获取动作
	GetActionByID(id uint) (*entity.SysAction, error)

	// ========== 表命令钩子 ==========
	// 获取表的所有命令钩子
	GetTableCmdsByTableID(tableID uint) ([]*entity.SysTableCmd, error)

	// 获取表的特定操作钩子
	GetTableCmdsByAction(tableID uint, action, event string) ([]*entity.SysTableCmd, error)
}
