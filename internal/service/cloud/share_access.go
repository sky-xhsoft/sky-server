package cloud

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
)

// shareAccessTokenTTL 加密分享访问凭证有效期
const shareAccessTokenTTL = 2 * time.Hour

const shareAccessTokenPrefix = "share"

// shareAccessFingerprint 当前访问密码的指纹；修改密码后旧凭证随即失效
func shareAccessFingerprint(share *entity.CloudShare) string {
	sum := sha256.Sum256([]byte(share.Password))
	return hex.EncodeToString(sum[:8])
}

// issueShareAccessToken 加密分享密码验证通过、或私密分享访问者校验通过后签发访问凭证（载荷：分享ID、过期时间、密码指纹）
func issueShareAccessToken(share *entity.CloudShare, now time.Time) (string, error) {
	payload := fmt.Sprintf("%s|%d|%d|%s", shareAccessTokenPrefix, share.ID,
		now.Add(shareAccessTokenTTL).Unix(), shareAccessFingerprint(share))
	return fieldcrypto.SealToken(payload)
}

// VerifyShareAccess 校验访问加密/私密分享内容（下载/预览/浏览文件夹）所需的凭证；公开分享直接放行
func (s *service) VerifyShareAccess(share *entity.CloudShare, token string) error {
	return verifyShareAccessToken(share, token, time.Now())
}

func verifyShareAccessToken(share *entity.CloudShare, token string, now time.Time) error {
	if !shareNeedsAccessToken(share) {
		return nil
	}
	if token == "" {
		if share.ShareType == "private" {
			return errors.New(errors.ErrUnauthorized, "访问凭证缺失，请重新打开分享页面")
		}
		return errors.New(errors.ErrUnauthorized, "请先输入访问密码")
	}

	invalid := errors.New(errors.ErrUnauthorized, "访问凭证无效，请重新输入访问密码")
	payload, err := fieldcrypto.OpenToken(token)
	if err != nil {
		return invalid
	}
	parts := strings.Split(payload, "|")
	if len(parts) != 4 || parts[0] != shareAccessTokenPrefix {
		return invalid
	}
	shareID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || uint(shareID) != share.ID {
		return invalid
	}
	expireAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return invalid
	}
	if subtle.ConstantTimeCompare([]byte(parts[3]), []byte(shareAccessFingerprint(share))) != 1 {
		return invalid
	}
	if now.Unix() > expireAt {
		return errors.New(errors.ErrUnauthorized, "访问凭证已过期，请重新输入访问密码")
	}
	return nil
}

// shareNeedsAccessToken 加密分享与私密分享的内容访问都需要访问凭证
func shareNeedsAccessToken(share *entity.CloudShare) bool {
	return share.ShareType == "password" || share.ShareType == "private"
}

// VerifyShareViewer 私密分享仅限分享者所在公司的登录用户访问；viewerCompanyID 为 0 表示未登录
func (s *service) VerifyShareViewer(ctx context.Context, share *entity.CloudShare, viewerCompanyID uint) error {
	if share.ShareType != "private" {
		return nil
	}
	if viewerCompanyID == 0 {
		return errors.New(errors.ErrUnauthorized, "该分享仅限分享者所在公司的成员访问，请先登录")
	}
	var sharerCompanyID uint
	if err := s.db.WithContext(ctx).Raw("SELECT SYS_COMPANY_ID FROM sys_user WHERE ID = ?", share.SharerID).Scan(&sharerCompanyID).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询分享者失败", err)
	}
	if sharerCompanyID == 0 || sharerCompanyID != viewerCompanyID {
		return errors.New(errors.ErrUnauthorized, "该分享仅限分享者所在公司的成员访问")
	}
	return nil
}

// maxShareDepth 校验分享范围时向上查找的最大层级（防止异常数据形成环导致死循环）
const maxShareDepth = 64

// VerifyShareItem 校验要访问的条目位于分享范围内：
// 文件分享只能访问被分享的文件本身；文件夹分享只能访问该文件夹及其子孙条目，且须属于分享者、未删除
func (s *service) VerifyShareItem(ctx context.Context, share *entity.CloudShare, itemID uint) error {
	if itemID == share.ResourceID {
		return nil
	}
	notInShare := errors.New(errors.ErrResourceNotFound, "文件不在该分享中")
	if share.ResourceType != "folder" {
		return notInShare
	}
	current := itemID
	for depth := 0; depth < maxShareDepth; depth++ {
		var item entity.CloudItem
		if err := s.db.WithContext(ctx).Select("ID, PARENT_ID, OWNER_ID").
			Where("ID = ? AND IS_ACTIVE = ?", current, "Y").First(&item).Error; err != nil {
			return notInShare
		}
		if item.OwnerID != share.SharerID || item.ParentID == nil {
			return notInShare
		}
		if *item.ParentID == share.ResourceID {
			return nil
		}
		current = *item.ParentID
	}
	return notInShare
}
