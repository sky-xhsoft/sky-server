package helper

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CreateTestDB 创建测试数据库（内存SQLite）
func CreateTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
