package jwk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"

	"github.com/go-jose/go-jose/v4"
	"github.com/jackc/pgx/v5"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/logger"
)

type UseCase struct {
	cfg     *config.Config
	aead    aead.Cipher
	jwkRepo *Repository
	logger  *logger.Logger
}

func NewUseCase(
	cfg *config.Config,
	aead aead.Cipher,
	jwkRepo *Repository,
	logger *logger.Logger,
) *UseCase {
	return &UseCase{
		cfg:     cfg,
		aead:    aead,
		jwkRepo: jwkRepo,
		logger:  logger,
	}
}

func (u *UseCase) GenerateJWK(alg jose.SignatureAlgorithm, use string, kid ...string) (*jose.JSONWebKey, error) {
	jwk := &jose.JSONWebKey{
		Use:                         use,
		Algorithm:                   string(alg),
		Certificates:                []*x509.Certificate{}, // unused
		CertificateThumbprintSHA256: []byte{},              // unused
		CertificateThumbprintSHA1:   []byte{},              // unused
	}

	if len(kid) > 0 {
		jwk.KeyID = kid[0]
	} else {
		jwk.KeyID = x.RandomUUID()
	}

	switch alg {
	case jose.HS256, jose.HS512:
		keyLen := 256
		key := make([]byte, keyLen)
		_, err := rand.Read(key)
		if err != nil {
			return nil, err
		}
		jwk.Key = key
	case jose.RS256, jose.RS512:
		bits := 2048
		privateKey, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, err
		}
		jwk.Key = privateKey
	default:
		return nil, errors.New("unsupported algorithm")
	}

	return jwk, nil
}

func (u *UseCase) CreateJWK(ctx context.Context, set Set, jwk *jose.JSONWebKey) error {
	active := false
	_, err := u.jwkRepo.GetActiveKey(ctx, set)
	if errors.Is(err, pgx.ErrNoRows) {
		active = true
	} else {
		return err
	}

	jwkBytes, err := json.Marshal(jwk)
	if err != nil {
		return err
	}

	encrypted, err := u.aead.Encrypt(ctx, jwkBytes, nil)
	if err != nil {
		return err
	}

	err = u.jwkRepo.Create(ctx, &KeyData{
		KeyID:     jwk.KeyID,
		SetID:     set,
		Key:       encrypted,
		Active:    active,
		CreatedAt: x.NowUTC(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *UseCase) GetKey(ctx context.Context, set Set, kid ...string) (*jose.JSONWebKey, error) {
	var key *KeyData
	var err error

	if len(kid) > 0 {
		key, err = u.jwkRepo.GetInactiveVerificationKey(ctx, set, kid[0])
	} else {
		key, err = u.jwkRepo.GetActiveKey(ctx, set)
	}
	if err != nil {
		return nil, err
	}

	jwkBytes, err := u.aead.Decrypt(ctx, key.Key, nil)
	if err != nil {
		return nil, err
	}

	jwk := &jose.JSONWebKey{}
	err = json.Unmarshal(jwkBytes, &jwk)
	if err != nil {
		return nil, err
	}

	return jwk, nil
}

func (u *UseCase) GetOrCreateJWKFn(set Set) jwt.GetPrivateKeyFn {
	// this function always return *jose.JSONWebKey key type
	// the any-type is used for extensibility of the core signer
	return func(ctx context.Context, kid ...string) (any, error) {
		jwk, err := u.GetKey(ctx, set, kid...)
		if err == nil {
			return jwk, nil
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		jwk, err = u.GenerateJWK(jose.HS256, "sig")
		if err != nil {
			return nil, err
		}

		err = u.CreateJWK(ctx, set, jwk)
		if err != nil {
			return nil, err
		}

		return jwk, nil
	}
}

func (u *UseCase) ActiveKey(ctx context.Context, set Set, kid string) (*jose.JSONWebKey, error) {
	return nil, nil
}
