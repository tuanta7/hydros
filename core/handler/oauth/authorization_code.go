package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeRequest(ctx context.Context, req *core.AuthorizeRequest) (err error) {
	if !req.ResponseTypes.ExactOne("code") {
		return nil
	}

	client := req.Client
	if client == nil {
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	registered := client.GetRedirectURIs()
	if req.RedirectURI.String() == "" && len(registered) == 1 {
		req.RedirectURI, err = url.Parse(registered[0]) // use the client only valid registered redirect_uri
		if err != nil {
			return core.ErrInvalidRequest.WithHint("Invalid redirect_uri \"%s\".", registered[0]).WithWrap(err)
		}
	}

	if req.RedirectURI.String() != "" {
		ok := x.IsMatchingURI(req.RedirectURI, registered)
		if !ok {
			return core.ErrInvalidRequest.WithHint("The 'redirect_uri' parameter does not match any of the OAuth 2.0 Client's pre-registered redirect urls.")
		}
	}

	if len(req.ResponseTypes) == 0 {
		return core.ErrUnsupportedResponseType.WithHint("The request is missing the 'response_type' parameter.")
	}

	if err = validateResponseType(req, client.GetResponseTypes()); err != nil {
		return err
	}

	if err = validateResponseMode(req, client.GetResponseModes()); err != nil {
		return err
	}

	scopeStrategy := h.config.GetScopeStrategy()
	for _, scope := range req.Scope {
		if !scopeStrategy(client.GetScopes(), scope) {
			return core.ErrInvalidScope.WithHint("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope)
		}
	}

	audienceStrategy := h.config.GetAudienceStrategy()
	if err = audienceStrategy(client.GetAudience(), req.Audience); err != nil {
		return err
	}

	if len(req.State) < h.config.GetMinParameterEntropy() {
		return core.ErrInvalidState.WithHint("Request parameter 'state' must be at least be %d characters long to ensure sufficient entropy.", h.config.GetMinParameterEntropy())
	}

	return nil
}

func validateResponseMode(req *core.AuthorizeRequest, registered []core.ResponseMode) error {
	found := false
	for _, mode := range registered {
		if mode == req.ResponseMode {
			found = true
			break
		}
	}

	if !found {
		return core.ErrUnsupportedResponseMode.WithHint("The client is not allowed to request response_mode '%s'.", req.ResponseMode)
	}

	return nil
}

func validateResponseType(req *core.AuthorizeRequest, registered []string) error {
	var found bool
	for _, t := range registered {
		if req.ResponseTypes.ExactAll(x.SpaceSplit(t)...) {
			found = true
			break
		}
	}

	if !found {
		return core.ErrUnsupportedResponseType.WithHint("The client is not allowed to request response_type '%s'.", req.ResponseTypes)
	}
	return nil
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

func (h *AuthorizationCodeGrantHandler) HandleAuthorizeResponse(
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

	// duplicate check as in HandleAuthorizeRequest, just to be sure after the login flow
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
