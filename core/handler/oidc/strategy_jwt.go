package oidc

import (
	"context"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

type IDTokenStrategyConfigurator interface {
	core.IDTokenIssuerProvider
	core.IDTokenLifetimeProvider
	core.MinParameterEntropyProvider
}

type IDTokenStrategy struct {
	cfg IDTokenStrategyConfigurator
	strategy.JWTSigner
}

func NewIDTokenStrategy(cfg IDTokenStrategyConfigurator, jwtSigner strategy.JWTSigner) *IDTokenStrategy {
	return &IDTokenStrategy{
		cfg:       cfg,
		JWTSigner: jwtSigner,
	}
}

func (i *IDTokenStrategy) GenerateIDToken(ctx context.Context, lifetime time.Duration, tr *core.TokenRequest) (string, error) {
	if lifetime == 0 {
		lifetime = time.Hour
	}

	oidcSession, ok := tr.Session.(OpenIDConnectSession)
	if !ok {
		return "", core.ErrServerError.WithDebug("Failed to generate id token because session must be of type OpenIDConnectSession")
	}

	claims := oidcSession.IDTokenClaims()
	if s, _ := claims.GetSubject(); s == "" {
		return "", core.ErrServerError.WithDebug("Failed to generate id token because session subject is empty.")
	}

	if tr.GrantType.ExactOne("refresh_token") {
	}

	if claims.ExpiresAt.IsZero() {
		claims.ExpiresAt = gojwt.NewNumericDate(x.NowUTC().Add(lifetime))
	}

	if claims.ExpiresAt.Before(x.NowUTC()) {
		return "", core.ErrServerError.WithDebug("Failed to generate id token because expiry claim can not be in the past.")
	}

	if claims.AuthTime.IsZero() {
		claims.AuthTime = x.NowUTC().Truncate(time.Second)
	}

	if claims.Issuer == "" {
		claims.Issuer = i.cfg.GetIDTokenIssuer()
	}

	claims.Audience = append(claims.Audience, tr.Client.GetID())
	claims.IssuedAt = gojwt.NewNumericDate(x.NowUTC())

	token, _, err := i.JWTSigner.Generate(ctx, claims)
	return token, err
}
