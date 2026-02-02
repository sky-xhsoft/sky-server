package live

import (
	"context"
	"fmt"
)

// AuditManager 审核管理器
// 注意：腾讯云直播审核 API 可能需要单独申请开通
type AuditManager struct {
	// client *live.Client
}

// NewAuditManager 创建审核管理器
func NewAuditManager() *AuditManager {
	return &AuditManager{}
}

// CreateKeywordLibRequest 创建关键词库请求
type CreateKeywordLibRequest struct {
	LibName string // 词库名称
	LibType int64  // 词库类型：1-黑名单，2-白名单
}

// CreateKeywordLib 创建关键词库
// 注意：此功能需要联系腾讯云开通审核功能
func (m *AuditManager) CreateKeywordLib(ctx context.Context, req *CreateKeywordLibRequest) (int64, error) {
	return 0, fmt.Errorf("audit feature not implemented: please contact Tencent Cloud to enable audit service")
}

// CreateKeywordsRequest 创建关键词请求
type CreateKeywordsRequest struct {
	LibID    int64    // 词库ID
	Keywords []string // 关键词列表
}

// CreateKeywords 创建关键词
func (m *AuditManager) CreateKeywords(ctx context.Context, req *CreateKeywordsRequest) error {
	return fmt.Errorf("audit feature not implemented: please contact Tencent Cloud to enable audit service")
}

// DeleteKeywords 删除关键词
func (m *AuditManager) DeleteKeywords(ctx context.Context, libID int64, keywords []string) error {
	return fmt.Errorf("audit feature not implemented: please contact Tencent Cloud to enable audit service")
}

// KeywordInfo 关键词信息
type KeywordInfo struct {
	Keyword    string
	CreateTime string
}

// DescribeKeywords 查询关键词列表
func (m *AuditManager) DescribeKeywords(ctx context.Context, libID int64) ([]*KeywordInfo, error) {
	return nil, fmt.Errorf("audit feature not implemented: please contact Tencent Cloud to enable audit service")
}
