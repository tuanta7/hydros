package storage

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type AuthorizeCodeStorage interface {
}

type AccessTokenStorage interface {
	CreateSession(ctx context.Context, signature string, req *core.TokenRequest) error
	GetSession(ctx context.Context, signature string, session *core.Session) (*core.TokenRequest, error)
	DeleteSession(ctx context.Context, signature string) error
}

type RefreshTokenStorage interface {
	RotateRefreshToken(ctx context.Context, requestID string, signature string) (err error)
}
