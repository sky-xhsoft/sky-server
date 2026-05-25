package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Compression 响应压缩中间件
// 对大于 1KB 的响应自动进行 gzip 压缩，减少 API 响应体积
func Compression() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico",
		".mp4", ".avi", ".mov", ".mkv",
		".mp3", ".wav", ".ogg",
		".zip", ".gz", ".tar", ".rar", ".7z",
	}))
}