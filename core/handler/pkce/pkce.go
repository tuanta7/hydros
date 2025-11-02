package pkce

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
)

type ProofKeyForCodeExchangeConfigurator interface {
	core.EnforcePKCEProvider
	core.EnforcePKCEForPublicClientsProvider
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
	if !req.ResponseTypes.IncludeAll("code") {
		return nil
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) HandleAuthorizeResponse(
	ctx context.Context,
	req *core.AuthorizeRequest,
	resp *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.IncludeAll("code") {
		return nil
	}

	client := req.Client
	if client == nil {
		// should never happen because NewAuthorizeRequest already checks this
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	challenge := req.CodeChallenge
	err := h.validateChallenge(challenge, client)
	if err != nil {
		return err
	}

	challengeMethod := req.CodeChallenge
	err = h.validateMethod(challengeMethod)
	if err != nil {
		return err
	}

	if challenge == "" && challengeMethod == "" {
		return nil
	}

	code := resp.Code
	if code == "" {
		return core.ErrServerError.WithDebug("The PKCE handler must be loaded after the authorize code handler.")
	}

	signature := h.codeStrategy.AuthorizeCodeSignature(ctx, code)
	err = h.storage.CreatePKCERequestSession(ctx, signature, req)
	if err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) validateChallenge(challenge string, client core.Client) error {
	if challenge != "" {
		return nil
	}

	if h.config.GetEnforcePKCE() {
		return core.ErrInvalidRequest.
			WithHint("Clients must include a code_challenge when performing the authorize code flow, but it is missing.").
			WithDebug("The server is configured in a way that enforces PKCE for clients.")
	}

	if h.config.GetEnforcePKCEForPublicClients() && client.IsPublic() {
		return core.ErrInvalidRequest.
			WithHint("This client must include a code_challenge when performing the authorize code flow, but it is missing.").
			WithDebug("The server is configured in a way that enforces PKCE for this client.")
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) validateMethod(challengeMethod string) error {
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
