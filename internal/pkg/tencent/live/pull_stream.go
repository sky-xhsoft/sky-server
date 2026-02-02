package live

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
)

// RegionClientProvider 地域客户端提供者接口
type RegionClientProvider interface {
	GetLiveClientForRegion(region string) (*live.Client, error)
}

// PullStreamManager 拉流管理器
type PullStreamManager struct {
	client         *live.Client
	clientProvider RegionClientProvider
}

// NewPullStreamManager 创建拉流管理器
func NewPullStreamManager(client *live.Client) *PullStreamManager {
	return &PullStreamManager{
		client: client,
	}
}

// NewPullStreamManagerWithProvider 创建支持多地域的拉流管理器
func NewPullStreamManagerWithProvider(client *live.Client, provider RegionClientProvider) *PullStreamManager {
	return &PullStreamManager{
		client:         client,
		clientProvider: provider,
	}
}

// CreatePullStreamTaskRequest 创建拉流任务请求
type CreatePullStreamTaskRequest struct {
	SourceType string   // 拉流源类型：PullLivePushLive（直播），PullVodPushLive（点播）
	SourceURLs []string // 拉流源URL列表
	DomainName string   // 推流域名
	AppName    string   // 推流应用名称
	StreamName string   // 推流流名称
	StartTime  string   // 开始时间，格式：yyyy-mm-dd HH:MM:SS
	EndTime    string   // 结束时间，格式：yyyy-mm-dd HH:MM:SS
	Operator   string   // 操作者
	Comment    string   // 任务描述
	Region     string   // 任务创建所在地域
}

// CreatePullStreamTask 创建拉流任务
func (m *PullStreamManager) CreatePullStreamTask(ctx context.Context, req *CreatePullStreamTaskRequest) (string, error) {
	// 根据地域选择客户端
	client := m.client
	if req.Region != "" && m.clientProvider != nil {
		regionClient, err := m.clientProvider.GetLiveClientForRegion(req.Region)
		if err != nil {
			return "", fmt.Errorf("failed to get client for region %s: %w", req.Region, err)
		}
		client = regionClient
	}

	request := live.NewCreateLivePullStreamTaskRequest()
	request.SourceType = common.StringPtr(req.SourceType)

	sourceURLs := make([]*string, len(req.SourceURLs))
	for i, url := range req.SourceURLs {
		sourceURLs[i] = common.StringPtr(url)
	}
	request.SourceUrls = sourceURLs

	request.DomainName = common.StringPtr(req.DomainName)
	request.AppName = common.StringPtr(req.AppName)
	request.StreamName = common.StringPtr(req.StreamName)
	request.StartTime = common.StringPtr(req.StartTime)
	request.EndTime = common.StringPtr(req.EndTime)
	request.Operator = common.StringPtr(req.Operator)
	request.Comment = common.StringPtr(req.Comment)

	response, err := client.CreateLivePullStreamTask(request)
	if err != nil {
		return "", fmt.Errorf("failed to create pull stream task: %w", err)
	}

	return *response.Response.TaskId, nil
}

// DeletePullStreamTask 删除拉流任务
func (m *PullStreamManager) DeletePullStreamTask(ctx context.Context, taskID string, operator string) error {
	request := live.NewDeleteLivePullStreamTaskRequest()
	request.TaskId = common.StringPtr(taskID)
	request.Operator = common.StringPtr(operator)

	_, err := m.client.DeleteLivePullStreamTask(request)
	if err != nil {
		return fmt.Errorf("failed to delete pull stream task: %w", err)
	}

	return nil
}

// PullStreamTaskInfo 拉流任务信息
type PullStreamTaskInfo struct {
	TaskID                string   `json:"taskId"`
	SourceType            string   `json:"sourceType"`
	SourceURLs            []string `json:"sourceUrls"`
	DomainName            string   `json:"domainName"`
	AppName               string   `json:"appName"`
	StreamName            string   `json:"streamName"`
	PushArgs              string   `json:"pushArgs,omitempty"`
	StartTime             string   `json:"startTime"`
	EndTime               string   `json:"endTime"`
	Region                string   `json:"region,omitempty"`
	VodLoopTimes          int64    `json:"vodLoopTimes,omitempty"`
	VodRefreshType        string   `json:"vodRefreshType,omitempty"`
	CreateTime            string   `json:"createTime"`
	UpdateTime            string   `json:"updateTime,omitempty"`
	CreateBy              string   `json:"createBy,omitempty"`
	UpdateBy              string   `json:"updateBy,omitempty"`
	CallbackUrl           string   `json:"callbackUrl,omitempty"`
	CallbackEvents        []string `json:"callbackEvents,omitempty"`
	CallbackInfo          string   `json:"callbackInfo,omitempty"`
	ErrorInfo             string   `json:"errorInfo,omitempty"`
	Status                string   `json:"status"`
	Comment               string   `json:"comment,omitempty"`
	BackupSourceType      string   `json:"backupSourceType,omitempty"`
	BackupSourceUrl       string   `json:"backupSourceUrl,omitempty"`
	VodLocalMode          int64    `json:"vodLocalMode,omitempty"`
	RecordTemplateId      string   `json:"recordTemplateId,omitempty"`
	BackupToUrl           string   `json:"backupToUrl,omitempty"`
	TranscodeTemplateName string   `json:"transcodeTemplateName,omitempty"`
}

// DescribePullStreamTasks 查询拉流任务列表
func (m *PullStreamManager) DescribePullStreamTasks(ctx context.Context, taskID *string) ([]*PullStreamTaskInfo, error) {
	request := live.NewDescribeLivePullStreamTasksRequest()
	if taskID != nil {
		request.TaskId = common.StringPtr(*taskID)
	}

	response, err := m.client.DescribeLivePullStreamTasks(request)
	if err != nil {
		return nil, fmt.Errorf("failed to describe pull stream tasks: %w", err)
	}

	tasks := make([]*PullStreamTaskInfo, 0, len(response.Response.TaskInfos))
	for _, task := range response.Response.TaskInfos {
		sourceURLs := make([]string, len(task.SourceUrls))
		for i, url := range task.SourceUrls {
			sourceURLs[i] = *url
		}

		info := &PullStreamTaskInfo{
			TaskID:     *task.TaskId,
			SourceType: *task.SourceType,
			SourceURLs: sourceURLs,
			DomainName: *task.DomainName,
			AppName:    *task.AppName,
			StreamName: *task.StreamName,
			StartTime:  *task.StartTime,
			EndTime:    *task.EndTime,
			Status:     *task.Status,
			CreateTime: *task.CreateTime,
		}

		// 可选字段
		if task.PushArgs != nil {
			info.PushArgs = *task.PushArgs
		}
		if task.Region != nil {
			info.Region = *task.Region
		}
		if task.VodLoopTimes != nil {
			info.VodLoopTimes = *task.VodLoopTimes
		}
		if task.VodRefreshType != nil {
			info.VodRefreshType = *task.VodRefreshType
		}
		if task.UpdateTime != nil {
			info.UpdateTime = *task.UpdateTime
		}
		if task.CreateBy != nil {
			info.CreateBy = *task.CreateBy
		}
		if task.UpdateBy != nil {
			info.UpdateBy = *task.UpdateBy
		}
		if task.CallbackUrl != nil {
			info.CallbackUrl = *task.CallbackUrl
		}
		if task.CallbackEvents != nil {
			callbackEvents := make([]string, len(task.CallbackEvents))
			for i, event := range task.CallbackEvents {
				callbackEvents[i] = *event
			}
			info.CallbackEvents = callbackEvents
		}
		if task.CallbackInfo != nil {
			info.CallbackInfo = *task.CallbackInfo
		}
		if task.ErrorInfo != nil {
			info.ErrorInfo = *task.ErrorInfo
		}
		if task.Comment != nil {
			info.Comment = *task.Comment
		}
		if task.BackupSourceType != nil {
			info.BackupSourceType = *task.BackupSourceType
		}
		if task.BackupSourceUrl != nil {
			info.BackupSourceUrl = *task.BackupSourceUrl
		}
		if task.VodLocalMode != nil {
			info.VodLocalMode = *task.VodLocalMode
		}
		if task.RecordTemplateId != nil {
			info.RecordTemplateId = *task.RecordTemplateId
		}
		if task.BackupToUrl != nil {
			info.BackupToUrl = *task.BackupToUrl
		}
		if task.TranscodeTemplateName != nil {
			info.TranscodeTemplateName = *task.TranscodeTemplateName
		}

		tasks = append(tasks, info)
	}

	return tasks, nil
}

// RestartPullStreamTask 重启拉流任务
func (m *PullStreamManager) RestartPullStreamTask(ctx context.Context, taskID string, operator string) error {
	request := live.NewRestartLivePullStreamTaskRequest()
	request.TaskId = common.StringPtr(taskID)
	request.Operator = common.StringPtr(operator)

	_, err := m.client.RestartLivePullStreamTask(request)
	if err != nil {
		return fmt.Errorf("failed to restart pull stream task: %w", err)
	}

	return nil
}

// UpdatePullStreamTaskRequest 更新拉流任务请求
type UpdatePullStreamTaskRequest struct {
	TaskID     string   // 任务ID
	SourceType string   // 拉流源类型
	SourceURLs []string // 拉流源URL列表
	StartTime  string   // 开始时间
	EndTime    string   // 结束时间
	Operator   string   // 操作者
	Comment    string   // 任务描述
	Status     string   // 任务状态：enable-启用，pause-暂停
}

// UpdatePullStreamTask 更新拉流任务
func (m *PullStreamManager) UpdatePullStreamTask(ctx context.Context, req *UpdatePullStreamTaskRequest) error {
	request := live.NewModifyLivePullStreamTaskRequest()
	request.TaskId = common.StringPtr(req.TaskID)
	request.Operator = common.StringPtr(req.Operator)

	if len(req.SourceURLs) > 0 {
		sourceURLs := make([]*string, len(req.SourceURLs))
		for i, url := range req.SourceURLs {
			sourceURLs[i] = common.StringPtr(url)
		}
		request.SourceUrls = sourceURLs
	}

	if req.StartTime != "" {
		request.StartTime = common.StringPtr(req.StartTime)
	}

	if req.EndTime != "" {
		request.EndTime = common.StringPtr(req.EndTime)
	}

	if req.Status != "" {
		request.Status = common.StringPtr(req.Status)
	}

	_, err := m.client.ModifyLivePullStreamTask(request)
	if err != nil {
		return fmt.Errorf("failed to update pull stream task: %w", err)
	}

	return nil
}

// PullStreamTaskStatus 拉流任务状态
type PullStreamTaskStatus struct {
	TaskID      string `json:"taskId"`
	RunStatus   string `json:"runStatus"`
	FileURL     string `json:"fileUrl,omitempty"`
	LoopedTimes int64  `json:"loopedTimes,omitempty"`
	OffsetTime  int64  `json:"offsetTime,omitempty"`
	ReportTime  string `json:"reportTime,omitempty"`
}

// DescribePullStreamTaskStatus 查询拉流任务状态
func (m *PullStreamManager) DescribePullStreamTaskStatus(ctx context.Context, taskID string) (*PullStreamTaskStatus, error) {
	request := live.NewDescribeLivePullStreamTaskStatusRequest()
	request.TaskId = common.StringPtr(taskID)

	response, err := m.client.DescribeLivePullStreamTaskStatus(request)
	if err != nil {
		return nil, fmt.Errorf("failed to describe pull stream task status: %w", err)
	}

	status := &PullStreamTaskStatus{
		TaskID: taskID,
	}

	if response.Response.TaskStatusInfo != nil {
		info := response.Response.TaskStatusInfo

		if info.RunStatus != nil {
			status.RunStatus = *info.RunStatus
		}

		if info.FileUrl != nil {
			status.FileURL = *info.FileUrl
		}

		if info.LoopedTimes != nil {
			status.LoopedTimes = *info.LoopedTimes
		}

		if info.OffsetTime != nil {
			status.OffsetTime = *info.OffsetTime
		}

		if info.ReportTime != nil {
			status.ReportTime = *info.ReportTime
		}
	}

	return status, nil
}
