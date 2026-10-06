package jwt

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newBlacklistJWT(t *testing.T) (*JWT, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	cli := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return New("test-secret", WithRedis(cli, "bl:")), mr
}

func TestTokenHash(t *testing.T) {
	h := TokenHash("eyJ.token.value")
	if len(h) != 64 || strings.Contains(h, "eyJ") {
		t.Fatalf("摘要应为 64 位十六进制且不含原文, got %q", h)
	}
	if TokenHash("eyJ.token.value") != h {
		t.Fatal("同一 token 的摘要应稳定")
	}
}

func TestBlacklist_StoresHashNotToken(t *testing.T) {
	j, mr := newBlacklistJWT(t)
	token := "eyJhbGciOi.payload.sig"

	if err := j.BlacklistToken(token, time.Minute); err != nil {
		t.Fatal(err)
	}
	if !j.IsBlacklisted(token) {
		t.Fatal("加入黑名单后应被识别")
	}
	for _, k := range mr.Keys() {
		if strings.Contains(k, token) {
			t.Fatalf("Redis 中不应保存 token 原文: %s", k)
		}
	}
}

func TestBlacklist_ByHash(t *testing.T) {
	j, _ := newBlacklistJWT(t)
	token := "eyJhbGciOi.other.sig"

	if err := j.BlacklistTokenHash(TokenHash(token), time.Minute); err != nil {
		t.Fatal(err)
	}
	if !j.IsBlacklisted(token) {
		t.Fatal("按摘要加入黑名单后，用原 token 检查应命中")
	}
}

func TestBlacklist_LegacyRawKey(t *testing.T) {
	j, _ := newBlacklistJWT(t)
	token := "eyJhbGciOi.legacy.sig"

	// 升级前的黑名单记录以 token 原文作键
	if err := j.redisCli.Set(context.Background(), "bl:"+token, "1", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if !j.IsBlacklisted(token) {
		t.Fatal("升级前加入黑名单的 token 仍应被识别")
	}
	if j.IsBlacklisted("eyJhbGciOi.not.listed") {
		t.Fatal("未加入黑名单的 token 不应命中")
	}
}
