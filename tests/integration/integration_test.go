package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/middleware"
)

func TestIntegration_HealthEndpoint(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	resp := env.Request("GET", "/health", nil, nil)
	env.AssertStatus(resp, http.StatusOK)

	var data map[string]interface{}
	env.ParseJSON(resp, &data)
	if data["status"] != "ok" {
		t.Errorf("Expected status=ok, got %v", data["status"])
	}
}

func TestIntegration_MiddlewareChain(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// 应用多个中间件
	env.Engine.Use(middleware.RequestID())
	env.Engine.Use(middleware.Compression())

	env.Engine.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": "test"})
	})

	resp := env.Request("GET", "/api/test", nil, nil)
	env.AssertStatus(resp, http.StatusOK)

	// 验证 RequestID 中间件添加了 X-Request-ID 头
	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header not set")
	}
}

func TestIntegration_AuthMiddleware(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	authGroup := env.Engine.Group("/api")
	authGroup.Use(MockAuthMiddleware(123, 1))

	authGroup.GET("/me", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(200, gin.H{"userID": userID})
	})

	resp := env.Request("GET", "/api/me", nil, nil)
	env.AssertStatus(resp, http.StatusOK)

	var data map[string]interface{}
	env.ParseJSON(resp, &data)

	// JSON 解析数字时会变成 float64
	if uid, ok := data["userID"].(float64); !ok || uid != 123 {
		t.Errorf("Expected userID=123, got %v", data["userID"])
	}
}

func TestIntegration_JSONPostRequest(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.Engine.POST("/api/echo", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	})

	payload := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	resp := env.Request("POST", "/api/echo", payload, nil)
	env.AssertStatus(resp, http.StatusOK)

	var result map[string]interface{}
	env.ParseJSON(resp, &result)

	if result["name"] != "test" {
		t.Errorf("Expected name=test, got %v", result["name"])
	}
}

func TestIntegration_ErrorResponse(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.Engine.GET("/api/error", func(c *gin.Context) {
		c.JSON(500, gin.H{
			"code":    50000,
			"message": "Internal error",
		})
	})

	resp := env.Request("GET", "/api/error", nil, nil)
	env.AssertStatus(resp, http.StatusInternalServerError)

	var data map[string]json.RawMessage
	env.ParseJSON(resp, &data)

	if _, ok := data["message"]; !ok {
		t.Error("Expected message field in error response")
	}
}

func TestIntegration_MakeJSONBody(t *testing.T) {
	body := MakeJSONBody(map[string]interface{}{
		"key": "value",
		"num": 42,
	})

	if body == "" {
		t.Error("Body should not be empty")
	}

	// 验证可以解析回来
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if parsed["key"] != "value" {
		t.Errorf("Expected key=value, got %v", parsed["key"])
	}
}
