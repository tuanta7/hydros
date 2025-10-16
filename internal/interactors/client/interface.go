package client

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/domain"
	clientrepo "github.com/tuanta7/oauth-server/internal/sources/postgres/client"
)

var (
	_ Repository = (clientrepo.Repository)(nil)
)

type Repository interface {
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
}
