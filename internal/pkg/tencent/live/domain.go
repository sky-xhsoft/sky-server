package live

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
	"go.uber.org/zap"
)

// DomainManager 域名管理器
type DomainManager struct {
	client *live.Client
}

// NewDomainManager 创建域名管理器
func NewDomainManager(client *live.Client) *DomainManager {
	return &DomainManager{
		client: client,
	}
}

// AddDomainRequest 添加域名请求
type AddDomainRequest struct {
	DomainName string // 域名
	DomainType int64  // 域名类型：0-推流域名，1-播放域名
}

// AddDomain 添加域名
func (m *DomainManager) AddDomain(ctx context.Context, req *AddDomainRequest) error {
	request := live.NewAddLiveDomainRequest()
	request.DomainName = common.StringPtr(req.DomainName)
	request.DomainType = common.Uint64Ptr(uint64(req.DomainType))

	_, err := m.client.AddLiveDomain(request)
	if err != nil {
		logger.Error("Failed to add domain",
			zap.String("domainName", req.DomainName),
			zap.Int64("domainType", req.DomainType),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add domain: %w", err)
	}

	return nil
}

// DeleteDomain 删除域名
func (m *DomainManager) DeleteDomain(ctx context.Context, domainName string) error {
	// 先查询域名信息获取域名类型
	domainInfo, err := m.DescribeDomain(ctx, domainName)
	if err != nil {
		logger.Error("Failed to get domain info",
			zap.String("domainName", domainName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get domain info: %w", err)
	}

	request := live.NewDeleteLiveDomainRequest()
	request.DomainName = common.StringPtr(domainName)
	request.DomainType = common.Uint64Ptr(uint64(domainInfo.Type))

	_, err = m.client.DeleteLiveDomain(request)
	if err != nil {
		logger.Error("Failed to delete domain",
			zap.String("domainName", domainName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	return nil
}

// DomainInfo 域名信息
type DomainInfo struct {
	Name         string
	Type         int64
	Status       int64
	CreateTime   string
	UpdateTime   string
	Region       string // 区域
	BCName       int64  // CNAME配置状态：0-未配置成功，1-配置成功
	TargetDomain string // CNAME目标域名
}

// DescribeDomain 查询域名信息
func (m *DomainManager) DescribeDomain(ctx context.Context, domainName string) (*DomainInfo, error) {
	request := live.NewDescribeLiveDomainRequest()
	request.DomainName = common.StringPtr(domainName)

	response, err := m.client.DescribeLiveDomain(request)
	if err != nil {
		logger.Error("Failed to describe domain",
			zap.String("domainName", domainName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to describe domain: %w", err)
	}

	if response.Response.DomainInfo == nil {
		logger.Error("Domain not found",
			zap.String("domainName", domainName),
		)
		return nil, fmt.Errorf("domain not found")
	}

	info := &DomainInfo{
		Name:       *response.Response.DomainInfo.Name,
		Type:       int64(*response.Response.DomainInfo.Type),
		Status:     int64(*response.Response.DomainInfo.Status),
		CreateTime: *response.Response.DomainInfo.CreateTime,
		BCName:     0, // 默认未配置
	}

	// 获取 CNAME 目标域名（腾讯云 API 返回的字段）
	if response.Response.DomainInfo.TargetDomain != nil {
		info.TargetDomain = *response.Response.DomainInfo.TargetDomain
	}

	// 获取 CNAME 配置状态（腾讯云 API 返回的字段）
	if response.Response.DomainInfo.BCName != nil {
		info.BCName = int64(*response.Response.DomainInfo.BCName)
	}

	return info, nil
}

// ListDomains 查询域名列表
func (m *DomainManager) ListDomains(ctx context.Context, domainType *int64) ([]*DomainInfo, error) {
	request := live.NewDescribeLiveDomainsRequest()
	if domainType != nil {
		request.DomainType = common.Uint64Ptr(uint64(*domainType))
	}

	response, err := m.client.DescribeLiveDomains(request)
	if err != nil {
		logger.Error("Failed to list domains",
			zap.Reflect("domainType", domainType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	domains := make([]*DomainInfo, 0, len(response.Response.DomainList))
	for _, domain := range response.Response.DomainList {
		info := &DomainInfo{
			Name:       *domain.Name,
			Type:       int64(*domain.Type),
			Status:     int64(*domain.Status),
			CreateTime: *domain.CreateTime,
			BCName:     0, // 默认未配置
		}

		// 获取 CNAME 目标域名（腾讯云 API 返回的字段）
		if domain.TargetDomain != nil {
			info.TargetDomain = *domain.TargetDomain
		}

		// 获取 CNAME 配置状态（腾讯云 API 返回的字段）
		if domain.BCName != nil {
			info.BCName = int64(*domain.BCName)
		}

		domains = append(domains, info)
	}

	return domains, nil
}

// EnableDomain 启用域名
func (m *DomainManager) EnableDomain(ctx context.Context, domainName string) error {
	request := live.NewEnableLiveDomainRequest()
	request.DomainName = common.StringPtr(domainName)

	_, err := m.client.EnableLiveDomain(request)
	if err != nil {
		logger.Error("Failed to enable domain",
			zap.String("domainName", domainName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to enable domain: %w", err)
	}

	return nil
}

// ForbidDomain 禁用域名
func (m *DomainManager) ForbidDomain(ctx context.Context, domainName string) error {
	request := live.NewForbidLiveDomainRequest()
	request.DomainName = common.StringPtr(domainName)

	_, err := m.client.ForbidLiveDomain(request)
	if err != nil {
		logger.Error("Failed to forbid domain",
			zap.String("domainName", domainName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to forbid domain: %w", err)
	}

	return nil
}

// DomainOwnerVerifyResult 域名归属验证结果
type DomainOwnerVerifyResult struct {
	Status     int64  // 验证状态：0-未验证，1-验证成功，2-验证失败
	Content    string // 验证内容（DNS记录值）
	MainDomain string // 主域名
}

// AuthenticateDomainOwner 验证域名归属
func (m *DomainManager) AuthenticateDomainOwner(ctx context.Context, domainName string, verifyType string) (*DomainOwnerVerifyResult, error) {
	request := live.NewAuthenticateDomainOwnerRequest()
	request.DomainName = common.StringPtr(domainName)
	request.VerifyType = common.StringPtr(verifyType) // dnsCheck 或 fileCheck

	response, err := m.client.AuthenticateDomainOwner(request)
	if err != nil {
		logger.Error("Failed to authenticate domain owner",
			zap.String("domainName", domainName),
			zap.String("verifyType", verifyType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to authenticate domain owner: %w", err)
	}

	result := &DomainOwnerVerifyResult{
		Status: int64(*response.Response.Status),
	}

	if response.Response.Content != nil {
		result.Content = *response.Response.Content
	}
	if response.Response.MainDomain != nil {
		result.MainDomain = *response.Response.MainDomain
	}

	return result, nil
}
