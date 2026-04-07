package cloud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// multipartUploadService 分片上传服务实现
type multipartUploadService struct {
	db                    *gorm.DB
	companyStorageManager *storage.CompanyStorageManager
	cloudService          Service
	userRepo              repository.UserRepository
	defaultChunkSize      int // 默认分片大小（字节）
	sessionExpireHours    int // 会话过期时间（小时）
}

// NewMultipartUploadService 创建分片上传服务
func NewMultipartUploadService(db *gorm.DB, companyStorageManager *storage.CompanyStorageManager, userRepo repository.UserRepository, cloudService Service, defaultChunkSize int, sessionExpireHours int) MultipartUploadService {
	// 设置默认值
	if defaultChunkSize <= 0 {
		defaultChunkSize = 5 * 1024 * 1024 // 默认 5MB
	}
	if sessionExpireHours <= 0 {
		sessionExpireHours = 24 // 默认 24 小时
	}

	return &multipartUploadService{
		db:                    db,
		companyStorageManager: companyStorageManager,
		userRepo:              userRepo,
		cloudService:          cloudService,
		defaultChunkSize:      defaultChunkSize,
		sessionExpireHours:    sessionExpireHours,
	}
}

// getCompanyStorage 根据用户ID获取公司存储实例
func (s *multipartUploadService) getCompanyStorage(ctx context.Context, userID uint) (storage.Storage, error) {
	// 从用户ID获取用户信息
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "获取用户信息失败", err)
	}

	// 获取公司存储实例
	companyStorage, err := s.companyStorageManager.GetStorage(ctx, user.SysCompanyID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrStorage, "获取公司存储实例失败", err)
	}

	return companyStorage, nil
}

// getUsernameByID 根据用户ID获取用户名
func (s *multipartUploadService) getUsernameByID(ctx context.Context, userID uint) string {
	var user entity.SysUser
	if err := s.db.WithContext(ctx).Select("USERNAME").Where("ID = ?", userID).First(&user).Error; err != nil {
		// 如果查询失败，返回默认值
		return fmt.Sprintf("user_%d", userID)
	}
	return user.Username
}
