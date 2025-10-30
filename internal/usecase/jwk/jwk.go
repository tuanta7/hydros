package jwk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/internal/domain"
)

type UseCase struct {
	cfg *config.Config
}

func NewUseCase() *UseCase {
	return &UseCase{}
}

func (u *UseCase) GenerateJWK(ctx context.Context, algorithm domain.Algorithm, keySize int) (*domain.JSONWebKey, error) {
	switch algorithm {
	case domain.AlgorithmRS256, domain.AlgorithmRS512:
		privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
		if err != nil {
			return nil, err
		}

		return &domain.JSONWebKey{
			Key:       privateKey,
			Algorithm: algorithm,
		}, nil
	case domain.AlgorithmHS256, domain.AlgorithmHS512:
		key := make([]byte, keySize)
		_, err := rand.Read(key)
		if err != nil {
			return nil, err
		}

		return &domain.JSONWebKey{
			Key:       key,
			Algorithm: algorithm,
		}, nil
	}

	return nil, errors.New("unsupported algorithm")
}

func (u *UseCase) CreateJWK(ctx context.Context, jwk *domain.JSONWebKey) error {
	if jwk == nil {
		return errors.New("jwk cannot be nil")
	}

	if jwk.Set == "" {
		return errors.New("jwk set cannot be empty")
	}

	if jwk.Use == "" {
		jwk.Use = "sig"
	}

	if jwk.KeyID == "" {
		jwk.KeyID = strings.Replace(uuid.NewString(), "-", "", -1)
	}

	return nil
}

func (u *UseCase) GetOrCreateJWKFn(ctx context.Context, set domain.Set, kid ...string) jwt.GetPrivateKeyFn {
	return func(ctx context.Context, kid ...string) (core.JSONWebKey, error) {
		return domain.JSONWebKey{
			KeyID:     "fake-key-id",
			Key:       "a14a1ef90f33ffc4aa7a47739dd042a2",
			Algorithm: domain.AlgorithmHS256,
			Use:       "sig",
			Set:       set,
			Active:    true,
		}, nil
	}
}
