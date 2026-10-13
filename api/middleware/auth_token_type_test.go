package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
)

// 只接受访问令牌，刷新令牌不能直接访问接口
func TestAuthRequired_TokenType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	j := jwt.New("test-secret-key-for-unit-tests-only")
	r := gin.New()
	r.GET("/me", AuthRequired(j), func(c *gin.Context) { c.Status(http.StatusOK) })

	access, _ := j.GenerateToken(1, 1, "admin", "web", "dev", time.Hour)
	refresh, _ := j.GenerateRefreshToken(1, 1, "admin", "web", "dev", time.Hour)
	for _, tc := range []struct {
		name  string
		token string
		want  int
	}{
		{"访问令牌", access, http.StatusOK},
		{"刷新令牌", refresh, http.StatusUnauthorized},
	} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", "Bearer "+tc.token)
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s: 期望 %d，实际 %d %s", tc.name, tc.want, w.Code, w.Body.String())
		}
	}
}
