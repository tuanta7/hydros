package jwk

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/pkg/postgres"
)

type Repository struct {
	table    string
	pgClient postgres.Client
}

func NewKeyRepository(pgc postgres.Client) *Repository {
	return &Repository{
		table:    "jwk",
		pgClient: pgc,
	}
}

func (r *Repository) List(ctx context.Context, limit uint64) ([]*KeyData, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.QueryProvider(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys, err := pgx.CollectRows(rows, postgres.ToObject[KeyData])
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *Repository) Create(ctx context.Context, key *KeyData) error {
	m := key.ColumnMap()
	var columns []string
	var values []any

	for k, v := range m {
		columns = append(columns, k)
		values = append(values, v)
	}

	query, args, err := r.pgClient.SQLBuilder().
		Insert(r.table).
		Columns(columns...).
		Values(values...).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pgClient.QueryProvider(ctx).Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetActiveKey(ctx context.Context, set Set) (*KeyData, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Where(
			squirrel.And{
				squirrel.Eq{"active": true},
				squirrel.Eq{"sid": set},
			},
		).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.QueryProvider(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	key, err := pgx.CollectOneRow(rows, postgres.ToObject[KeyData])
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (r *Repository) GetInactiveVerificationKey(ctx context.Context, set Set, kid string) (*KeyData, error) {
	return nil, errors.New("using inactive key to verify token is not supported yet")
}
