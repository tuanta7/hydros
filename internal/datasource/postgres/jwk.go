package postgres

import "github.com/tuanta7/hydros/pkg/adapter/postgres"

type JWKRepository struct {
	table    string
	pgClient postgres.Client
}

func NewKeyRepository(pgc postgres.Client) JWKRepository {
	return JWKRepository{
		table:    "jwk",
		pgClient: pgc,
	}
}
