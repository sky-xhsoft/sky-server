package cloud

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"gorm.io/gorm"
)

// Service 云盘服务接口
type Service interface {
	// 文件夹管理
	CreateFolder(ctx context.Context, req *CreateFolderRequest, userID uint) (*entity.CloudFolder, error)
	ListFolders(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudFolder, error)
	GetFolderTree(ctx context.Context, userID uint) ([]*FolderNode, error)
	DeleteFolder(ctx context.Context, folderID uint, userID uint) error
	RenameFolder(ctx context.Context, folderID uint, newName string, userID uint) error

	// 文件管理
	UploadFile(ctx context.Context, req *UploadFileRequest, userID uint) (*entity.CloudFile, error)
	DownloadFile(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, *entity.CloudFile, error)
	DeleteFile(ctx context.Context, fileID uint, userID uint) error
	MoveFile(ctx context.Context, fileID uint, targetFolderID *uint, userID uint) error
	RenameFile(ctx context.Context, fileID uint, newName string, userID uint) error
	ListFiles(ctx context.Context, folderID *uint, userID uint, page, pageSize int) ([]*entity.CloudFile, int64, error)

	// 文件分享
	CreateShare(ctx context.Context, req *CreateShareRequest, userID uint) (*entity.CloudShare, error)
	GetShareInfo(ctx context.Context, shareCode string, password string) (*ShareInfo, error)
	AccessShare(ctx context.Context, shareCode string, password string) (*entity.CloudShare, error)
	CancelShare(ctx context.Context, shareID uint, userID uint) error

	// 配额管理
	GetUserQuota(ctx context.Context, userID uint) (*entity.CloudQuota, error)
	CheckQuota(ctx context.Context, userID uint, fileSize int64) error
	UpdateQuota(ctx context.Context, userID uint, sizeDelta int64, fileDelta int) error
}

// service 云盘服务实现
type service struct {
	db      *gorm.DB
	storage storage.Storage
}

// NewService 创建云盘服务
func NewService(db *gorm.DB, storage storage.Storage) Service {
	return &service{
		db:      db,
		storage: storage,
	}
}

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	Name        string `json:"name" binding:"required"`
	ParentID    *uint  `json:"parentId"`
	Description string `json:"description"`
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	FileName    string
	FolderID    *uint
	FileSize    int64
	FileType    string
	Reader      io.Reader
	StorageType string // local 或 oss
}

// CreateShareRequest 创建分享请求
type CreateShareRequest struct {
	ResourceType string `json:"resourceType" binding:"required"` // file 或 folder
	ResourceID   uint   `json:"resourceId" binding:"required"`
	ShareType    string `json:"shareType" binding:"required"` // public, password, private
	Password     string `json:"password"`
	ExpireDays   int    `json:"expireDays"`   // 过期天数（0=永久）
	MaxDownloads int    `json:"maxDownloads"` // 最大下载次数（0=无限制）
}

// ShareInfo 分享信息
type ShareInfo struct {
	Share        *entity.CloudShare  `json:"share"`
	ResourceType string              `json:"resourceType"`
	File         *entity.CloudFile   `json:"file,omitempty"`
	Folder       *entity.CloudFolder `json:"folder,omitempty"`
	Sharer       string              `json:"sharer"` // 分享者名称
}

// FolderNode 文件夹树节点
type FolderNode struct {
	*entity.CloudFolder
	Children []*FolderNode `json:"children"`
}

// CreateFolder 创建文件夹
func (s *service) CreateFolder(ctx context.Context, req *CreateFolderRequest, userID uint) (*entity.CloudFolder, error) {
	// 构建路径
	path := "/" + req.Name
	if req.ParentID != nil {
		parent, err := s.getFolderByID(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		// 检查权限
		if parent.OwnerID != userID {
			return nil, errors.New(errors.ErrPermissionDenied, "无权限在此文件夹创建子文件夹")
		}
		path = parent.Path + "/" + req.Name
	}

	// 检查同名文件夹
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudFolder{}).
		Where("PARENT_ID = ? AND NAME = ? AND OWNER_ID = ? AND IS_ACTIVE = ?", req.ParentID, req.Name, userID, "Y").
		Count(&count).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "检查文件夹失败", err)
	}
	if count > 0 {
		return nil, errors.New(errors.ErrResourceExists, "同名文件夹已存在")
	}

	// 创建文件夹
	folder := &entity.CloudFolder{
		BaseModel: entity.BaseModel{
			CreateBy: fmt.Sprintf("user_%d", userID),
			UpdateBy: fmt.Sprintf("user_%d", userID),
			IsActive: "Y",
		},
		Name:        req.Name,
		ParentID:    req.ParentID,
		Path:        path,
		OwnerID:     userID,
		Description: req.Description,
		FileCount:   0,
		TotalSize:   0,
	}

	if err := s.db.WithContext(ctx).Create(folder).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件夹失败", err)
	}

	// 更新配额
	s.UpdateQuota(ctx, userID, 0, 0) // 文件夹数量+1在UpdateQuota中处理

	return folder, nil
}

// ListFolders 列出文件夹
func (s *service) ListFolders(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudFolder, error) {
	var folders []*entity.CloudFolder

	query := s.db.WithContext(ctx).
		Where("OWNER_ID = ? AND IS_ACTIVE = ?", userID, "Y")

	if parentID == nil {
		query = query.Where("PARENT_ID IS NULL")
	} else {
		query = query.Where("PARENT_ID = ?", *parentID)
	}

	if err := query.Order("NAME ASC").Find(&folders).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件夹失败", err)
	}

	return folders, nil
}

// GetFolderTree 获取文件夹树
func (s *service) GetFolderTree(ctx context.Context, userID uint) ([]*FolderNode, error) {
	// 查询所有文件夹
	var folders []*entity.CloudFolder
	if err := s.db.WithContext(ctx).
		Where("OWNER_ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Order("PATH ASC").
		Find(&folders).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件夹失败", err)
	}

	// 构建树
	return s.buildFolderTree(folders, nil), nil
}

// buildFolderTree 构建文件夹树
func (s *service) buildFolderTree(folders []*entity.CloudFolder, parentID *uint) []*FolderNode {
	var nodes []*FolderNode

	for _, folder := range folders {
		// 比较父ID
		if (parentID == nil && folder.ParentID == nil) ||
			(parentID != nil && folder.ParentID != nil && *parentID == *folder.ParentID) {
			node := &FolderNode{
				CloudFolder: folder,
				Children:    s.buildFolderTree(folders, &folder.ID),
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// DeleteFolder 删除文件夹
func (s *service) DeleteFolder(ctx context.Context, folderID uint, userID uint) error {
	folder, err := s.getFolderByID(ctx, folderID)
	if err != nil {
		return err
	}

	// 检查权限
	if folder.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "无权限删除此文件夹")
	}

	// 检查是否有子文件夹
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudFolder{}).
		Where("PARENT_ID = ? AND IS_ACTIVE = ?", folderID, "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查子文件夹失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceConflict, "文件夹不为空，无法删除")
	}

	// 检查是否有文件
	if err := s.db.WithContext(ctx).Model(&entity.CloudFile{}).
		Where("FOLDER_ID = ? AND IS_ACTIVE = ?", folderID, "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查文件失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceConflict, "文件夹包含文件，无法删除")
	}

	// 软删除
	if err := s.db.WithContext(ctx).Model(&entity.CloudFolder{}).
		Where("ID = ?", folderID).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除文件夹失败", err)
	}

	return nil
}

// UploadFile 上传文件
func (s *service) UploadFile(ctx context.Context, req *UploadFileRequest, userID uint) (*entity.CloudFile, error) {
	// 检查配额
	if err := s.CheckQuota(ctx, userID, req.FileSize); err != nil {
		return nil, err
	}

	// 构建存储路径
	ext := filepath.Ext(req.FileName)
	storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dateDir := time.Now().Format("2006/01/02")
	storagePath := fmt.Sprintf("cloud/%d/%s/%s", userID, dateDir, storageName)

	// 上传到存储
	accessURL, err := s.storage.Upload(ctx, storagePath, req.Reader, req.FileType)
	if err != nil {
		return nil, err
	}

	// 构建文件路径
	path := "/" + req.FileName
	if req.FolderID != nil {
		folder, err := s.getFolderByID(ctx, *req.FolderID)
		if err != nil {
			return nil, err
		}
		if folder.OwnerID != userID {
			return nil, errors.New(errors.ErrPermissionDenied, "无权限上传到此文件夹")
		}
		path = folder.Path + "/" + req.FileName
	}

	// 创建文件记录
	file := &entity.CloudFile{
		BaseModel: entity.BaseModel{
			CreateBy: fmt.Sprintf("user_%d", userID),
			UpdateBy: fmt.Sprintf("user_%d", userID),
			IsActive: "Y",
		},
		FileName:    req.FileName,
		FolderID:    req.FolderID,
		Path:        path,
		StorageType: req.StorageType,
		StoragePath: storagePath,
		FileSize:    req.FileSize,
		FileType:    req.FileType,
		FileExt:     ext,
		OwnerID:     userID,
		AccessURL:   accessURL,
	}

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		// 删除已上传的文件
		s.storage.Delete(ctx, storagePath)
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
	}

	// 更新配额
	s.UpdateQuota(ctx, userID, req.FileSize, 1)

	return file, nil
}

// CreateShare 创建分享
func (s *service) CreateShare(ctx context.Context, req *CreateShareRequest, userID uint) (*entity.CloudShare, error) {
	// 生成分享码
	shareCode := s.generateShareCode()

	// 计算过期时间
	var expireTime string
	if req.ExpireDays > 0 {
		expireTime = time.Now().AddDate(0, 0, req.ExpireDays).Format("2006-01-02 15:04:05")
	}

	// 创建分享记录
	share := &entity.CloudShare{
		BaseModel: entity.BaseModel{
			CreateBy: fmt.Sprintf("user_%d", userID),
			UpdateBy: fmt.Sprintf("user_%d", userID),
			IsActive: "Y",
		},
		ShareCode:    shareCode,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		SharerID:     userID,
		ShareType:    req.ShareType,
		Password:     req.Password,
		ExpireTime:   expireTime,
		MaxDownloads: req.MaxDownloads,
		Status:       "active",
	}

	if err := s.db.WithContext(ctx).Create(share).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建分享失败", err)
	}

	return share, nil
}

// GetUserQuota 获取用户配额
func (s *service) GetUserQuota(ctx context.Context, userID uint) (*entity.CloudQuota, error) {
	var quota entity.CloudQuota
	err := s.db.WithContext(ctx).
		Where("USER_ID = ? AND IS_ACTIVE = ?", userID, "Y").
		First(&quota).Error

	if err == gorm.ErrRecordNotFound {
		// 创建默认配额
		quota = entity.CloudQuota{
			BaseModel: entity.BaseModel{
				CreateBy: fmt.Sprintf("user_%d", userID),
				UpdateBy: fmt.Sprintf("user_%d", userID),
				IsActive: "Y",
			},
			UserID:      userID,
			TotalQuota:  10 * 1024 * 1024 * 1024, // 默认10GB
			UsedSpace:   0,
			FileCount:   0,
			FolderCount: 0,
			MaxFileSize: 100 * 1024 * 1024, // 默认100MB
			QuotaType:   "standard",
		}
		s.db.WithContext(ctx).Create(&quota)
	} else if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询配额失败", err)
	}

	return &quota, nil
}

// CheckQuota 检查配额
func (s *service) CheckQuota(ctx context.Context, userID uint, fileSize int64) error {
	quota, err := s.GetUserQuota(ctx, userID)
	if err != nil {
		return err
	}

	if quota.UsedSpace+fileSize > quota.TotalQuota {
		return errors.New(errors.ErrResourceConflict, "存储空间不足")
	}

	if fileSize > quota.MaxFileSize {
		return errors.New(errors.ErrInvalidParam, fmt.Sprintf("文件大小超过限制（最大%dMB）", quota.MaxFileSize/(1024*1024)))
	}

	return nil
}

// UpdateQuota 更新配额
func (s *service) UpdateQuota(ctx context.Context, userID uint, sizeDelta int64, fileDelta int) error {
	return s.db.WithContext(ctx).Model(&entity.CloudQuota{}).
		Where("USER_ID = ?", userID).
		Updates(map[string]interface{}{
			"USED_SPACE": gorm.Expr("USED_SPACE + ?", sizeDelta),
			"FILE_COUNT": gorm.Expr("FILE_COUNT + ?", fileDelta),
		}).Error
}

// 辅助方法
func (s *service) getFolderByID(ctx context.Context, folderID uint) (*entity.CloudFolder, error) {
	var folder entity.CloudFolder
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", folderID, "Y").First(&folder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "文件夹不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件夹失败", err)
	}
	return &folder, nil
}

func (s *service) generateShareCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 其他方法的存根实现
func (s *service) RenameFolder(ctx context.Context, folderID uint, newName string, userID uint) error {
	return nil
}

func (s *service) DownloadFile(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, *entity.CloudFile, error) {
	return nil, nil, nil
}

func (s *service) DeleteFile(ctx context.Context, fileID uint, userID uint) error {
	return nil
}

func (s *service) MoveFile(ctx context.Context, fileID uint, targetFolderID *uint, userID uint) error {
	return nil
}

func (s *service) RenameFile(ctx context.Context, fileID uint, newName string, userID uint) error {
	return nil
}

func (s *service) ListFiles(ctx context.Context, folderID *uint, userID uint, page, pageSize int) ([]*entity.CloudFile, int64, error) {
	return nil, 0, nil
}

func (s *service) GetShareInfo(ctx context.Context, shareCode string, password string) (*ShareInfo, error) {
	return nil, nil
}

func (s *service) AccessShare(ctx context.Context, shareCode string, password string) (*entity.CloudShare, error) {
	return nil, nil
}

func (s *service) CancelShare(ctx context.Context, shareID uint, userID uint) error {
	return nil
}
