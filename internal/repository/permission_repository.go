package repository

import "github.com/sky-xhsoft/sky-server/internal/model/entity"

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// 获取用户所在的权限组
	GetUserGroups(userID uint) ([]*entity.SysUserGroups, error)

	// 获取权限组的详细权限
	GetGroupPermissions(groupID uint) ([]*entity.SysGroupPrem, error)

	// 获取安全目录信息
	GetDirectory(directoryID uint) (*entity.SysDirectory, error)

	// 获取用户在指定目录的权限
	GetUserDirectoryPermission(userID, directoryID uint) (*entity.SysGroupPrem, error)

	// 获取用户所有权限（包括所有权限组）
	GetUserAllPermissions(userID uint) ([]*entity.SysGroupPrem, error)

	// 获取权限组信息
	GetGroup(groupID uint) (*entity.SysGroups, error)

	// 获取用户所属的所有权限组信息
	GetUserGroupsInfo(userID uint) ([]*entity.SysGroups, error)
}
