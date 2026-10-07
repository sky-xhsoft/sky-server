package handler

import (
	"fmt"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
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
		utils.FailFromError(c, "创建文件夹失败: ", err)
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
		utils.FailFromError(c, "查询文件夹失败: ", err)
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
		utils.FailFromError(c, "查询文件夹树失败: ", err)
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
		utils.FailFromError(c, "删除文件夹失败: ", err)
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
		utils.FailFromError(c, "重命名文件夹失败: ", err)
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
		utils.FailFromError(c, "打开文件失败: ", err)
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
		utils.FailFromError(c, "上传文件失败: ", err)
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
		utils.FailFromError(c, "下载文件失败: ", err)
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.Name)
	c.Header("Content-Type", *fileInfo.FileType)
	c.Header("Content-Length", strconv.FormatInt(*fileInfo.FileSize, 10))

	// 流式传输文件
	if _, err := io.Copy(c.Writer, reader); err != nil {
		utils.FailFromError(c, "传输文件失败: ", err)
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
		utils.FailFromError(c, "删除文件失败: ", err)
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
		utils.FailFromError(c, "移动文件失败: ", err)
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
		utils.FailFromError(c, "重命名文件失败: ", err)
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
		utils.FailFromError(c, "查询文件失败: ", err)
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
		utils.FailFromError(c, "创建分享失败: ", err)
		return
	}

	utils.Created(c, share)
}

// ShareListItem 分享列表项（扩展了资源名称和链接）
type ShareListItem struct {
	*entity.CloudShare
	FileName  string `json:"fileName"`  // 文件/文件夹名称
	ShareLink string `json:"shareLink"` // 完整分享链接
}

// GetMyShares 获取我的分享列表
// @Summary 获取我的分享列表
// @Description 获取当前用户创建的所有分享
// @Tags Cloud
// @Accept json
// @Produce json
// @Success 200 {array} ShareListItem
// @Router /api/v1/cloud/shares [get]
func (h *CloudHandler) GetMyShares(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	shares, err := h.cloudService.GetUserShares(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.FailFromError(c, "获取分享列表失败: ", err)
		return
	}

	// 构建分享链接的基础URL
	baseURL := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	// 转换为扩展的分享列表项
	shareList := make([]ShareListItem, 0, len(shares))
	for _, share := range shares {
		item := ShareListItem{
			CloudShare: share,
			ShareLink:  fmt.Sprintf("%s://%s/share/%s", scheme, baseURL, share.ShareCode),
		}

		// 根据资源类型获取资源名称
		if share.ResourceType == "file" {
			// 查询文件名
			fileItem, err := h.cloudService.GetItem(c.Request.Context(), share.ResourceID, userID.(uint))
			if err == nil && fileItem != nil {
				item.FileName = fileItem.Name
			} else {
				item.FileName = "未知文件"
			}
		} else if share.ResourceType == "folder" {
			// 查询文件夹名
			folderItem, err := h.cloudService.GetItem(c.Request.Context(), share.ResourceID, userID.(uint))
			if err == nil && folderItem != nil {
				item.FileName = folderItem.Name
			} else {
				item.FileName = "未知文件夹"
			}
		}

		shareList = append(shareList, item)
	}

	utils.Success(c, shareList)
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

	if !h.checkShareViewer(c, code) {
		return
	}

	shareInfo, err := h.cloudService.GetShareInfo(c.Request.Context(), code, password)
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
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

	if !h.checkShareViewer(c, code) {
		return
	}

	share, err := h.cloudService.AccessShare(c.Request.Context(), code, req.Password)
	if err != nil {
		writeShareError(c, "访问分享失败: ", err)
		return
	}

	utils.Success(c, share)
}

// writeShareError 分享公开接口的错误响应：访问密码错误 400，缺少/无效访问凭证 403，分享不存在/失效/过期 404，其余 500
func writeShareError(c *gin.Context, prefix string, err error) {
	var appErr *errors.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case errors.ErrInvalidParam:
			utils.BadRequest(c, appErr.Message)
			return
		case errors.ErrResourceNotFound:
			utils.NotFound(c, appErr.Message)
			return
		case errors.ErrUnauthorized:
			// 用 403 而非 401：前端遇到 401 会按登录过期处理并跳转登录页，而分享访问者通常未登录
			utils.Forbidden(c, appErr.Message)
			return
		}
	}
	utils.InternalError(c, prefix+err.Error())
}

// shareViewerCompanyID 访问者所在公司（可选鉴权解析出的登录用户），未登录为 0
func shareViewerCompanyID(c *gin.Context) uint {
	// 必须是已登录用户：域名租户中间件会给匿名请求也设置 companyID（按域名识别或默认公司），
	// 不能据此认定访问者属于该公司
	if _, loggedIn := c.Get("userID"); !loggedIn {
		return 0
	}
	if v, ok := c.Get("companyID"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// checkShareViewer 私密分享先校验访问者，失败时已写入错误响应
func (h *CloudHandler) checkShareViewer(c *gin.Context, code string) bool {
	share, err := h.cloudService.GetShareByCode(c.Request.Context(), code)
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
		return false
	}
	if err := h.cloudService.VerifyShareViewer(c.Request.Context(), share, shareViewerCompanyID(c)); err != nil {
		writeShareError(c, "", err)
		return false
	}
	return true
}

// shareAccessToken 读取加密分享访问凭证：优先请求头，其次查询参数（<img>/<video> 等无法设置请求头）
func shareAccessToken(c *gin.Context) string {
	if token := c.GetHeader("X-Share-Token"); token != "" {
		return token
	}
	return c.Query("token")
}

// DownloadShareFile 下载分享的文件
// @Summary 下载分享的文件
// @Description 通过分享码下载文件（不需要认证）
// @Tags Cloud
// @Produce octet-stream
// @Param code path string true "分享码"
// @Success 200 {file} binary
// @Router /api/v1/cloud/shares/{code}/download [get]
func (h *CloudHandler) DownloadShareFile(c *gin.Context) {
	code := c.Param("code")

	// 获取分享记录（不验证密码）
	share, err := h.cloudService.GetShareByCode(c.Request.Context(), code)
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
		return
	}
	if err := h.cloudService.VerifyShareViewer(c.Request.Context(), share, shareViewerCompanyID(c)); err != nil {
		writeShareError(c, "", err)
		return
	}
	// 加密/私密分享须携带签发的访问凭证
	if err := h.cloudService.VerifyShareAccess(share, shareAccessToken(c)); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 只支持文件分享下载
	if share.ResourceType != "file" {
		utils.BadRequest(c, "只能下载文件分享")
		return
	}

	// 计入下载次数（与上限检查原子完成）
	if err := h.cloudService.ClaimShareDownload(c.Request.Context(), share.ID); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 获取文件ID
	fileID := share.ResourceID

	// 下载文件（不需要用户ID验证，因为是公开分享）
	reader, fileInfo, err := h.cloudService.DownloadFileByID(c.Request.Context(), fileID)
	if err != nil {
		utils.FailFromError(c, "下载文件失败: ", err)
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.Name)
	c.Header("Content-Type", *fileInfo.FileType)
	c.Header("Content-Length", strconv.FormatInt(*fileInfo.FileSize, 10))

	// 流式传输文件
	if _, err := io.Copy(c.Writer, reader); err != nil {
		utils.FailFromError(c, "传输文件失败: ", err)
		return
	}
}

// GetShareDownloadURL 点击下载时获取下载地址：校验访问权限并计入下载次数（达到上限则拒绝）后返回文件直链；
// 文件没有直链时返回空 url，前端改用流式下载接口
// @Summary 获取分享文件下载地址
// @Tags Cloud
// @Produce json
// @Param code path string true "分享码"
// @Param fileId query int false "文件夹分享中的文件ID"
// @Router /api/v1/cloud/shares/{code}/download-url [post]
func (h *CloudHandler) GetShareDownloadURL(c *gin.Context) {
	ctx := c.Request.Context()
	share, err := h.cloudService.GetShareByCode(ctx, c.Param("code"))
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
		return
	}
	if err := h.cloudService.VerifyShareViewer(ctx, share, shareViewerCompanyID(c)); err != nil {
		writeShareError(c, "", err)
		return
	}
	if err := h.cloudService.VerifyShareAccess(share, shareAccessToken(c)); err != nil {
		writeShareError(c, "", err)
		return
	}

	fileID := share.ResourceID
	if share.ResourceType == "folder" {
		id, err := strconv.ParseUint(c.Query("fileId"), 10, 32)
		if err != nil || id == 0 {
			utils.BadRequest(c, "文件ID格式错误")
			return
		}
		if err := h.cloudService.VerifyShareItem(ctx, share, uint(id)); err != nil {
			writeShareError(c, "", err)
			return
		}
		fileID = uint(id)
	}

	info, err := h.cloudService.PrepareShareDownload(ctx, share, fileID)
	if err != nil {
		writeShareError(c, "获取下载地址失败: ", err)
		return
	}
	utils.Success(c, info)
}

// DownloadShareFolderFile 下载分享文件夹中的文件
// @Summary 下载分享文件夹中的文件
// @Description 通过分享码下载文件夹中的单个文件（不需要认证）
// @Tags Cloud
// @Produce octet-stream
// @Param code path string true "分享码"
// @Param fileId path int true "文件ID"
// @Success 200 {file} binary
// @Router /api/v1/cloud/shares/{code}/files/{fileId}/download [get]
func (h *CloudHandler) DownloadShareFolderFile(c *gin.Context) {
	code := c.Param("code")
	fileIDStr := c.Param("fileId")

	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "文件ID格式错误")
		return
	}

	// 获取分享记录（不验证密码）
	share, err := h.cloudService.GetShareByCode(c.Request.Context(), code)
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
		return
	}
	if err := h.cloudService.VerifyShareViewer(c.Request.Context(), share, shareViewerCompanyID(c)); err != nil {
		writeShareError(c, "", err)
		return
	}
	// 加密/私密分享须携带签发的访问凭证
	if err := h.cloudService.VerifyShareAccess(share, shareAccessToken(c)); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 只支持文件夹分享
	if share.ResourceType != "folder" {
		utils.BadRequest(c, "只能下载文件夹分享中的文件")
		return
	}
	// 文件须位于被分享的文件夹内
	if err := h.cloudService.VerifyShareItem(c.Request.Context(), share, uint(fileID)); err != nil {
		writeShareError(c, "", err)
		return
	}
	// 文件夹内每下载一个文件计一次分享下载
	if err := h.cloudService.ClaimShareDownload(c.Request.Context(), share.ID); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 下载文件
	reader, fileInfo, err := h.cloudService.DownloadFileByID(c.Request.Context(), uint(fileID))
	if err != nil {
		utils.FailFromError(c, "下载文件失败: ", err)
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.Name)
	if fileInfo.FileType != nil {
		c.Header("Content-Type", *fileInfo.FileType)
	} else {
		c.Header("Content-Type", "application/octet-stream")
	}
	if fileInfo.FileSize != nil {
		c.Header("Content-Length", strconv.FormatInt(*fileInfo.FileSize, 10))
	}

	// 流式传输文件
	if _, err := io.Copy(c.Writer, reader); err != nil {
		utils.FailFromError(c, "传输文件失败: ", err)
		return
	}
}

// PreviewShareFolderFile 预览分享文件夹中的文件
// @Summary 预览分享文件夹中的文件
// @Description 通过分享码预览文件夹中的单个文件（不需要认证）
// @Tags Cloud
// @Produce octet-stream
// @Param code path string true "分享码"
// @Param fileId path int true "文件ID"
// @Success 200 {file} binary
// @Router /api/v1/cloud/shares/{code}/files/{fileId}/preview [get]
func (h *CloudHandler) PreviewShareFolderFile(c *gin.Context) {
	code := c.Param("code")
	fileIDStr := c.Param("fileId")

	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "文件ID格式错误")
		return
	}

	// 获取分享记录（不验证密码）
	share, err := h.cloudService.GetShareByCode(c.Request.Context(), code)
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
		return
	}
	if err := h.cloudService.VerifyShareViewer(c.Request.Context(), share, shareViewerCompanyID(c)); err != nil {
		writeShareError(c, "", err)
		return
	}
	// 加密/私密分享须携带签发的访问凭证
	if err := h.cloudService.VerifyShareAccess(share, shareAccessToken(c)); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 只支持文件夹分享
	if share.ResourceType != "folder" {
		utils.BadRequest(c, "只能预览文件夹分享中的文件")
		return
	}
	// 文件须位于被分享的文件夹内
	if err := h.cloudService.VerifyShareItem(c.Request.Context(), share, uint(fileID)); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 下载文件（用于预览）
	reader, fileInfo, err := h.cloudService.DownloadFileByID(c.Request.Context(), uint(fileID))
	if err != nil {
		utils.FailFromError(c, "获取文件失败: ", err)
		return
	}
	defer reader.Close()

	// 根据文件类型设置 Content-Type
	contentType := "application/octet-stream"
	if fileInfo.FileType != nil {
		contentType = *fileInfo.FileType
	}

	// 设置响应头（inline 表示在浏览器中预览）
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename="+fileInfo.Name)
	if fileInfo.FileSize != nil {
		c.Header("Content-Length", strconv.FormatInt(*fileInfo.FileSize, 10))
	}

	// 流式传输文件
	if _, err := io.Copy(c.Writer, reader); err != nil {
		utils.FailFromError(c, "传输文件失败: ", err)
		return
	}
}

// GetShareFolderContent 获取分享文件夹的内容
// @Summary 获取分享文件夹的内容
// @Description 通过分享码获取文件夹内容（不需要认证，但需要先验证密码）
// @Tags Cloud
// @Accept json
// @Produce json
// @Param code path string true "分享码"
// @Param parentId query int false "父文件夹ID（0或空表示根目录）"
// @Success 200 {object} FolderContentResponse
// @Router /api/v1/cloud/shares/{code}/content [get]
func (h *CloudHandler) GetShareFolderContent(c *gin.Context) {
	code := c.Param("code")

	// 获取分享记录（不验证密码）
	share, err := h.cloudService.GetShareByCode(c.Request.Context(), code)
	if err != nil {
		writeShareError(c, "获取分享信息失败: ", err)
		return
	}
	if err := h.cloudService.VerifyShareViewer(c.Request.Context(), share, shareViewerCompanyID(c)); err != nil {
		writeShareError(c, "", err)
		return
	}
	// 加密/私密分享须携带签发的访问凭证
	if err := h.cloudService.VerifyShareAccess(share, shareAccessToken(c)); err != nil {
		writeShareError(c, "", err)
		return
	}

	// 只支持文件夹分享
	if share.ResourceType != "folder" {
		utils.BadRequest(c, "只能查看文件夹分享内容")
		return
	}

	// 获取 parentId 参数
	parentIDStr := c.Query("parentId")
	var parentID *uint
	if parentIDStr != "" && parentIDStr != "0" {
		id, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			utils.BadRequest(c, "parentId 格式错误")
			return
		}
		pid := uint(id)
		// 只能浏览被分享的文件夹及其子文件夹
		if err := h.cloudService.VerifyShareItem(c.Request.Context(), share, pid); err != nil {
			writeShareError(c, "", err)
			return
		}
		parentID = &pid
	} else {
		// 如果没有指定 parentId，使用分享的根文件夹ID
		pid := share.ResourceID
		parentID = &pid
	}

	// 获取文件夹内容（使用分享者的 userID）
	folders, err := h.cloudService.ListFolders(c.Request.Context(), parentID, share.SharerID)
	if err != nil {
		utils.FailFromError(c, "查询文件夹失败: ", err)
		return
	}

	files, _, err := h.cloudService.ListFiles(c.Request.Context(), parentID, share.SharerID, 1, 10000)
	if err != nil {
		utils.FailFromError(c, "查询文件失败: ", err)
		return
	}

	utils.Success(c, gin.H{
		"folders": folders,
		"files":   files,
	})
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
		utils.FailFromError(c, "取消分享失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "取消分享成功"})
}

// UpdateShare 更新分享
// @Summary 更新分享信息
// @Description 更新分享的访问密码和有效期
// @Tags Cloud
// @Accept json
// @Produce json
// @Param id path int true "分享ID"
// @Param request body UpdateShareRequest true "更新分享请求"
// @Success 200 {object} entity.CloudShare
// @Router /api/v1/cloud/shares/{id} [put]
func (h *CloudHandler) UpdateShare(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req UpdateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 构建更新请求
	updateReq := &cloud.UpdateShareRequest{
		Password:   req.Password,
		ExpireDays: req.ExpireDays,
	}

	share, err := h.cloudService.UpdateShare(c.Request.Context(), uint(id), updateReq, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "更新分享失败: ", err)
		return
	}

	utils.Success(c, share)
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
		utils.FailFromError(c, "获取配额信息失败: ", err)
		return
	}

	utils.Success(c, quota)
}

// 请求结构体定义

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	ParentID    *uint  `json:"parentId"`                      // 父文件夹ID（nil表示根目录）
	FolderName  string `json:"folderName" binding:"required"` // 文件夹名称
	Description string `json:"description"`                   // 描述
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

// UpdateShareRequest 更新分享请求
type UpdateShareRequest struct {
	Password   *string `json:"password"`   // 访问密码（nil表示不修改）
	ExpireDays *int    `json:"expireDays"` // 过期天数（0表示永久，nil表示不修改）
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

// GetFolderContent 获取文件夹内容（子文件夹+文件）
// @Summary 获取文件夹内容
// @Description 获取指定文件夹下的子文件夹和文件列表
// @Tags Cloud
// @Accept json
// @Produce json
// @Param folderId query int false "文件夹ID（0或空表示根目录）"
// @Success 200 {object} FolderContentResponse
// @Router /api/v1/cloud/folders/content [get]
func (h *CloudHandler) GetFolderContent(c *gin.Context) {
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

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	uid := userID.(uint)

	// 并发获取文件夹和文件
	type foldersResult struct {
		folders []*entity.CloudFolder
		err     error
	}
	type filesResult struct {
		files []*entity.CloudItem
		total int64
		err   error
	}

	foldersChan := make(chan foldersResult, 1)
	filesChan := make(chan filesResult, 1)

	// 获取子文件夹
	go func() {
		folders, err := h.cloudService.ListFolders(c.Request.Context(), folderID, uid)
		foldersChan <- foldersResult{folders: folders, err: err}
	}()

	// 获取文件列表（不分页，获取全部）
	go func() {
		files, total, err := h.cloudService.ListFiles(c.Request.Context(), folderID, uid, 1, 10000)
		filesChan <- filesResult{files: files, total: total, err: err}
	}()

	// 等待结果
	foldersRes := <-foldersChan
	filesRes := <-filesChan

	// 检查错误
	if foldersRes.err != nil {
		utils.InternalError(c, "查询文件夹失败: "+foldersRes.err.Error())
		return
	}
	if filesRes.err != nil {
		utils.InternalError(c, "查询文件失败: "+filesRes.err.Error())
		return
	}

	utils.Success(c, gin.H{
		"folders": foldersRes.folders,
		"files":   filesRes.files,
	})
}

// FolderContentResponse 文件夹内容响应
type FolderContentResponse struct {
	Folders []*entity.CloudFolder `json:"folders"` // 子文件夹列表
	Files   []*entity.CloudItem   `json:"files"`   // 文件列表
}
