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
	CreateAuthorizeCodeSession(ctx context.Context, signature string, req *core.Request) (err error)
	GetAuthorizationCodeSession(ctx context.Context, signature string, session core.Session) (*core.Request, error)
	InvalidateAuthorizeCodeSession(ctx context.Context, signature string) (err error)
}

type AccessTokenStorage interface {
	CreateAccessTokenSession(ctx context.Context, signature string, req *core.Request) error
	GetAccessTokenSession(ctx context.Context, signature string, session core.Session) (*core.Request, error)
	DeleteAccessTokenSession(ctx context.Context, signature string) error
}

type RefreshTokenStorage interface {
	GetRefreshTokenSession(ctx context.Context, signature string, session core.Session) (*core.Request, error)
	RotateRefreshToken(ctx context.Context, requestID string, signature string) (err error)
}
