package oidc

import (
	"context"
	"strconv"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

type OpenIDConnectAuthorizationCodeFlowHandler struct {
	config        OpenIDConnectPromptConfigurator
	tokenStrategy strategy.OpenIDConnectTokenStrategy
	storage       storage.OpenIDConnectRequestStorage
}

func NewOpenIDConnectAuthorizationCodeFlowHandler(
	config OpenIDConnectPromptConfigurator,
	tokenStrategy strategy.OpenIDConnectTokenStrategy,
	storage storage.OpenIDConnectRequestStorage,
) *OpenIDConnectAuthorizationCodeFlowHandler {
	return &OpenIDConnectAuthorizationCodeFlowHandler{
		config:        config,
		tokenStrategy: tokenStrategy,
		storage:       storage,
	}
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleAuthorizeRequest(ctx context.Context, req *core.AuthorizeRequest) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core.ErrUnsupportedResponseType
	}

	if !req.RequestedScope.IncludeAll("openid") {
		return nil // not an openid request
	}

	req.Prompt = x.SplitSpace(req.Form.Get("prompt"))
	req.Nonce = req.Form.Get("nonce")
	req.MaxAge = -1
	if m := req.Form.Get("max_age"); len(m) > 0 {
		maxAge, err := strconv.ParseInt(m, 10, 64)
		if err != nil {
			return core.ErrInvalidRequest.WithHint("Invalid value for 'max_age' parameter").WithWrap(err)
		}

		req.MaxAge = maxAge
	}

	requestURI := req.Form.Get("request_uri")
	request := req.Form.Get("request")
	if request == "" && requestURI == "" {
		return nil
	}

	if request != "" && requestURI != "" {
		return core.ErrInvalidRequest.WithHint("OpenID Connect parameters 'request' and 'request_uri' were both given, but you can use at most one.")
	}

	// TODO: support JWT Request Object
	oidcClient, ok := req.Client.(core.OpenIDConnectClient)
	if !ok {
		if requestURI != "" {
			return core.ErrRequestNotSupported.WithHint("OpenID Connect 'request_uri' context was given, but the  OAuth 2.0 Client does not implement advanced OpenID Connect capabilities.")
		}
		return core.ErrRequestNotSupported.WithHint("OpenID Connect 'request' context was given, but the  OAuth 2.0 Client does not implement advanced OpenID Connect capabilities.")
	}

	if oidcClient.GetJWKs() == nil || oidcClient.GetJWKsURI() == "" {
		return core.ErrInvalidRequest.WithHint("OpenID Connect 'request' or 'request_uri' context was given, but the OAuth 2.0 Client does not have any JSON Web Keys registered.")
	}

	return nil
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleAuthorizeResponse(
	ctx context.Context,
	req *core.AuthorizeRequest,
	res *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core.ErrUnsupportedResponseType
	}

	if !req.RequestedScope.IncludeAll("openid") {
		return nil
	}

	if len(res.Code) == 0 {
		return core.ErrMisconfiguration.WithDebug("The authorization code has not been issued yet, indicating a broken code configuration.")
	}

	if req.RedirectURI.String() == "" {
		return core.ErrInvalidRequest.WithHint("The 'redirect_uri' parameter is required when using OpenID Connect 1.0.")
	}

	if err := validatePrompt(h.config, req, h.tokenStrategy); err != nil {
		return err
	}

	if err := h.storage.CreateOpenIDConnectSession(ctx, res.Code, req.Sanitize(
		"grant_type",
		"max_age",
		"prompt",
		"acr_values",
		"id_token_hint",
		"nonce",
	)); err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	return nil
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	return core.ErrUnknownRequest
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleTokenResponse(ctx context.Context, req *core.TokenRequest, res *core.TokenResponse) error {
	token, err := h.tokenStrategy.GenerateIDToken(ctx, 0, req)
	if err != nil {
		return err
	}

	res.IDToken = token
	return nil
}
