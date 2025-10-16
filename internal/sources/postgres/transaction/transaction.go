package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuanta7/oauth-server/pkg/adapters/postgres"
)

type Transaction interface {
	Commit() error
	Rollback() error
}

type Manager struct {
	pool postgres.Pool
}

func (m *Manager) Do(fn func(tx Transaction) error) error {
	ctx := context.Background()
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	t := &Transaction{tx: tx}
	if err := fn(t); err != nil {
		_ = t.Rollback()
		return err
	}
	return t.Commit()
}
