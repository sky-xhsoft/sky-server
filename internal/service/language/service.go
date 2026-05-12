package language

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// ListRequest 翻译列表请求
type ListRequest struct {
	Language string `form:"language"`
	Module   string `form:"module"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// AddRequest 添加翻译请求
type AddRequest struct {
	Key      string `json:"key" binding:"required"`
	Language string `json:"language" binding:"required"`
	Value    string `json:"value" binding:"required"`
	Module   string `json:"module"`
}

// UpdateRequest 更新翻译请求
type UpdateRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Key      string `json:"key" binding:"required"`
	Language string `json:"language" binding:"required"`
	Value    string `json:"value" binding:"required"`
	Module   string `json:"module"`
}

// Service 语言服务接口
type Service interface {
	// 列出所有语言
	ListLanguages(companyID uint) ([]*entity.SysLanguage, error)
	// 获取翻译
	GetTranslation(companyID uint, key string, language string) (*entity.SysTranslation, error)
	// 列出翻译
	ListTranslations(companyID uint, req *ListRequest) ([]*entity.SysTranslation, int, error)
	// 获取模块翻译
	GetModuleTranslations(companyID uint, module string, language string) ([]*entity.SysTranslation, error)
	// 添加翻译
	AddTranslation(translation *AddRequest, userID uint, username string, companyID uint) error
	// 更新翻译
	UpdateTranslation(req *UpdateRequest, userID uint, username string, companyID uint) error
	// 删除翻译
	DeleteTranslation(id uint) error
	// 导出翻译
	ExportTranslations(companyID uint, language string) ([]byte, error)
	// 导入翻译
	ImportTranslations(companyID uint, data []byte, language string, userID uint, username string) error
}

type service struct {
	langRepo    repository.LanguageRepository
	transRepo   repository.TranslationRepository
	redisClient *redis.Client
	logger      *zap.Logger
	ctx         context.Context
	cacheTTL    time.Duration
}

// NewService 创建语言服务
func NewService(langRepo repository.LanguageRepository, transRepo repository.TranslationRepository, redisClient *redis.Client, logger *zap.Logger, cacheTTL int) Service {
	return &service{
		langRepo:    langRepo,
		transRepo:   transRepo,
		redisClient: redisClient,
		logger:      logger,
		ctx:         context.Background(),
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
	}
}

// ListLanguages 列出所有语言
func (s *service) ListLanguages(companyID uint) ([]*entity.SysLanguage, error) {
	cacheKey := "languages:" + strconv.FormatUint(uint64(companyID), 10)

	var languages []*entity.SysLanguage

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cached), &languages); err == nil {
				return languages, nil
			}
		}
	}

	var err error
	languages, err = s.langRepo.ListLanguages(companyID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询语言列表失败", err)
	}

	if s.redisClient != nil {
		data, _ := json.Marshal(languages)
		s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)
	}

	return languages, nil
}

// GetTranslation 获取翻译
func (s *service) GetTranslation(companyID uint, key string, language string) (*entity.SysTranslation, error) {
	cacheKey := "translation:" + strconv.FormatUint(uint64(companyID), 10) + ":" + key + ":" + language

	var translation *entity.SysTranslation

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(s.ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cached), &translation); err == nil && translation != nil {
				return translation, nil
			}
		}
	}

	// 从数据库查询
	translation, err := s.transRepo.GetTranslationByKey(companyID, key, language)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询翻译失败", err)
	}

	if translation == nil && language != "zh-CN" {
		// 如果没找到，尝试查找中文
		translation, _ = s.transRepo.GetTranslationByKey(companyID, key, "zh-CN")
	}

	if s.redisClient != nil && translation != nil {
		data, _ := json.Marshal(translation)
		s.redisClient.Set(s.ctx, cacheKey, data, s.cacheTTL)
	}

	return translation, nil
}

// ListTranslations 列出翻译
func (s *service) ListTranslations(companyID uint, req *ListRequest) ([]*entity.SysTranslation, int, error) {
	translations, total, err := s.transRepo.ListTranslations(companyID, req.Module, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询翻译列表失败", err)
	}
	return translations, total, nil
}

// GetModuleTranslations 获取模块翻译
func (s *service) GetModuleTranslations(companyID uint, module string, language string) ([]*entity.SysTranslation, error) {
	translations, err := s.transRepo.GetTranslationsByModule(companyID, module, language)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询翻译列表失败", err)
	}
	return translations, nil
}

// AddTranslation 添加翻译
func (s *service) AddTranslation(req *AddRequest, userID uint, username string, companyID uint) error {
	now := time.Now()

	translation := &entity.SysTranslation{
		Key:        req.Key,
		Language:   req.Language,
		Value:      req.Value,
		Module:     req.Module,
		IsActive:   "Y",
		CreateBy:   username,
		CreateTime: now,
		UpdateBy:   username,
		UpdateTime: now,
	}
	translation.SysCompanyID = companyID

	if err := s.transRepo.AddTranslation(translation); err != nil {
		return errors.Wrap(errors.ErrDatabase, "添加翻译失败", err)
	}

	// 清除缓存
	s.clearTranslationCache(companyID)

	return nil
}

// UpdateTranslation 更新翻译
func (s *service) UpdateTranslation(req *UpdateRequest, userID uint, username string, companyID uint) error {
	// 查询现有翻译
	translation, err := s.transRepo.GetTranslationByID(req.ID)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "翻译不存在", err)
	}
	if translation == nil {
		return errors.ResourceNotFound
	}

	now := time.Now()
	translation.Key = req.Key
	translation.Language = req.Language
	translation.Value = req.Value
	translation.Module = req.Module
	translation.UpdateBy = username
	translation.UpdateTime = now

	if err := s.transRepo.UpdateTranslation(translation); err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新翻译失败", err)
	}

	s.clearTranslationCache(companyID)

	return nil
}

// DeleteTranslation 删除翻译
func (s *service) DeleteTranslation(id uint) error {
	if err := s.transRepo.DeleteTranslation(id); err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除翻译失败", err)
	}

	s.clearTranslationCache(0) // 清除所有公司的缓存

	return nil
}

// ExportTranslations 导出翻译
func (s *service) ExportTranslations(companyID uint, language string) ([]byte, error) {
	translations, err := s.transRepo.GetTranslationsByModule(companyID, "", language)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询翻译列表失败", err)
	}

	// 转换为键值对格式
	kvMap := make(map[string]string)
	for _, t := range translations {
		kvMap[t.Key] = t.Value
	}

	return json.MarshalIndent(kvMap, "", "  ")
}

// ImportTranslations 导入翻译
func (s *service) ImportTranslations(companyID uint, data []byte, language string, userID uint, username string) error {
	var kvMap map[string]string
	if err := json.Unmarshal(data, &kvMap); err != nil {
		return errors.Wrap(errors.ErrInvalidParam, "解析翻译数据失败", err)
	}

	var translations []*entity.SysTranslation
	now := time.Now()
	for key, value := range kvMap {
		trans := &entity.SysTranslation{
			Key:        key,
			Language:   language,
			Value:      value,
			IsActive:   "Y",
			CreateBy:   username,
			CreateTime: now,
			UpdateBy:   username,
			UpdateTime: now,
		}
		trans.SysCompanyID = companyID
		translations = append(translations, trans)
	}

	if err := s.transRepo.BatchAddTranslations(translations); err != nil {
		return errors.Wrap(errors.ErrDatabase, "导入翻译失败", err)
	}

	s.clearTranslationCache(companyID)

	return nil
}

// clearTranslationCache 清除翻译缓存
func (s *service) clearTranslationCache(companyID uint) {
	if s.redisClient == nil {
		return
	}

	if companyID == 0 {
		// 清除所有翻译缓存
		pattern := "translation:*"
		iter := s.redisClient.Scan(s.ctx, 0, pattern, 0).Iterator()

		keys := []string{}
		for iter.Next(s.ctx) {
			keys = append(keys, iter.Val())
		}

		if len(keys) > 0 {
			s.redisClient.Del(s.ctx, keys...)
		}

		pattern = "languages:*"
		iter = s.redisClient.Scan(s.ctx, 0, pattern, 0).Iterator()

		keys = []string{}
		for iter.Next(s.ctx) {
			keys = append(keys, iter.Val())
		}

		if len(keys) > 0 {
			s.redisClient.Del(s.ctx, keys...)
		}
	} else {
		// 清除指定公司的翻译缓存
		companyIDStr := strconv.FormatUint(uint64(companyID), 10)
		pattern := "translation:" + companyIDStr + ":*"
		iter := s.redisClient.Scan(s.ctx, 0, pattern, 0).Iterator()

		keys := []string{}
		for iter.Next(s.ctx) {
			keys = append(keys, iter.Val())
		}

		if len(keys) > 0 {
			s.redisClient.Del(s.ctx, keys...)
		}

		s.redisClient.Del(s.ctx, "languages:"+companyIDStr)
	}
}
