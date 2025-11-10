package strategy

import (
	"context"
	"time"

	"github.com/tuanta7/hydros/core"
)

type OpenIDConnectTokenStrategy interface {
	GenerateIDToken(ctx context.Context, lifetime time.Duration, tr *core.TokenRequest) (token string, err error)
}
