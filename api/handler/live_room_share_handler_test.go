package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
	"gorm.io/gorm"
)

// shareLiveService 只实现分享页用到的域名查询：按公司返回播放域名
type shareLiveService struct {
	live.Service
	companies []uint // 被查询的公司
}

func (s *shareLiveService) ListDomains(ctx context.Context, domainType *int64) ([]*tencentLive.DomainInfo, error) {
	id, _ := ctx.Value(contextutil.CompanyIDKey).(uint)
	s.companies = append(s.companies, id)
	return []*tencentLive.DomainInfo{{Name: "new.play.test", Type: 1, Status: 1}, {Name: "play.test", Type: 1, Status: 1, BCName: 1}}, nil
}

func (s *shareLiveService) GetDomainAuthKey(ctx context.Context, domainName, keyType string) (*tencentLive.AuthKeyInfo, error) {
	return &tencentLive.AuthKeyInfo{MasterKey: "k"}, nil
}

var shareNow = time.Date(2026, 10, 12, 21, 0, 0, 0, time.Local)

func setupShareTest(t *testing.T) (*gin.Engine, *gorm.DB, *shareLiveService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.LiveRoom{}, &entity.BillLiveSession{}, &entity.SysCompany{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&entity.SysCompany{BaseModel: entity.BaseModel{ID: 52}, Name: "skyzhou", Code: "sky"})
	svc := &shareLiveService{}
	h := NewLiveRoomShareHandler(db, svc, live.NewPushURLService())
	h.now = func() time.Time { return shareNow }

	r := gin.New()
	auth := func(c *gin.Context) { c.Set("companyID", uint(52)); c.Set("username", "skyzhou"); c.Next() }
	r.GET("/live/rooms/:id/share", auth, h.GetShare)
	r.PUT("/live/rooms/:id/share", auth, h.UpdateShare)
	r.GET("/live/share/:code", h.GetPublicShare)
	return r, db, svc
}

func shareReq(t *testing.T, r *gin.Engine, method, path, body string) (int, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	data, _ := out["data"].(map[string]any)
	return w.Code, data
}

func createShareRoom(db *gorm.DB, id, company uint, status string) {
	start := shareNow.Add(2 * time.Hour)
	db.Create(&entity.LiveRoom{BaseModel: entity.BaseModel{ID: id, SysCompanyID: company, IsActive: "Y"}, RoomName: "发布会", BroadcastFormat: "live", Status: status, StartTime: &start, Description: "简介", ShareEnabled: "N"})
}

func TestShareStatus(t *testing.T) {
	ago := func(d time.Duration) *time.Time { v := shareNow.Add(-d); return &v }
	cases := []struct {
		status string
		ended  *time.Time
		want   string
	}{
		{"live", nil, ShareStatusLive},
		{"draft", nil, ShareStatusUpcoming},
		{"scheduled", nil, ShareStatusUpcoming},
		{"ended", ago(30 * time.Minute), ShareStatusEnded},
		{"ended", ago(time.Hour), ShareStatusEnded},
		{"ended", ago(61 * time.Minute), ShareStatusInvalid},
		{"ended", nil, ShareStatusInvalid},
		{"archived", nil, ShareStatusInvalid},
	}
	for _, c := range cases {
		if got := shareStatus(&entity.LiveRoom{Status: c.status}, c.ended, shareNow); got != c.want {
			t.Errorf("%s %v：得到 %s，应为 %s", c.status, c.ended, got, c.want)
		}
	}
}

func TestShare_EnableDisableReset(t *testing.T) {
	r, db, _ := setupShareTest(t)
	createShareRoom(db, 99, 52, "draft")
	createShareRoom(db, 98, 51, "draft")

	if code, data := shareReq(t, r, http.MethodGet, "/live/rooms/99/share", ""); code != 200 || data["enabled"] != false || data["code"] != "" {
		t.Fatalf("初始：%d %v", code, data)
	}
	_, data := shareReq(t, r, http.MethodPut, "/live/rooms/99/share", `{"enabled":true}`)
	first, _ := data["code"].(string)
	if data["enabled"] != true || len(first) != 24 {
		t.Fatalf("开启应生成分享码：%v", data)
	}
	// 关闭保留分享码，再开启沿用
	shareReq(t, r, http.MethodPut, "/live/rooms/99/share", `{"enabled":false}`)
	if _, data := shareReq(t, r, http.MethodPut, "/live/rooms/99/share", `{"enabled":true}`); data["code"] != first {
		t.Fatalf("再次开启应沿用分享码：%v", data)
	}
	// 重新生成：旧链接失效
	_, data = shareReq(t, r, http.MethodPut, "/live/rooms/99/share", `{"enabled":true,"reset":true}`)
	if data["code"] == first {
		t.Fatal("重新生成后分享码应变化")
	}
	if _, view := shareReq(t, r, http.MethodGet, "/live/share/"+first, ""); view["status"] != ShareStatusInvalid {
		t.Fatalf("旧链接应失效：%v", view)
	}
	// 其他公司的直播间
	if code, _ := shareReq(t, r, http.MethodPut, "/live/rooms/98/share", `{"enabled":true}`); code != http.StatusNotFound {
		t.Fatalf("其他公司的直播间应 404：%d", code)
	}
}

func TestShare_PublicView(t *testing.T) {
	r, db, svc := setupShareTest(t)
	createShareRoom(db, 99, 52, "draft")
	_, data := shareReq(t, r, http.MethodPut, "/live/rooms/99/share", `{"enabled":true}`)
	path := "/live/share/" + data["code"].(string)

	// 未开始：有开始时间与公司名，没有播放地址
	_, view := shareReq(t, r, http.MethodGet, path, "")
	if view["status"] != ShareStatusUpcoming || view["roomName"] != "发布会" || view["companyName"] != "skyzhou" || view["startTime"] == nil || view["playUrls"] != nil {
		t.Fatalf("未开始：%v", view)
	}

	// 直播中：用直播间所属公司、CNAME 已生效的播放域名生成地址
	db.Model(&entity.LiveRoom{}).Where("ID = 99").Update("STATUS", "live")
	_, view = shareReq(t, r, http.MethodGet, path, "")
	urls, _ := view["playUrls"].(map[string]any)
	if view["status"] != ShareStatusLive || !strings.HasPrefix(urls["flv"].(string), "http://play.test/live/99.flv?txSecret=") || urls["hls"] == nil || urls["rtmp"] != nil {
		t.Fatalf("直播中：%v", view)
	}
	if len(svc.companies) == 0 || svc.companies[0] != 52 {
		t.Fatalf("应按直播间所属公司查询域名：%v", svc.companies)
	}

	// 已结束：以最后一次断流时间为准，1 小时内显示已结束，超过失效
	db.Model(&entity.LiveRoom{}).Where("ID = 99").Update("STATUS", "ended")
	end := shareNow.Add(-40 * time.Minute)
	db.Create(&entity.BillLiveSession{SysCompanyID: 52, RoomID: 99, StartTime: end.Add(-time.Hour), EndTime: &end, LastSeenTime: end})
	if _, view = shareReq(t, r, http.MethodGet, path, ""); view["status"] != ShareStatusEnded || view["endedAt"] == nil || view["playUrls"] != nil {
		t.Fatalf("结束 40 分钟：%v", view)
	}
	db.Model(&entity.BillLiveSession{}).Where("ROOM_ID = 99").Update("END_TIME", shareNow.Add(-2*time.Hour))
	if _, view = shareReq(t, r, http.MethodGet, path, ""); view["status"] != ShareStatusInvalid || view["roomName"] != nil {
		t.Fatalf("结束 2 小时应失效且不返回直播间信息：%v", view)
	}

	// 关闭分享、直播间删除、乱码
	db.Model(&entity.LiveRoom{}).Where("ID = 99").Updates(map[string]any{"STATUS": "live", "SHARE_ENABLED": "N"})
	if _, view = shareReq(t, r, http.MethodGet, path, ""); view["status"] != ShareStatusInvalid {
		t.Fatalf("关闭分享：%v", view)
	}
	db.Model(&entity.LiveRoom{}).Where("ID = 99").Updates(map[string]any{"SHARE_ENABLED": "Y", "IS_ACTIVE": "N"})
	if _, view = shareReq(t, r, http.MethodGet, path, ""); view["status"] != ShareStatusInvalid {
		t.Fatalf("直播间删除：%v", view)
	}
	if _, view = shareReq(t, r, http.MethodGet, "/live/share/abc", ""); view["status"] != ShareStatusInvalid {
		t.Fatalf("乱码：%v", view)
	}
}
