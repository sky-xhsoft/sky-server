package volcengine

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// speech 相关常量
const (
	LiveAPIEndpoint   = "https://live.volcengineapi.com"
	SpeechTaskVersion = "2023-01-01"
)

// CreateSpeechTaskRequest 创建声影同传任务请求
// 对应火山引擎 CreateSpeechTask 接口
type CreateSpeechTaskRequest struct {
	Name     string          `json:"Name"`
	Rule     *SpeechTaskRule `json:"Rule"`
	RoomID   string          `json:"roomId,omitempty"`   // 本地直播间ID
	RoomName string          `json:"roomName,omitempty"` // 本地直播间名称
}

// SpeechTaskRule 声影同传任务规则（字段与火山控制台「新建任务」提交的参数一致）
// Pipe：si=声影同传；realtime=实时同传或原文 ASR 实时流式；mix=原文 ASR 整句
type SpeechTaskRule struct {
	Pipe      string              `json:"Pipe"`
	Delay     int                 `json:"Delay"`              // 播出延迟秒：声影同传 1-60，原文 ASR 整句 0-20，realtime 0-60（建议 2-3）
	Source    *SpeechTaskSource   `json:"Source"`             // 拉流源
	EndTime   string              `json:"EndTime"`            // 结束时间 ISO8601
	StartTime string              `json:"StartTime"`          // 开始时间 ISO8601
	Output    *SpeechTaskOutput   `json:"Output"`             // 转推输出
	SI        *SpeechTaskSI       `json:"SI,omitempty"`       // 声影同传：数字人口型
	RealTime  *SpeechTaskRealTime `json:"RealTime,omitempty"` // 同传声音、原声配置
	Extra     *SpeechTaskExtra    `json:"Extra,omitempty"`    // 降噪、音量均衡、原声音量
	Subtitle  *SpeechTaskSubtitle `json:"Subtitle,omitempty"` // 字幕参数（含关键词库、易错词、热词）
	Factory   *SpeechTaskFactory  `json:"Factory,omitempty"`  // 固定音色
	// 兼容旧版：词库放在 Rule 下（新版放在 Subtitle 下）
	GlossaryMap map[string]string `json:"GlossaryMap,omitempty"`
	ReplaceMap  map[string]string `json:"ReplaceMap,omitempty"`
	HotWords    []string          `json:"HotWords,omitempty"`
}

// SpeechTaskSource 拉流源
type SpeechTaskSource struct {
	Url string `json:"Url"`
}

// SpeechTaskOutput 转推输出
type SpeechTaskOutput struct {
	Url []string `json:"Url"`
}

// SpeechTaskSI 声影同传数字人参数
type SpeechTaskSI struct {
	LipSync bool `json:"LipSync"`
}

// SpeechTaskRealTime 同传声音与原声配置
type SpeechTaskRealTime struct {
	VoiceShow bool  `json:"VoiceShow"`           // 同传声音
	Separator *bool `json:"Separator,omitempty"` // 实时同传：true 仅保留翻译音频，false 原声自动叠加翻译
}

// SpeechTaskExtra 扩展参数
type SpeechTaskExtra struct {
	AMix      *SpeechTaskAMix      `json:"AMix,omitempty"`
	Separator *SpeechTaskSeparator `json:"Separator,omitempty"`
}

// SpeechTaskAMix 音频混合
type SpeechTaskAMix struct {
	BGVolume     *float64 `json:"BGVolume,omitempty"` // 原声音量 / 背景音（0-1）
	AudioEnhance bool     `json:"AudioEnhance"`       // 音量均衡
}

// SpeechTaskSeparator 降噪
type SpeechTaskSeparator struct {
	HandleType int  `json:"HandleType"` // 1 开启降噪，0 关闭
	Separator  bool `json:"Separator,omitempty"`
}

// SpeechTaskFactory 固定音色（火山技术支持提供，未公开）：{"ASR":"sail","TTS":"minimax","SID":"<SpeakerID>"}；自动复制时 TTS/SID 为空
type SpeechTaskFactory struct {
	ASR string `json:"ASR,omitempty"`
	TTS string `json:"TTS"`
	SID string `json:"SID"`
}

// SpeechTaskSubtitle 字幕参数
type SpeechTaskSubtitle struct {
	Position    *SpeechTaskSubtitlePosition `json:"Position,omitempty"`
	Original    *SpeechTaskSubtitleStyle    `json:"Original,omitempty"`    // 原文样式
	Translation []*SpeechTaskSubtitleStyle  `json:"Translation"`           // 译文样式列表（原文 ASR 为空数组）
	GlossaryMap map[string]string           `json:"GlossaryMap,omitempty"` // 关键词库：原文 → 译文
	ReplaceMap  map[string]string           `json:"ReplaceMap,omitempty"`  // 易错词：译文 → 纠正
	HotWords    []string                    `json:"HotWords,omitempty"`    // 热词库（原文 ASR）
}

// SpeechTaskSubtitlePosition 字幕位置（边距为 0-1 的比例）
type SpeechTaskSubtitlePosition struct {
	Relative string  `json:"Relative,omitempty"`
	MarginTb float64 `json:"MarginTb"` // 底部边距
	MarginLr float64 `json:"MarginLr"` // 左右边距
}

// SpeechTaskSubtitleStyle 字幕样式
type SpeechTaskSubtitleStyle struct {
	Language      string                 `json:"Language"`
	Font          string                 `json:"Font,omitempty"`
	FontSize      int                    `json:"FontSize,omitempty"`
	FontColor     string                 `json:"FontColor,omitempty"`
	Alignment     int                    `json:"Alignment,omitempty"`
	MaxRowNumber  int                    `json:"MaxRowNumber"` // 0 为自适应
	MaxCharNumber int                    `json:"MaxCharNumber,omitempty"`
	Hidden        bool                   `json:"Hidden"`
	ShowPriority  int                    `json:"ShowPriority"` // 原文在上：原文 100、译文 0；原文在下相反
	Box           *SpeechTaskSubtitleBox `json:"Box,omitempty"`
	Border        *SpeechTaskSubtitleBox `json:"Border,omitempty"`
}

// SpeechTaskSubtitleBox 字幕底色 / 描边：颜色格式 #RRGGBB@透明度
type SpeechTaskSubtitleBox struct {
	W      float64  `json:"W,omitempty"`
	Color  string   `json:"Color,omitempty"`
	Filter *float64 `json:"Filter,omitempty"` // 底色透明度（0-1）
}

// CreateSpeechTaskResponse 创建声影同传任务响应
type CreateSpeechTaskResponse struct {
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
		Action    string `json:"Action"`
		Version   string `json:"Version"`
		Service   string `json:"Service"`
		Region    string `json:"Region"`
	} `json:"ResponseMetadata"`
	Result *CreateSpeechTaskResult `json:"Result"`
}

// CreateSpeechTaskResult 创建结果（火山返回 Result.Data.TaskID 嵌套结构）
type CreateSpeechTaskResult struct {
	Code    int                   `json:"Code"`
	Message string                `json:"Message"`
	Data    *CreateSpeechTaskData `json:"Data"`
}

// CreateSpeechTaskData 创建任务返回数据
type CreateSpeechTaskData struct {
	TaskID   string `json:"TaskID"`
	TaskName string `json:"TaskName"`
}

// DeleteSpeechTaskRequest 删除声影同传任务请求
type DeleteSpeechTaskRequest struct {
	TaskID     string `json:"TaskID"`
	WithRecord bool   `json:"WithRecord,omitempty"`
}

// DeleteSpeechTaskResponse 删除响应
type DeleteSpeechTaskResponse struct {
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
	} `json:"ResponseMetadata"`
	Result *DeleteSpeechTaskResult `json:"Result"`
}

// DeleteSpeechTaskResult 删除结果
type DeleteSpeechTaskResult struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

// CreateSpeechTask 创建声影同传任务
func (c *Client) CreateSpeechTask(ctx context.Context, req *CreateSpeechTaskRequest) (string, error) {
	// sanitize: 火山仅接受 GetSpeechConfig 返回的有效字体 Key，SimHei 等非法值会导致 params invalid，
	// 因此剥离非白名单字体，让火山使用默认字体 arial。
	const validFont = "arial"
	if req.Rule != nil && req.Rule.Subtitle != nil {
		sanitizeFont := func(st *SpeechTaskSubtitleStyle) {
			if st == nil {
				return
			}
			if st.Font != "" && st.Font != validFont {
				st.Font = ""
			}
		}
		sanitizeFont(req.Rule.Subtitle.Original)
		for _, t := range req.Rule.Subtitle.Translation {
			sanitizeFont(t)
		}
	}
	url := fmt.Sprintf("%s?Action=CreateSpeechTask&Version=%s", LiveAPIEndpoint, SpeechTaskVersion)
	var resp CreateSpeechTaskResponse
	if err := c.PostRequest(ctx, url, req, &resp); err != nil {
		return "", err
	}
	if resp.Result == nil {
		return "", fmt.Errorf("创建声影同传任务失败: 响应为空")
	}
	taskID := ""
	if resp.Result.Data != nil {
		taskID = resp.Result.Data.TaskID
	}
	if taskID == "" {
		// 失败：Code != 0 或 Message 非 success；记录火山完整响应（含 RequestId，提工单用）
		raw, _ := json.Marshal(&resp)
		zap.L().Error("火山 CreateSpeechTask 失败响应", zap.ByteString("response", raw))
		meta := fmt.Sprintf("Code=%d, RequestId=%s", resp.Result.Code, resp.ResponseMetadata.RequestID)
		if resp.Result.Message != "" && resp.Result.Message != "success" {
			return "", fmt.Errorf("创建声影同传任务失败: %s（%s）", resp.Result.Message, meta)
		}
		return "", fmt.Errorf("创建声影同传任务失败: 未返回任务ID（%s）", meta)
	}
	return taskID, nil
}

// DeleteSpeechTask 删除声影同传任务
func (c *Client) DeleteSpeechTask(ctx context.Context, req *DeleteSpeechTaskRequest) error {
	url := fmt.Sprintf("%s?Action=DeleteSpeechTask&Version=%s", LiveAPIEndpoint, SpeechTaskVersion)
	var resp DeleteSpeechTaskResponse
	if err := c.PostRequest(ctx, url, req, &resp); err != nil {
		return err
	}
	if resp.Result != nil && resp.Result.Code != 0 {
		return fmt.Errorf("删除声影同传任务失败: code=%d msg=%s", resp.Result.Code, resp.Result.Message)
	}
	return nil
}

// SpeechTaskInfo 火山上的同传任务信息（GetSpeechTask / SearchSpeechTask）
// Status：init 准备中、prepare 启动中、running 运行中、prestop 关闭中、done 已结束
type SpeechTaskInfo struct {
	TaskID    string `json:"TaskID"`
	Name      string `json:"Name"`
	Status    string `json:"Status"`
	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
	StoppedAt string `json:"StoppedAt"`
	Rule      string `json:"Rule"`
}

type speechTaskResult[T any] struct {
	Result *struct {
		Code    int    `json:"Code"`
		Message string `json:"Message"`
		Data    T      `json:"Data"`
	} `json:"Result"`
}

func (r *speechTaskResult[T]) check(action string) error {
	if r.Result == nil {
		return fmt.Errorf("%s 失败: 响应为空", action)
	}
	if r.Result.Code != 0 {
		return fmt.Errorf("%s 失败: code=%d msg=%s", action, r.Result.Code, r.Result.Message)
	}
	return nil
}

// GetSpeechTask 查询单个同传任务
func (c *Client) GetSpeechTask(ctx context.Context, taskID string) (*SpeechTaskInfo, error) {
	url := fmt.Sprintf("%s?Action=GetSpeechTask&Version=%s", LiveAPIEndpoint, SpeechTaskVersion)
	var resp speechTaskResult[SpeechTaskInfo]
	if err := c.PostRequest(ctx, url, map[string]any{"TaskID": taskID}, &resp); err != nil {
		return nil, err
	}
	if err := resp.check("查询声影同传任务"); err != nil {
		return nil, err
	}
	return &resp.Result.Data, nil
}

// SearchSpeechTaskRequest 搜索同传任务（CreateTimeSt / CreateTimeEd 为东八区 "2006-01-02 15:04:05"）
type SearchSpeechTaskRequest struct {
	TaskStatus   []string `json:"TaskStatus,omitempty"`
	CreateTimeSt string   `json:"CreateTimeSt,omitempty"`
	CreateTimeEd string   `json:"CreateTimeEd,omitempty"`
	Page         int      `json:"Page"`
	PageSize     int      `json:"PageSize"`
}

// SearchSpeechTask 搜索同传任务，返回总数与本页任务
func (c *Client) SearchSpeechTask(ctx context.Context, req *SearchSpeechTaskRequest) (int, []SpeechTaskInfo, error) {
	url := fmt.Sprintf("%s?Action=SearchSpeechTask&Version=%s", LiveAPIEndpoint, SpeechTaskVersion)
	var resp speechTaskResult[struct {
		Count int              `json:"Count"`
		Tasks []SpeechTaskInfo `json:"Tasks"`
	}]
	if err := c.PostRequest(ctx, url, req, &resp); err != nil {
		return 0, nil, err
	}
	if err := resp.check("搜索声影同传任务"); err != nil {
		return 0, nil, err
	}
	return resp.Result.Data.Count, resp.Result.Data.Tasks, nil
}

// RestartSpeechTask 重启运行中或启动中的同传任务；reloadSubtitle 为 true 时不中断直播，仅重启字幕算法模块
func (c *Client) RestartSpeechTask(ctx context.Context, taskID string, reloadSubtitle bool) error {
	url := fmt.Sprintf("%s?Action=RestartSpeechTask&Version=%s", LiveAPIEndpoint, SpeechTaskVersion)
	body := map[string]any{"TaskID": taskID}
	if reloadSubtitle {
		body["SetSiStatus"] = "reload"
	}
	var resp speechTaskResult[json.RawMessage]
	if err := c.PostRequest(ctx, url, body, &resp); err != nil {
		return err
	}
	return resp.check("重启声影同传任务")
}
