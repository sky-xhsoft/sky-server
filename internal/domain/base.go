package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Entity 领域实体基类
type Entity struct {
	ID         uint           `gorm:"primaryKey"`
	SysCompanyID uint        `gorm:"column:sys_company_id;comment:公司ID"`
	CreateBy   string        `gorm:"column:create_by;comment:创建人"`
	CreateTime time.Time     `gorm:"column:create_time;comment:创建时间;autoCreateTime"`
	UpdateBy   string        `gorm:"column:update_by;comment:更新人"`
	UpdateTime time.Time     `gorm:"column:update_time;comment:更新时间;autoUpdateTime"`
	IsActive   string        `gorm:"column:is_active;comment:是否有效;default:Y"`
}

// BaseDomainService 领域服务基类
type BaseDomainService interface {
	WithContext(ctx context.Context)
	GetDB() *gorm.DB
}

// DomainRepository 领域仓库接口
type DomainRepository interface {
	GetByID(ctx context.Context, id uint) (Entity, error)
	GetList(ctx context.Context, filters map[string]interface{}) ([]Entity, error)
	GetPage(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]Entity, int64, error)
	Create(ctx context.Context, entity Entity) error
	Update(ctx context.Context, entity Entity) error
	Delete(ctx context.Context, id uint) error
}

// DomainEvent 领域事件基类
type DomainEvent struct {
	EventID     string        `json:"event_id"`
	EventType   string        `json:"event_type"`
	SourceID    string        `json:"source_id"`
	SourceType  string        `json:"source_type"`
	Data        interface{}   `json:"data"`
	Timestamp   time.Time     `json:"timestamp"`
	Version     int           `json:"version"`
}

// EventHandler 事件处理接口
type EventHandler interface {
	Handle(event *DomainEvent) error
	GetSupportedEvents() []string
}

// DomainError 领域错误
type DomainError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// NewDomainError 创建领域错误
func NewDomainError(code int, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error 实现error接口
func (e *DomainError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}