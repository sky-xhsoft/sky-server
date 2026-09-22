package crud

import (
	"context"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/mask"
	"github.com/sky-xhsoft/sky-server/internal/pkg/transaction"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"gorm.io/gorm"
)

// Submit 提交记录
func (s *service) Submit(ctx context.Context, tableName string, id uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查提交权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermSubmit)
	if err != nil {
		return errors.Wrap(errors.ErrPermissionDenied, "检查权限失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "没有提交权限")
	}

	// 检查表是否支持提交操作
	if !mask.HasAction(table.Mask, mask.ActionSubmit) {
		return errors.New(errors.ErrInvalidParam, "该表不支持提交操作")
	}

	// 查询记录是否存在
	var record map[string]interface{}
	if err := s.db.Table(tableName).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").Take(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrResourceNotFound, "记录不存在")
		}
		return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
	}

	// 在事务中执行提交操作
	return transaction.WithTransaction(s.db, func(tx *gorm.DB) error {
		// 执行 begin 钩子
		if err := s.executeHooksInTx(ctx, tx, table.ID, "S", "begin", record); err != nil {
			return errors.Wrap(errors.ErrActionExecute, "执行提交前钩子失败", err)
		}

		// 更新记录状态（如果表有 STATUS 字段）
		columns, err := s.metadataService.GetColumns(table.ID)
		if err != nil {
			return errors.Wrap(errors.ErrDatabase, "获取列信息失败", err)
		}

		hasStatusField := false
		for _, col := range columns {
			if col.DbName == "STATUS" {
				hasStatusField = true
				break
			}
		}

		if hasStatusField {
			if err := tx.Table(tableName).Where("ID = ?", id).Updates(map[string]interface{}{
				"STATUS":      "submitted",
				"UPDATE_BY":   userID,
				"UPDATE_TIME": time.Now(),
			}).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "更新记录状态失败", err)
			}
			record["STATUS"] = "submitted"
		}

		// 执行 end 钩子
		if err := s.executeHooksInTx(ctx, tx, table.ID, "S", "end", record); err != nil {
			return errors.Wrap(errors.ErrActionExecute, "执行提交后钩子失败", err)
		}

		return nil
	})
}

// Unsubmit 反提交记录
func (s *service) Unsubmit(ctx context.Context, tableName string, id uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查反提交权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermUnsubmit)
	if err != nil {
		return errors.Wrap(errors.ErrPermissionDenied, "检查权限失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "没有反提交权限")
	}

	// 检查表是否支持反提交操作
	if !mask.HasAction(table.Mask, mask.ActionUnsubmit) {
		return errors.New(errors.ErrInvalidParam, "该表不支持反提交操作")
	}

	// 查询记录是否存在
	var record map[string]interface{}
	if err := s.db.Table(tableName).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").Take(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrResourceNotFound, "记录不存在")
		}
		return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
	}

	// 在事务中执行反提交操作
	return transaction.WithTransaction(s.db, func(tx *gorm.DB) error {
		// 执行 begin 钩子
		if err := s.executeHooksInTx(ctx, tx, table.ID, "U", "begin", record); err != nil {
			return errors.Wrap(errors.ErrActionExecute, "执行反提交前钩子失败", err)
		}

		// 更新记录状态（如果表有 STATUS 字段）
		columns, err := s.metadataService.GetColumns(table.ID)
		if err != nil {
			return errors.Wrap(errors.ErrDatabase, "获取列信息失败", err)
		}

		hasStatusField := false
		for _, col := range columns {
			if col.DbName == "STATUS" {
				hasStatusField = true
				break
			}
		}

		if hasStatusField {
			if err := tx.Table(tableName).Where("ID = ?", id).Updates(map[string]interface{}{
				"STATUS":      "draft",
				"UPDATE_BY":   userID,
				"UPDATE_TIME": time.Now(),
			}).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "更新记录状态失败", err)
			}
			record["STATUS"] = "draft"
		}

		// 执行 end 钩子
		if err := s.executeHooksInTx(ctx, tx, table.ID, "U", "end", record); err != nil {
			return errors.Wrap(errors.ErrActionExecute, "执行反提交后钩子失败", err)
		}

		return nil
	})
}

// Void 作废记录
func (s *service) Void(ctx context.Context, tableName string, id uint, userID uint) error {
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermDelete)
	if err != nil {
		return errors.Wrap(errors.ErrPermissionDenied, "检查权限失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "没有作废权限")
	}

	if !mask.HasAction(table.Mask, mask.ActionVoid) {
		return errors.New(errors.ErrInvalidParam, "该表不支持作废操作")
	}

	var record map[string]interface{}
	if err := s.db.Table(tableName).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").Take(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrResourceNotFound, "记录不存在或已失效")
		}
		return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
	}

	return transaction.WithTransaction(s.db, func(tx *gorm.DB) error {
		if err := s.executeHooksInTx(ctx, tx, table.ID, "V", "begin", record); err != nil {
			return errors.Wrap(errors.ErrActionExecute, "执行作废前钩子失败", err)
		}

		if err := tx.Table(tableName).Where("ID = ?", id).Updates(map[string]interface{}{
			"STATUS":      "voided",
			"IS_ACTIVE":   "N",
			"UPDATE_BY":   userID,
			"UPDATE_TIME": time.Now(),
		}).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "作废记录失败", err)
		}

		if err := s.executeHooksInTx(ctx, tx, table.ID, "V", "end", record); err != nil {
			return errors.Wrap(errors.ErrActionExecute, "执行作废后钩子失败", err)
		}

		return nil
	})
}
