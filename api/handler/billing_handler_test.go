package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/billing"
	"gorm.io/gorm"
)

func billingRouter(t *testing.T, companyID uint) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.LiveCallbackEvent{}, &entity.BillLiveSession{}, &entity.BillUsageDaily{}, &entity.BillFluxReconcile{}, &entity.BillMeterState{}); err != nil {
		t.Fatal(err)
	}
	db.Exec("CREATE TABLE sys_company (ID INTEGER PRIMARY KEY, NAME TEXT)")
	db.Create(&entity.BillUsageDaily{SysCompanyID: 50, UsageDate: "2026-10-10", Meter: entity.MeterLiveDuration, RoomID: 101, Quantity: 30, Unit: "分钟", SourceKey: "a"})
	h := NewBillingHandler(billing.NewMeter(db, billing.NewTencentAccounts(nil)))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if companyID != 0 {
			c.Set("companyID", companyID)
		}
	})
	r.GET("/summary", h.Summary)
	r.GET("/sessions", h.Sessions)
	r.GET("/reconcile", h.Reconcile)
	r.POST("/recompute", h.Recompute)
	return r
}

func call(r *gin.Engine, method, path, body string) (int, map[string]any) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestBillingHandler_CompanyScope(t *testing.T) {
	// 公司 50 只能看到自己的用量
	code, out := call(billingRouter(t, 50), "GET", "/summary?month=2026-10", "")
	if code != http.StatusOK || out["data"].(map[string]any)["liveMinutes"].(float64) != 30 {
		t.Fatalf("本公司用量：%d %v", code, out)
	}
	code, out = call(billingRouter(t, 51), "GET", "/summary?month=2026-10", "")
	if code != http.StatusOK || out["data"].(map[string]any)["liveMinutes"].(float64) != 0 {
		t.Fatalf("其他公司看不到：%d %v", code, out)
	}
	if code, _ := call(billingRouter(t, 0), "GET", "/summary", ""); code != http.StatusUnauthorized {
		t.Fatalf("没有公司信息应拒绝，got %d", code)
	}
	if code, _ := call(billingRouter(t, 50), "GET", "/summary?month=2026-13", ""); code != http.StatusBadRequest {
		t.Fatalf("非法月份应返回 400，got %d", code)
	}
}

func TestBillingHandler_PlatformOnly(t *testing.T) {
	r := billingRouter(t, 50)
	if code, _ := call(r, "GET", "/reconcile", ""); code != http.StatusForbidden {
		t.Fatalf("非平台公司不能看对账，got %d", code)
	}
	if code, _ := call(r, "POST", "/recompute", `{"from":"2026-10-01","to":"2026-10-01"}`); code != http.StatusForbidden {
		t.Fatalf("非平台公司不能重新计量，got %d", code)
	}
	p := billingRouter(t, platformCompanyID)
	if code, out := call(p, "GET", "/reconcile?month=2026-10", ""); code != http.StatusOK || len(out["data"].(map[string]any)["companies"].([]any)) != 1 {
		t.Fatalf("平台查看全部公司：%d %v", code, out)
	}
	if code, _ := call(p, "POST", "/recompute", `{"from":"2026-10-01"}`); code != http.StatusBadRequest {
		t.Fatalf("缺少结束日期应返回 400，got %d", code)
	}
	if code, _ := call(p, "POST", "/recompute", `{"from":"2026-10-01","to":"2026-10-01","meters":["flux"]}`); code != http.StatusBadRequest {
		t.Fatalf("不支持的计费点应返回 400，got %d", code)
	}
	if code, out := call(p, "POST", "/recompute", `{"from":"2026-10-01","to":"2026-10-01","meters":["live_duration"]}`); code != http.StatusOK {
		t.Fatalf("重新计量：%d %v", code, out)
	}
}
