package live

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/config"
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
	tencentClient *tencent.Client
	domainMgr     *tencentLive.DomainManager
	streamMgr     *tencentLive.StreamManager
	pullStreamMgr *tencentLive.PullStreamManager
	auditMgr      *tencentLive.AuditManager
	recordMgr     *tencentLive.RecordManager
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

// NewService 创建直播服务
func NewService(cfg *config.TencentCloudConfig) (Service, error) {
	// 创建腾讯云客户端
	client, err := tencent.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	liveClient := client.GetLiveClient()

	return &liveService{
		tencentClient: client,
		domainMgr:     tencentLive.NewDomainManager(liveClient),
		streamMgr:     tencentLive.NewStreamManager(liveClient),
		pullStreamMgr: tencentLive.NewPullStreamManagerWithProvider(liveClient, client),
		auditMgr:      tencentLive.NewAuditManager(),
		recordMgr:     tencentLive.NewRecordManager(liveClient),
		redisClient:   redisRepo.Client,
	}, nil
}

// 域名管理实现
func (s *liveService) AddDomain(ctx context.Context, req *tencentLive.AddDomainRequest) error {
	err := s.domainMgr.AddDomain(ctx, req)
	if err != nil {
		return err
	}
	// 清除缓存
	s.clearDomainCache(ctx)
	return nil
}

func (s *liveService) DeleteDomain(ctx context.Context, domainName string) error {
	err := s.domainMgr.DeleteDomain(ctx, domainName)
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
	domainInfo, err := s.domainMgr.DescribeDomain(ctx, domainName)
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
	domains, err := s.domainMgr.ListDomains(ctx, nil) // 获取所有域名
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
	err := s.domainMgr.EnableDomain(ctx, domainName)
	if err != nil {
		return err
	}
	// 清除缓存
	s.clearDomainCache(ctx)
	s.clearDomainDetailCache(ctx, domainName)
	return nil
}

func (s *liveService) ForbidDomain(ctx context.Context, domainName string) error {
	err := s.domainMgr.ForbidDomain(ctx, domainName)
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
	return s.domainMgr.AuthenticateDomainOwner(ctx, domainName, verifyType)
}

// 流管理实现
func (s *liveService) DescribeStreamOnlineList(ctx context.Context, req *tencentLive.DescribeStreamOnlineListRequest) (*tencentLive.DescribeStreamOnlineListResponse, error) {
	return s.streamMgr.DescribeStreamOnlineList(ctx, req)
}

func (s *liveService) DescribeStreamHistoryList(ctx context.Context, req *tencentLive.DescribeStreamHistoryListRequest) (*tencentLive.DescribeStreamHistoryListResponse, error) {
	return s.streamMgr.DescribeStreamHistoryList(ctx, req)
}

func (s *liveService) DescribeStreamEventList(ctx context.Context, req *tencentLive.DescribeStreamEventListRequest) (*tencentLive.DescribeStreamEventListResponse, error) {
	return s.streamMgr.DescribeStreamEventList(ctx, req)
}

func (s *liveService) DropLiveStream(ctx context.Context, req *tencentLive.DropLiveStreamRequest) error {
	return s.streamMgr.DropLiveStream(ctx, req)
}

func (s *liveService) ResumeLiveStream(ctx context.Context, req *tencentLive.ResumeLiveStreamRequest) error {
	return s.streamMgr.ResumeLiveStream(ctx, req)
}

// 拉流管理实现
func (s *liveService) CreatePullStreamTask(ctx context.Context, req *tencentLive.CreatePullStreamTaskRequest) (string, error) {
	return s.pullStreamMgr.CreatePullStreamTask(ctx, req)
}

func (s *liveService) DeletePullStreamTask(ctx context.Context, taskID string, operator string) error {
	return s.pullStreamMgr.DeletePullStreamTask(ctx, taskID, operator)
}

func (s *liveService) DescribePullStreamTasks(ctx context.Context, req *tencentLive.DescribePullStreamTasksRequest) (*tencentLive.DescribePullStreamTasksResponse, error) {
	return s.pullStreamMgr.DescribePullStreamTasks(ctx, req)
}

func (s *liveService) UpdatePullStreamTask(ctx context.Context, req *tencentLive.UpdatePullStreamTaskRequest) error {
	return s.pullStreamMgr.UpdatePullStreamTask(ctx, req)
}

func (s *liveService) DescribePullStreamTaskStatus(ctx context.Context, taskID string) (*tencentLive.PullStreamTaskStatus, error) {
	return s.pullStreamMgr.DescribePullStreamTaskStatus(ctx, taskID)
}

func (s *liveService) RestartPullStreamTask(ctx context.Context, taskID string, operator string) error {
	return s.pullStreamMgr.RestartPullStreamTask(ctx, taskID, operator)
}

func (s *liveService) DescribePullTransformPushInfoList(ctx context.Context, req *tencentLive.DescribePullTransformPushInfoListRequest) (*tencentLive.DescribePullTransformPushInfoListResponse, error) {
	return s.pullStreamMgr.DescribePullTransformPushInfoList(ctx, req)
}

// 审核管理实现
func (s *liveService) CreateKeywordLib(ctx context.Context, req *tencentLive.CreateKeywordLibRequest) (int64, error) {
	return s.auditMgr.CreateKeywordLib(ctx, req)
}

func (s *liveService) CreateKeywords(ctx context.Context, req *tencentLive.CreateKeywordsRequest) error {
	return s.auditMgr.CreateKeywords(ctx, req)
}

func (s *liveService) DeleteKeywords(ctx context.Context, libID int64, keywords []string) error {
	return s.auditMgr.DeleteKeywords(ctx, libID, keywords)
}

func (s *liveService) DescribeKeywords(ctx context.Context, libID int64) ([]*tencentLive.KeywordInfo, error) {
	return s.auditMgr.DescribeKeywords(ctx, libID)
}

// 录制管理实现
func (s *liveService) CreateRecordTemplate(ctx context.Context, req *tencentLive.CreateRecordTemplateRequest) (int64, error) {
	return s.recordMgr.CreateRecordTemplate(ctx, req)
}

func (s *liveService) DeleteRecordTemplate(ctx context.Context, templateID int64) error {
	return s.recordMgr.DeleteRecordTemplate(ctx, templateID)
}

func (s *liveService) DescribeRecordTemplates(ctx context.Context) ([]*tencentLive.RecordTemplateInfo, error) {
	return s.recordMgr.DescribeRecordTemplates(ctx)
}

func (s *liveService) CreateRecordRule(ctx context.Context, req *tencentLive.CreateRecordRuleRequest) error {
	return s.recordMgr.CreateRecordRule(ctx, req)
}

func (s *liveService) DeleteRecordRule(ctx context.Context, domainName, appName, streamName string) error {
	return s.recordMgr.DeleteRecordRule(ctx, domainName, appName, streamName)
}

func (s *liveService) CreateRecordTask(ctx context.Context, req *tencentLive.CreateRecordTaskRequest) (string, error) {
	return s.recordMgr.CreateRecordTask(ctx, req)
}

func (s *liveService) StopRecordTask(ctx context.Context, taskID string) error {
	return s.recordMgr.StopRecordTask(ctx, taskID)
}

func (s *liveService) DeleteRecordTask(ctx context.Context, taskID string) error {
	return s.recordMgr.DeleteRecordTask(ctx, taskID)
}
