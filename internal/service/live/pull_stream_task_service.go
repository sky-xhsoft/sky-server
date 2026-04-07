package live

import (
	"context"
	"errors"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// PullStreamTaskService 拉流任务服务接口
type PullStreamTaskService interface {
	// CreatePullStreamTask 创建拉流任务
	CreatePullStreamTask(ctx context.Context, task *entity.PullStreamTask) error
	// UpdatePullStreamTask 更新拉流任务
	UpdatePullStreamTask(ctx context.Context, task *entity.PullStreamTask) error
	// DeletePullStreamTask 删除拉流任务
	DeletePullStreamTask(ctx context.Context, taskID string) error
	// GetPullStreamTask 获取拉流任务详情
	GetPullStreamTask(ctx context.Context, taskID string) (*entity.PullStreamTask, error)
	// ListPullStreamTasks 查询拉流任务列表
	ListPullStreamTasks(ctx context.Context, filter *PullStreamTaskFilter) ([]*entity.PullStreamTask, int64, error)
	// DisablePullStreamTask 禁用拉流任务
	DisablePullStreamTask(ctx context.Context, taskID string) error
	// EnablePullStreamTask 启用拉流任务
	EnablePullStreamTask(ctx context.Context, taskID string) error
}

// PullStreamTaskFilter 拉流任务查询过滤器
type PullStreamTaskFilter struct {
	Region        string
	SourceType    string
	Status        string
	Operator      string
	Keyword       string // 搜索关键词（任务备注）
	StartTimeFrom *time.Time
	StartTimeTo   *time.Time
	EndTimeFrom   *time.Time
	EndTimeTo     *time.Time
	Page          int
	PageSize      int
}

type pullStreamTaskService struct {
	db *gorm.DB
}

// NewPullStreamTaskService 创建拉流任务服务实例
func NewPullStreamTaskService(db *gorm.DB) PullStreamTaskService {
	return &pullStreamTaskService{
		db: db,
	}
}

// CreatePullStreamTask 创建拉流任务
func (s *pullStreamTaskService) CreatePullStreamTask(ctx context.Context, task *entity.PullStreamTask) error {
	if task.TaskID == "" {
		return errors.New("任务ID不能为空")
	}
	if task.Region == "" {
		return errors.New("地域不能为空")
	}
	if task.SourceType == "" {
		return errors.New("内容类型不能为空")
	}
	if task.SourceURL == "" {
		return errors.New("直播源地址不能为空")
	}
	if task.TargetURL == "" {
		return errors.New("目标地址不能为空")
	}
	if task.StartTime.IsZero() {
		return errors.New("开始时间不能为空")
	}
	if task.EndTime.IsZero() {
		return errors.New("结束时间不能为空")
	}
	if task.Operator == "" {
		return errors.New("操作者不能为空")
	}

	// 设置默认值
	if task.Status == "" {
		task.Status = "enable"
	}
	if task.IsActive == "" {
		task.IsActive = "Y"
	}

	return s.db.WithContext(ctx).Create(task).Error
}

// UpdatePullStreamTask 更新拉流任务
func (s *pullStreamTaskService) UpdatePullStreamTask(ctx context.Context, task *entity.PullStreamTask) error {
	if task.TaskID == "" {
		return errors.New("任务ID不能为空")
	}

	// 检查任务是否存在
	var existingTask entity.PullStreamTask
	if err := s.db.WithContext(ctx).Where("TASK_ID = ? AND IS_ACTIVE = ?", task.TaskID, "Y").First(&existingTask).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("拉流任务不存在")
		}
		return err
	}

	return s.db.WithContext(ctx).Model(&entity.PullStreamTask{}).Where("TASK_ID = ?", task.TaskID).Updates(task).Error
}

// DeletePullStreamTask 删除拉流任务
func (s *pullStreamTaskService) DeletePullStreamTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return errors.New("任务ID不能为空")
	}

	// 直接删除
	return s.db.WithContext(ctx).Where("TASK_ID = ?", taskID).Delete(&entity.PullStreamTask{}).Error
}

// GetPullStreamTask 获取拉流任务详情
func (s *pullStreamTaskService) GetPullStreamTask(ctx context.Context, taskID string) (*entity.PullStreamTask, error) {
	if taskID == "" {
		return nil, errors.New("任务ID不能为空")
	}

	var task entity.PullStreamTask
	if err := s.db.WithContext(ctx).Where("TASK_ID = ? AND IS_ACTIVE = ?", taskID, "Y").First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("拉流任务不存在")
		}
		return nil, err
	}

	return &task, nil
}

// ListPullStreamTasks 查询拉流任务列表
func (s *pullStreamTaskService) ListPullStreamTasks(ctx context.Context, filter *PullStreamTaskFilter) ([]*entity.PullStreamTask, int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.PullStreamTask{}).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	if filter.Region != "" {
		query = query.Where("REGION = ?", filter.Region)
	}
	if filter.SourceType != "" {
		query = query.Where("SOURCE_TYPE = ?", filter.SourceType)
	}
	if filter.Status != "" {
		query = query.Where("STATUS = ?", filter.Status)
	}
	if filter.Operator != "" {
		query = query.Where("OPERATOR = ?", filter.Operator)
	}
	if filter.Keyword != "" {
		query = query.Where("COMMENT LIKE ?", "%"+filter.Keyword+"%")
	}
	if filter.StartTimeFrom != nil {
		query = query.Where("START_TIME >= ?", filter.StartTimeFrom)
	}
	if filter.StartTimeTo != nil {
		query = query.Where("START_TIME <= ?", filter.StartTimeTo)
	}
	if filter.EndTimeFrom != nil {
		query = query.Where("END_TIME >= ?", filter.EndTimeFrom)
	}
	if filter.EndTimeTo != nil {
		query = query.Where("END_TIME <= ?", filter.EndTimeTo)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var tasks []*entity.PullStreamTask
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(filter.PageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// DisablePullStreamTask 禁用拉流任务
func (s *pullStreamTaskService) DisablePullStreamTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return errors.New("任务ID不能为空")
	}

	// 检查任务是否存在
	var task entity.PullStreamTask
	if err := s.db.WithContext(ctx).Where("TASK_ID = ? AND IS_ACTIVE = ?", taskID, "Y").First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("拉流任务不存在")
		}
		return err
	}

	// 更新状态为禁用
	return s.db.WithContext(ctx).Model(&entity.PullStreamTask{}).Where("TASK_ID = ?", taskID).Updates(map[string]interface{}{
		"STATUS":    "pause",
		"IS_ACTIVE": "N",
	}).Error
}

// EnablePullStreamTask 启用拉流任务
func (s *pullStreamTaskService) EnablePullStreamTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return errors.New("任务ID不能为空")
	}

	// 检查任务是否存在
	var task entity.PullStreamTask
	if err := s.db.WithContext(ctx).Where("TASK_ID = ?", taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("拉流任务不存在")
		}
		return err
	}

	// 更新状态为启用
	return s.db.WithContext(ctx).Model(&entity.PullStreamTask{}).Where("TASK_ID = ?", taskID).Updates(map[string]interface{}{
		"STATUS":    "enable",
		"IS_ACTIVE": "Y",
	}).Error
}
