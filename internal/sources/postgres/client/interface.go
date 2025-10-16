package client

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/domain"
	"github.com/tuanta7/oauth-server/internal/sources/postgres/transaction"
)

type Repository interface {
	WithTx(tx transaction.Transaction) Repository
	List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error)
}
