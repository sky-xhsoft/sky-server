package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"gorm.io/gorm"
)

// Service 审计日志服务接口
type Service interface {
	// 记录审计日志
	Log(ctx context.Context, log *entity.AuditLog) error

	// 异步记录审计日志(不阻塞主流程)
	LogAsync(log *entity.AuditLog)

	// 查询审计日志
	QueryLogs(ctx context.Context, req *QueryRequest) ([]*entity.AuditLog, int64, error)

	// 获取单条日志
	GetLog(ctx context.Context, id uint) (*entity.AuditLog, error)

	// 按用户查询日志
	GetUserLogs(ctx context.Context, userID uint, page, pageSize int) ([]*entity.AuditLog, int64, error)

	// 按资源查询日志
	GetResourceLogs(ctx context.Context, resource, resourceID string, page, pageSize int) ([]*entity.AuditLog, int64, error)

	// 统计接口
	GetStatistics(ctx context.Context, req *StatisticsRequest) (*Statistics, error)

	// 清理过期日志
	CleanExpiredLogs(ctx context.Context, beforeDate time.Time) (int64, error)
}

// QueryRequest 查询请求
type QueryRequest struct {
	UserID       uint      `json:"userId"`       // 用户ID
	Username     string    `json:"username"`     // 用户名
	Action       string    `json:"action"`       // 操作类型
	Resource     string    `json:"resource"`     // 资源类型
	ResourceID   string    `json:"resourceId"`   // 资源ID
	Status       string    `json:"status"`       // 状态
	IP           string    `json:"ip"`           // IP地址
	StartTime    time.Time `json:"startTime"`    // 开始时间
	EndTime      time.Time `json:"endTime"`      // 结束时间
	Page         int       `json:"page"`         // 页码
	PageSize     int       `json:"pageSize"`     // 每页大小
	SortBy       string    `json:"sortBy"`       // 排序字段
	SortOrder    string    `json:"sortOrder"`    // 排序方向
}

// StatisticsRequest 统计请求
type StatisticsRequest struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	GroupBy   string    `json:"groupBy"` // action, resource, user, date
}

// Statistics 统计结果
type Statistics struct {
	TotalCount    int64                    `json:"totalCount"`    // 总数
	SuccessCount  int64                    `json:"successCount"`  // 成功数
	FailureCount  int64                    `json:"failureCount"`  // 失败数
	ByAction      map[string]int64         `json:"byAction"`      // 按操作类型统计
	ByResource    map[string]int64         `json:"byResource"`    // 按资源类型统计
	ByUser        map[string]int64         `json:"byUser"`        // 按用户统计
	ByDate        map[string]int64         `json:"byDate"`        // 按日期统计
	TopUsers      []UserStat               `json:"topUsers"`      // 活跃用户TOP10
	TopActions    []ActionStat             `json:"topActions"`    // 热门操作TOP10
}

// UserStat 用户统计
type UserStat struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

// ActionStat 操作统计
type ActionStat struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// service 审计日志服务实现
type service struct {
	db      *gorm.DB
	logChan chan *entity.AuditLog // 异步日志通道
}

// NewService 创建审计日志服务
func NewService(db *gorm.DB) Service {
	s := &service{
		db:      db,
		logChan: make(chan *entity.AuditLog, 1000), // 缓冲1000条日志
	}

	// 启动异步日志处理goroutine
	go s.processAsyncLogs()

	return s
}

// Log 记录审计日志(同步)
func (s *service) Log(ctx context.Context, log *entity.AuditLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	if err := s.db.WithContext(ctx).Create(log).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "记录审计日志失败", err)
	}

	return nil
}

// LogAsync 异步记录审计日志(不阻塞)
func (s *service) LogAsync(log *entity.AuditLog) {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// 非阻塞发送到通道
	select {
	case s.logChan <- log:
		// 成功发送
	default:
		// 通道已满,丢弃日志(或者可以选择同步写入)
		// 这里选择丢弃以避免阻塞主流程
	}
}

// processAsyncLogs 处理异步日志
func (s *service) processAsyncLogs() {
	// 批量插入配置
	batchSize := 100
	batchTimeout := 5 * time.Second

	var batch []*entity.AuditLog
	timer := time.NewTimer(batchTimeout)
	defer timer.Stop()

	for {
		select {
		case log := <-s.logChan:
			batch = append(batch, log)

			// 达到批量大小，立即写入
			if len(batch) >= batchSize {
				s.writeBatch(batch)
				batch = nil
				timer.Reset(batchTimeout)
			}

		case <-timer.C:
			// 超时，写入当前批次
			if len(batch) > 0 {
				s.writeBatch(batch)
				batch = nil
			}
			timer.Reset(batchTimeout)
		}
	}
}

// writeBatch 批量写入日志
func (s *service) writeBatch(logs []*entity.AuditLog) {
	if len(logs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.db.WithContext(ctx).CreateInBatches(logs, len(logs)).Error; err != nil {
		// 记录错误但不影响主流程
		// TODO: 可以考虑将失败的日志写入文件
	}
}

// QueryLogs 查询审计日志
func (s *service) QueryLogs(ctx context.Context, req *QueryRequest) ([]*entity.AuditLog, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	query := s.db.WithContext(ctx).Model(&entity.AuditLog{})

	// 应用过滤条件
	if req.UserID > 0 {
		query = query.Where("USER_ID = ?", req.UserID)
	}
	if req.Username != "" {
		query = query.Where("USERNAME LIKE ?", "%"+req.Username+"%")
	}
	if req.Action != "" {
		query = query.Where("ACTION = ?", req.Action)
	}
	if req.Resource != "" {
		query = query.Where("RESOURCE = ?", req.Resource)
	}
	if req.ResourceID != "" {
		query = query.Where("RESOURCE_ID = ?", req.ResourceID)
	}
	if req.Status != "" {
		query = query.Where("STATUS = ?", req.Status)
	}
	if req.IP != "" {
		query = query.Where("IP = ?", req.IP)
	}
	if !req.StartTime.IsZero() {
		query = query.Where("CREATED_AT >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("CREATED_AT <= ?", req.EndTime)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询审计日志总数失败", err)
	}

	// 排序
	sortBy := "CREATED_AT"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder == "ASC" {
		sortOrder = "ASC"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var logs []*entity.AuditLog
	if err := query.Limit(req.PageSize).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询审计日志列表失败", err)
	}

	return logs, total, nil
}

// GetLog 获取单条日志
func (s *service) GetLog(ctx context.Context, id uint) (*entity.AuditLog, error) {
	var log entity.AuditLog
	if err := s.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "审计日志不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询审计日志失败", err)
	}

	return &log, nil
}

// GetUserLogs 按用户查询日志
func (s *service) GetUserLogs(ctx context.Context, userID uint, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	return s.QueryLogs(ctx, &QueryRequest{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetResourceLogs 按资源查询日志
func (s *service) GetResourceLogs(ctx context.Context, resource, resourceID string, page, pageSize int) ([]*entity.AuditLog, int64, error) {
	return s.QueryLogs(ctx, &QueryRequest{
		Resource:   resource,
		ResourceID: resourceID,
		Page:       page,
		PageSize:   pageSize,
	})
}

// GetStatistics 获取统计信息
func (s *service) GetStatistics(ctx context.Context, req *StatisticsRequest) (*Statistics, error) {
	query := s.db.WithContext(ctx).Model(&entity.AuditLog{})

	if !req.StartTime.IsZero() {
		query = query.Where("CREATED_AT >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("CREATED_AT <= ?", req.EndTime)
	}

	stats := &Statistics{
		ByAction:   make(map[string]int64),
		ByResource: make(map[string]int64),
		ByUser:     make(map[string]int64),
		ByDate:     make(map[string]int64),
	}

	// 总数统计
	query.Count(&stats.TotalCount)

	// 成功/失败统计
	s.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Where("STATUS = ?", entity.StatusSuccess).
		Count(&stats.SuccessCount)
	s.db.WithContext(ctx).Model(&entity.AuditLog{}).
		Where("STATUS = ?", entity.StatusFailure).
		Count(&stats.FailureCount)

	// 按操作类型统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	query.Select("ACTION as action, COUNT(*) as count").
		Group("ACTION").
		Scan(&actionStats)
	for _, stat := range actionStats {
		stats.ByAction[stat.Action] = stat.Count
	}

	// 按资源类型统计
	var resourceStats []struct {
		Resource string
		Count    int64
	}
	query.Select("RESOURCE as resource, COUNT(*) as count").
		Group("RESOURCE").
		Scan(&resourceStats)
	for _, stat := range resourceStats {
		stats.ByResource[stat.Resource] = stat.Count
	}

	// TOP活跃用户
	var topUsers []struct {
		UserID   uint
		Username string
		Count    int64
	}
	query.Select("USER_ID as user_id, USERNAME as username, COUNT(*) as count").
		Group("USER_ID, USERNAME").
		Order("count DESC").
		Limit(10).
		Scan(&topUsers)
	for _, user := range topUsers {
		stats.TopUsers = append(stats.TopUsers, UserStat{
			UserID:   user.UserID,
			Username: user.Username,
			Count:    user.Count,
		})
	}

	// TOP操作
	var topActions []struct {
		Action string
		Count  int64
	}
	query.Select("ACTION as action, COUNT(*) as count").
		Group("ACTION").
		Order("count DESC").
		Limit(10).
		Scan(&topActions)
	for _, action := range topActions {
		stats.TopActions = append(stats.TopActions, ActionStat{
			Action: action.Action,
			Count:  action.Count,
		})
	}

	return stats, nil
}

// CleanExpiredLogs 清理过期日志
func (s *service) CleanExpiredLogs(ctx context.Context, beforeDate time.Time) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("CREATED_AT < ?", beforeDate).
		Delete(&entity.AuditLog{})

	if result.Error != nil {
		return 0, errors.Wrap(errors.ErrDatabase, "清理过期日志失败", result.Error)
	}

	return result.RowsAffected, nil
}

// LogBuilder 审计日志构建器(辅助类)
type LogBuilder struct {
	log *entity.AuditLog
}

// NewLogBuilder 创建日志构建器
func NewLogBuilder() *LogBuilder {
	return &LogBuilder{
		log: &entity.AuditLog{
			Status:    entity.StatusSuccess,
			CreatedAt: time.Now(),
		},
	}
}

// WithUser 设置用户信息
func (b *LogBuilder) WithUser(userID uint, username string) *LogBuilder {
	b.log.UserID = userID
	b.log.Username = username
	return b
}

// WithAction 设置操作类型
func (b *LogBuilder) WithAction(action string) *LogBuilder {
	b.log.Action = action
	return b
}

// WithResource 设置资源信息
func (b *LogBuilder) WithResource(resource, resourceID, resourceName string) *LogBuilder {
	b.log.Resource = resource
	b.log.ResourceID = resourceID
	b.log.ResourceName = resourceName
	return b
}

// WithRequest 设置请求信息
func (b *LogBuilder) WithRequest(method, path, ip, userAgent string) *LogBuilder {
	b.log.Method = method
	b.log.Path = path
	b.log.IP = ip
	b.log.UserAgent = userAgent
	return b
}

// WithRequestBody 设置请求体
func (b *LogBuilder) WithRequestBody(body interface{}) *LogBuilder {
	if body != nil {
		data, _ := json.Marshal(body)
		b.log.RequestBody = string(data)
	}
	return b
}

// WithResponseBody 设置响应体
func (b *LogBuilder) WithResponseBody(body interface{}) *LogBuilder {
	if body != nil {
		data, _ := json.Marshal(body)
		b.log.ResponseBody = string(data)
	}
	return b
}

// WithOldValue 设置修改前的值
func (b *LogBuilder) WithOldValue(value interface{}) *LogBuilder {
	if value != nil {
		data, _ := json.Marshal(value)
		b.log.OldValue = string(data)
	}
	return b
}

// WithNewValue 设置修改后的值
func (b *LogBuilder) WithNewValue(value interface{}) *LogBuilder {
	if value != nil {
		data, _ := json.Marshal(value)
		b.log.NewValue = string(data)
	}
	return b
}

// WithStatus 设置状态
func (b *LogBuilder) WithStatus(status string) *LogBuilder {
	b.log.Status = status
	return b
}

// WithError 设置错误信息
func (b *LogBuilder) WithError(err error) *LogBuilder {
	if err != nil {
		b.log.Status = entity.StatusFailure
		b.log.ErrorMessage = err.Error()
	}
	return b
}

// WithDuration 设置执行时长(毫秒)
func (b *LogBuilder) WithDuration(duration int64) *LogBuilder {
	b.log.Duration = duration
	return b
}

// WithTags 设置标签
func (b *LogBuilder) WithTags(tags string) *LogBuilder {
	b.log.Tags = tags
	return b
}

// WithCompanyID 设置公司ID
func (b *LogBuilder) WithCompanyID(companyID uint) *LogBuilder {
	b.log.SysCompanyID = companyID
	return b
}

// Build 构建日志对象
func (b *LogBuilder) Build() *entity.AuditLog {
	return b.log
}
