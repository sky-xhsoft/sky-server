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
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// Service 云盘服务接口
type Service interface {
	// CloudItem 统一管理（新）
	CreateItem(ctx context.Context, item *entity.CloudItem) error
	GetItem(ctx context.Context, itemID uint, userID uint) (*entity.CloudItem, error)
	ListItems(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudItem, error)
	UpdateItem(ctx context.Context, item *entity.CloudItem) error
	DeleteItem(ctx context.Context, itemID uint, userID uint) error
	DeleteItemsByParentID(ctx context.Context, parentID uint, userID uint) error
	MoveItem(ctx context.Context, itemID uint, targetParentID *uint, userID uint) error
	RenameItem(ctx context.Context, itemID uint, newName string, userID uint) error

	// 文件夹管理
	CreateFolder(ctx context.Context, req *CreateFolderRequest, userID uint) (*entity.CloudFolder, error)
	ListFolders(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudFolder, error)
	GetFolderTree(ctx context.Context, userID uint) ([]*FolderNode, error)
	DeleteFolder(ctx context.Context, folderID uint, userID uint) error
	RenameFolder(ctx context.Context, folderID uint, newName string, userID uint) error

	// 文件管理
	UploadFile(ctx context.Context, req *UploadFileRequest, userID uint) (*entity.CloudItem, error)
	DownloadFile(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, *entity.CloudItem, error)
	DeleteFile(ctx context.Context, fileID uint, userID uint) error
	MoveFile(ctx context.Context, fileID uint, targetFolderID *uint, userID uint) error
	RenameFile(ctx context.Context, fileID uint, newName string, userID uint) error
	ListFiles(ctx context.Context, folderID *uint, userID uint, page, pageSize int) ([]*entity.CloudItem, int64, error)

	// 文件分享
	CreateShare(ctx context.Context, req *CreateShareRequest, userID uint) (*entity.CloudShare, error)
	GetShareInfo(ctx context.Context, shareCode string, password string) (*ShareInfo, error)
	GetShareByCode(ctx context.Context, shareCode string) (*entity.CloudShare, error)
	GetUserShares(ctx context.Context, userID uint) ([]*entity.CloudShare, error)
	AccessShare(ctx context.Context, shareCode string, password string) (*ShareInfo, error)
	CancelShare(ctx context.Context, shareID uint, userID uint) error
	DownloadFileByID(ctx context.Context, fileID uint) (io.ReadCloser, *entity.CloudItem, error)
	IncrementShareDownloadCount(ctx context.Context, shareID uint) error

	// 配额管理
	GetUserQuota(ctx context.Context, userID uint) (*entity.CloudQuota, error)
	CheckQuota(ctx context.Context, userID uint, fileSize int64) error
	UpdateQuota(ctx context.Context, userID uint, sizeDelta int64, fileDelta int, folderDelta int) error
	RecalculateQuota(ctx context.Context, userID uint) error
}

// service 云盘服务实现
type service struct {
	db                    *gorm.DB
	companyStorageManager *storage.CompanyStorageManager
	userRepo              repository.UserRepository
}

// NewService 创建云盘服务
func NewService(db *gorm.DB, companyStorageManager *storage.CompanyStorageManager, userRepo repository.UserRepository) Service {
	return &service{
		db:                    db,
		companyStorageManager: companyStorageManager,
		userRepo:              userRepo,
	}
}

// getCompanyStorage 根据用户ID获取公司存储实例
func (s *service) getCompanyStorage(ctx context.Context, userID uint) (storage.Storage, error) {
	// 从用户ID获取用户信息
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "获取用户信息失败", err)
	}

	// 获取公司存储实例
	companyStorage, err := s.companyStorageManager.GetStorage(ctx, user.SysCompanyID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrStorage, "获取公司存储实例失败", err)
	}

	return companyStorage, nil
}

// getUsernameByID 根据用户ID获取用户名
func (s *service) getUsernameByID(ctx context.Context, userID uint) string {
	var user entity.SysUser
	if err := s.db.WithContext(ctx).Select("USERNAME").Where("ID = ?", userID).First(&user).Error; err != nil {
		// 如果查询失败，返回默认值
		return fmt.Sprintf("user_%d", userID)
	}
	return user.Username
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
	File         *entity.CloudItem   `json:"file,omitempty"`
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
	if req.ParentID != nil && *req.ParentID != 0 {
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
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("PARENT_ID = ? AND NAME = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", req.ParentID, req.Name, userID, "Y", "folder").
		Count(&count).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "检查文件夹失败", err)
	}
	if count > 0 {
		return nil, errors.New(errors.ErrResourceExists, "同名文件夹已存在")
	}

	// 创建文件夹
	username := s.getUsernameByID(ctx, userID)
	item := &entity.CloudItem{
		BaseModel: entity.BaseModel{
			CreateBy: username,
			UpdateBy: username,
			IsActive: "Y",
		},
		ItemType:    "folder",
		Name:        req.Name,
		ParentID:    req.ParentID,
		Path:        path,
		OwnerID:     userID,
		Description: req.Description,
		FileCount:   0,
		TotalSize:   0,
	}

	if err := s.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件夹失败", err)
	}

	// 更新配额：文件夹数量+1
	s.UpdateQuota(ctx, userID, 0, 0, 1)

	// 转换为 CloudFolder 格式返回（兼容旧接口）
	return item.ToFolder(), nil
}

// ListFolders 列出文件夹
func (s *service) ListFolders(ctx context.Context, parentID *uint, userID uint) ([]*entity.CloudFolder, error) {
	// 从 cloud_item 表查询文件夹
	var items []*entity.CloudItem

	query := s.db.WithContext(ctx).
		Where("OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", userID, "Y", "folder")

	if parentID == nil {
		query = query.Where("PARENT_ID IS NULL")
	} else {
		query = query.Where("PARENT_ID = ?", *parentID)
	}

	if err := query.Order("NAME ASC").Find(&items).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件夹失败", err)
	}

	// 转换为 CloudFolder 格式（兼容旧接口）
	folders := make([]*entity.CloudFolder, 0, len(items))
	for _, item := range items {
		if folder := item.ToFolder(); folder != nil {
			folders = append(folders, folder)
		}
	}

	return folders, nil
}

// GetFolderTree 获取文件夹树
func (s *service) GetFolderTree(ctx context.Context, userID uint) ([]*FolderNode, error) {
	// 从 cloud_item 表查询所有文件夹
	var items []*entity.CloudItem
	if err := s.db.WithContext(ctx).
		Where("OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", userID, "Y", "folder").
		Order("PATH ASC").
		Find(&items).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件夹失败", err)
	}

	// 转换为 CloudFolder 格式
	folders := make([]*entity.CloudFolder, 0, len(items))
	for _, item := range items {
		if folder := item.ToFolder(); folder != nil {
			folders = append(folders, folder)
		}
	}

	// 构建树
	return s.buildFolderTreeOptimized(folders), nil
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

// buildFolderTreeOptimized 优化的文件夹树构建方法 - O(n)时间复杂度
// 使用 Map 预处理父子关系，避免每次递归都遍历整个列表
func (s *service) buildFolderTreeOptimized(folders []*entity.CloudFolder) []*FolderNode {
	// 1. 构建 parentID -> children 的映射表 O(n)
	childrenMap := make(map[uint][]*entity.CloudFolder)
	var roots []*entity.CloudFolder

	for _, folder := range folders {
		if folder.ParentID == nil {
			// 根节点（无父节点）
			roots = append(roots, folder)
		} else {
			// 子节点，添加到父节点的 children 列表中
			childrenMap[*folder.ParentID] = append(childrenMap[*folder.ParentID], folder)
		}
	}

	// 2. 递归构建树节点 O(n)
	var result []*FolderNode
	for _, root := range roots {
		result = append(result, s.buildNodeRecursively(root, childrenMap))
	}
	return result
}

// buildNodeRecursively 递归构建节点
// 使用预处理的 childrenMap，避免重复遍历
func (s *service) buildNodeRecursively(folder *entity.CloudFolder, childrenMap map[uint][]*entity.CloudFolder) *FolderNode {
	node := &FolderNode{CloudFolder: folder}

	// 从 Map 中直接获取子节点，O(1) 查找
	if children, ok := childrenMap[folder.ID]; ok {
		for _, child := range children {
			node.Children = append(node.Children, s.buildNodeRecursively(child, childrenMap))
		}
	}

	return node
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
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("PARENT_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", folderID, "Y", "folder").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查子文件夹失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceConflict, "文件夹不为空，无法删除")
	}

	// 检查是否有文件
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("PARENT_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", folderID, "file", "Y").
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查文件失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceConflict, "文件夹包含文件，无法删除")
	}

	// 软删除
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("ID = ? AND ITEM_TYPE = ?", folderID, "folder").
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除文件夹失败", err)
	}

	// 更新配额：文件夹数量-1
	s.UpdateQuota(ctx, userID, 0, 0, -1)

	return nil
}

// UploadFile 上传文件
func (s *service) UploadFile(ctx context.Context, req *UploadFileRequest, userID uint) (*entity.CloudItem, error) {
	// 检查配额
	if err := s.CheckQuota(ctx, userID, req.FileSize); err != nil {
		return nil, err
	}

	// 获取公司存储实例
	storageInstance, err := s.getCompanyStorage(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 构建存储路径
	ext := filepath.Ext(req.FileName)
	storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dateDir := time.Now().Format("2006/01/02")
	storagePath := fmt.Sprintf("cloud/%d/%s/%s", userID, dateDir, storageName)

	// 上传到存储
	accessURL, err := storageInstance.Upload(ctx, storagePath, req.Reader, req.FileType)
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

	// 创建文件记录 - 使用 CloudItem
	username := s.getUsernameByID(ctx, userID)
	// 获取正确的存储类型（与分片上传保持一致）
	var storageType string
	user, getuserErr := s.userRepo.GetUserByID(userID)
	if getuserErr == nil {
		companyConfig, getConfigErr := s.companyStorageManager.GetConfig(ctx, user.SysCompanyID)
		if getConfigErr == nil && companyConfig != nil {
			storageType = companyConfig.StorageType
		}
	}
	// 如果获取配置失败，根据存储实例类型推断
	if storageType == "" {
		switch storageInstance.(type) {
		case *storage.LocalStorage:
			storageType = "local"
		default:
			// 阿里云OSS和腾讯云COS都统一保存为 "oss"
			if _, ok := storageInstance.(interface{ IsAliyunOSS() bool }); ok {
				storageType = "oss"
			} else if _, ok := storageInstance.(interface{ IsTencentCOS() bool }); ok {
				storageType = "oss" // 腾讯云也保存为 oss
			} else {
				storageType = "local"
			}
		}
	} else {
		// 如果从配置获取到的是长格式类型，统一转换为 "oss"
		switch storageType {
		case "aliyunOSS":
			storageType = "oss"
		case "tencentCOS":
			storageType = "oss" // 腾讯云也转换为 oss
		case "cos":
			storageType = "oss" // 已有的 cos 也转换为 oss
		}
	}
	fileSize := req.FileSize
	fileType := req.FileType
	item := &entity.CloudItem{
		BaseModel: entity.BaseModel{
			CreateBy: username,
			UpdateBy: username,
			IsActive: "Y",
		},
		ItemType: "file",
		Name:     req.FileName,
		ParentID: req.FolderID,
		Path:     path,
		OwnerID:  userID,
		// 文件专属字段使用指针
		StorageType: &storageType,
		StoragePath: &storagePath,
		FileSize:    &fileSize,
		FileType:    &fileType,
		FileExt:     &ext,
		AccessURL:   &accessURL,
	}

	if err := s.db.WithContext(ctx).Create(item).Error; err != nil {
		// 删除已上传的文件
		storageInstance.Delete(ctx, storagePath)
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
	}

	// 更新配额：文件数量+1
	s.UpdateQuota(ctx, userID, req.FileSize, 1, 0)

	return item, nil
}

// CreateShare 创建分享
func (s *service) CreateShare(ctx context.Context, req *CreateShareRequest, userID uint) (*entity.CloudShare, error) {
	// 生成分享码
	shareCode := s.generateShareCode()

	// 计算过期时间
	var expireTime *time.Time
	if req.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, req.ExpireDays)
		expireTime = &t
	}

	// 创建分享记录
	username := s.getUsernameByID(ctx, userID)
	share := &entity.CloudShare{
		BaseModel: entity.BaseModel{
			CreateBy: username,
			UpdateBy: username,
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
		username := s.getUsernameByID(ctx, userID)
		quota = entity.CloudQuota{
			BaseModel: entity.BaseModel{
				CreateBy: username,
				UpdateBy: username,
				IsActive: "Y",
			},
			UserID:      userID,
			TotalQuota:  10 * 1024 * 1024 * 1024, // 默认10GB
			UsedSpace:   0,
			FileCount:   0,
			FolderCount: 0,
			MaxFileSize: 20 * 1024 * 1024 * 1024, // 默认20GB
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
func (s *service) UpdateQuota(ctx context.Context, userID uint, sizeDelta int64, fileDelta int, folderDelta int) error {
	updates := map[string]interface{}{
		"USED_SPACE": gorm.Expr("USED_SPACE + ?", sizeDelta),
		"FILE_COUNT": gorm.Expr("FILE_COUNT + ?", fileDelta),
	}

	// 只有当 folderDelta 不为 0 时才更新文件夹数量
	if folderDelta != 0 {
		updates["FOLDER_COUNT"] = gorm.Expr("FOLDER_COUNT + ?", folderDelta)
	}

	return s.db.WithContext(ctx).Model(&entity.CloudQuota{}).
		Where("USER_ID = ?", userID).
		Updates(updates).Error
}

// RecalculateQuota 重新计算用户配额（从数据库统计）
func (s *service) RecalculateQuota(ctx context.Context, userID uint) error {
	// 统计文件数量和总大小
	var fileStats struct {
		FileCount int64
		TotalSize int64
	}

	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Select("COUNT(*) as file_count, COALESCE(SUM(FILE_SIZE), 0) as total_size").
		Where("OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", userID, "file", "Y").
		Scan(&fileStats).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "统计文件失败", err)
	}

	// 统计文件夹数量
	var folderCount int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", userID, "folder", "Y").
		Count(&folderCount).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "统计文件夹失败", err)
	}

	// 更新配额表
	updates := map[string]interface{}{
		"USED_SPACE":   fileStats.TotalSize,
		"FILE_COUNT":   fileStats.FileCount,
		"FOLDER_COUNT": folderCount,
	}

	return s.db.WithContext(ctx).Model(&entity.CloudQuota{}).
		Where("USER_ID = ?", userID).
		Updates(updates).Error
}

// 辅助方法
func (s *service) getFolderByID(ctx context.Context, folderID uint) (*entity.CloudItem, error) {
	var item entity.CloudItem
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", folderID, "folder", "Y").
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "文件夹不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件夹失败", err)
	}
	return &item, nil
}

func (s *service) getFileByID(ctx context.Context, fileID uint) (*entity.CloudItem, error) {
	var item entity.CloudItem
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", fileID, "file", "Y").
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "文件不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件失败", err)
	}
	return &item, nil
}

func (s *service) updateChildrenPaths(ctx context.Context, oldPath, newPath string) error {
	// 更新所有子文件夹的路径
	if err := s.db.WithContext(ctx).Exec(`
		UPDATE cloud_folder
		SET PATH = CONCAT(?, SUBSTRING(PATH, ?))
		WHERE PATH LIKE ? AND IS_ACTIVE = 'Y'
	`, newPath, len(oldPath)+1, oldPath+"/%").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新子文件夹路径失败", err)
	}

	// 更新所有子文件的路径 - 使用 cloud_item 表
	if err := s.db.WithContext(ctx).Exec(`
		UPDATE cloud_item
		SET PATH = CONCAT(?, SUBSTRING(PATH, ?))
		WHERE PATH LIKE ? AND IS_ACTIVE = 'Y'
	`, newPath, len(oldPath)+1, oldPath+"/%").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新子文件路径失败", err)
	}

	return nil
}

func (s *service) generateShareCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RenameFolder 重命名文件夹
func (s *service) RenameFolder(ctx context.Context, folderID uint, newName string, userID uint) error {
	folder, err := s.getFolderByID(ctx, folderID)
	if err != nil {
		return err
	}

	// 检查权限
	if folder.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "无权限重命名此文件夹")
	}

	// 检查同名文件夹
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("PARENT_ID = ? AND NAME = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ? AND ID != ?",
			folder.ParentID, newName, userID, "Y", "folder", folderID).
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查文件夹失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceExists, "同名文件夹已存在")
	}

	// 更新文件夹名称和路径
	oldPath := folder.Path
	newPath := oldPath[:len(oldPath)-len(folder.Name)] + newName

	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("ID = ? AND ITEM_TYPE = ?", folderID, "folder").
		Updates(map[string]interface{}{
			"NAME":        newName,
			"PATH":        newPath,
			"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
			"UPDATE_TIME": time.Now(),
		}).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "重命名文件夹失败", err)
	}

	// 更新所有子文件夹和文件的路径
	if err := s.updateChildrenPaths(ctx, oldPath, newPath); err != nil {
		return err
	}

	return nil
}

// DownloadFile 下载文件
func (s *service) DownloadFile(ctx context.Context, fileID uint, userID uint) (io.ReadCloser, *entity.CloudItem, error) {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}

	// 检查权限
	if file.OwnerID != userID {
		return nil, nil, errors.New(errors.ErrPermissionDenied, "无权限下载此文件")
	}

	// StoragePath 是指针，需要解引用
	if file.StoragePath == nil {
		return nil, nil, errors.New(errors.ErrInvalidParam, "文件存储路径不存在")
	}

	// 获取公司存储实例
	storageInstance, err := s.getCompanyStorage(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	// 从存储中下载文件
	reader, err := storageInstance.Download(ctx, *file.StoragePath)
	if err != nil {
		return nil, nil, err
	}

	// 更新下载次数 - 使用 CloudItem 表并过滤 ITEM_TYPE = "file"
	s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("ID = ? AND ITEM_TYPE = ?", fileID, "file").
		Update("DOWNLOAD_COUNT", gorm.Expr("DOWNLOAD_COUNT + 1"))

	return reader, file, nil
}

// DeleteFile 删除文件
func (s *service) DeleteFile(ctx context.Context, fileID uint, userID uint) error {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	// 检查权限
	if file.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "无权限删除此文件")
	}

	// 软删除文件记录 - 使用 CloudItem 表并过滤 ITEM_TYPE = "file"
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("ID = ? AND ITEM_TYPE = ?", fileID, "file").
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除文件失败", err)
	}

	// 重新计算配额（从数据库统计）
	_ = s.RecalculateQuota(ctx, userID)

	// 异步删除物理文件（可选）
	if file.StoragePath != nil {
		// 获取公司存储实例
		storageInstance, err := s.getCompanyStorage(ctx, userID)
		if err == nil {
			go func() {
				_ = storageInstance.Delete(context.Background(), *file.StoragePath)
			}()
		}
	}

	return nil
}

// MoveFile 移动文件
func (s *service) MoveFile(ctx context.Context, fileID uint, targetFolderID *uint, userID uint) error {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	// 检查权限
	if file.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "无权限移动此文件")
	}

	// 检查目标文件夹
	var targetPath string
	if targetFolderID != nil {
		targetFolder, err := s.getFolderByID(ctx, *targetFolderID)
		if err != nil {
			return err
		}
		if targetFolder.OwnerID != userID {
			return errors.New(errors.ErrPermissionDenied, "无权限移动到此文件夹")
		}
		targetPath = targetFolder.Path
	}

	// 检查目标位置是否已有同名文件 - 使用 CloudItem 表并过滤 ITEM_TYPE = "file"
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("PARENT_ID = ? AND NAME = ? AND ITEM_TYPE = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ID != ?",
			targetFolderID, file.Name, "file", userID, "Y", fileID).
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查文件失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceExists, "目标文件夹中已存在同名文件")
	}

	// 更新文件位置和路径
	newPath := targetPath + "/" + file.Name
	if targetFolderID == nil {
		newPath = "/" + file.Name
	}

	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("ID = ? AND ITEM_TYPE = ?", fileID, "file").
		Updates(map[string]interface{}{
			"PARENT_ID":   targetFolderID,
			"PATH":        newPath,
			"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
			"UPDATE_TIME": time.Now(),
		}).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "移动文件失败", err)
	}

	return nil
}

// RenameFile 重命名文件
func (s *service) RenameFile(ctx context.Context, fileID uint, newName string, userID uint) error {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	// 检查权限
	if file.OwnerID != userID {
		return errors.New(errors.ErrPermissionDenied, "无权限重命名此文件")
	}

	// 检查同名文件 - 使用 CloudItem 表并过滤 ITEM_TYPE = "file"
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("PARENT_ID = ? AND NAME = ? AND ITEM_TYPE = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ID != ?",
			file.ParentID, newName, "file", userID, "Y", fileID).
		Count(&count).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "检查文件失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrResourceExists, "同名文件已存在")
	}

	// 更新文件名称和路径
	oldPath := file.Path
	newPath := oldPath[:len(oldPath)-len(file.Name)] + newName

	// 获取新的扩展名
	newExt := filepath.Ext(newName)

	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("ID = ? AND ITEM_TYPE = ?", fileID, "file").
		Updates(map[string]interface{}{
			"NAME":        newName,
			"FILE_EXT":    newExt,
			"PATH":        newPath,
			"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
			"UPDATE_TIME": time.Now(),
		}).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "重命名文件失败", err)
	}

	return nil
}

// ListFiles 列出文件
func (s *service) ListFiles(ctx context.Context, folderID *uint, userID uint, page, pageSize int) ([]*entity.CloudItem, int64, error) {
	var files []*entity.CloudItem
	var total int64

	query := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", userID, "file", "Y")

	if folderID == nil {
		query = query.Where("PARENT_ID IS NULL")
	} else {
		query = query.Where("PARENT_ID = ?", *folderID)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询文件总数失败", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("CREATE_TIME DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询文件失败", err)
	}

	return files, total, nil
}

// GetShareInfo 获取分享信息
func (s *service) GetShareInfo(ctx context.Context, shareCode string, password string) (*ShareInfo, error) {
	var share entity.CloudShare
	if err := s.db.WithContext(ctx).
		Where("SHARE_CODE = ? AND IS_ACTIVE = ?", shareCode, "Y").
		First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "分享不存在或已失效")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询分享失败", err)
	}

	// 检查状态
	if share.Status != "active" {
		return nil, errors.New(errors.ErrResourceNotFound, "分享已失效")
	}

	// 检查过期时间
	if share.ExpireTime != nil && share.ExpireTime.Before(time.Now()) {
		// 更新状态为过期
		s.db.WithContext(ctx).Model(&entity.CloudShare{}).
			Where("ID = ?", share.ID).
			Update("STATUS", "expired")
		return nil, errors.New(errors.ErrResourceNotFound, "分享已过期")
	}

	// 构建基本信息（不包含文件/文件夹详情）
	info := &ShareInfo{
		Share:        &share,
		ResourceType: share.ResourceType,
	}

	// 查询分享者信息
	var user entity.SysUser
	if err := s.db.WithContext(ctx).Where("ID = ?", share.SharerID).First(&user).Error; err == nil {
		info.Sharer = user.Username
	}

	// 始终加载文件/文件夹的基本信息（名称等），用于显示
	if share.ResourceType == "file" {
		var item entity.CloudItem
		if err := s.db.WithContext(ctx).
			Where("ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", share.ResourceID, "file", "Y").
			First(&item).Error; err == nil {
			info.File = &item
		}
	} else if share.ResourceType == "folder" {
		// 使用 CloudItem 表查询文件夹
		var item entity.CloudItem
		if err := s.db.WithContext(ctx).
			Where("ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", share.ResourceID, "folder", "Y").
			First(&item).Error; err == nil {
			// 转换为 CloudFolder 格式（兼容前端）
			folder := item.ToFolder()
			info.Folder = folder
		}
	}

	// 如果需要密码验证
	needPassword := share.ShareType == "password"
	passwordProvided := password != ""
	passwordCorrect := password == share.Password

	// 如果需要密码但没有提供，或者密码错误，只返回基本信息
	if needPassword && (!passwordProvided || !passwordCorrect) {
		// 如果密码错误，返回错误
		if passwordProvided && !passwordCorrect {
			return nil, errors.New(errors.ErrInvalidParam, "访问密码错误")
		}
		// 如果没有提供密码，返回基本信息（用于显示分享页面）
		return info, nil
	}

	// 密码验证通过或公开分享
	// 更新查看次数
	s.db.WithContext(ctx).Model(&entity.CloudShare{}).
		Where("ID = ?", share.ID).
		Update("VIEW_COUNT", gorm.Expr("VIEW_COUNT + 1"))

	return info, nil
}

// GetShareByCode 获取分享记录（不验证密码）
func (s *service) GetShareByCode(ctx context.Context, shareCode string) (*entity.CloudShare, error) {
	var share entity.CloudShare
	if err := s.db.WithContext(ctx).
		Where("SHARE_CODE = ? AND IS_ACTIVE = ?", shareCode, "Y").
		First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "分享不存在或已失效")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询分享失败", err)
	}

	// 检查状态
	if share.Status != "active" {
		return nil, errors.New(errors.ErrResourceNotFound, "分享已失效")
	}

	// 检查过期时间
	if share.ExpireTime != nil && share.ExpireTime.Before(time.Now()) {
		// 更新状态为过期
		s.db.WithContext(ctx).Model(&entity.CloudShare{}).
			Where("ID = ?", share.ID).
			Update("STATUS", "expired")
		return nil, errors.New(errors.ErrResourceNotFound, "分享已过期")
	}

	return &share, nil
}

// GetUserShares 获取用户的分享列表
func (s *service) GetUserShares(ctx context.Context, userID uint) ([]*entity.CloudShare, error) {
	var shares []*entity.CloudShare
	if err := s.db.WithContext(ctx).
		Where("SHARER_ID = ? AND IS_ACTIVE = ?", userID, "Y").
		Order("CREATE_TIME DESC").
		Find(&shares).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询分享列表失败", err)
	}
	return shares, nil
}

// AccessShare 访问分享
func (s *service) AccessShare(ctx context.Context, shareCode string, password string) (*ShareInfo, error) {
	// 获取分享信息（包含验证）
	info, err := s.GetShareInfo(ctx, shareCode, password)
	if err != nil {
		return nil, err
	}

	// 检查下载次数限制
	if info.Share.MaxDownloads > 0 && info.Share.DownloadCount >= info.Share.MaxDownloads {
		return nil, errors.New(errors.ErrResourceNotFound, "分享下载次数已达上限")
	}

	// 更新下载次数
	s.db.WithContext(ctx).Model(&entity.CloudShare{}).
		Where("ID = ?", info.Share.ID).
		Update("DOWNLOAD_COUNT", gorm.Expr("DOWNLOAD_COUNT + 1"))

	return info, nil
}

// DownloadFileByID 通过文件ID下载文件（不检查权限，用于分享下载）
func (s *service) DownloadFileByID(ctx context.Context, fileID uint) (io.ReadCloser, *entity.CloudItem, error) {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}

	// StoragePath 是指针，需要解引用
	if file.StoragePath == nil {
		return nil, nil, errors.New(errors.ErrInvalidParam, "文件存储路径不存在")
	}

	// 获取公司存储实例（通过文件所有者ID获取）
	storageInstance, err := s.getCompanyStorage(ctx, file.OwnerID)
	if err != nil {
		return nil, nil, err
	}

	// 从存储中下载文件
	reader, err := storageInstance.Download(ctx, *file.StoragePath)
	if err != nil {
		return nil, nil, err
	}

	return reader, file, nil
}

// IncrementShareDownloadCount 增加分享下载次数
func (s *service) IncrementShareDownloadCount(ctx context.Context, shareID uint) error {
	return s.db.WithContext(ctx).Model(&entity.CloudShare{}).
		Where("ID = ?", shareID).
		Update("DOWNLOAD_COUNT", gorm.Expr("DOWNLOAD_COUNT + 1")).Error
}

// CancelShare 取消分享
func (s *service) CancelShare(ctx context.Context, shareID uint, userID uint) error {
	var share entity.CloudShare
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND IS_ACTIVE = ?", shareID, "Y").
		First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrResourceNotFound, "分享不存在")
		}
		return errors.Wrap(errors.ErrDatabase, "查询分享失败", err)
	}

	// 检查权限
	if share.SharerID != userID {
		return errors.New(errors.ErrPermissionDenied, "无权限取消此分享")
	}

	// 更新分享状态
	if err := s.db.WithContext(ctx).Model(&entity.CloudShare{}).
		Where("ID = ?", shareID).
		Updates(map[string]interface{}{
			"STATUS":      "disabled",
			"IS_ACTIVE":   "N",
			"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
			"UPDATE_TIME": time.Now(),
		}).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "取消分享失败", err)
	}

	return nil
}
