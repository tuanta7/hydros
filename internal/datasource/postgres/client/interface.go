package client

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/domain"
)

type Repository interface {
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
}
