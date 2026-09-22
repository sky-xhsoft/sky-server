package volcengine

import (
	"context"
	"fmt"
)

// speech 相关常量
const (
	LiveAPIEndpoint   = "https://live.volcengineapi.com"
	SpeechTaskVersion = "2023-01-01"
)

// CreateSpeechTaskRequest 创建声影同传任务请求
// 对应火山引擎 CreateSpeechTask 接口
type CreateSpeechTaskRequest struct {
	Name string               `json:"Name"`
	Rule *SpeechTaskRule      `json:"Rule"`
	RoomID    string          `json:"roomId,omitempty"` // 本地直播间ID
	RoomName  string          `json:"roomName,omitempty"` // 本地直播间名称
}

// SpeechTaskRule 声影同传任务规则
type SpeechTaskRule struct {
	Pipe      string                `json:"Pipe"`      // si=声影同传, mix=仅字幕, realtime=实时同传
	Delay     int                   `json:"Delay,omitempty"`     // 延迟秒(声影同传模式)
	Source    *SpeechTaskSource     `json:"Source"`              // 拉流源
	EndTime   string                `json:"EndTime"`             // 结束时间 ISO8601
	StartTime string                `json:"StartTime"`           // 开始时间 ISO8601
	Output    *SpeechTaskOutput     `json:"Output"`              // 转推输出
	SI        *SpeechTaskSI         `json:"SI,omitempty"`        // 声影模式数字人参数
	RealTime  *SpeechTaskRealTime   `json:"RealTime,omitempty"`  // 实时模式参数
	Extra     *SpeechTaskExtra      `json:"Extra,omitempty"`     // 扩展参数
	Subtitle  *SpeechTaskSubtitle   `json:"Subtitle,omitempty"`  // 字幕参数
	GlossaryMap map[string]string   `json:"GlossaryMap,omitempty"` // 术语库
	ReplaceMap  map[string]string   `json:"ReplaceMap,omitempty"`  // 纠错词
	HotWords    []string            `json:"HotWords,omitempty"`    // 热词
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

// SpeechTaskRealTime 实时同传参数
type SpeechTaskRealTime struct {
	VoiceShow bool `json:"VoiceShow"`
}

// SpeechTaskExtra 扩展参数
type SpeechTaskExtra struct {
	AMix      *SpeechTaskAMix               `json:"AMix,omitempty"`
	Separator *SpeechTaskSeparator           `json:"Separator,omitempty"`
}

// SpeechTaskAMix 音频混合
type SpeechTaskAMix struct {
	BGVolume     float64 `json:"BGVolume,omitempty"`
	AudioEnhance bool    `json:"AudioEnhance,omitempty"`
}

// SpeechTaskSeparator 声道分离
type SpeechTaskSeparator struct {
	HandleType int  `json:"HandleType,omitempty"`
	Separator  bool `json:"Separator,omitempty"`
}

// SpeechTaskSubtitle 字幕参数
type SpeechTaskSubtitle struct {
	Position    *SpeechTaskSubtitlePosition `json:"Position,omitempty"`
	Original    *SpeechTaskSubtitleStyle    `json:"Original,omitempty"`    // 原文样式
	Translation []*SpeechTaskSubtitleStyle  `json:"Translation,omitempty"` // 译文样式列表
}

// SpeechTaskSubtitlePosition 字幕位置
type SpeechTaskSubtitlePosition struct {
	Relative string  `json:"Relative,omitempty"` // bottom/top
	MarginTb float64 `json:"MarginTb,omitempty"` // 上下边距
	MarginLr float64 `json:"MarginLr,omitempty"` // 左右边距
}

// SpeechTaskSubtitleStyle 字幕样式
type SpeechTaskSubtitleStyle struct {
	Language      string                  `json:"Language"`
	Font          string                  `json:"Font,omitempty"`
	FontSize      int                     `json:"FontSize,omitempty"`
	FontColor     string                  `json:"FontColor,omitempty"`
	Alignment     int                     `json:"Alignment,omitempty"`
	MaxRowNumber  int                     `json:"MaxRowNumber,omitempty"`
	MaxCharNumber int                     `json:"MaxCharNumber,omitempty"`
	Hidden        bool                    `json:"Hidden,omitempty"`
	ShowPriority  int                     `json:"ShowPriority,omitempty"`
	Box           *SpeechTaskSubtitleBox  `json:"Box,omitempty"`
	Border        *SpeechTaskSubtitleBox  `json:"Border,omitempty"`
}

// SpeechTaskSubtitleBox 字幕框/边框
type SpeechTaskSubtitleBox struct {
	W     float64 `json:"W,omitempty"`
	Color string  `json:"Color,omitempty"`
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
	Code    int                    `json:"Code"`
	Message string                 `json:"Message"`
	Data    *CreateSpeechTaskData  `json:"Data"`
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
		// 失败：Code != 0 或 Message 非 success
		if resp.Result.Message != "" && resp.Result.Message != "success" {
			return "", fmt.Errorf("创建声影同传任务失败: %s", resp.Result.Message)
		}
		return "", fmt.Errorf("创建声影同传任务失败: 未返回任务ID")
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