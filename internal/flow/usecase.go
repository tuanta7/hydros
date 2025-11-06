package flow

import (
	"context"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	flowRepo *Repository
	aead     aead.Cipher
	logger   *zapx.Logger
}

func NewUseCase(flowRepo *Repository, aead aead.Cipher, logger *zapx.Logger) *UseCase {
	return &UseCase{
		flowRepo: flowRepo,
		aead:     aead,
		logger:   logger,
	}
}

func (u *UseCase) EncodeFlow(ctx context.Context, f *Flow, a AdditionalData) (string, error) {
	return EncodeFlow(ctx, u.aead, f, a)
}

func (u *UseCase) DecodeFlow(ctx context.Context, encoded string, a AdditionalData) (*Flow, error) {
	return DecodeFlow(ctx, u.aead, encoded, a)
}

func (u *UseCase) GetLoginRequest(ctx context.Context, challenge string) (*Flow, error) {
	f, err := DecodeFlow(ctx, u.aead, challenge, AsLoginChallenge)
	if err != nil {
		return nil, err
	}

	// TODO: add login/consent timeout
	if f.RequestedAt.Add(time.Minute * 15).Before(x.NowUTC()) {
		return nil, core.ErrRequestUnauthorized.WithHint("The login request has expired, please try again.")
	}

	return f, nil
}

func (u *UseCase) VerifyAndInvalidateFlow(ctx context.Context, verifier string, as []byte) (*Flow, error) {
	f, err := DecodeFlow(ctx, u.aead, verifier, as)
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
