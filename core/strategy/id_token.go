package strategy

import (
	"context"
	"time"

	"github.com/tuanta7/hydros/core"
)

type OpenIDConnectTokenStrategy interface {
	GenerateIDToken(ctx context.Context, lifetime time.Duration, r *core.Request) (token string, err error)
	ValidateIDToken(ctx context.Context, token string) (err error)
}

func (js JWTStrategy) GenerateIDToken(ctx context.Context, lifetime time.Duration, r *core.Request) (token string, err error) {
	return "", nil
}

func (js JWTStrategy) ValidateIDToken(ctx context.Context, token string) (err error) {
	return nil
}
