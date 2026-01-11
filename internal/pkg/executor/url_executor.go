package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// URLExecutor URL调用执行器
type URLExecutor struct {
	client  *http.Client
	timeout time.Duration
}

// NewURLExecutor 创建URL执行器
func NewURLExecutor(timeout time.Duration) *URLExecutor {
	return &URLExecutor{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// URLRequest URL请求配置
type URLRequest struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`  // GET, POST, PUT, DELETE
	Headers map[string]string      `json:"headers"` // 请求头
	Body    map[string]interface{} `json:"body"`    // 请求体
	Params  map[string]interface{} `json:"params"`  // URL参数
}

// URLResponse URL响应
type URLResponse struct {
	StatusCode int                    `json:"statusCode"`
	Headers    map[string][]string    `json:"headers"`
	Body       string                 `json:"body"`
	BodyJSON   map[string]interface{} `json:"bodyJson"`
	Duration   time.Duration          `json:"duration"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error"`
}

// Execute 执行URL调用
func (e *URLExecutor) Execute(ctx context.Context, req *URLRequest) (*URLResponse, error) {
	start := time.Now()

	// 构建URL（添加查询参数）
	url := req.URL
	if len(req.Params) > 0 {
		url = e.buildURL(url, req.Params)
	}

	// 默认方法为GET
	method := req.Method
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// 构建请求体
	var bodyReader io.Reader
	if req.Body != nil && len(req.Body) > 0 {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInternal, "序列化请求体失败", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建HTTP请求失败", err)
	}

	// 设置请求头
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// 如果有请求体且没有设置Content-Type，默认使用application/json
	if bodyReader != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 执行请求
	resp, err := e.client.Do(httpReq)
	duration := time.Since(start)

	response := &URLResponse{
		Duration: duration,
		Success:  false,
	}

	if err != nil {
		response.Error = err.Error()
		return response, nil
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Sprintf("读取响应失败: %v", err)
		return response, nil
	}

	response.StatusCode = resp.StatusCode
	response.Headers = resp.Header
	response.Body = string(bodyBytes)

	// 尝试解析JSON响应
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var jsonBody map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
			response.BodyJSON = jsonBody
		}
	}

	// 判断是否成功（2xx状态码）
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response.Success = true
	} else {
		response.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return response, nil
}

// buildURL 构建带查询参数的URL
func (e *URLExecutor) buildURL(baseURL string, params map[string]interface{}) string {
	if len(params) == 0 {
		return baseURL
	}

	var queryParts []string
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%v", key, value))
	}

	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}

	return baseURL + separator + strings.Join(queryParts, "&")
}
