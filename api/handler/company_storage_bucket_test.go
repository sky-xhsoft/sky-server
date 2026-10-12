package handler

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
)

// storageConfService 返回指定公司配置
type storageConfService struct {
	fakeCompanyService
	conf *entity.SysCompanyConf
}

func (s *storageConfService) GetCompanyConfig(ctx context.Context, companyID uint) (*entity.SysCompanyConf, error) {
	return s.conf, nil
}

func setupStorageBucketTest(t *testing.T, conf *entity.SysCompanyConf) (*gin.Engine, *[]string, *string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	usedKey := new(string)
	created := &[]string{}
	origList, origCreate := listStorageBuckets, createStorageBucket
	listStorageBuckets = func(ctx context.Context, id, key string) ([]tencent.CosBucket, error) {
		*usedKey = id
		return []tencent.CosBucket{{Name: "files-1504307838", Region: "ap-nanjing"}, {Name: "logs-1504307838", Region: "ap-shanghai"}}, nil
	}
	createStorageBucket = func(ctx context.Context, id, key, bucket, region string) error {
		if bucket == "taken-1504307838" {
			return stderrors.New("PUT https://x: 409 BucketAlreadyExists")
		}
		*created = append(*created, bucket+"|"+region)
		return nil
	}
	// 不访问云平台：现有存储桶里有 APPID 时直接取，否则视为查不到
	origAppID := storageAppID
	storageAppID = func(ctx context.Context, id, key string, buckets []tencent.CosBucket) string {
		return tencent.BucketsAppID(buckets)
	}
	t.Cleanup(func() { listStorageBuckets, createStorageBucket, storageAppID = origList, origCreate, origAppID })

	h := NewCompanyHandler(&storageConfService{conf: conf}, nil)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("companyID", uint(52)); c.Next() })
	r.GET("/cos-buckets", h.ListStorageBuckets)
	r.POST("/cos-buckets", h.CreateStorageBucket)
	return r, created, usedKey
}

func doJSON(r *gin.Engine, method, path, body string) (int, map[string]any) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestStorageBuckets_KeyChoice(t *testing.T) {
	// 单独设置的 COS 密钥优先
	r, _, used := setupStorageBucketTest(t, &entity.SysCompanyConf{TencentCOSSecretID: "cos-id", TencentCOSSecretKey: "k", TencentCloudSecretID: "live-id", TencentCloudSecretKey: "k"})
	if code, _ := doJSON(r, http.MethodGet, "/cos-buckets", ""); code != http.StatusOK || *used != "cos-id" {
		t.Fatalf("应使用 COS 密钥：%d %s", code, *used)
	}
	// 未单独设置时用直播密钥
	r, _, used = setupStorageBucketTest(t, &entity.SysCompanyConf{TencentCloudSecretID: "live-id", TencentCloudSecretKey: "k"})
	if code, _ := doJSON(r, http.MethodGet, "/cos-buckets", ""); code != http.StatusOK || *used != "live-id" {
		t.Fatalf("应使用直播密钥：%d %s", code, *used)
	}
	// 都没有
	r, _, _ = setupStorageBucketTest(t, &entity.SysCompanyConf{})
	if code, out := doJSON(r, http.MethodGet, "/cos-buckets", ""); code != http.StatusBadRequest || !strings.Contains(out["message"].(string), "请先保存云账号密钥") {
		t.Fatalf("无密钥：%d %v", code, out)
	}
}

func TestStorageBuckets_ListAndCreate(t *testing.T) {
	conf := &entity.SysCompanyConf{TencentCloudSecretID: "live-id", TencentCloudSecretKey: "k", TencentCOSBucketName: "files-1504307838", TencentCOSRegion: "ap-nanjing"}
	r, created, _ := setupStorageBucketTest(t, conf)

	_, out := doJSON(r, http.MethodGet, "/cos-buckets", "")
	list := out["data"].([]any)
	first := list[0].(map[string]any)
	if len(list) != 2 || first["bucket"] != "files-1504307838" || first["current"] != true || first["url"] != "https://files-1504307838.cos.ap-nanjing.myqcloud.com" {
		t.Fatalf("列表：%v", list)
	}

	code, out := doJSON(r, http.MethodPost, "/cos-buckets", `{"bucket":"disk-2026","region":"ap-shanghai"}`)
	data, _ := out["data"].(map[string]any)
	if code != http.StatusOK || data["bucket"] != "disk-2026-1504307838" || data["url"] != "https://disk-2026-1504307838.cos.ap-shanghai.myqcloud.com" {
		t.Fatalf("新建：%d %v", code, out)
	}
	if len(*created) != 1 || (*created)[0] != "disk-2026-1504307838|ap-shanghai" {
		t.Fatalf("调用腾讯云：%v", *created)
	}

	for body, want := range map[string]string{
		`{"bucket":"files","region":"ap-nanjing"}`:    "已存在",
		`{"bucket":"taken","region":"ap-nanjing"}`:    "已被占用",
		`{"bucket":"Bad_Name","region":"ap-nanjing"}`: "只能包含小写字母",
		`{"bucket":"x","region":""}`:                  "请选择地域",
	} {
		if code, out := doJSON(r, http.MethodPost, "/cos-buckets", body); code != http.StatusBadRequest || !strings.Contains(out["message"].(string), want) {
			t.Errorf("%s：应提示「%s」，得到 %d %v", body, want, code, out)
		}
	}
}

// 账号下还没有存储桶：用密钥向云平台查询 APPID，名称自动补 -APPID
func TestStorageBuckets_CreateFirstBucketFetchesAppID(t *testing.T) {
	conf := &entity.SysCompanyConf{TencentCloudSecretID: "live-id", TencentCloudSecretKey: "k"}
	r, created, _ := setupStorageBucketTest(t, conf)
	listStorageBuckets = func(ctx context.Context, id, key string) ([]tencent.CosBucket, error) { return nil, nil }

	// 查不到 APPID：提示输入完整名称
	if code, out := doJSON(r, http.MethodPost, "/cos-buckets", `{"bucket":"first","region":"ap-nanjing"}`); code != http.StatusBadRequest || !strings.Contains(out["message"].(string), "请输入完整名称") {
		t.Fatalf("查不到 APPID：%d %v", code, out)
	}

	var usedKey string
	storageAppID = func(ctx context.Context, id, key string, buckets []tencent.CosBucket) string {
		usedKey = id
		return "1504307838"
	}
	code, out := doJSON(r, http.MethodPost, "/cos-buckets", `{"bucket":"first","region":"ap-nanjing"}`)
	data, _ := out["data"].(map[string]any)
	if code != http.StatusOK || data["bucket"] != "first-1504307838" || usedKey != "live-id" {
		t.Fatalf("自动取 APPID：%d %v key=%s", code, out, usedKey)
	}
	if len(*created) != 1 || (*created)[0] != "first-1504307838|ap-nanjing" {
		t.Fatalf("调用腾讯云：%v", *created)
	}
}
