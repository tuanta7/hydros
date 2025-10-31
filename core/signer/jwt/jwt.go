package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
)

type Configurator interface {
	core.AccessTokenIssuerProvider
}

type GetPrivateKeyFn func(ctx context.Context, kid ...string) (core.JSONWebKey, error)

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

func (s DefaultSigner) Generate(ctx context.Context, request *core.Request, tokenType core.TokenType) (string, string, error) {
	key, err := s.getPrivateKeyFn(ctx)
	if err != nil {
		return "", "", err
	}

	claims := &Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        strings.Replace(uuid.NewString(), "-", "", -1),
			Issuer:    s.config.GetAccessTokenIssuer(),
			Subject:   request.Session.GetSubject(),
			Audience:  gojwt.ClaimStrings(request.GrantedAudience),
			IssuedAt:  gojwt.NewNumericDate(x.NowUTC()),
			ExpiresAt: gojwt.NewNumericDate(request.Session.GetExpiresAt(tokenType)),
		},
		ClientID: request.Client.GetID(),
		Scope:    strings.Join(request.GrantedScope, " "),
	}

	algorithm := key.GetAlgorithm()
	if algorithm == nil || algorithm == gojwt.SigningMethodNone {
		return "", "", fmt.Errorf("invalid signing algorithm: %s", algorithm)
	}
	token := gojwt.NewWithClaims(algorithm, claims)

	privateKey := key.GetPrivateKey()
	if privateKey == nil {
		return "", "", errors.New("private key is not set")
	}

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", err
	}

	return signedToken, s.GetSignature(signedToken), nil
}

func (s DefaultSigner) GetSignature(token string) string {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	return parts[2]
}

func (s DefaultSigner) Validate(ctx context.Context, token string) (err error) {
	key, err := s.getPrivateKeyFn(ctx)
	if err != nil {
		return err
	}

	publicKey := key.GetPublicKey()
	if publicKey == nil {
		return errors.New("public key is not set")
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
