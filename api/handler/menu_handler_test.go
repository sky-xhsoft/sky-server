package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// recordingMenuService 记录菜单接口收到的系统类型
type recordingMenuService struct{ systemType string }

func (r *recordingMenuService) GetMenuTree(_ context.Context, _ uint, systemType string) ([]*entity.MenuNode, error) {
	r.systemType = systemType
	return nil, nil
}

func (r *recordingMenuService) GetUserMenuTree(_ context.Context, _, _ uint, systemType string) ([]*entity.MenuNode, error) {
	r.systemType = systemType
	return nil, nil
}

func (r *recordingMenuService) GetSubsystems(context.Context) ([]*entity.SysSubsystem, error) {
	return nil, nil
}

func TestMenuSystemTypeFallsBackToClientType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name, query, clientType, want string
	}{
		{"显式 type 优先", "?type=live", "cloud-web", "live"},
		{"未传 type 时按云盘登录端", "", "cloud-web", "cloud"},
		{"未传 type 时按直播登录端", "", "live-web", "live"},
		{"未传 type 时按系统中心登录端", "", "system-web", "system"},
		{"无法推断时不限制", "", "mobile", ""},
	}
	for _, path := range []string{"/menus/user/tree", "/menus/tree"} {
		for _, tc := range cases {
			svc := &recordingMenuService{}
			h := NewMenuHandler(svc)
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("userID", uint(1))
				c.Set("companyID", uint(1))
				c.Set("clientType", tc.clientType)
			})
			r.GET("/menus/user/tree", h.GetUserMenuTree)
			r.GET("/menus/tree", h.GetMenuTree)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path+tc.query, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("%s %s: HTTP %d", path, tc.name, w.Code)
			}
			if svc.systemType != tc.want {
				t.Errorf("%s %s: systemType = %q, want %q", path, tc.name, svc.systemType, tc.want)
			}
		}
	}
}
