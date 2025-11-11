package strategy

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

// TokenStrategy defines the methods used for managing token and authorization code
type TokenStrategy interface {
	AccessTokenStrategy
	RefreshTokenStrategy
	AuthorizeCodeStrategy
}

type AccessTokenStrategy interface {
	AccessTokenSignature(ctx context.Context, token string) string
	GenerateAccessToken(ctx context.Context, request *core.Request) (token string, signature string, err error)
	ValidateAccessToken(ctx context.Context, request *core.Request, token string) (err error)
}

type RefreshTokenStrategy interface {
	RefreshTokenSignature(ctx context.Context, token string) string
	GenerateRefreshToken(ctx context.Context, request *core.Request) (token string, signature string, err error)
	ValidateRefreshToken(ctx context.Context, request *core.Request, token string) (err error)
}

type AuthorizeCodeStrategy interface {
	AuthorizeCodeSignature(ctx context.Context, code string) string
	GenerateAuthorizeCode(ctx context.Context, request *core.Request) (code string, signature string, err error)
	ValidateAuthorizeCode(ctx context.Context, request *core.Request, code string) (err error)
}
