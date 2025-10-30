package oauth

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/strategy"
)

// JWTIntrospectionHandler is used to quickly eliminate invalid tokens so we don't have to query the database.
type JWTIntrospectionHandler struct {
	tokenStrategy strategy.AccessTokenStrategy
}

func NewJWTIntrospectionHandler(tokenStrategy strategy.AccessTokenStrategy) *JWTIntrospectionHandler {
	return &JWTIntrospectionHandler{
		tokenStrategy: tokenStrategy,
	}
}

func (h *JWTIntrospectionHandler) IntrospectToken(
	ctx context.Context,
	ir *core.IntrospectionRequest,
	tr *core.TokenRequest,
) (core.TokenType, error) {
	sig := h.tokenStrategy.AccessTokenSignature(ctx, ir.Token)
	if sig == "" {
		// this might be an opaque token
		return "", core.ErrUnknownRequest
	}

	err := h.tokenStrategy.ValidateAccessToken(ctx, tr, ir.Token)
	if err != nil {
		return "", err
	}

	return core.AccessToken, nil
}
