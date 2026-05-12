package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func (i *ipLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	lim, exists := i.limiters[ip]
	i.mu.RUnlock()
	if exists {
		return lim
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	lim, exists = i.limiters[ip]
	if !exists {
		lim = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = lim
	}
	return lim
}

func (i *ipLimiter) cleanupTick(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for range ticker.C {
		i.mu.Lock()
		i.limiters = make(map[string]*rate.Limiter)
		i.mu.Unlock()
	}
}

func NewIPRateLimiter(r rate.Limit, b int) *ipLimiter {
	il := &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
	go il.cleanupTick(10 * time.Minute)
	return il
}

func (i *ipLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		lim := i.getLimiter(ip)
		if !lim.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
