package flow

import (
	"context"
	"database/sql"
	stderr "errors"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	cfg      *config.Config
	flowRepo *Repository
	aead     aead.Cipher
	logger   *zapx.ZapLogger
}

func NewUseCase(cfg *config.Config, flowRepo *Repository, aead aead.Cipher, logger *zapx.ZapLogger) *UseCase {
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

func (u *UseCase) GetConsentRequest(ctx context.Context, challenge string) (*Flow, error) {
	f, err := DecodeFlow(ctx, u.aead, challenge, AsConsentChallenge)
	if err != nil {
		return nil, err
	}

	if f.RequestedAt.Add(u.cfg.GetConsentRequestMaxAge()).Before(x.NowUTC()) {
		return nil, core.ErrRequestUnauthorized.WithHint("The consent request has expired, please try again.")
	}

	return f, nil
}

func (u *UseCase) FindGrantedAndRememberedConsentRequest(ctx context.Context, client, user string) (*Flow, error) {
	flow, err := u.flowRepo.GetGrantedAndRememberedConsent(ctx, client, user)
	if stderr.Is(err, sql.ErrNoRows) {
		return nil, errors.ErrNoPreviousConsentFound
	} else if err != nil {
		return nil, err
	}

	consentRememberFor := 0
	if flow.ConsentRememberFor != nil {
		consentRememberFor = *flow.ConsentRememberFor
	}

	if consentRememberFor > 0 && flow.RequestedAt.Add(time.Duration(consentRememberFor)*time.Second).Before(x.NowUTC()) {
		return nil, errors.ErrNoPreviousConsentFound
	}

	return flow, nil
}

func (u *UseCase) SaveFlow(ctx context.Context, f *Flow) error {
	err := u.flowRepo.Create(ctx, f)
	if err != nil {
		u.logger.Error("unable to create flow",
			zap.Error(err),
			zap.Any("flow", f),
			zap.String("method", "flowRepo.Create"),
		)
		return err
	}

	return nil
}
