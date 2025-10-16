package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type client struct {
	pool       Pool
	sqlBuilder squirrel.StatementBuilderType
}

func (c *client) Pool() Pool {
	return c.pool
}

func (c *client) SQLBuilder() squirrel.StatementBuilderType {
	return c.sqlBuilder
}

func NewClient(dsn string, options ...Option) (Client, error) {
	pool, err := newPool(dsn, options...)
	if err != nil {
		return nil, err
	}

	return &client{
		pool:       pool,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func newPool(dsn string, options ...Option) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(dbConfig)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
