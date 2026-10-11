package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// SpeechGlossaryHandler 字幕同传词库智能提取
type SpeechGlossaryHandler struct {
	extractor liveService.GlossaryExtractor
}

// NewSpeechGlossaryHandler 创建处理器
func NewSpeechGlossaryHandler(extractor liveService.GlossaryExtractor) *SpeechGlossaryHandler {
	return &SpeechGlossaryHandler{extractor: extractor}
}

// Extract 整篇原文、译文提取关键词与易错词：POST /live/speech-tasks/glossary/extract
// 文章只用于本次提取，不保存
func (h *SpeechGlossaryHandler) Extract(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.Unauthorized(c, "未获取到公司信息")
		return
	}
	var req liveService.GlossaryExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	res, err := h.extractor.Extract(c.Request.Context(), companyID, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, res)
}
