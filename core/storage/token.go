package storage

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type TokenStorage interface {
	AccessTokenStorage
	RefreshTokenStorage
	AuthorizationCodeStorage
}

type AuthorizationCodeStorage interface {
	CreateAuthorizeCodeSession(ctx context.Context, code string, req *core.Request) (err error)
	GetAuthorizationCodeSession(ctx context.Context, code string, session core.Session) (*core.Request, error)
	InvalidateAuthorizeCodeSession(ctx context.Context, code string) (err error)
}

type AccessTokenStorage interface {
	CreateAccessTokenSession(ctx context.Context, signature string, req *core.TokenRequest) error
	GetAccessTokenSession(ctx context.Context, signature string, session core.Session) (*core.TokenRequest, error)
	DeleteAccessTokenSession(ctx context.Context, signature string) error
}

type RefreshTokenStorage interface {
	GetRefreshTokenSession(ctx context.Context, signature string, session core.Session) (*core.TokenRequest, error)
	RotateRefreshToken(ctx context.Context, requestID string, signature string) (err error)
}
