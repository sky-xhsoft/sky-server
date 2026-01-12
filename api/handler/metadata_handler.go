package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
)

// MetadataHandler 元数据处理器
type MetadataHandler struct {
	metadataService metadata.Service
}

// NewMetadataHandler 创建元数据处理器
func NewMetadataHandler(metadataService metadata.Service) *MetadataHandler {
	return &MetadataHandler{
		metadataService: metadataService,
	}
}

// GetTable 获取表定义
// @Summary 获取表定义
// @Description 根据表名获取表的元数据定义
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Success 200 {object} entity.SysTable
// @Router /api/v1/metadata/tables/{tableName} [get]
func (h *MetadataHandler) GetTable(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		utils.BadRequest(c, "表名不能为空")
		return
	}

	table, err := h.metadataService.GetTable(tableName)
	if err != nil {
		utils.InternalError(c, "获取表定义失败: "+err.Error())
		return
	}

	if table == nil {
		utils.NotFound(c, "表不存在")
		return
	}

	utils.Success(c, table)
}

// GetColumns 获取表的字段定义
// @Summary 获取表的字段定义
// @Description 根据表ID获取表的所有字段定义
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Success 200 {array} entity.SysColumn
// @Router /api/v1/metadata/tables/{tableId}/columns [get]
func (h *MetadataHandler) GetColumns(c *gin.Context) {
	tableIDStr := c.Param("tableName")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "表ID格式错误")
		return
	}

	columns, err := h.metadataService.GetColumns(uint(tableID))
	if err != nil {
		utils.InternalError(c, "获取字段定义失败: "+err.Error())
		return
	}

	utils.Success(c, columns)
}

// GetTableRefs 获取表的关系定义
// @Summary 获取表的关系定义
// @Description 根据表ID获取表的所有关联关系
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Success 200 {array} entity.SysTableRef
// @Router /api/v1/metadata/tables/{tableId}/refs [get]
func (h *MetadataHandler) GetTableRefs(c *gin.Context) {
	tableIDStr := c.Param("tableName")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "表ID格式错误")
		return
	}

	refs, err := h.metadataService.GetTableRefs(uint(tableID))
	if err != nil {
		utils.InternalError(c, "获取表关系失败: "+err.Error())
		return
	}

	utils.Success(c, refs)
}

// GetActions 获取表的动作定义
// @Summary 获取表的动作定义
// @Description 根据表ID获取表的所有可用动作
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Success 200 {array} entity.SysAction
// @Router /api/v1/metadata/tables/{tableId}/actions [get]
func (h *MetadataHandler) GetActions(c *gin.Context) {
	tableIDStr := c.Param("tableName")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "表ID格式错误")
		return
	}

	actions, err := h.metadataService.GetActions(uint(tableID))
	if err != nil {
		utils.InternalError(c, "获取动作定义失败: "+err.Error())
		return
	}

	utils.Success(c, actions)
}

// RefreshCache 刷新元数据缓存
// @Summary 刷新元数据缓存
// @Description 清空并重新加载元数据缓存
// @Tags 元数据
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/metadata/refresh [post]
func (h *MetadataHandler) RefreshCache(c *gin.Context) {
	if err := h.metadataService.RefreshCache(); err != nil {
		utils.InternalError(c, "刷新缓存失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "缓存刷新成功"})
}

// GetMetadataVersion 获取元数据版本
// @Summary 获取元数据版本
// @Description 获取当前元数据的版本号
// @Tags 元数据
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/metadata/version [get]
func (h *MetadataHandler) GetMetadataVersion(c *gin.Context) {
	version := h.metadataService.GetMetadataVersion()
	utils.Success(c, gin.H{"version": version})
}
