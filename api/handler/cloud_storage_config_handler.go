package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"go.uber.org/zap"
)

// CloudStorageConfigHandler 云存储配置处理器
type CloudStorageConfigHandler struct {
	storageConfigService cloud.StorageConfigService
}

// NewCloudStorageConfigHandler 创建云存储配置处理器
func NewCloudStorageConfigHandler(storageConfigService cloud.StorageConfigService) *CloudStorageConfigHandler {
	return &CloudStorageConfigHandler{
		storageConfigService: storageConfigService,
	}
}

// CreateConfigRequest 创建存储配置请求
type CreateConfigRequest struct {
	SysCompanyID uint   `json:"sysCompanyId" binding:"required"`                                 // 公司ID
	StorageType  string `json:"storageType" binding:"required,oneof=local aliyunOSS tencentCOS"` // 存储类型
	// 本地存储配置
	LocalBasePath string `json:"localBasePath" binding:"required_if=StorageType local"` // 本地存储基础路径
	LocalBaseURL  string `json:"localBaseUrl" binding:"required_if=StorageType local"`  // 本地存储基础URL
	// 阿里云OSS配置
	AliyunOSSEndpoint        string `json:"aliyunOSSEndpoint" binding:"required_if=StorageType aliyunOSS"`        // OSS端点
	AliyunOSSAccessKeyID     string `json:"aliyunOSSAccessKeyId" binding:"required_if=StorageType aliyunOSS"`     // AccessKey ID
	AliyunOSSAccessKeySecret string `json:"aliyunOSSAccessKeySecret" binding:"required_if=StorageType aliyunOSS"` // AccessKey Secret
	AliyunOSSBucketName      string `json:"aliyunOSSBucketName" binding:"required_if=StorageType aliyunOSS"`      // 存储桶名称
	AliyunOSSCDNDomain       string `json:"aliyunOSSCDNDomain"`                                                   // CDN域名
	// 腾讯云COS配置
	TencentCOSBucketURL  string `json:"tencentCOSBucketUrl" binding:"required_if=StorageType tencentCOS"`  // 存储桶URL
	TencentCOSSecretID   string `json:"tencentCOSSecretId" binding:"required_if=StorageType tencentCOS"`   // SecretID
	TencentCOSSecretKey  string `json:"tencentCOSSecretKey" binding:"required_if=StorageType tencentCOS"`  // SecretKey
	TencentCOSBucketName string `json:"tencentCOSBucketName" binding:"required_if=StorageType tencentCOS"` // 存储桶名称
	TencentCOSRegion     string `json:"tencentCOSRegion" binding:"required_if=StorageType tencentCOS"`     // 区域
	TencentCOSCDNDomain  string `json:"tencentCOSCDNDomain"`                                               // CDN域名
}

// UpdateConfigRequest 更新存储配置请求（继承自CreateConfigRequest，允许部分字段可选）
type UpdateConfigRequest struct {
	CreateConfigRequest
	ID uint `json:"id" binding:"required"` // 配置ID
}

// ConfigResponse 存储配置响应
type ConfigResponse struct {
	ID           uint   `json:"id"`           // 配置ID
	SysCompanyID uint   `json:"sysCompanyId"` // 公司ID
	StorageType  string `json:"storageType"`  // 存储类型
	// 本地存储配置
	LocalBasePath string `json:"localBasePath"` // 本地存储基础路径
	LocalBaseURL  string `json:"localBaseUrl"`  // 本地存储基础URL
	// 阿里云OSS配置
	AliyunOSSEndpoint        string `json:"aliyunOSSEndpoint"`        // OSS端点
	AliyunOSSAccessKeyID     string `json:"aliyunOSSAccessKeyId"`     // AccessKey ID
	AliyunOSSAccessKeySecret string `json:"aliyunOSSAccessKeySecret"` // AccessKey Secret（已脱敏）
	AliyunOSSBucketName      string `json:"aliyunOSSBucketName"`      // 存储桶名称
	AliyunOSSCDNDomain       string `json:"aliyunOSSCDNDomain"`       // CDN域名
	// 腾讯云COS配置
	TencentCOSBucketURL  string `json:"tencentCOSBucketUrl"`  // 存储桶URL
	TencentCOSSecretID   string `json:"tencentCOSSecretId"`   // SecretID（已脱敏）
	TencentCOSSecretKey  string `json:"tencentCOSSecretKey"`  // SecretKey（已脱敏）
	TencentCOSBucketName string `json:"tencentCOSBucketName"` // 存储桶名称
	TencentCOSRegion     string `json:"tencentCOSRegion"`     // 区域
	TencentCOSCDNDomain  string `json:"tencentCOSCDNDomain"`  // CDN域名
	// 基础字段
	CreateBy   string `json:"createBy"`   // 创建人
	CreateTime string `json:"createTime"` // 创建时间
	UpdateBy   string `json:"updateBy"`   // 更新人
	UpdateTime string `json:"updateTime"` // 更新时间
	IsActive   string `json:"isActive"`   // 是否激活
}

// toResponse 将实体转换为响应对象（脱敏敏感信息）
func toResponse(config *entity.CloudStorageConfig) *ConfigResponse {
	resp := &ConfigResponse{
		ID:                   config.ID,
		SysCompanyID:         config.SysCompanyID,
		StorageType:          config.StorageType,
		LocalBasePath:        config.LocalBasePath,
		LocalBaseURL:         config.LocalBaseURL,
		AliyunOSSEndpoint:    config.AliyunOSSEndpoint,
		AliyunOSSAccessKeyID: config.AliyunOSSAccessKeyID,
		AliyunOSSBucketName:  config.AliyunOSSBucketName,
		AliyunOSSCDNDomain:   config.AliyunOSSCDNDomain,
		TencentCOSBucketURL:  config.TencentCOSBucketURL,
		TencentCOSBucketName: config.TencentCOSBucketName,
		TencentCOSRegion:     config.TencentCOSRegion,
		TencentCOSCDNDomain:  config.TencentCOSCDNDomain,
		CreateBy:             config.CreateBy,
		CreateTime:           config.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateBy:             config.UpdateBy,
		UpdateTime:           config.UpdateTime.Format("2006-01-02 15:04:05"),
		IsActive:             config.IsActive,
	}

	// 脱敏敏感信息
	if resp.AliyunOSSAccessKeySecret != "" {
		if len(resp.AliyunOSSAccessKeySecret) > 8 {
			resp.AliyunOSSAccessKeySecret = resp.AliyunOSSAccessKeySecret[:4] + "****" + resp.AliyunOSSAccessKeySecret[len(resp.AliyunOSSAccessKeySecret)-4:]
		} else {
			resp.AliyunOSSAccessKeySecret = "****"
		}
	}

	if resp.TencentCOSSecretID != "" {
		if len(resp.TencentCOSSecretID) > 8 {
			resp.TencentCOSSecretID = resp.TencentCOSSecretID[:4] + "****" + resp.TencentCOSSecretID[len(resp.TencentCOSSecretID)-4:]
		} else {
			resp.TencentCOSSecretID = "****"
		}
	}

	if resp.TencentCOSSecretKey != "" {
		if len(resp.TencentCOSSecretKey) > 8 {
			resp.TencentCOSSecretKey = resp.TencentCOSSecretKey[:4] + "****" + resp.TencentCOSSecretKey[len(resp.TencentCOSSecretKey)-4:]
		} else {
			resp.TencentCOSSecretKey = "****"
		}
	}

	return resp
}

// CreateConfig 创建存储配置
// @Summary 创建存储配置
// @Description 创建公司的云存储配置
// @Tags Cloud-Storage-Config
// @Accept json
// @Produce json
// @Param request body CreateConfigRequest true "创建存储配置请求"
// @Success 200 {object} utils.Response{data=ConfigResponse}
// @Security BearerAuth
// @Router /api/v1/cloud/storage/config [post]
func (h *CloudStorageConfigHandler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	logger.Info("创建存储配置请求",
		zap.Uint("sysCompanyId", req.SysCompanyID),
		zap.String("storageType", req.StorageType))

	// 转换为实体
	config := &entity.CloudStorageConfig{
		SysCompanyID:             req.SysCompanyID,
		StorageType:              req.StorageType,
		LocalBasePath:            req.LocalBasePath,
		LocalBaseURL:             req.LocalBaseURL,
		AliyunOSSEndpoint:        req.AliyunOSSEndpoint,
		AliyunOSSAccessKeyID:     req.AliyunOSSAccessKeyID,
		AliyunOSSAccessKeySecret: req.AliyunOSSAccessKeySecret,
		AliyunOSSBucketName:      req.AliyunOSSBucketName,
		AliyunOSSCDNDomain:       req.AliyunOSSCDNDomain,
		TencentCOSBucketURL:      req.TencentCOSBucketURL,
		TencentCOSSecretID:       req.TencentCOSSecretID,
		TencentCOSSecretKey:      req.TencentCOSSecretKey,
		TencentCOSBucketName:     req.TencentCOSBucketName,
		TencentCOSRegion:         req.TencentCOSRegion,
		TencentCOSCDNDomain:      req.TencentCOSCDNDomain,
		IsActive:                 "Y", // 默认激活
	}

	// 创建配置
	if err := h.storageConfigService.CreateConfig(c.Request.Context(), config); err != nil {
		logger.Error("创建存储配置失败",
			zap.Uint("sysCompanyId", req.SysCompanyID),
			zap.String("storageType", req.StorageType),
			zap.Error(err))
		utils.InternalError(c, "创建存储配置失败: "+err.Error())
		return
	}

	// 返回创建成功的配置
	utils.Success(c, gin.H{
		"data":    toResponse(config),
		"message": "存储配置创建成功",
	})
}

// UpdateConfig 更新存储配置
// @Summary 更新存储配置
// @Description 更新公司的云存储配置
// @Tags Cloud-Storage-Config
// @Accept json
// @Produce json
// @Param request body UpdateConfigRequest true "更新存储配置请求"
// @Success 200 {object} utils.Response{data=ConfigResponse}
// @Security BearerAuth
// @Router /api/v1/cloud/storage/config/:id [put]
func (h *CloudStorageConfigHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "配置ID格式错误")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	logger.Info("更新存储配置请求",
		zap.Uint("id", uint(id)),
		zap.Uint("sysCompanyId", req.SysCompanyID),
		zap.String("storageType", req.StorageType))

	// 转换为实体
	config := &entity.CloudStorageConfig{
		ID:                       uint(id),
		SysCompanyID:             req.SysCompanyID,
		StorageType:              req.StorageType,
		LocalBasePath:            req.LocalBasePath,
		LocalBaseURL:             req.LocalBaseURL,
		AliyunOSSEndpoint:        req.AliyunOSSEndpoint,
		AliyunOSSAccessKeyID:     req.AliyunOSSAccessKeyID,
		AliyunOSSAccessKeySecret: req.AliyunOSSAccessKeySecret,
		AliyunOSSBucketName:      req.AliyunOSSBucketName,
		AliyunOSSCDNDomain:       req.AliyunOSSCDNDomain,
		TencentCOSBucketURL:      req.TencentCOSBucketURL,
		TencentCOSSecretID:       req.TencentCOSSecretID,
		TencentCOSSecretKey:      req.TencentCOSSecretKey,
		TencentCOSBucketName:     req.TencentCOSBucketName,
		TencentCOSRegion:         req.TencentCOSRegion,
		TencentCOSCDNDomain:      req.TencentCOSCDNDomain,
		IsActive:                 "Y",
	}

	// 更新配置
	if err := h.storageConfigService.UpdateConfig(c.Request.Context(), config); err != nil {
		logger.Error("更新存储配置失败",
			zap.Uint("id", uint(id)),
			zap.Uint("sysCompanyId", req.SysCompanyID),
			zap.String("storageType", req.StorageType),
			zap.Error(err))
		utils.InternalError(c, "更新存储配置失败: "+err.Error())
		return
	}

	// 返回更新后的配置
	utils.Success(c, gin.H{
		"data":    toResponse(config),
		"message": "存储配置更新成功",
	})
}

// GetCompanyConfig 获取公司存储配置
// @Summary 获取公司存储配置
// @Description 根据公司ID获取云存储配置
// @Tags Cloud-Storage-Config
// @Accept json
// @Produce json
// @Param companyId path int true "公司ID"
// @Success 200 {object} utils.Response{data=ConfigResponse}
// @Security BearerAuth
// @Router /api/v1/cloud/storage/config/company/:companyId [get]
func (h *CloudStorageConfigHandler) GetCompanyConfig(c *gin.Context) {
	companyIDStr := c.Param("companyId")
	companyID, err := strconv.ParseUint(companyIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "公司ID格式错误")
		return
	}

	logger.Debug("获取公司存储配置", zap.Uint("companyId", uint(companyID)))

	config, err := h.storageConfigService.GetCompanyConfig(c.Request.Context(), uint(companyID))
	if err != nil {
		logger.Error("获取公司存储配置失败",
			zap.Uint("companyId", uint(companyID)),
			zap.Error(err))
		utils.InternalError(c, "获取公司存储配置失败: "+err.Error())
		return
	}

	if config == nil {
		utils.NotFound(c, "未找到公司存储配置")
		return
	}

	utils.Success(c, gin.H{
		"data": toResponse(config),
	})
}

// GetAllConfigs 获取所有存储配置
// @Summary 获取所有存储配置
// @Description 获取所有有效的云存储配置
// @Tags Cloud-Storage-Config
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=[]ConfigResponse}
// @Security BearerAuth
// @Router /api/v1/cloud/storage/config [get]
func (h *CloudStorageConfigHandler) GetAllConfigs(c *gin.Context) {
	logger.Debug("获取所有存储配置")

	configs, err := h.storageConfigService.GetAllConfigs(c.Request.Context())
	if err != nil {
		logger.Error("获取所有存储配置失败", zap.Error(err))
		utils.InternalError(c, "获取所有存储配置失败: "+err.Error())
		return
	}

	// 转换为响应列表
	var responses []*ConfigResponse
	for _, config := range configs {
		responses = append(responses, toResponse(config))
	}

	utils.Success(c, gin.H{
		"data":  responses,
		"total": len(responses),
	})
}

// DeleteConfig 删除存储配置
// @Summary 删除存储配置
// @Description 软删除存储配置
// @Tags Cloud-Storage-Config
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/cloud/storage/config/:id [delete]
func (h *CloudStorageConfigHandler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "配置ID格式错误")
		return
	}

	logger.Info("删除存储配置请求", zap.Uint("id", uint(id)))

	if err := h.storageConfigService.DeleteConfig(c.Request.Context(), uint(id)); err != nil {
		logger.Error("删除存储配置失败",
			zap.Uint("id", uint(id)),
			zap.Error(err))
		utils.InternalError(c, "删除存储配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "存储配置删除成功",
	})
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

	logger.Info("刷新存储配置缓存请求", zap.Uint("companyId", uint(companyID)))

	if err := h.storageConfigService.RefreshCache(c.Request.Context(), uint(companyID)); err != nil {
		logger.Error("刷新存储配置缓存失败",
			zap.Uint("companyId", uint(companyID)),
			zap.Error(err))
		utils.InternalError(c, "刷新存储配置缓存失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "存储配置缓存刷新成功",
	})
}
