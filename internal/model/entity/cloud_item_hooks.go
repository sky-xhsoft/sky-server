package entity

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// ErrCloudFolderExists 同一目录下已存在同名文件夹
var ErrCloudFolderExists = errors.New("同名文件夹已存在")

// maxDuplicateSuffix 同名文件自动改名的最大序号
const maxDuplicateSuffix = 1000

// BeforeCreate 同一所有者、同一目录下名称唯一：
// 文件夹重名直接拒绝；文件重名自动改为“名称 (n).扩展名”（上传、分片合并、录制回调等所有入口统一生效）
func (c *CloudItem) BeforeCreate(tx *gorm.DB) error {
	if c.Name == "" || (c.IsActive != "" && c.IsActive != "Y") {
		return nil
	}
	db := tx.Session(&gorm.Session{NewDB: true})
	exists := func(name string) (bool, error) {
		q := db.Model(&CloudItem{}).Where("NAME = ? AND OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", name, c.OwnerID, c.ItemType, "Y")
		if c.ParentID == nil {
			q = q.Where("PARENT_ID IS NULL")
		} else {
			q = q.Where("PARENT_ID = ?", *c.ParentID)
		}
		var n int64
		err := q.Count(&n).Error
		return n > 0, err
	}

	taken, err := exists(c.Name)
	if err != nil || !taken {
		return err
	}
	if c.IsFolder() {
		return ErrCloudFolderExists
	}

	ext := filepath.Ext(c.Name)
	base := strings.TrimSuffix(c.Name, ext)
	for i := 1; i <= maxDuplicateSuffix; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		taken, err := exists(candidate)
		if err != nil {
			return err
		}
		if !taken {
			if strings.HasSuffix(c.Path, "/"+c.Name) {
				c.Path = strings.TrimSuffix(c.Path, c.Name) + candidate
			}
			c.Name = candidate
			return nil
		}
	}
	return nil
}
