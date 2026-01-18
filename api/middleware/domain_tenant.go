package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// DomainTenant 域名多租户中间件
// 根据请求的 Host 头自动识别公司并设置上下文
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
			//return
		}

		// 查询数据库，根据域名找到对应的公司
		var company entity.SysCompany
		err := db.Where("DOMAIN = ? AND IS_ACTIVE = ?", host, "Y").First(&company).Error

		if err == nil {
			// 找到公司，设置到上下文
			c.Set("companyID", company.ID)
			c.Set("companyName", company.Name)
			c.Set("companyDomain", host)
		}
		// 如果未找到，不设置（允许系统继续运行，可能使用其他方式识别公司）

		c.Next()
	}
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
