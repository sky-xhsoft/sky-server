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
	query := db.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND IS_ACTIVE = ?", "folder", name, spec.OwnerID, "Y")
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
