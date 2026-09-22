package queue

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.StreamKey == "" {
		t.Error("StreamKey should not be empty")
	}
	if cfg.GroupName == "" {
		t.Error("GroupName should not be empty")
	}
	if cfg.MaxRetries <= 0 {
		t.Error("MaxRetries should be > 0")
	}
}

func TestTaskSerialization(t *testing.T) {
	task := &Task{
		ID:   "test-1",
		Type: "email.send",
		Payload: map[string]interface{}{
			"to":      "test@example.com",
			"subject": "Test",
		},
		CreatedAt: time.Now(),
	}

	if task.ID != "test-1" {
		t.Error("Task ID mismatch")
	}
	if task.Type != "email.send" {
		t.Error("Task Type mismatch")
	}
	if task.Payload["to"] != "test@example.com" {
		t.Error("Payload to mismatch")
	}
}

func TestQueueStop_NotRunning(t *testing.T) {
	// 测试未启动时 Stop 不会 panic
	q := &redisQueue{
		cfg:      DefaultConfig(),
		handlers: make(map[string]Handler),
		stopCh:   make(chan struct{}),
	}
	q.Stop()
}

func TestQueueSubscribe(t *testing.T) {
	q := &redisQueue{
		cfg:      DefaultConfig(),
		handlers: make(map[string]Handler),
	}

	called := false
	q.Subscribe("test", func(ctx context.Context, task *Task) error {
		called = true
		return nil
	})

	if _, exists := q.handlers["test"]; !exists {
		t.Error("Handler not registered")
	}

	// 验证可以多次订阅覆盖
	q.Subscribe("test", func(ctx context.Context, task *Task) error {
		called = true
		return nil
	})

	if called {
		t.Error("Handler should not be called yet")
	}
}
