package crud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/transaction"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Create 创建记录
func (s *service) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}
	if err := s.requirePermission(ctx, userID, table, groups.PermCreate, "无创建权限"); err != nil {
		return nil, err
	}

	// 获取字段定义（在事务外，避免长时间持有锁）
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 验证和处理字段、生成ID、填审计字段与所属公司（在事务外）
	processedData, err := s.prepareCreate(ctx, table, columns, data, userID)
	if err != nil {
		return nil, err
	}

	logger.Debug("Prepared insert data", zap.Any("data", processedData))

	// 在事务中执行：before钩子 + 插入 + after钩子
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		return s.insertWithHooks(ctx, tx, table, data, processedData)
	})
	if err != nil {
		return nil, err
	}

	// 直接返回创建的记录数据（不通过 GetOne 查询，避免 IS_ACTIVE 过滤）
	// 因为刚创建的记录可能 IS_ACTIVE = 'N'，GetOne 会过滤掉
	return processedData, nil
}

// prepareCreate 新增前的字段处理：默认值、所属公司（服务端决定）、字段校验、ID、审计字段
func (s *service) prepareCreate(ctx context.Context, table *entity.SysTable, columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, errors.New(errors.ErrPermissionDenied, "无法确定当前用户")
	}

	// IS_ACTIVE 默认值（必须在 processFieldsForCreate 之前，避免 NullAble 误报）
	if _, exists := data[colIsActive]; !exists {
		data[colIsActive] = "Y"
	}

	// 所属公司由服务端决定，忽略请求里的值（必须在 processFieldsForCreate 之前，避免必填 / 外键校验失败）
	hasCompany := s.hasCompanyColumn(table.Name, columns)
	var company interface{}
	if hasCompany {
		company = createCompanyID(table.Name, data[colSysCompanyID], user.SysCompanyID)
		if n, ok := company.(int64); ok && n > 0 {
			// 平台为其他公司建记录：公司必须存在
			var cnt int64
			if err := s.db.Table("sys_company").Where("ID = ?", n).Count(&cnt).Error; err != nil || cnt == 0 {
				return nil, errors.New(errors.ErrInvalidParam, "所属公司不存在")
			}
		}
		data[colSysCompanyID] = company
	} else {
		delete(data, colSysCompanyID)
	}

	processedData, err := s.processFieldsForCreate(columns, data, userID)
	if err != nil {
		return nil, err
	}
	if err := checkUniqueColumns(s.db, table.Name, 0, processedData); err != nil {
		return nil, err
	}

	// 生成新的ID（在事务外，避免长时间持有锁）
	newID, err := s.idgenService.GetNextID(ctx, table.Name)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成ID失败", err)
	}
	processedData[colID] = newID

	// IS_ACTIVE: 如果前端传入了值则使用前端的值，否则默认为 "Y"
	if _, exists := processedData[colIsActive]; !exists {
		processedData[colIsActive] = "Y"
	}
	processedData[colCreateBy] = user.Username
	processedData[colCreateTime] = time.Now()
	if hasCompany {
		processedData[colSysCompanyID] = company
	}
	return processedData, nil
}

// insertWithHooks 在事务中执行：before钩子 + 插入 + after钩子
func (s *service) insertWithHooks(ctx context.Context, tx *gorm.DB, table *entity.SysTable, data, processedData map[string]interface{}) error {
	if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "begin", data); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
	}
	if err := tx.Table(table.Name).Create(&processedData).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建失败", err)
	}
	if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "end", processedData); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
	}
	return nil
}

// Update 更新记录
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}
	if err := s.requirePermission(ctx, userID, table, groups.PermUpdate, "无修改权限"); err != nil {
		return err
	}

	// 获取字段定义（在事务外）
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return err
	}
	if err := s.ensureWritable(ctx, table, columns, userID, id); err != nil {
		return err
	}

	// 验证和处理字段（在事务外）
	processedData, updateFields, err := s.prepareUpdate(table, columns, id, data, userID)
	if err != nil {
		return err
	}

	// 在事务中执行：before钩子 + 更新 + after钩子
	return transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		return s.updateWithHooks(ctx, tx, table, id, data, processedData, updateFields)
	})
}

// prepareUpdate 修改前的字段处理：所属公司不能改、字段校验、审计字段；返回要写的字段列表（支持零值更新）
func (s *service) prepareUpdate(table *entity.SysTable, columns []*entity.SysColumn, id uint, data map[string]interface{}, userID uint) (map[string]interface{}, []string, error) {
	// 添加ID到数据中供钩子使用
	data[colID] = id

	// 所属公司不能通过修改变更（整条提交时带着原值，允许）
	if v, ok := data[colSysCompanyID]; ok {
		if s.hasCompanyColumn(table.Name, columns) {
			if err := s.checkCompanyUnchanged(table.Name, id, v); err != nil {
				return nil, nil, err
			}
		}
		delete(data, colSysCompanyID)
	}

	processedData, err := s.processFieldsForUpdate(columns, data, userID)
	if err != nil {
		return nil, nil, err
	}
	if err := checkUniqueColumns(s.db, table.Name, id, processedData); err != nil {
		return nil, nil, err
	}
	delete(processedData, colSysCompanyID)
	// 请求的业务字段全部因不可编辑被跳过时明确报错（原先仍返回“更新成功”，掩盖了 MASK 配置错误）
	if skipped := skippedUpdateFields(columns, data, processedData); len(skipped) > 0 {
		return nil, nil, errors.New(errors.ErrInvalidParam, "以下字段不可编辑: "+strings.Join(skipped, "、"))
	}

	if user, err := s.userRepo.GetUserByID(userID); err == nil && user != nil {
		processedData[colUpdateBy] = user.Username
	}
	processedData[colUpdateTime] = time.Now()

	updateFields := make([]string, 0, len(processedData))
	for field := range processedData {
		updateFields = append(updateFields, field)
	}
	return processedData, updateFields, nil
}

// updateWithHooks 在事务中执行：before钩子 + 更新 + after钩子
func (s *service) updateWithHooks(ctx context.Context, tx *gorm.DB, table *entity.SysTable, id uint, data, processedData map[string]interface{}, updateFields []string) error {
	if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "begin", data); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
	}

	// 使用 Select 明确指定要更新的字段，包括零值；不过滤 IS_ACTIVE，允许更新任何状态的记录
	result := tx.Table(table.Name).
		Where("ID = ?", id).
		Select(updateFields).
		Updates(processedData)
	if result.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "更新失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.ErrResourceNotFound, "记录不存在")
	}

	processedData[colID] = id
	if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "end", processedData); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
	}
	return nil
}

// Delete 删除记录（物理删除）
func (s *service) Delete(ctx context.Context, tableName string, id uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查删除权限与写入范围
	if err := s.requirePermission(ctx, userID, table, groups.PermDelete, "无删除权限"); err != nil {
		return err
	}
	if err := s.ensureWritableByID(ctx, table, userID, id); err != nil {
		return err
	}

	// 检查外键引用（REF_ON_DELETE 保护）
	fkRefs, err := s.checkForeignKeyReferences(ctx, table.ID, id)
	if err != nil {
		return err
	}

	// 在事务中执行：级联处理 + before钩子 + 删除 + after钩子
	deleteData := map[string]interface{}{"ID": id}
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 处理 REF_ON_DELETE=cascade 和 setNull
		for _, ref := range fkRefs {
			switch ref.RefOnDelete {
			case "cascade":
				if err := tx.Table(ref.TableName).Where(ref.ColName+" = ?", id).Delete(nil).Error; err != nil {
					return errors.Wrap(errors.ErrDatabase, fmt.Sprintf("级联删除 %s 失败", ref.TableName), err)
				}
			case "setNull":
				if err := tx.Table(ref.TableName).Where(ref.ColName+" = ?", id).Update(ref.ColName, nil).Error; err != nil {
					return errors.Wrap(errors.ErrDatabase, fmt.Sprintf("级联置空 %s 失败", ref.TableName), err)
				}
			}
		}

		// 执行before钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
		}

		// 先删除所有1:1关联的子表数据
		tableRefs, err := s.metadataService.GetTableRefs(table.ID)
		if err == nil && len(tableRefs) > 0 {
			for _, ref := range tableRefs {
				// 只处理1:1关联
				if ref.AssocType == "1" {
					// 获取关联表信息
					refTable, err := s.metadataService.GetTableByID(uint(ref.RefTableID))
					if err != nil {
						continue
					}
					// 获取关联字段信息
					var refColumnName string
					if ref.RefColumnID > 0 {
						columns, err := s.metadataService.GetColumns(uint(ref.RefTableID))
						if err == nil {
							for _, col := range columns {
								if col.ID == uint(ref.RefColumnID) {
									refColumnName = col.DbName
									break
								}
							}
						}
					}
					// 如果没有找到关联字段，使用默认的主表ID字段
					if refColumnName == "" {
						refColumnName = table.Name + "_ID"
					}
					// 删除子表中关联的记录
					if err := tx.Table(refTable.Name).Where(refColumnName+" = ?", id).Delete(nil).Error; err != nil {
						return errors.Wrap(errors.ErrDatabase, "删除关联子表失败", err)
					}
				}
			}
		}

		// 执行物理删除（在事务中）
		result := tx.Table(table.Name).Where("ID = ?", id).Delete(nil)
		if result.Error != nil {
			return errors.Wrap(errors.ErrDatabase, "删除失败", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New(errors.ErrResourceNotFound, "记录不存在")
		}

		// 执行after钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
		}

		return nil
	})

	return err
}

// BatchDelete 批量删除
func (s *service) BatchDelete(ctx context.Context, tableName string, ids []uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查删除权限与写入范围
	if err := s.requirePermission(ctx, userID, table, groups.PermDelete, "无删除权限"); err != nil {
		return err
	}
	if err := s.ensureWritableByID(ctx, table, userID, ids...); err != nil {
		return err
	}

	// 检查每个ID的外键引用
	for _, id := range ids {
		_, err := s.checkForeignKeyReferences(ctx, table.ID, id)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, fmt.Sprintf("ID=%d: %s", id, err.Error()), err)
		}
	}

	// 在事务中执行批量删除
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 先删除所有1:1关联的子表数据
		tableRefs, err := s.metadataService.GetTableRefs(table.ID)
		if err == nil && len(tableRefs) > 0 {
			for _, ref := range tableRefs {
				// 只处理1:1关联
				if ref.AssocType == "1" {
					// 获取关联表信息
					refTable, err := s.metadataService.GetTableByID(uint(ref.RefTableID))
					if err != nil {
						continue
					}
					// 获取关联字段信息
					var refColumnName string
					if ref.RefColumnID > 0 {
						columns, err := s.metadataService.GetColumns(uint(ref.RefTableID))
						if err == nil {
							for _, col := range columns {
								if col.ID == uint(ref.RefColumnID) {
									refColumnName = col.DbName
									break
								}
							}
						}
					}
					// 如果没有找到关联字段，使用默认的主表ID字段
					if refColumnName == "" {
						refColumnName = table.Name + "_ID"
					}
					// 批量删除子表中关联的记录
					if err := tx.Table(refTable.Name).Where(refColumnName+" IN ?", ids).Delete(nil).Error; err != nil {
						return errors.Wrap(errors.ErrDatabase, "删除关联子表失败", err)
					}
				}
			}
		}

		// 对每个ID执行before钩子（在事务中）
		for _, id := range ids {
			deleteData := map[string]interface{}{"ID": id}
			if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
				return errors.Wrap(errors.ErrInternal, fmt.Sprintf("执行ID=%d的before钩子失败", id), err)
			}
		}

		// 执行批量物理删除（在事务中）
		result := tx.Table(table.Name).Where("ID IN ?", ids).Delete(nil)
		if result.Error != nil {
			return errors.Wrap(errors.ErrDatabase, "批量删除失败", result.Error)
		}

		// 对每个ID执行after钩子（在事务中）
		for _, id := range ids {
			deleteData := map[string]interface{}{"ID": id}
			if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
				return errors.Wrap(errors.ErrInternal, fmt.Sprintf("执行ID=%d的after钩子失败", id), err)
			}
		}

		return nil
	})

	return err
}
