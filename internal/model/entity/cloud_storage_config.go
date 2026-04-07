package entity

import "time"

// CloudStorageConfig 云盘存储配置
type CloudStorageConfig struct {
	ID           uint   `gorm:"column:ID;primaryKey" json:"id"`
	SysCompanyID uint   `gorm:"column:SYS_COMPANY_ID;not null;index" json:"sysCompanyId"`
	StorageType  string `gorm:"column:STORAGE_TYPE;size:20;default:local" json:"storageType"`

	// 本地存储配置
	LocalBasePath string `gorm:"column:LOCAL_BASE_PATH;size:500" json:"localBasePath"`
	LocalBaseURL  string `gorm:"column:LOCAL_BASE_URL;size:500" json:"localBaseURL"`

	// 阿里云OSS配置
	AliyunOSSEndpoint        string `gorm:"column:ALIYUN_OSS_ENDPOINT;size:255" json:"aliyunOssEndpoint"`
	AliyunOSSAccessKeyID     string `gorm:"column:ALIYUN_OSS_ACCESS_KEY_ID;size:255" json:"aliyunOssAccessKeyId"`
	AliyunOSSAccessKeySecret string `gorm:"column:ALIYUN_OSS_ACCESS_KEY_SECRET;size:255" json:"aliyunOssAccessKeySecret"`
	AliyunOSSBucketName      string `gorm:"column:ALIYUN_OSS_BUCKET_NAME;size:255" json:"aliyunOssBucketName"`
	AliyunOSSCDNDomain       string `gorm:"column:ALIYUN_OSS_CDN_DOMAIN;size:500" json:"aliyunOssCdnDomain"`

	// 腾讯云COS配置
	TencentCOSBucketURL  string `gorm:"column:TENCENT_COS_BUCKET_URL;size:500" json:"tencentCosBucketUrl"`
	TencentCOSSecretID   string `gorm:"column:TENCENT_COS_SECRET_ID;size:255" json:"tencentCosSecretId"`
	TencentCOSSecretKey  string `gorm:"column:TENCENT_COS_SECRET_KEY;size:255" json:"tencentCosSecretKey"`
	TencentCOSBucketName string `gorm:"column:TENCENT_COS_BUCKET_NAME;size:255" json:"tencentCosBucketName"`
	TencentCOSRegion     string `gorm:"column:TENCENT_COS_REGION;size:50" json:"tencentCosRegion"`
	TencentCOSCDNDomain  string `gorm:"column:TENCENT_COS_CDN_DOMAIN;size:500" json:"tencentCosCdnDomain"`

	IsActive   string    `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"`
	CreateBy   string    `gorm:"column:CREATE_BY;size:80" json:"createBy"`
	CreateTime time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateBy   string    `gorm:"column:UPDATE_BY;size:80" json:"updateBy"`
	UpdateTime time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

// TableName 指定表名
func (CloudStorageConfig) TableName() string {
	return "cloud_storage_config"
}
