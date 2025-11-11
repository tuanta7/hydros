package jwt

import (
	"context"
	"crypto/rsa"
	"errors"
	"strings"

	"github.com/go-jose/go-jose/v4"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/tuanta7/hydros/core"
)

var SupportedAlgorithm = map[string]gojwt.SigningMethod{
	"RS512": gojwt.SigningMethodRS512,
	"RS256": gojwt.SigningMethodRS256,
	"HS512": gojwt.SigningMethodHS512,
	"HS256": gojwt.SigningMethodHS256,
	//"ES256": gojwt.SigningMethodES256,
	//"ES512": gojwt.SigningMethodES512,
}

type Configurator interface {
	core.AccessTokenIssuerProvider
}

// GetPrivateKeyFn return the current active key or an inactive key but still allowed to validate token signature.
// The kid is optional, which is used to select the key that might be inactivated.
type GetPrivateKeyFn func(ctx context.Context, kid ...string) (any, error)

type DefaultSigner struct {
	config          Configurator
	getPrivateKeyFn GetPrivateKeyFn
}

func NewSigner(cfg Configurator, fn GetPrivateKeyFn) (*DefaultSigner, error) {
	return &DefaultSigner{
		config:          cfg,
		getPrivateKeyFn: fn,
	}, nil
}

func (s *DefaultSigner) Generate(ctx context.Context, claims gojwt.Claims, headers ...map[string]any) (string, string, error) {
	privateKey, algorithm, err := s.getSignKey(ctx)

	token := gojwt.NewWithClaims(algorithm, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", err
	}

	return signedToken, s.GetSignature(signedToken), nil
}

func (s *DefaultSigner) getSignKey(ctx context.Context) (any, gojwt.SigningMethod, error) {
	key, err := s.getPrivateKeyFn(ctx)
	if err != nil {
		return nil, nil, err
	}

	var privateKey any
	var algorithm gojwt.SigningMethod

	switch t := key.(type) {
	case *jose.JSONWebKey:
		privateKey = t.Key
		algorithm = SupportedAlgorithm[t.Algorithm]
	case jose.JSONWebKey:
		privateKey = t.Key
		algorithm = SupportedAlgorithm[t.Algorithm]
	case *rsa.PrivateKey:
		privateKey = t
		algorithm = gojwt.SigningMethodRS256
	case string:
		privateKey = []byte(t)
		algorithm = gojwt.SigningMethodHS256
	case []byte:
		privateKey = t
		algorithm = gojwt.SigningMethodHS256
	}

	return privateKey, algorithm, nil
}

func (s *DefaultSigner) GetSignature(token string) string {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	return parts[2]
}

func (s *DefaultSigner) Validate(ctx context.Context, token string) (err error) {
	publicKey, err := s.getVerificationKey(ctx)
	if err != nil {
		return err
	}

	parser := gojwt.Parser{}
	t, err := parser.ParseWithClaims(token, &Claims{}, func(t *gojwt.Token) (any, error) {
		return publicKey, nil
	})
	if err != nil {
		return toRFCErr(err)
	}

	if t.Method == gojwt.SigningMethodNone {
		return errors.New("token signing method none is not allowed")
	}

	return nil
}

func (s *DefaultSigner) getVerificationKey(ctx context.Context) (any, error) {
	key, err := s.getPrivateKeyFn(ctx)
	if err != nil {
		return nil, err
	}

	if t, ok := key.(*jose.JSONWebKey); ok {
		key = t.Key
	}

	if t, ok := key.(jose.JSONWebKey); ok {
		key = t.Key
	}

	switch t := key.(type) {
	case *rsa.PrivateKey:
		return t.Public(), nil
	case string:
		return t, nil
	case []byte:
		return t, nil
	}

	return nil, errors.New("public key is not set")
}

func toRFCErr(err error) *core.RFC6749Error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gojwt.ErrTokenMalformed):
		return core.ErrInvalidTokenFormat
	case errors.Is(err, gojwt.ErrSignatureInvalid) || errors.Is(err, gojwt.ErrTokenUnverifiable):
		return core.ErrTokenSignatureMismatch
	case errors.Is(err, gojwt.ErrTokenExpired):
		return core.ErrTokenExpired
	case errors.Is(err, gojwt.ErrTokenInvalidAudience) ||
		errors.Is(err, gojwt.ErrTokenUsedBeforeIssued) ||
		errors.Is(err, gojwt.ErrTokenInvalidIssuer) ||
		errors.Is(err, gojwt.ErrTokenInvalidClaims) ||
		errors.Is(err, gojwt.ErrTokenNotValidYet) ||
		errors.Is(err, gojwt.ErrTokenInvalidId):
		return core.ErrTokenClaim
	default:
		return core.ErrRequestUnauthorized
	}
}
