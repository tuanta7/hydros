package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TransactionManager struct {
	client Client
}

func NewTransactionManager(pgc Client) *TransactionManager {
	return &TransactionManager{
		client: pgc,
	}
}

type Statement func(ctx context.Context, tx pgx.Tx) error

func (tm *TransactionManager) WithTransaction(ctx context.Context, isoLevel pgx.TxIsoLevel, txFunc Statement) error {
	tx, err := tm.client.Pool().BeginTx(ctx, pgx.TxOptions{
		IsoLevel: isoLevel,
	})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	return txFunc(ctx, tx)
}
