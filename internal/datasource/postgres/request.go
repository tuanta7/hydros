package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
)

var tableName = map[core.TokenType]string{
	core.AccessToken:       "access_token",
	core.RefreshToken:      "refresh_token",
	core.IDToken:           "id_token",
	core.AuthorizationCode: "code",
	core.PKCECode:          "pkce",
}

type RequestSessionRepo struct {
	pgClient postgres.Client
}

func NewRequestSessionRepo(pgClient postgres.Client) *RequestSessionRepo {
	return &RequestSessionRepo{
		pgClient: pgClient,
	}
}

func (r *RequestSessionRepo) Create(ctx context.Context, tokenType core.TokenType, session *domain.RequestSessionData) error {
	data := session.ColumnMap()
	var columns []string
	var values []any

	for k, v := range data {
		columns = append(columns, k)
		values = append(values, v)
	}

	query, args, err := r.pgClient.SQLBuilder().
		Insert(tableName[tokenType]).
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

func (r *RequestSessionRepo) GetBySignature(ctx context.Context, tokenType core.TokenType, signature string) (*domain.RequestSessionData, error) {
	query, args, err := r.pgClient.SQLBuilder().
		Select("*").
		From(tableName[tokenType]).
		Where(squirrel.Eq{"signature": signature}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pgClient.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	session, err := pgx.CollectOneRow(rows, toObject[domain.RequestSessionData])
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *RequestSessionRepo) DeleteBySignature(ctx context.Context, tokenType core.TokenType, signature string) error {
	query, args, err := r.pgClient.SQLBuilder().
		Delete(tableName[tokenType]).
		Where(squirrel.Eq{"signature": signature}).
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
