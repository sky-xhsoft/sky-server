package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/pref"
)

// PrefHandler 用户环境变量处理器
type PrefHandler struct {
	prefService pref.Service
}

// NewPrefHandler 创建用户环境变量处理器
func NewPrefHandler(prefService pref.Service) *PrefHandler {
	return &PrefHandler{prefService: prefService}
}

// getUID 从上下文取当前用户ID
func (h *PrefHandler) getUID(c *gin.Context) uint {
	if v, ok := c.Get("userID"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// Get 读取单个环境变量 GET /user-prefs?module=&name=&default=
func (h *PrefHandler) Get(c *gin.Context) {
	uid := h.getUID(c)
	module := c.Query("module")
	name := c.Query("name")
	def := c.Query("default")
	val, err := h.prefService.Get(c.Request.Context(), uid, module, name, def)
	if err != nil {
		utils.InternalError(c, "查询环境变量失败: "+err.Error())
		return
	}
	if val == "" {
		val = def
	}
	utils.Success(c, map[string]string{"value": val})
}

// GetAll 读取某模块全部 GET /user-prefs/all?module=
func (h *PrefHandler) GetAll(c *gin.Context) {
	uid := h.getUID(c)
	module := c.Query("module")
	m, err := h.prefService.GetAll(c.Request.Context(), uid, module)
	if err != nil {
		utils.InternalError(c, "查询失败: "+err.Error())
		return
	}
	utils.Success(c, m)
}

// Set 写入环境变量 PUT /user-prefs (body: {module,name,value})
func (h *PrefHandler) Set(c *gin.Context) {
	uid := h.getUID(c)
	var req struct {
		Module string `json:"module"`
		Name   string `json:"name"`
		Value  string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.Name == "" {
		utils.BadRequest(c, "name 不能为空")
		return
	}
	if err := h.prefService.Set(c.Request.Context(), uid, req.Module, req.Name, req.Value); err != nil {
		utils.InternalError(c, "写入失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// Delete 删除单个 DELETE /user-prefs?module=&name=
func (h *PrefHandler) Delete(c *gin.Context) {
	uid := h.getUID(c)
	module := c.Query("module")
	name := c.Query("name")
	if err := h.prefService.Delete(c.Request.Context(), uid, module, name); err != nil {
		utils.InternalError(c, "删除失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// List 分页列出 GET /user-prefs/list?page=&pageSize=
func (h *PrefHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	uid := h.getUID(c)
	rows, total, err := h.prefService.List(c.Request.Context(), uid, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"data": rows, "total": total})
}