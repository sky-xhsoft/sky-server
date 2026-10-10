package cloud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// checkCompanyQuota 检查用户所属公司的配额（sys_company_conf 的 MAX_* 字段，单位字节）。
// 存储空间未配置（0）时默认 100GB；单文件大小为 0 表示不限制；文件数不做限制（MAX_FILES 不生效）
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
	err := s.db.WithContext(ctx).Select("MAX_STORAGE_SIZE, MAX_FILE_SIZE").
		Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = ?", companyID, "Y").First(&conf).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.Wrap(errors.ErrDatabase, "查询公司配置失败", err)
	}

	if conf.MaxFileSize > 0 && fileSize > conf.MaxFileSize {
		return errors.New(errors.ErrInvalidParam, fmt.Sprintf("文件大小超过公司限制（最大%s）", formatBytes(conf.MaxFileSize)))
	}
	limit := conf.MaxStorageSize
	if limit <= 0 {
		limit = DefaultCompanyStorageQuota
	}

	usedSpace, _, err := s.companyUsage(ctx, companyID)
	if err != nil {
		return err
	}
	if usedSpace+fileSize > limit {
		return errors.New(errors.ErrResourceConflict, fmt.Sprintf("公司存储空间不足（上限%s）", formatBytes(limit)))
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
