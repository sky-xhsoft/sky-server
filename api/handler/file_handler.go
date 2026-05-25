package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/file"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService file.Service
}

// NewFileHandler 创建文件处理器
func NewFileHandler(fileService file.Service) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile 上传单个文件
// @Summary 上传文件
// @Tags file
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传文件"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "未找到上传文件",
		})
		return
	}

	// 获取分类
	category := c.DefaultPostForm("category", "default")

	// 获取用户ID
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    errors.ErrUnauthorized,
			"message": "未授权",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "用户ID类型错误",
		})
		return
	}

	// 获取上传IP
	uploadIP := c.ClientIP()

	// 上传文件
	sysFile, err := h.fileService.UploadFile(c.Request.Context(), fileHeader, category, userID, uploadIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "上传成功",
		"data":    sysFile,
	})
}

// UploadMultipleFiles 批量上传文件
// @Summary 批量上传多个文件
// @Tags file
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "上传文件列表"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/upload/multiple [post]
func (h *FileHandler) UploadMultipleFiles(c *gin.Context) {
	// 获取表单
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "解析表单失败: " + err.Error(),
		})
		return
	}

	// 获取所有文件
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "未找到上传文件",
		})
		return
	}

	// 获取分类
	category := c.DefaultPostForm("category", "default")

	// 获取用户ID
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    errors.ErrUnauthorized,
			"message": "未授权",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrInternal,
			"message": "用户ID类型错误",
		})
		return
	}

	// 获取上传IP
	uploadIP := c.ClientIP()

	// 批量上传
	uploadedFiles, err := h.fileService.UploadMultipleFiles(c.Request.Context(), files, category, userID, uploadIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "上传成功",
		"data": gin.H{
			"total":   len(files),
			"success": len(uploadedFiles),
			"files":   uploadedFiles,
		},
	})
}

// DownloadFile 下载文件
// @Summary 下载指定文件
// @Tags file
// @Produce octet-stream
// @Param id path int true "文件ID"
// @Success 200 {file} binary
// @Router /api/v1/files/download/{id} [get]
func (h *FileHandler) DownloadFile(c *gin.Context) {
	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的文件ID",
		})
		return
	}

	// 获取用户ID
	userID, _ := c.Get("userID")

	// 获取文件信息并更新下载次数
	sysFile, err := h.fileService.DownloadFile(c.Request.Context(), uint(fileID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+sysFile.FileName)
	c.Header("Content-Type", sysFile.FileType)

	// 返回文件
	c.File(sysFile.FilePath)
}

// PreviewFile 预览文件（不增加下载次数）
// @Summary 在线预览文件
// @Tags file
// @Produce octet-stream
// @Param id path int true "文件ID"
// @Success 200 {file} binary
// @Router /api/v1/files/preview/{id} [get]
func (h *FileHandler) PreviewFile(c *gin.Context) {
	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的文件ID",
		})
		return
	}

	// 获取文件信息
	sysFile, err := h.fileService.GetFile(c.Request.Context(), uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	// 设置响应头（inline模式，浏览器尝试预览）
	c.Header("Content-Type", sysFile.FileType)
	c.Header("Content-Disposition", "inline; filename="+sysFile.FileName)

	// 返回文件
	c.File(sysFile.FilePath)
}

// DeleteFile 删除文件
// @Summary 删除指定文件
// @Tags file
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的文件ID",
		})
		return
	}

	// 获取用户ID
	userID, _ := c.Get("userID")

	// 删除文件
	if err := h.fileService.DeleteFile(c.Request.Context(), uint(fileID), userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// GetFile 获取文件信息
// @Summary 获取文件元数据信息
// @Tags file
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的文件ID",
		})
		return
	}

	// 获取文件信息
	sysFile, err := h.fileService.GetFile(c.Request.Context(), uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    sysFile,
	})
}

// ListFiles 查询文件列表
// @Summary 分页查询文件列表
// @Tags file
// @Accept json
// @Produce json
// @Param request body file.ListFilesRequest true "查询请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/files/list [post]
func (h *FileHandler) ListFiles(c *gin.Context) {
	var req file.ListFilesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 查询文件列表
	files, total, err := h.fileService.ListFiles(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
			"list":     files,
		},
	})
}

// GetFileByPath 根据存储名称直接获取文件（用于访问URL）
// @Summary 通过存储名访问文件
// @Tags file
// @Produce octet-stream
// @Param storageName path string true "文件存储名"
// @Success 200 {file} binary
// @Router /api/v1/files/access/{storageName} [get]
func (h *FileHandler) GetFileByPath(c *gin.Context) {
	storageName := c.Param("storageName")
	if storageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "文件名不能为空",
		})
		return
	}

	// 从数据库查询文件信息
	sysFile, err := h.fileService.GetFileByStorageName(c.Request.Context(), storageName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    errors.GetCode(err),
			"message": "文件不存在",
		})
		return
	}

	// 设置响应头（inline模式，浏览器尝试预览）
	c.Header("Content-Type", sysFile.FileType)
	c.Header("Content-Disposition", "inline; filename="+sysFile.FileName)
	c.Header("Cache-Control", "public, max-age=31536000") // 缓存1年

	// 返回文件
	c.File(sysFile.FilePath)
}
