package mask

import "testing"

func TestParseMask_FullMask(t *testing.T) {
	m := ParseMask("1111111111")
	if !m.AddVisible || !m.AddEditable || !m.EditVisible || !m.EditEditable ||
		!m.ListVisible || !m.ListEditable || !m.ImportVisible || !m.ExportVisible ||
		!m.PrintVisible || !m.Reserved {
		t.Error("All bits should be true for full mask")
	}
}

func TestParseMask_AllZeros(t *testing.T) {
	m := ParseMask("0000000000")
	if m.AddVisible || m.EditEditable || m.ListVisible {
		t.Error("All bits should be false for zero mask")
	}
}

func TestParseMask_ShortString(t *testing.T) {
	// 短字符串应自动补零
	m := ParseMask("11")
	if !m.AddVisible || !m.AddEditable {
		t.Error("First two bits should be true")
	}
	if m.EditVisible || m.ListVisible {
		t.Error("Subsequent bits should be false")
	}
}

func TestParseMask_LongString(t *testing.T) {
	// 长字符串应截断
	m := ParseMask("11111111110000")
	if !m.Reserved {
		t.Error("First 10 bits should be 1")
	}
}

func TestParseMask_ToString_RoundTrip(t *testing.T) {
	original := "1010101010"
	m := ParseMask(original)
	if m.ToString() != original {
		t.Errorf("Round-trip failed: %s -> %s", original, m.ToString())
	}
}

func TestIsVisible(t *testing.T) {
	// 1010101010: AddVisible=1, AddEditable=0, EditVisible=1, EditEditable=0,
	//             ListVisible=1, ListEditable=0, ImportVisible=1, ExportVisible=0,
	//             PrintVisible=1, Reserved=0
	m := ParseMask("1010101010")
	if !m.IsVisible("add") {
		t.Error("add should be visible")
	}
	if !m.IsVisible("edit") {
		t.Error("edit should be visible")
	}
	if !m.IsVisible("list") {
		t.Error("list should be visible")
	}
	if !m.IsVisible("import") {
		t.Error("import should be visible")
	}
	if m.IsVisible("export") {
		t.Error("export should NOT be visible (bit=0)")
	}
	if m.IsVisible("unknown") {
		t.Error("unknown operation should return false")
	}
}

func TestIsEditable(t *testing.T) {
	m := ParseMask("0101000000")
	if !m.AddEditable {
		t.Error("add should be editable")
	}
	if !m.IsEditable("add") {
		t.Error("IsEditable(add) should be true")
	}
	if m.IsEditable("export") {
		t.Error("export is not a valid editable target")
	}
}

func TestCanAccess(t *testing.T) {
	m := ParseMask("1111000000")
	if !m.CanAccess("add") {
		t.Error("add should be both visible and editable")
	}
	if m.CanAccess("list") {
		t.Error("list should not be accessible (both bits 0)")
	}
}

func TestHasAction(t *testing.T) {
	tests := []struct {
		mask     string
		action   rune
		expected bool
	}{
		{"AMDQ", 'A', true},
		{"AMDQ", 'M', true},
		{"AMDQ", 'D', true},
		{"AMDQ", 'X', false},
		{"AMDQSUV", 'S', true},
		{"AMDQSUV", 'U', true},
		{"AMDQSUV", 'V', true},
		{"AMDQSUV", 'I', false},
		{"AMDQSUVIE", 'I', true},
		{"AMDQSUVIE", 'E', true},
		{"", 'A', false},
	}

	for _, tt := range tests {
		got := HasAction(tt.mask, tt.action)
		if got != tt.expected {
			t.Errorf("HasAction(%q, %q): expected %v, got %v", tt.mask, tt.action, tt.expected, got)
		}
	}
}

func TestActionConstants(t *testing.T) {
	// 验证操作常量定义正确
	if ActionAdd != 'A' {
		t.Error("ActionAdd should be 'A'")
	}
	if ActionModify != 'M' {
		t.Error("ActionModify should be 'M'")
	}
	if ActionDelete != 'D' {
		t.Error("ActionDelete should be 'D'")
	}
	if ActionExport != 'E' {
		t.Error("ActionExport should be 'E'")
	}
}
