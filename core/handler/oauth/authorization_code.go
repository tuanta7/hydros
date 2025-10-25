package oauth

import (
	"context"
	"errors"

	core2 "github.com/tuanta7/hydros/core"
)

type AuthorizationCodeGrantHandler struct {
}

func NewAuthorizationCodeGrantHandler() *AuthorizationCodeGrantHandler {
	return &AuthorizationCodeGrantHandler{}
}

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeRequest(
	ctx context.Context,
	req *core2.AuthorizeRequest,
	res *core2.AuthorizeResponse,
) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core2.ErrUnknownRequest
	}

	return nil
}

func (h *AuthorizationCodeGrantHandler) AuthenticateClient(
	ctx context.Context,
	req *core2.TokenRequest,
) error {
	if req.Client != nil && req.Client.IsPublic() {
		return nil
	}

	// TODO: support confidential client
	return errors.New("confidential client is not supported yet")
}

func (h *AuthorizationCodeGrantHandler) HandleTokenRequest(
	ctx context.Context,
	req *core2.TokenRequest,
	res *core2.TokenResponse,
) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core2.ErrUnknownRequest
	}

	return nil
}
