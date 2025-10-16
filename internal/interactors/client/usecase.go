package client

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/domain"
	"github.com/tuanta7/oauth-server/internal/sources/postgres/transaction"
)

type UseCase struct {
	clientRepo Repository
	txManager  transaction.Manager
}

func NewUseCase(ur Repository, txm transaction.Manager) *UseCase {
	return &UseCase{
		clientRepo: ur,
		txManager:  txm,
	}
}

func (uc *UseCase) List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error) {
	return uc.clientRepo.List(ctx, page, pageSize)
}
