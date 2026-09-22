package cloud

import (
	"context"

	"gorm.io/gorm"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// SetItemPermissionRequest 设置项目权限请求
type SetItemPermissionRequest struct {
	TargetType string `json:"targetType" binding:"required,oneof=user group role"` // user / group / role
	TargetID   uint   `json:"targetId" binding:"required"`                         // 对象ID
	PermRead   bool   `json:"permRead"`                                            // 读权限
	PermWrite  bool   `json:"permWrite"`                                           // 写权限
}

// SetItemPermission 设置/更新项目权限（仅所有者或拥有写权限者）
func (s *service) SetItemPermission(ctx context.Context, itemID uint, req *SetItemPermissionRequest, userID uint) error {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	// 仅所有者可授权
	if item.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "仅所有者可设置权限")
	}

	perm := &entity.CloudItemPermission{
		ItemID:       itemID,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		PermRead:     "N",
		PermWrite:    "N",
		SysCompanyID: item.SysCompanyID,
		CreateBy:     s.getUsernameByID(ctx, userID),
		UpdateBy:     s.getUsernameByID(ctx, userID),
		IsActive:     "Y",
	}
	if req.PermRead {
		perm.PermRead = "Y"
	}
	if req.PermWrite {
		perm.PermWrite = "Y"
	}

	// upsert：命中任意行（含隐藏）则更新；否则新建，避免 unique 冲突
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.CloudItemPermission
		err := tx.Where("ITEM_ID = ? AND TARGET_TYPE = ? AND TARGET_ID = ?",
			itemID, req.TargetType, req.TargetID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			return tx.Create(perm).Error
		}
		if err != nil {
			return err
		}
		existing.PermRead = perm.PermRead
		existing.PermWrite = perm.PermWrite
		existing.UpdateBy = perm.UpdateBy
		existing.SysCompanyID = item.SysCompanyID
		existing.IsActive = "Y"
		return tx.Save(&existing).Error
	})
}

// ListItemPermissions 列出项目权限（仅所有者）
func (s *service) ListItemPermissions(ctx context.Context, itemID uint, userID uint) ([]*entity.CloudItemPermission, error) {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.OwnerID != userID {
		return nil, errors.New(errors.ErrPermissionDenied, "仅所有者可查看权限")
	}
	var perms []*entity.CloudItemPermission
	if err := s.db.WithContext(ctx).
		Where("ITEM_ID = ? AND IS_ACTIVE = ?", itemID, "Y").
		Order("ID ASC").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// RemoveItemPermission 删除项目权限（真删除，避免 unique 索引被软删记录占用）
func (s *service) RemoveItemPermission(ctx context.Context, itemID uint, permID uint, userID uint) error {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "仅所有者可删除权限")
	}
	return s.db.WithContext(ctx).
		Where("ID = ? AND ITEM_ID = ?", permID, itemID).
		Delete(&entity.CloudItemPermission{}).Error
}

// HasItemWritePermission 判断用户是否有某项目的写权限（所有者或权限表命中写）
func (s *service) HasItemWritePermission(ctx context.Context, itemID uint, userID uint) (bool, error) {
	item, err := s.getItemByID(ctx, itemID)
	if err != nil {
		return false, err
	}
	if item.OwnerID == userID {
		return true, nil
	}
	var cnt int64
	if err := s.db.WithContext(ctx).
		Model(&entity.CloudItemPermission{}).
		Where("ITEM_ID = ? AND TARGET_TYPE = ? AND TARGET_ID = ? AND PERM_WRITE = ? AND IS_ACTIVE = ?",
			itemID, "user", userID, "Y", "Y").Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// canAccessRead 判断用户是否有项目读权限（所有者或权限表命中读，且同公司/公用）
func (s *service) canAccessRead(ctx context.Context, item *entity.CloudItem, userID uint, companyID uint) bool {
	if item.OwnerID == userID {
		return true
	}
	if !s.companyScopeOK(item.SysCompanyID, companyID) {
		return false
	}
	var cnt int64
	if err := s.db.WithContext(ctx).
		Model(&entity.CloudItemPermission{}).
		Where("ITEM_ID = ? AND TARGET_TYPE = ? AND TARGET_ID = ? AND PERM_READ = ? AND IS_ACTIVE = ?",
			item.ID, "user", userID, "Y", "Y").Count(&cnt).Error; err != nil {
		return false
	}
	return cnt > 0
}

// companyScopeOK 公司范围：公用(0/-1)对所有公司可见，否则必须同公司
func (s *service) companyScopeOK(itemCompany, userCompany uint) bool {
	if itemCompany == 0 || itemCompany == ^uint(0) || itemCompany == userCompany {
		return true
	}
	return false
}

// getUserCompanyID 根据用户ID获取所属公司ID
func (s *service) getUserCompanyID(ctx context.Context, userID uint) uint {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return 0
	}
	return user.SysCompanyID
}

// GetUserCompanyID 导出方法：根据用户ID获取所属公司ID
func (s *service) GetUserCompanyID(ctx context.Context, userID uint) uint {
	return s.getUserCompanyID(ctx, userID)
}

// getItemByID 按ID获取项目（不论属主）
func (s *service) getItemByID(ctx context.Context, itemID uint) (*entity.CloudItem, error) {
	var item entity.CloudItem
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", itemID, "Y").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
