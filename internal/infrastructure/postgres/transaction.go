package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// WithTx stores a GORM transaction in the context.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext retrieves the transaction from context, falling back to the provided db.
func TxFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return db
}

// RunInTx executes fn inside a database transaction.
// The transaction is committed if fn returns nil, rolled back otherwise.
func RunInTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := WithTx(ctx, tx)
		return fn(txCtx)
	})
}
