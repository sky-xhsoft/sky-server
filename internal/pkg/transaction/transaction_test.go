package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTxDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Exec("CREATE TABLE test_tx (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return db
}

func TestNewExecutor(t *testing.T) {
	db := setupTxDB(t)
	exec := NewExecutor(db)
	if exec == nil {
		t.Fatal("expected non-nil executor")
	}
}

func TestExecutor_Execute_Success(t *testing.T) {
	db := setupTxDB(t)
	exec := NewExecutor(db)

	err := exec.Execute(context.Background(), func(tx *gorm.DB) error {
		return tx.Exec("INSERT INTO test_tx (id, name) VALUES (1, 'test')").Error
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	db.Table("test_tx").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestExecutor_Execute_Rollback(t *testing.T) {
	db := setupTxDB(t)
	exec := NewExecutor(db)

	wantErr := errors.New("rollback test")
	err := exec.Execute(context.Background(), func(tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO test_tx (id, name) VALUES (1, 'test')").Error; err != nil {
			return err
		}
		return wantErr
	})
	if err != wantErr {
		t.Fatalf("expected rollback error, got: %v", err)
	}

	var count int64
	db.Table("test_tx").Count(&count)
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

func TestExecutor_ExecuteWithResult(t *testing.T) {
	db := setupTxDB(t)
	exec := NewExecutor(db)

	result, err := exec.ExecuteWithResult(context.Background(), func(tx *gorm.DB) (interface{}, error) {
		if err := tx.Exec("INSERT INTO test_tx (id, name) VALUES (2, 'result')").Error; err != nil {
			return nil, err
		}
		return "success", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
}

func TestRunInTransaction_Success(t *testing.T) {
	db := setupTxDB(t)

	err := RunInTransaction(db, func(tx *gorm.DB) error {
		return tx.Exec("INSERT INTO test_tx (id, name) VALUES (3, 'tx')").Error
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInTransaction_Error(t *testing.T) {
	db := setupTxDB(t)

	err := RunInTransaction(db, func(tx *gorm.DB) error {
		return errors.New("tx failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithTx_GetFromContext(t *testing.T) {
	db := setupTxDB(t)
	ctx := context.Background()
	ctx = WithTx(ctx, db)

	tx := GetTxFromContext(ctx)
	if tx != db {
		t.Error("expected same db instance from context")
	}
}

func TestGetTxFromContext_NoTx(t *testing.T) {
	ctx := context.Background()
	tx := GetTxFromContext(ctx)
	if tx != nil {
		t.Error("expected nil when no tx in context")
	}
}

func TestWithTransaction(t *testing.T) {
	db := setupTxDB(t)

	err := WithTransaction(db, func(tx *gorm.DB) error {
		return tx.Exec("INSERT INTO test_tx (id, name) VALUES (4, 'wtx')").Error
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	db.Table("test_tx").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestExecutor_NilDB(t *testing.T) {
	exec := NewExecutor(nil)
	if exec == nil {
		t.Fatal("executor should not be nil even with nil db")
	}
}
