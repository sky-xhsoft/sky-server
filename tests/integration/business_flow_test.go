package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestIntegration_LoginAndAccessFlow 测试完整的登录-访问受保护资源流程
func TestIntegration_LoginAndAccessFlow(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// 模拟登录接口
	env.Engine.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request"})
			return
		}

		username, _ := req["username"].(string)
		password, _ := req["password"].(string)

		if username == "admin" && password == "admin123" {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "success",
				"data": gin.H{
					"token":        "fake-jwt-token",
					"refreshToken": "fake-refresh-token",
					"expiresIn":    3600,
					"user": gin.H{
						"id":       1,
						"username": "admin",
					},
				},
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Invalid credentials"})
	})

	// 受保护资源（需要 token）
	authGroup := env.Engine.Group("/api/v1/protected")
	authGroup.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || token != "Bearer fake-jwt-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Unauthorized"})
			return
		}
		c.Set("userID", uint(1))
		c.Next()
	})
	authGroup.GET("/me", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"userId": uid}})
	})

	// === 步骤 1: 失败的登录（无效凭据）===
	resp := env.Request("POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "wrong",
		"password": "wrong",
	}, nil)
	env.AssertStatus(resp, http.StatusUnauthorized)

	// === 步骤 2: 无 token 访问受保护资源 ===
	resp = env.Request("GET", "/api/v1/protected/me", nil, nil)
	env.AssertStatus(resp, http.StatusUnauthorized)

	// === 步骤 3: 成功登录 ===
	resp = env.Request("POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}, nil)
	env.AssertStatus(resp, http.StatusOK)

	var loginResp map[string]interface{}
	env.ParseJSON(resp, &loginResp)
	data, ok := loginResp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Login response should have data field, got: %v", loginResp)
	}
	token, _ := data["token"].(string)
	if token == "" {
		t.Fatal("Token should not be empty")
	}

	// === 步骤 4: 用 token 访问受保护资源 ===
	resp = env.Request("GET", "/api/v1/protected/me", nil, map[string]string{
		"Authorization": "Bearer " + token,
	})
	env.AssertStatus(resp, http.StatusOK)

	var meResp map[string]interface{}
	env.ParseJSON(resp, &meResp)
	if meData, ok := meResp["data"].(map[string]interface{}); ok {
		if uid, ok := meData["userId"].(float64); !ok || uid != 1 {
			t.Errorf("Expected userId=1, got %v", meData["userId"])
		}
	}

	// === 步骤 5: 错误的 token ===
	resp = env.Request("GET", "/api/v1/protected/me", nil, map[string]string{
		"Authorization": "Bearer invalid-token",
	})
	env.AssertStatus(resp, http.StatusUnauthorized)
}

// TestIntegration_CRUDFlow 测试基础 CRUD 流程
func TestIntegration_CRUDFlow(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// 内存存储模拟数据库
	items := make(map[uint]map[string]interface{})
	var nextID uint = 1

	api := env.Engine.Group("/api/v1/items")
	api.Use(MockAuthMiddleware(1, 1))

	api.GET("", func(c *gin.Context) {
		list := make([]map[string]interface{}, 0)
		for _, item := range items {
			list = append(list, item)
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"list": list, "total": len(list)}})
	})

	api.POST("", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		data["id"] = nextID
		items[nextID] = data
		nextID++
		c.JSON(http.StatusCreated, gin.H{"code": 200, "data": data})
	})

	api.DELETE("/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		var id uint
		json.Unmarshal([]byte(idStr), &id)
		delete(items, id)
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
	})

	// === 步骤 1: 初始列表为空 ===
	resp := env.Request("GET", "/api/v1/items", nil, nil)
	env.AssertStatus(resp, http.StatusOK)

	// === 步骤 2: 创建资源 ===
	resp = env.Request("POST", "/api/v1/items", map[string]interface{}{
		"name": "Test Item",
	}, nil)
	env.AssertStatus(resp, http.StatusCreated)

	// === 步骤 3: 列表应该有 1 条 ===
	resp = env.Request("GET", "/api/v1/items", nil, nil)
	env.AssertStatus(resp, http.StatusOK)
	var listResp map[string]interface{}
	env.ParseJSON(resp, &listResp)
	data := listResp["data"].(map[string]interface{})
	if total, ok := data["total"].(float64); !ok || total != 1 {
		t.Errorf("Expected total=1, got %v", data["total"])
	}
}
