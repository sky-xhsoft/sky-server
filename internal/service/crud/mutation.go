package crud

import (
	"context"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/transaction"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"gorm.io/gorm"
)

// Create 创建记录
func (s *service) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查创建权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermCreate)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无创建权限")
	}

	// 获取字段定义（在事务外，避免长时间持有锁）
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 验证和处理字段（在事务外）
	processedData, err := s.processFieldsForCreate(columns, data, userID)
	if err != nil {
		return nil, err
	}

	// 生成新的ID（在事务外，避免长时间持有锁）
	newID, err := s.idgenService.GetNextID(ctx, table.Name)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成ID失败", err)
	}
	processedData["ID"] = newID

	// 添加审计字段
	// IS_ACTIVE: 如果前端传入了值则使用前端的值，否则默认为 "Y"
	if _, exists := processedData["IS_ACTIVE"]; !exists {
		processedData["IS_ACTIVE"] = "Y"
	}

	// 获取用户信息以填充审计字段
	user, userErr := s.userRepo.GetUserByID(userID)
	if userErr == nil && user != nil {
		// 设置创建人和公司ID
		processedData["CREATE_BY"] = user.Username
		processedData["SYS_COMPANY_ID"] = user.SysCompanyID
		fmt.Printf("[DEBUG] 设置审计字段: CREATE_BY=%s, SYS_COMPANY_ID=%d\n", user.Username, user.SysCompanyID)
	} else {
		fmt.Printf("[DEBUG] 获取用户信息失败: userID=%d, err=%v\n", userID, userErr)
	}
	// 设置创建时间
	processedData["CREATE_TIME"] = time.Now()

	fmt.Printf("[DEBUG] 准备插入的数据: %+v\n", processedData)

	// 在事务中执行：before钩子 + 插入 + after钩子
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 执行before钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "begin", data); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
		}

		// 执行插入（在事务中，ID已经预先生成）
		if err := tx.Table(table.Name).Create(&processedData).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "创建失败", err)
		}

		// 执行after钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "end", processedData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 获取创建记录的ID（已在前面生成）
	//recordID := newID

	// 直接返回创建的记录数据（不通过 GetOne 查询，避免 IS_ACTIVE 过滤）
	// 因为刚创建的记录可能 IS_ACTIVE = 'N'，GetOne 会过滤掉
	return processedData, nil
}

// Update 更新记录
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查更新权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermUpdate)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "无修改权限")
	}

	// 添加ID到数据中供钩子使用
	data["ID"] = id

	// 获取字段定义（在事务外）
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return err
	}

	// 验证和处理字段（在事务外）
	processedData, err := s.processFieldsForUpdate(columns, data, userID)
	if err != nil {
		return err
	}

	// 添加审计字段
	// 获取用户信息以填充审计字段
	user, err := s.userRepo.GetUserByID(userID)
	if err == nil && user != nil {
		// 设置更新人
		processedData["UPDATE_BY"] = user.Username
	}
	// 设置更新时间
	processedData["UPDATE_TIME"] = time.Now()

	// 获取要更新的字段列表（支持零值更新）
	updateFields := make([]string, 0, len(processedData))
	for field := range processedData {
		updateFields = append(updateFields, field)
	}

	// 在事务中执行：before钩子 + 更新 + after钩子
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 执行before钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "begin", data); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
		}

		// 执行更新（在事务中，使用 Select 明确指定要更新的字段，包括零值）
		// 注意：不再过滤 IS_ACTIVE = 'Y'，允许更新任何状态的记录
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

		// 执行after钩子（在事务中）
		processedData["ID"] = id
		if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "end", processedData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
		}

		return nil
	})

	return err
}

// Delete 删除记录（物理删除）
func (s *service) Delete(ctx context.Context, tableName string, id uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查删除权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermDelete)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "无删除权限")
	}

	// 检查外键引用（REF_ON_DELETE 保护）
	if err := s.checkForeignKeyReferences(ctx, table.ID, id); err != nil {
		return err
	}

	// 在事务中执行：before钩子 + 删除 + after钩子
	deleteData := map[string]interface{}{"ID": id}
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
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

	// 检查删除权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermDelete)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "无删除权限")
	}

	// 检查每个ID的外键引用
	for _, id := range ids {
		if err := s.checkForeignKeyReferences(ctx, table.ID, id); err != nil {
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
