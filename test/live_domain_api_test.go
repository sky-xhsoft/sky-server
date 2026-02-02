package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// 测试配置
const (
	baseURL = "http://localhost:9090/api/v1"
	token   = "YOUR_JWT_TOKEN" // 替换为实际的 JWT token
)

// AddDomainRequest 添加域名请求
type AddDomainRequest struct {
	DomainName string `json:"domainName"`
	DomainType int64  `json:"domainType"`
}

// Response 响应结构
type Response struct {
	Code      int                    `json:"code"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

func main() {
	fmt.Println("=== 腾讯云直播域名管理 API 测试 ===\n")

	// 1. 添加域名
	fmt.Println("1. 测试添加域名...")
	addDomain("push.example.com", 0)

	// 2. 查询域名列表
	fmt.Println("\n2. 测试查询域名列表...")
	listDomains()

	// 3. 查询域名信息
	fmt.Println("\n3. 测试查询域名信息...")
	getDomain("push.example.com")

	// 4. 启用域名
	fmt.Println("\n4. 测试启用域名...")
	enableDomain("push.example.com")

	// 5. 禁用域名
	fmt.Println("\n5. 测试禁用域名...")
	forbidDomain("push.example.com")

	// 6. 删除域名
	fmt.Println("\n6. 测试删除域名...")
	deleteDomain("push.example.com")

	fmt.Println("\n=== 测试完成 ===")
}

// addDomain 添加域名
func addDomain(domainName string, domainType int64) {
	req := AddDomainRequest{
		DomainName: domainName,
		DomainType: domainType,
	}

	body, _ := json.Marshal(req)
	resp := doRequest("POST", "/live/domains", bytes.NewReader(body))
	printResponse(resp)
}

// listDomains 查询域名列表
func listDomains() {
	resp := doRequest("GET", "/live/domains?domainType=0", nil)
	printResponse(resp)
}

// getDomain 查询域名信息
func getDomain(domainName string) {
	resp := doRequest("GET", "/live/domains/"+domainName, nil)
	printResponse(resp)
}

// enableDomain 启用域名
func enableDomain(domainName string) {
	resp := doRequest("POST", "/live/domains/"+domainName+"/enable", nil)
	printResponse(resp)
}

// forbidDomain 禁用域名
func forbidDomain(domainName string) {
	resp := doRequest("POST", "/live/domains/"+domainName+"/forbid", nil)
	printResponse(resp)
}

// deleteDomain 删除域名
func deleteDomain(domainName string) {
	resp := doRequest("DELETE", "/live/domains/"+domainName, nil)
	printResponse(resp)
}

// doRequest 发送 HTTP 请求
func doRequest(method, path string, body io.Reader) *Response {
	url := baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return nil
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return nil
	}

	return &result
}

// printResponse 打印响应
func printResponse(resp *Response) {
	if resp == nil {
		return
	}

	fmt.Printf("状态码: %d\n", resp.Code)
	fmt.Printf("消息: %s\n", resp.Message)

	if resp.Data != nil {
		data, _ := json.MarshalIndent(resp.Data, "", "  ")
		fmt.Printf("数据: %s\n", string(data))
	}
}
