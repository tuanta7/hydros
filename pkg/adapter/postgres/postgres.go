package postgres

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type client struct {
	pool       *pgxpool.Pool
	sqlBuilder squirrel.StatementBuilderType
}

func (c *client) BeginTx(ctx context.Context, o pgx.TxOptions) (context.Context, error) {
	tx, err := c.pool.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, "tx_connection", tx), nil
}

func (c *client) QueryProvider(ctx context.Context) QueryProvider {
	txConn := ctx.Value("tx_connection")
	if txConn != nil {
		if tx, ok := txConn.(*pgxpool.Tx); ok {
			return tx
		}
	}

	return c.pool
}

func (c *client) SQLBuilder() squirrel.StatementBuilderType {
	return c.sqlBuilder
}

func (c *client) Close() {
	c.pool.Close()
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

func NewClientFromPool(pool *pgxpool.Pool) (Client, error) {
	if pool == nil {
		return nil, errors.New("pool must not be nil")
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
