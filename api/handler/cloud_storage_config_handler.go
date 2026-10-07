package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"go.uber.org/zap"
)

// CloudStorageConfigHandler 云存储配置处理器
//
// 实际使用的存储由公司配置（sys_company_conf.STORAGE_TYPE 等，“公司配置”页面维护）决定；
// cloud_storage_config 表不参与存储选择，原先针对它的增删改查接口已下线，避免两套配置互相误导。
// 这里只保留刷新：修改公司存储配置后让新配置立即生效。
type CloudStorageConfigHandler struct {
	storageConfigService cloud.StorageConfigService
}

// NewCloudStorageConfigHandler 创建云存储配置处理器
func NewCloudStorageConfigHandler(storageConfigService cloud.StorageConfigService) *CloudStorageConfigHandler {
	return &CloudStorageConfigHandler{
		storageConfigService: storageConfigService,
	}
}

// RefreshCache 刷新存储配置缓存
// @Summary 刷新存储配置缓存
// @Description 刷新指定公司的存储配置缓存
// @Tags Cloud-Storage-Config
// @Accept json
// @Produce json
// @Param companyId path int true "公司ID"
// @Success 200 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/cloud/storage/config/:companyId/refresh [post]
func (h *CloudStorageConfigHandler) RefreshCache(c *gin.Context) {
	companyIDStr := c.Param("companyId")
	companyID, err := strconv.ParseUint(companyIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "公司ID格式错误")
		return
	}

	if uint(companyID) != c.GetUint("companyID") {
		utils.Forbidden(c, "只能刷新本公司的存储配置")
		return
	}
	logger.Info("刷新存储配置缓存请求", zap.Uint("companyId", uint(companyID)))

	if err := h.storageConfigService.RefreshCache(c.Request.Context(), uint(companyID)); err != nil {
		logger.Error("刷新存储配置缓存失败",
			zap.Uint("companyId", uint(companyID)),
			zap.Error(err))
		utils.FailFromError(c, "刷新存储配置缓存失败: ", err)
		return
	}

	utils.Success(c, gin.H{
		"message": "存储配置缓存刷新成功",
	})
}
