package menu

import (
	"context"
	"fmt"
	"sort"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// Service 菜单服务接口
type Service interface {
	// GetMenuTree 获取完整菜单树（三级结构）
	GetMenuTree(ctx context.Context, companyID uint) ([]*entity.MenuNode, error)

	// GetUserMenuTree 获取用户权限过滤后的菜单树
	GetUserMenuTree(ctx context.Context, userID, companyID uint) ([]*entity.MenuNode, error)
}

type service struct {
	db *gorm.DB
}

// NewService 创建菜单服务实例
func NewService(db *gorm.DB) Service {
	return &service{
		db: db,
	}
}

// GetMenuTree 获取完整菜单树
func (s *service) GetMenuTree(ctx context.Context, companyID uint) ([]*entity.MenuNode, error) {
	// 1. 查询所有子系统（一级菜单）
	var subsystems []entity.SysSubsystem
	if err := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ?", "Y").
		Order("ORDERNO ASC, ID ASC").
		Find(&subsystems).Error; err != nil {
		return nil, fmt.Errorf("查询子系统失败: %w", err)
	}

	// 2. 查询所有表类别（二级菜单）
	var categories []entity.SysTableCategory
	if err := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ?", "Y").
		Order("ORDERNO ASC, ID ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询表类别失败: %w", err)
	}

	// 3. 查询所有菜单表（三级菜单）
	var tables []entity.SysTable
	if err := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ? AND IS_MENU = ?", "Y", "Y").
		Order("ORDERNO ASC, ID ASC").
		Find(&tables).Error; err != nil {
		return nil, fmt.Errorf("查询菜单表失败: %w", err)
	}

	// 4. 构建树形结构
	return s.buildMenuTree(subsystems, categories, tables), nil
}

// GetUserMenuTree 获取用户权限过滤后的菜单树
func (s *service) GetUserMenuTree(ctx context.Context, userID, companyID uint) ([]*entity.MenuNode, error) {
	// 1. 检查用户是否是管理员
	var isAdmin string
	if err := s.db.WithContext(ctx).
		Table("sys_user").
		Select("IS_ADMIN").
		Where("ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Scan(&isAdmin).Error; err != nil {
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	// 如果是管理员，直接返回完整菜单树
	if isAdmin == "Y" {
		return s.GetMenuTree(ctx, companyID)
	}

	// 2. 查询用户所属的权限组
	var groupIDs []uint
	if err := s.db.WithContext(ctx).
		Table("sys_user_groups").
		Where("SYS_USER_ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Pluck("SYS_GROUPS_ID", &groupIDs).Error; err != nil {
		return nil, fmt.Errorf("查询用户权限组失败: %w", err)
	}

	// 如果用户没有权限组，返回空菜单
	if len(groupIDs) == 0 {
		return []*entity.MenuNode{}, nil
	}

	// 2. 查询权限组对应的目录ID
	var directoryIDs []uint
	if err := s.db.WithContext(ctx).
		Table("sys_group_prem").
		Where("SYS_GROUPS_ID IN ? AND IS_ACTIVE = ?", groupIDs, "Y").
		Distinct("SYS_DIRECTORY_ID").
		Pluck("SYS_DIRECTORY_ID", &directoryIDs).Error; err != nil {
		return nil, fmt.Errorf("查询权限目录失败: %w", err)
	}

	// 如果没有目录权限，返回空菜单
	if len(directoryIDs) == 0 {
		return []*entity.MenuNode{}, nil
	}

	// 3. 查询目录对应的表ID
	var allowedTableIDs []uint
	if err := s.db.WithContext(ctx).
		Table("sys_directory").
		Where("ID IN ? AND IS_ACTIVE = ?", directoryIDs, "Y").
		Pluck("SYS_TABLE_ID", &allowedTableIDs).Error; err != nil {
		return nil, fmt.Errorf("查询目录表失败: %w", err)
	}

	// 4. 查询所有子系统
	var subsystems []entity.SysSubsystem
	if err := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ?", "Y").
		Order("ORDERNO ASC, ID ASC").
		Find(&subsystems).Error; err != nil {
		return nil, fmt.Errorf("查询子系统失败: %w", err)
	}

	// 5. 查询所有表类别
	var categories []entity.SysTableCategory
	if err := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ?", "Y").
		Order("ORDERNO ASC, ID ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询表类别失败: %w", err)
	}

	// 6. 只查询用户有权限的菜单表
	var tables []entity.SysTable
	query := s.db.WithContext(ctx).
		Where("IS_ACTIVE = ? AND IS_MENU = ?", "Y", "Y")

	if len(allowedTableIDs) > 0 {
		query = query.Where("ID IN ?", allowedTableIDs)
	} else {
		// 没有权限，返回空
		return []*entity.MenuNode{}, nil
	}

	if err := query.Order("ORDERNO ASC, ID ASC").Find(&tables).Error; err != nil {
		return nil, fmt.Errorf("查询菜单表失败: %w", err)
	}

	// 7. 构建树形结构（会自动过滤空分支）
	return s.buildMenuTree(subsystems, categories, tables), nil
}

// buildMenuTree 构建三级菜单树
func (s *service) buildMenuTree(
	subsystems []entity.SysSubsystem,
	categories []entity.SysTableCategory,
	tables []entity.SysTable,
) []*entity.MenuNode {
	// 1. 将 tables 按 category 分组
	tablesByCategory := make(map[uint][]*entity.MenuNode)
	for _, table := range tables {
		if table.SysTableCategoryID != nil {
			categoryID := *table.SysTableCategoryID
			tablesByCategory[categoryID] = append(tablesByCategory[categoryID], &entity.MenuNode{
				ID:          table.ID,
				Name:        table.Name,
				DisplayName: table.DisplayName,
				Icon:        table.IcoImg,
				URL:         table.URL,
				OrderNo:     table.OrderNo,
				Type:        "table",
			})
		}
	}

	// 2. 将 categories 按 subsystem 分组，并添加子节点
	categoriesBySubsystem := make(map[uint][]*entity.MenuNode)
	for _, category := range categories {
		// 只添加有子菜单的分类
		if children, exists := tablesByCategory[category.ID]; exists && len(children) > 0 {
			// 排序子菜单
			sort.Slice(children, func(i, j int) bool {
				if children[i].OrderNo != children[j].OrderNo {
					return children[i].OrderNo < children[j].OrderNo
				}
				return children[i].ID < children[j].ID
			})

			categoriesBySubsystem[category.SysSubsystemID] = append(
				categoriesBySubsystem[category.SysSubsystemID],
				&entity.MenuNode{
					ID:       category.ID,
					Name:     category.Name,
					Icon:     category.Icon,
					URL:      category.URL,
					OrderNo:  category.OrderNo,
					Type:     "category",
					Children: children,
				},
			)
		}
	}

	// 3. 构建最终的树形结构
	var result []*entity.MenuNode
	for _, subsystem := range subsystems {
		// 只添加有子菜单的子系统
		if children, exists := categoriesBySubsystem[subsystem.ID]; exists && len(children) > 0 {
			// 排序分类
			sort.Slice(children, func(i, j int) bool {
				if children[i].OrderNo != children[j].OrderNo {
					return children[i].OrderNo < children[j].OrderNo
				}
				return children[i].ID < children[j].ID
			})

			result = append(result, &entity.MenuNode{
				ID:       subsystem.ID,
				Name:     subsystem.Name,
				Icon:     subsystem.Icon,
				URL:      subsystem.URL,
				OrderNo:  subsystem.OrderNo,
				Type:     "subsystem",
				Children: children,
			})
		}
	}

	// 4. 排序子系统
	sort.Slice(result, func(i, j int) bool {
		if result[i].OrderNo != result[j].OrderNo {
			return result[i].OrderNo < result[j].OrderNo
		}
		return result[i].ID < result[j].ID
	})

	return result
}
