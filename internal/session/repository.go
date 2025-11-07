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

func (r *Repository) UpsertLoginSession(ctx context.Context, session *LoginSession) error {
	data := session.ColumnMap()
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
		Suffix(`
				ON CONFLICT (id) DO 
				UPDATE SET
					authenticated_at = EXCLUDED.authenticated_at, 
					subject = EXCLUDED.subject, 
					remember = EXCLUDED.remember, 
					identity_provider_session_id = EXCLUDED.identity_provider_session_id
		`).
		ToSql()
	if err != nil {
		return err
	}

	ct, err := r.pgClient.Pool().Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *Repository) DeleteLoginSession(ctx context.Context, id string) (*LoginSession, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Delete(r.table).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING *").
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
