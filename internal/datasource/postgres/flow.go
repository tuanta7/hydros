package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
)

type FlowRepository struct {
	table    string
	pgClient postgres.Client
}

func NewFlowRepository(pgc postgres.Client) *FlowRepository {
	return &FlowRepository{
		table:    "flow",
		pgClient: pgc,
	}
}

// Create is used to create a new login request flow
func (r *FlowRepository) Create(ctx context.Context, flow *domain.Flow) error {
	data := flow.ColumnMap()
	var columns []string
	var values []any

	for k, v := range data {
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

func (r *FlowRepository) Get(ctx context.Context, challenge string) (*domain.Flow, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Where(squirrel.Eq{"login_challenge": challenge}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flow, err := pgx.CollectOneRow(rows, toObject[domain.Flow])
	if err != nil {
		return nil, err
	}

	return flow, nil
}
