package storage

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type OpenIDConnectRequestStorage interface {
	CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, req *core.Request) error
	GetOpenIDConnectSession(ctx context.Context, authorizeCode string, session core.Session) (*core.Request, error)
	DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error
}
