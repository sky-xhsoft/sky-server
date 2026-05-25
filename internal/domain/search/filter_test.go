package search

import (
	"testing"
)

func TestSearchRequest_Validate_Defaults(t *testing.T) {
	req := &SearchRequest{Page: 0, PageSize: 0}
	req.Validate()
	if req.Page != 1 {
		t.Errorf("expected Page=1, got %d", req.Page)
	}
	if req.PageSize != 10 {
		t.Errorf("expected PageSize=10, got %d", req.PageSize)
	}
	if req.Order != "desc" {
		t.Errorf("expected Order='desc', got '%s'", req.Order)
	}
}

func TestSearchRequest_Validate_Caps(t *testing.T) {
	req := &SearchRequest{Page: 2, PageSize: 200}
	req.Validate()
	if req.PageSize != 100 {
		t.Errorf("expected PageSize capped at 100, got %d", req.PageSize)
	}
}

func TestSearchRequest_Validate_PreservesOrder(t *testing.T) {
	req := &SearchRequest{Page: 3, PageSize: 20, Order: "asc"}
	req.Validate()
	if req.Order != "asc" {
		t.Errorf("expected Order='asc', got '%s'", req.Order)
	}
}

func TestSearchRequest_HasFilters_Empty(t *testing.T) {
	req := &SearchRequest{}
	if req.HasFilters() {
		t.Error("empty request should have no filters")
	}
}

func TestSearchRequest_HasFilters_NonEmpty(t *testing.T) {
	req := &SearchRequest{
		Filters: []*SearchFilter{
			{Field: "name", Operator: OpEq, Value: "test"},
		},
	}
	if !req.HasFilters() {
		t.Error("request with filters should return true")
	}
}

func TestSearchFilter_AllOperators(t *testing.T) {
	operators := []Operator{OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpContains, OpStartsWith, OpEndsWith, OpIn, OpBetween, OpIsNull, OpIsNotNull}
	if len(operators) != 13 {
		t.Errorf("expected 13 operators, got %d", len(operators))
	}
}

func TestSearchFilter_Struct(t *testing.T) {
	f := &SearchFilter{
		Field:     "price",
		Operator:  OpBetween,
		ValueFrom: 10.0,
		ValueTo:   100.0,
	}
	if f.Field != "price" {
		t.Errorf("expected Field='price', got '%s'", f.Field)
	}
	if f.Operator != OpBetween {
		t.Errorf("expected OpBetween, got '%s'", f.Operator)
	}
}
