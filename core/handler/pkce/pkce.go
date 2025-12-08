package pkce

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	stderr "errors"
	"regexp"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
)

type ProofKeyForCodeExchangeConfigurator interface {
	core.EnablePKCEPlainChallengeMethodProvider
}

type ProofKeyForCodeExchangeHandler struct {
	config       ProofKeyForCodeExchangeConfigurator
	codeStrategy strategy.AuthorizeCodeStrategy
	storage      storage.PKCERequestStorage
}

func NewProofKeyForCodeExchangeHandler(
	config ProofKeyForCodeExchangeConfigurator,
	codeStrategy strategy.AuthorizeCodeStrategy,
	storage storage.PKCERequestStorage,
) *ProofKeyForCodeExchangeHandler {
	return &ProofKeyForCodeExchangeHandler{
		config:       config,
		storage:      storage,
		codeStrategy: codeStrategy,
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
	res *core.AuthorizeResponse,
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

	if len(res.Code) == 0 {
		return core.ErrMisconfiguration.WithDebug("The authorization code has not been issued yet, indicating a broken code configuration.")
	}

	signature := h.codeStrategy.AuthorizeCodeSignature(ctx, res.Code)
	if err := h.storage.CreatePKCERequestSession(ctx, signature, req.Request.Sanitize(
		"code_challenge",
		"code_challenge_method",
	)); err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	return nil
}

var verifierWrongFormat = regexp.MustCompile("[^\\w.\\-~]") //

func (h *ProofKeyForCodeExchangeHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
	}

	codeSignature := h.codeStrategy.AuthorizeCodeSignature(ctx, req.Code)
	pkceRequest, err := h.storage.GetPKCERequestSession(ctx, codeSignature, req.Session)
	if stderr.Is(err, core.ErrNotFound) {
		return core.ErrInvalidGrant.WithHint("Unable to find initial PKCE data tied to this request").WithWrap(err).WithDebug(err.Error())
	} else if err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	if err = h.storage.DeletePKCERequestSession(ctx, codeSignature); err != nil {
		return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
	}

	challenge := pkceRequest.Form.Get("code_challenge")
	method := pkceRequest.Form.Get("code_challenge_method")

	if len(challenge) == 0 {
		return core.ErrInvalidRequest.WithHint("The PKCE code challenge is missing.")
	}

	if err = h.validateChallengeMethod(method); err != nil {
		return err
	}

	verifier := req.CodeVerifier
	if l := len(verifier); l < 43 || l > 128 {
		return core.ErrInvalidRequest.WithHint("The PKCE code verifier must be between 43 and 128 characters long.")
	} else if verifierWrongFormat.MatchString(verifier) {
		return core.ErrInvalidGrant.WithHint("The PKCE code verifier must only contain [a-Z], [0-9], '-', '.', '_', '~'.")
	}

	switch method {
	case "S256":
		hash := sha256.New()
		if _, err = hash.Write([]byte(verifier)); err != nil {
			return core.ErrServerError.WithWrap(err).WithDebug(err.Error())
		}

		if base64.RawURLEncoding.EncodeToString(hash.Sum([]byte{})) != challenge {
			return core.ErrInvalidGrant.WithHint("The PKCE code verifier does not match the challenge.")
		}

	case "plain":
		fallthrough
	default:
		if verifier != challenge {
			return core.ErrInvalidGrant.WithHint("The PKCE code verifier does not match the challenge.")
		}
	}

	return nil
}

func (h *ProofKeyForCodeExchangeHandler) HandleTokenResponse(
	ctx context.Context,
	req *core.TokenRequest,
	_ *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("authorization_code") {
		return core.ErrUnknownRequest
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
