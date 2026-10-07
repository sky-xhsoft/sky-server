package cloud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// checkCompanyQuota 检查用户所属公司的配额（sys_company_conf 的 MAX_* 字段，单位字节，0 表示不限制）
func (s *service) checkCompanyQuota(ctx context.Context, userID uint, fileSize int64) error {
	var companyIDs []uint
	if err := s.db.WithContext(ctx).Table("sys_user").Where("ID = ?", userID).
		Pluck("SYS_COMPANY_ID", &companyIDs).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询用户公司失败", err)
	}
	if len(companyIDs) == 0 || companyIDs[0] == 0 {
		return nil
	}
	companyID := companyIDs[0]

	var conf entity.SysCompanyConf
	err := s.db.WithContext(ctx).Select("MAX_STORAGE_SIZE, MAX_FILE_SIZE, MAX_FILES").
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", companyID, "Y").First(&conf).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询公司配置失败", err)
	}

	if conf.MaxFileSize > 0 && fileSize > conf.MaxFileSize {
		return errors.New(errors.ErrInvalidParam, fmt.Sprintf("文件大小超过公司限制（最大%s）", formatBytes(conf.MaxFileSize)))
	}
	if conf.MaxStorageSize <= 0 && conf.MaxFiles <= 0 {
		return nil
	}

	var usage struct {
		UsedSpace int64
		FileCount int64
	}
	if err := s.db.WithContext(ctx).Table("cloud_quota q").
		Select("COALESCE(SUM(q.USED_SPACE), 0) AS used_space, COALESCE(SUM(q.FILE_COUNT), 0) AS file_count").
		Joins("JOIN sys_user u ON u.ID = q.USER_ID").
		Where("u.SYS_COMPANY_ID = ? AND q.IS_ACTIVE = ?", companyID, "Y").
		Scan(&usage).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "统计公司存储用量失败", err)
	}
	if conf.MaxStorageSize > 0 && usage.UsedSpace+fileSize > conf.MaxStorageSize {
		return errors.New(errors.ErrResourceConflict, fmt.Sprintf("公司存储空间不足（上限%s）", formatBytes(conf.MaxStorageSize)))
	}
	if conf.MaxFiles > 0 && usage.FileCount+1 > int64(conf.MaxFiles) {
		return errors.New(errors.ErrResourceConflict, fmt.Sprintf("公司文件数已达上限（%d个）", conf.MaxFiles))
	}
	return nil
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}
