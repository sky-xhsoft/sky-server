package crud

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/domain/search"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Exec("CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT, price REAL, quantity INTEGER, created_at DATETIME)").Error; err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return db
}

func TestBuildAdvancedSearch_NilRequest(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	query, err := s.buildAdvancedSearch(db.Model(&struct{}{}), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if query == nil {
		t.Error("expected non-nil query")
	}
}

func TestBuildAdvancedSearch_NoFilters(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{}
	query, err := s.buildAdvancedSearch(db.Model(&struct{}{}), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if query == nil {
		t.Error("expected non-nil query")
	}
}

func TestBuildAdvancedSearch_Eq(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "name", Operator: search.OpEq, Value: "test"},
		},
	}

	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]struct{}{})
	})
	if !strings.Contains(sql, "`name` =") {
		t.Errorf("SQL should contain '`name` =': %s", sql)
	}
}

func TestBuildAdvancedSearch_Ne(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "price", Operator: search.OpNe, Value: 0},
		},
	}

	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "`price` !=") {
		t.Errorf("SQL should contain '`price` !=': %s", sql)
	}
}

func TestBuildAdvancedSearch_Gt(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "price", Operator: search.OpGt, Value: 100},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "`price` >") {
		t.Errorf("SQL should contain '`price` >': %s", sql)
	}
}

func TestBuildAdvancedSearch_Gte(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "price", Operator: search.OpGte, Value: 100},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "`price` >=") {
		t.Errorf("SQL should contain '`price` >=': %s", sql)
	}
}

func TestBuildAdvancedSearch_Lt(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "quantity", Operator: search.OpLt, Value: 50},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "`quantity` <") {
		t.Errorf("SQL should contain '`quantity` <': %s", sql)
	}
}

func TestBuildAdvancedSearch_Lte(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "quantity", Operator: search.OpLte, Value: 50},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "`quantity` <=") {
		t.Errorf("SQL should contain '`quantity` <=': %s", sql)
	}
}

func TestBuildAdvancedSearch_Contains(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "name", Operator: search.OpContains, Value: "phone"},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "LIKE") {
		t.Errorf("SQL should contain LIKE: %s", sql)
	}
}

func TestBuildAdvancedSearch_StartsWith(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "name", Operator: search.OpStartsWith, Value: "A"},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "LIKE") {
		t.Errorf("SQL should contain LIKE: %s", sql)
	}
}

func TestBuildAdvancedSearch_EndsWith(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "name", Operator: search.OpEndsWith, Value: "Z"},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "LIKE") {
		t.Errorf("SQL should contain LIKE: %s", sql)
	}
}

func TestBuildAdvancedSearch_In(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "id", Operator: search.OpIn, Values: []interface{}{1, 2, 3}},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "IN") {
		t.Errorf("SQL should contain IN: %s", sql)
	}
}

func TestBuildAdvancedSearch_InEmpty(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "id", Operator: search.OpIn, Values: []interface{}{}},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if strings.Contains(sql, "IN") {
		t.Errorf("empty IN should be skipped: %s", sql)
	}
}

func TestBuildAdvancedSearch_Between(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "price", Operator: search.OpBetween, ValueFrom: 10, ValueTo: 100},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "BETWEEN") {
		t.Errorf("SQL should contain BETWEEN: %s", sql)
	}
}

func TestBuildAdvancedSearch_IsNull(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "created_at", Operator: search.OpIsNull},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "IS NULL") {
		t.Errorf("SQL should contain IS NULL: %s", sql)
	}
}

func TestBuildAdvancedSearch_IsNotNull(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "name", Operator: search.OpIsNotNull},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "IS NOT NULL") {
		t.Errorf("SQL should contain IS NOT NULL: %s", sql)
	}
}

func TestBuildAdvancedSearch_MultipleFilters(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "price", Operator: search.OpGte, Value: 10},
			{Field: "price", Operator: search.OpLte, Value: 100},
			{Field: "name", Operator: search.OpContains, Value: "test"},
		},
	}
	query, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Find(&[]struct{}{}) })
	if !strings.Contains(sql, "`price` >=") || !strings.Contains(sql, "`price` <=") || !strings.Contains(sql, "LIKE") {
		t.Errorf("SQL should contain all three conditions: %s", sql)
	}
}

func TestBuildAdvancedSearch_InvalidField(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "1; DROP TABLE", Operator: search.OpEq, Value: "x"},
		},
	}
	_, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err == nil {
		t.Error("expected error for invalid field name")
	}
}

func TestBuildAdvancedSearch_InvalidOperator(t *testing.T) {
	db := setupTestDB(t)
	s := &service{db: db}

	req := &search.SearchRequest{
		Filters: []*search.SearchFilter{
			{Field: "name", Operator: "invalid_op", Value: "x"},
		},
	}
	_, err := s.buildAdvancedSearch(db.Table("test_items"), req)
	if err == nil {
		t.Error("expected error for unsupported operator")
	}
}

func TestToString(t *testing.T) {
	if toString(nil) != "" {
		t.Error("toString(nil) should return empty string")
	}
	if toString("hello") != "hello" {
		t.Error("toString(string) should return the string")
	}
	if toString(123) != "123" {
		t.Error("toString(int) should return string representation")
	}
}

func TestSearchRequest_Validate(t *testing.T) {
	req := &search.SearchRequest{Page: 0, PageSize: 0}
	req.Validate()
	if req.Page != 1 {
		t.Errorf("Page should default to 1, got %d", req.Page)
	}
	if req.PageSize != 10 {
		t.Errorf("PageSize should default to 10, got %d", req.PageSize)
	}

	req = &search.SearchRequest{Page: 5, PageSize: 200}
	req.Validate()
	if req.PageSize != 100 {
		t.Errorf("PageSize should cap at 100, got %d", req.PageSize)
	}
}

func TestSearchRequest_HasFilters(t *testing.T) {
	req := &search.SearchRequest{}
	if req.HasFilters() {
		t.Error("empty request should have no filters")
	}

	req.Filters = []*search.SearchFilter{{Field: "name", Operator: search.OpEq, Value: "x"}}
	if !req.HasFilters() {
		t.Error("request with filters should return true")
	}
}
