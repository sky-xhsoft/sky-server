package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// DirectoryPermissionRequired 安全目录权限检查中间件
// directoryID: 目录ID
// requiredPerm: 需要的权限(位运算值)
func DirectoryPermissionRequired(groupService groups.Service, directoryID uint, requiredPerm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := groupService.CheckUserPermission(
			c.Request.Context(),
			userID.(uint),
			directoryID,
			requiredPerm,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    errors.ErrInternal,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    errors.ErrForbidden,
				"message": "无权限访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TablePermissionRequired 表权限检查中间件
// tableID: 表ID
// requiredPerm: 需要的权限(位运算值)
func TablePermissionRequired(groupService groups.Service, tableID uint, requiredPerm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// 检查表权限
		hasPermission, err := groupService.CheckUserTablePermission(
			c.Request.Context(),
			userID.(uint),
			tableID,
			requiredPerm,
		)

		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    errors.ErrForbidden,
				"message": "无权限访问该表",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DynamicTablePermissionRequired 动态表权限检查中间件(从路径参数获取表名)
// tableNameParam: 表名参数的key(如"tableName")
// requiredPerm: 需要的权限(位运算值)
func DynamicTablePermissionRequired(groupService groups.Service, tableNameParam string, requiredPerm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// 获取表名
		tableName := c.Param(tableNameParam)
		if tableName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    errors.ErrInvalidParam,
				"message": "表名参数缺失",
			})
			c.Abort()
			return
		}

		// TODO: 需要一个从表名查询表ID的方法
		// 这里暂时跳过，实际使用时需要实现
		// 或者改为直接使用表名而不是表ID

		c.Next()
	}
}

// GetUserPermission 获取用户权限(不阻止请求)
// 将用户在指定目录的权限值设置到context中
func GetUserPermission(groupService groups.Service, directoryID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Set("userPermission", 0)
			c.Set("userPermissionBits", map[string]bool{
				"read":   false,
				"create": false,
				"update": false,
				"delete": false,
				"export": false,
				"import": false,
			})
			c.Next()
			return
		}

		permission, err := groupService.GetUserDirectoryPermission(
			c.Request.Context(),
			userID.(uint),
			directoryID,
		)

		if err != nil {
			permission = 0
		}

		// 设置权限值
		c.Set("userPermission", permission)

		// 设置权限位映射(方便使用)
		c.Set("userPermissionBits", map[string]bool{
			"read":   (permission & groups.PermRead) > 0,
			"create": (permission & groups.PermCreate) > 0,
			"update": (permission & groups.PermUpdate) > 0,
			"delete": (permission & groups.PermDelete) > 0,
			"export": (permission & groups.PermExport) > 0,
			"import": (permission & groups.PermImport) > 0,
		})

		c.Next()
	}
}

// CheckPermissionBit 检查用户是否有指定的权限位
func CheckPermissionBit(c *gin.Context, permBit string) bool {
	permBits, exists := c.Get("userPermissionBits")
	if !exists {
		return false
	}

	bits, ok := permBits.(map[string]bool)
	if !ok {
		return false
	}

	return bits[permBit]
}

// GetDataFilter 获取用户数据过滤条件中间件
// 将用户的数据过滤条件设置到context中
func GetDataFilter(groupService groups.Service, directoryID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.Set("dataFilter", nil)
			c.Next()
			return
		}

		filter, err := groupService.GetUserDataFilter(
			c.Request.Context(),
			userIDVal.(uint),
			directoryID,
		)

		if err != nil {
			filter = nil
		}

		c.Set("dataFilter", filter)
		c.Next()
	}
}

// DirectoryIDFromParam 从路径参数中提取目录ID
func DirectoryIDFromParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.New(errors.ErrInvalidParam, "目录ID参数缺失")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, errors.New(errors.ErrInvalidParam, "目录ID格式错误")
	}

	return uint(id), nil
}
