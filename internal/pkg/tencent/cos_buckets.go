package tencent

import (
	"context"
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
