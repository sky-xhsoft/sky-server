package transaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Executor 事务执行器接口
type Executor interface {
	// Execute 在事务中执行函数
	Execute(ctx context.Context, fn func(tx *gorm.DB) error) error

	// ExecuteWithResult 在事务中执行函数并返回结果
	ExecuteWithResult(ctx context.Context, fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error)
}

// executor 事务执行器实现
type executor struct {
	db *gorm.DB
}

// NewExecutor 创建事务执行器
func NewExecutor(db *gorm.DB) Executor {
	return &executor{
		db: db,
	}
}

// Execute 在事务中执行函数
// 如果函数返回错误，自动回滚；否则自动提交
func (e *executor) Execute(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// ExecuteWithResult 在事务中执行函数并返回结果
// 如果函数返回错误，自动回滚；否则自动提交
func (e *executor) ExecuteWithResult(ctx context.Context, fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})
	return result, err
}

// WithTransaction 包装函数，使其支持在事务或非事务中执行
// 如果传入的 db 已经在事务中，直接使用；否则开启新事务
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	// 检查是否已在事务中
	if db.Statement != nil && db.Statement.DB != nil {
		// 已在事务中，直接执行
		return fn(db)
	}

	// 开启新事务
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// ExecuteInTransaction 确保函数在事务中执行
// 如果已在事务中，复用当前事务；否则创建新事务
func ExecuteInTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	// 从context检查是否已有事务
	if tx := GetTxFromContext(ctx); tx != nil {
		// 使用context中的事务
		return fn(tx)
	}

	// 开启新事务
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 将事务保存到context中（供嵌套调用使用）
		ctx = context.WithValue(ctx, txKey, tx)
		return fn(tx)
	})
}

// contextKey 用于在context中存储事务
type contextKey string

const txKey contextKey = "gorm:tx"

// GetTxFromContext 从context获取事务
func GetTxFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// WithTx 将事务放入context
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFunc 事务函数类型
type TxFunc func(tx *gorm.DB) error

// TxFuncWithResult 带返回值的事务函数类型
type TxFuncWithResult func(tx *gorm.DB) (interface{}, error)

// Rollback 回滚辅助函数
func Rollback(tx *gorm.DB, err error) error {
	if rbErr := tx.Rollback().Error; rbErr != nil {
		return fmt.Errorf("回滚失败: %v, 原始错误: %w", rbErr, err)
	}
	return err
}

// SafeCommit 安全提交事务
func SafeCommit(tx *gorm.DB) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic时回滚
			tx.Rollback()
			err = fmt.Errorf("事务提交panic: %v", r)
		}
	}()

	return tx.Commit().Error
}

// RunInTransaction 在事务中运行函数（简化版）
func RunInTransaction(db *gorm.DB, fn TxFunc) error {
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // 重新抛出panic
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
