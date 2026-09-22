package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// DomainTenant 域名多租户中间件
// 根据请求的 Host 头自动识别公司并设置上下文
// 支持两种域名识别方式：
//   1. {companyId}.zhibo.skyzhou.cn  —— 子域第一段为数字公司ID，如 1.zhibo.skyzhou.cn → 1、46.zhibo.skyzhou.cn → 46
//   2. DB 中 sys_company.DOMAIN 精确匹配
// 若都未识别，默认公司 1
func DomainTenant(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求的 Host（包含域名和端口）
		host := c.Request.Host

		// 移除端口号（如果有）
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// 如果 host 为空或为 localhost，不进行域名识别
		if host == "" || host == "localhost" || strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "192.168.") {
			c.Set("companyID", 1)
			c.Next()
			return
		}

		// 优先：解析 {companyId}.zhibo.skyzhou.cn 子域
		if companyID, ok := parseSubdomainCompanyID(host); ok {
			c.Set("companyID", companyID)
			c.Set("companyDomain", host)
			c.Next()
			return
		}

		// 其次：查询数据库，根据域名找到对应的公司
		var company entity.SysCompany
		err := db.Where("DOMAIN = ? AND IS_ACTIVE = ?", host, "Y").First(&company).Error

		if err == nil {
			// 找到公司，设置到上下文
			c.Set("companyID", company.ID)
			c.Set("companyName", company.Name)
			c.Set("companyDomain", host)
		} else {
			// 如果未找到，默认使用公司ID=1（默认公司）
			c.Set("companyID", uint(1))
		}

		c.Next()
	}
}

// subdomainBase 域名基线：{companyId} 等子域都挂在这个域下
const subdomainBase = ".zhibo.skyzhou.cn"

// parseSubdomainCompanyID 尝试解析 {companyId} + subdomainBase 形式的子域。
// 例：1.zhibo.skyzhou.cn → (1, true)；46.zhibo.skyzhou.cn → (46, true)；其它 → (0, false)
func parseSubdomainCompanyID(host string) (uint, bool) {
	if !strings.HasSuffix(host, subdomainBase) {
		return 0, false
	}
	firstPart := strings.TrimSuffix(host, subdomainBase)
	// 仅允许形如 "46" 的纯数字第一段；若带额外子域（如 a.1.zhibo.skyzhou.cn）则不支持
	if firstPart == "" || strings.Contains(firstPart, ".") {
		return 0, false
	}
	var id uint
	n, err := fmt.Sscanf(firstPart, "%d", &id)
	if err != nil || n != 1 || id == 0 {
		return 0, false
	}
	return id, true
}

// GetCompanyID 从上下文中获取公司 ID
func GetCompanyID(c *gin.Context) *uint {
	if companyID, exists := c.Get("companyID"); exists {
		if id, ok := companyID.(uint); ok {
			return &id
		}
	}
	return nil
}

// GetCompanyName 从上下文中获取公司名称
func GetCompanyName(c *gin.Context) string {
	if companyName, exists := c.Get("companyName"); exists {
		if name, ok := companyName.(string); ok {
			return name
		}
	}
	return ""
}

// RequireCompany 要求必须识别到公司的中间件
func RequireCompany() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("companyID"); !exists {
			c.JSON(403, gin.H{
				"code":    403,
				"message": "无法识别公司域名，请使用正确的域名访问",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
