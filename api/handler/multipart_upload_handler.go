package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
)

// MultipartUploadHandler 分片上传处理器
type MultipartUploadHandler struct {
	multipartService cloud.MultipartUploadService
}

// NewMultipartUploadHandler 创建分片上传处理器
func NewMultipartUploadHandler(multipartService cloud.MultipartUploadService) *MultipartUploadHandler {
	return &MultipartUploadHandler{
		multipartService: multipartService,
	}
}

// InitUpload 初始化分片上传
// @Summary 初始化分片上传
// @Description 初始化分片上传会话，支持断点续传（如果存在未完成的会话会返回已上传的分片）
// @Tags Cloud-Multipart
// @Accept json
// @Produce json
// @Param request body InitUploadRequest true "初始化上传请求"
// @Success 200 {object} cloud.UploadSessionInfo
// @Router /api/v1/cloud/files/multipart/init [post]
func (h *MultipartUploadHandler) InitUpload(c *gin.Context) {
	var req InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 将 folderId=0 视为根目录（nil）
	if req.FolderID != nil && *req.FolderID == 0 {
		req.FolderID = nil
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	initReq := &cloud.InitUploadRequest{
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileMD5:     req.FileMD5,
		FileType:    req.FileType,
		ChunkSize:   req.ChunkSize,
		FolderID:    req.FolderID,
		StorageType: req.StorageType,
	}

	sessionInfo, err := h.multipartService.InitUpload(c.Request.Context(), initReq, userID.(uint))
	if err != nil {
		utils.InternalError(c, "初始化上传失败: "+err.Error())
		return
	}

	utils.Success(c, sessionInfo)
}

// UploadChunk 上传单个分片
// @Summary 上传单个分片
// @Description 上传文件的单个分片，支持断点续传（已上传的分片会自动跳过）
// @Tags Cloud-Multipart
// @Accept multipart/form-data
// @Produce json
// @Param sessionId formData int true "会话ID"
// @Param chunkIndex formData int true "分片索引"
// @Param chunkMd5 formData string true "分片MD5"
// @Param chunkData formData file true "分片数据"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/files/multipart/upload [post]
func (h *MultipartUploadHandler) UploadChunk(c *gin.Context) {
	// 获取表单参数
	sessionIDStr := c.PostForm("sessionId")
	chunkIndexStr := c.PostForm("chunkIndex")
	chunkMD5 := c.PostForm("chunkMd5")

	// 验证参数
	if sessionIDStr == "" || chunkIndexStr == "" || chunkMD5 == "" {
		utils.BadRequest(c, "缺少必要参数: sessionId, chunkIndex, chunkMd5")
		return
	}

	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "sessionId 格式错误")
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		utils.BadRequest(c, "chunkIndex 格式错误")
		return
	}

	// 获取分片文件
	fileHeader, err := c.FormFile("chunkData")
	if err != nil {
		utils.BadRequest(c, "未找到分片数据")
		return
	}

	// 打开分片文件
	file, err := fileHeader.Open()
	if err != nil {
		utils.InternalError(c, "打开分片文件失败: "+err.Error())
		return
	}
	defer file.Close()

	// 读取分片数据
	chunkData, err := io.ReadAll(file)
	if err != nil {
		utils.InternalError(c, "读取分片数据失败: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 上传分片
	uploadReq := &cloud.UploadChunkRequest{
		SessionID:  uint(sessionID),
		ChunkIndex: chunkIndex,
		ChunkData:  chunkData,
		ChunkMD5:   chunkMD5,
	}

	if err := h.multipartService.UploadChunk(c.Request.Context(), uploadReq, userID.(uint)); err != nil {
		utils.InternalError(c, "上传分片失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":    "分片上传成功",
		"chunkIndex": chunkIndex,
		"uploaded":   true,
	})
}

// GetUploadStatus 获取上传状态
// @Summary 获取上传状态
// @Description 获取分片上传的当前状态，包括已上传的分片列表和进度
// @Tags Cloud-Multipart
// @Accept json
// @Produce json
// @Param sessionId query int true "会话ID"
// @Success 200 {object} cloud.UploadStatus
// @Router /api/v1/cloud/files/multipart/status [get]
func (h *MultipartUploadHandler) GetUploadStatus(c *gin.Context) {
	sessionIDStr := c.Query("sessionId")
	if sessionIDStr == "" {
		utils.BadRequest(c, "缺少参数: sessionId")
		return
	}

	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "sessionId 格式错误")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	status, err := h.multipartService.GetUploadStatus(c.Request.Context(), uint(sessionID), userID.(uint))
	if err != nil {
		utils.InternalError(c, "获取上传状态失败: "+err.Error())
		return
	}

	utils.Success(c, status)
}

// CompleteUpload 完成上传
// @Summary 完成分片上传
// @Description 完成所有分片上传后，合并分片并创建文件记录
// @Tags Cloud-Multipart
// @Accept json
// @Produce json
// @Param request body CompleteUploadRequest true "完成上传请求"
// @Success 200 {object} entity.CloudFile
// @Router /api/v1/cloud/files/multipart/complete [post]
func (h *MultipartUploadHandler) CompleteUpload(c *gin.Context) {
	var req CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	file, err := h.multipartService.CompleteUpload(c.Request.Context(), req.SessionID, userID.(uint))
	if err != nil {
		utils.InternalError(c, "完成上传失败: "+err.Error())
		return
	}

	utils.Success(c, file)
}

// AbortUpload 取消上传
// @Summary 取消分片上传
// @Description 取消上传会话，清理临时文件
// @Tags Cloud-Multipart
// @Accept json
// @Produce json
// @Param sessionId path int true "会话ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/cloud/files/multipart/{sessionId} [delete]
func (h *MultipartUploadHandler) AbortUpload(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "sessionId 格式错误")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.multipartService.AbortUpload(c.Request.Context(), uint(sessionID), userID.(uint)); err != nil {
		utils.InternalError(c, "取消上传失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "取消上传成功"})
}

// ResumeUpload 恢复上传
// @Summary 恢复上传（断点续传）
// @Description 根据文件MD5恢复之前未完成的上传会话
// @Tags Cloud-Multipart
// @Accept json
// @Produce json
// @Param request body ResumeUploadRequest true "恢复上传请求"
// @Success 200 {object} cloud.UploadSessionInfo
// @Router /api/v1/cloud/files/multipart/resume [post]
func (h *MultipartUploadHandler) ResumeUpload(c *gin.Context) {
	var req ResumeUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	sessionInfo, err := h.multipartService.ResumeUpload(c.Request.Context(), req.FileMD5, userID.(uint))
	if err != nil {
		utils.InternalError(c, "恢复上传失败: "+err.Error())
		return
	}

	utils.Success(c, sessionInfo)
}

// 请求结构体定义

// InitUploadRequest 初始化上传请求
type InitUploadRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	FileSize    int64  `json:"fileSize" binding:"required"`
	FileMD5     string `json:"fileMd5" binding:"required"`
	FileType    string `json:"fileType"`
	ChunkSize   int    `json:"chunkSize"`
	FolderID    *uint  `json:"folderId"`
	StorageType string `json:"storageType"` // local 或 oss
}

// CompleteUploadRequest 完成上传请求
type CompleteUploadRequest struct {
	SessionID uint `json:"sessionId" binding:"required"`
}

// ResumeUploadRequest 恢复上传请求
type ResumeUploadRequest struct {
	FileMD5 string `json:"fileMd5" binding:"required"`
}
