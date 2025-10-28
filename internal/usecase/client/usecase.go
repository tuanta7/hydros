package client

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	clientRepo Repository
	logger     *zapx.Logger
}

func NewUseCase(clientRepo Repository, logger *zapx.Logger) *UseCase {
	return &UseCase{
		clientRepo: clientRepo,
		logger:     logger,
	}
}

func (u *UseCase) ListClients(ctx context.Context, page, pageSize uint64) ([]core.Client, error) {
	clients, err := u.clientRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	var result []core.Client
	for _, client := range clients {
		result = append(result, client)
	}

	return result, nil
}

func (u *UseCase) GetClient(ctx context.Context, id string) (core.Client, error) {
	client, err := u.clientRepo.Get(ctx, id)
	if err != nil {
		u.logger.Error("cannot get client",
			zap.Error(err),
			zap.String("method", "clientRepo.Get"),
		)
		return nil, err
	}

	return client, nil
}
