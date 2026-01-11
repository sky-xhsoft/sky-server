package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"gorm.io/gorm"
)

// Service 文件服务接口
type Service interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, file *multipart.FileHeader, category string, userID uint, uploadIP string) (*entity.SysFile, error)

	// UploadMultipleFiles 批量上传文件
	UploadMultipleFiles(ctx context.Context, files []*multipart.FileHeader, category string, userID uint, uploadIP string) ([]*entity.SysFile, error)

	// DownloadFile 下载文件（返回文件路径）
	DownloadFile(ctx context.Context, fileID uint, userID uint) (*entity.SysFile, error)

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, fileID uint, userID uint) error

	// GetFile 获取文件信息
	GetFile(ctx context.Context, fileID uint) (*entity.SysFile, error)

	// ListFiles 查询文件列表
	ListFiles(ctx context.Context, req *ListFilesRequest) ([]*entity.SysFile, int64, error)

	// GetFileByMD5 根据MD5获取文件（用于秒传）
	GetFileByMD5(ctx context.Context, md5 string) (*entity.SysFile, error)
}

// ListFilesRequest 文件列表查询请求
type ListFilesRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	Category  string `json:"category"`
	FileName  string `json:"fileName"`
	FileType  string `json:"fileType"`
	CreateBy  string `json:"createBy"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// service 文件服务实现
type service struct {
	db           *gorm.DB
	uploadDir    string // 上传目录
	maxFileSize  int64  // 最大文件大小（字节）
	allowedExts  []string // 允许的文件扩展名
}

// Config 文件服务配置
type Config struct {
	UploadDir   string   // 上传目录
	MaxFileSize int64    // 最大文件大小（字节）
	AllowedExts []string // 允许的文件扩展名
}

// NewService 创建文件服务
func NewService(db *gorm.DB, cfg *Config) Service {
	// 设置默认值
	if cfg.UploadDir == "" {
		cfg.UploadDir = "./uploads"
	}
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 100 * 1024 * 1024 // 默认100MB
	}
	if len(cfg.AllowedExts) == 0 {
		cfg.AllowedExts = []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".txt", ".zip", ".rar"}
	}

	// 确保上传目录存在
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		panic(fmt.Sprintf("创建上传目录失败: %v", err))
	}

	return &service{
		db:          db,
		uploadDir:   cfg.UploadDir,
		maxFileSize: cfg.MaxFileSize,
		allowedExts: cfg.AllowedExts,
	}
}

// UploadFile 上传文件
func (s *service) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, category string, userID uint, uploadIP string) (*entity.SysFile, error) {
	// 验证文件大小
	if fileHeader.Size > s.maxFileSize {
		return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("文件大小超过限制（最大%dMB）", s.maxFileSize/(1024*1024)))
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedExt(ext) {
		return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("不支持的文件类型: %s", ext))
	}

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "打开文件失败", err)
	}
	defer file.Close()

	// 计算文件MD5
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "计算文件MD5失败", err)
	}
	md5Sum := hex.EncodeToString(hash.Sum(nil))

	// 检查是否已存在相同MD5的文件（秒传）
	existingFile, err := s.GetFileByMD5(ctx, md5Sum)
	if err == nil && existingFile != nil {
		// 文件已存在，创建新的文件记录但共享同一个物理文件
		newFile := &entity.SysFile{
			BaseModel: entity.BaseModel{
				CreateBy: fmt.Sprintf("user_%d", userID),
				UpdateBy: fmt.Sprintf("user_%d", userID),
				IsActive: "Y",
			},
			FileName:      fileHeader.Filename,
			StorageName:   existingFile.StorageName,
			FilePath:      existingFile.FilePath,
			FileSize:      fileHeader.Size,
			FileType:      fileHeader.Header.Get("Content-Type"),
			FileExt:       ext,
			StorageType:   "local",
			AccessURL:     existingFile.AccessURL,
			MD5:           md5Sum,
			UploadIP:      uploadIP,
			DownloadCount: 0,
			Category:      category,
		}

		if err := s.db.WithContext(ctx).Create(newFile).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
		}

		return newFile, nil
	}

	// 生成存储文件名（UUID + 扩展名）
	storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 按日期创建子目录
	dateDir := time.Now().Format("2006/01/02")
	targetDir := filepath.Join(s.uploadDir, dateDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建目录失败", err)
	}

	// 完整的文件路径
	fullPath := filepath.Join(targetDir, storageName)

	// 重置文件读取位置
	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "重置文件位置失败", err)
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建文件失败", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		// 删除已创建的文件
		os.Remove(fullPath)
		return nil, errors.Wrap(errors.ErrInternal, "保存文件失败", err)
	}

	// 生成访问URL
	accessURL := fmt.Sprintf("/api/v1/files/download/%s", storageName)

	// 创建文件记录
	sysFile := &entity.SysFile{
		BaseModel: entity.BaseModel{
			CreateBy: fmt.Sprintf("user_%d", userID),
			UpdateBy: fmt.Sprintf("user_%d", userID),
			IsActive: "Y",
		},
		FileName:      fileHeader.Filename,
		StorageName:   storageName,
		FilePath:      fullPath,
		FileSize:      fileHeader.Size,
		FileType:      fileHeader.Header.Get("Content-Type"),
		FileExt:       ext,
		StorageType:   "local",
		AccessURL:     accessURL,
		MD5:           md5Sum,
		UploadIP:      uploadIP,
		DownloadCount: 0,
		Category:      category,
	}

	if err := s.db.WithContext(ctx).Create(sysFile).Error; err != nil {
		// 删除已上传的文件
		os.Remove(fullPath)
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
	}

	return sysFile, nil
}

// UploadMultipleFiles 批量上传文件
func (s *service) UploadMultipleFiles(ctx context.Context, files []*multipart.FileHeader, category string, userID uint, uploadIP string) ([]*entity.SysFile, error) {
	result := make([]*entity.SysFile, 0, len(files))

	for _, file := range files {
		sysFile, err := s.UploadFile(ctx, file, category, userID, uploadIP)
		if err != nil {
			// 继续处理其他文件，但记录错误
			continue
		}
		result = append(result, sysFile)
	}

	if len(result) == 0 {
		return nil, errors.New(errors.ErrInternal, "所有文件上传失败")
	}

	return result, nil
}

// DownloadFile 下载文件
func (s *service) DownloadFile(ctx context.Context, fileID uint, userID uint) (*entity.SysFile, error) {
	sysFile, err := s.GetFile(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(sysFile.FilePath); os.IsNotExist(err) {
		return nil, errors.New(errors.ErrResourceNotFound, "文件不存在")
	}

	// 更新下载次数
	if err := s.db.WithContext(ctx).Model(&entity.SysFile{}).
		Where("ID = ?", fileID).
		UpdateColumn("DOWNLOAD_COUNT", gorm.Expr("DOWNLOAD_COUNT + 1")).Error; err != nil {
		// 下载次数更新失败不影响下载
	}

	return sysFile, nil
}

// DeleteFile 删除文件
func (s *service) DeleteFile(ctx context.Context, fileID uint, userID uint) error {
	sysFile, err := s.GetFile(ctx, fileID)
	if err != nil {
		return err
	}

	// 检查是否有其他记录引用同一个物理文件（相同MD5）
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.SysFile{}).
		Where("MD5 = ? AND IS_ACTIVE = ?", sysFile.MD5, "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查文件引用失败", err)
	}

	// 软删除文件记录
	if err := s.db.WithContext(ctx).Model(&entity.SysFile{}).
		Where("ID = ?", fileID).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除文件记录失败", err)
	}

	// 如果只有一个引用，删除物理文件
	if count == 1 {
		if err := os.Remove(sysFile.FilePath); err != nil && !os.IsNotExist(err) {
			// 物理文件删除失败，记录日志但不返回错误
		}
	}

	return nil
}

// GetFile 获取文件信息
func (s *service) GetFile(ctx context.Context, fileID uint) (*entity.SysFile, error) {
	var sysFile entity.SysFile
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", fileID, "Y").
		First(&sysFile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "文件不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件失败", err)
	}

	return &sysFile, nil
}

// ListFiles 查询文件列表
func (s *service) ListFiles(ctx context.Context, req *ListFilesRequest) ([]*entity.SysFile, int64, error) {
	// 设置默认分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	query := s.db.WithContext(ctx).Model(&entity.SysFile{}).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	if req.Category != "" {
		query = query.Where("CATEGORY = ?", req.Category)
	}
	if req.FileName != "" {
		query = query.Where("FILE_NAME LIKE ?", "%"+req.FileName+"%")
	}
	if req.FileType != "" {
		query = query.Where("FILE_TYPE LIKE ?", "%"+req.FileType+"%")
	}
	if req.CreateBy != "" {
		query = query.Where("CREATE_BY = ?", req.CreateBy)
	}
	if req.StartTime != "" {
		query = query.Where("CREATE_TIME >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("CREATE_TIME <= ?", req.EndTime)
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询文件总数失败", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var files []*entity.SysFile
	if err := query.Order("CREATE_TIME DESC").Limit(req.PageSize).Offset(offset).Find(&files).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询文件列表失败", err)
	}

	return files, total, nil
}

// GetFileByMD5 根据MD5获取文件
func (s *service) GetFileByMD5(ctx context.Context, md5 string) (*entity.SysFile, error) {
	var sysFile entity.SysFile
	if err := s.db.WithContext(ctx).
		Where("MD5 = ? AND IS_ACTIVE = ?", md5, "Y").
		First(&sysFile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件失败", err)
	}

	return &sysFile, nil
}

// isAllowedExt 检查文件扩展名是否允许
func (s *service) isAllowedExt(ext string) bool {
	for _, allowedExt := range s.allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}
