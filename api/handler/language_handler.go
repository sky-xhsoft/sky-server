package handler

import (
	"io"
	"strconv"

	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/language"

	"github.com/gin-gonic/gin"
)

// LanguageHandler 语言管理处理器
type LanguageHandler struct {
	langService language.Service
}

// NewLanguageHandler 创建语言管理处理器
func NewLanguageHandler(langService language.Service) *LanguageHandler {
	return &LanguageHandler{
		langService: langService,
	}
}

// ListLanguages 获取语言列表
// @Summary 获取语言列表
// @Description 获取系统支持的语言列表
// @Tags 语言管理
// @Accept json
// @Produce json
// @Success 200 {array} entity.SysLanguage
// @Router /api/v1/languages [get]
func (h *LanguageHandler) ListLanguages(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	languages, err := h.langService.ListLanguages(companyIDUint)
	if err != nil {
		utils.InternalError(c, "获取语言列表失败: "+err.Error())
		return
	}

	utils.Success(c, languages)
}

// ListTranslations 获取翻译列表
// @Summary 获取翻译列表
// @Description 获取翻译列表，支持分页和筛选
// @Tags 语言管理
// @Accept json
// @Produce json
// @Param language query string false "语言代码"
// @Param module query string false "模块名称"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} utils.Response
// @Router /api/v1/languages/translations [get]
func (h *LanguageHandler) ListTranslations(c *gin.Context) {
	var req language.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	companyID, _ := c.Get("companyID")
	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	translations, total, err := h.langService.ListTranslations(companyIDUint, &req)
	if err != nil {
		utils.InternalError(c, "获取翻译列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"items":    translations,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

// GetTranslation 获取翻译
// @Summary 获取翻译
// @Description 根据key获取翻译
// @Tags 语言管理
// @Accept json
// @Produce json
// @Param key path string true "翻译键"
// @Param language query string false "语言代码，默认zh-CN"
// @Success 200 {object} entity.SysTranslation
// @Router /api/v1/languages/translations/{key} [get]
func (h *LanguageHandler) GetTranslation(c *gin.Context) {
	key := c.Param("key")
	language := c.Query("language")
	if language == "" {
		language = "zh-CN"
	}

	companyID, _ := c.Get("companyID")
	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	translation, err := h.langService.GetTranslation(companyIDUint, key, language)
	if err != nil {
		utils.InternalError(c, "获取翻译失败: "+err.Error())
		return
	}

	utils.Success(c, translation)
}

// AddTranslation 添加翻译
// @Summary 添加翻译
// @Description 添加新的翻译
// @Tags 语言管理
// @Accept json
// @Produce json
// @Param request body language.AddRequest true "添加翻译请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/languages/translations [post]
func (h *LanguageHandler) AddTranslation(c *gin.Context) {
	var req language.AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	companyID, _ := c.Get("companyID")

	var userIDUint uint
	if id, ok := userID.(uint); ok {
		userIDUint = id
	}

	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	err := h.langService.AddTranslation(&req, userIDUint, username.(string), companyIDUint)
	if err != nil {
		utils.InternalError(c, "添加翻译失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "添加成功"})
}

// UpdateTranslation 更新翻译
// @Summary 更新翻译
// @Description 更新翻译内容
// @Tags 语言管理
// @Accept json
// @Produce json
// @Param id path int true "翻译ID"
// @Param request body language.UpdateRequest true "更新翻译请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/languages/translations/{id} [put]
func (h *LanguageHandler) UpdateTranslation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req language.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	req.ID = uint(id)

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	companyID, _ := c.Get("companyID")

	var userIDUint uint
	if id, ok := userID.(uint); ok {
		userIDUint = id
	}

	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	err = h.langService.UpdateTranslation(&req, userIDUint, username.(string), companyIDUint)
	if err != nil {
		utils.InternalError(c, "更新翻译失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "更新成功"})
}

// DeleteTranslation 删除翻译
// @Summary 删除翻译
// @Description 删除指定翻译
// @Tags 语言管理
// @Accept json
// @Produce json
// @Param id path int true "翻译ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/languages/translations/{id} [delete]
func (h *LanguageHandler) DeleteTranslation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	err = h.langService.DeleteTranslation(uint(id))
	if err != nil {
		utils.InternalError(c, "删除翻译失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// ExportTranslations 导出翻译
// @Summary 导出翻译
// @Description 导出指定语言的翻译为JSON格式
// @Tags 语言管理
// @Accept json
// @Produce octet-stream
// @Param language query string true "语言代码"
// @Success 200 {file} file
// @Router /api/v1/languages/translations/export [get]
func (h *LanguageHandler) ExportTranslations(c *gin.Context) {
	language := c.Query("language")
	if language == "" {
		utils.BadRequest(c, "请指定语言")
		return
	}

	companyID, _ := c.Get("companyID")
	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	data, err := h.langService.ExportTranslations(companyIDUint, language)
	if err != nil {
		utils.InternalError(c, "导出翻译失败: "+err.Error())
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=translations_"+language+".json")
	c.String(200, string(data))
}

// ImportTranslations 导入翻译
// @Summary 导入翻译
// @Description 导入翻译，支持JSON格式
// @Tags 语言管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "JSON翻译文件"
// @Param language formData string true "语言代码"
// @Success 200 {object} utils.Response
// @Router /api/v1/languages/translations/import [post]
func (h *LanguageHandler) ImportTranslations(c *gin.Context) {
	language := c.PostForm("language")
	if language == "" {
		utils.BadRequest(c, "请指定语言")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请上传文件")
		return
	}

	f, err := file.Open()
	if err != nil {
		utils.BadRequest(c, "打开文件失败: "+err.Error())
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		utils.BadRequest(c, "读取文件失败: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	companyID, _ := c.Get("companyID")

	var userIDUint uint
	if id, ok := userID.(uint); ok {
		userIDUint = id
	}

	var companyIDUint uint
	if id, ok := companyID.(uint); ok {
		companyIDUint = id
	}

	err = h.langService.ImportTranslations(companyIDUint, data, language, userIDUint, username.(string))
	if err != nil {
		utils.InternalError(c, "导入翻译失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "导入成功"})
}
