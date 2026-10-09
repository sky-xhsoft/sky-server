package cloud

import (
	"context"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// DefaultCompanyStorageQuota 公司未配置存储上限（MAX_STORAGE_SIZE <= 0）时的默认存储空间：100GB
const DefaultCompanyStorageQuota int64 = 100 << 30

// maxCategoryFiles 按类型列出文件的最大条数
const maxCategoryFiles = 5000

// 文件分类（与前端云盘左侧分类一致）：video / image / document（PDF、Office、文本）/ other
var (
	categoryVideoExts = setOf("mp4 avi mov wmv flv mkv webm m4v mpg mpeg ts 3gp m3u8 hls")
	categoryImageExts = setOf("jpg jpeg png gif bmp webp svg ico heic tif tiff")
	categoryDocExts   = setOf("pdf doc docx xls xlsx ppt pptx wps et dps odt ods odp rtf epub ofd xps txt md log json xml csv yaml yml")
)

func setOf(s string) map[string]bool {
	m := map[string]bool{}
	for _, v := range strings.Fields(s) {
		m[v] = true
	}
	return m
}

// FileCategoryOf 按扩展名（其次 MIME）归类文件
func FileCategoryOf(ext, mime string) string {
	e := strings.TrimPrefix(strings.ToLower(ext), ".")
	m := strings.ToLower(mime)
	switch {
	case categoryVideoExts[e] || (e == "" && strings.HasPrefix(m, "video/")):
		return "video"
	case categoryImageExts[e] || (e == "" && strings.HasPrefix(m, "image/")):
		return "image"
	case categoryDocExts[e] || (e == "" && (m == "application/pdf" || strings.HasPrefix(m, "text/"))):
		return "document"
	}
	return "other"
}

// DriveStats 用户云盘统计（全部文件，不限当前目录）
type DriveStats struct {
	All       int64 `json:"all"`
	Video     int64 `json:"video"`
	Image     int64 `json:"image"`
	Document  int64 `json:"document"`
	Other     int64 `json:"other"`
	Folders   int64 `json:"folders"`
	TotalSize int64 `json:"totalSize"`
}

// GetDriveStats 统计用户全部文件按类型的数量
func (s *service) GetDriveStats(ctx context.Context, userID uint) (*DriveStats, error) {
	var rows []struct {
		FileExt  string
		FileType string
		Cnt      int64
		Size     int64
	}
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Select("COALESCE(FILE_EXT, '') AS file_ext, COALESCE(FILE_TYPE, '') AS file_type, COUNT(*) AS cnt, COALESCE(SUM(FILE_SIZE), 0) AS size").
		Where("OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", userID, "file", "Y").
		Group("FILE_EXT, FILE_TYPE").
		Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计文件失败", err)
	}

	stats := &DriveStats{}
	for _, r := range rows {
		stats.All += r.Cnt
		stats.TotalSize += r.Size
		switch FileCategoryOf(r.FileExt, r.FileType) {
		case "video":
			stats.Video += r.Cnt
		case "image":
			stats.Image += r.Cnt
		case "document":
			stats.Document += r.Cnt
		default:
			stats.Other += r.Cnt
		}
	}
	if err := s.db.WithContext(ctx).Model(&entity.CloudItem{}).
		Where("OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", userID, "folder", "Y").
		Count(&stats.Folders).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计文件夹失败", err)
	}
	return stats, nil
}

// ListFilesByCategory 列出用户全部文件中某一类型的文件（不限目录），最近创建的在前
func (s *service) ListFilesByCategory(ctx context.Context, userID uint, category string) ([]*entity.CloudItem, error) {
	switch category {
	case "video", "image", "document", "other":
	default:
		return nil, errors.New(errors.ErrInvalidParam, "未知的文件分类")
	}
	var files []*entity.CloudItem
	if err := s.db.WithContext(ctx).
		Where("OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ?", userID, "file", "Y").
		Order("CREATE_TIME DESC").
		Find(&files).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询文件失败", err)
	}
	result := make([]*entity.CloudItem, 0, len(files))
	for _, f := range files {
		ext, mime := "", ""
		if f.FileExt != nil {
			ext = *f.FileExt
		}
		if f.FileType != nil {
			mime = *f.FileType
		}
		if FileCategoryOf(ext, mime) == category {
			result = append(result, f)
			if len(result) >= maxCategoryFiles {
				break
			}
		}
	}
	return result, nil
}

// CompanyStorage 公司存储空间（同一公司所有用户共享）
type CompanyStorage struct {
	CompanyID  uint  `json:"companyId"`
	TotalQuota int64 `json:"totalQuota"`
	UsedSpace  int64 `json:"usedSpace"`
	FileCount  int64 `json:"fileCount"`
}

// companyStorageLimit 公司存储上限：已配置取配置值，否则默认 100GB
func (s *service) companyStorageLimit(ctx context.Context, companyID uint) (int64, error) {
	var conf entity.SysCompanyConf
	err := s.db.WithContext(ctx).Select("MAX_STORAGE_SIZE").
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", companyID, "Y").First(&conf).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, errors.Wrap(errors.ErrDatabase, "查询公司配置失败", err)
	}
	if conf.MaxStorageSize > 0 {
		return conf.MaxStorageSize, nil
	}
	return DefaultCompanyStorageQuota, nil
}

// companyUsage 公司已用空间与文件数（各用户配额记录之和）
func (s *service) companyUsage(ctx context.Context, companyID uint) (int64, int64, error) {
	var usage struct {
		UsedSpace int64
		FileCount int64
	}
	if err := s.db.WithContext(ctx).Table("cloud_quota q").
		Select("COALESCE(SUM(q.USED_SPACE), 0) AS used_space, COALESCE(SUM(q.FILE_COUNT), 0) AS file_count").
		Joins("JOIN sys_user u ON u.ID = q.USER_ID").
		Where("u.SYS_COMPANY_ID = ? AND q.IS_ACTIVE = ?", companyID, "Y").
		Scan(&usage).Error; err != nil {
		return 0, 0, errors.Wrap(errors.ErrDatabase, "统计公司存储用量失败", err)
	}
	return usage.UsedSpace, usage.FileCount, nil
}

// GetCompanyStorage 用户所属公司的存储空间与用量
func (s *service) GetCompanyStorage(ctx context.Context, userID uint) (*CompanyStorage, error) {
	companyID := s.getUserCompanyID(ctx, userID)
	limit, err := s.companyStorageLimit(ctx, companyID)
	if err != nil {
		return nil, err
	}
	used, files, err := s.companyUsage(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return &CompanyStorage{CompanyID: companyID, TotalQuota: limit, UsedSpace: used, FileCount: files}, nil
}
