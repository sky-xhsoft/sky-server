package groups

import (
	"testing"
)

func TestPermissionBitConstants(t *testing.T) {
	if PermNone != 0 {
		t.Error("PermNone should be 0")
	}
	if PermRead != 1 {
		t.Error("PermRead should be 1")
	}
	if PermCreate != 2 {
		t.Error("PermCreate should be 2")
	}
	if PermUpdate != 4 {
		t.Error("PermUpdate should be 4")
	}
	if PermDelete != 8 {
		t.Error("PermDelete should be 8")
	}
	if PermAll != 255 {
		t.Errorf("PermAll should be 255, got %d", PermAll)
	}
}

func TestPermissionBitOperations(t *testing.T) {
	// 组合多个权限
	combined := PermRead | PermCreate | PermUpdate
	if combined != 7 {
		t.Errorf("Read|Create|Update should be 7, got %d", combined)
	}

	// 检查权限位
	if (combined & PermRead) == 0 {
		t.Error("Combined should include Read")
	}
	if (combined & PermDelete) != 0 {
		t.Error("Combined should NOT include Delete")
	}

	// PermAll 应包含所有
	if (PermAll & PermRead) == 0 {
		t.Error("PermAll should include Read")
	}
	if (PermAll & PermDelete) == 0 {
		t.Error("PermAll should include Delete")
	}
	if (PermAll & PermImport) == 0 {
		t.Error("PermAll should include Import")
	}
}

func TestPermissionMaskCheck(t *testing.T) {
	// 模拟权限检查逻辑：(userPerm & required) == required
	userPerm := PermRead | PermCreate

	// 应该有读权限
	if (userPerm & PermRead) != PermRead {
		t.Error("Should have read permission")
	}

	// 不应该有删除权限
	if (userPerm & PermDelete) == PermDelete {
		t.Error("Should NOT have delete permission")
	}

	// 多权限同时检查
	readCreate := PermRead | PermCreate
	if (userPerm & readCreate) != readCreate {
		t.Error("Should have both read and create permissions")
	}
}
