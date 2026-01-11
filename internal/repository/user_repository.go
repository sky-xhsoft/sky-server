package repository

import "github.com/sky-xhsoft/sky-server/internal/model/entity"

// UserRepository 用户仓储接口
type UserRepository interface {
	// 根据用户名获取用户
	GetUserByUsername(username string) (*entity.SysUser, error)

	// 根据ID获取用户
	GetUserByID(id uint) (*entity.SysUser, error)

	// 创建用户会话
	CreateSession(session *entity.SysUserSession) error

	// 根据设备ID获取会话
	GetSessionByDeviceID(userID uint, deviceID string) (*entity.SysUserSession, error)

	// 更新会话
	UpdateSession(session *entity.SysUserSession) error

	// 获取用户所有活跃会话
	GetActiveSessions(userID uint) ([]*entity.SysUserSession, error)

	// 根据Token获取会话
	GetSessionByToken(token string) (*entity.SysUserSession, error)

	// 删除会话（登出）
	DeleteSession(id uint) error

	// 删除用户所有会话
	DeleteAllSessions(userID uint) error
}
