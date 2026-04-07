package validator

import (
	"testing"
)

// AssertError 断言有错误
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", msg)
	}
}

// AssertNotNil 断言不是nil
func AssertNotNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected value not to be nil", msg)
	}
}

// TestSimpleValidation 测试简单验证
func TestSimpleValidation(t *testing.T) {
	type TestStruct struct {
		Username string `json:"username" validate:"required,min=3,max=20"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Age      int    `json:"age" validate:"required,min=18,max=100"`
	}

	testCases := []struct {
		name     string
		data     interface{}
		expected bool // true if valid
	}{
		{
			name: "Valid data",
			data: &TestStruct{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Age:      25,
			},
			expected: true,
		},
		{
			name: "Missing required fields",
			data: &TestStruct{
				Email:    "test@example.com",
				Password: "password123",
				Age:      25,
			},
			expected: false,
		},
		{
			name: "Invalid email format",
			data: &TestStruct{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
				Age:      25,
			},
			expected: false,
		},
		{
			name: "Password too short",
			data: &TestStruct{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "12345",
				Age:      25,
			},
			expected: false,
		},
		{
			name: "Age too young",
			data: &TestStruct{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Age:      17,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStruct(tc.data)
			if tc.expected && err != nil {
				t.Errorf("Expected valid data, but got error: %v", err)
			} else if !tc.expected && err == nil {
				t.Error("Expected validation error, but got nil")
			} else if !tc.expected && err != nil {
				t.Logf("Validation error (expected): %v", err)
			}
		})
	}
}

// TestValidationErrorFormat 测试验证错误格式
func TestValidationErrorFormat(t *testing.T) {
	type TestStruct struct {
		Username string `json:"username" validate:"required,min=3"`
		Email    string `json:"email" validate:"required,email"`
	}

	data := &TestStruct{
		Username: "ab",
		Email:    "invalid",
	}

	err := ValidateStruct(data)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if len(validationErr.Errors) == 0 {
		t.Error("Expected validation errors, but got none")
	}

	t.Logf("Validation errors: %v", validationErr.Errors)
}

// TestRequiredValidation 测试必填字段验证
func TestRequiredValidation(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required"`
		Phone string `json:"phone" validate:"required"`
	}

	data := &TestStruct{
		Email: "test@example.com",
	}

	err := ValidateStruct(data)
	AssertError(t, err, "Validation failed")
	AssertNotNil(t, err, "Validation error should not be nil")
}

// TestEmailValidation 测试邮箱验证
func TestEmailValidation(t *testing.T) {
	type TestStruct struct {
		Email string `json:"email" validate:"required,email"`
	}

	testCases := []struct {
		name  string
		email string
		valid bool
	}{
		{name: "Valid email", email: "test@example.com", valid: true},
		{name: "Valid email with subdomain", email: "user.name+tag@example.co.uk", valid: true},
		{name: "Invalid format", email: "not-an-email", valid: false},
		{name: "Missing domain", email: "user@", valid: false},
		{name: "Missing local part", email: "@example.com", valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := &TestStruct{Email: tc.email}
			err := ValidateStruct(data)
			if tc.valid && err != nil {
				t.Errorf("Expected valid email %q, got error %v", tc.email, err)
			} else if !tc.valid && err == nil {
				t.Errorf("Expected invalid email %q to fail validation", tc.email)
			}
		})
	}
}

// TestNumberValidation 测试数字验证
func TestNumberValidation(t *testing.T) {
	type TestStruct struct {
		Age      int     `json:"age" validate:"required,min=18,max=100"`
		Price    float64 `json:"price" validate:"required,min=0,max=1000.00"`
		Quantity int     `json:"quantity" validate:"required,min=1,max=100"`
	}

	testCases := []struct {
		name     string
		data     *TestStruct
		expected bool
	}{
		{
			name:     "Valid values",
			data:     &TestStruct{Age: 25, Price: 99.99, Quantity: 5},
			expected: true,
		},
		{
			name:     "Age too young",
			data:     &TestStruct{Age: 17, Price: 99.99, Quantity: 5},
			expected: false,
		},
		{
			name:     "Price too high",
			data:     &TestStruct{Age: 25, Price: 1001.00, Quantity: 5},
			expected: false,
		},
		{
			name:     "Quantity too low",
			data:     &TestStruct{Age: 25, Price: 99.99, Quantity: 0},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStruct(tc.data)
			if tc.expected && err != nil {
				t.Errorf("Expected valid data, got: %v", err)
			} else if !tc.expected && err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}

// TestTagNames 测试标签名称提取
func TestTagNames(t *testing.T) {
	type TestStruct struct {
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"email_address" validate:"required,email"`
	}

	data := &TestStruct{}
	err := ValidateStruct(data)

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatal("Expected ValidationError")
	}

	if _, hasFullName := validationErr.Errors["full_name"]; !hasFullName {
		t.Error("Expected error for full_name")
	}

	if _, hasEmail := validationErr.Errors["email_address"]; !hasEmail {
		t.Error("Expected error for email_address")
	}

	t.Logf("Validation errors: %v", validationErr.Errors)
}
