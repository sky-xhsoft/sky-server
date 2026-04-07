package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置结构
// CloudConfig 云盘配置
type CloudConfig struct {
	Enabled        bool          `mapstructure:"enabled"`        // 是否启用云盘功能
	AuthStrategy   string        `mapstructure:"authStrategy"`   // 用户认证策略: sso, standalone, both
	DefaultStorage string        `mapstructure:"defaultStorage"` // 默认存储类型
	Storage        StorageConfig `mapstructure:"storage"`        // 存储配置（作为全局默认配置）
}

type Config struct {
	App             AppConfig             `mapstructure:"app"`
	Database        DatabaseConfig        `mapstructure:"database"`
	Redis           RedisConfig           `mapstructure:"redis"`
	JWT             JWTConfig             `mapstructure:"jwt"`
	Log             LogConfig             `mapstructure:"log"`
	CORS            CORSConfig            `mapstructure:"cors"`
	Cache           CacheConfig           `mapstructure:"cache"`
	Action          ActionConfig          `mapstructure:"action"`
	RateLimit       RateLimitConfig       `mapstructure:"rateLimit"`
	Upload          UploadConfig          `mapstructure:"upload"`
	File            FileConfig            `mapstructure:"file"`
	MultipartUpload MultipartUploadConfig `mapstructure:"multipartUpload"`
	Cloud           CloudConfig           `mapstructure:"cloud"`   // 云盘配置
	Storage         StorageConfig         `mapstructure:"storage"` // 存储配置（兼容旧版本）
	Swagger         SwaggerConfig         `mapstructure:"swagger"`
	Security        SecurityConfig        `mapstructure:"security"`
	Monitoring      MonitoringConfig      `mapstructure:"monitoring"`
	TencentCloud    TencentCloudConfig    `mapstructure:"tencentCloud"`
	Volcengine      VolcengineConfig      `mapstructure:"volcengine"` // 火山引擎配置
}

// StorageConfig 存储配置
type StorageConfig struct {
	Default    string             `mapstructure:"default"`    // 默认存储类型
	Local      LocalStorageConfig `mapstructure:"local"`      // 本地存储配置
	AliyunOSS  AliyunOSSConfig    `mapstructure:"aliyunOSS"`  // 阿里云OSS配置
	TencentCOS TencentCOSConfig   `mapstructure:"tencentCOS"` // 腾讯云COS配置
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	BasePath string `mapstructure:"basePath"` // 基础路径
	BaseURL  string `mapstructure:"baseURL"`  // 基础URL
}

// AliyunOSSConfig 阿里云OSS配置
type AliyunOSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`        // OSS endpoint
	AccessKeyID     string `mapstructure:"accessKeyID"`     // AccessKey ID
	AccessKeySecret string `mapstructure:"accessKeySecret"` // AccessKey Secret
	BucketName      string `mapstructure:"bucketName"`      // Bucket名称
	CDNDomain       string `mapstructure:"cdnDomain"`       // CDN加速域名
}

// TencentCOSConfig 腾讯云COS配置
type TencentCOSConfig struct {
	BucketURL  string `mapstructure:"bucketURL"`  // Bucket URL
	SecretID   string `mapstructure:"secretID"`   // Secret ID
	SecretKey  string `mapstructure:"secretKey"`  // Secret Key
	BucketName string `mapstructure:"bucketName"` // Bucket名称
	Region     string `mapstructure:"region"`     // 区域
	CDNDomain  string `mapstructure:"cdnDomain"`  // CDN加速域名
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parseTime"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime int    `mapstructure:"connMaxLifetime"`
	LogLevel        string `mapstructure:"logLevel"` // SQL日志级别: silent, error, warn, info
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"poolSize"`
	MinIdleConns int    `mapstructure:"minIdleConns"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpire  int    `mapstructure:"accessTokenExpire"`
	RefreshTokenExpire int    `mapstructure:"refreshTokenExpire"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"filePath"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Compress   bool   `mapstructure:"compress"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allowOrigins"`
	AllowMethods     []string `mapstructure:"allowMethods"`
	AllowHeaders     []string `mapstructure:"allowHeaders"`
	ExposeHeaders    []string `mapstructure:"exposeHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MetadataTTL   int `mapstructure:"metadataTTL"`
	DictTTL       int `mapstructure:"dictTTL"`
	PermissionTTL int `mapstructure:"permissionTTL"`
}

// ActionConfig 动作配置
type ActionConfig struct {
	ScriptTimeout int `mapstructure:"scriptTimeout"` // 脚本执行超时时间（秒）
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
	BurstSize         int  `mapstructure:"burstSize"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxSize      int64    `mapstructure:"maxSize"`
	AllowedTypes []string `mapstructure:"allowedTypes"`
	UploadPath   string   `mapstructure:"uploadPath"`
}

// FileConfig 文件管理配置
type FileConfig struct {
	UploadDir   string   `mapstructure:"uploadDir"`   // 上传目录
	MaxFileSize int64    `mapstructure:"maxFileSize"` // 最大文件大小（字节）
	AllowedExts []string `mapstructure:"allowedExts"` // 允许的文件扩展名
}

// MultipartUploadConfig 分片上传配置
type MultipartUploadConfig struct {
	ChunkSize          int `mapstructure:"chunkSize"`          // 默认分片大小（字节）
	SessionExpireHours int `mapstructure:"sessionExpireHours"` // 会话过期时间（小时）
	CleanupInterval    int `mapstructure:"cleanupInterval"`    // 清理任务执行间隔（秒）
}

// SwaggerConfig Swagger配置
type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
	Host        string `mapstructure:"host"`
	BasePath    string `mapstructure:"basePath"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	PasswordCost        int      `mapstructure:"passwordCost"`
	AllowedBashCommands []string `mapstructure:"allowedBashCommands"`
	BashTimeout         int      `mapstructure:"bashTimeout"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	SlowQueryEnabled   bool `mapstructure:"slowQueryEnabled"`
	SlowQueryThreshold int  `mapstructure:"slowQueryThreshold"`
}

// TencentCloudConfig 腾讯云配置
type TencentCloudConfig struct {
	SecretID    string     `mapstructure:"secretID"`    // 腾讯云 SecretID
	SecretKey   string     `mapstructure:"secretKey"`   // 腾讯云 SecretKey
	Region      string     `mapstructure:"region"`      // 区域，如 ap-guangzhou
	CallbackKey string     `mapstructure:"callbackKey"` // 回调密钥，用于验证回调签名
	Live        LiveConfig `mapstructure:"live"`        // 直播配置
}

// LiveConfig 直播配置
type LiveConfig struct {
	Enabled      bool   `mapstructure:"enabled"`      // 是否启用直播功能
	PushDomain   string `mapstructure:"pushDomain"`   // 推流域名
	PlayDomain   string `mapstructure:"playDomain"`   // 播放域名
	AppName      string `mapstructure:"appName"`      // 应用名称
	StreamKey    string `mapstructure:"streamKey"`    // 推流密钥
	RecordBucket string `mapstructure:"recordBucket"` // 录制文件存储桶
	RecordRegion string `mapstructure:"recordRegion"` // 录制文件存储区域
	CallbackURL  string `mapstructure:"callbackURL"`  // 回调URL
}

// VolcengineConfig 火山引擎配置
type VolcengineConfig struct {
	AccessKeyId     string `mapstructure:"accessKeyId"`     // 火山引擎访问密钥ID
	AccessKeySecret string `mapstructure:"accessKeySecret"` // 火山引擎访问密钥Secret
	Region          string `mapstructure:"region"`          // 区域，如 cn-north-1
	Service         string `mapstructure:"service"`         // 服务名称，如 live
}

// Load 加载配置文件
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GetDSN 获取MySQL DSN连接字符串
func (c *MySQLConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
		c.ParseTime,
		c.Loc,
	)
}

// GetAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
