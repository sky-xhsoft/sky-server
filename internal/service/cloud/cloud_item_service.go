package cloud

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
)

// CloudItem 相关方法

// CreateItem 创建项目（文件或文件夹）
func (s *service) CreateItem(ctx context.Context, item *entity.CloudItem) error {
	return s.db.WithContext(ctx).Create(item).Error
}

// GetItem 获取项目（文件或文件夹）
func (s *service) GetItem(ctx context.Context, itemID uint, userID uint) (*entity.CloudItem, error) {
	var item entity.CloudItem
	err := s.db.WithContext(ctx).
		Where("ID = ? AND OWNER_ID = ? AND IS_ACTIVE = ?", itemID, userID, "Y").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ListItems 列出项目（文件+文件夹）
func (s *service) ListItems(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudItem, error) {
	var items []*entity.CloudItem
	query := s.db.WithContext(ctx).
		Where("OWNER_ID = ? AND IS_ACTIVE = ?", userID, "Y")

	if parentID == nil {
		query = query.Where("PARENT_ID IS NULL")
	} else {
		query = query.Where("PARENT_ID = ?", *parentID)
	}

	err := query.Order("ITEM_TYPE DESC, NAME ASC").Find(&items).Error // 文件夹优先
	if err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateItem 更新项目
func (s *service) UpdateItem(ctx context.Context, item *entity.CloudItem) error {
	return s.db.WithContext(ctx).
		Where("ID = ? AND OWNER_ID = ? AND IS_ACTIVE = ?", item.ID, item.OwnerID, "Y").
		Updates(item).Error
}

// DeleteItem 删除项目（软删除）
func (s *service) DeleteItem(ctx context.Context, itemID uint, userID uint) error {
	// 先获取项目信息
	item, err := s.GetItem(ctx, itemID, userID)
	if err != nil {
		return err
	}

	// 如果是文件，删除物理存储
	if item.IsFile() && item.StoragePath != nil {
		_ = s.storage.Delete(ctx, *item.StoragePath)
	}

	// 软删除数据库记录
	return s.db.WithContext(ctx).
		Model(&entity.CloudItem{}).
		Where("ID = ? AND OWNER_ID = ?", itemID, userID).
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
				_ = s.storage.Delete(ctx, *child.StoragePath)
			}
		}

		// 删除当前项目
		if err := s.DeleteItem(ctx, child.ID, userID); err != nil {
			return err
		}

		// 更新配额
		if child.IsFile() && child.FileSize != nil {
			_ = s.UpdateQuota(ctx, userID, -*child.FileSize, -1, 0)
		} else if child.IsFolder() {
			_ = s.UpdateQuota(ctx, userID, 0, 0, -1)
		}
	}

	return nil
}

// MoveItem 移动项目到新的父文件夹
func (s *service) MoveItem(ctx context.Context, itemID uint, targetParentID *uint, userID uint) error {
	// 检查项目是否存在
	item, err := s.GetItem(ctx, itemID, userID)
	if err != nil {
		return errors.ResourceNotFound
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

	// 更新父文件夹
	item.ParentID = targetParentID
	// 重新计算路径
	if targetParentID == nil {
		item.Path = "/" + item.Name
	} else {
		var parent entity.CloudItem
		if err := s.db.WithContext(ctx).First(&parent, *targetParentID).Error; err != nil {
			return err
		}
		item.Path = parent.Path + "/" + item.Name
	}

	return s.UpdateItem(ctx, item)
}

// RenameItem 重命名项目
func (s *service) RenameItem(ctx context.Context, itemID uint, newName string, userID uint) error {
	item, err := s.GetItem(ctx, itemID, userID)
	if err != nil {
		return errors.ResourceNotFound
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

	// 更新名称和路径
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

	return s.UpdateItem(ctx, item)
}

// GetStorage 获取存储引擎（供 handler 使用）
func (s *service) GetStorage() storage.Storage {
	return s.storage
}
