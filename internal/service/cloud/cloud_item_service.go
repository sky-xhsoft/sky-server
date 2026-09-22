package cloud

import (
	"context"

	"gorm.io/gorm"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"go.uber.org/zap"
)

// CloudItem 相关方法

// CreateItem 创建项目（文件或文件夹）— 写入公司归属 (SysCompanyID)
func (s *service) CreateItem(ctx context.Context, item *entity.CloudItem) error {
	if item.OwnerID != 0 {
		item.SysCompanyID = s.getUserCompanyID(ctx, item.OwnerID)
	} else if item.SysCompanyID == 0 {
		item.SysCompanyID = 0 // 无属主则公用
	}
	return s.db.WithContext(ctx).Create(item).Error
}

// GetItem 获取项目（文件或文件夹）— 公司隔离 + 用户权限判定
func (s *service) GetItem(ctx context.Context, itemID uint, userID uint) (*entity.CloudItem, error) {
	companyID := s.getUserCompanyID(ctx, userID)
	var item entity.CloudItem
	err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", itemID, "Y").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	// 权限判定：所有者 或 权限表命中读，且满足公司范围
	if !s.canAccessRead(ctx, &item, userID, companyID) {
		return nil, errors.PermissionDenied
	}
	return &item, nil
}

// ListItems 列出项目（文件+文件夹）— 公司隔离 + 用户权限过滤
func (s *service) ListItems(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudItem, error) {
	companyID := s.getUserCompanyID(ctx, userID)
	var allItems []*entity.CloudItem
	query := s.db.WithContext(ctx).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y")

	if parentID == nil {
		query = query.Where("PARENT_ID IS NULL")
	} else {
		query = query.Where("PARENT_ID = ?", *parentID)
	}

	if err := query.Order("ITEM_TYPE DESC, NAME ASC").Find(&allItems).Error; err != nil {
		return nil, err
	}

	// 过滤：所有者 或 权限表命中读
	items := make([]*entity.CloudItem, 0, len(allItems))
	for _, it := range allItems {
		if s.canAccessRead(ctx, it, userID, companyID) {
			items = append(items, it)
		}
	}
	return items, nil
}

// UpdateItem 更新项目 — 写权限判定
func (s *service) UpdateItem(ctx context.Context, item *entity.CloudItem) error {
	// 写权限判定（item.OwnerID 为操作者）
	hasWrite, err := s.HasItemWritePermission(ctx, item.ID, item.OwnerID)
	if err != nil || !hasWrite {
		return errors.PermissionDenied
	}
	return s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", item.ID, "Y").
		Updates(item).Error
}

// DeleteItem 删除项目（软删除）
func (s *service) DeleteItem(ctx context.Context, itemID uint, userID uint) error {
	// 先获取项目信息（会做读权限判定）
	item, err := s.GetItem(ctx, itemID, userID)
	if err != nil {
		return err
	}

	// 写权限判定：所有者或权限表命中写
	hasWrite, err := s.HasItemWritePermission(ctx, itemID, userID)
	if err != nil || !hasWrite {
		return errors.PermissionDenied
	}

	// 如果是文件，删除物理存储
	if item.IsFile() && item.StoragePath != nil {
		// 获取公司存储实例
		storageInstance, err := s.getCompanyStorage(ctx, userID)
		if err == nil {
			if err := storageInstance.Delete(ctx, *item.StoragePath); err != nil {
				logger.Warn("Storage delete failed for item", zap.Uint("itemId", item.ID), zap.Error(err))
			}
		}
	}

	// 软删除数据库记录
	return s.db.WithContext(ctx).
		Model(&entity.CloudItem{}).
		Where("ID = ? AND IS_ACTIVE = ?", itemID, "Y").
		Update("IS_ACTIVE", "N").Error
}

// DeleteItemsByParentID 递归删除子项目
func (s *service) DeleteItemsByParentID(ctx context.Context, parentID uint, userID uint) error {
	// 查找所有子项目
	var children []*entity.CloudItem
	err := s.db.WithContext(ctx).
		Where("PARENT_ID = ? AND OWNER_ID = ? AND IS_ACTIVE = ?", parentID, userID, "Y").
		Find(&children).Error
	if err != nil {
		return err
	}

	// 递归删除子项目
	for _, child := range children {
		if child.IsFolder() {
			// 如果是文件夹，递归删除其子项目
			if err := s.DeleteItemsByParentID(ctx, child.ID, userID); err != nil {
				return err
			}
		} else if child.IsFile() {
			// 如果是文件，删除存储
			if child.StoragePath != nil {
				// 获取公司存储实例
				storageInstance, err := s.getCompanyStorage(ctx, userID)
				if err == nil {
					if err := storageInstance.Delete(ctx, *child.StoragePath); err != nil {
						logger.Warn("Storage delete failed for child item", zap.Uint("itemId", child.ID), zap.Error(err))
					}
				}
			}
		}

		// 删除当前项目
		if err := s.DeleteItem(ctx, child.ID, userID); err != nil {
			return err
		}

		// 更新配额
		if child.IsFile() && child.FileSize != nil {
			if err := s.UpdateQuota(ctx, userID, -*child.FileSize, -1, 0); err != nil {
				logger.Warn("UpdateQuota failed for file", zap.Uint("userId", userID), zap.Error(err))
			}
		} else if child.IsFolder() {
			if err := s.UpdateQuota(ctx, userID, 0, 0, -1); err != nil {
				logger.Warn("UpdateQuota failed for folder", zap.Uint("userId", userID), zap.Error(err))
			}
		}
	}

	return nil
}

// MoveItem 移动项目到新的父文件夹
func (s *service) MoveItem(ctx context.Context, itemID uint, targetParentID *uint, userID uint) error {
	// 检查项目是否存在（含读权限判定）
	item, err := s.GetItem(ctx, itemID, userID)
	if err != nil {
		return errors.ResourceNotFound
	}

	// 写权限判定
	hasWrite, err := s.HasItemWritePermission(ctx, itemID, userID)
	if err != nil || !hasWrite {
		return errors.PermissionDenied
	}

	// 检查目标文件夹是否存在（如果不是根目录）
	if targetParentID != nil && *targetParentID != 0 {
		targetFolder, err := s.GetItem(ctx, *targetParentID, userID)
		if err != nil {
			return errors.New(errors.ErrResourceNotFound, "目标文件夹不存在")
		}
		if !targetFolder.IsFolder() {
			return errors.New(errors.ErrInvalidParam, "目标必须是文件夹")
		}
	}

	// 如果是移动到根目录，设置为 nil
	if targetParentID != nil && *targetParentID == 0 {
		targetParentID = nil
	}

	// 计算新路径
	var newPath string
	if targetParentID == nil {
		newPath = "/" + item.Name
	} else {
		var parent entity.CloudItem
		if err := s.db.WithContext(ctx).First(&parent, *targetParentID).Error; err != nil {
			return err
		}
		newPath = parent.Path + "/" + item.Name
	}

	// 使用事务确保原子性
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		oldPath := item.Path

		// 更新项目本身
		item.ParentID = targetParentID
		item.Path = newPath
		if err := tx.Save(item).Error; err != nil {
			return err
		}

		// 如果是文件夹，递归更新所有子项目的路径
		if item.IsFolder() {
			// 查询所有子项目（路径以 oldPath/ 开头）
			var children []entity.CloudItem
			if err := tx.Where("PATH LIKE ?", oldPath+"/%").Find(&children).Error; err != nil {
				return err
			}

			// 批量更新子项目路径
			for i := range children {
				children[i].Path = newPath + children[i].Path[len(oldPath):]
				if err := tx.Save(&children[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// RenameItem 重命名项目
func (s *service) RenameItem(ctx context.Context, itemID uint, newName string, userID uint) error {
	item, err := s.GetItem(ctx, itemID, userID)
	if err != nil {
		return errors.ResourceNotFound
	}

	// 写权限判定
	hasWrite, err := s.HasItemWritePermission(ctx, itemID, userID)
	if err != nil || !hasWrite {
		return errors.PermissionDenied
	}

	// 检查同级是否已存在同名项目
	var count int64
	query := s.db.WithContext(ctx).
		Model(&entity.CloudItem{}).
		Where("NAME = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ID != ?", newName, userID, "Y", itemID)

	if item.ParentID == nil {
		query = query.Where("PARENT_ID IS NULL")
	} else {
		query = query.Where("PARENT_ID = ?", *item.ParentID)
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New(30002, "同名项目已存在")
	}

	oldPath := item.Path
	oldName := item.Name
	item.Name = newName

	// 重新计算路径
	if item.ParentID == nil {
		item.Path = "/" + newName
	} else {
		// 替换路径中的最后一部分
		parentPath := item.Path[:len(item.Path)-len(oldName)-1]
		item.Path = parentPath + "/" + newName
	}
	newPath := item.Path

	// 使用事务确保原子性
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新项目本身
		if err := tx.Save(item).Error; err != nil {
			return err
		}

		// 如果是文件夹，递归更新所有子项目的路径
		if item.IsFolder() {
			// 查询所有子项目（路径以 oldPath/ 开头）
			var children []entity.CloudItem
			if err := tx.Where("PATH LIKE ?", oldPath+"/%").Find(&children).Error; err != nil {
				return err
			}

			// 批量更新子项目路径
			for i := range children {
				children[i].Path = newPath + children[i].Path[len(oldPath):]
				if err := tx.Save(&children[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// GetStorage 获取存储引擎（供 handler 使用，已废弃）
// 注意：建议使用 getCompanyStorage 方法获取用户对应的存储实例
func (s *service) GetStorage() storage.Storage {
	// 由于不再直接存储 storage 字段，这里返回 nil 或默认存储
	// 但为了兼容性，我们可以尝试返回一个默认存储
	return nil
}
