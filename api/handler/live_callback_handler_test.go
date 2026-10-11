package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
)

// TestMain 初始化全局 logger（verifySign 过期分支会调用 logger.Warn，未初始化会 nil panic）
func TestMain(m *testing.M) {
	if _, err := logger.Init(&config.LogConfig{Level: "error"}); err != nil {
		panic("init logger: " + err.Error())
	}
	os.Exit(m.Run())
}

// signCallback 复刻 handler 的验签算法，用于构造合法/非法签名：MD5(key + t)
func signCallback(key string, t int64) string {
	expected := fmt.Sprintf("%s%d", key, t)
	hash := md5.Sum([]byte(expected))
	return hex.EncodeToString(hash[:])
}

func newTestCallbackKey(key string) *LiveCallbackHandler {
	return &LiveCallbackHandler{callbackKey: key}
}

// TestVerifySign_Valid_FutureExpiry t 为未来到期时间 + 正确签名 → 应通过
func TestVerifySign_Valid_FutureExpiry(t *testing.T) {
	h := newTestCallbackKey("secret-key")
	tExpiry := time.Now().Unix() + 300 // 5分钟后到期
	sign := signCallback("secret-key", tExpiry)
	if !h.verifySign(tExpiry, sign) {
		t.Fatalf("合法签名应通过: t=%d", tExpiry)
	}
}

// TestVerifySign_Expired t 已过期(过去时间戳) + 正确签名 → 应拒绝
func TestVerifySign_Expired(t *testing.T) {
	h := newTestCallbackKey("secret-key")
	tPast := time.Now().Unix() - 1
	sign := signCallback("secret-key", tPast)
	if h.verifySign(tPast, sign) {
		t.Fatalf("已过期签名应被拒绝: t=%d", tPast)
	}
}

// TestVerifySign_WrongSign 正确时间 + 错误签名 → 应拒绝
func TestVerifySign_WrongSign(t *testing.T) {
	h := newTestCallbackKey("secret-key")
	tExpiry := time.Now().Unix() + 300
	if h.verifySign(tExpiry, "deadbeef") {
		t.Fatalf("错误签名应被拒绝")
	}
}

// TestVerifySign_WrongKey 用错误密钥签的同一时间 → 应拒绝
func TestVerifySign_WrongKey(t *testing.T) {
	h := newTestCallbackKey("real-key")
	tExpiry := time.Now().Unix() + 300
	sign := signCallback("wrong-key", tExpiry)
	if h.verifySign(tExpiry, sign) {
		t.Fatalf("错误密钥签名应被拒绝")
	}
}

// TestVerifySign_FarFuture_Rejected 超出现效窗口(+24h)的 t → 应拒绝(有效期内上界强制生效)
func TestVerifySign_FarFuture_Rejected(t *testing.T) {
	h := newTestCallbackKey("secret-key")
	tFarFuture := time.Now().Unix() + 24*3600 // 24小时后,远超10分钟窗口
	sign := signCallback("secret-key", tFarFuture)
	if h.verifySign(tFarFuture, sign) {
		t.Fatalf("超出10分钟窗口的远期签名应被拒绝")
	}
}

// TestVerifySign_WithinWindow_Accepted 未来估算在窗口边缘(如 +5分钟)正确签名 → 应通过
func TestVerifySign_WithinWindow_Accepted(t *testing.T) {
	h := newTestCallbackKey("secret-key")
	tWithin := time.Now().Unix() + 300 // 5分钟后,在10分钟窗口内
	sign := signCallback("secret-key", tWithin)
	if !h.verifySign(tWithin, sign) {
		t.Fatalf("窗口内合法签名应通过")
	}
}

// 腾讯云断流回调的 push_duration 为字符串、errcode 为数字；两种写法都要能解析
func TestDisconnectStreamEventFlexibleNumbers(t *testing.T) {
	cases := []struct {
		body     string
		duration int64
		reason   string
	}{
		{`{"app":"push.example.com","appname":"live","stream_id":"s1","push_duration":"51196","errcode":1,"errmsg":"推流断开"}`, 51, "1 推流断开"},
		{`{"stream_id":"s1","push_duration":61000,"errcode":"0","errmsg":""}`, 61, ""},
		{`{"stream_id":"s1","push_duration":"","duration":30}`, 30, ""},
	}
	for _, c := range cases {
		var e DisconnectStreamEvent
		if err := json.Unmarshal([]byte(c.body), &e); err != nil {
			t.Fatalf("解析 %s 失败: %v", c.body, err)
		}
		e.normalize()
		if e.Duration != c.duration || e.Reason != c.reason || e.StreamName != "s1" {
			t.Fatalf("%s: duration=%d reason=%q stream=%q", c.body, e.Duration, e.Reason, e.StreamName)
		}
	}
	var e DisconnectStreamEvent
	if err := json.Unmarshal([]byte(`{"push_duration":"abc"}`), &e); err == nil {
		t.Fatal("非数字的 push_duration 应报错")
	}
}
