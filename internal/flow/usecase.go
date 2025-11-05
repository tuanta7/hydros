package flow

import (
	"context"

	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	flowRepo *Repository
	aead     aead.Cipher
	logger   *zapx.Logger
}

func NewUseCase(flowRepo *Repository, logger *zapx.Logger) *UseCase {
	return &UseCase{
		flowRepo: flowRepo,
		logger:   logger,
	}
}

func (u *UseCase) VerifyAndInvalidateConsentRequest(ctx context.Context, verifier string) (*Flow, error) {
	f, err := DecodeFlow(ctx, u.aead, verifier, []byte(""))
	if err != nil {
		return nil, err
	}

	err = u.flowRepo.Create(ctx, f)
	if err != nil {
		u.logger.Error("cannot create flow",
			zap.Error(err),
			zap.String("method", "flowRepo.Create"),
		)
		return nil, err
	}

	return f, nil
}
