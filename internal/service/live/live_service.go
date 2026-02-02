package live

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/config"
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

	// 流管理
	DescribeStreamOnlineList(ctx context.Context, req *tencentLive.DescribeStreamOnlineListRequest) (*tencentLive.DescribeStreamOnlineListResponse, error)
	DescribeStreamHistoryList(ctx context.Context, req *tencentLive.DescribeStreamHistoryListRequest) (*tencentLive.DescribeStreamHistoryListResponse, error)
	DescribeStreamEventList(ctx context.Context, req *tencentLive.DescribeStreamEventListRequest) (*tencentLive.DescribeStreamEventListResponse, error)
	DropLiveStream(ctx context.Context, req *tencentLive.DropLiveStreamRequest) error
	ResumeLiveStream(ctx context.Context, req *tencentLive.ResumeLiveStreamRequest) error

	// 拉流管理
	CreatePullStreamTask(ctx context.Context, req *tencentLive.CreatePullStreamTaskRequest) (string, error)
	DeletePullStreamTask(ctx context.Context, taskID string, operator string) error
	DescribePullStreamTasks(ctx context.Context, taskID *string) ([]*tencentLive.PullStreamTaskInfo, error)
	UpdatePullStreamTask(ctx context.Context, req *tencentLive.UpdatePullStreamTaskRequest) error
	DescribePullStreamTaskStatus(ctx context.Context, taskID string) (*tencentLive.PullStreamTaskStatus, error)
	RestartPullStreamTask(ctx context.Context, taskID string, operator string) error

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
}

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
	}, nil
}

// 域名管理实现
func (s *liveService) AddDomain(ctx context.Context, req *tencentLive.AddDomainRequest) error {
	return s.domainMgr.AddDomain(ctx, req)
}

func (s *liveService) DeleteDomain(ctx context.Context, domainName string) error {
	return s.domainMgr.DeleteDomain(ctx, domainName)
}

func (s *liveService) DescribeDomain(ctx context.Context, domainName string) (*tencentLive.DomainInfo, error) {
	return s.domainMgr.DescribeDomain(ctx, domainName)
}

func (s *liveService) ListDomains(ctx context.Context, domainType *int64) ([]*tencentLive.DomainInfo, error) {
	return s.domainMgr.ListDomains(ctx, domainType)
}

func (s *liveService) EnableDomain(ctx context.Context, domainName string) error {
	return s.domainMgr.EnableDomain(ctx, domainName)
}

func (s *liveService) ForbidDomain(ctx context.Context, domainName string) error {
	return s.domainMgr.ForbidDomain(ctx, domainName)
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

func (s *liveService) DescribePullStreamTasks(ctx context.Context, taskID *string) ([]*tencentLive.PullStreamTaskInfo, error) {
	return s.pullStreamMgr.DescribePullStreamTasks(ctx, taskID)
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
