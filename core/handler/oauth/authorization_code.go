package oauth

import (
	"context"
	stderr "errors"
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

	if err = scopeStrategy(client.GetScopes(), ar.RequestedScope); err != nil {
		return err
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
	if err := scopeStrategy(client.GetScopes(), req.RequestedScope); err != nil {
		return err
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
		return core.ErrServerError.WithHint("session cannot be nil")
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

func (h *AuthorizationCodeGrantHandler) HandleTokenRequest(ctx context.Context, tokenRequest *core.TokenRequest) error {
	if !tokenRequest.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
	}

	client := tokenRequest.Client
	if client == nil {
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	if !client.GetGrantTypes().IncludeAll("authorization_code") {
		return core.ErrUnauthorizedClient.WithHint("The OAuth 2.0 Client is not allowed to use authorization grant \"authorization_code\".")
	}

	code := tokenRequest.Code
	signature := h.tokenStrategy.AuthorizeCodeSignature(ctx, code)
	authorizeRequest, err := h.tokenStorage.GetAuthorizationCodeSession(ctx, signature, tokenRequest.Session)
	if stderr.Is(err, core.ErrInvalidAuthorizationCode) {
		if authorizeRequest == nil {
			return core.ErrServerError.WithHint("Misconfigured code lead to an error that prohibited the OAuth 2.0 Framework from processing this request.").
				WithDebug("GetAuthorizationCodeSession must return a value for when returning \"ErrInvalidatedAuthorizeCode\".")
		}

		requestID := authorizeRequest.ID
		hint := "The authorization code has already been used."
		debug := ""

		if re := h.tokenStorage.RevokeAccessToken(ctx, requestID); re != nil {
			hint += " Additionally, an error occurred during processing the access token revocation."
			debug += "Revocation of access_token lead to error " + re.Error() + "."
		}

		if re := h.tokenStorage.RevokeRefreshToken(ctx, requestID); re != nil {
			hint += " Additionally, an error occurred during processing the refresh token revocation."
			debug += "Revocation of refresh_token lead to error " + re.Error() + "."
		}

		return core.ErrInvalidGrant.WithHint(hint).WithDebug(debug)
	} else if stderr.Is(err, core.ErrNotFound) {
		return core.ErrInvalidGrant.WithWrap(err).WithDebug(err.Error())
	} else if err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	if err = h.tokenStrategy.ValidateAuthorizeCode(ctx, authorizeRequest, code); err != nil {
		return core.ErrInvalidGrant.WithWrap(err).WithDebug(err.Error())
	}

	// overwrite the request
	tokenRequest.ID = authorizeRequest.ID
	tokenRequest.Session = authorizeRequest.Session
	tokenRequest.RequestedScope = authorizeRequest.RequestedScope
	tokenRequest.RequestedAudience = authorizeRequest.RequestedAudience

	if authorizeRequest.Client.GetID() != client.GetID() {
		return core.ErrInvalidGrant.WithHint("The OAuth 2.0 Client ID from this request does not match the one from the authorize request.")
	}

	return nil
}

func (h *AuthorizationCodeGrantHandler) HandleTokenResponse(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
	}

	return nil
}
