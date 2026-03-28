package transaction

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionManager 事务管理器
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Transaction 在事务中执行函数
func (tm *TransactionManager) Transaction(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx, tx)
	})
}

// TransactionWithOptions 在事务中执行函数（带选项）
func (tm *TransactionManager) TransactionWithOptions(
	ctx context.Context,
	opts *gorm.Session,
	fn func(ctx context.Context, tx *gorm.DB) error,
) error {
	return tm.db.WithContext(ctx).Session(opts).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx, tx)
	})
}

// Begin 手动开始事务
func (tm *TransactionManager) Begin(ctx context.Context) (*gorm.DB, error) {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, tx.Error
	}
	return tx, nil
}

// Commit 提交事务
func (tm *TransactionManager) Commit(tx *gorm.DB) error {
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}
	return nil
}

// Rollback 回滚事务
func (tm *TransactionManager) Rollback(tx *gorm.DB) error {
	if err := tx.Rollback().Error; err != nil {
		logger.Error("Failed to rollback transaction", zap.Error(err))
		return err
	}
	return nil
}

// ExecuteInNewTransaction 在新事务中执行
func (tm *TransactionManager) ExecuteInNewTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	return tm.Transaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		return fn(ctx)
	})
}

// ExecuteWithNestedTransaction 支持嵌套事务（使用保存点）
func (tm *TransactionManager) ExecuteWithNestedTransaction(
	ctx context.Context,
	fn func(ctx context.Context, savePoint func(name string) error, rollbackTo func(name string) error) error,
) error {
	return tm.Transaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		savePointCounter := 0
		savePoint := func(name string) error {
			if name == "" {
				savePointCounter++
				name = fmt.Sprintf("sp_%d", savePointCounter)
			}
			return tx.Exec(fmt.Sprintf("SAVEPOINT %s", name)).Error
		}

		rollbackTo := func(name string) error {
			return tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)).Error
		}

		return fn(ctx, savePoint, rollbackTo)
	})
}

type txKey struct{}

// GetTxFromContext 从上下文中获取事务
func GetTxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok
}

// GetDB 获取数据库连接（优先使用上下文中的事务）
func (tm *TransactionManager) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := GetTxFromContext(ctx); ok {
		return tx
	}
	return tm.db.WithContext(ctx)
}

// TransactionCallback 事务回调接口
type TransactionCallback interface {
	BeforeCommit(ctx context.Context) error
	AfterCommit(ctx context.Context)
	BeforeRollback(ctx context.Context)
	AfterRollback(ctx context.Context, err error)
}

// TransactionWithCallback 带回调的事务执行
func (tm *TransactionManager) TransactionWithCallback(
	ctx context.Context,
	callback TransactionCallback,
	fn func(ctx context.Context, tx *gorm.DB) error,
) error {
	var err error

	txErr := tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)

		// 执行前回调
		if callback != nil {
			if cbErr := callback.BeforeCommit(txCtx); cbErr != nil {
				return cbErr
			}
		}

		// 执行事务函数
		err = fn(txCtx, tx)
		return err
	})

	if callback != nil {
		if txErr != nil {
			callback.BeforeRollback(ctx)
			callback.AfterRollback(ctx, txErr)
		} else {
			callback.AfterCommit(ctx)
		}
	}

	return txErr
}

// TransactionContext 事务上下文
type TransactionContext struct {
	tx     *gorm.DB
	active bool
}

// NewTransactionContext 创建事务上下文
func NewTransactionContext(db *gorm.DB) (*TransactionContext, error) {
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &TransactionContext{
		tx:     tx,
		active: true,
	}, nil
}

// Tx 获取事务
func (tc *TransactionContext) Tx() *gorm.DB {
	return tc.tx
}

// Active 是否已激活
func (tc *TransactionContext) Active() bool {
	return tc.active
}

// Commit 提交
func (tc *TransactionContext) Commit() error {
	if !tc.active {
		return fmt.Errorf("transaction not active")
	}
	if err := tc.tx.Commit().Error; err != nil {
		return err
	}
	tc.active = false
	return nil
}

// Rollback 回滚
func (tc *TransactionContext) Rollback() error {
	if !tc.active {
		return nil
	}
	if err := tc.tx.Rollback().Error; err != nil {
		return err
	}
	tc.active = false
	return nil
}
