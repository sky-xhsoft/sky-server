package entity

import "time"

// CloudFolder 云盘文件夹
type CloudFolder struct {
	BaseModel
	Name        string     `gorm:"column:NAME;size:255;not null" json:"name"`              // 文件夹名称
	ParentID    *uint      `gorm:"column:PARENT_ID;index" json:"parentId"`                 // 父文件夹ID
	Path        string     `gorm:"column:PATH;size:1000;not null;index" json:"path"`       // 完整路径
	OwnerID     uint       `gorm:"column:OWNER_ID;index;not null" json:"ownerId"`          // 所有者ID
	IsPublic    string     `gorm:"column:IS_PUBLIC;size:1;default:N" json:"isPublic"`      // 是否公开 Y/N
	ShareCode   *string    `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"` // 分享码
	ShareExpire *time.Time `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`   // 分享过期时间
	Description string     `gorm:"column:DESCRIPTION;size:500" json:"description"`         // 描述
	FileCount   int        `gorm:"column:FILE_COUNT;default:0" json:"fileCount"`           // 文件数量
	TotalSize   int64      `gorm:"column:TOTAL_SIZE;default:0" json:"totalSize"`           // 总大小（字节）
}

// TableName 指定表名
func (CloudFolder) TableName() string {
	return "cloud_folder"
}

// CloudFile 云盘文件
type CloudFile struct {
	BaseModel
	FileName      string     `gorm:"column:FILE_NAME;size:255;not null" json:"fileName"`                    // 文件名
	FolderID      *uint      `gorm:"column:FOLDER_ID;index" json:"folderId"`                                // 文件夹ID
	Path          string     `gorm:"column:PATH;size:1000;not null" json:"path"`                            // 完整路径
	StorageType   string     `gorm:"column:STORAGE_TYPE;size:20;not null;default:local" json:"storageType"` // 存储类型: local, oss
	StoragePath   string     `gorm:"column:STORAGE_PATH;size:500;not null" json:"storagePath"`              // 存储路径
	FileSize      int64      `gorm:"column:FILE_SIZE;not null" json:"fileSize"`                             // 文件大小（字节）
	FileType      string     `gorm:"column:FILE_TYPE;size:100" json:"fileType"`                             // 文件MIME类型
	FileExt       string     `gorm:"column:FILE_EXT;size:20" json:"fileExt"`                                // 文件扩展名
	MD5           string     `gorm:"column:MD5;size:32;index" json:"md5"`                                   // MD5值
	OwnerID       uint       `gorm:"column:OWNER_ID;index;not null" json:"ownerId"`                         // 所有者ID
	IsPublic      string     `gorm:"column:IS_PUBLIC;size:1;default:N" json:"isPublic"`                     // 是否公开 Y/N
	ShareCode     *string    `gorm:"column:SHARE_CODE;size:50;uniqueIndex" json:"shareCode"`                // 分享码
	ShareExpire   *time.Time `gorm:"column:SHARE_EXPIRE;type:datetime" json:"shareExpire"`                  // 分享过期时间
	AccessURL     string     `gorm:"column:ACCESS_URL;size:500" json:"accessUrl"`                           // 访问URL
	Thumbnail     string     `gorm:"column:THUMBNAIL;size:500" json:"thumbnail"`                            // 缩略图URL
	DownloadCount int        `gorm:"column:DOWNLOAD_COUNT;default:0" json:"downloadCount"`                  // 下载次数
	Tags          string     `gorm:"column:TAGS;size:500" json:"tags"`                                      // 标签（逗号分隔）
	Description   string     `gorm:"column:DESCRIPTION;size:500" json:"description"`                        // 描述
}

// TableName 指定表名
func (CloudFile) TableName() string {
	return "cloud_file"
}

// CloudShare 云盘分享记录
type CloudShare struct {
	BaseModel
	ShareCode     string     `gorm:"column:SHARE_CODE;size:50;uniqueIndex;not null" json:"shareCode"` // 分享码
	ResourceType  string     `gorm:"column:RESOURCE_TYPE;size:20;not null" json:"resourceType"`       // 资源类型: file, folder
	ResourceID    uint       `gorm:"column:RESOURCE_ID;index;not null" json:"resourceId"`             // 资源ID
	SharerID      uint       `gorm:"column:SHARER_ID;index;not null" json:"sharerId"`                 // 分享者ID
	ShareType     string     `gorm:"column:SHARE_TYPE;size:20;not null" json:"shareType"`             // 分享类型: public, password, private
	Password      string     `gorm:"column:PASSWORD;size:50" json:"password"`                         // 访问密码
	ExpireTime    *time.Time `gorm:"column:EXPIRE_TIME;type:datetime" json:"expireTime"`              // 过期时间
	MaxDownloads  int        `gorm:"column:MAX_DOWNLOADS;default:0" json:"maxDownloads"`              // 最大下载次数（0=无限制）
	DownloadCount int        `gorm:"column:DOWNLOAD_COUNT;default:0" json:"downloadCount"`            // 已下载次数
	ViewCount     int        `gorm:"column:VIEW_COUNT;default:0" json:"viewCount"`                    // 查看次数
	Status        string     `gorm:"column:STATUS;size:20;default:active" json:"status"`              // 状态: active, expired, disabled
}

// TableName 指定表名
func (CloudShare) TableName() string {
	return "cloud_share"
}

// CloudQuota 云盘配额
type CloudQuota struct {
	BaseModel
	UserID      uint   `gorm:"column:USER_ID;uniqueIndex;not null" json:"userId"`           // 用户ID
	TotalQuota  int64  `gorm:"column:TOTAL_QUOTA;not null" json:"totalQuota"`               // 总配额（字节）
	UsedSpace   int64  `gorm:"column:USED_SPACE;default:0" json:"usedSpace"`                // 已用空间（字节）
	FileCount   int    `gorm:"column:FILE_COUNT;default:0" json:"fileCount"`                // 文件数量
	FolderCount int    `gorm:"column:FOLDER_COUNT;default:0" json:"folderCount"`            // 文件夹数量
	MaxFileSize int64  `gorm:"column:MAX_FILE_SIZE;default:0" json:"maxFileSize"`           // 单文件最大大小（字节）
	QuotaType   string `gorm:"column:QUOTA_TYPE;size:20;default:standard" json:"quotaType"` // 配额类型: standard, premium
}

// TableName 指定表名
func (CloudQuota) TableName() string {
	return "cloud_quota"
}

// CloudUploadSession 云盘上传会话（用于分片上传和断点续传）
type CloudUploadSession struct {
	BaseModel
	FileID         string     `gorm:"column:FILE_ID;size:64;not null;index" json:"fileId"`                   // 文件唯一标识（MD5）
	UserID         uint       `gorm:"column:USER_ID;index;not null" json:"userId"`                           // 用户ID
	FileName       string     `gorm:"column:FILE_NAME;size:255;not null" json:"fileName"`                    // 文件名
	FileSize       int64      `gorm:"column:FILE_SIZE;not null" json:"fileSize"`                             // 文件总大小（字节）
	FileType       string     `gorm:"column:FILE_TYPE;size:100" json:"fileType"`                             // 文件MIME类型
	FolderID       *uint      `gorm:"column:FOLDER_ID;index" json:"folderId"`                                // 目标文件夹ID
	ChunkSize      int        `gorm:"column:CHUNK_SIZE;not null;default:5242880" json:"chunkSize"`           // 分片大小（默认5MB）
	TotalChunks    int        `gorm:"column:TOTAL_CHUNKS;not null" json:"totalChunks"`                       // 总分片数
	UploadedChunks string     `gorm:"column:UPLOADED_CHUNKS;type:text" json:"uploadedChunks"`                // 已上传的分片索引（JSON数组）
	Status         string     `gorm:"column:STATUS;size:20;not null;default:uploading;index" json:"status"`  // 状态：uploading,paused,completed,failed
	StorageType    string     `gorm:"column:STORAGE_TYPE;size:20;not null;default:local" json:"storageType"` // 存储类型：local,oss
	StoragePath    string     `gorm:"column:STORAGE_PATH;size:500" json:"storagePath"`                       // 临时存储路径
	ExpireTime     *time.Time `gorm:"column:EXPIRE_TIME;not null;index" json:"expireTime"`                   // 过期时间
	ErrorMessage   string     `gorm:"column:ERROR_MESSAGE;type:text" json:"errorMessage"`                    // 错误信息
}

// TableName 指定表名
func (CloudUploadSession) TableName() string {
	return "cloud_upload_session"
}

// CloudChunkRecord 云盘分片记录
type CloudChunkRecord struct {
	ID         uint       `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`                            // 主键ID
	SessionID  uint       `gorm:"column:SESSION_ID;index;not null" json:"sessionId"`                       // 会话ID
	ChunkIndex int        `gorm:"column:CHUNK_INDEX;not null;index" json:"chunkIndex"`                     // 分片索引（从0开始）
	ChunkSize  int        `gorm:"column:CHUNK_SIZE;not null" json:"chunkSize"`                             // 分片大小（字节）
	ChunkMD5   string     `gorm:"column:CHUNK_MD5;size:32;not null" json:"chunkMd5"`                       // 分片MD5
	ChunkPath  string     `gorm:"column:CHUNK_PATH;size:500" json:"chunkPath"`                             // 分片存储路径
	Uploaded   bool       `gorm:"column:UPLOADED;not null;default:false;index" json:"uploaded"`            // 是否已上传
	UploadTime *time.Time `gorm:"column:UPLOAD_TIME;type:datetime" json:"uploadTime"`                      // 上传时间
	RetryCount int        `gorm:"column:RETRY_COUNT;not null;default:0" json:"retryCount"`                 // 重试次数
	CreateTime time.Time  `gorm:"column:CREATE_TIME;not null;default:CURRENT_TIMESTAMP" json:"createTime"` // 创建时间
	UpdateTime time.Time  `gorm:"column:UPDATE_TIME;not null;default:CURRENT_TIMESTAMP" json:"updateTime"` // 修改时间
}

// TableName 指定表名
func (CloudChunkRecord) TableName() string {
	return "cloud_chunk_record"
}
