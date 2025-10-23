package oauth

import (
	"context"
	"errors"

	"github.com/tuanta7/hydros/internal/core"
)

type AuthorizationCodeGrantHandler struct {
}

func NewAuthorizationCodeGrantHandler() *AuthorizationCodeGrantHandler {
	return &AuthorizationCodeGrantHandler{}
}

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeRequest(
	ctx context.Context,
	req *core.AuthorizeRequest,
	res *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core.ErrUnknownRequest
	}

	return nil
}

func (h *AuthorizationCodeGrantHandler) AuthenticateClient(
	ctx context.Context,
	req *core.TokenRequest,
) error {
	if req.Client != nil && req.Client.IsPublic() {
		return nil
	}

	// TODO: support confidential client
	return errors.New("confidential client is not supported yet")
}

func (h *AuthorizationCodeGrantHandler) HandleTokenRequest(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
	}

	return nil
}
