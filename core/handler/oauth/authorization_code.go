package oauth

import (
	"context"
	"fmt"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
)

type AuthorizationCodeGrantHandler struct {
	tokenStrategy            strategy.TokenStrategy
	authorizationCodeStorage storage.AuthorizationCodeStorage
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

func (h *AuthorizationCodeGrantHandler) HandleTokenRequest(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
	}

	code := req.Code
	signature := h.tokenStrategy.AuthorizeCodeSignature(ctx, code)
	authorizeRequest, err := h.authorizationCodeStorage.GetAuthorizationCodeSession(ctx, signature, req.Session)
	if err != nil {
		return err
	}

	// TODO
	fmt.Println("authorizeRequest:", authorizeRequest)

	return nil
}
