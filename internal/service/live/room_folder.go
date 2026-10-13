package live

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 直播间云盘目录下的子目录
const (
	RoomFolderRecording = "直播录制"
	RoomFolderClip      = "直播切片"
)

// RoomFolderSpec 直播间云盘目录的归属与存储位置
type RoomFolderSpec struct {
	RoomName        string
	CreateBy        string
	CompanyID       uint
	OwnerID         uint  // 直播间创建人的用户 ID，目录属于该用户
	StorageFolderID *uint // 本次目录的存储位置（录制或切片各自配置的云盘文件夹 ID），为空表示云盘根目录
}

// RoomCreatorID 查找直播间创建人的用户 ID。用户名只在公司内唯一（不同公司可以同名），因此：
//  1. 先在直播间所属公司内按用户名查找；
//  2. 公司内没有（如创建人已调到其他公司）时，只有全库唯一一个同名用户才使用，同名多个时不猜，返回 false。
//
// 不按公司区分会把云盘目录与文件记到其他公司的同名用户名下（该用户看不到，另一家公司反而能看到）。
func RoomCreatorID(db *gorm.DB, username string, companyID uint) (uint, bool) {
	if username == "" {
		return 0, false
	}
	var users []entity.SysUser
	if err := db.Select("ID").Where("USERNAME = ? AND SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", username, companyID, "Y").
		Limit(1).Find(&users).Error; err == nil && len(users) == 1 {
		return users[0].ID, true
	}
	if err := db.Select("ID").Where("USERNAME = ? AND IS_ACTIVE = ?", username, "Y").
		Limit(2).Find(&users).Error; err == nil && len(users) == 1 {
		return users[0].ID, true
	}
	return 0, false
}

// EnsureRoomFolder 确保「存储位置/直播间名称[/子目录]」存在，返回最末一级目录。
// 录制与切片可配置不同的存储位置，调用方按子目录传入对应的 StorageFolderID。
// 存储位置须为直播间创建人自己的有效文件夹，否则放在云盘根目录，避免写入他人目录。
func EnsureRoomFolder(db *gorm.DB, spec RoomFolderSpec, sub string) (*entity.CloudItem, error) {
	parent := resolveStorageFolder(db, spec)
	roomFolder, err := ensureFolder(db, spec, parent, spec.RoomName)
	if err != nil || sub == "" {
		return roomFolder, err
	}
	return ensureFolder(db, spec, roomFolder, sub)
}

// resolveStorageFolder 校验直播间配置的存储位置；无效时返回 nil（云盘根目录）
func resolveStorageFolder(db *gorm.DB, spec RoomFolderSpec) *entity.CloudItem {
	if spec.StorageFolderID == nil || *spec.StorageFolderID == 0 {
		return nil
	}
	var folder entity.CloudItem
	err := db.Where("ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ? AND OWNER_ID = ? AND SYS_COMPANY_ID IN (?, 0)",
		*spec.StorageFolderID, "folder", "Y", spec.OwnerID, spec.CompanyID).First(&folder).Error
	if err != nil {
		logger.Warn("直播间存储位置无效（不存在、已删除或不属于直播间创建人），改为云盘根目录",
			zap.String("roomName", spec.RoomName),
			zap.Uint("storageFolderId", *spec.StorageFolderID),
			zap.Uint("ownerId", spec.OwnerID))
		return nil
	}
	return &folder
}

// ensureFolder 在 parent（nil 为根目录）下查找或创建名为 name 的文件夹
func ensureFolder(db *gorm.DB, spec RoomFolderSpec, parent *entity.CloudItem, name string) (*entity.CloudItem, error) {
	query := db.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND SYS_COMPANY_ID = ? AND IS_ACTIVE = ?",
		"folder", name, spec.OwnerID, spec.CompanyID, "Y")
	path := "/" + name
	var parentID *uint
	if parent != nil {
		id := parent.ID
		parentID = &id
		path = parent.Path + "/" + name
		query = query.Where("PARENT_ID = ?", id)
	} else {
		query = query.Where("PARENT_ID IS NULL")
	}

	var folder entity.CloudItem
	if err := query.First(&folder).Error; err == nil {
		return &folder, nil
	}

	folder = entity.CloudItem{
		BaseModel: entity.BaseModel{
			SysCompanyID: spec.CompanyID,
			CreateBy:     spec.CreateBy,
			UpdateBy:     spec.CreateBy,
			IsActive:     "Y",
		},
		ItemType: "folder",
		Name:     name,
		ParentID: parentID,
		Path:     path,
		OwnerID:  spec.OwnerID,
	}
	if err := db.Create(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}
