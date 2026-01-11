package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// GroupsHandler 权限组处理器
type GroupsHandler struct {
	groupService groups.Service
}

// NewGroupsHandler 创建权限组处理器
func NewGroupsHandler(groupService groups.Service) *GroupsHandler {
	return &GroupsHandler{
		groupService: groupService,
	}
}

// CreateGroupRequest 创建权限组请求
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Sgrade      int    `json:"sgrade"`
}

// UpdateGroupRequest 更新权限组请求
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Sgrade      int    `json:"sgrade"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	Permissions []*groups.GroupPermission `json:"permissions" binding:"required"`
}

// AssignGroupsToUserRequest 分配权限组给用户请求
type AssignGroupsToUserRequest struct {
	DirectoryIDs []uint `json:"directoryIds" binding:"required"`
}

// CreateGroup 创建权限组
// @Summary 创建权限组
// @Description 创建新的权限组
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param group body CreateGroupRequest true "权限组信息"
// @Success 200 {object} map[string]interface{}
// @Router /groups [post]
// @Security BearerAuth
func (h *GroupsHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	group := &entity.SysGroups{
		BaseModel: entity.BaseModel{
			CreateBy: username.(string),
			UpdateBy: username.(string),
			IsActive: "Y",
		},
		Name:        req.Name,
		Description: req.Description,
		Sgrade:      req.Sgrade,
	}

	if err := h.groupService.CreateGroup(c.Request.Context(), group); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "创建权限组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id": group.ID,
		},
	})
}

// UpdateGroup 更新权限组
// @Summary 更新权限组
// @Description 更新权限组信息
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param id path int true "权限组ID"
// @Param group body UpdateGroupRequest true "权限组信息"
// @Success 200 {object} map[string]interface{}
// @Router /groups/{id} [put]
// @Security BearerAuth
func (h *GroupsHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的权限组ID",
		})
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	group := &entity.SysGroups{
		BaseModel: entity.BaseModel{
			ID:       uint(id),
			UpdateBy: username.(string),
		},
		Name:        req.Name,
		Description: req.Description,
		Sgrade:      req.Sgrade,
	}

	if err := h.groupService.UpdateGroup(c.Request.Context(), group); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "更新权限组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteGroup 删除权限组
// @Summary 删除权限组
// @Description 删除权限组
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param id path int true "权限组ID"
// @Success 200 {object} map[string]interface{}
// @Router /groups/{id} [delete]
// @Security BearerAuth
func (h *GroupsHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的权限组ID",
		})
		return
	}

	if err := h.groupService.DeleteGroup(c.Request.Context(), uint(id)); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "删除权限组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetGroup 获取权限组详情
// @Summary 获取权限组详情
// @Description 根据ID获取权限组详情
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param id path int true "权限组ID"
// @Success 200 {object} map[string]interface{}
// @Router /groups/{id} [get]
// @Security BearerAuth
func (h *GroupsHandler) GetGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的权限组ID",
		})
		return
	}

	group, err := h.groupService.GetGroup(c.Request.Context(), uint(id))
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
			"message": "获取权限组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    group,
	})
}

// ListGroups 查询权限组列表
// @Summary 查询权限组列表
// @Description 查询权限组列表,支持分页和过滤
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param name query string false "权限组名称(模糊查询)"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /groups [get]
// @Security BearerAuth
func (h *GroupsHandler) ListGroups(c *gin.Context) {
	req := &groups.ListGroupsRequest{
		Name: c.Query("name"),
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

	groupList, total, err := h.groupService.ListGroups(c.Request.Context(), req)
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
			"message": "查询权限组列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  groupList,
			"total": total,
			"page":  req.Page,
			"size":  req.PageSize,
		},
	})
}

// AssignPermissions 分配权限给权限组
// @Summary 分配权限给权限组
// @Description 分配权限给权限组
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param id path int true "权限组ID"
// @Param permissions body AssignPermissionsRequest true "权限列表"
// @Success 200 {object} map[string]interface{}
// @Router /groups/{id}/permissions [post]
// @Security BearerAuth
func (h *GroupsHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的权限组ID",
		})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.groupService.AssignPermissions(c.Request.Context(), uint(id), req.Permissions); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "分配权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetGroupPermissions 获取权限组的权限列表
// @Summary 获取权限组的权限列表
// @Description 获取权限组的权限列表
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param id path int true "权限组ID"
// @Success 200 {object} map[string]interface{}
// @Router /groups/{id}/permissions [get]
// @Security BearerAuth
func (h *GroupsHandler) GetGroupPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的权限组ID",
		})
		return
	}

	permissions, err := h.groupService.GetGroupPermissions(c.Request.Context(), uint(id))
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
			"message": "查询权限列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    permissions,
	})
}

// AssignGroupsToUser 分配权限组给用户
// @Summary 分配权限组给用户
// @Description 分配权限组给用户
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Param groups body AssignGroupsToUserRequest true "目录ID列表"
// @Success 200 {object} map[string]interface{}
// @Router /groups/users/{userId} [post]
// @Security BearerAuth
func (h *GroupsHandler) AssignGroupsToUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的用户ID",
		})
		return
	}

	var req AssignGroupsToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.groupService.AssignGroupsToUser(c.Request.Context(), uint(userID), req.DirectoryIDs); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "分配权限组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetUserGroups 获取用户的权限组列表
// @Summary 获取用户的权限组列表
// @Description 获取用户的权限组列表
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /groups/users/{userId} [get]
// @Security BearerAuth
func (h *GroupsHandler) GetUserGroups(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的用户ID",
		})
		return
	}

	groupList, err := h.groupService.GetUserGroups(c.Request.Context(), uint(userID))
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
			"message": "查询用户权限组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    groupList,
	})
}

// CheckPermission 检查用户权限
// @Summary 检查用户权限
// @Description 检查用户是否有指定的权限
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param directoryId query int true "目录ID"
// @Param permission query int true "权限值"
// @Success 200 {object} map[string]interface{}
// @Router /permissions/check [post]
// @Security BearerAuth
func (h *GroupsHandler) CheckPermission(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    errors.ErrUnauthorized,
			"message": "未登录",
		})
		return
	}

	directoryID, err := strconv.ParseUint(c.Query("directoryId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "目录ID参数错误",
		})
		return
	}

	permission, err := strconv.Atoi(c.Query("permission"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "权限值参数错误",
		})
		return
	}

	hasPermission, err := h.groupService.CheckUserPermission(
		c.Request.Context(),
		userID.(uint),
		uint(directoryID),
		permission,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "检查权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"hasPermission": hasPermission,
		},
	})
}

// GetUserPermission 获取用户权限
// @Summary 获取用户权限
// @Description 获取用户在指定目录的权限值
// @Tags 权限组管理
// @Accept json
// @Produce json
// @Param directoryId query int false "目录ID"
// @Success 200 {object} map[string]interface{}
// @Router /permissions/user [get]
// @Security BearerAuth
func (h *GroupsHandler) GetUserPermission(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    errors.ErrUnauthorized,
			"message": "未登录",
		})
		return
	}

	directoryIDStr := c.Query("directoryId")
	if directoryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "目录ID参数缺失",
		})
		return
	}

	directoryID, err := strconv.ParseUint(directoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "目录ID参数错误",
		})
		return
	}

	permission, err := h.groupService.GetUserDirectoryPermission(
		c.Request.Context(),
		userID.(uint),
		uint(directoryID),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "获取用户权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"permission": permission,
			"permissions": map[string]bool{
				"read":   (permission & groups.PermRead) > 0,
				"create": (permission & groups.PermCreate) > 0,
				"update": (permission & groups.PermUpdate) > 0,
				"delete": (permission & groups.PermDelete) > 0,
				"export": (permission & groups.PermExport) > 0,
				"import": (permission & groups.PermImport) > 0,
			},
		},
	})
}
