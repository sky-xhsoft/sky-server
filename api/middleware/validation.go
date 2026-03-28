package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/validator"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
)

// Validate 结构体参数验证中间件
func Validate(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 根据Content-Type选择绑定方式
		if err := c.ShouldBind(obj); err != nil {
			utils.BadRequest(c, "请求参数错误: "+err.Error())
			c.Abort()
			return
		}

		// 使用自定义验证器进行验证
		if err := validator.ValidateStruct(obj); err != nil {
			if valErr, ok := err.(*validator.ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "参数验证失败",
					"errors":  valErr.Errors,
				})
				c.Abort()
				return
			}
			utils.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		// 将验证后的对象存入上下文
		c.Set("validated", obj)
		c.Next()
	}
}

// ValidateJSON JSON参数验证中间件
func ValidateJSON(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定JSON数据
		if err := c.ShouldBindJSON(obj); err != nil {
			utils.BadRequest(c, "请求参数错误: "+err.Error())
			c.Abort()
			return
		}

		// 使用自定义验证器进行验证
		if err := validator.ValidateStruct(obj); err != nil {
			if valErr, ok := err.(*validator.ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "参数验证失败",
					"errors":  valErr.Errors,
				})
				c.Abort()
				return
			}
			utils.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		// 将验证后的对象存入上下文
		c.Set("validated", obj)
		c.Next()
	}
}

// ValidateQuery Query参数验证中间件
func ValidateQuery(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定Query数据
		if err := c.ShouldBindQuery(obj); err != nil {
			utils.BadRequest(c, "请求参数错误: "+err.Error())
			c.Abort()
			return
		}

		// 使用自定义验证器进行验证
		if err := validator.ValidateStruct(obj); err != nil {
			if valErr, ok := err.(*validator.ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "参数验证失败",
					"errors":  valErr.Errors,
				})
				c.Abort()
				return
			}
			utils.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		// 将验证后的对象存入上下文
		c.Set("validated", obj)
		c.Next()
	}
}

// ValidateURI URI参数验证中间件
func ValidateURI(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定URI数据
		if err := c.ShouldBindUri(obj); err != nil {
			utils.BadRequest(c, "请求参数错误: "+err.Error())
			c.Abort()
			return
		}

		// 使用自定义验证器进行验证
		if err := validator.ValidateStruct(obj); err != nil {
			if valErr, ok := err.(*validator.ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "参数验证失败",
					"errors":  valErr.Errors,
				})
				c.Abort()
				return
			}
			utils.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		// 将验证后的对象存入上下文
		c.Set("validated", obj)
		c.Next()
	}
}

// GetValidated 从上下文中获取验证后的对象
func GetValidated(c *gin.Context) interface{} {
	if val, exists := c.Get("validated"); exists {
		return val
	}
	return nil
}
