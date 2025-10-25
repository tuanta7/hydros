package client

import (
	"context"

	"github.com/tuanta7/hydros/internal/domain"
)

type UseCase struct {
	clientRepo Repository
}

func NewUseCase(ur Repository) *UseCase {
	return &UseCase{
		clientRepo: ur,
	}
}

func (u *UseCase) List(ctx context.Context, page, pageSize uint64) ([]*domain.Client, error) {
	return u.clientRepo.List(ctx, page, pageSize)
}

func (u *UseCase) Get(ctx context.Context, id string) (*domain.Client, error) {
	return nil, nil
}
