package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/audit"
)

// AuditLogger 审计日志中间件
func AuditLogger(auditService audit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 读取请求体(需要保存以便后续使用)
		var requestBody string
		if c.Request.Body != nil && shouldLogBody(c.Request.Method) {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// 重新设置请求体,供后续handler使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 创建响应写入器以捕获响应
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 继续处理请求
		c.Next()

		// 计算执行时长
		duration := time.Since(startTime).Milliseconds()

		// 获取用户信息
		var userID uint
		var username string
		if uid, exists := c.Get("userID"); exists {
			userID = uid.(uint)
		}
		if uname, exists := c.Get("username"); exists {
			username = uname.(string)
		}

		// 判断操作状态
		status := entity.StatusSuccess
		if c.Writer.Status() >= 400 {
			status = entity.StatusFailure
		}

		// 解析操作类型和资源
		action, resource := parseActionAndResource(c.Request.Method, c.Request.URL.Path)

		// 构建审计日志
		log := audit.NewLogBuilder().
			WithUser(userID, username).
			WithAction(action).
			WithResource(resource, "", "").
			WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.Request.UserAgent()).
			WithStatus(status).
			WithDuration(duration).
			Build()

		// 设置请求体(过滤敏感信息)
		if requestBody != "" && len(requestBody) < 10000 { // 限制大小
			log.RequestBody = filterSensitiveData(requestBody)
		}

		// 设置响应体(仅在失败时记录,或者根据配置)
		if status == entity.StatusFailure && responseWriter.body.Len() < 10000 {
			log.ResponseBody = responseWriter.body.String()
		}

		// 异步记录日志(不阻塞请求)
		auditService.LogAsync(log)
	}
}

// responseBodyWriter 响应体写入器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldLogBody 判断是否应该记录请求体
func shouldLogBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// parseActionAndResource 解析操作类型和资源类型
func parseActionAndResource(method, path string) (action, resource string) {
	// 根据HTTP方法映射操作类型
	switch method {
	case "GET":
		if strings.Contains(path, "/query") || strings.Contains(path, "/list") {
			action = entity.ActionQuery
		} else {
			action = entity.ActionRead
		}
	case "POST":
		if strings.Contains(path, "/execute") {
			action = entity.ActionExecute
		} else if strings.Contains(path, "/login") {
			action = entity.ActionLogin
		} else if strings.Contains(path, "/logout") {
			action = entity.ActionLogout
		} else if strings.Contains(path, "/start") {
			action = entity.ActionStartProcess
		} else if strings.Contains(path, "/complete") {
			action = entity.ActionCompleteTask
		} else if strings.Contains(path, "/claim") {
			action = entity.ActionClaimTask
		} else if strings.Contains(path, "/transfer") {
			action = entity.ActionTransferTask
		} else if strings.Contains(path, "/publish") {
			action = entity.ActionPublishWorkflow
		} else {
			action = entity.ActionCreate
		}
	case "PUT", "PATCH":
		action = entity.ActionUpdate
	case "DELETE":
		action = entity.ActionDelete
	default:
		action = "unknown"
	}

	// 根据路径解析资源类型
	if strings.Contains(path, "/auth") || strings.Contains(path, "/users") {
		resource = entity.ResourceUser
	} else if strings.Contains(path, "/data") {
		resource = entity.ResourceTable
	} else if strings.Contains(path, "/actions") {
		resource = entity.ResourceAction
	} else if strings.Contains(path, "/workflow") {
		resource = entity.ResourceWorkflow
	} else if strings.Contains(path, "/tasks") {
		resource = entity.ResourceTask
	} else if strings.Contains(path, "/dicts") {
		resource = entity.ResourceDict
	} else if strings.Contains(path, "/sequences") {
		resource = entity.ResourceSequence
	} else {
		resource = "unknown"
	}

	return
}

// filterSensitiveData 过滤敏感数据
func filterSensitiveData(data string) string {
	// 移除密码等敏感字段
	// 这里使用简单的字符串替换,生产环境应使用更复杂的JSON解析和过滤
	sensitiveFields := []string{"password", "token", "secret", "accessToken", "refreshToken"}

	filtered := data
	for _, field := range sensitiveFields {
		// 简单替换包含敏感字段的值
		if strings.Contains(strings.ToLower(filtered), strings.ToLower(field)) {
			// TODO: 使用JSON解析替换具体字段值为***
			filtered = strings.ReplaceAll(filtered, field, field+":[FILTERED]")
		}
	}

	return filtered
}
