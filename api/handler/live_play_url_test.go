package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
)

// playAuthLiveService 播放鉴权：enable 控制域名是否开启播放鉴权
type playAuthLiveService struct {
	live.Service
	enable int64
}

func (s *playAuthLiveService) GetDomainAuthKey(ctx context.Context, domainName, keyType string) (*tencentLive.AuthKeyInfo, error) {
	return &tencentLive.AuthKeyInfo{Enable: s.enable, MasterKey: "play-key", DomainName: domainName}, nil
}

func playURLFor(t *testing.T, enable int64, body string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewLivePushURLHandler(live.NewPushURLService(), &playAuthLiveService{enable: enable}, nil)
	r := gin.New()
	r.POST("/play-url/generate", func(c *gin.Context) { c.Set("companyID", uint(52)); h.GeneratePlayURL(c) })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/play-url/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var out struct {
		Data map[string]string `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out.Data["flv"]
}

func TestGeneratePlayURL_PlayAuth(t *testing.T) {
	body := `{"playDomain":"live.skyzhou.cn","appName":"live","streamName":"101","expireTime":1893456000}`
	// 预览页不传播放密钥：域名开启播放鉴权时用本公司的鉴权主 KEY 签名
	if u := playURLFor(t, 1, body); !strings.Contains(u, "txSecret=") || !strings.HasPrefix(u, "http://live.skyzhou.cn/live/101.flv") {
		t.Fatalf("开启鉴权应带签名：%s", u)
	}
	// 未开启鉴权：不加签名
	if u := playURLFor(t, 0, body); strings.Contains(u, "txSecret=") {
		t.Fatalf("未开启鉴权不应带签名：%s", u)
	}
	// 显式传入的密钥优先
	withKey := `{"playDomain":"live.skyzhou.cn","appName":"live","streamName":"101","playKey":"own","expireTime":1893456000}`
	a, b := playURLFor(t, 1, withKey), playURLFor(t, 1, body)
	if a == b || !strings.Contains(a, "txSecret=") {
		t.Fatalf("显式密钥应优先：%s / %s", a, b)
	}
}
