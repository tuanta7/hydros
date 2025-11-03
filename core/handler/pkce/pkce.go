package pkce

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
)

type ProofKeyForCodeExchangeConfigurator interface {
	core.EnablePKCEPlainChallengeMethodProvider
}

type ProofKeyForCodeExchangeHandler struct {
	config       ProofKeyForCodeExchangeConfigurator
	storage      storage.PKCERequestStorage
	codeStrategy strategy.AuthorizeCodeStrategy
}

func NewProofKeyForCodeExchangeHandler(config ProofKeyForCodeExchangeConfigurator) *ProofKeyForCodeExchangeHandler {
	return &ProofKeyForCodeExchangeHandler{
		config: config,
	}
}

func (h *ProofKeyForCodeExchangeHandler) HandleAuthorizeRequest(ctx context.Context, req *core.AuthorizeRequest) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core.ErrUnsupportedResponseType
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) HandleAuthorizeResponse(
	ctx context.Context,
	req *core.AuthorizeRequest,
	resp *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.ExactOne("code") {
		return core.ErrUnsupportedResponseType
	}

	client := req.Client
	if client == nil {
		// should never happen because NewAuthorizeRequest already checks this
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	if req.CodeChallenge == "" {
		return core.ErrInvalidRequest.WithHint("PKCE is required for OAuth 2.1.")
	}

	if err := h.validateChallengeMethod(req.CodeChallengeMethod); err != nil {
		return err
	} else if req.CodeChallengeMethod == "" {
		req.CodeChallengeMethod = "plain"
	}

	code := resp.Code
	if code == "" {
		return core.ErrServerError.WithDebug("The PKCE handler must be loaded after the authorize code handler.")
	}

	signature := h.codeStrategy.AuthorizeCodeSignature(ctx, code)
	if err := h.storage.CreatePKCERequestSession(ctx, signature, req.Request.Sanitize(
		"code_challenge",
		"code_challenge_method",
	)); err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) validateChallengeMethod(challengeMethod string) error {
	switch challengeMethod {
	case "S256":
		return nil
	case "plain":
		fallthrough
	case "":
		if h.config.IsEnablePKCEPlainChallengeMethod() {
			return nil
		}

		return core.ErrInvalidRequest.
			WithHint("Clients must use code_challenge_method=S256, plain is not allowed.").
			WithDebug("The server is configured in a way that enforces PKCE S256 as challenge method for clients.")
	}

	return core.ErrInvalidRequest.WithHint("The code_challenge_method is not supported, use S256 instead.")
}

func (h *ProofKeyForCodeExchangeHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) HandleTokenResponse(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	return nil
}
