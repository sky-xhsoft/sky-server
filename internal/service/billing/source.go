// Package billing 直播计费。一期只计量不扣费：推流段落（直播时长）、按直播间的播放流量、存储空间
package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
)

// GlobalAccount 平台腾讯云账号（未配置自有密钥的公司都使用它）
const GlobalAccount = "global"

// LiveAPI 计量用到的腾讯云直播接口（测试中可替换）
type LiveAPI interface {
	// OnlineStreams 当前正在推流的流名
	OnlineStreams(ctx context.Context) (map[string]bool, error)
	// PlayDomains 账号下的播放域名
	PlayDomains(ctx context.Context) ([]string, error)
	// DayStreamFlux 某播放域名某天每条流的流量（监控数据，MB）
	DayStreamFlux(ctx context.Context, domain, day string) (map[string]float64, error)
	// DomainBillingFlux 某播放域名某天的计费流量（MB）
	DomainBillingFlux(ctx context.Context, domain, day string) (float64, error)
}

// Accounts 公司 → 腾讯云账号
type Accounts interface {
	// Of 公司使用的账号：平台账号为 GlobalAccount，自有密钥为 tc:<密钥 ID 摘要>
	Of(ctx context.Context, companyID uint) string
	// API 账号对应的接口
	API(ctx context.Context, account string) (LiveAPI, error)
}

// dataRegion 直播数据接口不区分地域，固定用支持直播接口的地域，避免公司配置的地域（如 ap-nanjing）不支持
const dataRegion = "ap-guangzhou"

// tencentAccounts 基于公司腾讯云配置的账号实现。账号按密钥 ID 区分：多个公司填了同一套密钥时是同一个账号，
// 填的是平台密钥时就是平台账号；账号标识只保存密钥 ID 的摘要，不保存密钥
type tencentAccounts struct {
	mgr     *tencent.CompanyTencentManager
	mu      sync.Mutex
	clients map[string]*tencent.Client // 账号 → 客户端（Of 时记录）
}

// NewTencentAccounts 公司腾讯云账号
func NewTencentAccounts(mgr *tencent.CompanyTencentManager) Accounts {
	return &tencentAccounts{mgr: mgr, clients: map[string]*tencent.Client{}}
}

// accountKey 密钥 ID 的摘要
func accountKey(secretID string) string {
	sum := sha256.Sum256([]byte(secretID))
	return "tc:" + hex.EncodeToString(sum[:])[:10]
}

func (a *tencentAccounts) Of(ctx context.Context, companyID uint) string {
	if a.mgr == nil {
		return GlobalAccount
	}
	global := a.mgr.GetGlobalClient()
	client := global
	if a.mgr.HasCompanyConfig(ctx, companyID) {
		if c, err := a.mgr.GetClient(ctx, companyID); err == nil && c != nil {
			client = c
		}
	}
	key := GlobalAccount
	if client != nil && client != global && client.GetConfig() != nil &&
		(global == nil || global.GetConfig() == nil || client.GetConfig().SecretID != global.GetConfig().SecretID) {
		key = accountKey(client.GetConfig().SecretID)
	}
	a.mu.Lock()
	a.clients[key] = client
	a.mu.Unlock()
	return key
}

func (a *tencentAccounts) API(ctx context.Context, account string) (LiveAPI, error) {
	if a.mgr == nil {
		return nil, fmt.Errorf("未配置腾讯云")
	}
	a.mu.Lock()
	client, ok := a.clients[account]
	a.mu.Unlock()
	if !ok && account == GlobalAccount {
		client, ok = a.mgr.GetGlobalClient(), true
	}
	if !ok || client == nil || client.GetConfig() == nil || client.GetConfig().SecretID == "" {
		return nil, fmt.Errorf("账号 %s 未配置腾讯云密钥", account)
	}
	lc, err := client.GetLiveClientForRegion(dataRegion)
	if err != nil {
		return nil, err
	}
	return &tencentLive{c: lc}, nil
}

type tencentLive struct{ c *live.Client }

func (t *tencentLive) OnlineStreams(ctx context.Context) (map[string]bool, error) {
	out := map[string]bool{}
	for page := uint64(1); page <= 100; page++ {
		req := live.NewDescribeLiveStreamOnlineListRequest()
		req.PageNum = common.Uint64Ptr(page)
		req.PageSize = common.Uint64Ptr(1000)
		resp, err := t.c.DescribeLiveStreamOnlineListWithContext(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, s := range resp.Response.OnlineInfo {
			if s.StreamName != nil {
				out[*s.StreamName] = true
			}
		}
		if resp.Response.TotalPage == nil || page >= *resp.Response.TotalPage {
			break
		}
	}
	return out, nil
}

func (t *tencentLive) PlayDomains(ctx context.Context) ([]string, error) {
	var out []string
	for page := uint64(1); page <= 20; page++ {
		req := live.NewDescribeLiveDomainsRequest()
		req.DomainType = common.Uint64Ptr(1) // 1 播放域名
		req.PageNum = common.Uint64Ptr(page)
		req.PageSize = common.Uint64Ptr(100)
		resp, err := t.c.DescribeLiveDomainsWithContext(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, d := range resp.Response.DomainList {
			if d.Name != nil {
				out = append(out, *d.Name)
			}
		}
		if resp.Response.AllCount == nil || page*100 >= *resp.Response.AllCount {
			break
		}
	}
	return out, nil
}

func (t *tencentLive) DayStreamFlux(ctx context.Context, domain, day string) (map[string]float64, error) {
	out := map[string]float64{}
	for page := uint64(1); page <= 1000; page++ {
		req := live.NewDescribeStreamDayPlayInfoListRequest()
		req.DayTime = common.StringPtr(day)
		req.PlayDomain = common.StringPtr(domain)
		req.PageNum = common.Uint64Ptr(page)
		req.PageSize = common.Uint64Ptr(1000)
		resp, err := t.c.DescribeStreamDayPlayInfoListWithContext(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, s := range resp.Response.DataInfoList {
			if s.StreamName != nil && s.TotalFlux != nil {
				out[*s.StreamName] += *s.TotalFlux
			}
		}
		if resp.Response.TotalPage == nil || page >= *resp.Response.TotalPage {
			break
		}
	}
	return out, nil
}

func (t *tencentLive) DomainBillingFlux(ctx context.Context, domain, day string) (float64, error) {
	req := live.NewDescribeBillBandwidthAndFluxListRequest()
	req.StartTime = common.StringPtr(day + " 00:00:00")
	req.EndTime = common.StringPtr(day + " 23:59:59")
	req.PlayDomains = []*string{common.StringPtr(domain)}
	req.Granularity = common.Uint64Ptr(1440)
	resp, err := t.c.DescribeBillBandwidthAndFluxListWithContext(ctx, req)
	if err != nil {
		return 0, err
	}
	if resp.Response.SumFlux == nil {
		return 0, nil
	}
	return *resp.Response.SumFlux, nil
}

// roomIDOfStream 流名对应的直播间 ID：流名即直播间 ID，转码流为「ID_模板名」
func roomIDOfStream(stream string) uint {
	if i := strings.IndexByte(stream, '_'); i > 0 {
		stream = stream[:i]
	}
	var id uint
	for _, ch := range stream {
		if ch < '0' || ch > '9' {
			return 0
		}
		id = id*10 + uint(ch-'0')
		if id > 1<<31 {
			return 0
		}
	}
	return id
}
