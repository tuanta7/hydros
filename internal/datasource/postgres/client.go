package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
)

type ClientRepository interface {
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
	Create(ctx context.Context, client *domain.Client) error
	Get(ctx context.Context, id string) (*domain.Client, error)
}

type clientRepository struct {
	table    string
	pgClient postgres.Client
}

func NewClientRepository(pgc postgres.Client) ClientRepository {
	return &clientRepository{
		table:    "client",
		pgClient: pgc,
	}
}

func (r *clientRepository) List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Offset(pageSize * (page - 1)).
		Limit(pageSize).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients, err := pgx.CollectRows(rows, toObject[domain.Client])
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *clientRepository) Create(ctx context.Context, client *domain.Client) error {
	m := client.ColumnMap()
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

func (r *clientRepository) Get(ctx context.Context, id string) (*domain.Client, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	client, err := pgx.CollectOneRow(rows, toObject[domain.Client])
	if err != nil {
		return nil, err
	}

	return client, nil
}
