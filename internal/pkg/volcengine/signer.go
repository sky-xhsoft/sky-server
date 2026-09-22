package volcengine

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Signer 火山引擎API签名器
type Signer struct {
	AccessKeyId     string
	AccessKeySecret string
	Region          string
	Service         string
}

// NewSigner 创建火山引擎API签名器
func NewSigner(accessKeyId, accessKeySecret, region, service string) *Signer {
	return &Signer{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		Region:          region,
		Service:         service,
	}
}

// SignRequest 给HTTP请求签名
func (s *Signer) SignRequest(req *http.Request, body []byte) error {
	now := time.Now().UTC()

	// 先设置参与签名的头（Host、X-Date、X-Content-Sha256），再构建CanonicalRequest
	req.Header.Set("X-Date", now.Format("20060102T150405Z"))
	if len(body) > 0 {
		hash := sha256.Sum256(body)
		req.Header.Set("X-Content-Sha256", hex.EncodeToString(hash[:]))
	}

	// 步骤1: 创建CanonicalRequest
	canonicalRequest := s.buildCanonicalRequest(req, body)
	_, signedHeaders := s.getCanonicalHeaders(req)

	// 步骤2: 创建StringToSign
	stringToSign := s.buildStringToSign(canonicalRequest, now)

	// 步骤3: 计算签名
	signature := s.calculateSignature(stringToSign, now)

	// 步骤4: 添加Authorization头（SignedHeaders 用实际参与签名的头）
	authorization := fmt.Sprintf("HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.AccessKeyId, s.getCredentialScope(now), signedHeaders, signature)
	req.Header.Set("Authorization", authorization)

	// 把 host 头也写到请求头中（便于 Go http 发送时保持一致）
	if req.Header.Get("Host") == "" && req.Host != "" {
		req.Header.Set("Host", req.Host)
	}

	return nil
}

// buildCanonicalRequest 构建规范请求
func (s *Signer) buildCanonicalRequest(req *http.Request, body []byte) string {
	var buf bytes.Buffer

	// HTTP请求方法
	buf.WriteString(req.Method)
	buf.WriteString("\n")

	// CanonicalURI
	buf.WriteString(s.getCanonicalURI(req.URL))
	buf.WriteString("\n")

	// CanonicalQueryString
	buf.WriteString(s.getCanonicalQueryString(req.URL))
	buf.WriteString("\n")

	// CanonicalHeaders
	canonicalHeaders, signedHeaders := s.getCanonicalHeaders(req)
	buf.WriteString(canonicalHeaders)
	buf.WriteString("\n")

	// SignedHeaders
	buf.WriteString(signedHeaders)
	buf.WriteString("\n")

	// HashedPayload
	if len(body) > 0 {
		hash := sha256.Sum256(body)
		buf.WriteString(hex.EncodeToString(hash[:]))
	} else {
		buf.WriteString("UNSIGNED-PAYLOAD")
	}

	return buf.String()
}

// getCanonicalURI 获取规范URI
func (s *Signer) getCanonicalURI(u *url.URL) string {
	if u.Path == "" {
		return "/"
	}
	// 火山引擎规范：对 path 做 RFC3986 编码，保留 "/"（对每段分别编码后拼接）
	segments := strings.Split(u.Path, "/")
	encoded := make([]string, 0, len(segments))
	for _, seg := range segments {
		encoded = append(encoded, url.PathEscape(seg))
	}
	return strings.Join(encoded, "/")
}

// getCanonicalQueryString 获取规范查询字符串
func (s *Signer) getCanonicalQueryString(u *url.URL) string {
	query := u.Query()
	var keys []string
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		vals := query[k]
		sort.Strings(vals)
		for _, v := range vals {
			parts = append(parts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
		}
	}
	return strings.Join(parts, "&")
}

// getCanonicalHeaders 获取规范请求头
func (s *Signer) getCanonicalHeaders(req *http.Request) (canonicalHeaders, signedHeaders string) {
	// host 头：Go 的 Host 在 req.Host 字段，需单独并入
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	headerMap := make(map[string][]string)
	headerMap["host"] = []string{host}

	for k, v := range req.Header {
		key := strings.ToLower(k)
		// 只签名以下类型的请求头
		if key == "content-type" ||
			strings.HasPrefix(key, "x-") {
			headerMap[key] = v
		}
	}

	// headers 按小写 ASCII 排序
	var headers []string
	for k := range headerMap {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	for _, k := range headers {
		buf.WriteString(k)
		buf.WriteString(":")
		vals := headerMap[k]
		for i, v := range vals {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(strings.TrimSpace(v))
		}
		buf.WriteString("\n")
	}

	return buf.String(), strings.Join(headers, ";")
}

// buildStringToSign 构建待签名字符串
func (s *Signer) buildStringToSign(canonicalRequest string, t time.Time) string {
	var buf bytes.Buffer

	// 算法
	buf.WriteString("HMAC-SHA256")
	buf.WriteString("\n")

	// 请求日期
	buf.WriteString(t.Format("20060102T150405Z"))
	buf.WriteString("\n")

	// 凭证范围
	buf.WriteString(s.getCredentialScope(t))
	buf.WriteString("\n")

	// 规范请求的哈希
	hash := sha256.Sum256([]byte(canonicalRequest))
	buf.WriteString(hex.EncodeToString(hash[:]))

	return buf.String()
}

// getCredentialScope 获取凭证范围
func (s *Signer) getCredentialScope(t time.Time) string {
	return fmt.Sprintf("%s/%s/%s/request", t.Format("20060102"), s.Region, s.Service)
}

// calculateSignature 计算签名
func (s *Signer) calculateSignature(stringToSign string, t time.Time) string {
	// 步骤1: 计算kDate，火山引擎 kSecret 即为 Secret Access Key 本身
	kDate := hmacSHA256([]byte(s.AccessKeySecret), []byte(t.Format("20060102")))

	// 步骤2: 计算kRegion
	kRegion := hmacSHA256(kDate, []byte(s.Region))

	// 步骤3: 计算kService
	kService := hmacSHA256(kRegion, []byte(s.Service))

	// 步骤4: 计算kSigning
	kSigning := hmacSHA256(kService, []byte("request"))

	// 步骤5: 计算签名
	signature := hmacSHA256(kSigning, []byte(stringToSign))

	return hex.EncodeToString(signature)
}

// hmacSHA256 计算HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
