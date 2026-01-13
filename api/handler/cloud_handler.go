package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
)

// CloudHandler 云盘处理器
type CloudHandler struct {
	cloudService cloud.Service
}

// NewCloudHandler 创建云盘处理器
func NewCloudHandler(cloudService cloud.Service) *CloudHandler {
	return &CloudHandler{
		cloudService: cloudService,
	}
}

// CreateFolder 创建文件夹
// @Summary 创建文件夹
// @Description 在指定父文件夹下创建新文件夹
// @Tags Cloud
// @Accept json
// @Produce json
// @Param request body CreateFolderRequest true "创建文件夹请求"
// @Success 200 {object} entity.CloudFolder
// @Router /api/v1/cloud/folders [post]
func (h *CloudHandler) CreateFolder(c *gin.Context) {
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 将 parentId=0 视为根目录（nil）
	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	createReq := &cloud.CreateFolderRequest{
		Name:        req.FolderName,
		ParentID:    req.ParentID,
		Description: req.Description,
	}

	folder, err := h.cloudService.CreateFolder(c.Request.Context(), createReq, userID.(uint))
	if err != nil {
		utils.InternalError(c, "创建文件夹失败: "+err.Error())
		return
	}

	utils.Created(c, folder)
}

// ListFolders 列出文件夹
// @Summary 列出文件夹
// @Description 列出指定父文件夹下的所有子文件夹
// @Tags Cloud
// @Accept json
// @Produce json
// @Param parentId query int false "父文件夹ID（0或空表示根目录）"
// @Success 200 {array} entity.CloudFolder
// @Router /api/v1/cloud/folders [get]
func (h *CloudHandler) ListFolders(c *gin.Context) {
	parentIDStr := c.Query("parentId")
	var parentID *uint = nil
	if parentIDStr != "" {
		id, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			utils.BadRequest(c, "parentId 格式错误")
			return
		}
		// 将 0 视为根目录（nil）
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

	folders, err := h.cloudService.ListFolders(c.Request.Context(), parentID, userID.(uint))
	if err != nil {
		utils.InternalError(c, "查询文件夹失败: "+err.Error())
		return
	}

	utils.Success(c, folders)
}

// GetFolderTree 获取文件夹树
// @Summary 获取文件夹树
// @Description 获取用户的完整文件夹树结构
// @Tags Cloud
// @Accept json
// @Produce json
// @Success 200 {array} cloud.FolderNode
// @Router /api/v1/cloud/folders/tree [get]
func (h *CloudHandler) GetFolderTree(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	tree, err := h.cloudService.GetFolderTree(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.InternalError(c, "查询文件夹树失败: "+err.Error())
		return
	}

	utils.Success(c, tree)
}

// DeleteFolder 删除文件夹
// @Summary 删除文件夹
// @Description 删除指定文件夹（包括其中的所有文件和子文件夹）
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "文件夹ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/folders/{id} [delete]
func (h *CloudHandler) DeleteFolder(c *gin.Context) {
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

	if err := h.cloudService.DeleteFolder(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		utils.InternalError(c, "删除文件夹失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除文件夹成功"})
}

// RenameFolder 重命名文件夹
// @Summary 重命名文件夹
// @Description 重命名指定文件夹
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "文件夹ID"
// @Param request body RenameFolderRequest true "重命名请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/folders/{id}/rename [put]
func (h *CloudHandler) RenameFolder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req RenameFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.RenameFolder(c.Request.Context(), uint(id), req.NewName, userID.(uint)); err != nil {
		utils.InternalError(c, "重命名文件夹失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "重命名文件夹成功"})
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到指定文件夹
// @Tags Cloud
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param folderId formData int false "文件夹ID（0或空表示根目录）"
// @Success 200 {object} entity.CloudFile
// @Router /api/v1/cloud/files/upload [post]
func (h *CloudHandler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "未找到上传文件")
		return
	}

	folderIDStr := c.PostForm("folderId")
	var folderID *uint = nil
	if folderIDStr != "" {
		id, err := strconv.ParseUint(folderIDStr, 10, 32)
		if err != nil {
			utils.BadRequest(c, "folderId 格式错误")
			return
		}
		// 将 0 视为根目录（nil）
		if id > 0 {
			fid := uint(id)
			folderID = &fid
		}
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		utils.InternalError(c, "打开文件失败: "+err.Error())
		return
	}
	defer file.Close()

	// 构造上传请求
	uploadReq := &cloud.UploadFileRequest{
		FileName:    fileHeader.Filename,
		FolderID:    folderID,
		FileSize:    fileHeader.Size,
		FileType:    fileHeader.Header.Get("Content-Type"),
		Reader:      file,
		StorageType: "local",
	}

	uploadedFile, err := h.cloudService.UploadFile(c.Request.Context(), uploadReq, userID.(uint))
	if err != nil {
		utils.InternalError(c, "上传文件失败: "+err.Error())
		return
	}

	utils.Created(c, uploadedFile)
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定文件
// @Tags Cloud
// @Accept json
// @Produce application/octet-stream
// @Param id path int true "文件ID"
// @Success 200 {file} binary
// @Router /api/v1/cloud/files/{id}/download [get]
func (h *CloudHandler) DownloadFile(c *gin.Context) {
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

	reader, fileInfo, err := h.cloudService.DownloadFile(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		utils.InternalError(c, "下载文件失败: "+err.Error())
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.FileName)
	c.Header("Content-Type", fileInfo.FileType)
	c.Header("Content-Length", strconv.FormatInt(fileInfo.FileSize, 10))

	// 流式传输文件
	if _, err := io.Copy(c.Writer, reader); err != nil {
		utils.InternalError(c, "传输文件失败: "+err.Error())
		return
	}
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定文件
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/files/{id} [delete]
func (h *CloudHandler) DeleteFile(c *gin.Context) {
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

	if err := h.cloudService.DeleteFile(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		utils.InternalError(c, "删除文件失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除文件成功"})
}

// MoveFile 移动文件
// @Summary 移动文件
// @Description 移动文件到指定文件夹
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Param request body MoveFileRequest true "移动请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/files/{id}/move [put]
func (h *CloudHandler) MoveFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req MoveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 将 targetFolderId=0 视为根目录（nil）
	if req.TargetFolderID != nil && *req.TargetFolderID == 0 {
		req.TargetFolderID = nil
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.MoveFile(c.Request.Context(), uint(id), req.TargetFolderID, userID.(uint)); err != nil {
		utils.InternalError(c, "移动文件失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "移动文件成功"})
}

// RenameFile 重命名文件
// @Summary 重命名文件
// @Description 重命名指定文件
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Param request body RenameFileRequest true "重命名请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/files/{id}/rename [put]
func (h *CloudHandler) RenameFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req RenameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.cloudService.RenameFile(c.Request.Context(), uint(id), req.NewName, userID.(uint)); err != nil {
		utils.InternalError(c, "重命名文件失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "重命名文件成功"})
}

// ListFiles 列出文件
// @Summary 列出文件
// @Description 列出指定文件夹下的所有文件（支持分页）
// @Tags Cloud
// @Accept json
// @Produce json
// @Param folderId query int false "文件夹ID（0或空表示根目录）"
// @Param page query int false "页码（默认1）"
// @Param pageSize query int false "每页大小（默认20）"
// @Success 200 {object} ListFilesResponse
// @Router /api/v1/cloud/files [get]
func (h *CloudHandler) ListFiles(c *gin.Context) {
	folderIDStr := c.Query("folderId")
	var folderID *uint = nil
	if folderIDStr != "" {
		id, err := strconv.ParseUint(folderIDStr, 10, 32)
		if err != nil {
			utils.BadRequest(c, "folderId 格式错误")
			return
		}
		// 将 0 视为根目录（nil）
		if id > 0 {
			fid := uint(id)
			folderID = &fid
		}
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	files, total, err := h.cloudService.ListFiles(c.Request.Context(), folderID, userID.(uint), page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询文件失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"data":     files,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateShare 创建分享
// @Summary 创建文件分享
// @Description 创建文件或文件夹分享链接
// @Tags Cloud
// @Accept json
// @Produce json
// @Param request body CreateShareRequest true "创建分享请求"
// @Success 200 {object} entity.CloudShare
// @Router /api/v1/cloud/shares [post]
func (h *CloudHandler) CreateShare(c *gin.Context) {
	var req CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	shareReq := &cloud.CreateShareRequest{
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		ShareType:    req.ShareType,
		Password:     req.Password,
		ExpireDays:   req.ExpireDays,
		MaxDownloads: req.MaxDownloads,
	}

	share, err := h.cloudService.CreateShare(c.Request.Context(), shareReq, userID.(uint))
	if err != nil {
		utils.InternalError(c, "创建分享失败: "+err.Error())
		return
	}

	utils.Created(c, share)
}

// GetShareInfo 获取分享信息
// @Summary 获取分享信息
// @Description 根据分享码获取分享信息
// @Tags Cloud
// @Accept json
// @Produce json
// @Param code path string true "分享码"
// @Param password query string false "访问密码（如果需要）"
// @Success 200 {object} cloud.ShareInfo
// @Router /api/v1/cloud/shares/{code} [get]
func (h *CloudHandler) GetShareInfo(c *gin.Context) {
	code := c.Param("code")
	password := c.Query("password")

	shareInfo, err := h.cloudService.GetShareInfo(c.Request.Context(), code, password)
	if err != nil {
		utils.InternalError(c, "获取分享信息失败: "+err.Error())
		return
	}

	utils.Success(c, shareInfo)
}

// AccessShare 访问分享
// @Summary 访问分享文件
// @Description 访问分享文件（需要密码验证）
// @Tags Cloud
// @Accept json
// @Produce json
// @Param code path string true "分享码"
// @Param request body AccessShareRequest true "访问请求"
// @Success 200 {object} entity.CloudShare
// @Router /api/v1/cloud/shares/{code}/access [post]
func (h *CloudHandler) AccessShare(c *gin.Context) {
	code := c.Param("code")

	var req AccessShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	share, err := h.cloudService.AccessShare(c.Request.Context(), code, req.Password)
	if err != nil {
		utils.InternalError(c, "访问分享失败: "+err.Error())
		return
	}

	utils.Success(c, share)
}

// CancelShare 取消分享
// @Summary 取消文件分享
// @Description 取消指定的文件分享
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "分享ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/shares/{id} [delete]
func (h *CloudHandler) CancelShare(c *gin.Context) {
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

	if err := h.cloudService.CancelShare(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		utils.InternalError(c, "取消分享失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "取消分享成功"})
}

// GetUserQuota 获取用户配额
// @Summary 获取用户配额信息
// @Description 获取当前用户的存储配额和使用情况
// @Tags Cloud
// @Accept json
// @Produce json
// @Success 200 {object} entity.CloudQuota
// @Router /api/v1/cloud/quota [get]
func (h *CloudHandler) GetUserQuota(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	quota, err := h.cloudService.GetUserQuota(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.InternalError(c, "获取配额信息失败: "+err.Error())
		return
	}

	utils.Success(c, quota)
}

// 请求结构体定义

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	ParentID    *uint  `json:"parentId"`             // 父文件夹ID（nil表示根目录）
	FolderName  string `json:"folderName" binding:"required"` // 文件夹名称
	Description string `json:"description"`          // 描述
}

// RenameFolderRequest 重命名文件夹请求
type RenameFolderRequest struct {
	NewName string `json:"newName" binding:"required"` // 新名称
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	TargetFolderID *uint `json:"targetFolderId"` // 目标文件夹ID
}

// RenameFileRequest 重命名文件请求
type RenameFileRequest struct {
	NewName string `json:"newName" binding:"required"` // 新名称
}

// CreateShareRequest 创建分享请求
type CreateShareRequest struct {
	ResourceType string `json:"resourceType" binding:"required"` // file 或 folder
	ResourceID   uint   `json:"resourceId" binding:"required"`   // 资源ID
	ShareType    string `json:"shareType" binding:"required"`    // public, password, private
	Password     string `json:"password"`                        // 访问密码（当shareType=password时需要）
	ExpireDays   int    `json:"expireDays"`                      // 过期天数（0表示永久）
	MaxDownloads int    `json:"maxDownloads"`                    // 最大下载次数（0表示无限制）
}

// AccessShareRequest 访问分享请求
type AccessShareRequest struct {
	Password string `json:"password"` // 访问密码（如果需要）
}

// ListFilesResponse 文件列表响应
type ListFilesResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}
