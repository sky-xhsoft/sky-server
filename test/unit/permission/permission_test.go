package permission

import (
	"reflect"
	"testing"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name        string
		userPerm    int
		requirePerm int
		expected    bool
	}{
		{"有读权限", ReadWrite, Read, true},
		{"有读写权限", ReadWrite, ReadWrite, true},
		{"没有提交权限", ReadWrite, Submit, false},
		{"全部权限包含读", All, Read, true},
		{"全部权限包含审核", All, Audit, true},
		{"无权限", None, Read, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasPermission(tt.userPerm, tt.requirePerm)
			if result != tt.expected {
				t.Errorf("HasPermission(%d, %d) = %v, want %v", tt.userPerm, tt.requirePerm, result, tt.expected)
			}
		})
	}
}

func TestAddPermission(t *testing.T) {
	tests := []struct {
		name     string
		userPerm int
		addPerm  int
		expected int
	}{
		{"添加写权限到读权限", Read, Write, ReadWrite},
		{"添加提交权限到读写权限", ReadWrite, Submit, ReadWriteSubmit},
		{"添加导出权限", AllNoExport, Export, All},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddPermission(tt.userPerm, tt.addPerm)
			if result != tt.expected {
				t.Errorf("AddPermission(%d, %d) = %d, want %d", tt.userPerm, tt.addPerm, result, tt.expected)
			}
		})
	}
}

func TestRemovePermission(t *testing.T) {
	tests := []struct {
		name       string
		userPerm   int
		removePerm int
		expected   int
	}{
		{"移除写权限", ReadWrite, Write, Read},
		{"移除提交权限", ReadWriteSubmit, Submit, ReadWrite},
		{"移除导出权限", All, Export, AllNoExport},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemovePermission(tt.userPerm, tt.removePerm)
			if result != tt.expected {
				t.Errorf("RemovePermission(%d, %d) = %d, want %d", tt.userPerm, tt.removePerm, result, tt.expected)
			}
		})
	}
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name     string
		perm     int
		expected []string
	}{
		{"读权限", Read, []string{"read"}},
		{"读写权限", ReadWrite, []string{"read", "write"}},
		{"读写提交权限", ReadWriteSubmit, []string{"read", "write", "submit"}},
		{"全部权限", All, []string{"read", "write", "submit", "audit", "export"}},
		{"无权限", None, []string(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePermission(tt.perm)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePermission(%d) = %v, want %v", tt.perm, result, tt.expected)
			}
		})
	}
}

func TestBuildPermission(t *testing.T) {
	tests := []struct {
		name     string
		perms    []string
		expected int
	}{
		{"读权限", []string{"read"}, Read},
		{"读写权限", []string{"read", "write"}, ReadWrite},
		{"读写提交权限", []string{"read", "write", "submit"}, ReadWriteSubmit},
		{"全部权限", []string{"read", "write", "submit", "audit", "export"}, All},
		{"无权限", []string{}, None},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPermission(tt.perms)
			if result != tt.expected {
				t.Errorf("BuildPermission(%v) = %d, want %d", tt.perms, result, tt.expected)
			}
		})
	}
}

func TestCanRead(t *testing.T) {
	tests := []struct {
		name     string
		perm     int
		expected bool
	}{
		{"有读权限", Read, true},
		{"有读写权限", ReadWrite, true},
		{"无读权限", None, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanRead(tt.perm)
			if result != tt.expected {
				t.Errorf("CanRead(%d) = %v, want %v", tt.perm, result, tt.expected)
			}
		})
	}
}

func TestCanWrite(t *testing.T) {
	tests := []struct {
		name     string
		perm     int
		expected bool
	}{
		{"有读写权限", ReadWrite, true},
		{"只有读权限", Read, false},
		{"无权限", None, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanWrite(tt.perm)
			if result != tt.expected {
				t.Errorf("CanWrite(%d) = %v, want %v", tt.perm, result, tt.expected)
			}
		})
	}
}
