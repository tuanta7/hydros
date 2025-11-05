package flow

import (
	"context"

	"github.com/tuanta7/hydros/internal/datasource/postgres"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	flowRepo *postgres.FlowRepository
	logger   *zapx.Logger
}

func NewUseCase(flowRepo *postgres.FlowRepository, logger *zapx.Logger) *UseCase {
	return &UseCase{
		flowRepo: flowRepo,
		logger:   logger,
	}
}

func (u *UseCase) CreateLoginRequest(ctx context.Context, flow *domain.Flow) error {
	err := u.flowRepo.Create(ctx, flow)
	if err != nil {
		u.logger.Error("cannot create flow",
			zap.Error(err),
			zap.String("method", "flowRepo.Create"),
		)
		return err
	}

	return nil
}

func (u *UseCase) GetLoginRequest(ctx context.Context, challenge string) (*domain.Flow, error) {
	flow, err := u.flowRepo.Get(ctx, challenge)
	if err != nil {
		u.logger.Error("cannot get flow",
			zap.Error(err),
			zap.String("method", "flowRepo.Get"),
		)
		return nil, err
	}

	return flow, nil
}
