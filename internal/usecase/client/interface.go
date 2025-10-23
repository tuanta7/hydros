package client

import (
	"context"

	clientrepo "github.com/tuanta7/hydros/internal/datasource/postgres/client"
	"github.com/tuanta7/hydros/internal/domain"
)

var (
	_ Repository = (clientrepo.Repository)(nil)
)

type Repository interface {
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
}
