package postgres

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
)

type JWKRepository struct {
	table    string
	pgClient postgres.Client
}

func NewKeyRepository(pgc postgres.Client) *JWKRepository {
	return &JWKRepository{
		table:    "jwk",
		pgClient: pgc,
	}
}

func (r *JWKRepository) List(ctx context.Context, limit uint64) ([]*domain.KeyData, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys, err := pgx.CollectRows(rows, toObject[domain.KeyData])
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *JWKRepository) Create(ctx context.Context, key *domain.KeyData) error {
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

	_, err = r.pgClient.Pool().Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *JWKRepository) GetActiveKey(ctx context.Context, set domain.Set) (*domain.KeyData, error) {
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

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	key, err := pgx.CollectOneRow(rows, toObject[domain.KeyData])
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (r *JWKRepository) GetInactiveVerificationKey(ctx context.Context, set domain.Set, kid string) (*domain.KeyData, error) {
	return nil, errors.New("using inactive key to verify token is not supported yet")
}
