package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
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
		utils.InternalError(c, "查询项目失败: "+err.Error())
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
func (h *CloudItemHandler) CreateItem(c *gin.Context) {
	var req CreateItemRequest
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
		utils.InternalError(c, "创建项目失败: "+err.Error())
		return
	}

	utils.Created(c, item)
}

// DeleteItem 删除项目
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
		utils.InternalError(c, "项目不存在")
		return
	}

	if item.IsFolder() {
		if err := h.cloudService.DeleteItemsByParentID(c.Request.Context(), item.ID, uid); err != nil {
			utils.InternalError(c, "删除子项目失败: "+err.Error())
			return
		}
	}

	if err := h.cloudService.DeleteItem(c.Request.Context(), uint(id), uid); err != nil {
		utils.InternalError(c, "删除项目失败: "+err.Error())
		return
	}

	if item.IsFile() && item.FileSize != nil {
		_ = h.cloudService.UpdateQuota(c.Request.Context(), uid, -*item.FileSize, -1, 0)
	} else if item.IsFolder() {
		_ = h.cloudService.UpdateQuota(c.Request.Context(), uid, 0, 0, -1)
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// RenameItem 重命名项目
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
		utils.InternalError(c, "重命名失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "重命名成功"})
}

// MoveItem 移动项目
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
		utils.InternalError(c, "移动失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "移动成功"})
}

// BatchDelete 批量删除
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
			if item.IsFile() && item.FileSize != nil {
				_ = h.cloudService.UpdateQuota(c.Request.Context(), uid, -*item.FileSize, -1, 0)
			} else if item.IsFolder() {
				_ = h.cloudService.UpdateQuota(c.Request.Context(), uid, 0, 0, -1)
			}
		}
	}

	utils.Success(c, BatchDeleteResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	})
}

// BatchMove 批量移动
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
	TargetParentID *uint  `json:"targetParentId"`
}

type BatchDeleteResponse struct {
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedItems  []string `json:"failedItems"`
}
