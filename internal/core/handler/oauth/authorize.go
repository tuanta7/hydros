package oauth

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/core"
)

type AuthorizationCodeGrantHandler struct {
}

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeRequest(
	ctx context.Context,
	req *core.AuthorizeRequest,
	res *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.ExactOne("code") {
		return nil
	}

	return nil
}
