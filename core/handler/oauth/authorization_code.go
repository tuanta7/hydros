package oauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

type AuthorizationCodeGrantConfigurator interface {
	strategy.ScopeStrategyProvider
	strategy.AudienceStrategyProvider
	core.AuthorizationCodeLifetimeProvider
}

type AuthorizationCodeGrantHandler struct {
	config        AuthorizationCodeGrantConfigurator
	tokenStrategy strategy.TokenStrategy
	tokenStorage  storage.TokenStorage
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
		// we don't need to return ErrUnknownRequest here, because the /authorize endpoint is only for the
		// authorization code grant type, there is no more handler to fall back to.
		return nil
	}

	req.DefaultResponseMode = core.ResponseModeQuery
	if !x.IsURISecure(req.RedirectURI) {
		return core.ErrInvalidRequest.WithHint("Redirect URL is using an insecure protocol, http is only allowed for hosts with suffix 'localhost', for example: http://app.localhost/.")
	}

	client := req.Client
	if client == nil {
		// should never happen because NewAuthorizeRequest already checks this
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	scopeStrategy := h.config.GetScopeStrategy()
	for _, scope := range req.Scope {
		if !scopeStrategy(client.GetScopes(), scope) {
			return core.ErrInvalidScope.WithHint("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope)
		}
	}

	audienceStrategy := h.config.GetAudienceStrategy()
	err := audienceStrategy(client.GetAudience(), req.Audience)
	if err != nil {
		return err
	}

	code, signature, err := h.tokenStrategy.GenerateAuthorizeCode(ctx, &req.Request)
	if err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	if req.Session == nil {
		return errors.New("session cannot be nil")
	}

	req.Session.SetExpiresAt(core.AuthorizationCode, x.NowUTC().Add(h.config.GetAuthorizationCodeLifetime()))
	err = h.tokenStorage.CreateAuthorizeCodeSession(ctx, signature, &req.Request)
	if err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	res.Code = code
	res.State = req.State
	res.Scope = strings.Join(req.GrantedScope, " ")

	req.HandledResponseTypes = req.HandledResponseTypes.Append("code")
	return nil
}

func (h *AuthorizationCodeGrantHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	if !req.GrantType.ExactOne("authorization_code") {
		// return ErrUnknownRequest here, so the next handler can try to handle it.
		return core.ErrUnknownRequest
	}

	code := req.Code
	signature := h.tokenStrategy.AuthorizeCodeSignature(ctx, code)
	authorizeRequest, err := h.tokenStorage.GetAuthorizationCodeSession(ctx, signature, req.Session)
	if err != nil {
		return err
	}

	// TODO
	fmt.Println("authorizeRequest:", authorizeRequest)

	return nil
}

func (h *AuthorizationCodeGrantHandler) HandleTokenResponse(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	return nil
}
