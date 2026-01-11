package menu

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"gorm.io/gorm"
)

// Service 菜单服务接口
type Service interface {
	// 创建菜单
	CreateMenu(ctx context.Context, menu *entity.Menu) error

	// 更新菜单
	UpdateMenu(ctx context.Context, menu *entity.Menu) error

	// 删除菜单
	DeleteMenu(ctx context.Context, id uint) error

	// 获取菜单
	GetMenu(ctx context.Context, id uint) (*entity.Menu, error)

	// 查询菜单列表
	ListMenus(ctx context.Context, req *ListMenusRequest) ([]*entity.Menu, int64, error)

	// 获取菜单树
	GetMenuTree(ctx context.Context, parentID uint) ([]*entity.MenuNode, error)

	// 获取用户菜单树(根据权限过滤)
	GetUserMenuTree(ctx context.Context, userID uint) ([]*entity.MenuNode, error)

	// 获取用户路由(前端路由配置)
	GetUserRouters(ctx context.Context, userID uint) ([]*entity.RouterVO, error)

	// 获取角色菜单ID列表
	GetRoleMenuIDs(ctx context.Context, roleID uint) ([]uint, error)

	// 分配菜单给角色
	AssignMenusToRole(ctx context.Context, roleID uint, menuIDs []uint) error
}

// ListMenusRequest 查询菜单列表请求
type ListMenusRequest struct {
	MenuName  string // 菜单名称(模糊查询)
	MenuType  string // 菜单类型
	Status    string // 状态
	ParentID  *uint  // 父菜单ID
	Page      int    // 页码
	PageSize  int    // 每页大小
	SortBy    string // 排序字段
	SortOrder string // 排序方向
}

// service 菜单服务实现
type service struct {
	db *gorm.DB
}

// NewService 创建菜单服务
func NewService(db *gorm.DB) Service {
	return &service{
		db: db,
	}
}

// CreateMenu 创建菜单
func (s *service) CreateMenu(ctx context.Context, menu *entity.Menu) error {
	if err := s.db.WithContext(ctx).Create(menu).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建菜单失败", err)
	}

	return nil
}

// UpdateMenu 更新菜单
func (s *service) UpdateMenu(ctx context.Context, menu *entity.Menu) error {
	// 检查菜单是否存在
	if _, err := s.GetMenu(ctx, menu.ID); err != nil {
		return err
	}

	// 不能将自己设置为父菜单
	if menu.ParentID == menu.ID {
		return errors.New(errors.ErrValidation, "不能将自己设置为父菜单")
	}

	if err := s.db.WithContext(ctx).Model(&entity.Menu{}).Where("ID = ?", menu.ID).Updates(menu).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新菜单失败", err)
	}

	return nil
}

// DeleteMenu 删除菜单
func (s *service) DeleteMenu(ctx context.Context, id uint) error {
	// 检查菜单是否存在
	if _, err := s.GetMenu(ctx, id); err != nil {
		return err
	}

	// 检查是否有子菜单
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.Menu{}).
		Where("PARENT_ID = ? AND IS_ACTIVE = ?", id, "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查子菜单失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrValidation, "该菜单存在子菜单,无法删除")
	}

	// 软删除
	if err := s.db.WithContext(ctx).Model(&entity.Menu{}).
		Where("ID = ?", id).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除菜单失败", err)
	}

	return nil
}

// GetMenu 获取菜单
func (s *service) GetMenu(ctx context.Context, id uint) (*entity.Menu, error) {
	var menu entity.Menu
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
		First(&menu).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "菜单不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询菜单失败", err)
	}

	return &menu, nil
}

// ListMenus 查询菜单列表
func (s *service) ListMenus(ctx context.Context, req *ListMenusRequest) ([]*entity.Menu, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	query := s.db.WithContext(ctx).Model(&entity.Menu{}).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	if req.MenuName != "" {
		query = query.Where("MENU_NAME LIKE ?", "%"+req.MenuName+"%")
	}
	if req.MenuType != "" {
		query = query.Where("MENU_TYPE = ?", req.MenuType)
	}
	if req.Status != "" {
		query = query.Where("STATUS = ?", req.Status)
	}
	if req.ParentID != nil {
		query = query.Where("PARENT_ID = ?", *req.ParentID)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询菜单总数失败", err)
	}

	// 排序
	sortBy := "SORT_ORDER"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "ASC"
	if req.SortOrder == "DESC" {
		sortOrder = "DESC"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var menus []*entity.Menu
	if err := query.Limit(req.PageSize).Offset(offset).Find(&menus).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询菜单列表失败", err)
	}

	return menus, total, nil
}

// GetMenuTree 获取菜单树
func (s *service) GetMenuTree(ctx context.Context, parentID uint) ([]*entity.MenuNode, error) {
	// 查询所有菜单
	var menus []*entity.Menu
	query := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ? AND STATUS = ?", "Y", entity.MenuStatusEnabled).
		Order("SORT_ORDER ASC")

	if parentID > 0 {
		// 查询指定父菜单下的菜单树
		query = query.Where("PARENT_ID = ? OR ID = ?", parentID, parentID)
	}

	if err := query.Find(&menus).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询菜单列表失败", err)
	}

	// 构建菜单树
	return s.buildMenuTree(menus, parentID), nil
}

// GetUserMenuTree 获取用户菜单树(根据权限过滤)
func (s *service) GetUserMenuTree(ctx context.Context, userID uint) ([]*entity.MenuNode, error) {
	// 查询用户有权限的菜单
	var menus []*entity.Menu

	err := s.db.WithContext(ctx).
		Table("sys_menu m").
		Distinct("m.*").
		Joins("LEFT JOIN sys_permission p ON m.PERM_CODE = p.PERM_CODE").
		Joins("LEFT JOIN sys_role_permission rp ON p.ID = rp.PERMISSION_ID").
		Joins("LEFT JOIN sys_user_role ur ON rp.ROLE_ID = ur.ROLE_ID").
		Where("m.IS_ACTIVE = ? AND m.STATUS = ? AND m.IS_VISIBLE = ?", "Y", entity.MenuStatusEnabled, "Y").
		Where("(ur.USER_ID = ? AND ur.IS_ACTIVE = ? AND rp.IS_ACTIVE = ?) OR m.PERM_CODE IS NULL OR m.PERM_CODE = ''",
			userID, "Y", "Y").
		Order("m.SORT_ORDER ASC").
		Find(&menus).Error

	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询用户菜单失败", err)
	}

	// 构建菜单树
	return s.buildMenuTree(menus, 0), nil
}

// GetUserRouters 获取用户路由(前端路由配置)
func (s *service) GetUserRouters(ctx context.Context, userID uint) ([]*entity.RouterVO, error) {
	// 获取用户菜单树
	menuTree, err := s.GetUserMenuTree(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为路由对象
	return s.buildRouters(menuTree), nil
}

// GetRoleMenuIDs 获取角色菜单ID列表
func (s *service) GetRoleMenuIDs(ctx context.Context, roleID uint) ([]uint, error) {
	var menuIDs []uint

	// 通过角色权限查询菜单
	err := s.db.WithContext(ctx).
		Table("sys_menu m").
		Select("m.ID").
		Joins("INNER JOIN sys_permission p ON m.PERM_CODE = p.PERM_CODE").
		Joins("INNER JOIN sys_role_permission rp ON p.ID = rp.PERMISSION_ID").
		Where("rp.ROLE_ID = ? AND rp.IS_ACTIVE = ? AND m.IS_ACTIVE = ?",
			roleID, "Y", "Y").
		Pluck("m.ID", &menuIDs).Error

	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询角色菜单失败", err)
	}

	return menuIDs, nil
}

// AssignMenusToRole 分配菜单给角色
func (s *service) AssignMenusToRole(ctx context.Context, roleID uint, menuIDs []uint) error {
	// 查询菜单关联的权限
	var permCodes []string
	err := s.db.WithContext(ctx).
		Table("sys_menu").
		Where("ID IN ? AND PERM_CODE IS NOT NULL AND PERM_CODE != ''", menuIDs).
		Pluck("PERM_CODE", &permCodes).Error

	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询菜单权限失败", err)
	}

	// 查询权限ID
	var permissionIDs []uint
	if len(permCodes) > 0 {
		err = s.db.WithContext(ctx).
			Table("sys_permission").
			Where("PERM_CODE IN ?", permCodes).
			Pluck("ID", &permissionIDs).Error

		if err != nil {
			return errors.Wrap(errors.ErrDatabase, "查询权限ID失败", err)
		}
	}

	// 这里应该调用角色服务的分配权限方法
	// 为了简化,这里直接返回权限ID列表
	// 实际应用中,应该在handler层调用roleService.AssignPermissions

	return nil
}

// buildMenuTree 构建菜单树
func (s *service) buildMenuTree(menus []*entity.Menu, parentID uint) []*entity.MenuNode {
	// 构建菜单映射
	menuMap := make(map[uint]*entity.MenuNode)
	for _, menu := range menus {
		menuMap[menu.ID] = &entity.MenuNode{
			Menu:     menu,
			Children: make([]*entity.MenuNode, 0),
		}
	}

	// 构建树结构
	var tree []*entity.MenuNode
	for _, node := range menuMap {
		if node.ParentID == parentID {
			// 根节点
			tree = append(tree, node)
		} else {
			// 子节点
			if parent, exists := menuMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return tree
}

// buildRouters 构建路由对象
func (s *service) buildRouters(menuNodes []*entity.MenuNode) []*entity.RouterVO {
	routers := make([]*entity.RouterVO, 0)

	for _, node := range menuNodes {
		// 按钮类型不需要生成路由
		if node.MenuType == entity.MenuTypeButton {
			continue
		}

		router := &entity.RouterVO{
			Name:      node.MenuName,
			Path:      node.Path,
			Hidden:    node.IsVisible != "Y",
			Redirect:  node.Redirect,
			Component: node.Component,
			Meta: entity.Meta{
				Title:      node.MenuName,
				Icon:       node.Icon,
				NoCache:    node.IsCache != "Y",
				AlwaysShow: node.AlwaysShow == "Y",
				Hidden:     node.IsVisible != "Y",
			},
		}

		// 递归构建子路由
		if len(node.Children) > 0 {
			router.Children = s.buildRouters(node.Children)
		}

		routers = append(routers, router)
	}

	return routers
}
