package oauth

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

type HMACStrategy struct {
	hmac   strategy.OpaqueSigner
	config core.LifetimeConfigProvider
}

func NewHMACStrategy(config core.LifetimeConfigProvider, hmac strategy.OpaqueSigner) *HMACStrategy {
	return &HMACStrategy{
		config: config,
		hmac:   hmac,
	}
}

func (hs *HMACStrategy) AccessTokenSignature(ctx context.Context, token string) string {
	return hs.hmac.GetSignature(token)
}

func (hs *HMACStrategy) GenerateAccessToken(ctx context.Context, request *core.Request) (token string, signature string, err error) {
	return hs.hmac.Generate(ctx, request)
}

func (hs *HMACStrategy) ValidateAccessToken(ctx context.Context, request *core.Request, token string) (err error) {
	exp := request.Session.GetExpiresAt(core.AccessToken)
	if expiredAt := request.RequestedAt.Add(hs.config.GetAccessTokenLifetime()); exp.IsZero() && expiredAt.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Access token expired at '%s'.", expiredAt)
	}

	if !exp.IsZero() && exp.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Access token expired at '%s'.", exp)
	}

	return hs.hmac.Validate(ctx, token)
}

func (hs *HMACStrategy) RefreshTokenSignature(ctx context.Context, token string) string {
	return hs.hmac.GetSignature(token)
}

func (hs *HMACStrategy) GenerateRefreshToken(ctx context.Context, request *core.Request) (token string, signature string, err error) {
	return hs.hmac.Generate(ctx, request)
}

func (hs *HMACStrategy) ValidateRefreshToken(ctx context.Context, request *core.Request, token string) (err error) {
	exp := request.Session.GetExpiresAt(core.RefreshToken)
	if !exp.IsZero() && exp.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Refresh token expired at '%s'.", exp)
	}

	return hs.hmac.Validate(ctx, token)
}

func (hs *HMACStrategy) AuthorizeCodeSignature(ctx context.Context, code string) string {
	return hs.hmac.GetSignature(code)
}

func (hs *HMACStrategy) GenerateAuthorizeCode(ctx context.Context, request *core.Request) (code string, signature string, err error) {
	return hs.hmac.Generate(ctx, request)
}

func (hs *HMACStrategy) ValidateAuthorizeCode(ctx context.Context, request *core.Request, code string) (err error) {
	exp := request.Session.GetExpiresAt(core.AuthorizationCode)
	if expiredAt := request.RequestedAt.Add(hs.config.GetAuthorizationCodeLifetime()); exp.IsZero() && expiredAt.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Authorize code expired at '%s'.", expiredAt)
	}

	if !exp.IsZero() && exp.Before(x.NowUTC()) {
		return core.ErrTokenExpired.WithHint("Authorize code expired at '%s'.", exp)
	}

	return hs.hmac.Validate(ctx, code)
}
