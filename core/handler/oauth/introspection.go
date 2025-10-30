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

type TokenStorage interface {
	storage.AccessTokenStorage
	storage.RefreshTokenStorage
}

type TokenIntrospectionHandler struct {
	config        TokenIntrospectionConfigurator
	scopeStrategy strategy.ScopeStrategy
	tokenStrategy strategy.TokenStrategy
	tokenStorage  TokenStorage
}

func NewTokenIntrospectionHandler(
	config TokenIntrospectionConfigurator,
	tokenStrategy strategy.TokenStrategy,
	tokenStorage TokenStorage,

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
) (core.TokenType, error) {
	if h.config.IsDisableRefreshTokenValidation() {
		signature := h.tokenStrategy.AccessTokenSignature(ctx, ir.Token)
		_, err := h.tokenStorage.GetAccessTokenSession(ctx, signature, tr.Session)
		if err != nil {
			return "", err
		}

		return core.AccessToken, nil
	}

	var tokenRequestDB *core.TokenRequest
	var err error
	var tokenType core.TokenType

	switch ir.TokenTypeHint {
	case core.RefreshToken:
		tokenType = core.RefreshToken
		signature := h.tokenStrategy.RefreshTokenSignature(ctx, ir.Token)
		tokenRequestDB, err = h.tokenStorage.GetRefreshTokenSession(ctx, signature, tr.Session)
		if err != nil {
			return "", err
		}

		err = h.tokenStrategy.ValidateRefreshToken(ctx, tokenRequestDB, ir.Token)
		if err != nil {
			return "", err
		}

	case core.IDToken:
		tokenType = core.IDToken
		fallthrough
	default:
		// default to access token
		tokenType = core.AccessToken
		signature := h.tokenStrategy.AccessTokenSignature(ctx, ir.Token)
		tokenRequestDB, err = h.tokenStorage.GetAccessTokenSession(ctx, signature, tr.Session)
		if err != nil {
			return "", err
		}

		err = h.tokenStrategy.ValidateAccessToken(ctx, tokenRequestDB, ir.Token)
		if err != nil {
			return "", err
		}
	}

	for _, scope := range ir.Scope {
		if !h.scopeStrategy(tokenRequestDB.GrantedScope, scope) {
			return "", core.ErrInvalidScope.WithHint("The request scope '%s' has not been granted or is not allowed to be requested.", scope)
		}
	}

	tr.Merge(&tokenRequestDB.Request)
	return tokenType, nil
}
