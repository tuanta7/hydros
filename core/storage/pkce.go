package storage

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type PKCERequestStorage interface {
	GetPKCERequestSession(ctx context.Context, signature string, session core.Session) (*core.Request, error)
	CreatePKCERequestSession(ctx context.Context, signature string, request *core.Request) error
	DeletePKCERequestSession(ctx context.Context, signature string) error
}
