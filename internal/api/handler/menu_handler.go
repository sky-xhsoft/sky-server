package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/menu"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	menuService menu.Service
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(menuService menu.Service) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
	}
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	MenuName   string `json:"menuName" binding:"required"`
	ParentID   uint   `json:"parentId"`
	MenuType   string `json:"menuType" binding:"required,oneof=dir menu button"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	PermCode   string `json:"permCode"`
	Icon       string `json:"icon"`
	SortOrder  int    `json:"sortOrder"`
	IsVisible  string `json:"isVisible" binding:"oneof=Y N"`
	IsCache    string `json:"isCache" binding:"oneof=Y N"`
	IsFrame    string `json:"isFrame" binding:"oneof=Y N"`
	Status     string `json:"status" binding:"required,oneof=enabled disabled"`
	Redirect   string `json:"redirect"`
	AlwaysShow string `json:"alwaysShow" binding:"oneof=Y N"`
	Remark     string `json:"remark"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	MenuName   string `json:"menuName"`
	ParentID   uint   `json:"parentId"`
	MenuType   string `json:"menuType" binding:"omitempty,oneof=dir menu button"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	PermCode   string `json:"permCode"`
	Icon       string `json:"icon"`
	SortOrder  int    `json:"sortOrder"`
	IsVisible  string `json:"isVisible" binding:"omitempty,oneof=Y N"`
	IsCache    string `json:"isCache" binding:"omitempty,oneof=Y N"`
	IsFrame    string `json:"isFrame" binding:"omitempty,oneof=Y N"`
	Status     string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	Redirect   string `json:"redirect"`
	AlwaysShow string `json:"alwaysShow" binding:"omitempty,oneof=Y N"`
	Remark     string `json:"remark"`
}

// CreateMenu 创建菜单
// @Summary 创建菜单
// @Description 创建新的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param menu body CreateMenuRequest true "菜单信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /menus [post]
// @Security BearerAuth
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户信息
	username, _ := c.Get("username")

	menuEntity := &entity.Menu{
		MenuName:   req.MenuName,
		ParentID:   req.ParentID,
		MenuType:   req.MenuType,
		Path:       req.Path,
		Component:  req.Component,
		PermCode:   req.PermCode,
		Icon:       req.Icon,
		SortOrder:  req.SortOrder,
		IsVisible:  req.IsVisible,
		IsCache:    req.IsCache,
		IsFrame:    req.IsFrame,
		Status:     req.Status,
		Redirect:   req.Redirect,
		AlwaysShow: req.AlwaysShow,
		Remark:     req.Remark,
		CreateBy:   username.(string),
		UpdateBy:   username.(string),
		IsActive:   "Y",
	}

	// 如果字段为空，设置默认值
	if menuEntity.IsVisible == "" {
		menuEntity.IsVisible = "Y"
	}
	if menuEntity.IsCache == "" {
		menuEntity.IsCache = "N"
	}
	if menuEntity.IsFrame == "" {
		menuEntity.IsFrame = "N"
	}
	if menuEntity.AlwaysShow == "" {
		menuEntity.AlwaysShow = "N"
	}

	if err := h.menuService.CreateMenu(c.Request.Context(), menuEntity); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "创建菜单失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id": menuEntity.ID,
		},
	})
}

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Description 更新菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Param menu body UpdateMenuRequest true "菜单信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /menus/{id} [put]
// @Security BearerAuth
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的菜单ID",
		})
		return
	}

	var req UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	menuEntity := &entity.Menu{
		ID:         uint(id),
		MenuName:   req.MenuName,
		ParentID:   req.ParentID,
		MenuType:   req.MenuType,
		Path:       req.Path,
		Component:  req.Component,
		PermCode:   req.PermCode,
		Icon:       req.Icon,
		SortOrder:  req.SortOrder,
		IsVisible:  req.IsVisible,
		IsCache:    req.IsCache,
		IsFrame:    req.IsFrame,
		Status:     req.Status,
		Redirect:   req.Redirect,
		AlwaysShow: req.AlwaysShow,
		Remark:     req.Remark,
		UpdateBy:   username.(string),
	}

	if err := h.menuService.UpdateMenu(c.Request.Context(), menuEntity); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "更新菜单失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteMenu 删除菜单
// @Summary 删除菜单
// @Description 删除菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /menus/{id} [delete]
// @Security BearerAuth
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的菜单ID",
		})
		return
	}

	if err := h.menuService.DeleteMenu(c.Request.Context(), uint(id)); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "删除菜单失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetMenu 获取菜单详情
// @Summary 获取菜单详情
// @Description 根据ID获取菜单详情
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /menus/{id} [get]
// @Security BearerAuth
func (h *MenuHandler) GetMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的菜单ID",
		})
		return
	}

	menuEntity, err := h.menuService.GetMenu(c.Request.Context(), uint(id))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "获取菜单失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    menuEntity,
	})
}

// ListMenus 查询菜单列表
// @Summary 查询菜单列表
// @Description 查询菜单列表,支持分页和过滤
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param menuName query string false "菜单名称(模糊查询)"
// @Param menuType query string false "菜单类型"
// @Param status query string false "状态"
// @Param parentId query int false "父菜单ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /menus [get]
// @Security BearerAuth
func (h *MenuHandler) ListMenus(c *gin.Context) {
	req := &menu.ListMenusRequest{
		MenuName: c.Query("menuName"),
		MenuType: c.Query("menuType"),
		Status:   c.Query("status"),
	}

	// 处理parentId
	if parentIDStr := c.Query("parentId"); parentIDStr != "" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil {
			pid := uint(parentID)
			req.ParentID = &pid
		}
	}

	// 处理分页参数
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			req.Page = p
		}
	}
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			req.PageSize = ps
		}
	}

	menus, total, err := h.menuService.ListMenus(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "查询菜单列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  menus,
			"total": total,
			"page":  req.Page,
			"size":  req.PageSize,
		},
	})
}

// GetMenuTree 获取菜单树
// @Summary 获取菜单树
// @Description 获取菜单树结构
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param parentId query int false "父菜单ID" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /menus/tree [get]
// @Security BearerAuth
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	var parentID uint
	if parentIDStr := c.Query("parentId"); parentIDStr != "" {
		if pid, err := strconv.ParseUint(parentIDStr, 10, 32); err == nil {
			parentID = uint(pid)
		}
	}

	tree, err := h.menuService.GetMenuTree(c.Request.Context(), parentID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "获取菜单树失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    tree,
	})
}

// GetUserMenuTree 获取当前用户菜单树
// @Summary 获取当前用户菜单树
// @Description 根据用户权限获取菜单树
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /menus/user/tree [get]
// @Security BearerAuth
func (h *MenuHandler) GetUserMenuTree(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    errors.ErrUnauthorized,
			"message": "未登录",
		})
		return
	}

	tree, err := h.menuService.GetUserMenuTree(c.Request.Context(), userID.(uint))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "获取用户菜单树失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    tree,
	})
}

// GetUserRouters 获取当前用户路由
// @Summary 获取当前用户路由
// @Description 获取当前用户的前端路由配置
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /menus/user/routers [get]
// @Security BearerAuth
func (h *MenuHandler) GetUserRouters(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    errors.ErrUnauthorized,
			"message": "未登录",
		})
		return
	}

	routers, err := h.menuService.GetUserRouters(c.Request.Context(), userID.(uint))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "获取用户路由失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    routers,
	})
}
