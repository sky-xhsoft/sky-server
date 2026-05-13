package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(10001, "test error")
	if err.Code != 10001 {
		t.Errorf("Code = %d, want 10001", err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("Message = %s, want test error", err.Message)
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("original")
	err := Wrap(40001, "db error", cause)
	if err.Code != 40001 {
		t.Errorf("Code = %d, want 40001", err.Code)
	}
	if err.Err != cause {
		t.Errorf("Wrapped error = %v, want %v", err.Err, cause)
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("original")
	err := Wrap(40001, "db error", cause)
	if errors.Unwrap(err) != cause {
		t.Error("Unwrap should return the original error")
	}
}

func TestIs(t *testing.T) {
	err1 := New(10001, "test")
	err2 := New(10001, "another")
	err3 := New(20001, "auth")

	if !Is(err1, err2) {
		t.Error("Same code should match")
	}
	if Is(err1, err3) {
		t.Error("Different codes should not match")
	}
}

func TestGetCode(t *testing.T) {
	err := New(10001, "test")
	if GetCode(err) != 10001 {
		t.Errorf("GetCode = %d, want 10001", GetCode(err))
	}
	if GetCode(nil) != 0 {
		t.Error("GetCode(nil) should be 0")
	}
	plain := errors.New("plain error")
	if GetCode(plain) != ErrInternal {
		t.Errorf("GetCode(plain) = %d, want ErrInternal", GetCode(plain))
	}
}

func TestAs(t *testing.T) {
	err := New(10001, "test")
	var target *AppError
	if !As(err, &target) {
		t.Error("As should return true for AppError")
	}
	if target.Code != 10001 {
		t.Errorf("target.Code = %d, want 10001", target.Code)
	}

	plain := errors.New("plain")
	if As(plain, &target) {
		t.Error("As should return false for plain error")
	}
}

func TestAppError_Error(t *testing.T) {
	err := New(10001, "test")
	s := err.Error()
	if s != "[10001] test" {
		t.Errorf("Error() = %s, want [10001] test", s)
	}

	err2 := Wrap(40001, "db error", errors.New("sql failed"))
	s2 := err2.Error()
	if s2 != "[40001] db error: sql failed" {
		t.Errorf("Error() = %s, want [40001] db error: sql failed", s2)
	}
}
