package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/sequence"
)

// SequenceHandler 序号处理器
type SequenceHandler struct {
	sequenceService sequence.Service
}

// NewSequenceHandler 创建序号处理器
func NewSequenceHandler(sequenceService sequence.Service) *SequenceHandler {
	return &SequenceHandler{
		sequenceService: sequenceService,
	}
}

// NextValue 获取下一个序号
// @Summary 获取下一个序号
// @Description 根据序号名称获取下一个序号值
// @Tags 序号
// @Accept json
// @Produce json
// @Param seqName path string true "序号名称"
// @Success 200 {object} utils.Response
// @Router /api/v1/sequences/{seqName}/next [post]
func (h *SequenceHandler) NextValue(c *gin.Context) {
	seqName := c.Param("seqName")
	if seqName == "" {
		utils.BadRequest(c, "序号名称不能为空")
		return
	}

	value, err := h.sequenceService.NextValue(seqName)
	if err != nil {
		utils.InternalError(c, "生成序号失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"value": value})
}

// BatchNextValue 批量获取序号
// @Summary 批量获取序号
// @Description 批量获取指定数量的序号
// @Tags 序号
// @Accept json
// @Produce json
// @Param request body BatchNextValueRequest true "批量获取请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/sequences/batch [post]
func (h *SequenceHandler) BatchNextValue(c *gin.Context) {
	var req BatchNextValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Count <= 0 || req.Count > 100 {
		utils.BadRequest(c, "数量必须在1-100之间")
		return
	}

	values := make([]string, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		value, err := h.sequenceService.NextValue(req.SeqName)
		if err != nil {
			utils.InternalError(c, "生成序号失败: "+err.Error())
			return
		}
		values = append(values, value)
	}

	utils.Success(c, gin.H{"values": values})
}

// GetCurrentValue 获取当前序号值
// @Summary 获取当前序号值
// @Description 获取序号的当前值（不递增）
// @Tags 序号
// @Accept json
// @Produce json
// @Param seqName path string true "序号名称"
// @Success 200 {object} utils.Response
// @Router /api/v1/sequences/{seqName}/current [get]
func (h *SequenceHandler) GetCurrentValue(c *gin.Context) {
	seqName := c.Param("seqName")
	if seqName == "" {
		utils.BadRequest(c, "序号名称不能为空")
		return
	}

	value, err := h.sequenceService.GetCurrentValue(seqName)
	if err != nil {
		utils.InternalError(c, "获取当前序号失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"value": value})
}

// ResetSequence 重置序号
// @Summary 重置序号
// @Description 重置序号到初始值
// @Tags 序号
// @Accept json
// @Produce json
// @Param seqName path string true "序号名称"
// @Success 200 {object} utils.Response
// @Router /api/v1/sequences/{seqName}/reset [post]
func (h *SequenceHandler) ResetSequence(c *gin.Context) {
	seqName := c.Param("seqName")
	if seqName == "" {
		utils.BadRequest(c, "序号名称不能为空")
		return
	}

	if err := h.sequenceService.ResetSequence(seqName); err != nil {
		utils.InternalError(c, "重置序号失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "序号重置成功"})
}

// BatchNextValueRequest 批量获取序号请求
type BatchNextValueRequest struct {
	SeqName string `json:"seqName" binding:"required"`
	Count   int    `json:"count" binding:"required"`
}
