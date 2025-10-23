package client

import (
	"context"

	"github.com/tuanta7/hydros/internal/domain"
)

type Repository interface {
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
}
