package oauth

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type UseCase struct {
	authorizeHandlers []core.AuthorizeHandler
	tokenHandlers     []core.TokenHandler
}

func (u *UseCase) HandleTokenEndpoint(ctx context.Context, req *core.TokenRequest, res *core.TokenResponse) error {
	for _, ah := range u.tokenHandlers {
		_ = ah.HandleTokenRequest(ctx, req, res)
	}

	return nil
}
