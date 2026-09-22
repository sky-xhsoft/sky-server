// Package queue 提供基于 Redis Stream 的异步消息队列实现
// 用于解耦耗时任务（邮件发送、文件处理、消息推送等）
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Task 表示一个异步任务
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	Retries   int                    `json:"retries"`
}

// Handler 任务处理器
type Handler func(ctx context.Context, task *Task) error

// Queue Redis Stream 队列接口
type Queue interface {
	// Publish 发布任务到队列
	Publish(ctx context.Context, taskType string, payload map[string]interface{}) error
	// Subscribe 订阅指定类型的任务
	Subscribe(taskType string, handler Handler)
	// Start 启动消费者
	Start(ctx context.Context) error
	// Stop 停止消费者
	Stop()
}

// Config 队列配置
type Config struct {
	StreamKey    string        // Redis Stream 键名
	GroupName    string        // 消费组名称
	ConsumerName string        // 消费者名称
	MaxRetries   int           // 最大重试次数
	BlockTimeout time.Duration // 阻塞超时
	BatchSize    int64         // 批量获取数量
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		StreamKey:    "sky:tasks",
		GroupName:    "sky-workers",
		ConsumerName: fmt.Sprintf("worker-%d", time.Now().UnixNano()),
		MaxRetries:   3,
		BlockTimeout: 5 * time.Second,
		BatchSize:    10,
	}
}

// redisQueue Redis Stream 实现
type redisQueue struct {
	client   *redis.Client
	cfg      *Config
	logger   *zap.Logger
	handlers map[string]Handler
	mu       sync.RWMutex
	stopCh   chan struct{}
	wg       sync.WaitGroup
	running  bool
}

// NewRedisQueue 创建基于 Redis Stream 的队列
func NewRedisQueue(client *redis.Client, cfg *Config, logger *zap.Logger) Queue {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &redisQueue{
		client:   client,
		cfg:      cfg,
		logger:   logger,
		handlers: make(map[string]Handler),
		stopCh:   make(chan struct{}),
	}
}

// Publish 发布任务
func (q *redisQueue) Publish(ctx context.Context, taskType string, payload map[string]interface{}) error {
	task := &Task{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      taskType,
		Payload:   payload,
		CreatedAt: time.Now(),
		Retries:   0,
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	_, err = q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: q.cfg.StreamKey,
		Values: map[string]interface{}{
			"type": taskType,
			"data": string(data),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("publish task: %w", err)
	}

	q.logger.Debug("Task published",
		zap.String("id", task.ID),
		zap.String("type", taskType))
	return nil
}

// Subscribe 注册任务处理器
func (q *redisQueue) Subscribe(taskType string, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[taskType] = handler
}

// Start 启动消费者
func (q *redisQueue) Start(ctx context.Context) error {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return fmt.Errorf("queue already running")
	}
	q.running = true
	q.mu.Unlock()

	// 创建消费组（已存在则忽略）
	err := q.client.XGroupCreateMkStream(ctx, q.cfg.StreamKey, q.cfg.GroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}

	q.wg.Add(1)
	go q.consume(ctx)

	q.logger.Info("Queue started",
		zap.String("stream", q.cfg.StreamKey),
		zap.String("group", q.cfg.GroupName),
		zap.String("consumer", q.cfg.ConsumerName))

	return nil
}

// Stop 停止消费者
func (q *redisQueue) Stop() {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return
	}
	q.running = false
	close(q.stopCh)
	q.mu.Unlock()

	q.wg.Wait()
	q.logger.Info("Queue stopped")
}

// consume 消费任务循环
func (q *redisQueue) consume(ctx context.Context) {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			q.consumeOnce(ctx)
		}
	}
}

// consumeOnce 消费一批任务
func (q *redisQueue) consumeOnce(ctx context.Context) {
	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.cfg.GroupName,
		Consumer: q.cfg.ConsumerName,
		Streams:  []string{q.cfg.StreamKey, ">"},
		Count:    q.cfg.BatchSize,
		Block:    q.cfg.BlockTimeout,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return
		}
		q.logger.Error("Read group failed", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			q.processMessage(ctx, msg)
		}
	}
}

// processMessage 处理单条消息
func (q *redisQueue) processMessage(ctx context.Context, msg redis.XMessage) {
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		q.logger.Warn("Invalid message data", zap.String("id", msg.ID))
		q.ack(ctx, msg.ID)
		return
	}

	var task Task
	if err := json.Unmarshal([]byte(dataStr), &task); err != nil {
		q.logger.Error("Unmarshal task failed", zap.Error(err), zap.String("id", msg.ID))
		q.ack(ctx, msg.ID)
		return
	}

	q.mu.RLock()
	handler, exists := q.handlers[task.Type]
	q.mu.RUnlock()

	if !exists {
		q.logger.Warn("No handler for task type", zap.String("type", task.Type))
		q.ack(ctx, msg.ID)
		return
	}

	if err := handler(ctx, &task); err != nil {
		q.logger.Error("Task handler failed",
			zap.String("id", task.ID),
			zap.String("type", task.Type),
			zap.Int("retries", task.Retries),
			zap.Error(err))

		// 重试机制
		if task.Retries < q.cfg.MaxRetries {
			task.Retries++
			if data, mErr := json.Marshal(&task); mErr == nil {
				q.client.XAdd(ctx, &redis.XAddArgs{
					Stream: q.cfg.StreamKey,
					Values: map[string]interface{}{
						"type": task.Type,
						"data": string(data),
					},
				})
			}
		}
	}

	q.ack(ctx, msg.ID)
}

// ack 确认消息已处理
func (q *redisQueue) ack(ctx context.Context, msgID string) {
	if err := q.client.XAck(ctx, q.cfg.StreamKey, q.cfg.GroupName, msgID).Err(); err != nil {
		q.logger.Warn("Ack failed", zap.Error(err), zap.String("id", msgID))
	}
}
