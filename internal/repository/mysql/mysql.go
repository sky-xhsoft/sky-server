package mysql

import (
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化MySQL连接
func Init(cfg *config.MySQLConfig, log *zap.Logger) (*gorm.DB, error) {
	// 配置GORM日志
	gormLogger := logger.New(
		&gormLoggerWriter{logger: log},
		logger.Config{
			SlowThreshold:             500 * time.Millisecond, // 慢查询阈值
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return db, nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// gormLoggerWriter GORM日志写入器，将GORM日志写入到zap
type gormLoggerWriter struct {
	logger *zap.Logger
}

// Printf 实现gorm/logger.Writer接口
func (w *gormLoggerWriter) Printf(format string, args ...interface{}) {
	w.logger.Sugar().Infof(format, args...)
}
