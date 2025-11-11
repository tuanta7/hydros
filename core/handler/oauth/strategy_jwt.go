package oauth

import (
	"context"
	"strings"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

type JWTStrategy struct {
	cfg  core.AccessTokenIssuerProvider
	hmac strategy.OpaqueSigner
	jwt  strategy.JWTSigner
}

func NewJWTStrategy(cfg core.AccessTokenIssuerProvider, hmac strategy.OpaqueSigner, jwt strategy.JWTSigner) *JWTStrategy {
	return &JWTStrategy{
		cfg:  cfg,
		hmac: hmac,
		jwt:  jwt,
	}
}

func (js *JWTStrategy) AccessTokenSignature(ctx context.Context, token string) string {
	return js.jwt.GetSignature(token)
}

func (js *JWTStrategy) GenerateAccessToken(ctx context.Context, request *core.Request) (token string, signature string, err error) {
	claims := &jwt.Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        x.RandomUUID(),
			Issuer:    js.cfg.GetAccessTokenIssuer(),
			Subject:   request.Session.GetSubject(),
			Audience:  gojwt.ClaimStrings(request.GrantedAudience),
			IssuedAt:  gojwt.NewNumericDate(x.NowUTC()),
			ExpiresAt: gojwt.NewNumericDate(request.Session.GetExpiresAt(core.AccessToken)),
		},
		ClientID: request.Client.GetID(),
		Scope:    strings.Join(request.GrantedScope, " "),
	}

	return js.jwt.Generate(ctx, claims)
}

func (js *JWTStrategy) ValidateAccessToken(ctx context.Context, request *core.Request, token string) (err error) {
	return js.jwt.Validate(ctx, token)
}

func (js *JWTStrategy) RefreshTokenSignature(ctx context.Context, token string) string {
	return js.hmac.GetSignature(token)
}

func (js *JWTStrategy) GenerateRefreshToken(ctx context.Context, request *core.Request) (token string, signature string, err error) {
	return js.hmac.Generate(ctx, request)
}

func (js *JWTStrategy) ValidateRefreshToken(ctx context.Context, request *core.Request, token string) (err error) {
	return js.hmac.Validate(ctx, token)
}

func (js *JWTStrategy) AuthorizeCodeSignature(ctx context.Context, code string) string {
	return js.hmac.GetSignature(code)
}

func (js *JWTStrategy) GenerateAuthorizeCode(ctx context.Context, request *core.Request) (code string, signature string, err error) {
	return js.hmac.Generate(ctx, request)
}

func (js *JWTStrategy) ValidateAuthorizeCode(ctx context.Context, request *core.Request, code string) (err error) {
	return js.hmac.Validate(ctx, code)
}
