package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"go.uber.org/zap"
)

// CloudItemHandler 云盘项目处理器（统一处理文件和文件夹）
type CloudItemHandler struct {
	cloudService cloud.Service
}

// NewCloudItemHandler 创建云盘项目处理器
func NewCloudItemHandler(cloudService cloud.Service) *CloudItemHandler {
	return &CloudItemHandler{
		cloudService: cloudService,
	}
}

// ListItems 列出项目（文件+文件夹）
// @Summary 列出云盘项目
// @Tags cloud
// @Produce json
// @Param folderId query int false "文件夹ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items [get]
func (h *CloudItemHandler) ListItems(c *gin.Context) {
	parentIDStr := c.Query("parentId")
	var parentID *uint = nil
	if parentIDStr != "" {
		id, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			utils.BadRequest(c, "parentId 格式错误")
			return
		}
		if id > 0 {
			pid := uint(id)
			parentID = &pid
		}
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	items, err := h.cloudService.ListItems(c.Request.Context(), parentID, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "查询项目失败: ", err)
		return
	}

	folders := []interface{}{}
	files := []interface{}{}

	for _, item := range items {
		if item.IsFolder() {
			folders = append(folders, item)
		} else if item.IsFile() {
			files = append(files, item)
		}
	}

	utils.Success(c, gin.H{
		"folders": folders,
		"files":   files,
	})
}

// CreateItem 创建项目
// @Summary 创建云盘项目（文件或文件夹）
// @Tags cloud
// @Accept json
// @Produce json
// @Param request body object true "创建请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items [post]
func (h *CloudItemHandler) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	name, err := cloud.ValidateItemName(req.Name)
	if err != nil {
		utils.FailFromError(c, "", err)
		return
	}
	req.Name = name

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	uid := userID.(uint)

	item := &entity.CloudItem{
		ItemType: req.ItemType,
		Name:     req.Name,
		ParentID: req.ParentID,
		OwnerID:  uid,
	}

	if req.ParentID == nil || *req.ParentID == 0 {
		item.Path = "/" + req.Name
		item.ParentID = nil
	} else {
		parent, err := h.cloudService.GetItem(c.Request.Context(), *req.ParentID, uid)
		if err != nil {
			utils.InternalError(c, "父文件夹不存在")
			return
		}
		item.Path = parent.Path + "/" + req.Name
	}

	if err := h.cloudService.CreateItem(c.Request.Context(), item); err != nil {
		utils.FailFromError(c, "创建项目失败: ", err)
		return
	}

	// 重新计算配额
	if err := h.cloudService.RecalculateQuota(c.Request.Context(), uid); err != nil {
		logger.Warn("RecalculateQuota failed", zap.Uint("userId", uid), zap.Error(err))
	}

	utils.Created(c, item)
}

// DeleteItem 删除项目
// @Summary 删除云盘项目
// @Tags cloud
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/{id} [delete]
func (h *CloudItemHandler) DeleteItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	uid := userID.(uint)

	item, err := h.cloudService.GetItem(c.Request.Context(), uint(id), uid)
	if err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	if item.IsFolder() {
		if err := h.cloudService.DeleteItemsByParentID(c.Request.Context(), item.ID, uid); err != nil {
			utils.FailFromError(c, "删除子项目失败: ", err)
			return
		}
	}

	if err := h.cloudService.DeleteItem(c.Request.Context(), uint(id), uid); err != nil {
		utils.FailFromError(c, "删除项目失败: ", err)
		return
	}

	// 重新计算配额
	if err := h.cloudService.RecalculateQuota(c.Request.Context(), uid); err != nil {
		logger.Warn("RecalculateQuota failed", zap.Uint("userId", uid), zap.Error(err))
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// RenameItem 重命名项目
// @Summary 重命名云盘项目
// @Tags cloud
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Param request body object true "新名称"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/{id}/rename [put]
func (h *CloudItemHandler) RenameItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req RenameItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.RenameItem(c.Request.Context(), uint(id), req.NewName, userID.(uint)); err != nil {
		utils.FailFromError(c, "重命名失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "重命名成功"})
}

// MoveItem 移动项目
// @Summary 移动云盘项目到目标文件夹
// @Tags cloud
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Param request body object true "目标文件夹ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/{id}/move [put]
func (h *CloudItemHandler) MoveItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req MoveItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.MoveItem(c.Request.Context(), uint(id), req.TargetParentID, userID.(uint)); err != nil {
		utils.FailFromError(c, "移动失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "移动成功"})
}

// BatchDelete 批量删除
// @Summary 批量删除云盘项目
// @Tags cloud
// @Accept json
// @Produce json
// @Param request body object true "项目ID列表"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/batch/delete [post]
func (h *CloudItemHandler) BatchDelete(c *gin.Context) {
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	uid := userID.(uint)
	successCount := 0
	failedItems := []string{}

	allIDs := append(req.ItemIDs, req.FileIDs...)
	allIDs = append(allIDs, req.FolderIDs...)

	for _, itemID := range allIDs {
		item, err := h.cloudService.GetItem(c.Request.Context(), itemID, uid)
		if err != nil {
			failedItems = append(failedItems, strconv.FormatUint(uint64(itemID), 10))
			continue
		}

		if item.IsFolder() {
			_ = h.cloudService.DeleteItemsByParentID(c.Request.Context(), item.ID, uid)
		}

		if err := h.cloudService.DeleteItem(c.Request.Context(), itemID, uid); err != nil {
			failedItems = append(failedItems, strconv.FormatUint(uint64(itemID), 10))
		} else {
			successCount++
		}
	}

	// 批量删除完成后，重新计算配额
	if successCount > 0 {
		if err := h.cloudService.RecalculateQuota(c.Request.Context(), uid); err != nil {
			logger.Warn("RecalculateQuota failed", zap.Uint("userId", uid), zap.Error(err))
		}
	}

	utils.Success(c, BatchDeleteResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	})
}

// BatchMove 批量移动
// @Summary 批量移动云盘项目
// @Tags cloud
// @Accept json
// @Produce json
// @Param request body object true "项目ID列表和目标文件夹"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/batch/move [post]
func (h *CloudItemHandler) BatchMove(c *gin.Context) {
	var req BatchMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	uid := userID.(uint)
	successCount := 0
	failedItems := []string{}

	allIDs := append(req.ItemIDs, req.FileIDs...)
	allIDs = append(allIDs, req.FolderIDs...)

	for _, itemID := range allIDs {
		if err := h.cloudService.MoveItem(c.Request.Context(), itemID, req.TargetParentID, uid); err != nil {
			failedItems = append(failedItems, strconv.FormatUint(uint64(itemID), 10))
		} else {
			successCount++
		}
	}

	utils.Success(c, BatchDeleteResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	})
}

// 请求结构体定义

type CreateItemRequest struct {
	ItemType string `json:"itemType" binding:"required"`
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parentId"`
}

type RenameItemRequest struct {
	NewName string `json:"newName" binding:"required"`
}

type MoveItemRequest struct {
	TargetParentID *uint `json:"targetParentId"`
}

type BatchDeleteRequest struct {
	ItemIDs   []uint `json:"itemIds"`
	FileIDs   []uint `json:"fileIds"`
	FolderIDs []uint `json:"folderIds"`
}

type BatchMoveRequest struct {
	ItemIDs        []uint `json:"itemIds"`
	FileIDs        []uint `json:"fileIds"`
	FolderIDs      []uint `json:"folderIds"`
	TargetParentID *uint  `json:"targetParentId"`
}

type BatchDeleteResponse struct {
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedItems  []string `json:"failedItems"`
}

// SetItemPermission 设置项目权限
// @Summary 设置云盘项目权限
// @Tags cloud
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Param request body cloud.SetItemPermissionRequest true "权限设置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/{id}/permissions [post]
func (h *CloudItemHandler) SetItemPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req cloud.SetItemPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.SetItemPermission(c.Request.Context(), uint(id), &req, userID.(uint)); err != nil {
		utils.FailFromError(c, "设置权限失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "权限设置成功"})
}

// ListItemPermissions 列出项目权限
// @Summary 列出云盘项目权限
// @Tags cloud
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/{id}/permissions [get]
func (h *CloudItemHandler) ListItemPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	perms, err := h.cloudService.ListItemPermissions(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		utils.FailFromError(c, "查询权限失败: ", err)
		return
	}

	utils.Success(c, gin.H{"permissions": perms})
}

// RemoveItemPermission 删除项目权限
// @Summary 删除云盘项目权限
// @Tags cloud
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Param permissionId path int true "权限ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cloud/items/{id}/permissions/{permissionId} [delete]
func (h *CloudItemHandler) RemoveItemPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}
	permStr := c.Param("permissionId")
	permID, err := strconv.ParseUint(permStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "权限ID格式错误")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.RemoveItemPermission(c.Request.Context(), uint(id), uint(permID), userID.(uint)); err != nil {
		utils.FailFromError(c, "删除权限失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "权限删除成功"})
}
