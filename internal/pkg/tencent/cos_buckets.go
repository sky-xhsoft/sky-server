package tencent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// cosServiceURL 列出账号下所有地域的存储桶
const cosServiceURL = "https://service.cos.myqcloud.com"

// CosBucket 腾讯云账号下的 COS 存储桶
type CosBucket struct {
	Name         string `json:"name"`   // 完整名称，含 -appid
	Region       string `json:"region"` // 如 ap-shanghai
	CreationDate string `json:"creationDate"`
}

// ListCosBuckets 用本客户端的密钥列出账号下全部存储桶（按地域、名称排序）
func (c *Client) ListCosBuckets(ctx context.Context) ([]CosBucket, error) {
	return ListCosBuckets(ctx, c.config.SecretID, c.config.SecretKey)
}

// CreateCosBucket 用本客户端的密钥创建存储桶，见 CreateCosBucket
func (c *Client) CreateCosBucket(ctx context.Context, bucket, region string) error {
	return CreateCosBucket(ctx, c.config.SecretID, c.config.SecretKey, bucket, region)
}

func cosHTTPClient(secretID, secretKey string) *http.Client {
	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: &cos.AuthorizationTransport{SecretID: secretID, SecretKey: secretKey},
	}
}

// ListCosBuckets 用指定密钥列出账号下全部存储桶（按地域、名称排序）
func ListCosBuckets(ctx context.Context, secretID, secretKey string) ([]CosBucket, error) {
	u, _ := url.Parse(cosServiceURL)
	client := cos.NewClient(&cos.BaseURL{ServiceURL: u}, cosHTTPClient(secretID, secretKey))

	var out []CosBucket
	opt := &cos.ServiceGetOptions{MaxKeys: 1000}
	for {
		res, _, err := client.Service.Get(ctx, opt)
		if err != nil {
			return nil, err
		}
		for _, b := range res.Buckets {
			out = append(out, CosBucket{Name: b.Name, Region: b.Region, CreationDate: b.CreationDate})
		}
		if !res.IsTruncated || res.NextMarker == "" || res.NextMarker == opt.Marker {
			break
		}
		opt.Marker = res.NextMarker
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Region != out[j].Region {
			return out[i].Region < out[j].Region
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// CreateCosBucket 用指定密钥创建存储桶（公开读写），并设置 CORS 允许所有来源、请求头和常用方法。
// bucket 为完整名称（含 -APPID）。存储桶已创建但 CORS 设置失败时返回 ErrCosBucketCORS
func CreateCosBucket(ctx context.Context, secretID, secretKey, bucket, region string) error {
	u, err := url.Parse(CosBucketURL(bucket, region))
	if err != nil {
		return err
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, cosHTTPClient(secretID, secretKey))
	if _, err := client.Bucket.Put(ctx, &cos.BucketPutOptions{XCosACL: "public-read-write"}); err != nil {
		return err
	}
	_, err = client.Bucket.PutCORS(ctx, &cos.BucketPutCORSOptions{Rules: []cos.BucketCORSRule{{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "HEAD"},
		AllowedHeaders: []string{"*"},
		ExposeHeaders:  []string{"*"},
		MaxAgeSeconds:  600,
	}}})
	if err != nil {
		return fmt.Errorf("%w：%v", ErrCosBucketCORS, err)
	}
	return nil
}

// ErrCosBucketCORS 存储桶已创建，但设置跨域访问失败
var ErrCosBucketCORS = errors.New("存储桶已创建，但设置跨域访问（CORS）失败，请在腾讯云控制台手动设置")

// CosBucketURL 存储桶访问地址
func CosBucketURL(bucket, region string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
}

var (
	// 存储桶名称：小写字母、数字和 -，不能以 - 开头或结尾；地域如 ap-nanjing；APPID 为 10 位数字
	cosBucketNameRe  = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	cosRegionRe      = regexp.MustCompile(`^[a-z]+(-[a-z]+)+(-[0-9]+)?$`)
	cosAppIDSuffixRe = regexp.MustCompile(`-[0-9]{10}$`)
)

// cosBucketMaxLen 存储桶完整名称（名称-APPID）最长 50 个字符
const cosBucketMaxLen = 50

// ResolveNewCosBucket 校验新建存储桶的名称和地域，返回不含 / 含 -APPID 的名称。
// name 可不带 -APPID（取自账号现有存储桶 buckets）；名称本身可以以数字结尾，只有与账号 APPID 一致的后缀才视为 -APPID
func ResolveNewCosBucket(buckets []CosBucket, name, region string) (short, full, cleanRegion string, err error) {
	name = strings.ToLower(strings.TrimSpace(name))
	region = strings.TrimSpace(region)
	if !cosRegionRe.MatchString(region) {
		return "", "", "", errors.New("请选择地域")
	}
	appID := ""
	for _, b := range buckets {
		if i := strings.LastIndex(b.Name, "-"); i > 0 && cosAppIDSuffixRe.MatchString(b.Name) {
			appID = b.Name[i+1:]
			break
		}
	}
	full = name
	switch {
	case appID != "" && strings.HasSuffix(name, "-"+appID):
		name = strings.TrimSuffix(name, "-"+appID)
	case appID != "":
		full = name + "-" + appID
	case cosAppIDSuffixRe.MatchString(name):
		name = name[:strings.LastIndex(name, "-")]
	default:
		return "", "", "", errors.New("账号下还没有存储桶，无法确定 APPID，请输入完整名称（名称-APPID）")
	}
	if !cosBucketNameRe.MatchString(name) {
		return "", "", "", errors.New("存储桶名称只能包含小写字母、数字和 -，且不能以 - 开头或结尾")
	}
	if len(full) > cosBucketMaxLen {
		return "", "", "", fmt.Errorf("存储桶名称过长：加上 -APPID 后不能超过 %d 个字符", cosBucketMaxLen)
	}
	for _, b := range buckets {
		if b.Name == full {
			return "", "", "", fmt.Errorf("存储桶 %s 已存在（%s）", full, b.Region)
		}
	}
	return name, full, region, nil
}

// CosCreateError 腾讯云创建存储桶的常见错误转为中文提示（CORS 设置失败原样返回）
func CosCreateError(err error) error {
	if err == nil || errors.Is(err, ErrCosBucketCORS) {
		return err
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "BucketAlreadyExists"):
		return errors.New("存储桶名称已被占用（腾讯云内全局唯一），请换一个名称")
	case strings.Contains(msg, "TooManyBuckets"):
		return errors.New("腾讯云账号的存储桶数量已达上限（200 个）")
	case strings.Contains(msg, "arrears"), strings.Contains(msg, " 451 "):
		return errors.New("腾讯云账号欠费，充值后才能创建存储桶")
	case strings.Contains(msg, "AccessDenied"):
		return errors.New("腾讯云密钥没有创建存储桶的权限（需要 cos:PutBucket）")
	case strings.Contains(msg, "InvalidRegionName"), strings.Contains(msg, "no such host"):
		return errors.New("地域不存在或不支持")
	}
	return fmt.Errorf("创建存储桶失败：%w", err)
}
