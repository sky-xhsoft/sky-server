package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// UserStatusLookup 查询用户是否启用、是否管理员（路由注册时注入，中间件不直接依赖数据库）
type UserStatusLookup func(userID uint) (active bool, admin bool, err error)

// userStatusTTL 用户状态缓存时间：账号停用后最迟这么久旧令牌失效
const userStatusTTL = 30 * time.Second

type cachedUserStatus struct {
	active bool
	admin  bool
	at     time.Time
}

var (
	userStatusLookup UserStatusLookup
	userStatusCache  sync.Map // userID -> cachedUserStatus
)

// SetUserStatusLookup 设置用户状态查询函数；未设置时不校验用户状态
func SetUserStatusLookup(fn UserStatusLookup) {
	userStatusLookup = fn
	userStatusCache = sync.Map{}
}

// InvalidateUserStatus 清除某个用户的状态缓存（停用、改管理员标志后调用可立即生效）
func InvalidateUserStatus(userID uint) {
	userStatusCache.Delete(userID)
}

// resolveUserStatus 返回用户状态；ok=false 表示无法判断（未注入查询函数或查询出错），此时不拦截
func resolveUserStatus(userID uint) (active, admin, ok bool) {
	if userStatusLookup == nil {
		return true, false, false
	}
	if v, found := userStatusCache.Load(userID); found {
		st := v.(cachedUserStatus)
		if time.Since(st.at) < userStatusTTL {
			return st.active, st.admin, true
		}
	}
	active, admin, err := userStatusLookup(userID)
	if err != nil {
		return true, false, false
	}
	userStatusCache.Store(userID, cachedUserStatus{active: active, admin: admin, at: time.Now()})
	return active, admin, true
}

// AdminRequired 仅允许管理员访问（须在 AuthRequired 之后使用）
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAdmin, _ := c.Get("isAdmin"); isAdmin != true {
			c.JSON(http.StatusForbidden, gin.H{
				"code":      pkgerrors.ErrPermissionDenied,
				"message":   "仅管理员可操作",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
