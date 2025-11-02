package oidc

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

var oidcParameters = []string{"grant_type", "max_age", "prompt", "acr_values", "id_token_hint", "nonce"}

type OpenIDConnectAuthorizationCodeFlowHandler struct {
}

func NewOpenIDConnectAuthorizationCodeFlowHandler() *OpenIDConnectAuthorizationCodeFlowHandler {
	return &OpenIDConnectAuthorizationCodeFlowHandler{}
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleAuthorizeRequest(ctx context.Context, req *core.AuthorizeRequest) error {
	if !req.Scope.IncludeAll("openid") {
		return nil
	}

	if req.RedirectURI.String() == "" {
		return core.ErrInvalidRequest.WithHint("The 'redirect_uri' parameter is required when using OpenID Connect 1.0.")
	}

	requestURI := req.Form.Get("request_uri")
	request := req.Form.Get("request")
	if request == "" && requestURI == "" {
		return nil
	}

	// TODO: support JWT Request Object
	if request != "" && requestURI != "" {
		return core.ErrInvalidRequest.WithHint("OpenID Connect parameters 'request' and 'request_uri' were both given, but you can use at most one.")
	}

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
	resp *core.AuthorizeResponse,
) error {
	if req.ResponseTypes.ExactOne("code") && !(req.Scope.IncludeAll("openid")) {
		return nil
	}

	return nil
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	return nil
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleTokenResponse(ctx context.Context, req *core.TokenRequest, res *core.TokenResponse) error {
	return nil
}
