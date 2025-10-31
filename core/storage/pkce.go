package storage

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type PKCERequestStorage interface {
	GetPKCERequestSession(ctx context.Context, signature string, session core.Session) (*core.AuthorizeRequest, error)
	CreatePKCERequestSession(ctx context.Context, signature string, request *core.AuthorizeRequest) error
	DeletePKCERequestSession(ctx context.Context, signature string) error
}
