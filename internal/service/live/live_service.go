package live

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

// Service 直播服务接口
type Service interface {
	// 域名管理
	AddDomain(ctx context.Context, req *tencentLive.AddDomainRequest) error
	DeleteDomain(ctx context.Context, domainName string) error
	DescribeDomain(ctx context.Context, domainName string) (*tencentLive.DomainInfo, error)
	ListDomains(ctx context.Context, domainType *int64) ([]*tencentLive.DomainInfo, error)
	EnableDomain(ctx context.Context, domainName string) error
	ForbidDomain(ctx context.Context, domainName string) error
	AuthenticateDomainOwner(ctx context.Context, domainName string, verifyType string) (*tencentLive.DomainOwnerVerifyResult, error)
	GetDomainAuthKey(ctx context.Context, domainName, keyType string) (*tencentLive.AuthKeyInfo, error)

	// 流管理
	DescribeStreamOnlineList(ctx context.Context, req *tencentLive.DescribeStreamOnlineListRequest) (*tencentLive.DescribeStreamOnlineListResponse, error)
	DescribeStreamHistoryList(ctx context.Context, req *tencentLive.DescribeStreamHistoryListRequest) (*tencentLive.DescribeStreamHistoryListResponse, error)
	DescribeStreamEventList(ctx context.Context, req *tencentLive.DescribeStreamEventListRequest) (*tencentLive.DescribeStreamEventListResponse, error)
	DropLiveStream(ctx context.Context, req *tencentLive.DropLiveStreamRequest) error
	ResumeLiveStream(ctx context.Context, req *tencentLive.ResumeLiveStreamRequest) error
	GetStreamMetrics(ctx context.Context, req *tencentLive.GetStreamMetricsRequest) (*tencentLive.StreamMetrics, error)

	// 拉流管理
	CreatePullStreamTask(ctx context.Context, req *tencentLive.CreatePullStreamTaskRequest) (string, error)
	DeletePullStreamTask(ctx context.Context, taskID string, operator string) error
	DescribePullStreamTasks(ctx context.Context, req *tencentLive.DescribePullStreamTasksRequest) (*tencentLive.DescribePullStreamTasksResponse, error)
	UpdatePullStreamTask(ctx context.Context, req *tencentLive.UpdatePullStreamTaskRequest) error
	DescribePullStreamTaskStatus(ctx context.Context, taskID string) (*tencentLive.PullStreamTaskStatus, error)
	RestartPullStreamTask(ctx context.Context, taskID string, operator string) error
	DescribePullTransformPushInfoList(ctx context.Context, req *tencentLive.DescribePullTransformPushInfoListRequest) (*tencentLive.DescribePullTransformPushInfoListResponse, error)

	// 审核管理
	CreateKeywordLib(ctx context.Context, req *tencentLive.CreateKeywordLibRequest) (int64, error)
	CreateKeywords(ctx context.Context, req *tencentLive.CreateKeywordsRequest) error
	DeleteKeywords(ctx context.Context, libID int64, keywords []string) error
	DescribeKeywords(ctx context.Context, libID int64) ([]*tencentLive.KeywordInfo, error)

	// 录制管理
	CreateRecordTemplate(ctx context.Context, req *tencentLive.CreateRecordTemplateRequest) (int64, error)
	DeleteRecordTemplate(ctx context.Context, templateID int64) error
	DescribeRecordTemplates(ctx context.Context) ([]*tencentLive.RecordTemplateInfo, error)
	CreateRecordRule(ctx context.Context, req *tencentLive.CreateRecordRuleRequest) error
	DeleteRecordRule(ctx context.Context, domainName, appName, streamName string) error
	CreateRecordTask(ctx context.Context, req *tencentLive.CreateRecordTaskRequest) (string, error)
	StopRecordTask(ctx context.Context, taskID string) error
	DeleteRecordTask(ctx context.Context, taskID string) error
}

// liveService 直播服务实现
type liveService struct {
	tencentMgr    *tencent.CompanyTencentManager // 公司级腾讯云客户端管理器
	tencentClient *tencent.Client                // 兼容旧模式的全局客户端
	redisClient   *redis.Client
}

const (
	// 域名列表缓存key前缀
	domainListCacheKeyPrefix = "live:domains:list:"
	// 域名详情缓存key前缀
	domainDetailCacheKeyPrefix = "live:domain:detail:"
	// 缓存过期时间（5分钟）
	domainCacheExpiration = 24 * time.Hour
)

// NewService 创建直播服务（已弃用，请使用 NewServiceWithCompanyManager）
func NewService(cfg *config.TencentCloudConfig, redisClient *redis.Client) (Service, error) {
	client, err := tencent.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &liveService{
		tencentClient: client,
		redisClient:   redisClient,
	}, nil
}

// NewServiceWithCompanyManager 创建支持公司级配置的直播服务
func NewServiceWithCompanyManager(tencentMgr *tencent.CompanyTencentManager, redisClient *redis.Client) (Service, error) {
	defaultClient := tencentMgr.GetDefaultClient()

	return &liveService{
		tencentMgr:    tencentMgr,
		tencentClient: defaultClient,
		redisClient:   redisClient,
	}, nil
}

// getTencentClient 从context中获取公司ID，返回对应的腾讯云客户端
func (s *liveService) getTencentClient(ctx context.Context) *tencent.Client {
	if s.tencentMgr == nil {
		return s.tencentClient
	}

	// 尝试从context中获取companyID
	if companyIDVal := ctx.Value(contextutil.CompanyIDKey); companyIDVal != nil {
		if companyID, ok := companyIDVal.(uint); ok {
			client, err := s.tencentMgr.GetClient(ctx, companyID)
			if err == nil && client != nil {
				return client
			}
		}
	}

	// 如果没有companyID或获取失败，返回默认客户端
	return s.tencentMgr.GetDefaultClient()
}

// getDomainManager 获取域名管理器
func (s *liveService) getDomainManager(ctx context.Context) *tencentLive.DomainManager {
	client := s.getTencentClient(ctx)
	return tencentLive.NewDomainManager(client.GetLiveClient())
}

// getStreamManager 获取流管理器
func (s *liveService) getStreamManager(ctx context.Context) *tencentLive.StreamManager {
	client := s.getTencentClient(ctx)
	return tencentLive.NewStreamManager(client.GetLiveClient())
}

// getPullStreamManager 获取拉流管理器
func (s *liveService) getPullStreamManager(ctx context.Context) *tencentLive.PullStreamManager {
	client := s.getTencentClient(ctx)
	return tencentLive.NewPullStreamManagerWithProvider(client.GetLiveClient(), client)
}

// getRecordManager 获取录制管理器
func (s *liveService) getRecordManager(ctx context.Context) *tencentLive.RecordManager {
	client := s.getTencentClient(ctx)
	return tencentLive.NewRecordManager(client.GetLiveClient())
}

// getAuditManager 获取审核管理器
func (s *liveService) getAuditManager() *tencentLive.AuditManager {
	return tencentLive.NewAuditManager()
}

// 域名管理实现
func (s *liveService) AddDomain(ctx context.Context, req *tencentLive.AddDomainRequest) error {
	err := s.getDomainManager(ctx).AddDomain(ctx, req)
	if err != nil {
		return err
	}
	// 清除缓存
	s.clearDomainCache(ctx)
	return nil
}

func (s *liveService) DeleteDomain(ctx context.Context, domainName string) error {
	err := s.getDomainManager(ctx).DeleteDomain(ctx, domainName)
	if err != nil {
		return err
	}
	// 清除缓存
	s.clearDomainCache(ctx)
	s.clearDomainDetailCache(ctx, domainName)
	return nil
}

func (s *liveService) DescribeDomain(ctx context.Context, domainName string) (*tencentLive.DomainInfo, error) {
	// 尝试从缓存获取
	if s.redisClient != nil {
		cacheKey := s.domainDetailCacheKey(ctx, domainName)
		cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var domainInfo tencentLive.DomainInfo
			if err := json.Unmarshal([]byte(cachedData), &domainInfo); err == nil {
				return &domainInfo, nil
			}
		}
	}

	// 从腾讯云获取
	domainInfo, err := s.getDomainManager(ctx).DescribeDomain(ctx, domainName)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if s.redisClient != nil && domainInfo != nil {
		cacheKey := s.domainDetailCacheKey(ctx, domainName)
		if data, err := json.Marshal(domainInfo); err == nil {
			s.redisClient.Set(ctx, cacheKey, data, domainCacheExpiration)
		}
	}

	return domainInfo, nil
}

func (s *liveService) ListDomains(ctx context.Context, domainType *int64) ([]*tencentLive.DomainInfo, error) {
	// 尝试从缓存获取
	if s.redisClient != nil {
		cachedData, err := s.redisClient.Get(ctx, s.domainListCacheKey(ctx)).Result()
		if err == nil && cachedData != "" {
			var domains []*tencentLive.DomainInfo
			if err := json.Unmarshal([]byte(cachedData), &domains); err == nil {
				// 旧版本写入的缓存可能含系统域名，读取时同样过滤
				domains = excludeSystemDomains(domains)
				// 如果指定了域名类型，需要过滤
				if domainType != nil {
					filtered := make([]*tencentLive.DomainInfo, 0)
					for _, domain := range domains {
						if domain.Type == *domainType {
							filtered = append(filtered, domain)
						}
					}
					return filtered, nil
				}
				return domains, nil
			}
		}
	}

	// 从腾讯云获取
	domains, err := s.getDomainManager(ctx).ListDomains(ctx, nil) // 获取所有域名
	if err != nil {
		return nil, err
	}

	filteredDomains := excludeSystemDomains(domains)

	if s.redisClient != nil {
		cacheKey := s.domainListCacheKey(ctx)
		if data, err := json.Marshal(filteredDomains); err == nil {
			s.redisClient.Set(ctx, cacheKey, data, domainCacheExpiration)
		}
	}

	// 如果指定了域名类型，需要过滤
	if domainType != nil {
		filtered := make([]*tencentLive.DomainInfo, 0)
		for _, domain := range filteredDomains {
			if domain.Type == *domainType {
				filtered = append(filtered, domain)
			}
		}
		return filtered, nil
	}

	return filteredDomains, nil
}

func (s *liveService) EnableDomain(ctx context.Context, domainName string) error {
	err := s.getDomainManager(ctx).EnableDomain(ctx, domainName)
	if err != nil {
		return err
	}
	// 清除缓存
	s.clearDomainCache(ctx)
	s.clearDomainDetailCache(ctx, domainName)
	return nil
}

func (s *liveService) ForbidDomain(ctx context.Context, domainName string) error {
	err := s.getDomainManager(ctx).ForbidDomain(ctx, domainName)
	if err != nil {
		return err
	}
	// 清除缓存
	s.clearDomainCache(ctx)
	s.clearDomainDetailCache(ctx, domainName)
	return nil
}

// domainListCacheKey 按公司生成域名列表缓存key
func (s *liveService) domainListCacheKey(ctx context.Context) string {
	return domainListCacheKeyPrefix + s.companyID(ctx)
}

// domainDetailCacheKey 按公司+域名生成详情缓存key
func (s *liveService) domainDetailCacheKey(ctx context.Context, domainName string) string {
	return domainDetailCacheKeyPrefix + s.companyID(ctx) + ":" + domainName
}

// companyID 从context中提取公司ID，未取到用1
func (s *liveService) companyID(ctx context.Context) string {
	if v := ctx.Value(contextutil.CompanyIDKey); v != nil {
		if id, ok := v.(uint); ok && id != 0 {
			return strconv.FormatUint(uint64(id), 10)
		}
	}
	return "1"
}

// clearDomainCache 清除当前公司域名列表缓存
func (s *liveService) clearDomainCache(ctx context.Context) {
	if s.redisClient != nil {
		s.redisClient.Del(ctx, s.domainListCacheKey(ctx))
	}
}

// clearDomainDetailCache 清除当前公司指定域名详情缓存
func (s *liveService) clearDomainDetailCache(ctx context.Context, domainName string) {
	if s.redisClient != nil {
		cacheKey := s.domainDetailCacheKey(ctx, domainName)
		s.redisClient.Del(ctx, cacheKey)
	}
}

func (s *liveService) AuthenticateDomainOwner(ctx context.Context, domainName string, verifyType string) (*tencentLive.DomainOwnerVerifyResult, error) {
	return s.getDomainManager(ctx).AuthenticateDomainOwner(ctx, domainName, verifyType)
}

func (s *liveService) GetDomainAuthKey(ctx context.Context, domainName, keyType string) (*tencentLive.AuthKeyInfo, error) {
	mgr := s.getDomainManager(ctx)
	// keyType: push-推流鉴权, play-播放鉴权
	if keyType == "play" {
		return mgr.GetPlayAuthKey(ctx, domainName)
	}
	return mgr.GetPushAuthKey(ctx, domainName)
}

// 流管理实现
func (s *liveService) DescribeStreamOnlineList(ctx context.Context, req *tencentLive.DescribeStreamOnlineListRequest) (*tencentLive.DescribeStreamOnlineListResponse, error) {
	return s.getStreamManager(ctx).DescribeStreamOnlineList(ctx, req)
}

func (s *liveService) DescribeStreamHistoryList(ctx context.Context, req *tencentLive.DescribeStreamHistoryListRequest) (*tencentLive.DescribeStreamHistoryListResponse, error) {
	return s.getStreamManager(ctx).DescribeStreamHistoryList(ctx, req)
}

func (s *liveService) DescribeStreamEventList(ctx context.Context, req *tencentLive.DescribeStreamEventListRequest) (*tencentLive.DescribeStreamEventListResponse, error) {
	return s.getStreamManager(ctx).DescribeStreamEventList(ctx, req)
}

func (s *liveService) DropLiveStream(ctx context.Context, req *tencentLive.DropLiveStreamRequest) error {
	return s.getStreamManager(ctx).DropLiveStream(ctx, req)
}

func (s *liveService) ResumeLiveStream(ctx context.Context, req *tencentLive.ResumeLiveStreamRequest) error {
	return s.getStreamManager(ctx).ResumeLiveStream(ctx, req)
}

func (s *liveService) GetStreamMetrics(ctx context.Context, req *tencentLive.GetStreamMetricsRequest) (*tencentLive.StreamMetrics, error) {
	return s.getStreamManager(ctx).GetStreamMetrics(ctx, req)
}

// 拉流管理实现
func (s *liveService) CreatePullStreamTask(ctx context.Context, req *tencentLive.CreatePullStreamTaskRequest) (string, error) {
	return s.getPullStreamManager(ctx).CreatePullStreamTask(ctx, req)
}

func (s *liveService) DeletePullStreamTask(ctx context.Context, taskID string, operator string) error {
	return s.getPullStreamManager(ctx).DeletePullStreamTask(ctx, taskID, operator)
}

func (s *liveService) DescribePullStreamTasks(ctx context.Context, req *tencentLive.DescribePullStreamTasksRequest) (*tencentLive.DescribePullStreamTasksResponse, error) {
	return s.getPullStreamManager(ctx).DescribePullStreamTasks(ctx, req)
}

func (s *liveService) UpdatePullStreamTask(ctx context.Context, req *tencentLive.UpdatePullStreamTaskRequest) error {
	return s.getPullStreamManager(ctx).UpdatePullStreamTask(ctx, req)
}

func (s *liveService) DescribePullStreamTaskStatus(ctx context.Context, taskID string) (*tencentLive.PullStreamTaskStatus, error) {
	return s.getPullStreamManager(ctx).DescribePullStreamTaskStatus(ctx, taskID)
}

func (s *liveService) RestartPullStreamTask(ctx context.Context, taskID string, operator string) error {
	return s.getPullStreamManager(ctx).RestartPullStreamTask(ctx, taskID, operator)
}

func (s *liveService) DescribePullTransformPushInfoList(ctx context.Context, req *tencentLive.DescribePullTransformPushInfoListRequest) (*tencentLive.DescribePullTransformPushInfoListResponse, error) {
	return s.getPullStreamManager(ctx).DescribePullTransformPushInfoList(ctx, req)
}

// 审核管理实现
func (s *liveService) CreateKeywordLib(ctx context.Context, req *tencentLive.CreateKeywordLibRequest) (int64, error) {
	return s.getAuditManager().CreateKeywordLib(ctx, req)
}

func (s *liveService) CreateKeywords(ctx context.Context, req *tencentLive.CreateKeywordsRequest) error {
	return s.getAuditManager().CreateKeywords(ctx, req)
}

func (s *liveService) DeleteKeywords(ctx context.Context, libID int64, keywords []string) error {
	return s.getAuditManager().DeleteKeywords(ctx, libID, keywords)
}

func (s *liveService) DescribeKeywords(ctx context.Context, libID int64) ([]*tencentLive.KeywordInfo, error) {
	return s.getAuditManager().DescribeKeywords(ctx, libID)
}

// 录制管理实现
func (s *liveService) CreateRecordTemplate(ctx context.Context, req *tencentLive.CreateRecordTemplateRequest) (int64, error) {
	return s.getRecordManager(ctx).CreateRecordTemplate(ctx, req)
}

func (s *liveService) DeleteRecordTemplate(ctx context.Context, templateID int64) error {
	return s.getRecordManager(ctx).DeleteRecordTemplate(ctx, templateID)
}

func (s *liveService) DescribeRecordTemplates(ctx context.Context) ([]*tencentLive.RecordTemplateInfo, error) {
	return s.getRecordManager(ctx).DescribeRecordTemplates(ctx)
}

func (s *liveService) CreateRecordRule(ctx context.Context, req *tencentLive.CreateRecordRuleRequest) error {
	return s.getRecordManager(ctx).CreateRecordRule(ctx, req)
}

func (s *liveService) DeleteRecordRule(ctx context.Context, domainName, appName, streamName string) error {
	return s.getRecordManager(ctx).DeleteRecordRule(ctx, domainName, appName, streamName)
}

func (s *liveService) CreateRecordTask(ctx context.Context, req *tencentLive.CreateRecordTaskRequest) (string, error) {
	return s.getRecordManager(ctx).CreateRecordTask(ctx, req)
}

func (s *liveService) StopRecordTask(ctx context.Context, taskID string) error {
	return s.getRecordManager(ctx).StopRecordTask(ctx, taskID)
}

func (s *liveService) DeleteRecordTask(ctx context.Context, taskID string) error {
	return s.getRecordManager(ctx).DeleteRecordTask(ctx, taskID)
}

// systemDomainSuffixes 腾讯云自动分配的系统域名后缀，不属于用户自有域名，不在域名列表中展示
var systemDomainSuffixes = []string{".myqcloud.com", ".tlivecloud.com"}

// isSystemDomain 判断是否为腾讯云系统域名（不区分大小写）
func isSystemDomain(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	for _, suffix := range systemDomainSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

// excludeSystemDomains 过滤掉腾讯云系统域名
func excludeSystemDomains(domains []*tencentLive.DomainInfo) []*tencentLive.DomainInfo {
	filtered := make([]*tencentLive.DomainInfo, 0, len(domains))
	for _, domain := range domains {
		if domain != nil && !isSystemDomain(domain.Name) {
			filtered = append(filtered, domain)
		}
	}
	return filtered
}
