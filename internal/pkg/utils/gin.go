package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// ParseUintParam 从URL路径参数解析uint
func ParseUintParam(c *gin.Context, paramName string) (uint, error) {
	paramStr := c.Param(paramName)
	if paramStr == "" {
		return 0, errors.Wrap(errors.ErrMissingParam, "参数不能为空: "+paramName, nil)
	}

	val, err := strconv.ParseUint(paramStr, 10, 32)
	if err != nil {
		return 0, errors.Wrap(errors.ErrInvalidParamType, "参数格式错误: "+paramName, err)
	}

	return uint(val), nil
}

// ParseUintQueryParam 从查询参数解析uint
func ParseUintQueryParam(c *gin.Context, paramName string) (uint, error) {
	paramStr := c.Query(paramName)
	if paramStr == "" {
		return 0, errors.Wrap(errors.ErrMissingParam, "参数不能为空: "+paramName, nil)
	}

	val, err := strconv.ParseUint(paramStr, 10, 32)
	if err != nil {
		return 0, errors.Wrap(errors.ErrInvalidParamType, "参数格式错误: "+paramName, err)
	}

	return uint(val), nil
}

// ParseIntParam 从URL路径参数解析int
func ParseIntParam(c *gin.Context, paramName string) (int, error) {
	paramStr := c.Param(paramName)
	if paramStr == "" {
		return 0, errors.Wrap(errors.ErrMissingParam, "参数不能为空: "+paramName, nil)
	}

	val, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, errors.Wrap(errors.ErrInvalidParamType, "参数格式错误: "+paramName, err)
	}

	return val, nil
}

// ParseIntQueryParam 从查询参数解析int
func ParseIntQueryParam(c *gin.Context, paramName string) (int, error) {
	paramStr := c.Query(paramName)
	if paramStr == "" {
		return 0, errors.Wrap(errors.ErrMissingParam, "参数不能为空: "+paramName, nil)
	}

	val, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, errors.Wrap(errors.ErrInvalidParamType, "参数格式错误: "+paramName, err)
	}

	return val, nil
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.Wrap(errors.ErrUnauthorized, "未授权", nil)
	}

	uid, ok := userID.(uint)
	if !ok {
		return 0, errors.Wrap(errors.ErrInvalidParamType, "用户ID格式错误", nil)
	}

	return uid, nil
}
