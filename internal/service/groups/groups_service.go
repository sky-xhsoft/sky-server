package groups

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// 权限位定义
const (
	PermNone   = 0      // 无权限
	PermRead   = 1 << 0 // 1 - 读取
	PermCreate = 1 << 1 // 2 - 创建
	PermUpdate = 1 << 2 // 4 - 更新
	PermDelete = 1 << 3 // 8 - 删除
	PermExport = 1 << 4 // 16 - 导出
	PermImport = 1 << 5 // 32 - 导入
	PermAll    = 63     // 111111 - 所有权限
)

// Service 权限组服务接口
type Service interface {
	// 权限组管理
	CreateGroup(ctx context.Context, group *entity.SysGroups) error
	UpdateGroup(ctx context.Context, group *entity.SysGroups) error
	DeleteGroup(ctx context.Context, id uint) error
	GetGroup(ctx context.Context, id uint) (*entity.SysGroups, error)
	ListGroups(ctx context.Context, req *ListGroupsRequest) ([]*entity.SysGroups, int64, error)

	// 安全目录管理
	CreateDirectory(ctx context.Context, dir *entity.SysDirectory) error
	UpdateDirectory(ctx context.Context, dir *entity.SysDirectory) error
	DeleteDirectory(ctx context.Context, id uint) error
	GetDirectory(ctx context.Context, id uint) (*entity.SysDirectory, error)
	ListDirectories(ctx context.Context, req *ListDirectoriesRequest) ([]*entity.SysDirectory, int64, error)
	GetDirectoryTree(ctx context.Context, parentID *uint) ([]*DirectoryNode, error)

	// 权限组明细管理
	AssignPermissions(ctx context.Context, groupID uint, permissions []*GroupPermission) error
	GetGroupPermissions(ctx context.Context, groupID uint) ([]*entity.SysGroupPrem, error)
	RemovePermissions(ctx context.Context, groupID uint, directoryIDs []uint) error

	// 用户权限组管理
	AssignGroupsToUser(ctx context.Context, userID uint, directoryIDs []uint) error
	GetUserGroups(ctx context.Context, userID uint) ([]*entity.SysGroups, error)
	RemoveUserGroups(ctx context.Context, userID uint, directoryIDs []uint) error

	// 权限检查
	CheckUserPermission(ctx context.Context, userID uint, directoryID uint, permission int) (bool, error)
	GetUserDirectoryPermission(ctx context.Context, userID uint, directoryID uint) (int, error)
	CheckUserTablePermission(ctx context.Context, userID uint, tableID uint, permission int) (bool, error)
	GetUserDataFilter(ctx context.Context, userID uint, directoryID uint) (map[string]interface{}, error)
}

// ListGroupsRequest 查询权限组请求
type ListGroupsRequest struct {
	Name     string
	Page     int
	PageSize int
}

// ListDirectoriesRequest 查询目录请求
type ListDirectoriesRequest struct {
	Name     string
	TableID  *uint
	ParentID *uint
	Page     int
	PageSize int
}

// GroupPermission 权限组权限
type GroupPermission struct {
	DirectoryID uint   `json:"directoryId"`
	Permission  int    `json:"permission"`  // 位运算权限值
	FilterObj   string `json:"filterObj"`   // JSON格式的过滤条件
}

// DirectoryNode 目录树节点
type DirectoryNode struct {
	*entity.SysDirectory
	Children []*DirectoryNode `json:"children"`
}

// service 权限组服务实现
type service struct {
	db *gorm.DB
}

// NewService 创建权限组服务
func NewService(db *gorm.DB) Service {
	return &service{
		db: db,
	}
}

// CreateGroup 创建权限组
func (s *service) CreateGroup(ctx context.Context, group *entity.SysGroups) error {
	if err := s.db.WithContext(ctx).Create(group).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建权限组失败", err)
	}
	return nil
}

// UpdateGroup 更新权限组
func (s *service) UpdateGroup(ctx context.Context, group *entity.SysGroups) error {
	// 检查权限组是否存在
	if _, err := s.GetGroup(ctx, group.ID); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Model(&entity.SysGroups{}).
		Where("ID = ?", group.ID).Updates(group).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新权限组失败", err)
	}
	return nil
}

// DeleteGroup 删除权限组
func (s *service) DeleteGroup(ctx context.Context, id uint) error {
	// 检查权限组是否存在
	if _, err := s.GetGroup(ctx, id); err != nil {
		return err
	}

	// 检查是否有用户使用该权限组
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.SysUserGroups{}).
		Where("SYS_DIRECTORY_ID = ? AND IS_ACTIVE = ?", id, "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查权限组使用失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrValidation, "该权限组已分配给用户,无法删除")
	}

	// 软删除
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除权限组
		if err := tx.Model(&entity.SysGroups{}).
			Where("ID = ?", id).
			Update("IS_ACTIVE", "N").Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "删除权限组失败", err)
		}

		// 删除权限组明细
		if err := tx.Model(&entity.SysGroupPrem{}).
			Where("SYS_GROUPS_ID = ?", id).
			Update("IS_ACTIVE", "N").Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "删除权限组明细失败", err)
		}

		return nil
	})
}

// GetGroup 获取权限组
func (s *service) GetGroup(ctx context.Context, id uint) (*entity.SysGroups, error) {
	var group entity.SysGroups
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
		First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "权限组不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询权限组失败", err)
	}
	return &group, nil
}

// ListGroups 查询权限组列表
func (s *service) ListGroups(ctx context.Context, req *ListGroupsRequest) ([]*entity.SysGroups, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	query := s.db.WithContext(ctx).Model(&entity.SysGroups{}).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	if req.Name != "" {
		query = query.Where("NAME LIKE ?", "%"+req.Name+"%")
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询权限组总数失败", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var groups []*entity.SysGroups
	if err := query.Order("ID DESC").Limit(req.PageSize).Offset(offset).Find(&groups).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询权限组列表失败", err)
	}

	return groups, total, nil
}

// CreateDirectory 创建安全目录
func (s *service) CreateDirectory(ctx context.Context, dir *entity.SysDirectory) error {
	if err := s.db.WithContext(ctx).Create(dir).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建安全目录失败", err)
	}
	return nil
}

// UpdateDirectory 更新安全目录
func (s *service) UpdateDirectory(ctx context.Context, dir *entity.SysDirectory) error {
	// 检查目录是否存在
	if _, err := s.GetDirectory(ctx, dir.ID); err != nil {
		return err
	}

	// 不能将自己设置为父目录
	if dir.ParentID != nil && *dir.ParentID == dir.ID {
		return errors.New(errors.ErrValidation, "不能将自己设置为父目录")
	}

	if err := s.db.WithContext(ctx).Model(&entity.SysDirectory{}).
		Where("ID = ?", dir.ID).Updates(dir).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新安全目录失败", err)
	}
	return nil
}

// DeleteDirectory 删除安全目录
func (s *service) DeleteDirectory(ctx context.Context, id uint) error {
	// 检查目录是否存在
	if _, err := s.GetDirectory(ctx, id); err != nil {
		return err
	}

	// 检查是否有子目录
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.SysDirectory{}).
		Where("PARENT_ID = ? AND IS_ACTIVE = ?", id, "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查子目录失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrValidation, "该目录存在子目录,无法删除")
	}

	// 软删除
	if err := s.db.WithContext(ctx).Model(&entity.SysDirectory{}).
		Where("ID = ?", id).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除安全目录失败", err)
	}
	return nil
}

// GetDirectory 获取安全目录
func (s *service) GetDirectory(ctx context.Context, id uint) (*entity.SysDirectory, error) {
	var dir entity.SysDirectory
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
		First(&dir).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "安全目录不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询安全目录失败", err)
	}
	return &dir, nil
}

// ListDirectories 查询目录列表
func (s *service) ListDirectories(ctx context.Context, req *ListDirectoriesRequest) ([]*entity.SysDirectory, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	query := s.db.WithContext(ctx).Model(&entity.SysDirectory{}).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	if req.Name != "" {
		query = query.Where("NAME LIKE ?", "%"+req.Name+"%")
	}
	if req.TableID != nil {
		query = query.Where("SYS_TABLE_ID = ?", *req.TableID)
	}
	if req.ParentID != nil {
		query = query.Where("PARENT_ID = ?", *req.ParentID)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询目录总数失败", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var dirs []*entity.SysDirectory
	if err := query.Order("ORDERNO ASC").Limit(req.PageSize).Offset(offset).Find(&dirs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询目录列表失败", err)
	}

	return dirs, total, nil
}

// GetDirectoryTree 获取目录树
func (s *service) GetDirectoryTree(ctx context.Context, parentID *uint) ([]*DirectoryNode, error) {
	// 查询所有目录
	var dirs []*entity.SysDirectory
	query := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ?", "Y").
		Order("ORDERNO ASC")

	if parentID != nil && *parentID > 0 {
		// 查询指定父目录下的目录树
		query = query.Where("PARENT_ID = ? OR ID = ?", *parentID, *parentID)
	}

	if err := query.Find(&dirs).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询目录列表失败", err)
	}

	// 构建目录树
	pid := uint(0)
	if parentID != nil {
		pid = *parentID
	}
	return s.buildDirectoryTree(dirs, pid), nil
}

// buildDirectoryTree 构建目录树
func (s *service) buildDirectoryTree(dirs []*entity.SysDirectory, parentID uint) []*DirectoryNode {
	// 构建目录映射
	dirMap := make(map[uint]*DirectoryNode)
	for _, dir := range dirs {
		dirMap[dir.ID] = &DirectoryNode{
			SysDirectory: dir,
			Children:     make([]*DirectoryNode, 0),
		}
	}

	// 构建树结构
	var tree []*DirectoryNode
	for _, node := range dirMap {
		if node.ParentID == nil || *node.ParentID == parentID {
			// 根节点
			tree = append(tree, node)
		} else {
			// 子节点
			if parent, exists := dirMap[*node.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return tree
}

// AssignPermissions 分配权限给权限组
func (s *service) AssignPermissions(ctx context.Context, groupID uint, permissions []*GroupPermission) error {
	// 检查权限组是否存在
	if _, err := s.GetGroup(ctx, groupID); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除原有权限
		if err := tx.Model(&entity.SysGroupPrem{}).
			Where("SYS_GROUPS_ID = ?", groupID).
			Update("IS_ACTIVE", "N").Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "删除原有权限失败", err)
		}

		// 添加新权限
		for _, perm := range permissions {
			groupPerm := &entity.SysGroupPrem{
				BaseModel: entity.BaseModel{
					IsActive: "Y",
				},
				SysGroupsID:    groupID,
				SysDirectoryID: perm.DirectoryID,
				Permission:     perm.Permission,
				FilterObj:      perm.FilterObj,
			}
			if err := tx.Create(groupPerm).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "添加权限失败", err)
			}
		}

		return nil
	})
}

// GetGroupPermissions 获取权限组权限
func (s *service) GetGroupPermissions(ctx context.Context, groupID uint) ([]*entity.SysGroupPrem, error) {
	var permissions []*entity.SysGroupPrem
	if err := s.db.WithContext(ctx).
		Where("SYS_GROUPS_ID = ? AND IS_ACTIVE = ?", groupID, "Y").
		Find(&permissions).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询权限组权限失败", err)
	}
	return permissions, nil
}

// RemovePermissions 移除权限组权限
func (s *service) RemovePermissions(ctx context.Context, groupID uint, directoryIDs []uint) error {
	if err := s.db.WithContext(ctx).Model(&entity.SysGroupPrem{}).
		Where("SYS_GROUPS_ID = ? AND SYS_DIRECTORY_ID IN ?", groupID, directoryIDs).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "移除权限失败", err)
	}
	return nil
}

// AssignGroupsToUser 分配权限组给用户
func (s *service) AssignGroupsToUser(ctx context.Context, userID uint, directoryIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除原有分配
		if err := tx.Model(&entity.SysUserGroups{}).
			Where("SYS_USER_ID = ?", userID).
			Update("IS_ACTIVE", "N").Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "删除原有分配失败", err)
		}

		// 添加新分配
		for _, dirID := range directoryIDs {
			userGroup := &entity.SysUserGroups{
				BaseModel: entity.BaseModel{
					IsActive: "Y",
				},
				SysUserID:      userID,
				SysDirectoryID: dirID,
			}
			if err := tx.Create(userGroup).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "分配权限组失败", err)
			}
		}

		return nil
	})
}

// GetUserGroups 获取用户权限组
func (s *service) GetUserGroups(ctx context.Context, userID uint) ([]*entity.SysGroups, error) {
	var groups []*entity.SysGroups
	err := s.db.WithContext(ctx).
		Table("sys_groups").
		Joins("INNER JOIN sys_user_groups ON sys_groups.ID = sys_user_groups.SYS_DIRECTORY_ID").
		Where("sys_user_groups.SYS_USER_ID = ? AND sys_groups.IS_ACTIVE = ? AND sys_user_groups.IS_ACTIVE = ?",
			userID, "Y", "Y").
		Find(&groups).Error
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询用户权限组失败", err)
	}
	return groups, nil
}

// RemoveUserGroups 移除用户权限组
func (s *service) RemoveUserGroups(ctx context.Context, userID uint, directoryIDs []uint) error {
	if err := s.db.WithContext(ctx).Model(&entity.SysUserGroups{}).
		Where("SYS_USER_ID = ? AND SYS_DIRECTORY_ID IN ?", userID, directoryIDs).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "移除用户权限组失败", err)
	}
	return nil
}

// CheckUserPermission 检查用户权限
func (s *service) CheckUserPermission(ctx context.Context, userID uint, directoryID uint, permission int) (bool, error) {
	// 检查用户是否是管理员
	var isAdmin string
	if err := s.db.WithContext(ctx).
		Table("sys_user").
		Select("IS_ADMIN").
		Where("ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Scan(&isAdmin).Error; err != nil {
		return false, errors.Wrap(errors.ErrDatabase, "查询用户信息失败", err)
	}

	// 如果是管理员，直接返回有权限
	if isAdmin == "Y" {
		return true, nil
	}

	userPerm, err := s.GetUserDirectoryPermission(ctx, userID, directoryID)
	if err != nil {
		return false, err
	}

	// 位运算检查权限
	return (userPerm & permission) == permission, nil
}

// GetUserDirectoryPermission 获取用户在目录的权限值
func (s *service) GetUserDirectoryPermission(ctx context.Context, userID uint, directoryID uint) (int, error) {
	// 检查用户是否是管理员
	var isAdmin string
	if err := s.db.WithContext(ctx).
		Table("sys_user").
		Select("IS_ADMIN").
		Where("ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Scan(&isAdmin).Error; err != nil {
		return 0, errors.Wrap(errors.ErrDatabase, "查询用户信息失败", err)
	}

	// 如果是管理员，返回全部权限
	if isAdmin == "Y" {
		return PermAll, nil
	}

	var permission int
	err := s.db.WithContext(ctx).
		Table("sys_group_prem").
		Select("sys_group_prem.PERMISSION").
		Joins("INNER JOIN sys_user_groups ON sys_group_prem.SYS_GROUPS_ID = sys_user_groups.SYS_DIRECTORY_ID").
		Where("sys_user_groups.SYS_USER_ID = ? AND sys_group_prem.SYS_DIRECTORY_ID = ? AND sys_group_prem.IS_ACTIVE = ? AND sys_user_groups.IS_ACTIVE = ?",
			userID, directoryID, "Y", "Y").
		Scan(&permission).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, errors.Wrap(errors.ErrDatabase, "查询用户权限失败", err)
	}

	return permission, nil
}

// CheckUserTablePermission 检查用户表权限
func (s *service) CheckUserTablePermission(ctx context.Context, userID uint, tableID uint, permission int) (bool, error) {
	// 检查用户是否是管理员
	var isAdmin string
	if err := s.db.WithContext(ctx).
		Table("sys_user").
		Select("IS_ADMIN").
		Where("ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Scan(&isAdmin).Error; err != nil {
		return false, errors.Wrap(errors.ErrDatabase, "查询用户信息失败", err)
	}

	// 如果是管理员，直接返回有权限
	if isAdmin == "Y" {
		return true, nil
	}

	// 查询表关联的目录
	var dir entity.SysDirectory
	err := s.db.WithContext(ctx).
		Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", tableID, "Y").
		First(&dir).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 表没有关联目录，默认无权限
			return false, nil
		}
		return false, errors.Wrap(errors.ErrDatabase, "查询表目录失败", err)
	}

	// 检查目录权限
	return s.CheckUserPermission(ctx, userID, dir.ID, permission)
}

// GetUserDataFilter 获取用户数据过滤条件
func (s *service) GetUserDataFilter(ctx context.Context, userID uint, directoryID uint) (map[string]interface{}, error) {
	var filterObj sql.NullString
	err := s.db.WithContext(ctx).
		Table("sys_group_prem").
		Select("sys_group_prem.FILTER_OBJ").
		Joins("INNER JOIN sys_user_groups ON sys_group_prem.SYS_GROUPS_ID = sys_user_groups.SYS_DIRECTORY_ID").
		Where("sys_user_groups.SYS_USER_ID = ? AND sys_group_prem.SYS_DIRECTORY_ID = ? AND sys_group_prem.IS_ACTIVE = ? AND sys_user_groups.IS_ACTIVE = ?",
			userID, directoryID, "Y", "Y").
		Scan(&filterObj).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询数据过滤条件失败", err)
	}

	if !filterObj.Valid || filterObj.String == "" {
		return nil, nil
	}

	// 解析JSON
	var filter map[string]interface{}
	if err := json.Unmarshal([]byte(filterObj.String), &filter); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "解析过滤条件失败", err)
	}

	return filter, nil
}
