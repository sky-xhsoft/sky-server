package entity

// SysGroups 权限组
type SysGroups struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255;not null" json:"name"`
	Description string `gorm:"column:DESCRIPTION;size:255" json:"description"`
	Sgrade      int    `gorm:"column:SGRADE" json:"sgrade"`
}

// TableName 指定表名
func (SysGroups) TableName() string {
	return "sys_groups"
}

// SysUserGroups 用户权限组关联
type SysUserGroups struct {
	BaseModel
	SysUserID      uint `gorm:"column:SYS_USER_ID;index:idx_user_groups;not null" json:"sysUserId"`
	SysDirectoryID uint `gorm:"column:SYS_DIRECTORY_ID;index:idx_user_groups;not null" json:"sysDirectoryId"`
}

// TableName 指定表名
func (SysUserGroups) TableName() string {
	return "sys_user_groups"
}

// SysDirectory 安全目录
type SysDirectory struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255;not null" json:"name"`
	SysTableID  *uint  `gorm:"column:SYS_TABLE_ID;index" json:"sysTableId"`
	ParentID    *uint  `gorm:"column:PARENT_ID;index" json:"parentId"`
	Orderno     int    `gorm:"column:ORDERNO" json:"orderno"`
	Description string `gorm:"column:DESCRIPTION;size:255" json:"description"`
}

// TableName 指定表名
func (SysDirectory) TableName() string {
	return "sys_directory"
}

// SysGroupPrem 权限组明细
type SysGroupPrem struct {
	BaseModel
	SysGroupsID    uint   `gorm:"column:SYS_GROUPS_ID;index;not null" json:"sysGroupsId"`
	SysDirectoryID uint   `gorm:"column:SYS_DIRECTORY_ID;index;not null" json:"sysDirectoryId"`
	Permission     int    `gorm:"column:PERMISSION;not null" json:"permission"` // 权限值（位运算）
	FilterObj      string `gorm:"column:FILTER_OBJ;size:255" json:"filterObj"`  // 数据过滤条件（JSON）
}

// TableName 指定表名
func (SysGroupPrem) TableName() string {
	return "sys_group_prem"
}

// SysCompany 公司（多租户）
type SysCompany struct {
	BaseModel
	Name        string  `gorm:"column:NAME;size:255;not null" json:"name"`
	Code        string  `gorm:"column:CODE;size:50;uniqueIndex" json:"code"`
	Domain      *string `gorm:"column:DOMAIN;size:255;uniqueIndex" json:"domain"` // 公司域名（用于多租户识别）
	Description string  `gorm:"column:DESCRIPTION;size:500" json:"description"`
	Status      string  `gorm:"column:STATUS;size:1;default:Y" json:"status"` // Y:启用, N:禁用
}

// TableName 指定表名
func (SysCompany) TableName() string {
	return "sys_company"
}

// SysCompanyConf 公司配置
// 与 sys_company 是 1:1 关系，保存公司的各项配置
type SysCompanyConf struct {
	BaseModel
	SysCompanyID  uint   `gorm:"column:SYS_COMPANY_ID;uniqueIndex" json:"sysCompanyId"` // 关联公司ID
	SecretID      string `gorm:"column:SECRET_ID;size:255" json:"secretId"`             // secretID
	SecretKey     string `gorm:"column:SECRET_KEY;size:255" json:"secretKey"`           // secretKey
	Region        string `gorm:"column:REGION;size:255" json:"region"`                  // region
	StorageType   string `gorm:"column:STORAGE_TYPE;size:20" json:"storageType"`        // 存储类型: local, aliyunOSS, tencentCOS
	LocalBasePath string `gorm:"column:LOCAL_BASE_PATH;size:500" json:"localBasePath"`  // 本地存储基础路径
	LocalBaseURL  string `gorm:"column:LOCAL_BASE_URL;size:500" json:"localBaseUrl"`    // 本地存储基础URL
	// 阿里云OSS配置
	AliyunOSSEndpoint        string `gorm:"column:ALIYUN_OSS_ENDPOINT;size:255" json:"aliyunOssEndpoint"`                 // 阿里云OSS Endpoint
	AliyunOSSAccessKeyID     string `gorm:"column:ALIYUN_OSS_ACCESS_KEY_ID;size:255" json:"aliyunOssAccessKeyId"`         // 阿里云OSS AccessKeyID
	AliyunOSSAccessKeySecret string `gorm:"column:ALIYUN_OSS_ACCESS_KEY_SECRET;size:255" json:"aliyunOssAccessKeySecret"` // 阿里云OSS AccessKeySecret
	AliyunOSSBucketName      string `gorm:"column:ALIYUN_OSS_BUCKET_NAME;size:255" json:"aliyunOssBucketName"`            // 阿里云OSS Bucket名称
	AliyunOSSCdnDomain       string `gorm:"column:ALIYUN_OSS_CDN_DOMAIN;size:500" json:"aliyunOssCdnDomain"`              // 阿里云OSS CDN域名
	// 腾讯云COS配置
	TencentCOSBucketURL  string `gorm:"column:TENCENT_COS_BUCKET_URL;size:500" json:"tencentCosBucketUrl"`   // 腾讯云COS Bucket URL
	TencentCOSSecretID   string `gorm:"column:TENCENT_COS_SECRET_ID;size:255" json:"tencentCosSecretId"`     // 腾讯云COS SecretID
	TencentCOSSecretKey  string `gorm:"column:TENCENT_COS_SECRET_KEY;size:255" json:"tencentCosSecretKey"`   // 腾讯云COS SecretKey
	TencentCOSBucketName string `gorm:"column:TENCENT_COS_BUCKET_NAME;size:255" json:"tencentCosBucketName"` // 腾讯云COS Bucket名称
	TencentCOSRegion     string `gorm:"column:TENCENT_COS_REGION;size:50" json:"tencentCosRegion"`           // 腾讯云COS 区域
	TencentCOSCdnDomain  string `gorm:"column:TENCENT_COS_CDN_DOMAIN;size:500" json:"tencentCosCdnDomain"`   // 腾讯云COS CDN域名
	// 腾讯云直播配置
	TencentCloudSecretID    string `gorm:"column:TENCENT_CLOUD_SECRET_ID;size:255" json:"tencentCloudSecretId"`       // 腾讯云SecretID
	TencentCloudSecretKey   string `gorm:"column:TENCENT_CLOUD_SECRET_KEY;size:255" json:"tencentCloudSecretKey"`     // 腾讯云SecretKey
	TencentCloudRegion      string `gorm:"column:TENCENT_CLOUD_REGION;size:255" json:"tencentCloudRegion"`            // 腾讯云区域
	TencentCloudCallbackKey string `gorm:"column:TENCENT_CLOUD_CALLBACK_KEY;size:255" json:"tencentCloudCallbackKey"` // 腾讯云回调密钥
	// 火山引擎配置
	VolcengineAccessKeyId     string `gorm:"column:VOLCENGINE_ACCESS_KEY_ID;size:255" json:"volcengineAccessKeyId"`         // 火山引擎访问密钥ID
	VolcengineAccessKeySecret string `gorm:"column:VOLCENGINE_ACCESS_KEY_SECRET;size:255" json:"volcengineAccessKeySecret"` // 火山引擎访问密钥Secret
	VolcengineRegion          string `gorm:"column:VOLCENGINE_REGION;size:50" json:"volcengineRegion"`                      // 火山引擎区域
	VolcengineService         string `gorm:"column:VOLCENGINE_SERVICE;size:50" json:"volcengineService"`                    // 火山引擎服务名称
}

// TableName 指定表名
func (SysCompanyConf) TableName() string {
	return "sys_company_conf"
}
