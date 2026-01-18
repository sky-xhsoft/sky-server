package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/crud"
)

// CrudHandler 通用CRUD处理器
type CrudHandler struct {
	crudService crud.Service
}

// NewCrudHandler 创建通用CRUD处理器
func NewCrudHandler(crudService crud.Service) *CrudHandler {
	return &CrudHandler{
		crudService: crudService,
	}
}

// GetOne 查询单条记录
// @Summary 查询单条记录
// @Description 根据ID查询单条记录
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path int true "记录ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/data/{tableName}/{id} [get]
func (h *CrudHandler) GetOne(c *gin.Context) {
	tableName := c.Param("tableName")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.crudService.GetOne(c.Request.Context(), tableName, uint(id), userID.(uint))
	if err != nil {
		utils.InternalError(c, "查询失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// GetList 查询列表
// @Summary 查询列表
// @Description 查询记录列表，支持分页、排序、过滤
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param request body crud.QueryRequest true "查询请求"
// @Success 200 {object} crud.QueryResponse
// @Router /api/v1/data/{tableName} [post]
func (h *CrudHandler) GetList(c *gin.Context) {
	tableName := c.Param("tableName")

	var req crud.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置表名（覆盖请求体中的表名）
	req.TableName = tableName

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.crudService.GetList(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		utils.InternalError(c, "查询失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// Create 创建记录
// @Summary 创建记录
// @Description 创建新记录
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param data body map[string]interface{} true "记录数据"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/data/{tableName} [post]
func (h *CrudHandler) Create(c *gin.Context) {
	tableName := c.Param("tableName")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.crudService.Create(c.Request.Context(), tableName, data, userID.(uint))
	if err != nil {
		utils.InternalError(c, "创建失败: "+err.Error())
		return
	}

	utils.Created(c, result)
}

// Update 更新记录
// @Summary 更新记录
// @Description 更新指定记录
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path int true "记录ID"
// @Param data body map[string]interface{} true "记录数据"
// @Success 200 {object} utils.Response
// @Router /api/v1/data/{tableName}/{id} [put]
func (h *CrudHandler) Update(c *gin.Context) {
	tableName := c.Param("tableName")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.crudService.Update(c.Request.Context(), tableName, uint(id), data, userID.(uint)); err != nil {
		utils.InternalError(c, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "更新成功"})
}

// Delete 删除记录
// @Summary 删除记录
// @Description 删除指定记录（软删除）
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path int true "记录ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/data/{tableName}/{id} [delete]
func (h *CrudHandler) Delete(c *gin.Context) {
	tableName := c.Param("tableName")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.crudService.Delete(c.Request.Context(), tableName, uint(id), userID.(uint)); err != nil {
		utils.InternalError(c, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// BatchDelete 批量删除
// @Summary 批量删除记录
// @Description 批量删除指定记录（软删除）
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param request body BatchDeleteRequest true "批量删除请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/data/{tableName}/batch-delete [post]
func (h *CrudHandler) BatchDelete(c *gin.Context) {
	tableName := c.Param("tableName")

	var req CRUDBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.crudService.BatchDelete(c.Request.Context(), tableName, req.IDs, userID.(uint)); err != nil {
		utils.InternalError(c, "批量删除失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "批量删除成功"})
}

// CRUDBatchDeleteRequest 批量删除请求
type CRUDBatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}
