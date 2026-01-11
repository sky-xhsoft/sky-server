package entity

import "time"

// SysFile 系统文件
type SysFile struct {
	BaseModel
	FileName     string    `gorm:"column:FILE_NAME;size:255;not null" json:"fileName"`                 // 原始文件名
	StorageName  string    `gorm:"column:STORAGE_NAME;size:255;not null;index" json:"storageName"`    // 存储文件名（唯一）
	FilePath     string    `gorm:"column:FILE_PATH;size:500;not null" json:"filePath"`                // 文件路径
	FileSize     int64     `gorm:"column:FILE_SIZE;not null" json:"fileSize"`                         // 文件大小（字节）
	FileType     string    `gorm:"column:FILE_TYPE;size:100" json:"fileType"`                         // 文件类型/MIME类型
	FileExt      string    `gorm:"column:FILE_EXT;size:20" json:"fileExt"`                            // 文件扩展名
	StorageType  string    `gorm:"column:STORAGE_TYPE;size:20;not null;default:local" json:"storageType"` // 存储类型：local, oss, s3
	BucketName   string    `gorm:"column:BUCKET_NAME;size:100" json:"bucketName"`                     // 存储桶名称（云存储）
	AccessURL    string    `gorm:"column:ACCESS_URL;size:500" json:"accessUrl"`                       // 访问URL
	ThumbnailURL string    `gorm:"column:THUMBNAIL_URL;size:500" json:"thumbnailUrl"`                 // 缩略图URL
	MD5          string    `gorm:"column:MD5;size:32;index" json:"md5"`                               // 文件MD5值
	UploadIP     string    `gorm:"column:UPLOAD_IP;size:50" json:"uploadIp"`                          // 上传IP
	DownloadCount int      `gorm:"column:DOWNLOAD_COUNT;default:0" json:"downloadCount"`              // 下载次数
	Category     string    `gorm:"column:CATEGORY;size:50" json:"category"`                           // 文件分类
	Description  string    `gorm:"column:DESCRIPTION;size:500" json:"description"`                    // 文件描述
	ExpireTime   *time.Time `gorm:"column:EXPIRE_TIME" json:"expireTime"`                             // 过期时间
}

// TableName 指定表名
func (SysFile) TableName() string {
	return "sys_file"
}
