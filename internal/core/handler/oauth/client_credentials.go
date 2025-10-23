package oauth

import (
	"context"

	"github.com/tuanta7/hydros/internal/core"
	"github.com/tuanta7/hydros/internal/core/strategy"
)

type ClientCredentialsGrantHandler struct {
	scopeStrategy strategy.ScopeStrategy
}

func (h *ClientCredentialsGrantHandler) HandleTokenRequest(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("client_credentials") {
		return core.ErrUnknownRequest
	}

	client := req.Client
	for _, scope := range req.RequestedScope {
		if !h.scopeStrategy(client.GetScopes(), scope) {
			return core.ErrInvalidScope
		}
	}

	return nil
}
