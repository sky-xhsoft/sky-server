// Package vendortext 把返回给用户的提示文字中的第三方云平台名称（腾讯云、火山引擎等）与服务地址替换为中性说法。
// 只用于提示类文字（错误信息、同步状态、体检结果）；业务数据中的地址（存储桶、播放地址等）不能经过这里。
package vendortext

import (
	"regexp"
	"strings"
)

const hostPattern = `(?:myqcloud\.com|tencentcloudapi\.com|qcloud\.com|tencent\.com|tencentcos\.cn|volces\.com|volcengineapi\.com|volcengine\.com|bytedance\.com|bytedanceapi\.com|byteplus\.com)`

var (
	// 带协议的地址整段替换；不带协议的主机名单独替换
	urlRe  = regexp.MustCompile(`(?i)https?://[^\s"'<>，。；）)]*` + hostPattern + `[^\s"'<>，。；）)]*`)
	hostRe = regexp.MustCompile(`(?i)[a-z0-9][a-z0-9.-]*\.` + hostPattern)

	wordReplacer = strings.NewReplacer(
		"TencentCloudSDKError", "CloudSDKError",
		"腾讯云", "云平台",
		"腾讯", "云平台",
		"火山方舟", "智能服务",
		"火山引擎", "云服务",
		"火山", "云服务",
	)
	enRe  = regexp.MustCompile(`(?i)tencent\s*cloud|volcano\s*engine|volcengine|byteplus|bytedance`)
	tcRe  = regexp.MustCompile(`(?i)tencent`)
	cosRe = regexp.MustCompile(`\bCOS\b`)
)

// Sanitize 返回去掉第三方云平台名称与服务地址后的提示文字
func Sanitize(s string) string {
	if s == "" || !Contains(s) {
		return s
	}
	s = urlRe.ReplaceAllStringFunc(s, func(u string) string {
		// 地址后紧跟的冒号等标点属于句子，不属于地址
		trimmed := strings.TrimRight(u, ":.,;")
		return "[云服务地址]" + u[len(trimmed):]
	})
	s = hostRe.ReplaceAllString(s, "[云服务地址]")
	s = wordReplacer.Replace(s)
	s = enRe.ReplaceAllString(s, "cloud service")
	s = tcRe.ReplaceAllString(s, "cloud")
	s = cosRe.ReplaceAllString(s, "对象存储")
	return s
}

var detectRe = regexp.MustCompile(`(?i)腾讯|火山|tencent|volc|byteplus|bytedance|qcloud|\bCOS\b`)

// Contains 文字中是否含有需要替换的第三方名称或地址
func Contains(s string) bool {
	return detectRe.MatchString(s)
}
