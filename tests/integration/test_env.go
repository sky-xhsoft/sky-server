// Package integration 提供集成测试基础设施
// 测试整个 API 流程，包括路由、中间件、服务等
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// TestEnvironment 集成测试环境
type TestEnvironment struct {
	T      *testing.T
	DB     *gorm.DB
	Engine *gin.Engine
	Ctx    context.Context
}

// NewTestEnvironment 创建新的测试环境
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// 内存 SQLite
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	// 自动迁移测试用到的实体
	if err := db.AutoMigrate(
		&entity.SysFile{},
	); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	return &TestEnvironment{
		T:      t,
		DB:     db,
		Engine: engine,
		Ctx:    context.Background(),
	}
}

// Request 执行 HTTP 请求并返回响应
func (env *TestEnvironment) Request(method, path string, body interface{}, headers map[string]string) *http.Response {
	env.T.Helper()

	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = bytes.NewBufferString(v)
		case []byte:
			reqBody = bytes.NewBuffer(v)
		default:
			data, err := json.Marshal(body)
			if err != nil {
				env.T.Fatalf("Failed to marshal body: %v", err)
			}
			reqBody = bytes.NewBuffer(data)
		}
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		env.T.Fatalf("Failed to create request: %v", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	env.Engine.ServeHTTP(w, req)
	return w.Result()
}

// AssertStatus 断言响应状态码
func (env *TestEnvironment) AssertStatus(resp *http.Response, expected int) {
	env.T.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		env.T.Errorf("Expected status %d, got %d. Body: %s", expected, resp.StatusCode, string(body))
	}
}

// ParseJSON 解析 JSON 响应
func (env *TestEnvironment) ParseJSON(resp *http.Response, target interface{}) {
	env.T.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		env.T.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(body, target); err != nil {
		env.T.Fatalf("Failed to unmarshal JSON: %v, body: %s", err, string(body))
	}
}

// WaitFor 等待条件成立（用于异步操作测试）
func (env *TestEnvironment) WaitFor(condition func() bool, timeout time.Duration, msg string) {
	env.T.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	env.T.Fatalf("WaitFor timeout: %s", msg)
}

// Cleanup 清理测试环境
func (env *TestEnvironment) Cleanup() {
	if env.DB != nil {
		if sqlDB, err := env.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}

// MakeJSONBody 辅助函数：将 map 转为 JSON body
func MakeJSONBody(data map[string]interface{}) string {
	b, _ := json.Marshal(data)
	return string(b)
}

// MockAuthMiddleware 模拟认证中间件，设置测试用户上下文
func MockAuthMiddleware(userID uint, companyID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("companyID", companyID)
		c.Set("username", fmt.Sprintf("test-user-%d", userID))
		c.Next()
	}
}