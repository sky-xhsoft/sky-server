package handler

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/file"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
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
	userID, _ := c.Get("userID")

	// 获取上传IP
	uploadIP := c.ClientIP()

	// 上传文件
	sysFile, err := h.fileService.UploadFile(c.Request.Context(), fileHeader, category, userID.(uint), uploadIP)
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
	userID, _ := c.Get("userID")

	// 获取上传IP
	uploadIP := c.ClientIP()

	// 批量上传
	uploadedFiles, err := h.fileService.UploadMultipleFiles(c.Request.Context(), files, category, userID.(uint), uploadIP)
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
func (h *FileHandler) GetFileByPath(c *gin.Context) {
	storageName := c.Param("storageName")
	if storageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "文件名不能为空",
		})
		return
	}

	// 简化版本：直接从uploads目录查找
	// 注意：这是不安全的，实际应该从数据库查询验证权限
	// TODO: 添加GetFileByStorageName方法进行验证
	baseDir := "./uploads"
	filePath := filepath.Join(baseDir, storageName)

	// 返回文件
	c.File(filePath)
}
