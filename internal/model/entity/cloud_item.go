package entity

import "time"

// CloudItem 云盘项目（文件+文件夹统一）
type CloudItem struct {
	BaseModel
	ItemType string `gorm:"column:ITEM_TYPE;size:20;not null;index" json:"itemType"`      // 项目类型: file, folder
	Name     string `gorm:"column:NAME;size:255;not null" json:"name"`                    // 名称（文件名或文件夹名）
	ParentID *uint  `gorm:"column:PARENT_ID;index:idx_parent_type" json:"parentId"`       // 父文件夹ID
	Path     string `gorm:"column:PATH;size:1000;not null;index" json:"path"`             // 完整路径
	OwnerID  uint   `gorm:"column:OWNER_ID;index:idx_owner_type;not null" json:"ownerId"` // 所有者ID

	// 文件专用字段（文件夹时为NULL）
	StorageType   *string `gorm:"column:STORAGE_TYPE;size:20" json:"storageType,omitempty"`  // 存储类型: local, oss（仅文件）
	StoragePath   *string `gorm:"column:STORAGE_PATH;size:500" json:"storagePath,omitempty"` // 存储路径（仅文件）
	FileSize      *int64  `gorm:"column:FILE_SIZE" json:"fileSize,omitempty"`                // 文件大小（字节，仅文件）
	FileType      *string `gorm:"column:FILE_TYPE;size:100" json:"fileType,omitempty"`       // 文件MIME类型（仅文件）
	FileExt       *string `gorm:"column:FILE_EXT;size:20" json:"fileExt,omitempty"`          // 文件扩展名（仅文件）
	MD5           *string `gorm:"column:MD5;size:32;index" json:"md5,omitempty"`             // MD5值（仅文件）
	AccessURL     *string `gorm:"column:ACCESS_URL;size:500" json:"accessUrl,omitempty"`     // 访问URL（仅文件）
	Thumbnail     *string `gorm:"column:THUMBNAIL;size:500" json:"thumbnail,omitempty"`      // 缩略图URL（仅文件）
	DownloadCount int     `gorm:"column:DOWNLOAD_COUNT;default:0" json:"downloadCount"`      // 下载次数（仅文件）
	Tags          *string `gorm:"column:TAGS;size:500" json:"tags,omitempty"`                // 标签（逗号分隔，仅文件）

	// 文件夹专用字段（文件时为NULL）
	FileCount int   `gorm:"column:FILE_COUNT;default:0" json:"fileCount"` // 文件数量（仅文件夹）
	TotalSize int64 `gorm:"column:TOTAL_SIZE;default:0" json:"totalSize"` // 总大小（字节，仅文件夹）

	// 共用字段
	IsPublic    string     `gorm:"column:IS_PUBLIC;size:1;default:N" json:"isPublic"`        // 是否公开 Y/N
	ShareCode   *string    `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"`   // 分享码
	ShareExpire *time.Time `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`     // 分享过期时间
	Description string     `gorm:"column:DESCRIPTION;size:500" json:"description,omitempty"` // 描述
}

// TableName 指定表名
func (CloudItem) TableName() string {
	return "cloud_item"
}

// IsFile 判断是否为文件
func (c *CloudItem) IsFile() bool {
	return c.ItemType == "file"
}

// IsFolder 判断是否为文件夹
func (c *CloudItem) IsFolder() bool {
	return c.ItemType == "folder"
}

// ToFile 转换为文件（兼容旧接口）
func (c *CloudItem) ToFile() *CloudFile {
	if !c.IsFile() {
		return nil
	}

	file := &CloudFile{
		BaseModel:     c.BaseModel,
		FileName:      c.Name,
		FolderID:      c.ParentID,
		Path:          c.Path,
		OwnerID:       c.OwnerID,
		IsPublic:      c.IsPublic,
		ShareCode:     c.ShareCode,
		ShareExpire:   c.ShareExpire,
		DownloadCount: c.DownloadCount,
		Description:   c.Description,
	}

	if c.StorageType != nil {
		file.StorageType = *c.StorageType
	}
	if c.StoragePath != nil {
		file.StoragePath = *c.StoragePath
	}
	if c.FileSize != nil {
		file.FileSize = *c.FileSize
	}
	if c.FileType != nil {
		file.FileType = *c.FileType
	}
	if c.FileExt != nil {
		file.FileExt = *c.FileExt
	}
	if c.MD5 != nil {
		file.MD5 = *c.MD5
	}
	if c.AccessURL != nil {
		file.AccessURL = *c.AccessURL
	}
	if c.Thumbnail != nil {
		file.Thumbnail = *c.Thumbnail
	}
	if c.Tags != nil {
		file.Tags = *c.Tags
	}

	return file
}

// ToFolder 转换为文件夹（兼容旧接口）
func (c *CloudItem) ToFolder() *CloudFolder {
	if !c.IsFolder() {
		return nil
	}

	return &CloudFolder{
		BaseModel:   c.BaseModel,
		Name:        c.Name,
		ParentID:    c.ParentID,
		Path:        c.Path,
		OwnerID:     c.OwnerID,
		IsPublic:    c.IsPublic,
		ShareCode:   c.ShareCode,
		ShareExpire: c.ShareExpire,
		Description: c.Description,
		FileCount:   c.FileCount,
		TotalSize:   c.TotalSize,
	}
}

// FromFile 从文件创建
func CloudItemFromFile(file *CloudFile) *CloudItem {
	item := &CloudItem{
		BaseModel:     file.BaseModel,
		ItemType:      "file",
		Name:          file.FileName,
		ParentID:      file.FolderID,
		Path:          file.Path,
		OwnerID:       file.OwnerID,
		IsPublic:      file.IsPublic,
		ShareCode:     file.ShareCode,
		ShareExpire:   file.ShareExpire,
		DownloadCount: file.DownloadCount,
		Description:   file.Description,
	}

	item.StorageType = &file.StorageType
	item.StoragePath = &file.StoragePath
	item.FileSize = &file.FileSize
	item.FileType = &file.FileType
	item.FileExt = &file.FileExt
	item.MD5 = &file.MD5
	item.AccessURL = &file.AccessURL
	item.Thumbnail = &file.Thumbnail
	item.Tags = &file.Tags

	return item
}

// FromFolder 从文件夹创建
func CloudItemFromFolder(folder *CloudFolder) *CloudItem {
	return &CloudItem{
		BaseModel:   folder.BaseModel,
		ItemType:    "folder",
		Name:        folder.Name,
		ParentID:    folder.ParentID,
		Path:        folder.Path,
		OwnerID:     folder.OwnerID,
		IsPublic:    folder.IsPublic,
		ShareCode:   folder.ShareCode,
		ShareExpire: folder.ShareExpire,
		Description: folder.Description,
		FileCount:   folder.FileCount,
		TotalSize:   folder.TotalSize,
	}
}
