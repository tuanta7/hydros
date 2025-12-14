package flow

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/pkg/postgres"
)

type Repository struct {
	table    string
	pgClient postgres.Client
}

func NewFlowRepository(pgc postgres.Client) *Repository {
	return &Repository{
		table:    "flow",
		pgClient: pgc,
	}
}

// Create is used to create a new login request flow
func (r *Repository) Create(ctx context.Context, flow *Flow) error {
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

	_, err = r.pgClient.QueryProvider(ctx).Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetGrantedAndRememberedConsent(ctx context.Context, client, subject string) (*Flow, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Where(
			squirrel.And{
				squirrel.Eq{"subject": subject},
				squirrel.Eq{"client_id": client},
				squirrel.Or{
					squirrel.Eq{"state": StateConsentGranted},
					squirrel.Eq{"state": StateConsentHandled},
				},
				squirrel.Eq{"consent_remember": true},
				squirrel.Eq{"consent_skip": false},
				squirrel.Eq{"consent_error": []byte("{}")},
			},
		).
		OrderBy("requested_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.QueryProvider(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flow, err := pgx.CollectOneRow(rows, postgres.ToObject[Flow])
	if err != nil {
		return nil, err
	}

	return flow, nil
}
