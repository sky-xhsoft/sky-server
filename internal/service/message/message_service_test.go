package message

import (
	"testing"
)

// 创建测试用 service 实例（无需 DB/WS）
func newServiceForTest() *service {
	return &service{}
}

func TestService_ReplaceVariables_Simple(t *testing.T) {
	s := newServiceForTest()
	template := "Hello {{name}}, your code is {{code}}."
	vars := map[string]interface{}{
		"name": "Alice",
		"code": "12345",
	}

	result := s.replaceVariables(template, vars)
	expected := "Hello Alice, your code is 12345."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestService_ReplaceVariables_NoMatch(t *testing.T) {
	s := newServiceForTest()
	template := "No variables here"
	vars := map[string]interface{}{"unused": "value"}

	result := s.replaceVariables(template, vars)
	if result != template {
		t.Errorf("Expected unchanged template, got %q", result)
	}
}

func TestService_ReplaceVariables_MissingVar(t *testing.T) {
	s := newServiceForTest()
	template := "Hi {{name}}, age {{age}}"
	vars := map[string]interface{}{"name": "Bob"}

	result := s.replaceVariables(template, vars)
	// age 未提供，{{age}} 应原样保留
	if result != "Hi Bob, age {{age}}" {
		t.Errorf("Expected partial replacement, got %q", result)
	}
}

func TestService_ReplaceVariables_NumberValues(t *testing.T) {
	s := newServiceForTest()
	template := "Count: {{count}}, Price: {{price}}"
	vars := map[string]interface{}{
		"count": 42,
		"price": 3.14,
	}

	result := s.replaceVariables(template, vars)
	// 数字值应转为字符串
	if result == template {
		t.Error("Variables should be replaced")
	}
}

func TestService_GetSenderName_Nil(t *testing.T) {
	s := newServiceForTest()
	name := s.getSenderName(nil)
	if name != "system" {
		t.Errorf("Expected 'system' for nil sender, got %q", name)
	}
}

func TestService_GetSenderName_WithID(t *testing.T) {
	s := newServiceForTest()
	uid := uint(42)
	name := s.getSenderName(&uid)
	if name != "user_42" {
		t.Errorf("Expected 'user_42', got %q", name)
	}
}

func TestService_UintsToString(t *testing.T) {
	s := newServiceForTest()
	tests := []struct {
		input    []uint
		expected string
	}{
		{[]uint{}, ""},
		{[]uint{1}, "1"},
		{[]uint{1, 2, 3}, "1,2,3"},
		{[]uint{100, 200}, "100,200"},
	}

	for _, tt := range tests {
		got := s.uintsToString(tt.input)
		if got != tt.expected {
			t.Errorf("uintsToString(%v): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}
