package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// DirectoryHandler 安全目录处理器
type DirectoryHandler struct {
	groupService groups.Service
}

// NewDirectoryHandler 创建安全目录处理器
func NewDirectoryHandler(groupService groups.Service) *DirectoryHandler {
	return &DirectoryHandler{
		groupService: groupService,
	}
}

// CreateDirectoryRequest 创建安全目录请求
type CreateDirectoryRequest struct {
	Name        string `json:"name" binding:"required"`
	SysTableID  *uint  `json:"sysTableId"`
	ParentID    *uint  `json:"parentId"`
	Orderno     int    `json:"orderno"`
	Description string `json:"description"`
}

// UpdateDirectoryRequest 更新安全目录请求
type UpdateDirectoryRequest struct {
	Name        string `json:"name"`
	SysTableID  *uint  `json:"sysTableId"`
	ParentID    *uint  `json:"parentId"`
	Orderno     int    `json:"orderno"`
	Description string `json:"description"`
}

// CreateDirectory 创建安全目录
// @Summary 创建安全目录
// @Description 创建新的安全目录
// @Tags 安全目录管理
// @Accept json
// @Produce json
// @Param directory body CreateDirectoryRequest true "安全目录信息"
// @Success 200 {object} map[string]interface{}
// @Router /directories [post]
// @Security BearerAuth
func (h *DirectoryHandler) CreateDirectory(c *gin.Context) {
	var req CreateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	dir := &entity.SysDirectory{
		BaseModel: entity.BaseModel{
			CreateBy: username.(string),
			UpdateBy: username.(string),
			IsActive: "Y",
		},
		Name:        req.Name,
		SysTableID:  req.SysTableID,
		ParentID:    req.ParentID,
		Orderno:     req.Orderno,
		Description: req.Description,
	}

	if err := h.groupService.CreateDirectory(c.Request.Context(), dir); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "创建安全目录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id": dir.ID,
		},
	})
}

// UpdateDirectory 更新安全目录
// @Summary 更新安全目录
// @Description 更新安全目录信息
// @Tags 安全目录管理
// @Accept json
// @Produce json
// @Param id path int true "目录ID"
// @Param directory body UpdateDirectoryRequest true "安全目录信息"
// @Success 200 {object} map[string]interface{}
// @Router /directories/{id} [put]
// @Security BearerAuth
func (h *DirectoryHandler) UpdateDirectory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的目录ID",
		})
		return
	}

	var req UpdateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	dir := &entity.SysDirectory{
		BaseModel: entity.BaseModel{
			ID:       uint(id),
			UpdateBy: username.(string),
		},
		Name:        req.Name,
		SysTableID:  req.SysTableID,
		ParentID:    req.ParentID,
		Orderno:     req.Orderno,
		Description: req.Description,
	}

	if err := h.groupService.UpdateDirectory(c.Request.Context(), dir); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "更新安全目录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteDirectory 删除安全目录
// @Summary 删除安全目录
// @Description 删除安全目录
// @Tags 安全目录管理
// @Accept json
// @Produce json
// @Param id path int true "目录ID"
// @Success 200 {object} map[string]interface{}
// @Router /directories/{id} [delete]
// @Security BearerAuth
func (h *DirectoryHandler) DeleteDirectory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的目录ID",
		})
		return
	}

	if err := h.groupService.DeleteDirectory(c.Request.Context(), uint(id)); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "删除安全目录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetDirectory 获取安全目录详情
// @Summary 获取安全目录详情
// @Description 根据ID获取安全目录详情
// @Tags 安全目录管理
// @Accept json
// @Produce json
// @Param id path int true "目录ID"
// @Success 200 {object} map[string]interface{}
// @Router /directories/{id} [get]
// @Security BearerAuth
func (h *DirectoryHandler) GetDirectory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的目录ID",
		})
		return
	}

	dir, err := h.groupService.GetDirectory(c.Request.Context(), uint(id))
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
			"message": "获取安全目录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    dir,
	})
}

// ListDirectories 查询安全目录列表
// @Summary 查询安全目录列表
// @Description 查询安全目录列表,支持分页和过滤
// @Tags 安全目录管理
// @Accept json
// @Produce json
// @Param name query string false "目录名称(模糊查询)"
// @Param tableId query int false "表ID"
// @Param parentId query int false "父目录ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /directories [get]
// @Security BearerAuth
func (h *DirectoryHandler) ListDirectories(c *gin.Context) {
	req := &groups.ListDirectoriesRequest{
		Name: c.Query("name"),
	}

	// 处理tableId
	if tableIDStr := c.Query("tableId"); tableIDStr != "" {
		tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
		if err == nil {
			tid := uint(tableID)
			req.TableID = &tid
		}
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

	dirList, total, err := h.groupService.ListDirectories(c.Request.Context(), req)
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
			"message": "查询目录列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  dirList,
			"total": total,
			"page":  req.Page,
			"size":  req.PageSize,
		},
	})
}

// GetDirectoryTree 获取目录树
// @Summary 获取目录树
// @Description 获取目录树结构
// @Tags 安全目录管理
// @Accept json
// @Produce json
// @Param parentId query int false "父目录ID" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /directories/tree [get]
// @Security BearerAuth
func (h *DirectoryHandler) GetDirectoryTree(c *gin.Context) {
	var parentID *uint
	if parentIDStr := c.Query("parentId"); parentIDStr != "" {
		if pid, err := strconv.ParseUint(parentIDStr, 10, 32); err == nil {
			pidVal := uint(pid)
			parentID = &pidVal
		}
	}

	tree, err := h.groupService.GetDirectoryTree(c.Request.Context(), parentID)
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
			"message": "获取目录树失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    tree,
	})
}
