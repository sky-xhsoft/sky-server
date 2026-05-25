package common

import (
	"testing"
)

func TestPageRequest_Validate_Defaults(t *testing.T) {
	p := &PageRequest{Page: 0, PageSize: 0}
	p.Validate()
	if p.Page != 1 {
		t.Errorf("expected Page=1, got %d", p.Page)
	}
	if p.PageSize != 10 {
		t.Errorf("expected PageSize=10, got %d", p.PageSize)
	}
}

func TestPageRequest_Validate_Caps(t *testing.T) {
	p := &PageRequest{Page: 5, PageSize: 200}
	p.Validate()
	if p.PageSize != 100 {
		t.Errorf("expected PageSize capped at 100, got %d", p.PageSize)
	}
}

func TestPageRequest_Validate_KeepsValid(t *testing.T) {
	p := &PageRequest{Page: 3, PageSize: 20}
	p.Validate()
	if p.Page != 3 {
		t.Errorf("expected Page=3, got %d", p.Page)
	}
	if p.PageSize != 20 {
		t.Errorf("expected PageSize=20, got %d", p.PageSize)
	}
}

func TestPageRequest_Offset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 20, 40},
		{1, 50, 0},
	}

	for _, tt := range tests {
		p := &PageRequest{Page: tt.page, PageSize: tt.pageSize}
		if off := p.Offset(); off != tt.expected {
			t.Errorf("page=%d pageSize=%d offset=%d, expected %d", tt.page, tt.pageSize, off, tt.expected)
		}
	}
}

func TestNewPageResponse(t *testing.T) {
	data := []string{"a", "b", "c"}
	resp := NewPageResponse(data, 30, 2, 10)
	if resp.Total != 30 {
		t.Errorf("expected Total=30, got %d", resp.Total)
	}
	if resp.Page != 2 {
		t.Errorf("expected Page=2, got %d", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("expected PageSize=10, got %d", resp.PageSize)
	}
	if len(resp.List) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.List))
	}
}
