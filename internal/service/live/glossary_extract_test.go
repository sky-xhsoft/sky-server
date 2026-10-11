package live

import (
	"context"
	"errors"
	"io"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/pkg/volcengine"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/pkg/ark"
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
	fc := &fakeChat{}
	g := &glossaryExtractor{newClient: func(k, m string) GlossaryChatClient {
		if k != "ak-test" || m != "ep-test" {
			t.Fatalf("应使用全局火山方舟配置，got %s %s", k, m)
		}
		return fc
	}}
	if withArk {
		g.apiKey, g.model = "ak-test", "ep-test"
	}
	return g, fc
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
	// 默认关闭深度思考；模型不支持 thinking 参数时去掉后重试
	var bodies []string
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if strings.Contains(string(b), `"thinking"`) {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":{"code":"InvalidParameter","message":"The parameter thinking is not supported"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv2.Close()
	c2 := ark.NewClient("ak-test", "ep-test")
	c2.BaseURL = srv2.URL
	if out, err := c2.Chat(context.Background(), []ark.Message{{Role: "user", Content: "hi"}}, 0.1, 100); err != nil || out != "ok" || len(bodies) != 2 || !strings.Contains(bodies[0], `"thinking":{"type":"disabled"}`) {
		t.Fatalf("应先关闭深度思考，不支持时去掉参数重试：%q %v %v", out, err, bodies)
	}

	c.APIKey = "wrong"
	_, err = c.Chat(context.Background(), []ark.Message{{Role: "user", Content: "hi"}}, 0.1, 100)
	var ae *ark.Error
	if !errors.As(err, &ae) || ae.Status != 401 || ae.Code != "AuthenticationError" {
		t.Fatalf("应返回方舟错误，got %v", err)
	}
}

// 所有公司共用一个火山账号：不能删除其他公司的同传任务
func TestDeleteSpeechTask_OnlyOwnCompany(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&SpeechTaskRecord{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&SpeechTaskRecord{CompanyID: 51, TaskID: "task-of-51", Name: "x"})
	svc := NewSpeechTaskService(nil, db)
	err = svc.DeleteSpeechTask(context.Background(), 50, &volcengine.DeleteSpeechTaskRequest{TaskID: "task-of-51"})
	if err == nil || !strings.Contains(err.Error(), "不属于本公司") {
		t.Fatalf("应拒绝删除其他公司的任务，got %v", err)
	}
	if _, err := svc.CreateSpeechTask(context.Background(), 50, &volcengine.CreateSpeechTaskRequest{Name: "x"}); err == nil || !strings.Contains(err.Error(), "未配置火山引擎") {
		t.Fatalf("未配置全局火山密钥应提示，got %v", err)
	}
}
