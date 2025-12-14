package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option func(*pgxpool.Config)

func WithMinConns(minConns int32) Option {
	return func(config *pgxpool.Config) {
		config.MinConns = minConns
	}
}

func WithMaxConns(maxConns int32) Option {
	return func(config *pgxpool.Config) {
		config.MaxConns = maxConns
	}
}

func WithQueryTrace(qt pgx.QueryTracer) Option {
	return func(config *pgxpool.Config) {
		config.ConnConfig.Tracer = qt
	}
}
