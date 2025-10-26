package oauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
)

type AuthorizationCodeGrantHandler struct {
	authorizationCodeStrategy strategy.TokenStrategy
	authorizationCodeStorage  storage.AuthorizationCodeStorage
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

	code := req.Code
	signature := h.authorizationCodeStrategy.GetSignature(code)
	authorizeRequest, err := h.authorizationCodeStorage.GetAuthorizationCodeSession(ctx, signature, req.Session)
	if err != nil {
		return err
	}

	// TODO
	fmt.Println("authorizeRequest:", authorizeRequest)

	return nil
}
