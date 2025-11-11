package oauth

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
)

type TokenIntrospectionConfigurator interface {
	core.DisableRefreshTokenValidationProvider
}

type TokenIntrospectionStorage interface {
	storage.AccessTokenStorage
	storage.RefreshTokenStorage
}

// TokenIntrospectionHandler will validate the token and return the token information. This handler is always needed
// to run since we need to check the token revocation status.
type TokenIntrospectionHandler struct {
	config        TokenIntrospectionConfigurator
	scopeStrategy strategy.ScopeStrategy
	tokenStrategy strategy.TokenStrategy
	tokenStorage  TokenIntrospectionStorage
}

func NewTokenIntrospectionHandler(
	config TokenIntrospectionConfigurator,
	tokenStrategy strategy.TokenStrategy,
	tokenStorage TokenIntrospectionStorage,

) *TokenIntrospectionHandler {
	return &TokenIntrospectionHandler{
		config:        config,
		scopeStrategy: strategy.ExactScopeStrategy,
		tokenStrategy: tokenStrategy,
		tokenStorage:  tokenStorage,
	}
}

func (h *TokenIntrospectionHandler) IntrospectToken(
	ctx context.Context,
	ir *core.IntrospectionRequest,
	tr *core.TokenRequest,
) (tokenType core.TokenType, err error) {
	if h.config.IsDisableRefreshTokenValidation() {
		signature := h.tokenStrategy.AccessTokenSignature(ctx, ir.Token)
		_, err = h.tokenStorage.GetAccessTokenSession(ctx, signature, tr.Session)
		if err != nil {
			return "", err
		}

		return core.AccessToken, nil
	}

	var handledTokenRequest *core.Request

	switch ir.TokenTypeHint {
	case core.RefreshToken:
		tokenType = core.RefreshToken
		signature := h.tokenStrategy.RefreshTokenSignature(ctx, ir.Token)
		handledTokenRequest, err = h.tokenStorage.GetRefreshTokenSession(ctx, signature, tr.Session)
		if err != nil {
			return "", err
		}

		err = h.tokenStrategy.ValidateRefreshToken(ctx, handledTokenRequest, ir.Token)
		if err != nil {
			return "", err
		}
	default:
		tokenType = core.AccessToken
		signature := h.tokenStrategy.AccessTokenSignature(ctx, ir.Token)
		handledTokenRequest, err = h.tokenStorage.GetAccessTokenSession(ctx, signature, tr.Session)
		if err != nil {
			return "", err
		}

		err = h.tokenStrategy.ValidateAccessToken(ctx, handledTokenRequest, ir.Token)
		if err != nil {
			return "", err
		}
	}

	if err = h.scopeStrategy(handledTokenRequest.GrantedScope, ir.Scope); err != nil {
		return "", err
	}

	tr.Merge(handledTokenRequest)
	return tokenType, nil
}
