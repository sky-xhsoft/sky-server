package live

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	redisRepo "github.com/sky-xhsoft/sky-server/internal/repository/redis"
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

	// 流管理
	DescribeStreamOnlineList(ctx context.Context, req *tencentLive.DescribeStreamOnlineListRequest) (*tencentLive.DescribeStreamOnlineListResponse, error)
	DescribeStreamHistoryList(ctx context.Context, req *tencentLive.DescribeStreamHistoryListRequest) (*tencentLive.DescribeStreamHistoryListResponse, error)
	DescribeStreamEventList(ctx context.Context, req *tencentLive.DescribeStreamEventListRequest) (*tencentLive.DescribeStreamEventListResponse, error)
	DropLiveStream(ctx context.Context, req *tencentLive.DropLiveStreamRequest) error
	ResumeLiveStream(ctx context.Context, req *tencentLive.ResumeLiveStreamRequest) error

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
	// 域名列表缓存key
	domainListCacheKey = "live:domains:list"
	// 域名详情缓存key前缀
	domainDetailCacheKeyPrefix = "live:domain:detail:"
	// 缓存过期时间（5分钟）
	domainCacheExpiration = 24 * time.Hour
)

// NewService 创建直播服务（已弃用，请使用 NewServiceWithCompanyManager）
func NewService(cfg *config.TencentCloudConfig) (Service, error) {
	// 创建腾讯云客户端
	client, err := tencent.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &liveService{
		tencentClient: client,
		redisClient:   redisRepo.Client,
	}, nil
}

// NewServiceWithCompanyManager 创建支持公司级配置的直播服务
func NewServiceWithCompanyManager(tencentMgr *tencent.CompanyTencentManager) (Service, error) {
	defaultClient := tencentMgr.GetDefaultClient()

	return &liveService{
		tencentMgr:    tencentMgr,
		tencentClient: defaultClient,
		redisClient:   redisRepo.Client,
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
		cacheKey := domainDetailCacheKeyPrefix + domainName
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
		cacheKey := domainDetailCacheKeyPrefix + domainName
		if data, err := json.Marshal(domainInfo); err == nil {
			s.redisClient.Set(ctx, cacheKey, data, domainCacheExpiration)
		}
	}

	return domainInfo, nil
}

func (s *liveService) ListDomains(ctx context.Context, domainType *int64) ([]*tencentLive.DomainInfo, error) {
	// 尝试从缓存获取
	if s.redisClient != nil {
		cachedData, err := s.redisClient.Get(ctx, domainListCacheKey).Result()
		if err == nil && cachedData != "" {
			var domains []*tencentLive.DomainInfo
			if err := json.Unmarshal([]byte(cachedData), &domains); err == nil {
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

	// 过滤掉以 myqcloud.com 结尾的域名
	filteredDomains := make([]*tencentLive.DomainInfo, 0)
	for _, domain := range domains {
		if !strings.HasSuffix(domain.Name, ".myqcloud.com") {
			filteredDomains = append(filteredDomains, domain)
		}
	}

	// 写入缓存（缓存过滤后的域名）
	if s.redisClient != nil && filteredDomains != nil {
		if data, err := json.Marshal(filteredDomains); err == nil {
			s.redisClient.Set(ctx, domainListCacheKey, data, domainCacheExpiration)
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

// clearDomainCache 清除域名列表缓存
func (s *liveService) clearDomainCache(ctx context.Context) {
	if s.redisClient != nil {
		s.redisClient.Del(ctx, domainListCacheKey)
	}
}

// clearDomainDetailCache 清除域名详情缓存
func (s *liveService) clearDomainDetailCache(ctx context.Context, domainName string) {
	if s.redisClient != nil {
		cacheKey := domainDetailCacheKeyPrefix + domainName
		s.redisClient.Del(ctx, cacheKey)
	}
}

func (s *liveService) AuthenticateDomainOwner(ctx context.Context, domainName string, verifyType string) (*tencentLive.DomainOwnerVerifyResult, error) {
	return s.getDomainManager(ctx).AuthenticateDomainOwner(ctx, domainName, verifyType)
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
