package jwk

import (
	"context"

	"github.com/go-jose/go-jose/v4"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/internal/datasource/postgres"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/zapx"
)

type UseCase struct {
	cfg     *config.Config
	aead    aead.Cipher
	jwkRepo *postgres.JWKRepository
	logger  *zapx.Logger
}

func NewUseCase(
	cfg *config.Config,
	aead aead.Cipher,
	jwkRepo *postgres.JWKRepository,
	logger *zapx.Logger,
) *UseCase {
	return &UseCase{
		cfg:     cfg,
		aead:    aead,
		jwkRepo: jwkRepo,
		logger:  logger,
	}
}

func (u *UseCase) GenerateJWK(alg jose.SignatureAlgorithm, use string, kid ...string) (*jose.JSONWebKey, error) {
	//bits := 0
	//if alg == jose.RS256 || alg == jose.RS384 || alg == jose.RS512 {
	//	bits = 4096
	//}
	return nil, nil
}

func (u *UseCase) CreateJWK(ctx context.Context, jwk *domain.JSONWebKey) error {
	//jwk.CreatedAt = x.NowUTC()
	//err := u.jwkRepo.Create(ctx, jwk)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (u *UseCase) GetOrCreateJWKFn(set domain.Set) jwt.GetPrivateKeyFn {
	return func(ctx context.Context, kid ...string) (any, error) {
		return jose.JSONWebKey{
			Key:       []byte("secret-key-for-hs256"),
			KeyID:     "key-id-for-hs256",
			Algorithm: "HS256",
		}, nil
	}
}
