package oidc

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type OpenIDConnectAuthorizationCodeFlowHandler struct{}

func NewOpenIDConnectAuthorizationCodeFlowHandler() *OpenIDConnectAuthorizationCodeFlowHandler {
	return &OpenIDConnectAuthorizationCodeFlowHandler{}
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleAuthorizeRequest(
	ctx context.Context,
	req *core.AuthorizeRequest,
	resp *core.AuthorizeResponse,
) error {
	return nil
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	return nil
}

func (h *OpenIDConnectAuthorizationCodeFlowHandler) HandleTokenResponse(ctx context.Context, req *core.TokenRequest, res *core.TokenResponse) error {
	return nil
}
