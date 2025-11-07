package flow

import (
	"context"

	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/zapx"
)

type UseCase struct {
	cfg      *config.Config
	flowRepo *Repository
	aead     aead.Cipher
	logger   *zapx.Logger
}

func NewUseCase(cfg *config.Config, flowRepo *Repository, aead aead.Cipher, logger *zapx.Logger) *UseCase {
	return &UseCase{
		cfg:      cfg,
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

	if f.RequestedAt.Add(u.cfg.GetConsentRequestMaxAge()).Before(x.NowUTC()) {
		return nil, core.ErrRequestUnauthorized.WithHint("The login request has expired, please try again.")
	}

	return f, nil
}

func (u *UseCase) VerifyAndInvalidateLoginRequest(ctx context.Context, verifier string) (*Flow, error) {
	f, err := DecodeFlow(ctx, u.aead, verifier, AsLoginVerifier)
	if err != nil {
		return nil, err
	}

	err = f.InvalidateLoginRequest()
	if err != nil {
		return nil, core.ErrInvalidRequest.WithDebug(err.Error())
	}

	return f, nil
}
