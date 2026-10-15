package company

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	apperrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name            string `json:"name" binding:"required"`
	Code            string `json:"code"`
	Email           string `json:"email"`
	Phone           string `json:"phone" binding:"required"`
	VerifyCode      string `json:"verifyCode" binding:"required"`
	ContactName     string `json:"contactName"`
	ContactPhone    string `json:"contactPhone"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword"`
	AdminName       string `json:"adminName"`
}

// UpdateCompanyConfigRequest 更新公司配置请求
// 使用指针字段实现部分更新：nil 表示不修改该字段
type UpdateCompanyConfigRequest struct {
	// 通用密钥
	SecretID  *string `json:"secretId"`
	SecretKey *string `json:"secretKey"`
	Region    *string `json:"region"`

	// 存储配置
	StorageType   *string `json:"storageType"`
	LocalBasePath *string `json:"localBasePath"`
	LocalBaseURL  *string `json:"localBaseUrl"`

	// 阿里云 OSS
	AliyunOSSEndpoint        *string `json:"aliyunOssEndpoint"`
	AliyunOSSAccessKeyID     *string `json:"aliyunOssAccessKeyId"`
	AliyunOSSAccessKeySecret *string `json:"aliyunOssAccessKeySecret"`
	AliyunOSSBucketName      *string `json:"aliyunOssBucketName"`
	AliyunOSSCdnDomain       *string `json:"aliyunOssCdnDomain"`

	// 腾讯云 COS
	TencentCOSBucketURL  *string `json:"tencentCosBucketUrl"`
	TencentCOSSecretID   *string `json:"tencentCosSecretId"`
	TencentCOSSecretKey  *string `json:"tencentCosSecretKey"`
	TencentCOSBucketName *string `json:"tencentCosBucketName"`
	TencentCOSRegion     *string `json:"tencentCosRegion"`
	TencentCOSCdnDomain  *string `json:"tencentCosCdnDomain"`

	// 腾讯云直播/云 API
	TencentCloudSecretID    *string `json:"tencentCloudSecretId"`
	TencentCloudSecretKey   *string `json:"tencentCloudSecretKey"`
	TencentCloudRegion      *string `json:"tencentCloudRegion"`
	TencentCloudCallbackKey *string `json:"tencentCloudCallbackKey"`

	// 火山引擎
	VolcengineAccessKeyId     *string `json:"volcengineAccessKeyId"`
	VolcengineAccessKeySecret *string `json:"volcengineAccessKeySecret"`
	VolcengineRegion          *string `json:"volcengineRegion"`
	VolcengineService         *string `json:"volcengineService"`

	// 本地化与限制
	DefaultLanguage *string `json:"defaultLanguage"`
	Timezone        *string `json:"timezone"`
	Currency        *string `json:"currency"`
	MaxStorageSize  *int64  `json:"maxStorageSize"`
	MaxFileSize     *int64  `json:"maxFileSize"`
	MaxFiles        *int    `json:"maxFiles"`
}

// 注册校验错误（handler 据此返回 400 / 409 与提示）
var (
	ErrInvalidPhone  = errors.New("手机号格式不正确")
	ErrPhoneTaken    = errors.New("该手机号已注册")
	ErrUsernameTaken = errors.New("该用户名已被使用")
)

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// ValidPhone 中国大陆手机号：1 开头、第二位 3～9、共 11 位数字
func ValidPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

// RegisterAvailability 注册前的用户名、手机号可用性
type RegisterAvailability struct {
	UsernameTaken bool `json:"usernameTaken"`
	PhoneValid    bool `json:"phoneValid"`
	PhoneTaken    bool `json:"phoneTaken"`
}

// Service 公司服务接口
type Service interface {
	Register(ctx context.Context, req *RegisterRequest) error
	// CheckRegisterAvailability 检查用户名、手机号能否用于注册（空值不检查）
	CheckRegisterAvailability(ctx context.Context, username, phone string) (*RegisterAvailability, error)
	GetCompanyConfig(ctx context.Context, companyID uint) (*entity.SysCompanyConf, error)
	UpdateCompanyConfig(ctx context.Context, companyID uint, req *UpdateCompanyConfigRequest) (*entity.SysCompanyConf, error)
}

type service struct {
	db       *gorm.DB
	confRepo repository.SysCompanyConfRepository
	// onConfigChanged 配置更新成功后的回调（用于刷新厂商客户端缓存，如腾讯云密钥）
	onConfigChanged func(companyID uint)
	// subdomainSuffix 注册时分配的公司子域后缀（如 ".zhibo.skyzhou.cn"）；空则不分配域名
	subdomainSuffix string
}

// NewService 创建公司服务
func NewService(db *gorm.DB, confRepo repository.SysCompanyConfRepository) Service {
	return &service{
		db:       db,
		confRepo: confRepo,
	}
}

// SetConfigChangedCallback 设置配置更新回调（供 main 注入腾讯云/火山等客户端缓存刷新）
func (s *service) SetConfigChangedCallback(fn func(companyID uint)) {
	s.onConfigChanged = fn
}

// ConfigChangedSetter 配置更新回调注入接口（主程序通过类型断言注入刷新回调）
type ConfigChangedSetter interface {
	SetConfigChangedCallback(fn func(companyID uint))
}

// SetSubdomainSuffix 设置注册时分配的公司子域后缀（供 main 按 app.subdomainBase 注入）
func (s *service) SetSubdomainSuffix(suffix string) {
	s.subdomainSuffix = suffix
}

// SubdomainSuffixSetter 公司子域后缀注入接口（主程序通过类型断言注入）
type SubdomainSuffixSetter interface {
	SetSubdomainSuffix(suffix string)
}

// defaultCompanyDomain 新公司的默认子域 {公司ID}<suffix>；suffix 为空（未配置 app.subdomainBase）时返回空串
func defaultCompanyDomain(companyID uint, suffix string) string {
	if suffix == "" || companyID == 0 {
		return ""
	}
	return fmt.Sprintf("%d%s", companyID, suffix)
}

// Register 执行注册流程
func (s *service) Register(ctx context.Context, req *RegisterRequest) error {
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("两次密码不一致")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Phone = strings.TrimSpace(req.Phone)
	if !ValidPhone(req.Phone) {
		return ErrInvalidPhone
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 0. 用户名、手机号全局唯一：业务表的 CREATE_BY 记录的是用户名，登录也按用户名 / 手机号查找，重复会串到其他账号
		if taken, err := userExists(tx, "USERNAME", req.Username); err != nil {
			return err
		} else if taken {
			return ErrUsernameTaken
		}
		if taken, err := userExists(tx, "PHONE", req.Phone); err != nil {
			return err
		} else if taken {
			return ErrPhoneTaken
		}

		// 1. 创建公司
		now := time.Now()
		result := tx.Exec(`
			INSERT INTO sys_company (NAME, CODE, CONTACT_NAME, CONTACT_PHONE, EMAIL, IS_ACTIVE, CREATE_TIME)
			VALUES (?, ?, ?, ?, ?, 'Y', ?)
		`, req.Name, req.Code, req.ContactName, req.ContactPhone, req.Email, now)
		if err := result.Error; err != nil {
			return fmt.Errorf("创建公司失败: %w", err)
		}
		var companyID uint
		tx.Raw("SELECT LAST_INSERT_ID()").Scan(&companyID)

		// 联系电话兜底：未填时用注册手机号
		contactPhone := req.ContactPhone
		if contactPhone == "" {
			contactPhone = req.Phone
		}

		// SYS_COMPANY_ID = ID（自身引用）；配置了子域后缀时分配默认子域（多公司域名登录）。
		// 未配置时不分配：公司有域名后，用户名登录只能从该域名进入，在默认域名下会被拒绝
		if domain := defaultCompanyDomain(companyID, s.subdomainSuffix); domain != "" {
			tx.Exec("UPDATE sys_company SET SYS_COMPANY_ID = ?, CONTACT_PHONE = ?, DOMAIN = ? WHERE ID = ?",
				companyID, contactPhone, domain, companyID)
		} else {
			tx.Exec("UPDATE sys_company SET SYS_COMPANY_ID = ?, CONTACT_PHONE = ? WHERE ID = ?",
				companyID, contactPhone, companyID)
		}

		// 2. 创建管理员用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("加密密码失败: %w", err)
		}

		userResult := tx.Exec(`
			INSERT INTO sys_user (USERNAME, PASSWORD, TRUE_NAME, EMAIL, PHONE, SYS_COMPANY_ID, IS_ADMIN, IS_ACTIVE, CREATE_TIME)
			VALUES (?, ?, ?, ?, ?, ?, 'Y', 'Y', ?)
		`, req.Username, string(hashedPassword), req.AdminName, req.Email, req.Phone, companyID, now)
		if err := userResult.Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}

		return nil
	})
}

// StorageTypeTencentCOS 云盘存储类型：腾讯云 COS（公司配置只支持此类型）
const StorageTypeTencentCOS = "tencentCOS"

// GetCompanyConfig 获取公司配置
func (s *service) GetCompanyConfig(ctx context.Context, companyID uint) (*entity.SysCompanyConf, error) {
	if companyID == 0 {
		return nil, fmt.Errorf("公司ID不能为空")
	}
	return s.confRepo.GetByCompanyID(ctx, companyID)
}

// UpdateCompanyConfig 更新公司配置
func (s *service) UpdateCompanyConfig(ctx context.Context, companyID uint, req *UpdateCompanyConfigRequest) (*entity.SysCompanyConf, error) {
	if companyID == 0 {
		return nil, fmt.Errorf("公司ID不能为空")
	}
	// 云盘存储只支持腾讯云 COS（不再支持服务器本地等）；统一写为 tencentCOS
	if req.StorageType != nil {
		switch strings.ToLower(strings.TrimSpace(*req.StorageType)) {
		case "tencentcos", "cos":
			v := StorageTypeTencentCOS
			req.StorageType = &v
		default:
			return nil, apperrors.New(apperrors.ErrValidation, "云盘存储只支持云存储（对象存储）")
		}
	}

	conf, err := s.confRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("查询公司配置失败: %w", err)
	}

	now := time.Now()
	username := ""
	if u, ok := ctx.Value("username").(string); ok {
		username = u
	}

	if conf == nil {
		// 创建默认配置
		conf = &entity.SysCompanyConf{
			BaseModel: entity.BaseModel{
				CreateBy: username,
				UpdateBy: username,
				IsActive: "Y",
			},
			SysCompanyID:    companyID,
			DefaultLanguage: "zh-CN",
			Timezone:        "Asia/Shanghai",
			Currency:        "CNY",
		}
	}
	conf.UpdateBy = username
	conf.UpdateTime = now

	applyString(req.SecretID, &conf.SecretID)
	applyString(req.SecretKey, &conf.SecretKey)
	applyString(req.Region, &conf.Region)
	applyString(req.StorageType, &conf.StorageType)
	applyString(req.LocalBasePath, &conf.LocalBasePath)
	applyString(req.LocalBaseURL, &conf.LocalBaseURL)
	applyString(req.AliyunOSSEndpoint, &conf.AliyunOSSEndpoint)
	applyString(req.AliyunOSSAccessKeyID, &conf.AliyunOSSAccessKeyID)
	applyString(req.AliyunOSSAccessKeySecret, &conf.AliyunOSSAccessKeySecret)
	applyString(req.AliyunOSSBucketName, &conf.AliyunOSSBucketName)
	applyString(req.AliyunOSSCdnDomain, &conf.AliyunOSSCdnDomain)
	applyString(req.TencentCOSBucketURL, &conf.TencentCOSBucketURL)
	applyString(req.TencentCOSSecretID, &conf.TencentCOSSecretID)
	applyString(req.TencentCOSSecretKey, &conf.TencentCOSSecretKey)
	applyString(req.TencentCOSBucketName, &conf.TencentCOSBucketName)
	applyString(req.TencentCOSRegion, &conf.TencentCOSRegion)
	applyString(req.TencentCOSCdnDomain, &conf.TencentCOSCdnDomain)
	applyString(req.TencentCloudSecretID, &conf.TencentCloudSecretID)
	applyString(req.TencentCloudSecretKey, &conf.TencentCloudSecretKey)
	applyString(req.TencentCloudRegion, &conf.TencentCloudRegion)
	applyString(req.TencentCloudCallbackKey, &conf.TencentCloudCallbackKey)
	applyString(req.VolcengineAccessKeyId, &conf.VolcengineAccessKeyId)
	applyString(req.VolcengineAccessKeySecret, &conf.VolcengineAccessKeySecret)
	applyString(req.VolcengineRegion, &conf.VolcengineRegion)
	applyString(req.VolcengineService, &conf.VolcengineService)
	applyString(req.DefaultLanguage, &conf.DefaultLanguage)
	applyString(req.Timezone, &conf.Timezone)
	applyString(req.Currency, &conf.Currency)
	applyInt64(req.MaxStorageSize, &conf.MaxStorageSize)
	applyInt64(req.MaxFileSize, &conf.MaxFileSize)
	applyInt(req.MaxFiles, &conf.MaxFiles)

	if conf.ID == 0 {
		conf.CreateTime = now
		if err := s.confRepo.Create(ctx, conf); err != nil {
			return nil, fmt.Errorf("创建公司配置失败: %w", err)
		}
	} else {
		if err := s.confRepo.Update(ctx, conf); err != nil {
			return nil, fmt.Errorf("更新公司配置失败: %w", err)
		}
	}

	// 配置已变更，触发回调刷新厂商客户端缓存（腾讯云/火山等密钥）
	if s.onConfigChanged != nil {
		s.onConfigChanged(companyID)
	}

	return conf, nil
}

func applyString(src *string, dst *string) {
	if src != nil {
		*dst = *src
	}
}

func applyInt64(src *int64, dst *int64) {
	if src != nil {
		*dst = *src
	}
}

func applyInt(src *int, dst *int) {
	if src != nil {
		*dst = *src
	}
}

// CheckRegisterAvailability 检查用户名、手机号能否用于注册
func (s *service) CheckRegisterAvailability(ctx context.Context, username, phone string) (*RegisterAvailability, error) {
	db := s.db.WithContext(ctx)
	res := &RegisterAvailability{}
	if username = strings.TrimSpace(username); username != "" {
		taken, err := userExists(db, "USERNAME", username)
		if err != nil {
			return nil, err
		}
		res.UsernameTaken = taken
	}
	if phone = strings.TrimSpace(phone); phone != "" {
		res.PhoneValid = ValidPhone(phone)
		if res.PhoneValid {
			taken, err := userExists(db, "PHONE", phone)
			if err != nil {
				return nil, err
			}
			res.PhoneTaken = taken
		}
	}
	return res, nil
}

// userExists 是否已有有效用户使用该用户名 / 手机号（column 只取 USERNAME 或 PHONE）
func userExists(db *gorm.DB, column, value string) (bool, error) {
	if column != "USERNAME" && column != "PHONE" {
		return false, fmt.Errorf("不支持的字段 %s", column)
	}
	var n int64
	if err := db.Table("sys_user").Where(column+" = ? AND IS_ACTIVE = ?", value, "Y").Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
