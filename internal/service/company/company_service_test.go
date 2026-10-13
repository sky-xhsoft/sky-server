package company

import (
	"context"
	"strings"
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
		Phone:           "13800000000",
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

func TestCompanyService_UpdateConfig_OnlyTencentCOS(t *testing.T) {
	svc := NewService(setupCompanyTestDB(t), nil)
	for _, v := range []string{"local", "aliyunOSS", ""} {
		st := v
		_, err := svc.UpdateCompanyConfig(context.Background(), 1, &UpdateCompanyConfigRequest{StorageType: &st})
		if err == nil || !strings.Contains(err.Error(), "云盘存储只支持云存储") {
			t.Errorf("存储类型 %q 应被拒绝，得到 %v", v, err)
		}
	}
}

// 用户名、手机号全局唯一（不同公司也不能重复），手机号格式校验
func TestCompanyService_Register_UniqueAndPhone(t *testing.T) {
	db := setupCompanyTestDB(t)
	svc := NewService(db, nil)
	base := func(username, phone string) *RegisterRequest {
		return &RegisterRequest{Name: "Co", Username: username, Phone: phone, Password: "Pass123!", ConfirmPassword: "Pass123!"}
	}
	if err := svc.Register(context.Background(), base("sajida", "13800000001")); err != nil {
		t.Fatalf("首次注册应成功: %v", err)
	}
	cases := []struct {
		req  *RegisterRequest
		want error
	}{
		{base("sajida", "13800000002"), ErrUsernameTaken},
		{base(" sajida ", "13800000003"), ErrUsernameTaken},
		{base("other", "13800000001"), ErrPhoneTaken},
		{base("other", "88888888888"), ErrInvalidPhone},
		{base("other", "1380000000"), ErrInvalidPhone},
	}
	for _, c := range cases {
		if err := svc.Register(context.Background(), c.req); err != c.want {
			t.Errorf("%s / %s: 期望 %v，实际 %v", c.req.Username, c.req.Phone, c.want, err)
		}
	}
	var companies int64
	db.Raw("SELECT COUNT(*) FROM sys_company").Scan(&companies)
	if companies != 1 {
		t.Errorf("校验失败时不应创建公司，实际 %d 家", companies)
	}

	res, err := svc.CheckRegisterAvailability(context.Background(), "sajida", "13800000001")
	if err != nil || !res.UsernameTaken || !res.PhoneTaken || !res.PhoneValid {
		t.Errorf("查重结果不对: %+v %v", res, err)
	}
	res, _ = svc.CheckRegisterAvailability(context.Background(), "new", "88888888888")
	if res.UsernameTaken || res.PhoneValid || res.PhoneTaken {
		t.Errorf("新用户名、格式错误的手机号: %+v", res)
	}
}
