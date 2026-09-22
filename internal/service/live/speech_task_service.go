package live

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/volcengine"
	"gorm.io/gorm"
)

// SpeechTaskRecord 本地声影同传任务记录(落库)
// 注意: live_speech_task 表列名为大写下划线, 需显式 gorm column 映射
type SpeechTaskRecord struct {
	ID          uint   `gorm:"column:ID;primaryKey" json:"id"`
	CompanyID   uint   `gorm:"column:SYS_COMPANY_ID" json:"-"`
	RoomID      uint   `gorm:"column:ROOM_ID" json:"roomId"`
	RoomName    string `gorm:"column:ROOM_NAME" json:"roomName"`
	TaskID      string `gorm:"column:TASK_ID" json:"taskId"`
	Name        string `gorm:"column:NAME" json:"name"`
	Mode        string `gorm:"column:MODE" json:"mode"`
	SourceURL   string `gorm:"column:SOURCE_URL" json:"sourceUrl"`
	TargetURLs  string `gorm:"column:TARGET_URLS" json:"targetUrls"`
	OutputLangs string `gorm:"column:OUTPUT_LANGS" json:"outputLangs"`
	Subtitle    bool   `gorm:"column:SUBTITLE" json:"subtitle"`
	StartTime   string `gorm:"column:START_TIME" json:"startTime"`
	EndTime     string `gorm:"column:END_TIME" json:"endTime"`
	Status      string `gorm:"column:STATUS" json:"status"`
	RuleJSON    string `gorm:"column:RULE_JSON" json:"-"`
	CreateTime  string `gorm:"column:CREATE_TIME" json:"createTime"`
}

// TableName 指定表名
func (SpeechTaskRecord) TableName() string {
	return "live_speech_task"
}

// SpeechTaskService 声影同传任务服务(对接火山引擎 CreateSpeechTask / DeleteSpeechTask)
type SpeechTaskService interface {
	// CreateSpeechTask 创建声影同传任务, 返回火山任务ID
	CreateSpeechTask(ctx context.Context, companyID uint, req *volcengine.CreateSpeechTaskRequest) (string, error)
	// DeleteSpeechTask 删除声影同传任务
	DeleteSpeechTask(ctx context.Context, companyID uint, req *volcengine.DeleteSpeechTaskRequest) error
	// ListSpeechTasks 查询本地声影同传任务列表(按公司+可选直播间)
	ListSpeechTasks(ctx context.Context, companyID, roomID uint) ([]SpeechTaskRecord, error)
}

// speechTaskService 声影同传任务服务实现
type speechTaskService struct {
	volcengineMgr *volcengine.CompanyVolcengineManager
	db            *gorm.DB
}

// NewSpeechTaskService 创建声影同传任务服务
func NewSpeechTaskService(mgr *volcengine.CompanyVolcengineManager, db *gorm.DB) SpeechTaskService {
	return &speechTaskService{volcengineMgr: mgr, db: db}
}

// getClient 获取公司级火山引擎客户端
func (s *speechTaskService) getClient(ctx context.Context, companyID uint) (*volcengine.Client, error) {
	if s.volcengineMgr == nil {
		return nil, fmt.Errorf("火山引擎客户端管理器未初始化")
	}
	return s.volcengineMgr.GetClient(ctx, companyID)
}

// CreateSpeechTask 创建声影同传任务
func (s *speechTaskService) CreateSpeechTask(ctx context.Context, companyID uint, req *volcengine.CreateSpeechTaskRequest) (string, error) {
	client, err := s.getClient(ctx, companyID)
	if err != nil {
		return "", err
	}
	taskID, err := client.CreateSpeechTask(ctx, req)
	if err != nil {
		return "", err
	}

	// 落库保存任务记录
	if req.Rule != nil {
		src := ""
		if req.Rule.Source != nil {
			src = req.Rule.Source.Url
		}
		targets := []string{}
		if req.Rule.Output != nil {
			targets = req.Rule.Output.Url
		}
		subtitle := req.Rule.Subtitle != nil
		ruleJSON, _ := json.Marshal(req.Rule)
		rec := &SpeechTaskRecord{
			CompanyID:   companyID,
			RoomID:      strToUint(req.RoomID),
			RoomName:    req.RoomName,
			TaskID:      taskID,
			Name:        req.Name,
			Mode:        req.Rule.Pipe,
			SourceURL:   src,
			TargetURLs:  strings.Join(targets, ","),
			OutputLangs: outputLangsFromRule(req.Rule),
			Subtitle:    subtitle,
			StartTime:   req.Rule.StartTime,
			EndTime:     req.Rule.EndTime,
			Status:      "created",
			RuleJSON:    string(ruleJSON),
			CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		}
		if s.db != nil {
			s.db.WithContext(ctx).Create(rec)
		}
	}
	return taskID, nil
}

// DeleteSpeechTask 删除声影同传任务
func (s *speechTaskService) DeleteSpeechTask(ctx context.Context, companyID uint, req *volcengine.DeleteSpeechTaskRequest) error {
	client, err := s.getClient(ctx, companyID)
	if err != nil {
		return err
	}
	if err := client.DeleteSpeechTask(ctx, req); err != nil {
		return err
	}
	// 删除本地记录
	if s.db != nil && req.TaskID != "" {
		s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND TASK_ID = ?", companyID, req.TaskID).
			Delete(&SpeechTaskRecord{})
	}
	return nil
}

// ListSpeechTasks 查询本地声影同传任务列表
func (s *speechTaskService) ListSpeechTasks(ctx context.Context, companyID, roomID uint) ([]SpeechTaskRecord, error) {
	var items []SpeechTaskRecord
	if s.db == nil {
		return items, nil
	}
	q := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ?", companyID)
	if roomID != 0 {
		q = q.Where("ROOM_ID = ?", roomID)
	}
	err := q.Order("id DESC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	// 填充 CreateTime 字段(DB datetime 列)
	for i := range items {
		items[i].CompanyID = companyID
	}
	return items, nil
}

// strToUint 字符串转uint
func strToUint(s string) uint {
	var v uint
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		v = v*10 + uint(ch-'0')
	}
	return v
}

// outputLangsFromRule 提取字幕目标语言(每个译文子条的 Language)
func outputLangsFromRule(rule *volcengine.SpeechTaskRule) string {
	if rule == nil || rule.Subtitle == nil {
		return ""
	}
	langs := []string{}
	for _, tr := range rule.Subtitle.Translation {
		if tr != nil && tr.Language != "" {
			langs = append(langs, tr.Language)
		}
	}
	return strings.Join(langs, ",")
}
