package domain

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventBus 事件总线
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

var (
	eventBusInstance *EventBus
	eventBusOnce     sync.Once
)

// GetEventBus 获取事件总线单例
func GetEventBus() *EventBus {
	eventBusOnce.Do(func() {
		eventBusInstance = &EventBus{
			handlers: make(map[string][]EventHandler),
		}
	})
	return eventBusInstance
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handlers, ok := eb.handlers[eventType]; ok {
		newHandlers := make([]EventHandler, 0, len(handlers))
		for _, h := range handlers {
			if h != handler {
				newHandlers = append(newHandlers, h)
			}
		}
		eb.handlers[eventType] = newHandlers
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(event *DomainEvent) error {
	eb.mu.RLock()
	handlers := make([]EventHandler, 0)
	if hs, ok := eb.handlers[event.EventType]; ok {
		handlers = append(handlers, hs...)
	}
	if hs, ok := eb.handlers["*"]; ok {
		handlers = append(handlers, hs...)
	}
	eb.mu.RUnlock()

	var lastErr error
	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// PublishAsync 异步发布事件
func (eb *EventBus) PublishAsync(event *DomainEvent) {
	go func() {
		_ = eb.Publish(event)
	}()
}

// NewDomainEvent 创建新的领域事件
func NewDomainEvent(eventType, sourceID, sourceType string, data interface{}) *DomainEvent {
	return &DomainEvent{
		EventID:    uuid.New().String(),
		EventType:  eventType,
		SourceID:   sourceID,
		SourceType: sourceType,
		Data:       data,
		Timestamp:  time.Now(),
		Version:    1,
	}
}

// EventStore 事件存储接口
type EventStore interface {
	Save(event *DomainEvent) error
	GetBySourceID(sourceID string) ([]*DomainEvent, error)
	GetByEventType(eventType string) ([]*DomainEvent, error)
	GetByTimeRange(start, end time.Time) ([]*DomainEvent, error)
}

// MemoryEventStore 内存事件存储实现
type MemoryEventStore struct {
	events []*DomainEvent
	mu     sync.RWMutex
}

// NewMemoryEventStore 创建内存事件存储
func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events: make([]*DomainEvent, 0),
	}
}

// Save 保存事件
func (mes *MemoryEventStore) Save(event *DomainEvent) error {
	mes.mu.Lock()
	defer mes.mu.Unlock()
	mes.events = append(mes.events, event)
	return nil
}

// GetBySourceID 根据源ID获取事件
func (mes *MemoryEventStore) GetBySourceID(sourceID string) ([]*DomainEvent, error) {
	mes.mu.RLock()
	defer mes.mu.RUnlock()

	result := make([]*DomainEvent, 0)
	for _, e := range mes.events {
		if e.SourceID == sourceID {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetByEventType 根据事件类型获取事件
func (mes *MemoryEventStore) GetByEventType(eventType string) ([]*DomainEvent, error) {
	mes.mu.RLock()
	defer mes.mu.RUnlock()

	result := make([]*DomainEvent, 0)
	for _, e := range mes.events {
		if e.EventType == eventType {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetByTimeRange 根据时间范围获取事件
func (mes *MemoryEventStore) GetByTimeRange(start, end time.Time) ([]*DomainEvent, error) {
	mes.mu.RLock()
	defer mes.mu.RUnlock()

	result := make([]*DomainEvent, 0)
	for _, e := range mes.events {
		if !e.Timestamp.Before(start) && e.Timestamp.Before(end) {
			result = append(result, e)
		}
	}
	return result, nil
}