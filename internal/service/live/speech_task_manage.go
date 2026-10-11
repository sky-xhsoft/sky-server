package live

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/volcengine"
)

// 同传任务状态（与火山一致）；本地新建后尚未同步时为 created
const (
	SpeechStatusCreated = "created"
	SpeechStatusInit    = "init"    // 准备中
	SpeechStatusPrepare = "prepare" // 启动中
	SpeechStatusRunning = "running" // 运行中
	SpeechStatusPrestop = "prestop" // 关闭中
	SpeechStatusDone    = "done"    // 已结束
)

// ErrSpeechTaskNotFound 任务不存在或不属于本公司
var ErrSpeechTaskNotFound = errors.New("任务不存在或不属于本公司")

// SpeechTaskQuery 同传任务列表查询
type SpeechTaskQuery struct {
	RoomID   uint
	Keyword  string // 按任务名称、直播间名称模糊匹配
	Status   string // 空为全部；active 为未结束（准备中 / 启动中 / 运行中 / 关闭中）
	Page     int
	PageSize int
}

// SpeechTaskPage 同传任务分页结果
type SpeechTaskPage struct {
	Items     []SpeechTaskRecord `json:"items"`
	Total     int                `json:"total"`
	SyncError string             `json:"syncError,omitempty"` // 未能从火山同步状态的原因
}

func isActiveSpeechStatus(s string) bool {
	return s != SpeechStatusDone
}

// parseLocalTime 本地记录的时间（数据库 datetime 读出为 RFC3339 或 "2006-01-02 15:04:05"）
func parseLocalTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// ListSpeechTasksPaged 本公司的同传任务（可按直播间、名称、状态筛选），并从火山同步未结束任务的实时状态
func (s *speechTaskService) ListSpeechTasksPaged(ctx context.Context, companyID uint, q SpeechTaskQuery) (*SpeechTaskPage, error) {
	page := &SpeechTaskPage{Items: []SpeechTaskRecord{}}
	if s.db == nil {
		return page, nil
	}
	db := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ?", companyID)
	if q.RoomID != 0 {
		db = db.Where("ROOM_ID = ?", q.RoomID)
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		db = db.Where("(NAME LIKE ? OR ROOM_NAME LIKE ?)", "%"+kw+"%", "%"+kw+"%")
	}
	var all []SpeechTaskRecord
	if err := db.Order("ID DESC").Find(&all).Error; err != nil {
		return nil, err
	}
	if err := s.syncStatuses(ctx, all); err != nil {
		page.SyncError = err.Error()
	}

	filtered := all[:0]
	for _, r := range all {
		switch {
		case q.Status == "":
		case q.Status == "active" && !isActiveSpeechStatus(r.Status):
			continue
		case q.Status != "active" && r.Status != q.Status && !(q.Status == SpeechStatusInit && r.Status == SpeechStatusCreated):
			continue
		}
		filtered = append(filtered, r)
	}
	page.Total = len(filtered)
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	start := (q.Page - 1) * q.PageSize
	if start < len(filtered) {
		end := start + q.PageSize
		if end > len(filtered) {
			end = len(filtered)
		}
		page.Items = append(page.Items, filtered[start:end]...)
	}
	for i := range page.Items {
		page.Items[i].CompanyID = companyID
	}
	return page, nil
}

// syncStatuses 未结束的任务按创建时间区间搜索一次火山（共用账号，按任务 ID 对应），状态有变化时写回本地
func (s *speechTaskService) syncStatuses(ctx context.Context, records []SpeechTaskRecord) error {
	var earliest time.Time
	active := map[string]*SpeechTaskRecord{}
	for i := range records {
		r := &records[i]
		if !isActiveSpeechStatus(r.Status) || r.TaskID == "" {
			continue
		}
		active[r.TaskID] = r
		if t, ok := parseLocalTime(r.CreateTime); ok && (earliest.IsZero() || t.Before(earliest)) {
			earliest = t
		}
	}
	if len(active) == 0 {
		return nil
	}
	client, err := s.getClient(ctx, 0)
	if err != nil {
		return err
	}
	req := &volcengine.SearchSpeechTaskRequest{PageSize: 100}
	if !earliest.IsZero() {
		req.CreateTimeSt = earliest.Add(-10 * time.Minute).Format("2006-01-02 15:04:05")
	}
	found := map[string]string{}
	for p := 1; p <= 10; p++ {
		req.Page = p
		total, tasks, err := client.SearchSpeechTask(ctx, req)
		if err != nil {
			return err
		}
		for _, t := range tasks {
			found[t.TaskID] = t.Status
		}
		if p*req.PageSize >= total || len(tasks) == 0 {
			break
		}
	}
	for id, r := range active {
		st, ok := found[id]
		if !ok {
			// 火山上已搜索不到（已被彻底删除），视为已结束
			st = SpeechStatusDone
		}
		if st != "" && st != r.Status {
			r.Status = st
			s.db.WithContext(ctx).Model(&SpeechTaskRecord{}).Where("ID = ?", r.ID).Update("STATUS", st)
		}
	}
	return nil
}

// ownTask 本公司的任务记录
func (s *speechTaskService) ownTask(ctx context.Context, companyID uint, taskID string) (*SpeechTaskRecord, error) {
	if s.db == nil || strings.TrimSpace(taskID) == "" {
		return nil, ErrSpeechTaskNotFound
	}
	var r SpeechTaskRecord
	if err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND TASK_ID = ?", companyID, taskID).First(&r).Error; err != nil {
		return nil, ErrSpeechTaskNotFound
	}
	return &r, nil
}

// StopSpeechTask 停止任务：火山删除但保留历史记录（WithRecord=false），本地记录保留，可复制新建
func (s *speechTaskService) StopSpeechTask(ctx context.Context, companyID uint, taskID string) error {
	r, err := s.ownTask(ctx, companyID, taskID)
	if err != nil {
		return err
	}
	if r.Status == SpeechStatusDone {
		return errors.New("任务已结束")
	}
	client, err := s.getClient(ctx, companyID)
	if err != nil {
		return err
	}
	if err := client.DeleteSpeechTask(ctx, &volcengine.DeleteSpeechTaskRequest{TaskID: taskID}); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&SpeechTaskRecord{}).Where("ID = ?", r.ID).Update("STATUS", SpeechStatusPrestop).Error
}

// RemoveSpeechTask 彻底删除：火山连同历史一起删除（运行中的任务火山会先停止），并删除本地记录
func (s *speechTaskService) RemoveSpeechTask(ctx context.Context, companyID uint, taskID string) error {
	r, err := s.ownTask(ctx, companyID, taskID)
	if err != nil {
		return err
	}
	client, err := s.getClient(ctx, companyID)
	if err != nil {
		return err
	}
	if err := client.DeleteSpeechTask(ctx, &volcengine.DeleteSpeechTaskRequest{TaskID: taskID, WithRecord: true}); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(&SpeechTaskRecord{}, r.ID).Error
}

// RestartSpeechTask 重启运行中或启动中的任务；reloadSubtitle 为 true 时不中断直播，仅重启字幕算法模块
func (s *speechTaskService) RestartSpeechTask(ctx context.Context, companyID uint, taskID string, reloadSubtitle bool) error {
	r, err := s.ownTask(ctx, companyID, taskID)
	if err != nil {
		return err
	}
	client, err := s.getClient(ctx, companyID)
	if err != nil {
		return err
	}
	info, err := client.GetSpeechTask(ctx, taskID)
	if err != nil {
		return err
	}
	if info.Status != r.Status {
		s.db.WithContext(ctx).Model(&SpeechTaskRecord{}).Where("ID = ?", r.ID).Update("STATUS", info.Status)
	}
	if info.Status != SpeechStatusRunning && info.Status != SpeechStatusPrepare {
		return errors.New("只有运行中或启动中的任务可以重启")
	}
	return client.RestartSpeechTask(ctx, taskID, reloadSubtitle)
}

// GetSpeechTaskRule 任务创建时提交的参数（用于复制新建）
func (s *speechTaskService) GetSpeechTaskRule(ctx context.Context, companyID uint, taskID string) (json.RawMessage, error) {
	r, err := s.ownTask(ctx, companyID, taskID)
	if err != nil {
		return nil, err
	}
	if r.RuleJSON == "" || !json.Valid([]byte(r.RuleJSON)) {
		return nil, errors.New("该任务没有保存创建参数，无法复制")
	}
	return json.RawMessage(r.RuleJSON), nil
}
