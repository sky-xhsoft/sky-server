package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置结构
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
	Swagger         SwaggerConfig         `mapstructure:"swagger"`
	Security        SecurityConfig        `mapstructure:"security"`
	Monitoring      MonitoringConfig      `mapstructure:"monitoring"`
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
