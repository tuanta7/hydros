package session

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
)

type Repository struct {
	table    string
	pgClient postgres.Client
}

func NewSessionRepository(pgc postgres.Client) *Repository {
	return &Repository{
		table:    "login_session",
		pgClient: pgc,
	}
}

func (r *Repository) GetRememberedLoginSession(ctx context.Context, id string) (*LoginSession, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(r.table).
		Where(squirrel.And{
			squirrel.Eq{"id": id},
			squirrel.Eq{"remember": true},
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	session, err := pgx.CollectOneRow(rows, postgres.ToObject[LoginSession])
	if err != nil {
		return nil, err
	}

	return session, nil
}
