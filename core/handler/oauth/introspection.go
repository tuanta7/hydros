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

type TokenIntrospectionHandler struct {
	config               TokenIntrospectionConfigurator
	scopeStrategy        strategy.ScopeStrategy
	accessTokenStrategy  strategy.TokenStrategy
	refreshTokenStrategy strategy.TokenStrategy
	accessTokenStorage   storage.AccessTokenStorage
	refreshTokenStorage  storage.RefreshTokenStorage
}

func NewTokenIntrospectionHandler(
	config TokenIntrospectionConfigurator,
	accessTokenStrategy strategy.TokenStrategy,
	refreshTokenStrategy strategy.TokenStrategy,
	accessTokenStorage storage.AccessTokenStorage,
	refreshTokenStorage storage.RefreshTokenStorage,
) *TokenIntrospectionHandler {
	return &TokenIntrospectionHandler{
		config:               config,
		scopeStrategy:        strategy.ExactScopeStrategy,
		accessTokenStrategy:  accessTokenStrategy,
		refreshTokenStrategy: refreshTokenStrategy,
		accessTokenStorage:   accessTokenStorage,
		refreshTokenStorage:  refreshTokenStorage,
	}
}

func (h *TokenIntrospectionHandler) IntrospectToken(
	ctx context.Context,
	ir *core.IntrospectionRequest,
	tr *core.TokenRequest,
) (core.TokenType, error) {
	if h.config.IsDisableRefreshTokenValidation() {
		signature := h.accessTokenStrategy.GetSignature(ir.Token)
		_, err := h.accessTokenStorage.GetAccessTokenSession(ctx, signature, tr.Session)
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
		signature := h.refreshTokenStrategy.GetSignature(ir.Token)
		tokenRequestDB, err = h.refreshTokenStorage.GetRefreshTokenSession(ctx, signature, tr.Session)
	case core.IDToken:
		tokenType = core.IDToken
		fallthrough
	default:
		// default to access token
		tokenType = core.AccessToken
		signature := h.accessTokenStrategy.GetSignature(ir.Token)
		tokenRequestDB, err = h.accessTokenStorage.GetAccessTokenSession(ctx, signature, tr.Session)
	}

	if err != nil {
		return "", err
	}

	for _, scope := range ir.Scope {
		if !h.scopeStrategy(tokenRequestDB.GrantedScope, scope) {
			return "", core.ErrInvalidScope.WithHint("The request scope '%s' has not been granted or is not allowed to be requested.", scope)
		}
	}

	mergeRequests(tr, tokenRequestDB)
	return tokenType, nil
}

// mergeRequests merges back old request values into the current one.
func mergeRequests(curr, old *core.TokenRequest) {
	curr.ID = old.ID
	curr.RequestedAt = old.RequestedAt
	curr.Scope = curr.Scope.Append(old.Scope...)
	curr.GrantedScope = curr.GrantedScope.Append(old.GrantedScope...)
	curr.Audience = curr.Audience.Append(old.Audience...)
	curr.GrantedAudience = curr.GrantedAudience.Append(old.GrantedAudience...)
	curr.Client = old.Client
	curr.Session = old.Session

	for k, v := range old.Form {
		curr.Form[k] = v
	}
}
