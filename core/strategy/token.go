package strategy

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
)

// TokenStrategy defines the methods used for managing token and authorization code
type TokenStrategy interface {
	AccessTokenStrategy
	RefreshTokenStrategy
	AuthorizeCodeStrategy
}

type AccessTokenStrategy interface {
	AccessTokenSignature(ctx context.Context, token string) string
	GenerateAccessToken(ctx context.Context, request *core.TokenRequest) (token string, signature string, err error)
	ValidateAccessToken(ctx context.Context, request *core.TokenRequest, token string) (err error)
}

type RefreshTokenStrategy interface {
	RefreshTokenSignature(ctx context.Context, token string) string
	GenerateRefreshToken(ctx context.Context, request *core.TokenRequest) (token string, signature string, err error)
	ValidateRefreshToken(ctx context.Context, request *core.TokenRequest, token string) (err error)
}

type AuthorizeCodeStrategy interface {
	AuthorizeCodeSignature(ctx context.Context, code string) string
	GenerateAuthorizeCode(ctx context.Context, request *core.TokenRequest) (code string, signature string, err error)
	ValidateAuthorizeCode(ctx context.Context, request *core.TokenRequest, code string) (err error)
}

type HMACStrategy struct {
	hmac   Signer
	config core.LifetimeConfigProvider
}

func NewHMACStrategy(config core.LifetimeConfigProvider, hmac Signer) *HMACStrategy {
	return &HMACStrategy{
		config: config,
		hmac:   hmac,
	}
}

func (hs HMACStrategy) AccessTokenSignature(ctx context.Context, token string) string {
	return hs.hmac.GetSignature(token)
}

func (hs HMACStrategy) GenerateAccessToken(ctx context.Context, request *core.TokenRequest) (token string, signature string, err error) {
	return hs.hmac.Generate(ctx, request, core.AccessToken)
}

func (hs HMACStrategy) ValidateAccessToken(ctx context.Context, request *core.TokenRequest, token string) (err error) {
	exp := request.Session.GetExpiresAt(core.AccessToken)
	if expiredAt := request.RequestedAt.Add(hs.config.GetAccessTokenLifetime()); exp.IsZero() && expiredAt.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Access token expired at '%s'.", expiredAt)
	}

	if !exp.IsZero() && exp.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Access token expired at '%s'.", exp)
	}

	return hs.hmac.Validate(ctx, token)
}

func (hs HMACStrategy) RefreshTokenSignature(ctx context.Context, token string) string {
	return hs.hmac.GetSignature(token)
}

func (hs HMACStrategy) GenerateRefreshToken(ctx context.Context, request *core.TokenRequest) (token string, signature string, err error) {
	return hs.hmac.Generate(ctx, request, core.RefreshToken)
}

func (hs HMACStrategy) ValidateRefreshToken(ctx context.Context, request *core.TokenRequest, token string) (err error) {
	exp := request.Session.GetExpiresAt(core.RefreshToken)
	if !exp.IsZero() && exp.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Refresh token expired at '%s'.", exp)
	}

	// exp = 0 (unlimited lifetime) or token is not expired
	return hs.hmac.Validate(ctx, token)
}

func (hs HMACStrategy) AuthorizeCodeSignature(ctx context.Context, code string) string {
	return hs.hmac.GetSignature(code)
}

func (hs HMACStrategy) GenerateAuthorizeCode(ctx context.Context, request *core.TokenRequest) (code string, signature string, err error) {
	return hs.hmac.Generate(ctx, request, core.AuthorizationCode)
}

func (hs HMACStrategy) ValidateAuthorizeCode(ctx context.Context, request *core.TokenRequest, code string) (err error) {
	exp := request.Session.GetExpiresAt(core.AuthorizationCode)
	if expiredAt := request.RequestedAt.Add(hs.config.GetAuthorizationCodeLifetime()); exp.IsZero() && expiredAt.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Authorize code expired at '%s'.", expiredAt)
	}

	if !exp.IsZero() && exp.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Authorize code expired at '%s'.", exp)
	}

	return hs.hmac.Validate(ctx, code)
}

type JWTStrategy struct {
	hmac Signer
	jwt  Signer
}

func NewJWTStrategy(hmac Signer, jwt Signer) *JWTStrategy {
	return &JWTStrategy{
		hmac: hmac,
		jwt:  jwt,
	}
}

func (js JWTStrategy) AccessTokenSignature(ctx context.Context, token string) string {
	return js.jwt.GetSignature(token)
}

func (js JWTStrategy) GenerateAccessToken(ctx context.Context, request *core.TokenRequest) (token string, signature string, err error) {
	return js.jwt.Generate(ctx, request, core.AccessToken)
}

func (js JWTStrategy) ValidateAccessToken(ctx context.Context, request *core.TokenRequest, token string) (err error) {
	return js.jwt.Validate(ctx, token)
}

func (js JWTStrategy) RefreshTokenSignature(ctx context.Context, token string) string {
	return js.hmac.GetSignature(token)
}

func (js JWTStrategy) GenerateRefreshToken(ctx context.Context, request *core.TokenRequest) (token string, signature string, err error) {
	return js.hmac.Generate(ctx, request, core.RefreshToken)
}

func (js JWTStrategy) ValidateRefreshToken(ctx context.Context, request *core.TokenRequest, token string) (err error) {
	return js.hmac.Validate(ctx, token)
}

func (js JWTStrategy) AuthorizeCodeSignature(ctx context.Context, code string) string {
	return js.hmac.GetSignature(code)
}

func (js JWTStrategy) GenerateAuthorizeCode(ctx context.Context, request *core.TokenRequest) (code string, signature string, err error) {
	return js.hmac.Generate(ctx, request, core.AuthorizationCode)
}

func (js JWTStrategy) ValidateAuthorizeCode(ctx context.Context, request *core.TokenRequest, code string) (err error) {
	return js.hmac.Validate(ctx, code)
}
