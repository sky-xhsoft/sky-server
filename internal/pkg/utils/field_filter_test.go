package utils

import (
	"reflect"
	"testing"
)

func TestFilterFields_EmptyFields(t *testing.T) {
	obj := map[string]interface{}{"id": 1, "name": "test"}
	result := FilterFields(obj, "")
	m, ok := result.(map[string]interface{})
	if !ok || m["id"] != 1 {
		t.Error("expected original object when fields is empty")
	}
}

func TestFilterFields_MapFiltering(t *testing.T) {
	obj := map[string]interface{}{"id": 1, "name": "test", "email": "test@example.com"}
	result := FilterFields(obj, "id,name")
	m := result.(map[string]interface{})
	if _, ok := m["email"]; ok {
		t.Error("email should be filtered out")
	}
	if m["id"] != 1 || m["name"] != "test" {
		t.Error("id and name should be preserved")
	}
}

func TestFilterFields_SliceFiltering(t *testing.T) {
	obj := []map[string]interface{}{
		{"id": 1, "name": "a", "extra": "x"},
		{"id": 2, "name": "b", "extra": "y"},
	}
	result := FilterFields(obj, "id,name")
	slice := result.([]map[string]interface{})
	if len(slice) != 2 {
		t.Error("slice length should be preserved")
	}
	if _, ok := slice[0]["extra"]; ok {
		t.Error("extra should be filtered out")
	}
}

func TestFilterFields_StructFiltering(t *testing.T) {
	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	user := User{ID: 1, Name: "test", Email: "test@example.com"}
	result := FilterFields(user, "id,name")
	m := result.(map[string]interface{})
	if _, ok := m["email"]; ok {
		t.Error("email should be filtered out of struct")
	}
	if !reflect.DeepEqual(m["id"], 1) {
		t.Error("id should be preserved")
	}
}

func TestParseFields(t *testing.T) {
	result := parseFields("id, name, email ")
	if !result["id"] || !result["name"] || !result["email"] {
		t.Error("parseFields should handle spaces")
	}
}

func TestSelectFields(t *testing.T) {
	result := SelectFields([]string{"id", "name"})
	if len(result) != 2 || !result["id"] || !result["name"] {
		t.Error("SelectFields should convert slice to map")
	}
}
