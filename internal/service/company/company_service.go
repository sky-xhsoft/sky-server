package company

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name            string `json:"name" binding:"required"`
	Code            string `json:"code"`
	Email           string `json:"email"`
	VerifyCode      string `json:"verifyCode"`
	ContactName     string `json:"contactName"`
	ContactPhone    string `json:"contactPhone"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword"`
	AdminName       string `json:"adminName"`
}
type Service interface {
	Register(ctx context.Context, req *RegisterRequest) error
}

type service struct {
	db *gorm.DB
}

// NewService 创建公司注册服务
func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

// Register 执行注册流程
func (s *service) Register(ctx context.Context, req *RegisterRequest) error {
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("两次密码不一致")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		// SYS_COMPANY_ID = ID（自身引用）
		tx.Exec("UPDATE sys_company SET SYS_COMPANY_ID = ? WHERE ID = ?", companyID, companyID)

		// 2. 创建管理员用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("加密密码失败: %w", err)
		}

		userResult := tx.Exec(`
			INSERT INTO sys_user (USERNAME, PASSWORD, TRUE_NAME, EMAIL, PHONE, SYS_COMPANY_ID, IS_ADMIN, IS_ACTIVE, CREATE_TIME)
			VALUES (?, ?, ?, ?, ?, ?, 'Y', 'Y', ?)
		`, req.Username, string(hashedPassword), req.AdminName, req.Email, req.ContactPhone, companyID, now)
		if err := userResult.Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}

		return nil
	})
}
