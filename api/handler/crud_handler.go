package handler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/domain/search"
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
		utils.FailFromError(c, "查询失败: ", err)
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

	// 首先解析为 map 以获取所有字段
	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 构建 QueryRequest
	req := crud.QueryRequest{
		TableName: tableName,
	}

	// 提取标准查询参数
	if page, ok := rawData["page"].(float64); ok {
		req.Page = int(page)
	}
	if pageSize, ok := rawData["pageSize"].(float64); ok {
		req.PageSize = int(pageSize)
	}
	// 支持多字段排序（逗号分隔）
	if orderBy, ok := rawData["orderBy"].(string); ok {
		req.OrderBy = orderBy // 例如: "NAME,CREATE_TIME"
	}
	if order, ok := rawData["order"].(string); ok {
		req.Order = order // 例如: "asc,desc"
	}

	// 提取 filters 字段（如果存在）
	if filters, ok := rawData["filters"].(map[string]interface{}); ok {
		req.Filters = filters
	} else {
		// 如果没有 filters 字段，将所有非标准字段作为过滤条件
		req.Filters = make(map[string]interface{})
		standardFields := map[string]bool{
			"tableName": true,
			"page":      true,
			"pageSize":  true,
			"orderBy":   true,
			"order":     true,
			"filters":   true,
			"include":   true,
		}
		for key, value := range rawData {
			if !standardFields[key] {
				req.Filters[key] = value
			}
		}
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.crudService.GetList(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "查询失败: ", err)
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
		utils.FailFromError(c, "创建失败: ", err)
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
		utils.FailFromError(c, "更新失败: ", err)
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
		utils.FailFromError(c, "删除失败: ", err)
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
		utils.FailFromError(c, "批量删除失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "批量删除成功"})
}

// CRUDBatchDeleteRequest 批量删除请求
type CRUDBatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// SaveWithDetailsRequest 同时保存主表和明细请求
type SaveWithDetailsRequest struct {
	TableName    string                 `json:"tableName" binding:"required"`  // 主表名
	MainRecord   map[string]interface{} `json:"mainRecord" binding:"required"` // 主表数据
	Details      []DetailTableRequest   `json:"details" binding:"required"`    // 子表数据列表
	Mode         string                 `json:"mode"`                          // 保存模式：create/update
	MainRecordID uint                   `json:"mainRecordId"`                  // 更新模式下的主表ID
}

// DetailTableRequest 子表请求数据
type DetailTableRequest struct {
	TableName string                   `json:"tableName" binding:"required"` // 子表名
	Records   []map[string]interface{} `json:"records" binding:"required"`   // 子表数据列表
	AssoType  string                   `json:"assoType" binding:"required"`  // 关联类型：1=1:1, n=1:n
	RefField  string                   `json:"refField" binding:"required"`  // 子表中关联主表ID的字段名
}

// SaveWithDetailsResponse 保存响应
type SaveWithDetailsResponse struct {
	MainRecord map[string]interface{} `json:"mainRecord"` // 保存后的主表数据
	Details    []DetailTableResponse  `json:"details"`    // 保存后的子表数据
}

// DetailTableResponse 子表响应数据
type DetailTableResponse struct {
	TableName string                   `json:"tableName"` // 子表名
	Records   []map[string]interface{} `json:"records"`   // 保存后的子表数据
}

// Submit 提交记录
// @Summary 提交记录
// @Description 提交指定记录，执行提交前后钩子
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path int true "记录ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/data/{tableName}/{id}/submit [post]
func (h *CrudHandler) Submit(c *gin.Context) {
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

	if err := h.crudService.Submit(c.Request.Context(), tableName, uint(id), userID.(uint)); err != nil {
		utils.FailFromError(c, "提交失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "提交成功"})
}

// SaveWithDetails 同时保存主表和明细
// @Summary 同时保存主表和明细
// @Description 同时保存主表记录和所有关联的子表明细记录，使用事务保证一致性
// @Tags CRUD
// @Accept json
// @Produce json
// @Param request body SaveWithDetailsRequest true "保存请求"
// @Success 200 {object} SaveWithDetailsResponse
// @Router /api/v1/data/save-with-details [post]
func (h *CrudHandler) SaveWithDetails(c *gin.Context) {
	var req SaveWithDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists || userID == nil {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 转换为service层的请求结构体
	serviceReq := &crud.SaveWithDetailsRequest{
		TableName:    req.TableName,
		MainRecord:   req.MainRecord,
		Mode:         req.Mode,
		MainRecordID: req.MainRecordID,
	}

	// 转换子表数据
	serviceReq.Details = make([]crud.DetailTableRequest, len(req.Details))
	for i, d := range req.Details {
		serviceReq.Details[i] = crud.DetailTableRequest{
			TableName: d.TableName,
			Records:   d.Records,
			AssoType:  d.AssoType,
			RefField:  d.RefField,
		}
	}

	result, err := h.crudService.SaveWithDetails(c.Request.Context(), serviceReq, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "保存失败: ", err)
		return
	}
	if result == nil {
		utils.Success(c, map[string]interface{}{
			"mainRecord": req.MainRecord,
			"details":    []interface{}{},
		})
		return
	}

	// 转换为handler层的响应结构体
	resp := SaveWithDetailsResponse{
		MainRecord: result.MainRecord,
		Details:    []DetailTableResponse{},
	}
	// 安全地转换子表数据
	if result.Details != nil {
		for _, d := range result.Details {
			resp.Details = append(resp.Details, DetailTableResponse{
				TableName: d.TableName,
				Records:   d.Records,
			})
		}
	}
	for i, d := range result.Details {
		resp.Details[i] = DetailTableResponse{
			TableName: d.TableName,
			Records:   d.Records,
		}
	}

	utils.Success(c, resp)
}

// Unsubmit 反提交记录
// @Summary 反提交记录
// @Description 反提交指定记录，执行反提交前后钩子
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path int true "记录ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/data/{tableName}/{id}/unsubmit [post]
func (h *CrudHandler) Unsubmit(c *gin.Context) {
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

	if err := h.crudService.Unsubmit(c.Request.Context(), tableName, uint(id), userID.(uint)); err != nil {
		utils.FailFromError(c, "反提交失败: ", err)
		return
	}

	utils.Success(c, gin.H{"message": "反提交成功"})
}

// Void 作废记录
func (h *CrudHandler) Void(c *gin.Context) {
	tableName := c.Param("tableName")
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}
	if err := h.crudService.Void(c.Request.Context(), tableName, uint(id), userID.(uint)); err != nil {
		utils.FailFromError(c, "作废失败: ", err)
		return
	}
	utils.Success(c, gin.H{"message": "作废成功"})
}

// Export 导出数据到Excel（支持查询参数过滤和排序）
func (h *CrudHandler) Export(c *gin.Context) {
	tableName := c.Param("tableName")
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var req crud.QueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = crud.QueryRequest{TableName: tableName}
	}
	// gin 的 ShouldBindQuery 无法将 ?filters={...}(JSON) 绑定到 map 字段，
	// 这里手动解析 filters 查询参数, 确保「带筛选条件的导出」只导出匹配行。
	if req.Filters == nil {
		if raw := c.Query("filters"); raw != "" {
			var fm map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &fm); err == nil && len(fm) > 0 {
				req.Filters = fm
			}
		}
	}

	f, err := h.crudService.Export(c.Request.Context(), tableName, req.Filters, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "导出失败: ", err)
		return
	}
	buf, _ := f.WriteToBuffer()
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, tableName))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// Import 从Excel导入数据
func (h *CrudHandler) Import(c *gin.Context) {
	tableName := c.Param("tableName")
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请上传Excel文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		utils.BadRequest(c, "无法读取文件")
		return
	}
	defer f.Close()
	count, err := h.crudService.Import(c.Request.Context(), tableName, f, userID.(uint))
	if err != nil && count == 0 {
		utils.FailFromError(c, "导入失败: ", err)
		return
	}
	if err != nil {
		// 部分成功：返回导入数量 + 失败提示（不走500，前端可直接展示）
		utils.Success(c, gin.H{"imported": count, "message": err.Error()})
		return
	}
	utils.Success(c, gin.H{"imported": count})
}

// ImportTemplate 导出导入模板Excel（供「下载导入模板」使用）
func (h *CrudHandler) ImportTemplate(c *gin.Context) {
	tableName := c.Param("tableName")
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}
	f, err := h.crudService.ImportTemplate(c.Request.Context(), tableName, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "获取模板失败: ", err)
		return
	}
	buf, _ := f.WriteToBuffer()
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_import_template.xlsx"`, tableName))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
func (h *CrudHandler) BatchUpdate(c *gin.Context) {
	tableName := c.Param("tableName")
	var req struct {
		IDs  []uint                 `json:"ids" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var failed int
	for _, id := range req.IDs {
		if err := h.crudService.Update(c.Request.Context(), tableName, id, req.Data, userID.(uint)); err != nil {
			failed++
		}
	}

	if failed > 0 {
		utils.Success(c, gin.H{"message": fmt.Sprintf("修改完成，成功 %d 条，失败 %d 条", len(req.IDs)-failed, failed)})
		return
	}
	utils.Success(c, gin.H{"message": fmt.Sprintf("成功修改 %d 条记录", len(req.IDs))})
}

// Search 高级搜索
// @Summary 高级搜索
// @Description 支持多条件组合搜索，支持 eq/ne/gt/gte/lt/lte/contains/startsWith/endsWith/in/between/isNull/isNotNull 运算符
// @Tags CRUD
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param request body search.SearchRequest true "搜索请求"
// @Success 200 {object} crud.QueryResponse
// @Router /api/v1/data/{tableName}/search [post]
func (h *CrudHandler) Search(c *gin.Context) {
	tableName := c.Param("tableName")

	var req search.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.crudService.Search(c.Request.Context(), &req, tableName, userID.(uint))
	if err != nil {
		utils.FailFromError(c, "搜索失败: ", err)
		return
	}

	utils.Success(c, result)
}
