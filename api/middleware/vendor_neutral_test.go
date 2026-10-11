package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func vendorNeutralRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(VendorNeutralMessages())
	r.GET("/err", func(c *gin.Context) {
		c.JSON(400, gin.H{"code": 400, "message": "获取腾讯云存储桶失败：" + errors.New("Get https://b-1250000000.cos.ap-guangzhou.myqcloud.com/: 403").Error()})
	})
	r.GET("/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 200, "data": gin.H{
			"id":         int64(9007199254740993),
			"bucketUrl":  "https://b-1250000000.cos.ap-guangzhou.myqcloud.com",
			"syncError":  "腾讯云上已不存在该模板",
			"issues":     []gin.H{{"level": "warning", "message": "请在腾讯云控制台开启"}},
			"cloudError": "[TencentCloudSDKError] Code=AuthFailure",
		}})
	})
	r.GET("/file", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("腾讯云"))
	})
	return r
}

func TestVendorNeutralMessages(t *testing.T) {
	r := vendorNeutralRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/err", nil))
	if w.Code != 400 || strings.Contains(w.Body.String(), "腾讯") || strings.Contains(w.Body.String(), "myqcloud") {
		t.Fatalf("错误提示未替换：%d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/data", nil))
	body := w.Body.String()
	for _, want := range []string{
		`"bucketUrl":"https://b-1250000000.cos.ap-guangzhou.myqcloud.com"`, // 业务数据原样
		`"id":9007199254740993`, // 大整数精度不丢
		`"syncError":"云平台上已不存在该模板"`,
		`"message":"请在云平台控制台开启"`,
		`"cloudError":"[CloudSDKError] Code=AuthFailure"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("响应缺少 %s：%s", want, body)
		}
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/file", nil))
	if w.Body.String() != "腾讯云" {
		t.Errorf("非 JSON 响应不应改动：%s", w.Body.String())
	}
}

func TestVendorNeutralMessages_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				c.JSON(500, gin.H{"code": 500, "message": "服务器内部错误"})
			}
		}()
		c.Next()
	})
	r.Use(VendorNeutralMessages())
	r.GET("/panic", func(c *gin.Context) { panic("boom") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if w.Code != 500 || !strings.Contains(w.Body.String(), "服务器内部错误") {
		t.Fatalf("panic 后外层应能写出错误响应：%d %q", w.Code, w.Body.String())
	}
}
