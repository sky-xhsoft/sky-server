package live

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
	"go.uber.org/zap"
)

// StreamManager 流管理器
type StreamManager struct {
	client *live.Client
}

// NewStreamManager 创建流管理器
func NewStreamManager(client *live.Client) *StreamManager {
	return &StreamManager{
		client: client,
	}
}

// StreamOnlineInfo 在线流信息
type StreamOnlineInfo struct {
	DomainName  string   `json:"domainName"`  // 域名
	AppName     string   `json:"appName"`     // 应用名称
	StreamName  string   `json:"streamName"`  // 流名称
	PushToDelay int64    `json:"pushToDelay"` // 推流到延迟
	PublishTime []string `json:"publishTime"` // 推流时间列表
}

// DescribeStreamOnlineListRequest 查询在线流列表请求
type DescribeStreamOnlineListRequest struct {
	DomainName *string // 域名（可选）
	AppName    *string // 应用名称（可选）
	PageNum    *int64  // 页码（可选，默认1）
	PageSize   *int64  // 每页大小（可选，默认10）
}

// DescribeStreamOnlineListResponse 查询在线流列表响应
type DescribeStreamOnlineListResponse struct {
	OnlineInfo []*StreamOnlineInfo `json:"onlineInfo"` // 在线流列表
	TotalNum   int64               `json:"totalNum"`   // 总数
	TotalPage  int64               `json:"totalPage"`  // 总页数
	PageNum    int64               `json:"pageNum"`    // 当前页码
	PageSize   int64               `json:"pageSize"`   // 每页大小
}

// DescribeStreamOnlineList 查询在线流列表
func (m *StreamManager) DescribeStreamOnlineList(ctx context.Context, req *DescribeStreamOnlineListRequest) (*DescribeStreamOnlineListResponse, error) {
	request := live.NewDescribeLiveStreamOnlineListRequest()

	// 设置可选参数
	if req.DomainName != nil {
		request.DomainName = req.DomainName
	}
	if req.AppName != nil {
		request.AppName = req.AppName
	}
	if req.PageNum != nil {
		request.PageNum = common.Uint64Ptr(uint64(*req.PageNum))
	}
	if req.PageSize != nil {
		request.PageSize = common.Uint64Ptr(uint64(*req.PageSize))
	}

	response, err := m.client.DescribeLiveStreamOnlineList(request)
	if err != nil {
		logger.Error("Failed to describe stream online list",
			zap.Reflect("request", req),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to describe stream online list: %w", err)
	}

	// 转换响应数据
	result := &DescribeStreamOnlineListResponse{
		OnlineInfo: make([]*StreamOnlineInfo, 0),
		TotalNum:   0,
		TotalPage:  0,
		PageNum:    1,
		PageSize:   10,
	}

	if response.Response.TotalNum != nil {
		result.TotalNum = int64(*response.Response.TotalNum)
	}
	if response.Response.TotalPage != nil {
		result.TotalPage = int64(*response.Response.TotalPage)
	}
	if response.Response.PageNum != nil {
		result.PageNum = int64(*response.Response.PageNum)
	}
	if response.Response.PageSize != nil {
		result.PageSize = int64(*response.Response.PageSize)
	}

	// 转换在线流信息
	if response.Response.OnlineInfo != nil {
		for _, info := range response.Response.OnlineInfo {
			streamInfo := &StreamOnlineInfo{}

			if info.DomainName != nil {
				streamInfo.DomainName = *info.DomainName
			}
			if info.AppName != nil {
				streamInfo.AppName = *info.AppName
			}
			if info.StreamName != nil {
				streamInfo.StreamName = *info.StreamName
			}
			if info.PushToDelay != nil {
				streamInfo.PushToDelay = *info.PushToDelay
			}
			// 转换推流时间列表
			if info.PublishTimeList != nil {
				streamInfo.PublishTime = make([]string, 0, len(info.PublishTimeList))
				for _, pt := range info.PublishTimeList {
					if pt.PublishTime != nil {
						streamInfo.PublishTime = append(streamInfo.PublishTime, *pt.PublishTime)
					}
				}
			}

			result.OnlineInfo = append(result.OnlineInfo, streamInfo)
		}
	}

	return result, nil
}

// StreamHistoryInfo 历史流信息
type StreamHistoryInfo struct {
	DomainName      string `json:"domainName"`      // 域名
	AppName         string `json:"appName"`         // 应用名称
	StreamName      string `json:"streamName"`      // 流名称
	StreamStartTime string `json:"streamStartTime"` // 推流开始时间
	StreamEndTime   string `json:"streamEndTime"`   // 推流结束时间
	StopReason      string `json:"stopReason"`      // 停止原因
	Duration        int64  `json:"duration"`        // 推流时长（秒）
	Resolution      string `json:"resolution"`      // 分辨率
}

// DescribeStreamHistoryListRequest 查询历史流列表请求
type DescribeStreamHistoryListRequest struct {
	DomainName string  // 域名（必填）
	StartTime  string  // 开始时间（必填，格式：2019-01-01T00:00:00Z）
	EndTime    string  // 结束时间（必填，格式：2019-01-01T23:59:59Z）
	AppName    *string // 应用名称（可选）
	StreamName *string // 流名称（可选）
	PageNum    *int64  // 页码（可选，默认1）
	PageSize   *int64  // 每页大小（可选，默认10）
}

// DescribeStreamHistoryListResponse 查询历史流列表响应
type DescribeStreamHistoryListResponse struct {
	HistoryInfo []*StreamHistoryInfo `json:"historyInfo"` // 历史流列表
	TotalNum    int64                `json:"totalNum"`    // 总数
	TotalPage   int64                `json:"totalPage"`   // 总页数
	PageNum     int64                `json:"pageNum"`     // 当前页码
	PageSize    int64                `json:"pageSize"`    // 每页大小
}

// DescribeStreamHistoryList 查询历史流列表
func (m *StreamManager) DescribeStreamHistoryList(ctx context.Context, req *DescribeStreamHistoryListRequest) (*DescribeStreamHistoryListResponse, error) {
	request := live.NewDescribeLiveStreamPublishedListRequest()

	// 设置必填参数
	request.DomainName = common.StringPtr(req.DomainName)
	request.StartTime = common.StringPtr(req.StartTime)
	request.EndTime = common.StringPtr(req.EndTime)

	// 设置可选参数
	if req.AppName != nil {
		request.AppName = req.AppName
	}
	if req.StreamName != nil {
		request.StreamName = req.StreamName
	}
	if req.PageNum != nil {
		request.PageNum = common.Uint64Ptr(uint64(*req.PageNum))
	}
	if req.PageSize != nil {
		request.PageSize = common.Uint64Ptr(uint64(*req.PageSize))
	}

	response, err := m.client.DescribeLiveStreamPublishedList(request)
	if err != nil {
		logger.Error("Failed to describe stream history list",
			zap.Reflect("request", req),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to describe stream history list: %w", err)
	}

	// 转换响应数据
	result := &DescribeStreamHistoryListResponse{
		HistoryInfo: make([]*StreamHistoryInfo, 0),
		TotalNum:    0,
		TotalPage:   0,
		PageNum:     1,
		PageSize:    10,
	}

	if response.Response.TotalNum != nil {
		result.TotalNum = int64(*response.Response.TotalNum)
	}
	if response.Response.TotalPage != nil {
		result.TotalPage = int64(*response.Response.TotalPage)
	}
	if response.Response.PageNum != nil {
		result.PageNum = int64(*response.Response.PageNum)
	}
	if response.Response.PageSize != nil {
		result.PageSize = int64(*response.Response.PageSize)
	}

	// 转换历史流信息
	if response.Response.PublishInfo != nil {
		for _, info := range response.Response.PublishInfo {
			streamInfo := &StreamHistoryInfo{}

			if info.DomainName != nil {
				streamInfo.DomainName = *info.DomainName
			}
			if info.AppName != nil {
				streamInfo.AppName = *info.AppName
			}
			if info.StreamName != nil {
				streamInfo.StreamName = *info.StreamName
			}
			if info.StreamStartTime != nil {
				streamInfo.StreamStartTime = *info.StreamStartTime
			}
			if info.StreamEndTime != nil {
				streamInfo.StreamEndTime = *info.StreamEndTime
			}
			if info.StopReason != nil {
				streamInfo.StopReason = *info.StopReason
			}
			if info.Duration != nil {
				streamInfo.Duration = int64(*info.Duration)
			}
			if info.Resolution != nil {
				streamInfo.Resolution = *info.Resolution
			}

			result.HistoryInfo = append(result.HistoryInfo, streamInfo)
		}
	}

	return result, nil
}

// DropLiveStreamRequest 断开直播推流请求
type DropLiveStreamRequest struct {
	StreamName string // 流名称（必填）
	DomainName string // 推流域名（必填）
	AppName    string // 推流路径，默认为 live（必填）
}

// DropLiveStream 断开直播推流
func (m *StreamManager) DropLiveStream(ctx context.Context, req *DropLiveStreamRequest) error {
	request := live.NewDropLiveStreamRequest()
	request.StreamName = common.StringPtr(req.StreamName)
	request.DomainName = common.StringPtr(req.DomainName)
	request.AppName = common.StringPtr(req.AppName)

	_, err := m.client.DropLiveStream(request)
	if err != nil {
		logger.Error("Failed to drop live stream",
			zap.String("streamName", req.StreamName),
			zap.String("domainName", req.DomainName),
			zap.String("appName", req.AppName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to drop live stream: %w", err)
	}

	return nil
}

// StreamEventInfo 推断流事件信息
type StreamEventInfo struct {
	DomainName      string `json:"domainName"`      // 推流域名
	AppName         string `json:"appName"`         // 应用名称
	StreamName      string `json:"streamName"`      // 流名称
	StreamStartTime string `json:"streamStartTime"` // 推流开始时间
	StreamEndTime   string `json:"streamEndTime"`   // 推流结束时间
	StopReason      string `json:"stopReason"`      // 停止原因
	Duration        uint64 `json:"duration"`        // 推流持续时长（秒）
	ClientIp        string `json:"clientIp"`        // 主播IP
	Resolution      string `json:"resolution"`      // 分辨率
}

// DescribeStreamEventListRequest 查询推断流事件请求
type DescribeStreamEventListRequest struct {
	StartTime  string  // 起始时间（必填，UTC格式，例如：2018-12-29T19:00:00Z）
	EndTime    string  // 结束时间（必填，UTC格式，例如：2018-12-29T20:00:00Z）
	AppName    *string // 推流路径（可选）
	DomainName *string // 推流域名（可选）
	StreamName *string // 流名称（可选）
	PageNum    *int64  // 页码（可选，默认1）
	PageSize   *int64  // 每页大小（可选，默认10）
	IsFilter   *int64  // 是否过滤（可选，0-不过滤，1-只返回开播成功的）
}

// DescribeStreamEventListResponse 查询推断流事件响应
type DescribeStreamEventListResponse struct {
	EventList []*StreamEventInfo `json:"eventList"` // 推断流事件列表
	TotalNum  int64              `json:"totalNum"`  // 总数
	TotalPage int64              `json:"totalPage"` // 总页数
	PageNum   int64              `json:"pageNum"`   // 当前页码
	PageSize  int64              `json:"pageSize"`  // 每页大小
}

// DescribeStreamEventList 查询推断流事件
func (m *StreamManager) DescribeStreamEventList(ctx context.Context, req *DescribeStreamEventListRequest) (*DescribeStreamEventListResponse, error) {
	request := live.NewDescribeLiveStreamEventListRequest()
	request.StartTime = common.StringPtr(req.StartTime)
	request.EndTime = common.StringPtr(req.EndTime)

	// 设置可选参数
	if req.AppName != nil {
		request.AppName = req.AppName
	}
	if req.DomainName != nil {
		request.DomainName = req.DomainName
	}
	if req.StreamName != nil {
		request.StreamName = req.StreamName
	}
	if req.PageNum != nil {
		request.PageNum = common.Uint64Ptr(uint64(*req.PageNum))
	}
	if req.PageSize != nil {
		request.PageSize = common.Uint64Ptr(uint64(*req.PageSize))
	}
	if req.IsFilter != nil {
		request.IsFilter = req.IsFilter
	}

	response, err := m.client.DescribeLiveStreamEventList(request)
	if err != nil {
		logger.Error("Failed to describe stream event list",
			zap.Reflect("request", req),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to describe stream event list: %w", err)
	}

	// 转换响应数据
	result := &DescribeStreamEventListResponse{
		EventList: make([]*StreamEventInfo, 0),
		TotalNum:  0,
		TotalPage: 0,
		PageNum:   1,
		PageSize:  10,
	}

	if response.Response.TotalNum != nil {
		result.TotalNum = int64(*response.Response.TotalNum)
	}
	if response.Response.TotalPage != nil {
		result.TotalPage = int64(*response.Response.TotalPage)
	}
	if response.Response.PageNum != nil {
		result.PageNum = int64(*response.Response.PageNum)
	}
	if response.Response.PageSize != nil {
		result.PageSize = int64(*response.Response.PageSize)
	}

	// 转换事件信息
	if response.Response.EventList != nil {
		for _, event := range response.Response.EventList {
			eventInfo := &StreamEventInfo{}

			if event.DomainName != nil {
				eventInfo.DomainName = *event.DomainName
			}
			if event.AppName != nil {
				eventInfo.AppName = *event.AppName
			}
			if event.StreamName != nil {
				eventInfo.StreamName = *event.StreamName
			}
			if event.StreamStartTime != nil {
				eventInfo.StreamStartTime = *event.StreamStartTime
			}
			if event.StreamEndTime != nil {
				eventInfo.StreamEndTime = *event.StreamEndTime
			}
			if event.StopReason != nil {
				eventInfo.StopReason = *event.StopReason
			}
			if event.Duration != nil {
				eventInfo.Duration = *event.Duration
			}
			if event.ClientIp != nil {
				eventInfo.ClientIp = *event.ClientIp
			}
			if event.Resolution != nil {
				eventInfo.Resolution = *event.Resolution
			}

			result.EventList = append(result.EventList, eventInfo)
		}
	}

	return result, nil
}

// ResumeLiveStreamRequest 恢复直播推流请求
type ResumeLiveStreamRequest struct {
	StreamName string // 流名称（必填）
	DomainName string // 推流域名（必填）
	AppName    string // 推流路径，默认为 live（必填）
}

// ResumeLiveStream 恢复直播推流
func (m *StreamManager) ResumeLiveStream(ctx context.Context, req *ResumeLiveStreamRequest) error {
	request := live.NewResumeLiveStreamRequest()
	request.StreamName = common.StringPtr(req.StreamName)
	request.DomainName = common.StringPtr(req.DomainName)
	request.AppName = common.StringPtr(req.AppName)

	_, err := m.client.ResumeLiveStream(request)
	if err != nil {
		logger.Error("Failed to resume live stream",
			zap.String("streamName", req.StreamName),
			zap.String("domainName", req.DomainName),
			zap.String("appName", req.AppName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to resume live stream: %w", err)
	}

	return nil
}
