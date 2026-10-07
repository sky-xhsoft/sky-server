package groups

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// fillDirectoryParents 按分类推导父目录：表目录的父级为同分类的分类目录
func fillDirectoryParents(dirs []*entity.SysDirectory) {
	categoryDir := make(map[uint]uint)
	for _, d := range dirs {
		if d.IsCategory() && d.SysTableCategoryID != nil {
			if _, exists := categoryDir[*d.SysTableCategoryID]; !exists {
				categoryDir[*d.SysTableCategoryID] = d.ID
			}
		}
	}
	for _, d := range dirs {
		d.ParentID = nil
		if d.IsCategory() || d.SysTableCategoryID == nil {
			continue
		}
		if parentID, ok := categoryDir[*d.SysTableCategoryID]; ok && parentID != d.ID {
			pid := parentID
			d.ParentID = &pid
		}
	}
}

// categoryOf 返回目录所属分类；目录不存在或无分类时 ok=false
func (s *service) categoryOf(ctx context.Context, dirID uint) (uint, bool) {
	var dir entity.SysDirectory
	if err := s.db.WithContext(ctx).Select("ID, SYS_TABLE_CATEGORY_ID").
		Where("ID = ? AND IS_ACTIVE = ?", dirID, "Y").First(&dir).Error; err != nil || dir.SysTableCategoryID == nil {
		return 0, false
	}
	return *dir.SysTableCategoryID, true
}

// whereChildrenOf 限定为某分类目录下的表目录
func whereChildrenOf(query *gorm.DB, category uint) *gorm.DB {
	return query.Where("SYS_TABLE_CATEGORY_ID = ? AND SYS_TABLE_ID IS NOT NULL AND SYS_TABLE_ID <> 0", category)
}
