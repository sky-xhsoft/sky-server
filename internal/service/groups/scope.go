package groups

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// 权限组的公司范围：context 中带公司ID（来自接口请求）时，
// 只能读取本公司和公共（0/-1）权限组，只能修改本公司权限组，只能给本公司用户分配；
// 不带公司ID（内部调用）时不限制。

func scopeCompany(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(contextutil.CompanyIDKey).(uint)
	return id, ok && id > 0
}

func isSharedCompany(id uint) bool {
	return id == 0 || id == ^uint(0)
}

// checkGroupScope 校验权限组是否在当前公司范围内；write 为 true 时公共权限组也不允许修改
func checkGroupScope(ctx context.Context, group *entity.SysGroups, write bool) error {
	company, ok := scopeCompany(ctx)
	if !ok || group.SysCompanyID == company {
		return nil
	}
	if isSharedCompany(group.SysCompanyID) {
		if write {
			return errors.New(errors.ErrPermissionDenied, "公共权限组不允许修改")
		}
		return nil
	}
	return errors.New(errors.ErrResourceNotFound, "权限组不存在")
}

// checkUserScope 校验目标用户属于当前公司
func (s *service) checkUserScope(ctx context.Context, userID uint) error {
	company, ok := scopeCompany(ctx)
	if !ok {
		return nil
	}
	var userCompany uint
	res := s.db.WithContext(ctx).Table("sys_user").Select("SYS_COMPANY_ID").Where("ID = ?", userID).Limit(1).Scan(&userCompany)
	if res.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "查询用户失败", res.Error)
	}
	if res.RowsAffected == 0 || userCompany != company {
		return errors.New(errors.ErrResourceNotFound, "用户不存在")
	}
	return nil
}

// checkGroupsReadable 校验一组权限组都可被当前公司使用（本公司或公共）
func (s *service) checkGroupsReadable(ctx context.Context, groupIDs []uint) error {
	for _, id := range groupIDs {
		if _, err := s.GetGroup(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
