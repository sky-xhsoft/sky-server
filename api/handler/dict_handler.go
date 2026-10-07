package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/dict"
)

// DictHandler 字典处理器
type DictHandler struct {
	dictService dict.Service
}

// NewDictHandler 创建字典处理器
func NewDictHandler(dictService dict.Service) *DictHandler {
	return &DictHandler{
		dictService: dictService,
	}
}

// GetDictItems 根据字典ID获取字典项
// @Summary 获取字典项
// @Description 根据字典ID获取所有字典项
// @Tags 字典
// @Accept json
// @Produce json
// @Param dictId path int true "字典ID"
// @Success 200 {array} entity.SysDictItem
// @Router /api/v1/dicts/{dictId}/items [get]
func (h *DictHandler) GetDictItems(c *gin.Context) {
	dictIDStr := c.Param("dictId")
	dictID, err := strconv.ParseUint(dictIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "字典ID格式错误")
		return
	}

	items, err := h.dictService.GetDictItems(uint(dictID))
	if err != nil {
		utils.FailFromError(c, "获取字典项失败: ", err)
		return
	}

	utils.Success(c, items)
}

// GetDictItemsByName 根据字典名称获取字典项
// @Summary 根据字典名称获取字典项
// @Description 根据字典名称获取所有字典项
// @Tags 字典
// @Accept json
// @Produce json
// @Param dictName path string true "字典名称"
// @Success 200 {array} entity.SysDictItem
// @Router /api/v1/dicts/name/{dictName}/items [get]
func (h *DictHandler) GetDictItemsByName(c *gin.Context) {
	dictName := c.Param("dictName")
	if dictName == "" {
		utils.BadRequest(c, "字典名称不能为空")
		return
	}

	items, err := h.dictService.GetDictItemsByName(dictName)
	if err != nil {
		utils.FailFromError(c, "获取字典项失败: ", err)
		return
	}

	utils.Success(c, items)
}

// GetDefaultValue 获取字典的默认值
// @Summary 获取字典的默认值
// @Description 根据字典名称获取默认值
// @Tags 字典
// @Accept json
// @Produce json
// @Param dictName path string true "字典名称"
// @Success 200 {object} utils.Response
// @Router /api/v1/dicts/{dictName}/default [get]
func (h *DictHandler) GetDefaultValue(c *gin.Context) {
	// 路由为 /dicts/:dictId/default（原先读取 dictName 参数，永远为空）；数字按字典ID查，否则按字典名称查
	param := c.Param("dictId")
	if param == "" {
		param = c.Param("dictName")
	}
	if param == "" {
		utils.BadRequest(c, "字典ID不能为空")
		return
	}

	if id, err := strconv.ParseUint(param, 10, 32); err == nil {
		items, err := h.dictService.GetDictItems(uint(id))
		if err != nil {
			utils.FailFromError(c, "获取默认值失败: ", err)
			return
		}
		defaultValue := ""
		for _, item := range items {
			if item.IsDefaultValue == "Y" {
				defaultValue = item.Value
				break
			}
		}
		if defaultValue == "" && len(items) > 0 {
			defaultValue = items[0].Value
		}
		utils.Success(c, gin.H{"defaultValue": defaultValue})
		return
	}

	defaultValue, err := h.dictService.GetDefaultValue(param)
	if err != nil {
		utils.FailFromError(c, "获取默认值失败: ", err)
		return
	}

	utils.Success(c, gin.H{"defaultValue": defaultValue})
}

// RefreshCache 刷新字典缓存
// @Summary 刷新字典缓存
// @Description 清空并重新加载所有字典缓存
// @Tags 字典
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/dicts/refresh [post]
func (h *DictHandler) RefreshCache(c *gin.Context) {
	if err := h.dictService.RefreshDictCache(); err != nil {
		utils.FailFromError(c, "刷新缓存失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "缓存刷新成功"})
}
