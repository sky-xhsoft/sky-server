package live

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/ark"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
	"gorm.io/gorm"
)

type fakeChat struct {
	replies []string
	err     error
	calls   []string
}

func (f *fakeChat) Chat(ctx context.Context, msgs []ark.Message, t float64, max int) (string, error) {
	f.calls = append(f.calls, msgs[len(msgs)-1].Content)
	if f.err != nil {
		return "", f.err
	}
	i := len(f.calls) - 1
	if i >= len(f.replies) {
		i = len(f.replies) - 1
	}
	return f.replies[i], nil
}

func setupGlossary(t *testing.T, withArk bool) (*glossaryExtractor, *fakeChat) {
	t.Helper()
	fieldcrypto.Setup("test-glossary-key")
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.SysCompanyConf{}); err != nil {
		t.Fatal(err)
	}
	conf := entity.SysCompanyConf{SysCompanyID: 50}
	if withArk {
		conf.ArkAPIKey, conf.ArkModel = "ak-test", "ep-test"
	}
	db.Create(&conf)
	fc := &fakeChat{}
	return &glossaryExtractor{db: db, newClient: func(k, m string) GlossaryChatClient {
		if k != "ak-test" || m != "ep-test" {
			t.Fatalf("应使用公司自己的方舟配置，got %s %s", k, m)
		}
		return fc
	}}, fc
}

const glossarySrc = "尊敬的各位来宾，欢迎来到子格博主题大会。子格博今天发布基于火山引擎的声影同传。"
const glossaryTgt = "Distinguished guests, welcome to the Zigebo Keynote. Today Zigebo launches simultaneous interpretation built on Volcengine."

func TestGlossaryExtract_FiltersAndDedupes(t *testing.T) {
	g, fc := setupGlossary(t, true)
	fc.replies = []string{"好的，结果如下：\n```json\n" + `{"keywords":[
		{"source":"子格博","target":"Zigebo","type":"品牌"},
		{"source":"火山引擎","target":"Volcengine","type":"品牌"},
		{"source":"子格博","target":"Zigebo","type":"品牌"},
		{"source":"不存在的词","target":"Ghost","type":"术语"},
		{"source":"主题大会","target":"Conference","type":"术语"}],
	 "typos":[{"wrong":"Zigbee","correct":"Zigebo","reason":"近音"},{"wrong":"Volcano Engine","correct":"Volcengine","reason":"直译"},
		{"wrong":"abc","correct":"NotInText","reason":"其他"},{"wrong":"Zigebo","correct":"Zigebo"}]}` + "\n```"}
	res, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: glossarySrc, TargetText: glossaryTgt, SourceLang: "zh", TargetLang: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Keywords) != 2 || res.Keywords[0].Source != "子格博" || res.Keywords[0].Count != 2 {
		t.Fatalf("应过滤臆造词与去重，并按出现次数排序：%+v", res.Keywords)
	}
	if len(res.Typos) != 2 {
		t.Fatalf("易错词应只保留正确写法出现在译文中、且易错与纠正不同的：%+v", res.Typos)
	}
	if !strings.Contains(fc.calls[0], "【译文】") {
		t.Fatal("请求应同时包含原文与译文")
	}
}

func TestGlossaryExtract_LongTextIsChunked(t *testing.T) {
	g, fc := setupGlossary(t, true)
	fc.replies = []string{`{"keywords":[{"source":"子格博","target":"Zigebo","type":"品牌"}],"typos":[]}`}
	src := strings.Repeat(glossarySrc+"\n", 220) // 约 9 千字，分 2 段
	tgt := strings.Repeat(glossaryTgt+"\n", 220)
	res, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: src, TargetText: tgt, SourceLang: "zh", TargetLang: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Chunks < 2 || len(fc.calls) != res.Chunks || len(res.Keywords) != 1 {
		t.Fatalf("长文应分段调用并合并去重：chunks=%d calls=%d kw=%d", res.Chunks, len(fc.calls), len(res.Keywords))
	}
}

func TestGlossaryExtract_Errors(t *testing.T) {
	g, _ := setupGlossary(t, false)
	if _, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: "a", TargetText: "b"}); !errors.Is(err, ErrNoArkConfig) {
		t.Fatalf("未配置方舟应提示，got %v", err)
	}
	g, fc := setupGlossary(t, true)
	if _, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: "a"}); err == nil {
		t.Fatal("同传模式缺少译文应报错")
	}
	if _, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: strings.Repeat("字", 30001), TargetText: "b"}); err == nil {
		t.Fatal("超过 3 万字应报错")
	}
	fc.err = &ark.Error{Status: 401, Code: "AuthenticationError", Message: "bad key"}
	_, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: glossarySrc, TargetText: glossaryTgt})
	if err == nil || !strings.Contains(err.Error(), "API Key 无效") {
		t.Fatalf("鉴权失败应给出中文提示，got %v", err)
	}
	fc.err = nil
	fc.replies = []string{"抱歉，我无法处理"}
	if _, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: glossarySrc, TargetText: glossaryTgt}); err == nil {
		t.Fatal("模型返回非 JSON 应报错")
	}
}

func TestGlossaryExtract_ASRSourceOnly(t *testing.T) {
	g, fc := setupGlossary(t, true)
	fc.replies = []string{`{"keywords":[],"typos":[{"wrong":"紫格博","correct":"子格博","reason":"近音"}]}`}
	res, err := g.Extract(context.Background(), 50, &GlossaryExtractRequest{SourceText: glossarySrc, SourceLang: "zh", TypoSide: "source"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Typos) != 1 || strings.Contains(fc.calls[0], "【译文】") {
		t.Fatalf("原文 ASR 只需原文，易错词按原文校验：%+v", res)
	}
}

func TestArkClient_Chat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" || r.Header.Get("Authorization") != "Bearer ak-test" {
			w.WriteHeader(401)
			_, _ = w.Write([]byte(`{"error":{"code":"AuthenticationError","message":"bad"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"keywords\":[]}"}}]}`))
	}))
	defer srv.Close()
	c := ark.NewClient("ak-test", "ep-test")
	c.BaseURL = srv.URL
	out, err := c.Chat(context.Background(), []ark.Message{{Role: "user", Content: "hi"}}, 0.1, 100)
	if err != nil || out != `{"keywords":[]}` {
		t.Fatalf("got %q %v", out, err)
	}
	c.APIKey = "wrong"
	_, err = c.Chat(context.Background(), []ark.Message{{Role: "user", Content: "hi"}}, 0.1, 100)
	var ae *ark.Error
	if !errors.As(err, &ae) || ae.Status != 401 || ae.Code != "AuthenticationError" {
		t.Fatalf("应返回方舟错误，got %v", err)
	}
}
