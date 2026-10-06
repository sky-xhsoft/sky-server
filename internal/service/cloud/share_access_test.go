package cloud

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
	"gorm.io/gorm"
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

func TestVerifyShareAccessToken_Private(t *testing.T) {
	fieldcrypto.Setup("test-share-token-key")
	now := time.Unix(1_700_000_000, 0)
	share := &entity.CloudShare{ShareType: "private"}
	share.ID = 30

	if err := verifyShareAccessToken(share, "", now); err == nil {
		t.Error("私密分享未携带凭证应被拒绝")
	}
	token, err := issueShareAccessToken(share, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyShareAccessToken(share, token, now.Add(time.Minute)); err != nil {
		t.Errorf("私密分享携带有效凭证应放行: %v", err)
	}
}

func TestVerifyShareViewer(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER)`)
	db.Exec(`INSERT INTO sys_user (ID, SYS_COMPANY_ID) VALUES (1, 10), (2, 0)`)
	svc := &service{db: db}

	private := &entity.CloudShare{ShareType: "private", SharerID: 1}
	orphan := &entity.CloudShare{ShareType: "private", SharerID: 2}
	public := &entity.CloudShare{ShareType: "public", SharerID: 1}

	tests := []struct {
		name    string
		share   *entity.CloudShare
		viewer  uint
		wantErr bool
	}{
		{"同公司登录用户", private, 10, false},
		{"其他公司用户", private, 11, true},
		{"未登录", private, 0, true},
		{"分享者无所属公司", orphan, 10, true},
		{"公开分享无需校验", public, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.VerifyShareViewer(context.Background(), tt.share, tt.viewer)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
