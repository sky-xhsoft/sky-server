package live

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
)

// RecordManager 录制管理器
type RecordManager struct {
	client *live.Client
}

// NewRecordManager 创建录制管理器
func NewRecordManager(client *live.Client) *RecordManager {
	return &RecordManager{
		client: client,
	}
}

// CreateRecordTemplateRequest 创建录制模板请求
type CreateRecordTemplateRequest struct {
	TemplateName string // 模板名称
	Description  string // 模板描述
	FlvParam     *RecordParam
	HlsParam     *RecordParam
	Mp4Param     *RecordParam
	AacParam     *RecordParam
}

// RecordParam 录制参数
type RecordParam struct {
	RecordInterval int64 // 录制间隔（秒）
	StorageTime    int64 // 存储时长（秒）
	Enable         int64 // 是否启用：0-禁用，1-启用
}

// CreateRecordTemplate 创建录制模板
func (m *RecordManager) CreateRecordTemplate(ctx context.Context, req *CreateRecordTemplateRequest) (int64, error) {
	request := live.NewCreateLiveRecordTemplateRequest()
	request.TemplateName = common.StringPtr(req.TemplateName)
	request.Description = common.StringPtr(req.Description)

	if req.FlvParam != nil {
		request.FlvParam = &live.RecordParam{
			RecordInterval: common.Int64Ptr(req.FlvParam.RecordInterval),
			StorageTime:    common.Int64Ptr(req.FlvParam.StorageTime),
			Enable:         common.Int64Ptr(req.FlvParam.Enable),
		}
	}

	if req.HlsParam != nil {
		request.HlsParam = &live.RecordParam{
			RecordInterval: common.Int64Ptr(req.HlsParam.RecordInterval),
			StorageTime:    common.Int64Ptr(req.HlsParam.StorageTime),
			Enable:         common.Int64Ptr(req.HlsParam.Enable),
		}
	}

	if req.Mp4Param != nil {
		request.Mp4Param = &live.RecordParam{
			RecordInterval: common.Int64Ptr(req.Mp4Param.RecordInterval),
			StorageTime:    common.Int64Ptr(req.Mp4Param.StorageTime),
			Enable:         common.Int64Ptr(req.Mp4Param.Enable),
		}
	}

	response, err := m.client.CreateLiveRecordTemplate(request)
	if err != nil {
		return 0, fmt.Errorf("failed to create record template: %w", err)
	}

	return *response.Response.TemplateId, nil
}

// DeleteRecordTemplate 删除录制模板
func (m *RecordManager) DeleteRecordTemplate(ctx context.Context, templateID int64) error {
	request := live.NewDeleteLiveRecordTemplateRequest()
	request.TemplateId = common.Int64Ptr(templateID)

	_, err := m.client.DeleteLiveRecordTemplate(request)
	if err != nil {
		return fmt.Errorf("failed to delete record template: %w", err)
	}

	return nil
}

// RecordTemplateInfo 录制模板信息
type RecordTemplateInfo struct {
	TemplateID   int64
	TemplateName string
	Description  string
	FlvParam     *RecordParam
	HlsParam     *RecordParam
	Mp4Param     *RecordParam
}

// DescribeRecordTemplates 查询录制模板列表
func (m *RecordManager) DescribeRecordTemplates(ctx context.Context) ([]*RecordTemplateInfo, error) {
	request := live.NewDescribeLiveRecordTemplatesRequest()

	response, err := m.client.DescribeLiveRecordTemplates(request)
	if err != nil {
		return nil, fmt.Errorf("failed to describe record templates: %w", err)
	}

	templates := make([]*RecordTemplateInfo, 0, len(response.Response.Templates))
	for _, tpl := range response.Response.Templates {
		info := &RecordTemplateInfo{
			TemplateID:   *tpl.TemplateId,
			TemplateName: *tpl.TemplateName,
			Description:  *tpl.Description,
		}

		if tpl.FlvParam != nil {
			info.FlvParam = &RecordParam{
				RecordInterval: *tpl.FlvParam.RecordInterval,
				StorageTime:    *tpl.FlvParam.StorageTime,
				Enable:         *tpl.FlvParam.Enable,
			}
		}

		if tpl.HlsParam != nil {
			info.HlsParam = &RecordParam{
				RecordInterval: *tpl.HlsParam.RecordInterval,
				StorageTime:    *tpl.HlsParam.StorageTime,
				Enable:         *tpl.HlsParam.Enable,
			}
		}

		if tpl.Mp4Param != nil {
			info.Mp4Param = &RecordParam{
				RecordInterval: *tpl.Mp4Param.RecordInterval,
				StorageTime:    *tpl.Mp4Param.StorageTime,
				Enable:         *tpl.Mp4Param.Enable,
			}
		}

		templates = append(templates, info)
	}

	return templates, nil
}

// CreateRecordRuleRequest 创建录制规则请求
type CreateRecordRuleRequest struct {
	DomainName string // 推流域名
	AppName    string // 推流应用名称
	StreamName string // 流名称
	TemplateID int64  // 模板ID
}

// CreateRecordRule 创建录制规则
func (m *RecordManager) CreateRecordRule(ctx context.Context, req *CreateRecordRuleRequest) error {
	request := live.NewCreateLiveRecordRuleRequest()
	request.DomainName = common.StringPtr(req.DomainName)
	request.AppName = common.StringPtr(req.AppName)
	request.StreamName = common.StringPtr(req.StreamName)
	request.TemplateId = common.Int64Ptr(req.TemplateID)

	_, err := m.client.CreateLiveRecordRule(request)
	if err != nil {
		return fmt.Errorf("failed to create record rule: %w", err)
	}

	return nil
}

// DeleteRecordRule 删除录制规则
func (m *RecordManager) DeleteRecordRule(ctx context.Context, domainName, appName, streamName string) error {
	request := live.NewDeleteLiveRecordRuleRequest()
	request.DomainName = common.StringPtr(domainName)
	request.AppName = common.StringPtr(appName)
	request.StreamName = common.StringPtr(streamName)

	_, err := m.client.DeleteLiveRecordRule(request)
	if err != nil {
		return fmt.Errorf("failed to delete record rule: %w", err)
	}

	return nil
}

// CreateRecordTaskRequest 创建录制任务请求
type CreateRecordTaskRequest struct {
	StreamName string // 流名称
	DomainName string // 推流域名
	AppName    string // 推流应用名称
	EndTime    int64  // 结束时间（Unix时间戳）
	StartTime  int64  // 开始时间（Unix时间戳）
	StreamType int64  // 流类型：0-直播流，1-混流
	TemplateID int64  // 模板ID
	Extension  string // 扩展字段
	Comment    string // 任务描述
}

// CreateRecordTask 创建录制任务
func (m *RecordManager) CreateRecordTask(ctx context.Context, req *CreateRecordTaskRequest) (string, error) {
	request := live.NewCreateRecordTaskRequest()
	request.StreamName = common.StringPtr(req.StreamName)
	request.DomainName = common.StringPtr(req.DomainName)
	request.AppName = common.StringPtr(req.AppName)
	request.EndTime = common.Uint64Ptr(uint64(req.EndTime))
	request.StartTime = common.Uint64Ptr(uint64(req.StartTime))
	request.StreamType = common.Uint64Ptr(uint64(req.StreamType))
	request.TemplateId = common.Uint64Ptr(uint64(req.TemplateID))

	response, err := m.client.CreateRecordTask(request)
	if err != nil {
		return "", fmt.Errorf("failed to create record task: %w", err)
	}

	return *response.Response.TaskId, nil
}

// StopRecordTask 停止录制任务
func (m *RecordManager) StopRecordTask(ctx context.Context, taskID string) error {
	request := live.NewStopRecordTaskRequest()
	request.TaskId = common.StringPtr(taskID)

	_, err := m.client.StopRecordTask(request)
	if err != nil {
		return fmt.Errorf("failed to stop record task: %w", err)
	}

	return nil
}

// DeleteRecordTask 删除录制任务
func (m *RecordManager) DeleteRecordTask(ctx context.Context, taskID string) error {
	request := live.NewDeleteRecordTaskRequest()
	request.TaskId = common.StringPtr(taskID)

	_, err := m.client.DeleteRecordTask(request)
	if err != nil {
		return fmt.Errorf("failed to delete record task: %w", err)
	}

	return nil
}
