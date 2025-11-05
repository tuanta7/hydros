package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
)

type SessionRepository struct {
	table    string
	pgClient postgres.Client
}

func NewSessionRepository(pgc postgres.Client) *SessionRepository {
	return &SessionRepository{
		table:    "login_session",
		pgClient: pgc,
	}
}

func (r *SessionRepository) GetRememberedLoginSession(ctx context.Context, id string) (*domain.LoginSession, error) {
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
	defer rows.Close()

	session, err := pgx.CollectOneRow(rows, toObject[domain.LoginSession])
	if err != nil {
		return nil, err
	}

	return session, nil
}
