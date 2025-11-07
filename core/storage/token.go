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
	CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, req *core.Request) (err error)
	GetRefreshTokenSession(ctx context.Context, signature string, session core.Session) (*core.Request, error)
	DeleteRefreshTokenSession(ctx context.Context, signature string) (err error)
	RotateRefreshToken(ctx context.Context, requestID string, signature string) (err error)
}

// IDTokenStorage is unused. ID Token is always JWT, so no storage is needed
type IDTokenStorage interface{}
