package metadata

import (
	"testing"
)

func TestForeignKeyOption(t *testing.T) {
	opt := &ForeignKeyOption{ID: 1, Label: "选项A"}
	if opt.ID != 1 {
		t.Errorf("expected ID=1, got %d", opt.ID)
	}
	if opt.Label != "选项A" {
		t.Errorf("expected Label='选项A', got '%s'", opt.Label)
	}
}

func TestForeignKeyQuery_Defaults(t *testing.T) {
	q := &ForeignKeyQuery{TableID: 10, Page: 0, PageSize: 0}
	if q.TableID != 10 {
		t.Errorf("expected TableID=10, got %d", q.TableID)
	}
	if q.Search != "" {
		t.Error("expected empty search")
	}
}

func TestForeignKeyDisplayQuery(t *testing.T) {
	colID := uint(5)
	q := &ForeignKeyDisplayQuery{TableID: 10, Value: "123", ColumnID: &colID}
	if q.TableID != 10 {
		t.Errorf("expected TableID=10, got %d", q.TableID)
	}
	if q.Value != "123" {
		t.Errorf("expected Value='123', got '%s'", q.Value)
	}
	if *q.ColumnID != 5 {
		t.Errorf("expected ColumnID=5, got %d", *q.ColumnID)
	}
}
