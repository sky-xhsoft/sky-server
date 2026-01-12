package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/menu"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	menuService menu.Service
}

// NewMenuHandler 创建菜单处理器实例
func NewMenuHandler(menuService menu.Service) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
	}
}

// GetMenuTree 获取完整菜单树
// @Summary 获取完整菜单树
// @Description 获取系统完整的三级菜单结构（子系统-表类别-表单）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "{"code":200,"data":[...],"message":"success"}"
// @Failure 401 {object} map[string]interface{} "{"code":401,"message":"未授权"}"
// @Failure 500 {object} map[string]interface{} "{"code":500,"message":"服务器错误"}"
// @Router /menus/tree [get]
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	// 从上下文获取公司ID
	companyID, exists := c.Get("companyID")
	if !exists {
		utils.Unauthorized(c, "未找到公司信息")
		return
	}

	// 获取菜单树
	menuTree, err := h.menuService.GetMenuTree(c.Request.Context(), companyID.(uint))
	if err != nil {
		utils.InternalError(c, "获取菜单树失败: "+err.Error())
		return
	}

	utils.Success(c, menuTree)
}

// GetUserMenuTree 获取用户菜单树
// @Summary 获取用户菜单树
// @Description 获取当前用户权限过滤后的菜单树
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "{"code":200,"data":[...],"message":"success"}"
// @Failure 401 {object} map[string]interface{} "{"code":401,"message":"未授权"}"
// @Failure 500 {object} map[string]interface{} "{"code":500,"message":"服务器错误"}"
// @Router /menus/user/tree [get]
func (h *MenuHandler) GetUserMenuTree(c *gin.Context) {
	// 从上下文获取用户ID和公司ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未找到用户信息")
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		utils.Unauthorized(c, "未找到公司信息")
		return
	}

	// 获取用户菜单树
	menuTree, err := h.menuService.GetUserMenuTree(
		c.Request.Context(),
		userID.(uint),
		companyID.(uint),
	)
	if err != nil {
		utils.InternalError(c, "获取用户菜单树失败: "+err.Error())
		return
	}

	utils.Success(c, menuTree)
}
