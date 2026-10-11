package tencent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
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
	u, _ := url.Parse(cosServiceURL)
	client := cos.NewClient(&cos.BaseURL{ServiceURL: u}, &http.Client{
		Timeout:   15 * time.Second,
		Transport: &cos.AuthorizationTransport{SecretID: c.config.SecretID, SecretKey: c.config.SecretKey},
	})

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

// CreateCosBucket 创建存储桶（公开读写），并设置 CORS 允许所有来源、请求头和常用方法。
// bucket 为完整名称（含 -APPID）。存储桶已创建但 CORS 设置失败时返回 ErrCosBucketCORS
func (c *Client) CreateCosBucket(ctx context.Context, bucket, region string) error {
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		return err
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Timeout:   15 * time.Second,
		Transport: &cos.AuthorizationTransport{SecretID: c.config.SecretID, SecretKey: c.config.SecretKey},
	})
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
