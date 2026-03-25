package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// userRepository 用户仓储MySQL实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByUsername(username string) (*entity.SysUser, error) {
	var user entity.SysUser
	err := r.db.Where("USERNAME = ? AND IS_ACTIVE = ?", username, "Y").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(id uint) (*entity.SysUser, error) {
	var user entity.SysUser
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateSession(session *entity.SysUserSession) error {
	return r.db.Create(session).Error
}

func (r *userRepository) GetSessionByDeviceID(userID uint, deviceID string) (*entity.SysUserSession, error) {
	var session entity.SysUserSession
	err := r.db.Where("USER_ID = ? AND DEVICE_ID = ? AND IS_ACTIVE = ?", userID, deviceID, "Y").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userRepository) UpdateSession(session *entity.SysUserSession) error {
	return r.db.Save(session).Error
}

func (r *userRepository) GetActiveSessions(userID uint) ([]*entity.SysUserSession, error) {
	var sessions []*entity.SysUserSession
	err := r.db.Where("USER_ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Order("LOGIN_TIME DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *userRepository) GetSessionByToken(token string) (*entity.SysUserSession, error) {
	var session entity.SysUserSession
	err := r.db.Where("TOKEN = ? AND IS_ACTIVE = ?", token, "Y").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userRepository) DeleteSession(id uint) error {
	return r.db.Model(&entity.SysUserSession{}).Where("ID = ?", id).Update("IS_ACTIVE", "N").Error
}

func (r *userRepository) DeleteSessionByDeviceID(userID uint, deviceID string) error {
	// 物理删除该设备的所有会话记录，避免唯一索引冲突
	return r.db.Where("USER_ID = ? AND DEVICE_ID = ?", userID, deviceID).Delete(&entity.SysUserSession{}).Error
}

func (r *userRepository) DeleteAllSessions(userID uint) error {
	return r.db.Model(&entity.SysUserSession{}).Where("USER_ID = ?", userID).Update("IS_ACTIVE", "N").Error
}

func (r *userRepository) GetCompanyByID(companyID uint) (*entity.SysCompany, error) {
	var company entity.SysCompany
	err := r.db.Where("ID = ? AND IS_ACTIVE = ?", companyID, "Y").First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *userRepository) GetCompanyConfByCompanyID(companyID uint) (*entity.SysCompanyConf, error) {
	var conf entity.SysCompanyConf
	err := r.db.Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", companyID, "Y").First(&conf).Error
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
