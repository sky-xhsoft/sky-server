package service

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// BaseService 服务基类
type BaseService struct {
	logger *zap.Logger
}

// NewBaseService 创建服务基类实例
func NewBaseService(logger *zap.Logger) *BaseService {
	if logger == nil {
		logger = zap.NewNop() // 使用空日志器作为默认值
	}
	return &BaseService{
		logger: logger,
	}
}

// Logger 获取日志器
func (s *BaseService) Logger() *zap.Logger {
	return s.logger
}

// Errorf 创建格式化的业务错误
func (s *BaseService) Errorf(code int, format string, args ...interface{}) error {
	return errors.New(code, fmt.Sprintf(format, args...))
}

// WrapError 包装错误
func (s *BaseService) WrapError(code int, message string, err error) error {
	return errors.Wrap(code, message, err)
}

// LogError 记录错误日志
func (s *BaseService) LogError(ctx context.Context, err error, fields ...zap.Field) {
	if err == nil {
		return
	}

	logFields := []zap.Field{}
	if ctx != nil {
		if requestID, ok := ctx.Value("requestID").(string); ok {
			logFields = append(logFields, zap.String("requestID", requestID))
		}
		if userID, ok := ctx.Value("userID").(uint); ok {
			logFields = append(logFields, zap.Uint("userID", userID))
		}
	}

	logFields = append(logFields, fields...)
	logFields = append(logFields, zap.Error(err))

	s.logger.Error("Service error", logFields...)
}

// LogInfo 记录信息日志
func (s *BaseService) LogInfo(ctx context.Context, message string, fields ...zap.Field) {
	logFields := []zap.Field{}
	if ctx != nil {
		if requestID, ok := ctx.Value("requestID").(string); ok {
			logFields = append(logFields, zap.String("requestID", requestID))
		}
		if userID, ok := ctx.Value("userID").(uint); ok {
			logFields = append(logFields, zap.Uint("userID", userID))
		}
	}

	logFields = append(logFields, fields...)

	s.logger.Info(message, logFields...)
}

// LogWarn 记录警告日志
func (s *BaseService) LogWarn(ctx context.Context, message string, fields ...zap.Field) {
	logFields := []zap.Field{}
	if ctx != nil {
		if requestID, ok := ctx.Value("requestID").(string); ok {
			logFields = append(logFields, zap.String("requestID", requestID))
		}
		if userID, ok := ctx.Value("userID").(uint); ok {
			logFields = append(logFields, zap.Uint("userID", userID))
		}
	}

	logFields = append(logFields, fields...)

	s.logger.Warn(message, logFields...)
}
