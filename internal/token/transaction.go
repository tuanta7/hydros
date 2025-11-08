package token

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (r *RequestSessionRepo) BeginTX(ctx context.Context) (context.Context, error) {
	return r.pgClient.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
}

func (r *RequestSessionRepo) Commit(ctx context.Context) error {
	q := r.pgClient.QueryProvider(ctx)
	tx, ok := q.(*pgxpool.Tx)
	if !ok {
		return errors.New("no transaction found")
	}

	return tx.Commit(ctx)
}

func (r *RequestSessionRepo) Rollback(ctx context.Context) error {
	q := r.pgClient.QueryProvider(ctx)
	tx, ok := q.(*pgxpool.Tx)
	if !ok {
		return errors.New("no transaction found")
	}

	return tx.Rollback(ctx)
}
