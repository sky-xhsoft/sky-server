package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
)

// MetadataHandler 元数据处理器
type MetadataHandler struct {
	metadataService metadata.Service
	dictRepo        repository.DictRepository
}

// NewMetadataHandler 创建元数据处理器
func NewMetadataHandler(metadataService metadata.Service, dictRepo repository.DictRepository) *MetadataHandler {
	return &MetadataHandler{
		metadataService: metadataService,
		dictRepo:        dictRepo,
	}
}

// GetTableConfig 获取表的完整配置（表信息+字段列表）
// @Summary 获取表的完整配置
// @Description 根据表ID获取表的元数据定义和所有字段定义
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/metadata/tables/{tableId}/config [get]
func (h *MetadataHandler) GetTableConfig(c *gin.Context) {
	tableIDStr := c.Param("tableId")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "表ID格式错误")
		return
	}

	// 获取表信息
	table, err := h.metadataService.GetTableByID(uint(tableID))
	if err != nil {
		utils.InternalError(c, "获取表定义失败: "+err.Error())
		return
	}

	if table == nil {
		utils.NotFound(c, "表不存在")
		return
	}

	// 获取字段列表
	columns, err := h.metadataService.GetColumns(uint(tableID))
	if err != nil {
		utils.InternalError(c, "获取字段定义失败: "+err.Error())
		return
	}

	// 收集所有需要的字典ID
	dictIDs := make(map[uint]bool)
	for _, col := range columns {
		if col.SetValueType == "select" && col.SysDictID != "" {
			// SysDictID 是字符串，需要转换为uint
			dictID, err := strconv.ParseUint(col.SysDictID, 10, 32)
			if err != nil {
				fmt.Printf("[WARN] 字段 %s 的字典ID格式错误: %s\n", col.DbName, col.SysDictID)
				continue
			}
			dictIDs[uint(dictID)] = true
		}
	}

	// 获取字典数据
	dictData := make(map[uint][]*entity.SysDictItem)
	for dictID := range dictIDs {
		items, err := h.dictRepo.GetDictItems(dictID)
		if err != nil {
			// 字典获取失败不影响整体返回，只记录日志
			fmt.Printf("[WARN] 获取字典 %d 失败: %v\n", dictID, err)
			continue
		}
		dictData[dictID] = items
	}

	// 返回完整配置
	utils.Success(c, gin.H{
		"table":    table,
		"columns":  columns,
		"dictData": dictData,
	})
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
	tableIDStr := c.Param("tableId")
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
	tableIDStr := c.Param("tableId")
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
	tableIDStr := c.Param("tableId")
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

// GetForeignKeyOptions 获取外键选项列表
// @Summary 获取外键选项列表
// @Description 根据表ID获取外键字段的可选项列表，支持搜索和分页
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableId query int true "关联表ID"
// @Param columnId query int false "显示字段ID（可选，默认使用表的DK字段）"
// @Param search query string false "搜索关键字"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页条数" default(100)
// @Success 200 {object} utils.Response
// @Router /api/v1/metadata/foreign-key-options [get]
func (h *MetadataHandler) GetForeignKeyOptions(c *gin.Context) {
	tableIDStr := c.Query("tableId")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "表ID格式错误")
		return
	}

	columnIDStr := c.Query("columnId")
	var columnID *uint
	if columnIDStr != "" {
		cid, err := strconv.ParseUint(columnIDStr, 10, 32)
		if err == nil {
			cidUint := uint(cid)
			columnID = &cidUint
		}
	}

	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))

	// Extract userID from context for company-based filtering
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	options, total, isDropdown, err := h.metadataService.GetForeignKeyOptions(uint(tableID), columnID, search, page, pageSize, userID.(uint))
	if err != nil {
		utils.InternalError(c, "获取外键选项失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":       options,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"isDropdown": isDropdown,
	})
}

// GetForeignKeyDisplayValue 获取外键显示值
// @Summary 获取外键显示值
// @Description 根据外键值获取对应的显示文本
// @Tags 元数据
// @Accept json
// @Produce json
// @Param tableId query int true "关联表ID"
// @Param value query string true "外键值"
// @Param columnId query int false "显示字段ID（可选）"
// @Success 200 {object} utils.Response
// @Router /api/v1/metadata/foreign-key-display-value [get]
func (h *MetadataHandler) GetForeignKeyDisplayValue(c *gin.Context) {
	tableIDStr := c.Query("tableId")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "表ID格式错误")
		return
	}

	value := c.Query("value")
	if value == "" {
		utils.BadRequest(c, "值不能为空")
		return
	}

	columnIDStr := c.Query("columnId")
	var columnID *uint
	if columnIDStr != "" {
		cid, err := strconv.ParseUint(columnIDStr, 10, 32)
		if err == nil {
			cidUint := uint(cid)
			columnID = &cidUint
		}
	}

	// Extract userID from context for company-based filtering
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	displayValue, err := h.metadataService.GetForeignKeyDisplayValue(uint(tableID), value, columnID, userID.(uint))
	if err != nil {
		utils.InternalError(c, "获取显示值失败: "+err.Error())
		return
	}

	utils.Success(c, displayValue)
}
