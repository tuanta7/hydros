package client

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/domain"
)

type UseCase struct {
	clientRepo Repository
}

func NewUseCase(ur Repository) *UseCase {
	return &UseCase{
		clientRepo: ur,
	}
}

func (uc *UseCase) List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error) {
	return uc.clientRepo.List(ctx, page, pageSize)
}
