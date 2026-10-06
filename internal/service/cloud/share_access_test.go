package cloud

import (
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
)

func newPasswordShare(id uint, password string) *entity.CloudShare {
	share := &entity.CloudShare{ShareType: "password", Password: password}
	share.ID = id
	return share
}

func TestVerifyShareAccessToken(t *testing.T) {
	fieldcrypto.Setup("test-share-token-key")
	now := time.Unix(1_700_000_000, 0)
	share := newPasswordShare(25, "abcd")

	token, err := issueShareAccessToken(share, now)
	if err != nil {
		t.Fatalf("issueShareAccessToken failed: %v", err)
	}

	tests := []struct {
		name    string
		share   *entity.CloudShare
		token   string
		at      time.Time
		wantErr bool
	}{
		{"有效凭证", share, token, now.Add(time.Minute), false},
		{"临近过期仍有效", share, token, now.Add(shareAccessTokenTTL), false},
		{"已过期", share, token, now.Add(shareAccessTokenTTL + time.Second), true},
		{"未携带凭证", share, "", now, true},
		{"伪造凭证", share, "garbage-token", now, true},
		{"其他分享的凭证", newPasswordShare(26, "abcd"), token, now, true},
		{"修改密码后旧凭证失效", newPasswordShare(25, "new-pass"), token, now, true},
		{"公开分享无需凭证", &entity.CloudShare{ShareType: "public"}, "", now, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyShareAccessToken(tt.share, tt.token, tt.at)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
