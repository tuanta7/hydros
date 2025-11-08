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
	core.MinParameterEntropyProvider
	core.AuthorizationCodeLifetimeProvider
}

type AuthorizationCodeGrantHandler struct {
	config        AuthorizationCodeGrantConfigurator
	tokenStrategy strategy.TokenStrategy
	tokenStorage  storage.TokenStorage
}

func NewAuthorizationCodeGrantHandler(
	config AuthorizationCodeGrantConfigurator,
	tokenStrategy strategy.TokenStrategy,
	tokenStorage storage.TokenStorage,
) *AuthorizationCodeGrantHandler {
	return &AuthorizationCodeGrantHandler{
		config:        config,
		tokenStrategy: tokenStrategy,
		tokenStorage:  tokenStorage,
	}
}

func validateResponseMode(ar *core.AuthorizeRequest, registered []core.ResponseMode) bool {
	for _, mode := range registered {
		if ar.ResponseMode == mode {
			return true
		}
	}
	return false
}

func validateResponseType(ar *core.AuthorizeRequest, registered []string) bool {
	for _, t := range registered {
		if ar.ResponseTypes.ExactAll(x.SplitSpace(t)...) {
			return true
		}
	}
	return false
}

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeRequest(ctx context.Context, ar *core.AuthorizeRequest) (err error) {
	if !ar.ResponseTypes.ExactOne("code") {
		return core.ErrUnsupportedResponseType.WithHint("The server only supports the response_type 'code'.")
	}

	client := ar.Client
	if client == nil {
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	if !validateResponseType(ar, client.GetResponseTypes()) {
		return core.ErrUnsupportedResponseType.WithHint("The client is not allowed to request response_type '%s'.", ar.ResponseTypes)
	}

	if !validateResponseMode(ar, client.GetResponseModes()) {
		return core.ErrUnsupportedResponseMode.WithHint("The client is not allowed to request response_mode '%s'.", ar.ResponseMode)
	}

	scopeStrategy := h.config.GetScopeStrategy()
	for _, scope := range ar.RequestedScope {
		if !scopeStrategy(client.GetScopes(), scope) {
			return core.ErrInvalidScope.WithHint("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope)
		}
	}

	audienceStrategy := h.config.GetAudienceStrategy()
	if err = audienceStrategy(client.GetAudience(), ar.RequestedAudience); err != nil {
		return err
	}

	return nil
}

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeResponse(
	ctx context.Context,
	req *core.AuthorizeRequest,
	res *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core.ErrUnsupportedResponseType.WithHint("The server only supports the response_type 'code'.")
	}

	req.DefaultResponseMode = core.ResponseModeQuery

	if !x.IsURISecure(req.RedirectURI) {
		return core.ErrInvalidRequest.WithHint("Redirect URL is using an insecure protocol, http is only allowed for hosts with suffix 'localhost', for example: http://app.localhost/.")
	}

	client := req.Client
	if client == nil {
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	scopeStrategy := h.config.GetScopeStrategy()
	for _, scope := range req.RequestedScope {
		if !scopeStrategy(client.GetScopes(), scope) {
			return core.ErrInvalidScope.WithHint("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope)
		}
	}

	audienceStrategy := h.config.GetAudienceStrategy()
	if err := audienceStrategy(client.GetAudience(), req.RequestedAudience); err != nil {
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
