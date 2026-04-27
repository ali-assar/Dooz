package tx

import (
	"context"

	"gorm.io/gorm"
)

type (
	TransactionFN func(ctx context.Context) error
	txKey         struct{}

	Transaction interface {
		Do(ctx context.Context, fn TransactionFN) error
		DB(ctx context.Context) *gorm.DB
	}
)

type tx struct {
	db *gorm.DB
}

func NewTransaction(db *gorm.DB) Transaction {
	return &tx{db: db}
}

func (m *tx) Do(ctx context.Context, fn TransactionFN) error {
	return m.db.WithContext(ctx).Transaction(func(t *gorm.DB) error {
		ct := context.WithValue(ctx, txKey{}, t)
		return fn(ct)
	})
}

func (m *tx) DB(ctx context.Context) *gorm.DB {
	t, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		return m.db.WithContext(ctx)
	}
	return t
}
