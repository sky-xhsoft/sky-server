package cloud

import (
	"context"
	stderrors "errors"
	"strings"

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
		// 调用方已指定公司时保留（如直播回调按直播间所在公司归属），否则取属主所在公司
		if item.SysCompanyID == 0 {
			item.SysCompanyID = s.getUserCompanyID(ctx, item.OwnerID)
		}
		if item.CreateBy == "" {
			item.CreateBy = s.getUsernameByID(ctx, item.OwnerID)
			item.UpdateBy = item.CreateBy
		}
	} else if item.SysCompanyID == 0 {
		item.SysCompanyID = 0 // 无属主则公用
	}
	if err := s.db.WithContext(ctx).Create(item).Error; err != nil {
		if stderrors.Is(err, entity.ErrCloudFolderExists) {
			return errors.New(errors.ErrResourceExists, "同名文件夹已存在")
		}
		return err
	}
	return nil
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

	// 能读取父文件夹时其内容全部可见（授权继承，也避免逐条查询权限）
	if parentID != nil {
		var parent entity.CloudItem
		if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", *parentID, "Y").First(&parent).Error; err == nil &&
			s.canAccessRead(ctx, &parent, userID, companyID) {
			return allItems, nil
		}
	}

	// 过滤：所有者 或 权限表命中读（含上级授权）
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

	// 如果是文件，删除物理存储（秒传的多条记录共用同一存储对象，仍被引用时保留）
	if item.IsFile() {
		s.deleteStorageIfUnreferenced(ctx, item)
	}

	// 软删除数据库记录
	return s.db.WithContext(ctx).
		Model(&entity.CloudItem{}).
		Where("ID = ? AND IS_ACTIVE = ?", itemID, "Y").
		Update("IS_ACTIVE", "N").Error
}

// DeleteItemsByParentID 递归删除子项目
func (s *service) DeleteItemsByParentID(ctx context.Context, parentID uint, userID uint) error {
	return s.deleteChildren(ctx, parentID, userID, map[uint]bool{parentID: true})
}

// deleteChildren 递归删除子项目；visited 防止异常数据（环）导致无限递归
func (s *service) deleteChildren(ctx context.Context, parentID uint, userID uint, visited map[uint]bool) error {
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
		if visited[child.ID] {
			continue
		}
		visited[child.ID] = true
		if child.IsFolder() {
			// 如果是文件夹，递归删除其子项目
			if err := s.deleteChildren(ctx, child.ID, userID, visited); err != nil {
				return err
			}
		}
		// 文件的物理存储由 DeleteItem 统一处理

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

	// 文件夹不能移动到自身或自己的子文件夹下（否则形成环，删除时会无限递归）
	if item.IsFolder() && targetParentID != nil {
		inside, err := s.isSelfOrDescendant(ctx, *targetParentID, item.ID)
		if err != nil {
			return err
		}
		if inside {
			return errors.New(errors.ErrInvalidParam, "不能将文件夹移动到其自身或子文件夹中")
		}
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

	// 使用事务确保原子性；只更新父目录与路径，不整行回写（避免覆盖创建人等字段）
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		oldPath := item.Path

		if err := tx.Model(&entity.CloudItem{}).Where("ID = ?", item.ID).
			Updates(map[string]interface{}{"PARENT_ID": targetParentID, "PATH": newPath}).Error; err != nil {
			return err
		}

		// 文件夹：按父子关系收集子孙条目（只含同一所有者），逐个改写路径前缀
		if item.IsFolder() {
			descendants, err := collectDescendants(tx, item.ID, item.OwnerID)
			if err != nil {
				return err
			}
			for _, child := range descendants {
				if !strings.HasPrefix(child.Path, oldPath+"/") {
					continue
				}
				if err := tx.Model(&entity.CloudItem{}).Where("ID = ?", child.ID).
					Update("PATH", newPath+child.Path[len(oldPath):]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// maxFolderDepth 目录层级上限，防止异常数据（环）导致死循环
const maxFolderDepth = 64

// isSelfOrDescendant 判断 folderID 是否就是 ancestorID 或位于其下
func (s *service) isSelfOrDescendant(ctx context.Context, folderID, ancestorID uint) (bool, error) {
	current := folderID
	for depth := 0; depth < maxFolderDepth; depth++ {
		if current == ancestorID {
			return true, nil
		}
		var parent struct {
			ParentID *uint `gorm:"column:PARENT_ID"`
		}
		res := s.db.WithContext(ctx).Table("cloud_item").Select("PARENT_ID").Where("ID = ?", current).Limit(1).Scan(&parent)
		if res.Error != nil {
			return false, res.Error
		}
		if res.RowsAffected == 0 || parent.ParentID == nil {
			return false, nil
		}
		current = *parent.ParentID
	}
	// 层级异常深（可能已有环），按“在其下”处理以拒绝移动
	return true, nil
}

// collectDescendants 按父子关系收集文件夹下的全部子孙条目（同一所有者），已访问集合防止环
func collectDescendants(tx *gorm.DB, folderID, ownerID uint) ([]entity.CloudItem, error) {
	var all []entity.CloudItem
	visited := map[uint]bool{folderID: true}
	queue := []uint{folderID}
	for len(queue) > 0 {
		var children []entity.CloudItem
		if err := tx.Select("ID, PATH, ITEM_TYPE").Where("PARENT_ID IN ? AND OWNER_ID = ? AND IS_ACTIVE = ?", queue, ownerID, "Y").
			Find(&children).Error; err != nil {
			return nil, err
		}
		queue = queue[:0]
		for _, child := range children {
			if visited[child.ID] {
				continue
			}
			visited[child.ID] = true
			all = append(all, child)
			if child.IsFolder() {
				queue = append(queue, child.ID)
			}
		}
	}
	return all, nil
}

// RenameItem 重命名项目
func (s *service) RenameItem(ctx context.Context, itemID uint, newName string, userID uint) error {
	newName, err := ValidateItemName(newName)
	if err != nil {
		return err
	}
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
	item.Name = newName

	// 按父目录重新计算路径（不从旧路径截取：旧数据路径可能为空或不一致）
	item.Path = itemChildPath(s.db.WithContext(ctx), item.ParentID, newName)
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

// deleteStorageIfUnreferenced 删除文件的存储对象；仍有其他有效记录引用同一存储路径（秒传）时保留。
// 存储对象位于文件所有者所属公司的存储中
func (s *service) deleteStorageIfUnreferenced(ctx context.Context, item *entity.CloudItem) {
	if item.StoragePath == nil || *item.StoragePath == "" {
		return
	}
	var refs int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("STORAGE_PATH = ? AND IS_ACTIVE = ? AND ID <> ?", *item.StoragePath, "Y", item.ID).
		Count(&refs).Error; err != nil || refs > 0 {
		return
	}
	storageInstance, err := s.getItemStorage(ctx, item)
	if err != nil {
		return
	}
	if err := storageInstance.Delete(ctx, *item.StoragePath); err != nil {
		logger.Warn("Storage delete failed for item", zap.Uint("itemId", item.ID), zap.Error(err))
	}
}

// itemChildPath 条目在父目录下的完整路径：父目录 PATH + "/" + 名称；根目录（或父目录无路径）为 "/" + 名称
func itemChildPath(db *gorm.DB, parentID *uint, name string) string {
	if parentID != nil {
		var parentPath string
		db.Table("cloud_item").Select("PATH").Where("ID = ?", *parentID).Scan(&parentPath)
		if parentPath != "" {
			return strings.TrimSuffix(parentPath, "/") + "/" + name
		}
	}
	return "/" + name
}
