package company

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupCompanyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	// 手动创建表，匹配 init.sql 的列名（CONTACT_NAME 等）
	createCompany := `
		CREATE TABLE sys_company (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_COMPANY_ID INTEGER,
			NAME TEXT,
			CODE TEXT,
			CONTACT_NAME TEXT,
			CONTACT_PHONE TEXT,
			EMAIL TEXT,
			IS_ACTIVE TEXT DEFAULT 'Y',
			CREATE_TIME DATETIME
		)
	`
	createUser := `
		CREATE TABLE sys_user (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			SYS_COMPANY_ID INTEGER,
			USERNAME TEXT,
			PASSWORD TEXT,
			TRUE_NAME TEXT,
			EMAIL TEXT,
			PHONE TEXT,
			IS_ADMIN TEXT DEFAULT 'N',
			IS_ACTIVE TEXT DEFAULT 'Y',
			CREATE_TIME DATETIME
		)
	`
	if err := db.Exec(createCompany).Error; err != nil {
		t.Fatalf("Failed to create sys_company: %v", err)
	}
	if err := db.Exec(createUser).Error; err != nil {
		t.Fatalf("Failed to create sys_user: %v", err)
	}
	return db
}

func TestCompanyService_Register_PasswordMismatch(t *testing.T) {
	db := setupCompanyTestDB(t)
	svc := NewService(db, nil)

	err := svc.Register(context.Background(), &RegisterRequest{
		Name:            "Test Co",
		Code:            "TC",
		Username:        "admin",
		Password:        "Pass123!",
		ConfirmPassword: "Different!",
		AdminName:       "Admin",
	})

	if err == nil {
		t.Fatal("Expected error for password mismatch")
	}
}

func TestCompanyService_Register_Success(t *testing.T) {
	db := setupCompanyTestDB(t)
	svc := NewService(db, nil)

	err := svc.Register(context.Background(), &RegisterRequest{
		Name:            "Test Co",
		Code:            "TC001",
		Email:           "test@example.com",
		Username:        "admin",
		Password:        "Pass123!",
		ConfirmPassword: "Pass123!",
		AdminName:       "Admin User",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证公司已创建
	var count int64
	db.Raw("SELECT COUNT(*) FROM sys_company").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 company, got %d", count)
	}

	// 验证管理员用户已创建（密码已加密）
	var password, isAdmin string
	db.Raw("SELECT PASSWORD, IS_ADMIN FROM sys_user WHERE USERNAME = ?", "admin").Row().Scan(&password, &isAdmin)
	if password == "Pass123!" {
		t.Error("Password should be hashed")
	}
	if isAdmin != "Y" {
		t.Errorf("Expected IS_ADMIN=Y, got %s", isAdmin)
	}
}
