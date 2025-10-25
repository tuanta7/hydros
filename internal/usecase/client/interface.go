package client

import (
	"context"

	"github.com/tuanta7/hydros/internal/datasource/postgres"
	"github.com/tuanta7/hydros/internal/domain"
)

var (
	_ Repository = (postgres.ClientRepository)(nil)
)

type Repository interface {
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
}
