package mask

import "strings"

// FieldMask 字段MASK结构（10位）
type FieldMask struct {
	AddVisible    bool // 位1: 新增可见
	AddEditable   bool // 位2: 新增可修改
	EditVisible   bool // 位3: 修改可见
	EditEditable  bool // 位4: 修改可修改
	ListVisible   bool // 位5: 列表可见
	ListEditable  bool // 位6: 列表可修改
	ImportVisible bool // 位7: 导入可见
	ExportVisible bool // 位8: 导出可见
	PrintVisible  bool // 位9: 打印可见
	Reserved      bool // 位10: 扩展预留
}

// ParseMask 从字符串解析MASK（10位）
func ParseMask(maskStr string) *FieldMask {
	// 确保字符串长度为10
	if len(maskStr) < 10 {
		maskStr = maskStr + strings.Repeat("0", 10-len(maskStr))
	}
	if len(maskStr) > 10 {
		maskStr = maskStr[:10]
	}

	mask := &FieldMask{}
	mask.AddVisible = maskStr[0] == '1'
	mask.AddEditable = maskStr[1] == '1'
	mask.EditVisible = maskStr[2] == '1'
	mask.EditEditable = maskStr[3] == '1'
	mask.ListVisible = maskStr[4] == '1'
	mask.ListEditable = maskStr[5] == '1'
	mask.ImportVisible = maskStr[6] == '1'
	mask.ExportVisible = maskStr[7] == '1'
	mask.PrintVisible = maskStr[8] == '1'
	mask.Reserved = maskStr[9] == '1'

	return mask
}

// ToString 生成MASK字符串
func (m *FieldMask) ToString() string {
	result := ""
	if m.AddVisible {
		result += "1"
	} else {
		result += "0"
	}
	if m.AddEditable {
		result += "1"
	} else {
		result += "0"
	}
	if m.EditVisible {
		result += "1"
	} else {
		result += "0"
	}
	if m.EditEditable {
		result += "1"
	} else {
		result += "0"
	}
	if m.ListVisible {
		result += "1"
	} else {
		result += "0"
	}
	if m.ListEditable {
		result += "1"
	} else {
		result += "0"
	}
	if m.ImportVisible {
		result += "1"
	} else {
		result += "0"
	}
	if m.ExportVisible {
		result += "1"
	} else {
		result += "0"
	}
	if m.PrintVisible {
		result += "1"
	} else {
		result += "0"
	}
	if m.Reserved {
		result += "1"
	} else {
		result += "0"
	}
	return result
}

// IsVisible 检查指定操作是否可见
func (m *FieldMask) IsVisible(operation string) bool {
	switch operation {
	case "add":
		return m.AddVisible
	case "edit":
		return m.EditVisible
	case "list":
		return m.ListVisible
	case "import":
		return m.ImportVisible
	case "export":
		return m.ExportVisible
	case "print":
		return m.PrintVisible
	default:
		return false
	}
}

// IsEditable 检查指定操作是否可编辑
func (m *FieldMask) IsEditable(operation string) bool {
	switch operation {
	case "add":
		return m.AddEditable
	case "edit":
		return m.EditEditable
	case "list":
		return m.ListEditable
	default:
		return false
	}
}

// CanAccess 检查字段在指定操作下是否可访问（可见且可编辑）
func (m *FieldMask) CanAccess(operation string) bool {
	return m.IsVisible(operation) && m.IsEditable(operation)
}
