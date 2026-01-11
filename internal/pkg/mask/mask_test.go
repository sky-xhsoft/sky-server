package mask

import (
	"testing"
)

func TestParseMask(t *testing.T) {
	tests := []struct {
		name     string
		maskStr  string
		expected *FieldMask
	}{
		{
			name:    "全部可见可编辑",
			maskStr: "1111111111",
			expected: &FieldMask{
				AddVisible:    true,
				AddEditable:   true,
				EditVisible:   true,
				EditEditable:  true,
				ListVisible:   true,
				ListEditable:  true,
				ImportVisible: true,
				ExportVisible: true,
				PrintVisible:  true,
				Reserved:      true,
			},
		},
		{
			name:    "新增可见可编辑，其他不可见",
			maskStr: "1100000000",
			expected: &FieldMask{
				AddVisible:    true,
				AddEditable:   true,
				EditVisible:   false,
				EditEditable:  false,
				ListVisible:   false,
				ListEditable:  false,
				ImportVisible: false,
				ExportVisible: false,
				PrintVisible:  false,
				Reserved:      false,
			},
		},
		{
			name:    "只读（列表可见）",
			maskStr: "0000100000",
			expected: &FieldMask{
				AddVisible:    false,
				AddEditable:   false,
				EditVisible:   false,
				EditEditable:  false,
				ListVisible:   true,
				ListEditable:  false,
				ImportVisible: false,
				ExportVisible: false,
				PrintVisible:  false,
				Reserved:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMask(tt.maskStr)
			if result.AddVisible != tt.expected.AddVisible ||
				result.AddEditable != tt.expected.AddEditable ||
				result.EditVisible != tt.expected.EditVisible ||
				result.EditEditable != tt.expected.EditEditable ||
				result.ListVisible != tt.expected.ListVisible ||
				result.ListEditable != tt.expected.ListEditable ||
				result.ImportVisible != tt.expected.ImportVisible ||
				result.ExportVisible != tt.expected.ExportVisible ||
				result.PrintVisible != tt.expected.PrintVisible ||
				result.Reserved != tt.expected.Reserved {
				t.Errorf("ParseMask() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestFieldMask_ToString(t *testing.T) {
	tests := []struct {
		name     string
		mask     *FieldMask
		expected string
	}{
		{
			name: "全部可见可编辑",
			mask: &FieldMask{
				AddVisible:    true,
				AddEditable:   true,
				EditVisible:   true,
				EditEditable:  true,
				ListVisible:   true,
				ListEditable:  true,
				ImportVisible: true,
				ExportVisible: true,
				PrintVisible:  true,
				Reserved:      true,
			},
			expected: "1111111111",
		},
		{
			name: "新增可见可编辑，其他不可见",
			mask: &FieldMask{
				AddVisible:  true,
				AddEditable: true,
			},
			expected: "1100000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mask.ToString()
			if result != tt.expected {
				t.Errorf("ToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldMask_IsVisible(t *testing.T) {
	mask := ParseMask("1111111111")

	tests := []struct {
		operation string
		expected  bool
	}{
		{"add", true},
		{"edit", true},
		{"list", true},
		{"import", true},
		{"export", true},
		{"print", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			result := mask.IsVisible(tt.operation)
			if result != tt.expected {
				t.Errorf("IsVisible(%s) = %v, want %v", tt.operation, result, tt.expected)
			}
		})
	}
}

func TestFieldMask_IsEditable(t *testing.T) {
	mask := ParseMask("1111111111")

	tests := []struct {
		operation string
		expected  bool
	}{
		{"add", true},
		{"edit", true},
		{"list", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			result := mask.IsEditable(tt.operation)
			if result != tt.expected {
				t.Errorf("IsEditable(%s) = %v, want %v", tt.operation, result, tt.expected)
			}
		})
	}
}
