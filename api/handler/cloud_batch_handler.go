package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
)

// CloudBatchDeleteRequest 云盘批量删除请求
type CloudBatchDeleteRequest struct {
	FileIDs   []uint `json:"fileIds"`
	FolderIDs []uint `json:"folderIds"`
}

// CloudBatchDeleteResponse 云盘批量删除响应
type CloudBatchDeleteResponse struct {
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedItems  []string `json:"failedItems,omitempty"`
}

// BatchDelete 批量删除
func (h *CloudHandler) BatchDelete(c *gin.Context) {
	var req CloudBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, errors.ErrInvalidParam, errors.New(errors.ErrInvalidParam, "参数错误"))
		return
	}

	userID := c.GetUint("userID")
	ctx := c.Request.Context()

	successCount := 0
	failedItems := []string{}

	// 批量删除文件
	for _, fileID := range req.FileIDs {
		if err := h.cloudService.DeleteFile(ctx, fileID, userID); err != nil {
			failedItems = append(failedItems, fmt.Sprintf("file_%d", fileID))
		} else {
			successCount++
		}
	}

	// 批量删除文件夹
	for _, folderID := range req.FolderIDs {
		if err := h.cloudService.DeleteFolder(ctx, folderID, userID); err != nil {
			failedItems = append(failedItems, fmt.Sprintf("folder_%d", folderID))
		} else {
			successCount++
		}
	}

	utils.Success(c, CloudBatchDeleteResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	})
}

// CloudBatchMoveRequest 云盘批量移动请求
type CloudBatchMoveRequest struct {
	FileIDs        []uint `json:"fileIds"`
	TargetFolderID *uint  `json:"targetFolderId"`
}

// CloudBatchMoveResponse 云盘批量移动响应
type CloudBatchMoveResponse struct {
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedItems  []string `json:"failedItems,omitempty"`
}

// BatchMove 批量移动
func (h *CloudHandler) BatchMove(c *gin.Context) {
	var req CloudBatchMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, errors.ErrInvalidParam, errors.New(errors.ErrInvalidParam, "参数错误"))
		return
	}

	userID := c.GetUint("userID")
	ctx := c.Request.Context()

	successCount := 0
	failedItems := []string{}

	for _, fileID := range req.FileIDs {
		if err := h.cloudService.MoveFile(ctx, fileID, req.TargetFolderID, userID); err != nil {
			failedItems = append(failedItems, fmt.Sprintf("file_%d", fileID))
		} else {
			successCount++
		}
	}

	utils.Success(c, CloudBatchMoveResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	})
}
