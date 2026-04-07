package test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestLogger 测试用的日志实现
type TestLogger struct {
	logger *zap.Logger
}

func NewTestLogger() *TestLogger {
	logger, _ := zap.NewDevelopment()
	return &TestLogger{logger: logger}
}

func (l *TestLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	l.logger.Sugar().Debugf(format, args...)
}

func (l *TestLogger) Info(ctx context.Context, format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}

func (l *TestLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	l.logger.Sugar().Warnf(format, args...)
}

func (l *TestLogger) Error(ctx context.Context, format string, args ...interface{}) {
	l.logger.Sugar().Errorf(format, args...)
}

// CreateTestDB 创建内存数据库用于测试
func CreateTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db
}

// CreateTestSQLiteDB 创建SQLite数据库用于测试
func CreateTestSQLiteDB(t *testing.T, dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	return db
}

// CreateTestSQLDB 创建SQL数据库连接
func CreateTestSQLDB(t *testing.T, driver, dsn string) *sql.DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		t.Fatalf("Failed to open SQL database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

// SetupGormTestDB 设置GORM测试数据库
func SetupGormTestDB(t *testing.T) *gorm.DB {
	db := CreateTestDB(t)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Logf("Failed to get SQL DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			t.Logf("Failed to close DB: %v", err)
		}
	})

	return db
}

// CreateTestContext 创建测试上下文
func CreateTestContext() context.Context {
	return context.Background()
}

// TestTableCleanup 清理测试数据表
func TestTableCleanup(db *gorm.DB, tables ...string) error {
	for _, table := range tables {
		if err := db.Exec("DROP TABLE IF EXISTS " + table).Error; err != nil {
			return err
		}
	}
	return nil
}

// TestTransaction 事务测试辅助函数
func TestTransaction(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB) error) {
	err := db.Transaction(fn)
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

// AssertNoError 断言没有错误
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// AssertError 断言有错误
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", msg)
	}
}

// AssertEqual 断言相等
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertNotNil 断言不是nil
func AssertNotNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected value not to be nil", msg)
	}
}

// AssertNil 断言是nil
func AssertNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value != nil {
		t.Fatalf("%s: expected nil but got %v", msg, value)
	}
}
