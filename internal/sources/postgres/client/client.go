package client

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/oauth-server/internal/domain"
	"github.com/tuanta7/oauth-server/internal/sources/postgres/transaction"
	"github.com/tuanta7/oauth-server/pkg/adapters/postgres"
)

type repository struct {
	tx       transaction.Transaction
	pgClient postgres.Client
}

func NewRepository(pgc postgres.Client) Repository {
	return &repository{
		pgClient: pgc,
	}
}

func (s *repository) WithTx(tx transaction.Transaction) Repository {
	s.tx = tx
	return s
}

func (s *repository) List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error) {
	query, args, err := s.pgClient.SQLBuilder().
		Select("*").
		From("").
		Offset(pageSize * (page - 1)).
		Limit(pageSize).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Client, error) {
		c, err := pgx.RowToStructByName[domain.Client](row)
		if err != nil {
			return nil, err
		}

		return &c, nil
	})

	return clients, nil
}

func (s *repository) Create(ctx context.Context, client *domain.Client) error {
	return nil
}
