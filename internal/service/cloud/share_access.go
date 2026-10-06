package cloud

import (
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

// issueShareAccessToken 密码验证通过后签发访问凭证（载荷：分享ID、过期时间、密码指纹）
func issueShareAccessToken(share *entity.CloudShare, now time.Time) (string, error) {
	payload := fmt.Sprintf("%s|%d|%d|%s", shareAccessTokenPrefix, share.ID,
		now.Add(shareAccessTokenTTL).Unix(), shareAccessFingerprint(share))
	return fieldcrypto.SealToken(payload)
}

// VerifyShareAccess 校验访问加密分享内容（下载/预览/浏览文件夹）所需的凭证；非加密分享直接放行
func (s *service) VerifyShareAccess(share *entity.CloudShare, token string) error {
	return verifyShareAccessToken(share, token, time.Now())
}

func verifyShareAccessToken(share *entity.CloudShare, token string, now time.Time) error {
	if share.ShareType != "password" {
		return nil
	}
	if token == "" {
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
