package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/ark"
	"github.com/sky-xhsoft/sky-server/internal/pkg/vendortext"
)

// 整篇提取的限制
const (
	glossaryMaxTextRunes   = 30000 // 原文、译文各自上限（与关键词库上限一致）
	glossaryChunkRunes     = 5000  // 单次交给模型的原文长度，超过则分段
	glossaryMaxTermRunes   = 50    // 单个词不超过 50 个字（易错词限制）
	glossaryMaxKeywords    = 300
	glossaryMaxTypos       = 100
	glossaryExtractTimeout = 3 * time.Minute
)

// ErrNoArkConfig 系统未配置火山方舟
var ErrNoArkConfig = errors.New("系统未配置智能服务（API Key 与推理接入点），请联系平台管理员")

var glossaryLangNames = map[string]string{
	"zh": "中文", "en": "英语", "de": "德语", "es": "西班牙语", "fr": "法语", "id": "印尼语", "ja": "日语", "pt": "葡萄牙语",
}

// GlossaryExtractRequest 整篇原文、译文提取词库
type GlossaryExtractRequest struct {
	SourceText string `json:"sourceText"`
	TargetText string `json:"targetText"`
	SourceLang string `json:"sourceLang"`
	TargetLang string `json:"targetLang"`
	TypoSide   string `json:"typoSide"` // target（默认：同传纠正译文）/ source（原文 ASR 纠正原文）
}

// GlossaryKeyword 关键词：原文 → 译文
type GlossaryKeyword struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Count  int    `json:"count"` // 在原文中出现的次数
}

// GlossaryTypo 易错词：易错写法 → 纠正
type GlossaryTypo struct {
	Wrong   string `json:"wrong"`
	Correct string `json:"correct"`
	Reason  string `json:"reason"`
}

// GlossaryExtractResult 提取结果
type GlossaryExtractResult struct {
	Keywords  []GlossaryKeyword `json:"keywords"`
	Typos     []GlossaryTypo    `json:"typos"`
	Chunks    int               `json:"chunks"`
	Failed    int               `json:"failed"` // 分段中失败的段数
	ElapsedMs int64             `json:"elapsedMs"`
}

// GlossaryChatClient 大模型对话（*ark.Client 实现；测试中可替换）
type GlossaryChatClient interface {
	Chat(ctx context.Context, messages []ark.Message, temperature float64, maxTokens int) (string, error)
}

// GlossaryExtractor 字幕同传词库智能提取：服务端调用火山方舟大模型切词对齐
type GlossaryExtractor interface {
	Extract(ctx context.Context, companyID uint, req *GlossaryExtractRequest) (*GlossaryExtractResult, error)
}

type glossaryExtractor struct {
	apiKey    string
	model     string
	newClient func(apiKey, model string) GlossaryChatClient
}

// NewGlossaryExtractor 创建词库提取服务：火山方舟为全局统一配置（配置文件 volcengine.ark）
func NewGlossaryExtractor(cfg config.VolcengineArkConfig) GlossaryExtractor {
	return &glossaryExtractor{apiKey: strings.TrimSpace(cfg.APIKey), model: strings.TrimSpace(cfg.Model),
		newClient: func(k, m string) GlossaryChatClient { return ark.NewClient(k, m) }}
}

func (g *glossaryExtractor) Extract(ctx context.Context, companyID uint, req *GlossaryExtractRequest) (*GlossaryExtractResult, error) {
	start := time.Now()
	src, tgt := strings.TrimSpace(req.SourceText), strings.TrimSpace(req.TargetText)
	if req.TypoSide != "source" {
		req.TypoSide = "target"
	}
	if src == "" {
		return nil, errors.New("请粘贴原文全文")
	}
	if tgt == "" && req.TypoSide == "target" {
		return nil, errors.New("请粘贴译文全文")
	}
	if utf8.RuneCountInString(src) > glossaryMaxTextRunes || utf8.RuneCountInString(tgt) > glossaryMaxTextRunes {
		return nil, fmt.Errorf("原文、译文各不能超过 %d 个字符", glossaryMaxTextRunes)
	}

	if g.apiKey == "" || g.model == "" {
		return nil, ErrNoArkConfig
	}
	client := g.newClient(g.apiKey, g.model)

	ctx, cancel := context.WithTimeout(ctx, glossaryExtractTimeout)
	defer cancel()

	srcChunks, tgtChunks := splitAligned(src, tgt)
	res := &GlossaryExtractResult{Chunks: len(srcChunks)}
	kwSeen := map[string]int{}
	typoSeen := map[string]bool{}
	var lastErr error
	for i := range srcChunks {
		out, err := g.extractChunk(ctx, client, req, srcChunks[i], tgtChunks[i])
		if err != nil {
			res.Failed++
			lastErr = err
			if ctx.Err() != nil {
				break
			}
			continue
		}
		for _, k := range out.Keywords {
			k.Source, k.Target, k.Type = cleanTerm(k.Source), cleanTerm(k.Target), strings.TrimSpace(k.Type)
			if k.Source == "" || k.Target == "" || !containsFold(src, k.Source) || (tgt != "" && !containsFold(tgt, k.Target)) {
				continue // 原文词、译文词须真实出现在文章中，过滤模型臆造
			}
			key := strings.ToLower(k.Source)
			if _, ok := kwSeen[key]; ok {
				continue
			}
			if k.Type == "" {
				k.Type = "术语"
			}
			k.Count = countFold(src, k.Source)
			kwSeen[key] = len(res.Keywords)
			res.Keywords = append(res.Keywords, k)
		}
		side := tgt
		if req.TypoSide == "source" {
			side = src
		}
		for _, t := range out.Typos {
			t.Wrong, t.Correct, t.Reason = cleanTerm(t.Wrong), cleanTerm(t.Correct), strings.TrimSpace(t.Reason)
			if t.Wrong == "" || t.Correct == "" || t.Wrong == t.Correct || !containsFold(side, t.Correct) {
				continue // 正确写法须出现在文章中
			}
			key := strings.ToLower(t.Wrong)
			if typoSeen[key] {
				continue
			}
			if t.Reason == "" {
				t.Reason = "其他"
			}
			typoSeen[key] = true
			res.Typos = append(res.Typos, t)
		}
	}
	if res.Failed == res.Chunks && lastErr != nil {
		return nil, glossaryError(lastErr)
	}
	sort.SliceStable(res.Keywords, func(i, j int) bool { return res.Keywords[i].Count > res.Keywords[j].Count })
	if len(res.Keywords) > glossaryMaxKeywords {
		res.Keywords = res.Keywords[:glossaryMaxKeywords]
	}
	if len(res.Typos) > glossaryMaxTypos {
		res.Typos = res.Typos[:glossaryMaxTypos]
	}
	if res.Keywords == nil {
		res.Keywords = []GlossaryKeyword{}
	}
	if res.Typos == nil {
		res.Typos = []GlossaryTypo{}
	}
	res.ElapsedMs = time.Since(start).Milliseconds()
	return res, nil
}

type glossaryModelOutput struct {
	Keywords []GlossaryKeyword `json:"keywords"`
	Typos    []GlossaryTypo    `json:"typos"`
}

func langName(code string) string {
	if n, ok := glossaryLangNames[code]; ok {
		return n
	}
	if code == "" {
		return "原文语言"
	}
	return code
}

func (g *glossaryExtractor) extractChunk(ctx context.Context, client GlossaryChatClient, req *GlossaryExtractRequest, src, tgt string) (*glossaryModelOutput, error) {
	srcName, tgtName := langName(req.SourceLang), langName(req.TargetLang)
	typoLang, typoText := tgtName+"译文", "译文"
	if req.TypoSide == "source" {
		typoLang, typoText = srcName+"原文（语音识别结果）", "原文"
	}
	system := "你是专业的同声传译术语专家，负责为直播同传准备词库。用户会提供" + srcName + "原文" +
		map[bool]string{true: "和对应的" + tgtName + "译文", false: ""}[tgt != ""] + "。请完成两件事：\n" +
		"1. keywords：抽取对翻译准确性重要的专有名词与术语（品牌、产品、人名、机构、地名、专业术语），以及反复出现且译法固定的词，给出原文词与译文中对应的译法。" +
		"source 必须原样出现在原文中，target 必须原样出现在译文中；不要抽取普通常用词。\n" +
		"2. typos：在" + typoLang + "中，语音识别或机器翻译容易把哪些词写错（近音、直译、拼写、大小写），给出“易错写法 → 正确写法”。correct 必须原样出现在" + typoText + "中，wrong 与 correct 不能相同。\n" +
		"每个词不超过 50 个字；keywords 最多 60 条，typos 最多 30 条。只输出 JSON，不要任何解释或代码块标记，格式：\n" +
		`{"keywords":[{"source":"","target":"","type":"品牌|产品|人名|机构|地名|术语|通用词"}],"typos":[{"wrong":"","correct":"","reason":"近音|直译|拼写|大小写|其他"}]}`
	user := "【原文】\n" + src
	if tgt != "" {
		user += "\n\n【译文】\n" + tgt
	}
	content, err := client.Chat(ctx, []ark.Message{{Role: "system", Content: system}, {Role: "user", Content: user}}, 0.1, 4096)
	if err != nil {
		return nil, err
	}
	return parseGlossaryJSON(content)
}

// parseGlossaryJSON 从模型回复中取出 JSON（兼容代码块、前后说明文字）
func parseGlossaryJSON(content string) (*glossaryModelOutput, error) {
	i, j := strings.Index(content, "{"), strings.LastIndex(content, "}")
	if i < 0 || j <= i {
		return nil, errors.New("模型返回的内容不是 JSON")
	}
	var out glossaryModelOutput
	if err := json.Unmarshal([]byte(content[i:j+1]), &out); err != nil {
		return nil, fmt.Errorf("模型返回的 JSON 无法解析：%w", err)
	}
	return &out, nil
}

// glossaryError 方舟错误转为中文提示
func glossaryError(err error) error {
	var ae *ark.Error
	if errors.As(err, &ae) {
		switch {
		case ae.Status == 401 || ae.Status == 403:
			return errors.New("智能服务 API Key 无效或无权限，请联系平台管理员")
		case ae.Status == 404:
			return errors.New("智能服务推理接入点不存在或未开通，请联系平台管理员")
		case ae.Status == 429:
			return errors.New("智能服务调用频率或额度超限，请稍后重试")
		}
		return fmt.Errorf("智能服务调用失败：%s", vendortext.Sanitize(ae.Message))
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("提取超时，请缩短文章后重试")
	}
	return fmt.Errorf("提取失败：%w", err)
}

func cleanTerm(s string) string {
	s = strings.TrimSpace(strings.Trim(s, "\"“”'‘’"))
	if utf8.RuneCountInString(s) > glossaryMaxTermRunes {
		return ""
	}
	return s
}

func containsFold(text, term string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(term))
}

func countFold(text, term string) int {
	return strings.Count(strings.ToLower(text), strings.ToLower(term))
}

// splitAligned 长文分段：按段落把原文、译文切成相同段数，各段按长度比例大致对齐
func splitAligned(src, tgt string) ([]string, []string) {
	n := int(math.Ceil(float64(utf8.RuneCountInString(src)) / glossaryChunkRunes))
	if n <= 1 {
		return []string{src}, []string{tgt}
	}
	return splitParagraphs(src, n), splitParagraphs(tgt, n)
}

// splitParagraphs 按段落把文本分成 n 段（尽量等长，不拆开段落）
func splitParagraphs(text string, n int) []string {
	paras := strings.Split(text, "\n")
	total := utf8.RuneCountInString(text)
	out := make([]string, 0, n)
	var cur []string
	acc := 0
	for _, p := range paras {
		cur = append(cur, p)
		acc += utf8.RuneCountInString(p) + 1
		if len(out) < n-1 && acc >= total*(len(out)+1)/n {
			out = append(out, strings.Join(cur, "\n"))
			cur = nil
		}
	}
	out = append(out, strings.Join(cur, "\n"))
	for len(out) < n {
		out = append(out, "")
	}
	return out
}
