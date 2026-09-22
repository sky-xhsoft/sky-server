package pref

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// Service 用户环境变量服务
type Service interface {
	// Get 获取用户环境变量值(带默认值)
	Get(ctx context.Context, userID uint, module, name, def string) (string, error)
	// Set 设置用户环境变量(upsert)
	Set(ctx context.Context, userID uint, module, name, value string) error
	// Delete 删除单个环境变量
	Delete(ctx context.Context, userID uint, module, name string) error
	// GetAll 获取用户某一模块的全部变量
	GetAll(ctx context.Context, userID uint, module string) (map[string]string, error)
	// GetAllGlobal 获取全局(用户0)所有模块变量
	GetAllGlobal(ctx context.Context) ([]*entity.SysUserPref, error)
	// ReplaceVariables 替换字符串中的 $变量$ 通配(核心能力)
	ReplaceVariables(ctx context.Context, str string, userID uint) (string, error)
	// List 分页列出(供元数据/CUD)
	List(ctx context.Context, userID uint, page, pageSize int) ([]*entity.SysUserPref, int64, error)
}

type service struct {
	db       *gorm.DB
	userRepo repository.UserRepository
}

// NewService 创建用户环境变量服务
func NewService(db *gorm.DB, userRepo repository.UserRepository) Service {
	return &service{db: db, userRepo: userRepo}
}

func (s *service) Get(ctx context.Context, userID uint, module, name, def string) (string, error) {
	// 用户级优先,其次是全局(用户0)
	for _, uid := range []uint{userID, 0} {
		var p entity.SysUserPref
		ierr := s.db.WithContext(ctx).Where("SYS_USER_ID = ? AND MODULE = ? AND NAME = ? AND IS_ACTIVE = 'Y'", uid, module, name).First(&p).Error
		if ierr == nil {
			return p.Value, nil
		}
		if ierr != gorm.ErrRecordNotFound {
			return def, ierr
		}
	}
	return def, nil
}

func (s *service) Set(ctx context.Context, userID uint, module, name, value string) error {
	var p entity.SysUserPref
	err := s.db.WithContext(ctx).Where("SYS_USER_ID = ? AND MODULE = ? AND NAME = ?", userID, module, name).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		p = entity.SysUserPref{SysUserID: userID, Module: module, Name: name, Value: value}
		return s.db.WithContext(ctx).Create(&p).Error
	}
	if err != nil {
		return err
	}
	p.Value = value
	p.IsActive = "Y"
	return s.db.WithContext(ctx).Save(&p).Error
}

func (s *service) Delete(ctx context.Context, userID uint, module, name string) error {
	return s.db.WithContext(ctx).
		Where("SYS_USER_ID = ? AND MODULE = ? AND NAME = ?", userID, module, name).
		Delete(&entity.SysUserPref{}).Error
}

func (s *service) GetAll(ctx context.Context, userID uint, module string) (map[string]string, error) {
	var rows []entity.SysUserPref
	q := s.db.WithContext(ctx).Where("(SYS_USER_ID = ? OR SYS_USER_ID = 0) AND IS_ACTIVE = 'Y'", userID)
	if module != "" {
		q = q.Where("MODULE = ?", module)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, r := range rows {
		key := r.Module + "." + r.Name
		out[key] = r.Value
	}
	return out, nil
}

func (s *service) GetAllGlobal(ctx context.Context) ([]*entity.SysUserPref, error) {
	var rows []*entity.SysUserPref
	if err := s.db.WithContext(ctx).Where("SYS_USER_ID = 0 AND IS_ACTIVE = 'Y'").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// List 分页列出(用于元数据/CUD 直查,通常带表级过滤)
func (s *service) List(ctx context.Context, userID uint, page, pageSize int) ([]*entity.SysUserPref, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := s.db.WithContext(ctx).Where("IS_ACTIVE = 'Y'")
	var total int64
	if err := q.Model(&entity.SysUserPref{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []*entity.SysUserPref
	if err := q.Order("ID DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// 内置变量名
const (
	VarUser    = "user"
	VarCompany = "company"
	VarNow     = "now"
	VarNowDate = "nowDate"
)

var varRe = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)(?:\$)?`)

// ReplaceVariables 替换 $变量$ 通配(仿 Bos QueryRequestImpl.replaceVariables)
// 支持: $user$(当前用户名)/$company$(当前公司名)/$now$(YYYY-MM-DD HH:MM:SS)/$nowDate$(YYYY-MM-DD)
// 以及用户环境变量(以 "env." 前缀或任意 module.name 命名的 sys_user_pref 条目)
func (s *service) ReplaceVariables(ctx context.Context, str string, userID uint) (string, error) {
	if !strings.Contains(str, "$") {
		return str, nil
	}

	// 内置变量
	now := time.Now()
	replacer := strings.NewReplacer(
		"$user$", s.resolveUserLabel(ctx, userID),
		"$user", s.resolveUserLabel(ctx, userID),
		"$company$", s.resolveCompanyLabel(ctx, userID),
		"$company", s.resolveCompanyLabel(ctx, userID),
		"$now$", now.Format("2006-01-02 15:04:05"),
		"$now", now.Format("2006-01-02 15:04:05"),
		"$nowDate$", now.Format("2006-01-02"),
		"$nowDate", now.Format("2006-01-02"),
	)

	out := replacer.Replace(str)

	// 用户环境变量 $module.name$ (通过 sys_user_pref)
	out = varRe.ReplaceAllStringFunc(out, func(m string) string {
		inner := strings.TrimPrefix(strings.TrimSuffix(m, "$"), "$")
		// 已在内置替换中被处理的不再动
		switch inner {
		case "user", "company", "now", "nowDate":
			return m
		}
		// 解析 module.name
		if strings.Contains(inner, ".") {
			parts := strings.SplitN(inner, ".", 2)
			val, err := s.Get(ctx, userID, parts[0], parts[1], "")
			if err == nil && val != "" {
				return val
			}
		}
		return m
	})

	return out, nil
}

func (s *service) resolveUserLabel(ctx context.Context, userID uint) string {
	if userID == 0 {
		return ""
	}
	u, err := s.userRepo.GetUserByID(userID)
	if err != nil || u == nil {
		return fmt.Sprintf("%d", userID)
	}
	if u.TrueName != "" {
		return u.TrueName
	}
	if u.Username != "" {
		return u.Username
	}
	return fmt.Sprintf("%d", userID)
}

func (s *service) resolveCompanyLabel(ctx context.Context, userID uint) string {
	if userID == 0 {
		return ""
	}
	u, err := s.userRepo.GetUserByID(userID)
	if err != nil || u == nil {
		return ""
	}
	if u.SysCompanyID == 0 {
		return ""
	}
	var c entity.SysCompany
	if cerr := s.db.WithContext(ctx).Where("ID = ?", u.SysCompanyID).First(&c).Error; cerr != nil {
		return fmt.Sprintf("%d", u.SysCompanyID)
	}
	return c.Name
}

// ErrorXXX 保留 errors 引用(避免未使用)
var _ = errors.ErrInvalidParam